package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/ascii85"
	"encoding/base32"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
)

/* ---------------- Base32（RFC 4648，标准字母表与 Extended Hex 字母表） ---------------- */

// utilBase32 Base32 编解码工具（对应 hutool Base32）。
// 标准 5 个字节编码为 8 个字符，不足 8n 用 "=" 补足，体积膨胀 3/5。
type utilBase32 struct{}

// Encode 标准 Base32 编码（字母表 A-Z2-7，带 "=" 填充，对应 hutool Base32.encode）。
// 场景：对二进制数据（如密钥、令牌）做大小写不敏感友好的文本化传输。
func (b *utilBase32) Encode(data []byte) string {
	return base32.StdEncoding.EncodeToString(data)
}

// EncodeHex 扩展十六进制风格 Base32 编码（字母表 0-9A-V，对应 hutool Base32.encodeHex）。
// 场景：编码结果需要与十六进制排序/展示习惯一致的场合（如 DNS 类系统）。
func (b *utilBase32) EncodeHex(data []byte) string {
	return base32.HexEncoding.EncodeToString(data)
}

// Decode 标准 Base32 解码，输入含非法字符或填充错误时返回 error。
func (b *utilBase32) Decode(s string) ([]byte, error) {
	return base32.StdEncoding.DecodeString(s)
}

// DecodeHex 扩展十六进制风格 Base32 解码。
func (b *utilBase32) DecodeHex(s string) ([]byte, error) {
	return base32.HexEncoding.DecodeString(s)
}

/* ---------------- Base58（Bitcoin 字母表，剔除易混淆的 0 O I l） ---------------- */

// Base58 编码字母表（Bitcoin 规范）
const b58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// b58Lookup 字符 -> 数值查找表，-1 表示非法字符
var b58Lookup = func() *[128]int8 {
	var table [128]int8
	for i := range table {
		table[i] = -1
	}
	for i := 0; i < len(b58Alphabet); i++ {
		table[b58Alphabet[i]] = int8(i)
	}
	return &table
}()

// utilBase58 Base58 编解码工具（对应 hutool Base58）。
// 规范见 https://en.bitcoin.it/wiki/Base58Check_encoding，
// 常用于比特币地址、IPFS CID 等需要人工抄写/比对的编码场景。
type utilBase58 struct{}

// Encode Base58 编码（大整数除法实现，前导 0x00 字节逐个映射为 '1'，对应 hutool Base58.encode）。
// 场景：把公钥哈希、交易 ID 等转为便于人工复制的字符串。
func (b *utilBase58) Encode(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	// 统计前导 0 的个数，每个 0x00 编码为字母表首字符 '1'
	zeroCount := 0
	for zeroCount < len(data) && data[zeroCount] == 0 {
		zeroCount++
	}
	// divmod 会原地修改数据，拷贝一份避免污染入参
	buf := make([]byte, len(data))
	copy(buf, data)
	encoded := make([]byte, len(buf)*2) // 58 进制比 256 进制膨胀不超过 1.38 倍，2 倍长度是上界
	outputStart := len(encoded)
	for inputStart := zeroCount; inputStart < len(buf); {
		outputStart--
		encoded[outputStart] = b58Alphabet[b58Divmod(buf, inputStart, 256, 58)]
		if buf[inputStart] == 0 {
			inputStart++ // 跳过已成 0 的前导位，优化除法次数
		}
	}
	// 跳过大数运算产生的前导 '1'（真正的前导 0 由下面循环负责还原）
	for outputStart < len(encoded) && encoded[outputStart] == b58Alphabet[0] {
		outputStart++
	}
	for zeroCount > 0 {
		zeroCount--
		outputStart--
		encoded[outputStart] = b58Alphabet[0]
	}
	return string(encoded[outputStart:])
}

// EncodeChecked 带校验和的 Base58 编码（Base58Check）：
// 输出 = version 字节 + data + checksum，其中 checksum 为 data 做两次 SHA-256 后取前 4 字节
// （对应 hutool Base58.encodeChecked）。
// 场景：生成钱包地址、需要防转录错误的短标识。
func (b *utilBase58) EncodeChecked(version byte, data []byte) string {
	buf := make([]byte, 0, len(data)+5)
	buf = append(buf, version)
	buf = append(buf, data...)
	buf = append(buf, b58Checksum(data)...)
	return b.Encode(buf)
}

// Decode Base58 解码，含非法字符（如 0、O、I、l、空格等）时返回 error。
func (b *utilBase58) Decode(s string) ([]byte, error) {
	if len(s) == 0 {
		return []byte{}, nil
	}
	// 先把 Base58 字符转成 58 进制的大数位
	input58 := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		digit := int8(-1)
		if c < 128 {
			digit = b58Lookup[c]
		}
		if digit < 0 {
			return nil, fmt.Errorf("base58 含非法字符 %q（位置 %d）", rune(c), i)
		}
		input58[i] = byte(digit)
	}
	// 统计前导 0（对应编码时的 '1'）
	zeros := 0
	for zeros < len(input58) && input58[zeros] == 0 {
		zeros++
	}
	// 58 进制转 256 进制
	decoded := make([]byte, len(s))
	outputStart := len(decoded)
	for inputStart := zeros; inputStart < len(input58); {
		outputStart--
		decoded[outputStart] = b58Divmod(input58, inputStart, 58, 256)
		if input58[inputStart] == 0 {
			inputStart++
		}
	}
	// 丢弃计算过程中多出来的前导 0
	for outputStart < len(decoded) && decoded[outputStart] == 0 {
		outputStart++
	}
	return decoded[outputStart-zeros:], nil
}

// DecodeChecked 校验并去除 Base58Check 的校验和（对应 hutool Base58.decodeChecked）：
//   - withVersion=true：结果保留 version 字节（原始 version + 载荷）；
//   - withVersion=false：结果去掉首字节，仅返回载荷。
//
// checksum 与末尾 4 字节不匹配、或数据过短时返回 error。
func (b *utilBase58) DecodeChecked(s string, withVersion bool) ([]byte, error) {
	data, err := b.Decode(s)
	if err != nil {
		return nil, err
	}
	if len(data) < 4 {
		return nil, errors.New("base58 数据过短，缺少 4 字节校验和")
	}
	end := len(data) - 4
	start := 0
	if withVersion {
		start = 1
	}
	if start > end {
		return nil, errors.New("base58 数据过短，无法提取载荷")
	}
	payload := data[start:end]
	if !bytes.Equal(b58Checksum(payload), data[end:]) {
		return nil, errors.New("base58 校验和（checksum）不匹配")
	}
	return payload, nil
}

// b58Divmod 大数除法：number 为 base 进制的大数位（每字节一位），
// 原地改为商并返回余数。firstDigit 指向第一个非零位以跳过前导零。
func b58Divmod(number []byte, firstDigit, base, divisor int) byte {
	remainder := 0
	for i := firstDigit; i < len(number); i++ {
		digit := int(number[i]) & 0xFF
		temp := remainder*base + digit
		number[i] = byte(temp / divisor)
		remainder = temp % divisor
	}
	return byte(remainder)
}

// b58Checksum 两次 SHA-256 后取前 4 字节作为校验和
func b58Checksum(data []byte) []byte {
	first := sha256.Sum256(data)
	second := sha256.Sum256(first[:])
	return second[:4]
}

/* ---------------- Base62（GMP 标准字母表与反转字母表，常用于短 URL） ---------------- */

// Base62 两种字母表：标准（0-9A-Za-z）与反转（0-9a-zA-Z），对照 hutool Base62Codec
const (
	b62AlphabetStandard = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	b62AlphabetInverted = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// 两种字母表对应的解码查找表，-1 表示非法字符
var (
	b62LookupStandard = b62BuildLookup(b62AlphabetStandard)
	b62LookupInverted = b62BuildLookup(b62AlphabetInverted)
)

// b62BuildLookup 构建字符 -> 数值查找表
func b62BuildLookup(alphabet string) *[128]int8 {
	var table [128]int8
	for i := range table {
		table[i] = -1
	}
	for i := 0; i < len(alphabet); i++ {
		table[alphabet[i]] = int8(i)
	}
	return &table
}

// utilBase62 Base62 编解码工具（对应 hutool Base62，底层来自 seruco/base62）。
// 常用于短链 ID、把二进制摘要压成更短的字符串。
type utilBase62 struct{}

// Encode Base62 编码（标准字母表 0-9A-Za-z，GMP 风格，对应 hutool Base62.encode）。
// 原始数据的前导 0x00 字节会映射为前导 '0' 字符。
// 场景：短链服务把自增 ID 的字节形式压成 62 进制短码。
func (b *utilBase62) Encode(data []byte) string {
	return b62EncodeWith(data, b62AlphabetStandard)
}

// EncodeInverted Base62 编码（反转字母表 0-9a-zA-Z，对应 hutool Base62.encodeInverted）。
// 场景：与部分 PHP/JS 库互通时使用的大小写相反风格。
func (b *utilBase62) EncodeInverted(data []byte) string {
	return b62EncodeWith(data, b62AlphabetInverted)
}

// Decode 标准 Base62 解码，含非法字符时返回 error。
func (b *utilBase62) Decode(s string) ([]byte, error) {
	return b62DecodeWith(s, b62LookupStandard)
}

// DecodeInverted 反转字母表 Base62 解码，含非法字符时返回 error。
func (b *utilBase62) DecodeInverted(s string) ([]byte, error) {
	return b62DecodeWith(s, b62LookupInverted)
}

// b62EncodeWith 指定字母表编码
func b62EncodeWith(data []byte, alphabet string) string {
	indices := b62Convert(data, 256, 62)
	out := make([]byte, len(indices))
	for i, v := range indices {
		out[i] = alphabet[v]
	}
	return string(out)
}

// b62DecodeWith 指定查找表解码
func b62DecodeWith(encoded string, lookup *[128]int8) ([]byte, error) {
	prepared := make([]byte, len(encoded))
	for i := 0; i < len(encoded); i++ {
		c := encoded[i]
		if c >= 128 || lookup[c] < 0 {
			return nil, fmt.Errorf("base62 含非法字符 %q（位置 %d）", rune(c), i)
		}
		prepared[i] = byte(lookup[c])
	}
	return b62Convert(prepared, 62, 256), nil
}

// b62Convert 任意进制大数转换（对照 hutool Base62Codec.convert）：
// 把 sourceBase 进制的大数位转成 targetBase 进制，结果高位在前，
// 并按输入的前导 0 字节个数在结果高位补 0。
func b62Convert(message []byte, sourceBase, targetBase int) []byte {
	var out []byte
	source := message
	for len(source) > 0 {
		quotient := make([]byte, 0, len(source))
		remainder := 0
		for _, b := range source {
			acc := int(b) + remainder*sourceBase
			digit := (acc - acc%targetBase) / targetBase
			remainder = acc % targetBase
			// 商的前导 0 不写入，保持大数位紧凑
			if len(quotient) > 0 || digit > 0 {
				quotient = append(quotient, byte(digit))
			}
		}
		out = append(out, byte(remainder))
		source = quotient
	}
	// 输入的前导 0 字节在输出高位补 0（全 0 输入保留长度-1 个）
	for i := 0; i < len(message)-1 && message[i] == 0; i++ {
		out = append(out, 0)
	}
	// 低位先产出，翻转为高位在前
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

/* ---------------- Base85（Ascii85 与 Z85 两种风格） ---------------- */

// Z85 字母表（ZeroMQ RFC 32，剔除引号、反斜杠等对代码不友好的字符），对照 hutool Base85Codec
const b85Z85Alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ.-:+=^!/*?&<>()[]{}@%$#"

// b85Z85Lookup Z85 解码查找表，-1 表示非法字符
var b85Z85Lookup = func() *[128]int8 {
	var table [128]int8
	for i := range table {
		table[i] = -1
	}
	for i := 0; i < len(b85Z85Alphabet); i++ {
		table[b85Z85Alphabet[i]] = int8(i)
	}
	return &table
}()

// utilBase85 Base85 编解码工具（对应 hutool Base85）。
// 相比 Base64 体积再省约 7%，编解码无需大数运算，速度更快。
type utilBase85 struct{}

// Encode Ascii85 编码（对应 hutool Base85.encode，标准 "!"~"u" 字母表）。
// 说明：Go 标准库对全 0 的 4 字节组会输出压缩字符 'z'，这是 Ascii85 规范行为。
// 场景：PDF 内嵌图像、git 二进制补丁等使用 Ascii85 的场合。
func (b *utilBase85) Encode(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	buf := make([]byte, ascii85.MaxEncodedLen(len(data)))
	n := ascii85.Encode(buf, data)
	return string(buf[:n])
}

// Decode Ascii85 解码（容忍空白字符，尾部不足 5 字符的分组按 flush 语义还原），
// 含非法字符时返回 error。
func (b *utilBase85) Decode(s string) ([]byte, error) {
	buf := make([]byte, ascii85.MaxEncodedLen(len(s)))
	n, _, err := ascii85.Decode(buf, []byte(s), true)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

// EncodeZ85 Z85 编码（ZeroMQ 字母表，对应 hutool Base85.encodeZ85 / Z85 规范）。
// Z85 要求数据长度为 4 的倍数，否则返回 error。
// 场景：ZeroMQ 的 CURVE 安全握手密钥（公钥/私钥）文本化交换。
func (b *utilBase85) EncodeZ85(data []byte) (string, error) {
	if len(data)%4 != 0 {
		return "", fmt.Errorf("z85 编码要求数据长度为 4 的倍数，当前 %d", len(data))
	}
	out := make([]byte, 0, len(data)/4*5)
	for i := 0; i < len(data); i += 4 {
		tuple := uint32(data[i])<<24 | uint32(data[i+1])<<16 | uint32(data[i+2])<<8 | uint32(data[i+3])
		var chunk [5]byte
		for j := 4; j >= 0; j-- {
			chunk[j] = b85Z85Alphabet[tuple%85]
			tuple /= 85
		}
		out = append(out, chunk[:]...)
	}
	return string(out), nil
}

// DecodeZ85 Z85 解码，长度非 5 的倍数或含非法字符时返回 error。
func (b *utilBase85) DecodeZ85(s string) ([]byte, error) {
	if len(s)%5 != 0 {
		return nil, fmt.Errorf("z85 解码要求字符串长度为 5 的倍数，当前 %d", len(s))
	}
	out := make([]byte, 0, len(s)/5*4)
	for i := 0; i < len(s); i += 5 {
		var tuple uint32
		for j := 0; j < 5; j++ {
			c := s[i+j]
			if c >= 128 || b85Z85Lookup[c] < 0 {
				return nil, fmt.Errorf("z85 含非法字符 %q（位置 %d）", rune(c), i+j)
			}
			tuple = tuple*85 + uint32(b85Z85Lookup[c])
		}
		out = append(out, byte(tuple>>24), byte(tuple>>16), byte(tuple>>8), byte(tuple))
	}
	return out, nil
}

/* ---------------- BCD 码（Binary-Coded Decimal，压缩十进制） ---------------- */

// utilBCD BCD 码工具（对应 hutool BCD）。
// 用 4 个 bit 存一位十进制数，一个字节装两位，常用于电信/银行定长报文。
type utilBCD struct{}

// StrToBcd ASCII 数字串转压缩 BCD 码（对应 hutool BCD.strToBcd）：
// 长度为奇数时按 hutool 规则在左侧补 '0' 凑偶；支持 0-9 与十六进制字母 a-f/A-F
// （hutool 允许字母参与半字节计算），其余字符返回 error。
// StrToBcd("1234") = [0x12,0x34]，StrToBcd("123") = [0x01,0x23]。
// 场景：组装 ISO8583 等定长 BCD 报文字段。
func (b *utilBCD) StrToBcd(asc string) ([]byte, error) {
	for i := 0; i < len(asc); i++ {
		c := asc[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return nil, fmt.Errorf("bcd 含非法字符 %q（仅支持 0-9 与 a-f/A-F）", rune(c))
		}
	}
	if len(asc)%2 != 0 {
		asc = "0" + asc
	}
	out := make([]byte, len(asc)/2)
	for p := 0; p < len(out); p++ {
		out[p] = byte(bcdNibbleValue(asc[2*p])<<4 + bcdNibbleValue(asc[2*p+1]))
	}
	return out, nil
}

// BcdToStr 压缩 BCD 码转数字串（对应 hutool BCD.bcdToStr）：
// 每个字节拆高低半字节逐个映射，0-9 还原为数字，10-15 映射为 'A'-'F'。
// BcdToStr([0x12,0x34]) = "1234"。
// 场景：解析 BCD 报文字段为可读数字串。
func (b *utilBCD) BcdToStr(bcd []byte) string {
	out := make([]byte, len(bcd)*2)
	for i := 0; i < len(bcd); i++ {
		out[i*2] = bcdNibbleChar(int(bcd[i]&0xf0) >> 4)
		out[i*2+1] = bcdNibbleChar(int(bcd[i] & 0x0f))
	}
	return string(out)
}

// bcdNibbleValue ASCII 字符转半字节值（对照 hutool ascToBcd：数字取本身，字母 a-f/A-F 取 10-15）
func bcdNibbleValue(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'z':
		return int(c-'a') + 0x0a
	default:
		return int(c-'A') + 0x0a
	}
}

// bcdNibbleChar 半字节值转字符：0-9 -> '0'-'9'，10-15 -> 'A'-'F'
func bcdNibbleChar(v int) byte {
	if v > 9 {
		return byte(v + 'A' - 10)
	}
	return byte(v + '0')
}

/* ---------------- Caesar 凯撒密码 ---------------- */

// 凯撒密码的 52 字符轮盘表（大小写成对排列），对照 hutool Caesar.TABLE
const codecCaesarTable = "AaBbCcDdEeFfGgHhIiJjKkLlMmNnOoPpQqRrSsTtUuVvWwXxYyZz"

// utilCaesar 凯撒密码工具（对应 hutool Caesar）。
// 仅对表内 52 个大小写字母按轮盘位移，数字、符号、中文等表外字符保持原样。
type utilCaesar struct{}

// Encode 凯撒加密：表内字符沿轮盘循环位移 offset 位（offset 可为负），表外字符原样保留。
// Encode("abc", 1) = "Bbd"（对照 hutool Caesar.encode）。
// 场景：简单混淆、CTF 练习，不具备真实安全性。
func (c *utilCaesar) Encode(message string, offset int) string {
	runes := []rune(message)
	for i, r := range runes {
		pos := strings.IndexRune(codecCaesarTable, r)
		if pos < 0 {
			continue
		}
		runes[i] = codecCaesarShift(pos, offset)
	}
	return string(runes)
}

// Decode 凯撒解密（位移取反），与 Encode 互为逆运算。
func (c *utilCaesar) Decode(cipherText string, offset int) string {
	runes := []rune(cipherText)
	for i, r := range runes {
		pos := strings.IndexRune(codecCaesarTable, r)
		if pos < 0 {
			continue
		}
		runes[i] = codecCaesarShift(pos, -offset)
	}
	return string(runes)
}

// codecCaesarShift 轮盘位移：表内位置 pos 循环移动 offset 位，负数取模修正
func codecCaesarShift(pos, offset int) rune {
	p := (pos + offset) % 52
	if p < 0 {
		p += 52
	}
	return rune(codecCaesarTable[p])
}

/* ---------------- Rot ROT-N 回转密码 ---------------- */

// utilRot RotN 移位密码工具（对应 hutool Rot），凯撒密码的变体，ROT13 最常用。
type utilRot struct{}

// Encode13 ROT13 编码（encodeNumber=true 时数字同时走 ROT5，对应 hutool Rot.encode13）。
// ROT13 的特性是加密即解密，Encode13 与 Decode13 互逆。
// 场景：论坛剧透遮挡、简单隐藏文本。
func (r *utilRot) Encode13(message string, encodeNumber bool) string {
	return r.Encode(message, 13, encodeNumber)
}

// Encode 通用 ROT-N 编码（对应 hutool Rot.encode）：
// 只位移 a-zA-Z（模 26），encodeNumber=true 时数字按模 10 位移，其它字符原样保留。
func (r *utilRot) Encode(message string, offset int, encodeNumber bool) string {
	runes := []rune(message)
	for i, c := range runes {
		if encodeNumber && c >= '0' && c <= '9' {
			runes[i] = rune('0' + (int(c-'0')+offset)%10)
		}
		switch {
		case c >= 'A' && c <= 'Z':
			runes[i] = rune('A' + (int(c-'A')+offset)%26)
		case c >= 'a' && c <= 'z':
			runes[i] = rune('a' + (int(c-'a')+offset)%26)
		}
	}
	return string(runes)
}

// Decode13 ROT13 解码（decodeNumber=true 时同时解码数字，对应 hutool Rot.decode13）。
func (r *utilRot) Decode13(rot string, decodeNumber bool) string {
	return r.Decode(rot, 13, decodeNumber)
}

// Decode 通用 ROT-N 解码（对应 hutool Rot.decode），位移方向与编码相反并做负数修正。
func (r *utilRot) Decode(rot string, offset int, decodeNumber bool) string {
	runes := []rune(rot)
	for i, c := range runes {
		t := int(c)
		if decodeNumber && t >= '0' && t <= '9' {
			t -= '0'
			t -= offset
			for t < 0 {
				t += 10
			}
			t += '0'
		}
		switch {
		case t >= 'A' && t <= 'Z':
			t -= 'A'
			t -= offset
			for t < 0 {
				t += 26
			}
			t += 'A'
		case t >= 'a' && t <= 'z':
			t -= 'a'
			t -= offset
			for t < 0 {
				t += 26
			}
			t += 'a'
		}
		runes[i] = rune(t)
	}
	return string(runes)
}

/* ---------------- Morse 摩斯密码 ---------------- */

// 莫斯电码表（code -> 电码），电码用 0/1 存储：0 为点、1 为划，对照 hutool Morse
var (
	morseAlphabet   = make(map[rune]string, 44) // 字符 -> 电码
	morseDictionary = make(map[string]rune, 44) // 电码 -> 字符
)

func init() {
	// 字母 A-Z
	morseRegister('A', "01")
	morseRegister('B', "1000")
	morseRegister('C', "1010")
	morseRegister('D', "100")
	morseRegister('E', "0")
	morseRegister('F', "0010")
	morseRegister('G', "110")
	morseRegister('H', "0000")
	morseRegister('I', "00")
	morseRegister('J', "0111")
	morseRegister('K', "101")
	morseRegister('L', "0100")
	morseRegister('M', "11")
	morseRegister('N', "10")
	morseRegister('O', "111")
	morseRegister('P', "0110")
	morseRegister('Q', "1101")
	morseRegister('R', "010")
	morseRegister('S', "000")
	morseRegister('T', "1")
	morseRegister('U', "001")
	morseRegister('V', "0001")
	morseRegister('W', "011")
	morseRegister('X', "1001")
	morseRegister('Y', "1011")
	morseRegister('Z', "1100")
	// 数字 0-9
	morseRegister('0', "11111")
	morseRegister('1', "01111")
	morseRegister('2', "00111")
	morseRegister('3', "00011")
	morseRegister('4', "00001")
	morseRegister('5', "00000")
	morseRegister('6', "10000")
	morseRegister('7', "11000")
	morseRegister('8', "11100")
	morseRegister('9', "11110")
	// 常用标点
	morseRegister('.', "010101")
	morseRegister(',', "110011")
	morseRegister('?', "001100")
	morseRegister('\'', "011110")
	morseRegister('!', "101011")
	morseRegister('/', "10010")
	morseRegister('(', "10110")
	morseRegister(')', "101101")
	morseRegister('&', "01000")
	morseRegister(':', "111000")
	morseRegister(';', "101010")
	morseRegister('=', "10001")
	morseRegister('+', "01010")
	morseRegister('-', "100001")
	morseRegister('_', "001101")
	morseRegister('"', "010010")
	morseRegister('$', "0001001")
	morseRegister('@', "011010")
}

// morseRegister 注册一条莫斯电码（0 为点、1 为划）
func morseRegister(abc rune, dict string) {
	morseAlphabet[abc] = dict
	morseDictionary[dict] = abc
}

// utilMorse 摩斯密码工具（对应 hutool Morse）。
// 编码约定：字符间以单个空格分隔，单词（空格）间以 " / " 分隔；
// 电码用 "." 表示点、"-" 表示划；大小写统一按大写处理，未知字符跳过。
type utilMorse struct{}

// Encode 文本转莫斯电码（电码表来自 hutool Morse：大写字母、数字与常用标点）。
// Encode("SOS") = "... --- ..."，Encode("HI YOU") = ".... .. / -.-- --- ..-"。
// 场景：信号演示、题目出题、简单文本趣味编码。
func (m *utilMorse) Encode(text string) string {
	var b strings.Builder
	wroteAny := false       // 是否已输出过有效电码（用于控制分隔符）
	pendingWordSep := false // 是否待插入单词分隔符 " / "
	for _, r := range strings.ToUpper(text) {
		if r == ' ' {
			if wroteAny {
				pendingWordSep = true
			}
			continue
		}
		code, ok := morseAlphabet[r]
		if !ok {
			continue // 未知字符跳过
		}
		if pendingWordSep {
			b.WriteString(" / ")
			pendingWordSep = false
		} else if wroteAny {
			b.WriteByte(' ')
		}
		// 0 -> 点（.），1 -> 划（-）
		b.WriteString(strings.ReplaceAll(strings.ReplaceAll(code, "0", "."), "1", "-"))
		wroteAny = true
	}
	return b.String()
}

// Decode 莫斯电码转文本：按 " / " 拆单词、按空格拆字符后逐个解码，
// 未知电码跳过（对照 hutool 的容错语义）。Decode("... --- ...") = "SOS"。
func (m *utilMorse) Decode(morse string) string {
	words := strings.Split(morse, " / ")
	parts := make([]string, 0, len(words))
	for _, word := range words {
		var sb strings.Builder
		// Fields 同时容忍多余空格/制表符
		for _, letter := range strings.Fields(word) {
			// 点 -> 0，划 -> 1 后查反表
			code := strings.ReplaceAll(strings.ReplaceAll(letter, ".", "0"), "-", "1")
			if c, ok := morseDictionary[code]; ok {
				sb.WriteRune(c)
			}
		}
		if sb.Len() > 0 {
			parts = append(parts, sb.String())
		}
	}
	return strings.Join(parts, " ")
}

/* ---------------- Punycode（RFC 3492，中文域名 xn-- 转码） ---------------- */

// RFC 3492 算法常量，对照 hutool PunyCode
const (
	punyTmin        = 1
	punyTmax        = 26
	punyBase        = 36
	punyInitialN    = 128
	punyInitialBias = 72
	punyDamp        = 700
	punySkew        = 38
	punyDelimiter   = '-'
)

// punyPrefix IDN 的 ACE 前缀
const punyPrefix = "xn--"

// utilPunycode Punycode 编解码工具（RFC 3492，对照 hutool PunyCode）。
// 用于把非 ASCII（如中文）域名标签转成 DNS 可用的 xn-- 形式及其反向还原。
type utilPunycode struct{}

// Encode 将字符串按 RFC 3492 编码为 Punycode（对照 hutool PunyCode.encode）：
// withPrefix=true 时结果加 "xn--" 前缀；纯 ASCII 输入直接原样返回（不加前缀）。
// 场景：中文域名注册/解析前的转码。
func (p *utilPunycode) Encode(input string, withPrefix bool) string {
	runes := []rune(input)
	n, delta, bias := punyInitialN, 0, punyInitialBias
	var output []rune
	// 1. 先把全部基础码点（ASCII）拷贝到输出
	b := 0
	for _, c := range runes {
		if c < 0x80 {
			output = append(output, c)
			b++
		}
	}
	if b > 0 {
		if b == len(runes) {
			// 全部为基础码点，无需编码（对照 hutool 提前返回，不加前缀）
			return string(output)
		}
		output = append(output, punyDelimiter)
	}
	// 2. 逐个处理剩余的非基础码点
	h := b
	for h < len(runes) {
		m := rune(math.MaxInt32)
		for _, c := range runes {
			if c >= rune(n) && c < m {
				m = c
			}
		}
		if int(m)-n > (math.MaxInt32-delta)/(h+1) {
			panic("punycode: 编码溢出（OVERFLOW）")
		}
		delta += (int(m) - n) * (h + 1)
		n = int(m)
		for _, c := range runes {
			if c < rune(n) {
				delta++
				if delta == 0 {
					panic("punycode: 编码溢出（OVERFLOW）")
				}
			}
			if c == rune(n) {
				q := delta
				for k := punyBase; ; k += punyBase {
					var t int
					switch {
					case k <= bias:
						t = punyTmin
					case k >= bias+punyTmax:
						t = punyTmax
					default:
						t = k - bias
					}
					if q < t {
						break
					}
					output = append(output, punyDigitToRune(t+(q-t)%(punyBase-t)))
					q = (q - t) / (punyBase - t)
				}
				output = append(output, punyDigitToRune(q))
				bias = punyAdapt(delta, h+1, h == b)
				delta = 0
				h++
			}
		}
		delta++
		n++
	}
	encoded := string(output)
	if withPrefix {
		encoded = punyPrefix + encoded
	}
	return encoded
}

// Decode Punycode 解码（对照 hutool PunyCode.decode）：
// 自动去掉可能存在的 "xn--" 前缀（忽略大小写），非法输入返回 error。
func (p *utilPunycode) Decode(input string) (string, error) {
	if len(input) >= len(punyPrefix) && strings.EqualFold(input[:len(punyPrefix)], punyPrefix) {
		input = input[len(punyPrefix):]
	}
	runes := []rune(input)
	n, i, bias := punyInitialN, 0, punyInitialBias
	output := make([]rune, 0, len(runes))
	// 找最后一个分隔符，其之前为基础码点
	d := -1
	for idx, c := range runes {
		if c == punyDelimiter {
			d = idx
		}
	}
	if d > 0 {
		for _, c := range runes[:d] {
			if c < 0x80 {
				output = append(output, c)
			}
		}
		d++
	} else {
		d = 0
	}
	length := len(runes)
	for d < length {
		oldi := i
		w := 1
		for k := punyBase; ; k += punyBase {
			if d == length {
				return "", errors.New("punycode 解码失败：输入不完整（BAD_INPUT）")
			}
			c := runes[d]
			d++
			digit, err := punyRuneToDigit(c)
			if err != nil {
				return "", err
			}
			if digit > (math.MaxInt32-i)/w {
				return "", errors.New("punycode 解码溢出（OVERFLOW）")
			}
			i += digit * w
			var t int
			switch {
			case k <= bias:
				t = punyTmin
			case k >= bias+punyTmax:
				t = punyTmax
			default:
				t = k - bias
			}
			if digit < t {
				break
			}
			w *= punyBase - t
		}
		bias = punyAdapt(i-oldi, len(output)+1, oldi == 0)
		if i/(len(output)+1) > math.MaxInt32-n {
			return "", errors.New("punycode 解码溢出（OVERFLOW）")
		}
		n += i / (len(output) + 1)
		i = i % (len(output) + 1)
		// 在位置 i 插入码点 n
		output = append(output, 0)
		copy(output[i+1:], output[i:])
		output[i] = rune(n)
		i++
	}
	return string(output), nil
}

// EncodeDomain 域名转 Punycode（对照 hutool PunyCode.encodeDomain）：
// 按 "." 拆分标签，仅对含非 ASCII 字符的标签做 xn-- 编码，纯 ASCII 标签原样保留。
// EncodeDomain("你好.com") = "xn--6qq986b.com"。
func (p *utilPunycode) EncodeDomain(domain string) string {
	labels := strings.Split(domain, ".")
	for idx, label := range labels {
		labels[idx] = p.Encode(label, true)
	}
	return strings.Join(labels, ".")
}

// DecodeDomain Punycode 域名还原（对照 hutool PunyCode.decodeDomain）：
// 按 "." 拆分标签，仅对 "xn--" 开头（忽略大小写）的标签解码，任一标签非法返回 error。
func (p *utilPunycode) DecodeDomain(domain string) (string, error) {
	labels := strings.Split(domain, ".")
	for idx, label := range labels {
		if strings.HasPrefix(strings.ToLower(label), punyPrefix) {
			decoded, err := p.Decode(label)
			if err != nil {
				return "", err
			}
			labels[idx] = decoded
		}
	}
	return strings.Join(labels, "."), nil
}

// punyAdapt bias 调整函数（RFC 3492 的 adapt）
func punyAdapt(delta, numpoints int, first bool) int {
	if first {
		delta /= punyDamp
	} else {
		delta /= 2
	}
	delta += delta / numpoints
	k := 0
	for delta > ((punyBase-punyTmin)*punyTmax)/2 {
		delta /= punyBase - punyTmin
		k += punyBase
	}
	return k + (punyBase-punyTmin+1)*delta/(delta+punySkew)
}

// punyDigitToRune 数字转编码字符：0-25 -> 'a'-'z'，26-35 -> '0'-'9'
func punyDigitToRune(d int) rune {
	if d < 26 {
		return rune(d + 'a')
	}
	if d < 36 {
		return rune(d - 26 + '0')
	}
	panic("punycode: 非法数字（BAD_DIGIT）")
}

// punyRuneToDigit 编码字符转数字（大小写不敏感），非法字符返回 error
func punyRuneToDigit(c rune) (int, error) {
	switch {
	case c >= '0' && c <= '9':
		return int(c-'0') + 26, nil
	case c >= 'a' && c <= 'z':
		return int(c - 'a'), nil
	case c >= 'A' && c <= 'Z':
		return int(c - 'A'), nil
	}
	return 0, fmt.Errorf("punycode 含非法字符 %q", c)
}

/* ---------------- Radix 自定义字母表进制转换 ---------------- */

// 34 进制乱序字母表（不含易混淆的 I、O，对照 hutool RadixUtil.RADIXS_SHUFFLE_34）
const Radix34 = "H3UM16TDFPSBZJ90CW28QYRE45AXKNGV7L"

// 59 进制乱序字母表（不含易混淆的 I、O、l，对照 hutool RadixUtil.RADIXS_SHUFFLE_59）
const Radix59 = "vh9wGkfK8YmqbsoENP3764SeCX0dVzrgy1HRtpnTaLjJW2xQiZAcBMUFDu5"

// utilRadix 自定义进制转换工具（对应 hutool RadixUtil）。
// 用途：按自定义字母表把 ID 编码成邀请码/短链码，让有规律的数字看起来无规律。
type utilRadix struct{}

// Encode 把整型数值按自定义字母表编码（对照 hutool RadixUtil.encode）：
//   - radixs 为自定义进制字符表（如 utils.Radix34），长度至少 2 且无重复字符，否则返回 error；
//   - num 不支持负数，负数返回 error。
//
// Encode(Radix34, 100) 按字母表把 100 转成 34 进制串。
// 场景：根据 ID 生成尽量短的邀请码且不希望被直接猜测出与 ID 的关联。
func (r *utilRadix) Encode(radixs string, num int64) (string, error) {
	rs := []rune(radixs)
	rl := len(rs)
	if rl < 2 {
		return "", errors.New("自定义进制最少两个字符")
	}
	if num < 0 {
		return "", errors.New("暂不支持负数")
	}
	seen := make(map[rune]bool, rl)
	for _, c := range rs {
		if seen[c] {
			return "", fmt.Errorf("自定义进制字符表含重复字符 %q", c)
		}
		seen[c] = true
	}
	// long 的 64 进制结果最长 64 位
	buf := make([]rune, 64)
	i := len(buf)
	for {
		i--
		buf[i] = rs[num%int64(rl)]
		num /= int64(rl)
		if num == 0 {
			break
		}
	}
	return string(buf[i:]), nil
}

// DecodeToInt 把 Encode 的结果还原为数值（对照 hutool RadixUtil.decodeToInt）：
// radixs 必须与编码时一致，长度不足、编码串为空或含字母表之外的字符均返回 error。
func (r *utilRadix) DecodeToInt(radixs, encodeStr string) (int64, error) {
	rl := len([]rune(radixs))
	if rl < 2 {
		return 0, errors.New("自定义进制最少两个字符")
	}
	if encodeStr == "" {
		return 0, errors.New("encodeStr 为空")
	}
	var res int64
	for _, c := range encodeStr {
		idx := strings.IndexRune(radixs, c)
		if idx < 0 {
			return 0, fmt.Errorf("编码串含非法字符 %q", c)
		}
		res = res*int64(rl) + int64(idx)
	}
	return res, nil
}

/* ---------------- Hashids（数字组 <-> 短 ID 的可逆混淆编码） ---------------- */

// Hashids 算法常量，对照 hutool Hashids（来自 davidafsilva/java-hashids）
const (
	hidsLotteryMod         = 100  // 彩票值计算的模数
	hidsGuardThreshold     = 12.0 // guard 数量 = ceil(字母表长度 / 12)
	hidsSeparatorThreshold = 3.5  // 分隔符数量阈值
	hidsMinAlphabetLength  = 16   // 字母表最小长度
)

// 默认字母表与默认分隔符（对照 hutool Hashids.DEFAULT_ALPHABET / DEFAULT_SEPARATORS）
var (
	hidsDefaultAlphabet   = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	hidsDefaultSeparators = []rune("cfhistuCFHISTU")
)

// Hashids Hashids 协议实现：把一组非负整数编码成短且唯一的字符串，并可无损解码。
// 编码结果与盐（salt）相关，可配置最小长度；常用于防止连续 ID 被爬虫遍历。
// 实例内部状态在构造时固定，方法层面用互斥锁保证并发安全（对照 hutool Hashids）。
type Hashids struct {
	mu         sync.Mutex
	alphabet   []rune        // 参与编码的字母表（构造后固定）
	separators []rune        // 多个数字编码时的分界符
	sepsSet    map[rune]bool // 分隔符集合（加速解码判断）
	salt       []rune        // 盐
	guards     []rune        // 补齐 minLength 时可能加入首尾的字符
	minLength  int           // 编码结果最小长度
}

// newHashids 按盐、最小长度与字母表创建 Hashids（供 utils.NewHashids 调用）：
//   - alphabet 为空时使用 hutool 默认字母表；
//   - minLength < 0 按 0 处理（不限制最小长度）；
//   - 自定义字母表非法（去重后少于 16 个字符或含空格）时 panic
//     （对照 hutool 构造函数抛出 IllegalArgumentException）。
func newHashids(salt string, minLength int, alphabet string) *Hashids {
	if minLength < 0 {
		minLength = 0
	}
	saltRunes := []rune(salt)
	if alphabet == "" {
		alphabet = string(hidsDefaultAlphabet)
	}
	alphabetRunes := []rune(alphabet)

	// 过滤并洗牌分隔符：仅保留出现在字母表中的分隔符
	tmpSeparators := hidsShuffle(hidsFilterSeparators(hidsDefaultSeparators, alphabetRunes), saltRunes)

	// 校验并过滤字母表（去重、剔除分隔符字符）
	tmpAlphabet, err := hidsValidateAlphabet(alphabetRunes, tmpSeparators)
	if err != nil {
		panic(err.Error())
	}

	// 分隔符数量不足时从字母表头部补充（对照 hutool，注意这里的整数除法语义）
	if len(tmpSeparators) == 0 || float64(len(tmpAlphabet)/len(tmpSeparators)) > hidsSeparatorThreshold {
		minSeparatorsSize := int(math.Ceil(float64(len(tmpAlphabet)) / hidsSeparatorThreshold))
		if minSeparatorsSize > len(tmpSeparators) {
			missing := minSeparatorsSize - len(tmpSeparators)
			head := append([]rune(nil), tmpAlphabet[:missing]...)
			tmpSeparators = append(tmpSeparators, head...)
			tmpAlphabet = tmpAlphabet[missing:]
		}
	}

	// 洗牌最终字母表
	hidsShuffle(tmpAlphabet, saltRunes)

	guardsLen := int(math.Ceil(float64(len(tmpAlphabet)) / hidsGuardThreshold))
	h := &Hashids{
		minLength: minLength,
		salt:      append([]rune(nil), saltRunes...),
		sepsSet:   make(map[rune]bool, len(tmpSeparators)),
	}
	if len(alphabetRunes) < 3 { // 对照 hutool：极小字母表时 guard 取自分隔符（常规入参不会走到）
		h.guards = append([]rune(nil), tmpSeparators[:guardsLen]...)
		h.separators = append([]rune(nil), tmpSeparators[guardsLen:]...)
		h.alphabet = append([]rune(nil), tmpAlphabet...)
	} else {
		h.guards = append([]rune(nil), tmpAlphabet[:guardsLen]...)
		h.separators = append([]rune(nil), tmpSeparators...)
		h.alphabet = append([]rune(nil), tmpAlphabet[guardsLen:]...)
	}
	for _, c := range h.separators {
		h.sepsSet[c] = true
	}
	return h
}

// Encode 把一组非负整数编码为 Hashids 字符串（对照 hutool Hashids.encode）：
// 空参数或含负数返回 error。Encode(347) 可能得到 "yr8" 之类的短 ID。
// 场景：把数据库自增主键混淆后对外暴露。
func (h *Hashids) Encode(numbers ...int64) (string, error) {
	if len(numbers) == 0 {
		return "", errors.New("hashids 编码至少需要一个数字")
	}
	for _, num := range numbers {
		if num < 0 {
			return "", fmt.Errorf("hashids 不支持负数: %d", num)
		}
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.encodeImpl(numbers), nil
}

// Decode 把 Hashids 字符串解码回数字组，非法 hash（含字母表之外的字符或校验失败）返回 error。
func (h *Hashids) Decode(hash string) ([]int64, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.decodeImpl([]rune(hash))
}

// EncodeHex 把 16 进制串按 16 位大整数分组编码（对照 hutool Hashids.encodeFromHex）：
// 每 12 个 hex 字符一组（可带 0x/0X 前缀），非法 hex 返回 error。
// 场景：对 16 进制形式的 ID（如部分 Mongo ObjectId 的十六进制串）做混淆。
func (h *Hashids) EncodeHex(hexStr string) (string, error) {
	trimmed := strings.TrimPrefix(hexStr, "0x")
	trimmed = strings.TrimPrefix(trimmed, "0X")
	if trimmed == "" {
		return "", errors.New("hashids 16 进制串不能为空")
	}
	numbers := make([]int64, 0, len(trimmed)/12+1)
	for i := 0; i < len(trimmed); i += 12 {
		end := i + 12
		if end > len(trimmed) {
			end = len(trimmed)
		}
		// 前补 '1' 保证值为正，取低 64 位（对照 hutool BigInteger("1"+chunk,16).longValue()）
		v, err := strconv.ParseUint("1"+trimmed[i:end], 16, 64)
		if err != nil {
			return "", fmt.Errorf("hashids 非法 16 进制串 %q: %w", hexStr, err)
		}
		numbers = append(numbers, int64(v))
	}
	return h.Encode(numbers...)
}

// DecodeToHex 把 Hashids 字符串解码为 16 进制串（对照 hutool Hashids.decodeToHex），
// 与 EncodeHex 互逆，非法 hash 返回 error。
func (h *Hashids) DecodeToHex(hash string) (string, error) {
	values, err := h.Decode(hash)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, v := range values {
		// 编码时前补了 '1'，此处去掉 hex 首位即可还原（对照 hutool 的 hex(1, length) 截取）
		hexStr := strconv.FormatUint(uint64(v), 16)
		b.WriteString(hexStr[1:])
	}
	return b.String(), nil
}

// encodeImpl 编码核心（无锁，Encode 与 Decode 校验共用；空数字组按 hutool 语义返回补齐串）
func (h *Hashids) encodeImpl(numbers []int64) string {
	current := append([]rune(nil), h.alphabet...)

	// 彩票值：各数字对 (下标+100) 取模后累加
	var lotteryID int64
	for i, num := range numbers {
		lotteryID += num % int64(i+hidsLotteryMod)
	}
	lottery := current[int(lotteryID)%len(current)]

	global := make([]rune, 0, len(numbers)*2+8)
	for idx, num := range numbers {
		// 每编码一个数字前，按 [lottery+salt+alphabet] 派生新字母表
		hidsDeriveNewAlphabet(current, h.salt, lottery)

		initialLength := len(global)
		global = hidsAppendNumber(global, num, current)

		// 第一个数字编码完成后把彩票字符插到最前面
		if idx == 0 {
			global = hidsInsertAt(global, 0, []rune{lottery})
		}

		// 还有剩余数字时追加分隔符（对照 hutool 的 charAt(initialLength)+1 取模）
		if idx+1 < len(numbers) {
			n := num % (int64(global[initialLength]) + 1)
			global = append(global, h.separators[int(n)%len(h.separators)])
		}
	}

	// 长度不足时在首尾插入 guard 字符
	if h.minLength > len(global) && len(h.guards) > 0 {
		guardIdx := int((lotteryID + int64(lottery)) % int64(len(h.guards)))
		global = hidsInsertAt(global, 0, []rune{h.guards[guardIdx]})
		if h.minLength > len(global) {
			at := 2
			if at >= len(global) { // 防越界（hutool 此处在极端入参下会下标越界）
				at = len(global) - 1
			}
			guardIdx = int((lotteryID + int64(global[at])) % int64(len(h.guards)))
			global = append(global, h.guards[guardIdx])
		}
	}

	// 仍不足 minLength 时用洗牌后的字母表向两侧填充
	paddingLeft := h.minLength - len(global)
	for paddingLeft > 0 {
		salt := append([]rune(nil), current...)
		hidsShuffle(current, salt)

		half := len(current) / 2
		initialSize := len(global)
		if paddingLeft > len(current) {
			// 整个字母表包住当前编码
			offset := half
			if len(current)%2 != 0 {
				offset++ // 奇数长度时后半段多取中间一个字符
			}
			global = hidsInsertAt(global, 0, current[half:half+offset])
			global = hidsInsertAt(global, offset+initialSize, current[:half])
			paddingLeft -= len(current)
		} else {
			// 精确补齐：前半段砍掉 excess/2，剩余部分分别贴到两侧
			excess := len(current) + len(global) - h.minLength
			secondStart := half + hidsFloorDiv(excess, 2)
			secondLen := len(current) - secondStart
			firstLen := paddingLeft - secondLen
			global = hidsInsertAt(global, 0, current[secondStart:])
			global = hidsInsertAt(global, secondLen+initialSize, current[:firstLen])
			paddingLeft = 0
		}
	}
	return string(global)
}

// decodeImpl 解码核心（无锁，调用方持锁）
func (h *Hashids) decodeImpl(hash []rune) ([]int64, error) {
	// 定位 hash 中出现的 guard 位置，取前两个作为编码区间边界
	var guardsIdx []int
	for i, c := range hash {
		for _, g := range h.guards {
			if g == c {
				guardsIdx = append(guardsIdx, i)
				break
			}
		}
	}
	startIdx, endIdx := 0, len(hash)
	if len(guardsIdx) > 0 {
		startIdx = guardsIdx[0] + 1
		if len(guardsIdx) > 1 {
			endIdx = guardsIdx[1]
		}
	}

	var decoded []int64
	if len(hash) > 0 && startIdx < endIdx {
		lottery := hash[startIdx]

		block := make([]rune, 0, len(hash))
		// 基础解码盐：lottery + salt（不足字母表长度时由当前字母表补齐）
		decodeSalt := make([]rune, len(h.alphabet))
		decodeSalt[0] = lottery
		saltLength := len(h.salt)
		if saltLength >= len(h.alphabet) {
			saltLength = len(h.alphabet) - 1
		}
		copy(decodeSalt[1:1+saltLength], h.salt[:saltLength])
		saltLeft := len(h.alphabet) - saltLength - 1

		current := append([]rune(nil), h.alphabet...)

		// 按分隔符切块，每块用各自洗牌后的字母表解码
		for i := startIdx + 1; i < endIdx; i++ {
			if !h.sepsSet[hash[i]] {
				block = append(block, hash[i])
				if i < endIdx-1 { // 未到区间末尾则继续积攒
					continue
				}
			}
			if len(block) > 0 {
				if saltLeft > 0 {
					copy(decodeSalt[len(h.alphabet)-saltLeft:], current[:saltLeft])
				}
				hidsShuffle(current, decodeSalt)
				n, err := hidsParseBlock(block, current)
				if err != nil {
					return nil, err
				}
				decoded = append(decoded, n)
				block = block[:0]
			}
		}
	}

	// 校验：重新编码必须与原 hash 完全一致（对照 hutool 的 Objects.equals 校验）
	if string(hash) != h.encodeImpl(decoded) {
		return nil, fmt.Errorf("hashids 非法 hash: %s", string(hash))
	}
	return decoded, nil
}

// hidsAppendNumber 把数字按字母表编码追加到 sb（高位在前，对照 hutool translate）
func hidsAppendNumber(sb []rune, n int64, alphabet []rune) []rune {
	base := int64(len(alphabet))
	var digits []rune
	for {
		digits = append(digits, alphabet[n%base])
		n /= base
		if n == 0 {
			break
		}
	}
	for i := len(digits) - 1; i >= 0; i-- {
		sb = append(sb, digits[i])
	}
	return sb
}

// hidsParseBlock 把一段编码按字母表还原为数字（对照 hutool translate 反向），
// 含字母表之外的字符返回 error
func hidsParseBlock(block, alphabet []rune) (int64, error) {
	mapping := make(map[rune]int64, len(alphabet))
	for i, c := range alphabet {
		if _, ok := mapping[c]; !ok { // 重复字符取首个下标（对照 hutool 的 reducing 语义）
			mapping[c] = int64(i)
		}
	}
	var number int64
	mult := int64(1)
	for i := len(block) - 1; i >= 0; i-- {
		idx, ok := mapping[block[i]]
		if !ok {
			return 0, fmt.Errorf("hashids 含字母表之外的字符 %q", string(block[i]))
		}
		number += idx * mult
		mult *= int64(len(alphabet))
	}
	return number, nil
}

// hidsDeriveNewAlphabet 就地派生新字母表：以 [lottery+salt+alphabet] 为盐洗牌（对照 hutool）
func hidsDeriveNewAlphabet(alphabet, salt []rune, lottery rune) {
	newSalt := make([]rune, len(alphabet))
	newSalt[0] = lottery
	spaceLeft := len(newSalt) - 1
	offset := 1
	if len(salt) > 0 && spaceLeft > 0 {
		n := len(salt)
		if n > spaceLeft {
			n = spaceLeft
		}
		copy(newSalt[offset:offset+n], salt[:n])
		spaceLeft -= n
		offset += n
	}
	if spaceLeft > 0 {
		copy(newSalt[offset:offset+spaceLeft], alphabet[:spaceLeft])
	}
	hidsShuffle(alphabet, newSalt)
}

// hidsFilterSeparators 仅保留出现在字母表中的分隔符（对照 hutool filterSeparators）
func hidsFilterSeparators(separators, alphabet []rune) []rune {
	valid := make(map[rune]bool, len(alphabet))
	for _, c := range alphabet {
		valid[c] = true
	}
	out := make([]rune, 0, len(separators))
	for _, c := range separators {
		if valid[c] {
			out = append(out, c)
		}
	}
	return out
}

// hidsValidateAlphabet 校验字母表：去重后长度不得小于 16、不能含空格，
// 返回剔除分隔符字符后的去重字母表（保持原有顺序，对照 hutool validateAndFilterAlphabet）
func hidsValidateAlphabet(alphabet, separators []rune) ([]rune, error) {
	if len(alphabet) < hidsMinAlphabetLength {
		return nil, fmt.Errorf("hashids 字母表至少需要 %d 个字符，当前 %d 个", hidsMinAlphabetLength, len(alphabet))
	}
	invalid := make(map[rune]bool, len(separators))
	for _, c := range separators {
		invalid[c] = true
	}
	seen := make(map[rune]bool, len(alphabet))
	unique := make([]rune, 0, len(alphabet))
	for i, c := range alphabet {
		if c == ' ' {
			return nil, fmt.Errorf("hashids 字母表不能包含空格（位置 %d）", i)
		}
		if !invalid[c] && !seen[c] {
			seen[c] = true
			unique = append(unique, c)
		}
	}
	return unique, nil
}

// hidsShuffle hutool 特有的洗牌算法（就地交换，盐为空时不洗牌，对照 hutool shuffle）
func hidsShuffle(alphabet, salt []rune) []rune {
	if len(salt) == 0 {
		return alphabet
	}
	for i, v, p := len(alphabet)-1, 0, 0; i > 0; i, v = i-1, v+1 {
		v %= len(salt)
		z := int(salt[v])
		p += z
		j := (z + v + p) % i
		alphabet[j], alphabet[i] = alphabet[i], alphabet[j]
	}
	return alphabet
}

// hidsInsertAt 在 pos 处插入 ins 并返回新切片
func hidsInsertAt(dst []rune, pos int, ins []rune) []rune {
	out := make([]rune, 0, len(dst)+len(ins))
	out = append(out, dst[:pos]...)
	out = append(out, ins...)
	out = append(out, dst[pos:]...)
	return out
}

// hidsFloorDiv 向下取整除法（对应 Java Math.floorDiv）
func hidsFloorDiv(a, b int) int {
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
}
