package utils

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"sort"
	"strings"
	"time"
)

// utilCrypto 哈希与编码工具。
//
// 注意：MD5/SHA1 属于弱哈希，仅用于数据校验、缓存键、兼容旧系统；
// 涉及密码存储请使用 bcrypt/argon2 等专用算法，接口签名推荐 HmacSha256。
type utilCrypto struct{}

// MD5 计算 32 位小写十六进制 MD5 摘要。
// MD5("abc") = "900150983cd24fb0d6963f7d28e17f72"。
// 场景：文件/报文完整性校验、缓存键生成。
func (c *utilCrypto) MD5(input string) string {
	sum := md5.Sum([]byte(input))
	return hex.EncodeToString(sum[:])
}

// MD5WithSalt 将 input 与 salt 拼接后再取 MD5。
// 场景：兼容旧系统的口令摘要存储（新系统建议改用 bcrypt）。
func (c *utilCrypto) MD5WithSalt(input, salt string) string {
	return c.MD5(input + salt)
}

// MD5File 流式计算文件 MD5，大文件也不会占用过多内存。
// 场景：上传文件秒传判断、下载文件完整性校验。
func (c *utilCrypto) MD5File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// SHA1 计算 40 位小写十六进制 SHA1 摘要。场景：Git 风格内容标识、旧系统兼容。
func (c *utilCrypto) SHA1(input string) string {
	sum := sha1.Sum([]byte(input))
	return hex.EncodeToString(sum[:])
}

// SHA256 计算 64 位小写十六进制 SHA256 摘要。场景：数据指纹、区块链式校验。
func (c *utilCrypto) SHA256(input string) string {
	sum := sha256.Sum256([]byte(input))
	return hex.EncodeToString(sum[:])
}

// HmacSha256 计算 HMAC-SHA256 并返回小写十六进制串。
// 场景：开放接口签名，如 hex(HMAC-SHA256(secret, "GET\n/path\n" + timestamp))。
func (c *utilCrypto) HmacSha256(key, data string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

// Base64Encode 标准 Base64 编码（带填充）。
// Base64Encode("hello") = "aGVsbG8="。场景：HTTP Basic、简单二进制转文本。
func (c *utilCrypto) Base64Encode(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

// Base64Decode 标准 Base64 解码，输入非法时返回错误。
func (c *utilCrypto) Base64Decode(input string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Base64UrlEncode URL 安全的 Base64 编码（用 - _ 替代 + /，带填充）。
// 场景：JWT 载荷、放入 URL/Cookie/文件名的编码。
func (c *utilCrypto) Base64UrlEncode(input string) string {
	return base64.URLEncoding.EncodeToString([]byte(input))
}

// Base64UrlDecode URL 安全的 Base64 解码，输入非法时返回错误。
func (c *utilCrypto) Base64UrlDecode(input string) (string, error) {
	data, err := base64.URLEncoding.DecodeString(input)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Uuid 生成 UUID v4 字符串（36 位，含 4 个连字符），基于 crypto/rand 随机源。
// 随机源异常时返回空串。场景：请求 ID、消息去重键、分布式场景的临时 ID。
func (c *utilCrypto) Uuid() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return ""
	}
	b[6] = (b[6] & 0x0f) | 0x40 // 高 4 位置为版本号 4
	b[8] = (b[8] & 0x3f) | 0x80 // 高 2 位置为变体 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// ------------------ 摘要扩展（[]byte 与 SHA512 变体） ------------------

// MD5Bytes 计算 []byte 数据的 32 位小写十六进制 MD5 摘要。
// 场景：对二进制内容（图片、报文字节流）求校验值，避免先转字符串。
func (c *utilCrypto) MD5Bytes(data []byte) string {
	sum := md5.Sum(data)
	return hex.EncodeToString(sum[:])
}

// SHA1Bytes 计算 []byte 数据的 40 位小写十六进制 SHA1 摘要。场景：二进制内容指纹、旧系统兼容。
func (c *utilCrypto) SHA1Bytes(data []byte) string {
	sum := sha1.Sum(data)
	return hex.EncodeToString(sum[:])
}

// SHA256Bytes 计算 []byte 数据的 64 位小写十六进制 SHA256 摘要。场景：二进制内容指纹、下载校验。
func (c *utilCrypto) SHA256Bytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// Sha512 计算 []byte 数据的 128 位小写十六进制 SHA512 摘要。场景：大文件指纹、证书摘要。
func (c *utilCrypto) Sha512(data []byte) string {
	sum := sha512.Sum512(data)
	return hex.EncodeToString(sum[:])
}

// Sha512Bytes 同 Sha512，命名对齐 MD5Bytes/SHA1Bytes/SHA256Bytes 系列。场景：同上。
func (c *utilCrypto) Sha512Bytes(data []byte) string {
	return c.Sha512(data)
}

// Sha512Str 计算字符串的 SHA512 摘要（hex 小写）。场景：口令摘要兼容、长文本指纹。
func (c *utilCrypto) Sha512Str(str string) string {
	return c.Sha512([]byte(str))
}

// ------------------ HMAC 扩展 ------------------

// HmacMd5 计算 HMAC-MD5 并返回小写十六进制串。场景：兼容旧系统的消息认证码（建议新系统用 HmacSha256）。
func (c *utilCrypto) HmacMd5(key, data string) string {
	mac := hmac.New(md5.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

// HmacSha1 计算 HMAC-SHA1 并返回小写十六进制串。场景：OAuth 1.0 签名、旧接口认证。
func (c *utilCrypto) HmacSha1(key, data string) string {
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

// HmacSha512 计算 HMAC-SHA512 并返回小写十六进制串。场景：高安全要求的消息认证、JWT HS512。
func (c *utilCrypto) HmacSha512(key, data string) string {
	mac := hmac.New(sha512.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

// ------------------ 参数签名（对应 hutool SignUtil.signParams 系列） ------------------

// SignParamsMd5 对参数做 MD5 签名：key 升序排序后依次拼接 k+v（无分隔符），
// 再依次拼接 otherParams 各项，最后取 MD5（hex 小写）。
// 对应 hutool SignUtil.signParamsMd5（默认 separator 与 keyValueSeparator 均为空串）。
// 场景：开放平台 API 签名，如 md5(sorted(params) + "&secret=xxx")。
func (c *utilCrypto) SignParamsMd5(params map[string]string, otherParams ...string) string {
	return c.MD5Bytes([]byte(cryptoSignParamsPlain(params, otherParams)))
}

// SignParamsSha1 对参数做 SHA1 签名，拼接规则同 SignParamsMd5（对应 hutool SignUtil.signParamsSha1）。
func (c *utilCrypto) SignParamsSha1(params map[string]string, otherParams ...string) string {
	return c.SHA1Bytes([]byte(cryptoSignParamsPlain(params, otherParams)))
}

// SignParamsSha256 对参数做 SHA256 签名，拼接规则同 SignParamsMd5（对应 hutool SignUtil.signParamsSha256）。
// 场景：支付回调等要求强摘要的参数签名。
func (c *utilCrypto) SignParamsSha256(params map[string]string, otherParams ...string) string {
	return c.SHA256Bytes([]byte(cryptoSignParamsPlain(params, otherParams)))
}

// cryptoSignParamsPlain 按 hutool MapUtil.sortJoin 的默认规则拼接参数：
// key 升序排序后依次拼接 k+v，无任何分隔符，最后把 otherParams 各项原样追加到末尾。
// （Go 的 map 值无 null 概念，hutool 的忽略 null 逻辑不适用。）
func cryptoSignParamsPlain(params map[string]string, otherParams []string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString(params[k])
	}
	for _, v := range otherParams {
		sb.WriteString(v)
	}
	return sb.String()
}

// ------------------ OTP 一次性密码（RFC 4226 / RFC 6238，对应 hutool HOTP/TOTP） ------------------

// cryptoOtpModDivisors 口令位数对应的取模除数表，下标即位数（对照 hutool HOTP.MOD_DIVISORS）。
var cryptoOtpModDivisors = [9]uint32{1, 10, 100, 1000, 10000, 100000, 1000000, 10000000, 100000000}

// Hotp 生成 RFC 4226 基于事件计数的一次性密码（HOTP，HMAC-SHA1，动态截断算法对照 hutool HOTP）。
// secretB32 为 Base32 编码密钥（容忍小写与缺失填充）；digits 为位数，6~8 之外的值按 6 处理。
// 场景：计数型硬件令牌、双因素认证的事件同步模式。
func (c *utilCrypto) Hotp(secretB32 string, counter int64, digits int) (string, error) {
	key, err := cryptoDecodeBase32(secretB32)
	if err != nil {
		return "", err
	}
	return cryptoHotpCode(key, uint64(counter), digits), nil
}

// Totp 生成 RFC 6238 基于时间的一次性密码（TOTP），时间步长 30 秒（Google Authenticator 兼容）。
// digits 含义同 Hotp。场景：登录二次验证、敏感操作动态口令。
func (c *utilCrypto) Totp(secretB32 string, t time.Time, digits int) (string, error) {
	return c.Hotp(secretB32, t.Unix()/30, digits)
}

// TotpValidate 校验一次性密码，允许与当前时间相差前后 window 个步长（应对时钟漂移，建议 1）。
// code 非法密钥返回 error；位数取 code 的实际长度（6~8 之外按 6 处理），不匹配时返回 false。
// 场景：服务端校验用户输入的 6 位动态口令。
func (c *utilCrypto) TotpValidate(secretB32, code string, window int) (bool, error) {
	key, err := cryptoDecodeBase32(secretB32)
	if err != nil {
		return false, err
	}
	now := time.Now().Unix() / 30
	for i := -window; i <= window; i++ {
		counter := now + int64(i)
		if counter < 0 {
			continue
		}
		if cryptoHotpCode(key, uint64(counter), len(code)) == code {
			return true, nil
		}
	}
	return false, nil
}

// GoogleSecretKey 生成 Google Authenticator 兼容的随机 Base32 密钥与 otpauth 扫码链接。
// numBytes 为密钥随机字节数（hutool 语义），<=0 时按常用的 20 字节处理。
// 返回 uri 形如 otpauth://totp/{account}?secret={secret}。场景：TOTP 绑定页生成二维码内容。
func (c *utilCrypto) GoogleSecretKey(account string, numBytes int) (secret, uri string, err error) {
	if numBytes <= 0 {
		numBytes = 20
	}
	buf := make([]byte, numBytes)
	if _, err = rand.Read(buf); err != nil {
		return "", "", err
	}
	secret = base32.StdEncoding.EncodeToString(buf)
	uri = fmt.Sprintf("otpauth://totp/%s?secret=%s", account, secret)
	return secret, uri, nil
}

// cryptoHotpCode 按 RFC 4226 动态截断算法由 HMAC-SHA1 结果生成 digits 位数字口令（hutool HOTP.truncate）。
func cryptoHotpCode(key []byte, counter uint64, digits int) string {
	if digits < 6 || digits > 8 {
		digits = 6
	}
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], counter)
	mac := hmac.New(sha1.New, key)
	mac.Write(buf[:])
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	code := (uint32(sum[offset]&0x7f)<<24 |
		uint32(sum[offset+1])<<16 |
		uint32(sum[offset+2])<<8 |
		uint32(sum[offset+3])) % cryptoOtpModDivisors[digits]
	return fmt.Sprintf("%0*d", digits, code)
}

// cryptoDecodeBase32 解码 Base32 密钥，容忍小写字母与缺失的 '=' 填充
// （Google Authenticator 生成的密钥常见此两种形态），解码前统一转大写并去掉填充符。
func cryptoDecodeBase32(s string) ([]byte, error) {
	s = strings.ToUpper(strings.TrimSpace(s))
	s = strings.TrimRight(s, "=")
	return base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(s)
}

// ------------------ PBKDF2 密钥派生（对应 hutool PBKDF2） ------------------

// Pbkdf2Hex 计算 PBKDF2 并返回小写 hex，固定使用 hutool PBKDF2 默认参数：
// HmacSHA1、迭代 1000 次、派生密钥 512 位（64 字节）。结果与 hutool new PBKDF2().encryptHex 互通。
// 场景：旧口令库迁移、与 Java 服务端互相校验派生密钥。
func (c *utilCrypto) Pbkdf2Hex(password, salt []byte) string {
	key, err := pbkdf2.Key(sha1.New, string(password), salt, 1000, 64)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(key)
}

// Pbkdf2Key PBKDF2 通用版本：迭代次数、派生长度与哈希算法均可指定，薄封装标准库 crypto/pbkdf2。
// 示例：Pbkdf2Key("password", salt, 4096, 32, sha256.New)。
// 场景：自定义安全参数的密钥派生（如 AES 密钥由口令派生）。
func (c *utilCrypto) Pbkdf2Key(password string, salt []byte, iterations, keyLen int, newHash func() hash.Hash) ([]byte, error) {
	return pbkdf2.Key(newHash, password, salt, iterations, keyLen)
}
