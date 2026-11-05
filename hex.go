package utils

import (
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// utilHex 十六进制工具。
//
// 覆盖字节/字符串与 hex 互转、数值与 hex 互转、IEEE754 位模式还原、
// hex 分组格式化与颜色值编解码，方法逐个对照 hutool HexUtil 移植。
type utilHex struct{}

/* ---------------- 编码与解码 ---------------- */

// Encode 字节数组转小写十六进制字符串。Encode([]byte{0x01, 0xAB}) = "01ab"。
// 场景：报文、摘要等字节流以文本形式展示或传输。
func (x *utilHex) Encode(data []byte) string {
	return hex.EncodeToString(data)
}

// EncodeUpper 字节数组转大写十六进制字符串。EncodeUpper([]byte{0x01, 0xAB}) = "01AB"。
// 场景：协议要求大写 hex 的场合，如某些硬件协议、单据号。
func (x *utilHex) EncodeUpper(data []byte) string {
	return strings.ToUpper(hex.EncodeToString(data))
}

// Decode 十六进制字符串解码为字节数组；长度为奇数或含非 hex 字符时返回错误
// （注意：hutool 对奇数长度会补 0 凑偶，此处按严格校验返回错误）。
// Decode("e8af95") = [0xe8, 0xaf, 0x95]。场景：还原 Encode 编码的字节流。
func (x *utilHex) Decode(hexStr string) ([]byte, error) {
	return hex.DecodeString(hexStr)
}

// EncodeStr 字符串按 UTF-8 字节转小写十六进制。EncodeStr("中") = "e4b8ad"。
// 场景：把文本转成 hex 串存储或传输。
func (x *utilHex) EncodeStr(str string) string {
	return hex.EncodeToString([]byte(str))
}

// DecodeStr 十六进制字符串按 UTF-8 还原为字符串，输入非法时返回错误。
// DecodeStr("e4b8ad") = "中"。场景：还原 EncodeStr 编码的内容。
func (x *utilHex) DecodeStr(hexStr string) (string, error) {
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// IsHexNumber 判断字符串是否为合法的十六进制数：支持 0x/0X 与 # 前缀，也可无前缀；
// 空串、负号开头或含非 hex 字符返回 false（对应 hutool isHexNumber）。
// IsHexNumber("0xFF") = true，IsHexNumber("0xG1") = false，IsHexNumber("-11") = false。
// 场景：解析配置或表达式前判断数值形态。
func (x *utilHex) IsHexNumber(value string) bool {
	if value == "" || strings.HasPrefix(value, "-") {
		return false
	}
	idx := 0
	if strings.HasPrefix(value, "0x") || strings.HasPrefix(value, "0X") {
		idx = 2
	} else if strings.HasPrefix(value, "#") {
		idx = 1
	}
	body := value[idx:]
	if body == "" {
		return false
	}
	for i := 0; i < len(body); i++ {
		c := body[i]
		isHex := c >= '0' && c <= '9' ||
			c >= 'a' && c <= 'f' ||
			c >= 'A' && c <= 'F'
		if !isHex {
			return false
		}
	}
	return true
}

/* ---------------- 数值转换 ---------------- */

// ToUnicodeHex 字符转 4 位十六进制形式（不足 4 位前面补 0），如 '中' -> "4e2d"。
// 注意与 hutool toUnicodeHex 的差别：此处不含 "\u" 前缀，需要时自行拼接（"\\u" + ToUnicodeHex(r)）。
// 场景：生成 Unicode 转义串，如中文转 \u4e2d 形式的 properties 文件内容。
func (x *utilHex) ToUnicodeHex(r rune) string {
	return fmt.Sprintf("%04x", uint32(r))
}

// ToHex 整数转十六进制字符串（无符号输出，负数按补码形式，对应 Java Long.toHexString）。
// ToHex(255) = "ff"，ToHex(-1) = "ffffffffffffffff"。
// 场景：把数值以 hex 形式写入协议报文或日志。
func (x *utilHex) ToHex(num int64) string {
	return strconv.FormatUint(uint64(num), 16)
}

// HexToInt 十六进制字符串转 int64，自动去除 0x/0X/# 前缀，非法输入返回错误（对应 hutool hexToInt）。
// HexToInt("0xff") = 255。场景：解析 hex 字面量配置、报文中的数值字段。
func (x *utilHex) HexToInt(hexStr string) (int64, error) {
	return strconv.ParseInt(hexRemovePrefix(hexStr), 16, 64)
}

// HexToFloat32 把 8 位十六进制串按 IEEE754 单精度位模式还原为 float32（对应 hutool hexToFloat）。
// HexToFloat32("3f800000") = 1。场景：还原 ToHex 转储的浮点位模式、解析二进制协议 dump。
func (x *utilHex) HexToFloat32(hexStr string) (float32, error) {
	u, err := strconv.ParseUint(hexRemovePrefix(hexStr), 16, 32)
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(uint32(u)), nil
}

// HexToFloat64 把 16 位十六进制串按 IEEE754 双精度位模式还原为 float64（对应 hutool hexToDouble）。
// HexToFloat64("3ff0000000000000") = 1。场景：同 HexToFloat32，用于双精度。
func (x *utilHex) HexToFloat64(hexStr string) (float64, error) {
	u, err := strconv.ParseUint(hexRemovePrefix(hexStr), 16, 64)
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(u), nil
}

/* ---------------- 格式化与颜色 ---------------- */

// Format 格式化 hex 字符串：每 2 位一组用空格分隔，长度为奇数时前面补 0（对应 hutool format）。
// Format("010203") = "01 02 03"，Format("123") = "01 23"。
// 场景：日志中打印字节流，便于逐字节核对。
func (x *utilHex) Format(hexStr string) string {
	if hexStr == "" {
		return ""
	}
	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}
	var b strings.Builder
	b.Grow(len(hexStr) + len(hexStr)/2 - 1)
	for i := 0; i < len(hexStr); i += 2 {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(hexStr[i : i+2])
	}
	return b.String()
}

// EncodeColor 把 RGB 三个分量编码为 "#RRGGBB" 大写十六进制颜色串（对应 hutool encodeColor）。
// EncodeColor(255, 255, 255) = "#FFFFFF"。场景：生成前端颜色值。
func (x *utilHex) EncodeColor(r, g, b uint8) string {
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

// DecodeColor 解析十六进制颜色串为 r/g/b 三个分量，支持 "#FFFFFF" 与 "FFFFFF" 两种形式
// （对应 hutool decodeColor）；长度不为 6 或含非 hex 字符时返回错误。
// DecodeColor("#FF8000") = (255, 128, 0, nil)。场景：解析配置中的颜色值。
func (x *utilHex) DecodeColor(colorHex string) (r, g, b uint8, err error) {
	s := strings.TrimPrefix(colorHex, "#")
	if len(s) != 6 {
		return 0, 0, 0, fmt.Errorf("颜色值 %q 格式非法：需要 6 位十六进制（可带 # 前缀）", colorHex)
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("颜色值 %q 含非十六进制字符：%w", colorHex, err)
	}
	return uint8(v >> 16), uint8(v >> 8), uint8(v), nil
}

// hexRemovePrefix 去除十六进制字符串的常见前缀 0x、0X、#（对应 hutool removeHexPrefix）
func hexRemovePrefix(hexStr string) string {
	if len(hexStr) > 1 {
		switch hexStr[0] {
		case '0':
			if hexStr[1] == 'x' || hexStr[1] == 'X' {
				return hexStr[2:]
			}
		case '#':
			return hexStr[1:]
		}
	}
	return hexStr
}
