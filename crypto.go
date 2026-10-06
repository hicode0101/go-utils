package utils

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
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
