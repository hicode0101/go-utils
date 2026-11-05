package utils

import (
	"crypto/sha1"
	"encoding/base32"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

/* ---------------- 摘要：MD5 家族 ---------------- */

func TestCryptoMd5AndFile(t *testing.T) {
	// RFC 1321 黄金值
	if got := Crypto.MD5("abc"); got != "900150983cd24fb0d6963f7d28e17f72" {
		t.Errorf("MD5(abc) = %q, want \"900150983cd24fb0d6963f7d28e17f72\"", got)
	}
	// MD5WithSalt 等价于拼接后摘要，且与无盐摘要不同
	if got := Crypto.MD5WithSalt("abc", "salt"); got != Crypto.MD5("abcsalt") {
		t.Errorf("MD5WithSalt(abc, salt) = %q, want md5(abcsalt) = %q", got, Crypto.MD5("abcsalt"))
	}
	if Crypto.MD5WithSalt("abc", "salt") == Crypto.MD5("abc") {
		t.Error("加盐摘要不应等于无盐摘要")
	}

	// MD5File：写临时文件后求摘要，与同内容字符串摘要一致
	path := filepath.Join(t.TempDir(), "data.txt")
	content := "hello 文件摘要 123"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("写临时文件失败: %v", err)
	}
	got, err := Crypto.MD5File(path)
	if err != nil {
		t.Fatalf("MD5File 失败: %v", err)
	}
	if want := Crypto.MD5(content); got != want {
		t.Errorf("MD5File = %q, want %q", got, want)
	}
	// 错误路径：文件不存在
	if _, err := Crypto.MD5File(filepath.Join(t.TempDir(), "no_such_file.txt")); err == nil {
		t.Error("MD5File 打开不存在的文件应返回错误")
	}
}

/* ---------------- 摘要：SHA 家族与 []byte/SHA512 变体 ---------------- */

func TestCryptoShaDigests(t *testing.T) {
	// NIST 标准向量（"abc"）
	if got := Crypto.SHA1("abc"); got != "a9993e364706816aba3e25717850c26c9cd0d89d" {
		t.Errorf("SHA1(abc) = %q", got)
	}
	if got := Crypto.SHA256("abc"); got != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Errorf("SHA256(abc) = %q", got)
	}

	// []byte 版本与 string 版本一致（用含非 UTF-8 字节的二进制数据验证）
	bin := []byte{0x00, 0x01, 0xfb, 0xff, 0xfe, 'a', 'b', 0x00}
	s := string(bin)
	if got, want := Crypto.MD5Bytes(bin), Crypto.MD5(s); got != want {
		t.Errorf("MD5Bytes 与 MD5 不一致: %q vs %q", got, want)
	}
	if got, want := Crypto.SHA1Bytes(bin), Crypto.SHA1(s); got != want {
		t.Errorf("SHA1Bytes 与 SHA1 不一致: %q vs %q", got, want)
	}
	if got, want := Crypto.SHA256Bytes(bin), Crypto.SHA256(s); got != want {
		t.Errorf("SHA256Bytes 与 SHA256 不一致: %q vs %q", got, want)
	}

	// SHA512 黄金值：128 位小写 hex
	const sha512Abc = "ddaf35a193617abacc417349ae20413112e6fa4e89a97ea20a9eeee64b55d39a" +
		"2192992a274fc1a836ba3c23a3feebbd454d4423643ce80e2a9ac94fa54ca49f"
	if got := Crypto.Sha512([]byte("abc")); got != sha512Abc {
		t.Errorf("Sha512(abc) = %q", got)
	}
	// Sha512Bytes / Sha512Str 与 Sha512 一致，长度均为 128
	if got := Crypto.Sha512Bytes(bin); len(got) != 128 || got != Crypto.Sha512(bin) {
		t.Errorf("Sha512Bytes 长度 %d, 与 Sha512 不一致", len(got))
	}
	if got := Crypto.Sha512Str(s); len(got) != 128 || got != Crypto.Sha512(bin) {
		t.Errorf("Sha512Str 长度 %d, 与 Sha512 不一致", len(got))
	}
	if got := Crypto.Sha512Str("abc"); got != strings.ToLower(got) {
		t.Errorf("Sha512Str 应输出小写 hex: %q", got)
	}
}

/* ---------------- HMAC 扩展 ---------------- */

func TestCryptoHmac(t *testing.T) {
	// 广为引用的测试向量（RFC 4231 同族）：key="key"，data=速老狐狸句子
	const key = "key"
	const data = "The quick brown fox jumps over the lazy dog"
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"HmacMd5", Crypto.HmacMd5(key, data), "80070713463e7749b90c2dc24911e275"},
		{"HmacSha1", Crypto.HmacSha1(key, data), "de7c9b85b8b78aa6bc8a7a36f70a90701c9db4d9"},
		{"HmacSha256", Crypto.HmacSha256(key, data), "f7bc83f430538424b13298e6aa6fb143ef4d59a14946175997479dbc2d1a3cd8"},
		{"HmacSha512", Crypto.HmacSha512(key, data), "b42af09057bac1e2d41708e48a902e09b5ff7f12ab428a4fe86653c73dd248fb" +
			"82f948a549f7b791a5b41915ee4d1ec3935357e4e2317250d0372afa2ebeeb3a"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}
	// 换 key 结果应不同
	if Crypto.HmacSha256("other", data) == Crypto.HmacSha256(key, data) {
		t.Error("不同密钥的 HMAC 不应相同")
	}
	if got := Crypto.HmacSha256(key, data); len(got) != 64 || got != strings.ToLower(got) {
		t.Errorf("HmacSha256 应为 64 位小写 hex: %q", got)
	}
}

/* ---------------- Base64 与 URL 安全 Base64 ---------------- */

func TestCryptoBase64(t *testing.T) {
	// 标准编码黄金值与往返
	if got := Crypto.Base64Encode("hello"); got != "aGVsbG8=" {
		t.Errorf("Base64Encode(hello) = %q, want \"aGVsbG8=\"", got)
	}
	if dec, err := Crypto.Base64Decode("aGVsbG8="); err != nil || dec != "hello" {
		t.Errorf("Base64Decode 往返失败: %q, %v", dec, err)
	}
	// 二进制数据往返（含 0 字节与非 UTF-8 字节）
	bin := string([]byte{0x00, 0xfb, 0xff, 0xfe, 0x0f})
	if dec, err := Crypto.Base64Decode(Crypto.Base64Encode(bin)); err != nil || dec != bin {
		t.Errorf("Base64 二进制往返失败: %q, %v", dec, err)
	}
	// 错误路径：非法输入
	if _, err := Crypto.Base64Decode("aGVsbG8"); err == nil { // 缺填充
		t.Error("Base64Decode 缺填充应返回错误")
	}
	if _, err := Crypto.Base64Decode("a=b"); err == nil { // 非法字符
		t.Error("Base64Decode 非法字符应返回错误")
	}

	// URL 安全编码：+/ 替换为 -_（0xfb 0xff 0xfe 的标准编码为 +//+）
	tri := string([]byte{0xfb, 0xff, 0xfe})
	std := Crypto.Base64Encode(tri)
	if std != "+//+" {
		t.Fatalf("测试前提失败: 标准编码应为 \"+//+\", got %q", std)
	}
	urlEnc := Crypto.Base64UrlEncode(tri)
	if urlEnc != "-__-" {
		t.Errorf("Base64UrlEncode = %q, want \"-__-\"", urlEnc)
	}
	if urlEnc == std {
		t.Error("URL 安全编码应与标准编码存在 +/ → -_ 的字符差异")
	}
	if strings.ContainsAny(urlEnc, "+/") {
		t.Errorf("URL 安全编码不应含 + 或 /: %q", urlEnc)
	}
	if dec, err := Crypto.Base64UrlDecode(urlEnc); err != nil || dec != tri {
		t.Errorf("Base64UrlDecode 往返失败: %q, %v", dec, err)
	}
	// URL 解码对标准字母表字符报错
	if _, err := Crypto.Base64UrlDecode(std); err == nil {
		t.Error("Base64UrlDecode 遇到 + 字符应返回错误")
	}
	// 含 + / 数据的长报文往返
	mix := string([]byte{0xfb, 0xef, 0xbe, 0xfe, 0xfa, 0x0f, 0x00, 0xff})
	if dec, err := Crypto.Base64UrlDecode(Crypto.Base64UrlEncode(mix)); err != nil || dec != mix {
		t.Errorf("Base64Url 长报文往返失败: %q, %v", dec, err)
	}
}

/* ---------------- UUID v4 ---------------- */

func TestCryptoUuid(t *testing.T) {
	u := Crypto.Uuid()
	if len(u) != 36 {
		t.Fatalf("Uuid 长度 = %d, want 36", len(u))
	}
	if strings.Count(u, "-") != 4 {
		t.Errorf("Uuid 连字符数 = %d, want 4: %q", strings.Count(u, "-"), u)
	}
	// 格式 8-4-4-4-12
	parts := strings.Split(u, "-")
	if len(parts[0]) != 8 || len(parts[1]) != 4 || len(parts[2]) != 4 || len(parts[3]) != 4 || len(parts[4]) != 12 {
		t.Errorf("Uuid 分段长度错误: %q", u)
	}
	if u[14] != '4' {
		t.Errorf("Uuid 版本位应为 4: %q", u)
	}
	if v := u[19]; v != '8' && v != '9' && v != 'a' && v != 'b' {
		t.Errorf("Uuid 变体位应为 8/9/a/b: %q", u)
	}
	hexPart := strings.ReplaceAll(u, "-", "")
	if hexPart != strings.ToLower(hexPart) || strings.ContainsAny(hexPart, "gHIJKLhijklm") {
		t.Errorf("Uuid 应为小写 hex: %q", u)
	}
	// 随机性：两次生成不应相同
	if again := Crypto.Uuid(); again == u {
		t.Errorf("两次 Uuid 不应相同: %q", u)
	}
}

/* ---------------- 参数签名 ---------------- */

func TestCryptoSignParams(t *testing.T) {
	// 乱序 map 同结果，且等于手动按 key 升序拼接后的摘要
	m1 := map[string]string{"b": "2", "a": "1", "c": "3"}
	m2 := map[string]string{"c": "3", "a": "1", "b": "2"}
	if got1, got2 := Crypto.SignParamsMd5(m1), Crypto.SignParamsMd5(m2); got1 != got2 {
		t.Errorf("乱序 map 签名应一致: %q vs %q", got1, got2)
	}
	if got, want := Crypto.SignParamsMd5(m1), Crypto.MD5("a1b2c3"); got != want {
		t.Errorf("SignParamsMd5 = %q, want md5(a1b2c3) = %q", got, want)
	}

	// otherParams 追加语义：等于手动拼接后摘要
	got := Crypto.SignParamsMd5(map[string]string{"a": "1"}, "&secret=x", "tail")
	if want := Crypto.MD5("a1&secret=xtail"); got != want {
		t.Errorf("SignParamsMd5(otherParams) = %q, want md5(a1&secret=xtail) = %q", got, want)
	}

	// 空 map：对空串摘要；追加 otherParams 后等于其摘要
	if got := Crypto.SignParamsMd5(map[string]string{}); got != "d41d8cd98f00b204e9800998ecf8427e" {
		t.Errorf("SignParamsMd5(空 map) = %q, want md5(\"\")", got)
	}
	if got, want := Crypto.SignParamsMd5(map[string]string{}, "abc"), Crypto.MD5("abc"); got != want {
		t.Errorf("SignParamsMd5(空 map, abc) = %q, want %q", got, want)
	}

	// SHA1 / SHA256 变体与对应摘要一致
	if got, want := Crypto.SignParamsSha1(m1), Crypto.SHA1("a1b2c3"); got != want {
		t.Errorf("SignParamsSha1 = %q, want %q", got, want)
	}
	if got, want := Crypto.SignParamsSha256(m2), Crypto.SHA256("a1b2c3"); got != want {
		t.Errorf("SignParamsSha256 = %q, want %q", got, want)
	}
	// nil map 等同空 map
	if got, want := Crypto.SignParamsSha1(nil), Crypto.SHA1(""); got != want {
		t.Errorf("SignParamsSha1(nil) = %q, want %q", got, want)
	}
}

/* ---------------- HOTP（RFC 4226） ---------------- */

func TestCryptoHotp(t *testing.T) {
	// RFC 4226 Appendix D 官方向量：secret 为 ASCII "12345678901234567890" 的 Base32
	const secret = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	want := []string{
		"755224", "287082", "359152", "969429", "338314",
		"254676", "287922", "162583", "399871", "520489",
	}
	for i, w := range want {
		got, err := Crypto.Hotp(secret, int64(i), 6)
		if err != nil {
			t.Fatalf("HOTP(counter=%d) 失败: %v", i, err)
		}
		if got != w {
			t.Errorf("HOTP(counter=%d) = %q, want %q", i, got, w)
		}
	}

	// 容忍小写 Base32
	if got, err := Crypto.Hotp(strings.ToLower(secret), 0, 6); err != nil || got != "755224" {
		t.Errorf("小写 Base32 应得到同值: %q, %v", got, err)
	}
	// 容忍缺失填充："MFRGG"（"abc" 无填充）与 "MFRGG===" 等价
	a, errA := Crypto.Hotp("MFRGG", 1, 6)
	b, errB := Crypto.Hotp("MFRGG===", 1, 6)
	if errA != nil || errB != nil || a != b {
		t.Errorf("缺失填充应得到同值: %q vs %q, %v %v", a, b, errA, errB)
	}

	// 位数 6~8 之外的值按 6 处理
	if got, _ := Crypto.Hotp(secret, 0, 5); got != "755224" {
		t.Errorf("digits=5 应回退为 6 位: %q", got)
	}
	if got, _ := Crypto.Hotp(secret, 0, 9); got != "755224" {
		t.Errorf("digits=9 应回退为 6 位: %q", got)
	}
	if got, _ := Crypto.Hotp(secret, 0, 8); len(got) != 8 {
		t.Errorf("digits=8 应输出 8 位: %q", got)
	}

	// 错误路径：非法 Base32
	if _, err := Crypto.Hotp("abc!", 0, 6); err == nil {
		t.Error("非法 Base32 密钥应返回错误")
	}
}

/* ---------------- TOTP（RFC 6238）与校验 ---------------- */

func TestCryptoTotp(t *testing.T) {
	// RFC 6238 Appendix B SHA-1 官方向量（8 位，secret 同 RFC 4226）
	const secret = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	cases := []struct {
		ts   int64
		want string
	}{
		{59, "94287082"},
		{1111111109, "07081804"},
		{1111111111, "14050471"},
		{1234567890, "89005924"},
		{2000000000, "69279037"},
		{20000000000, "65353130"},
	}
	for _, c := range cases {
		got, err := Crypto.Totp(secret, time.Unix(c.ts, 0), 8)
		if err != nil {
			t.Fatalf("Totp(T=%d) 失败: %v", c.ts, err)
		}
		if got != c.want {
			t.Errorf("Totp(T=%d) = %q, want %q", c.ts, got, c.want)
		}
	}
	// 错误路径：非法 Base32
	if _, err := Crypto.Totp("abc!", time.Now(), 6); err == nil {
		t.Error("Totp 非法 Base32 密钥应返回错误")
	}

	// TotpValidate：当前时刻生成 → 通过
	now := time.Now()
	code, err := Crypto.Totp(secret, now, 6)
	if err != nil {
		t.Fatalf("Totp 失败: %v", err)
	}
	if ok, err := Crypto.TotpValidate(secret, code, 1); err != nil || !ok {
		t.Errorf("当前口令应通过校验: %v, %v", ok, err)
	}
	// 相邻步长（+30s）在 window=1 内应通过
	codeNext, _ := Crypto.Totp(secret, now.Add(30*time.Second), 6)
	if ok, err := Crypto.TotpValidate(secret, codeNext, 1); err != nil || !ok {
		t.Errorf("相邻步长口令在 window=1 下应通过: %v, %v", ok, err)
	}
	// 超出 window（+90s = +3 步）不应通过
	codeFar, _ := Crypto.Totp(secret, now.Add(90*time.Second), 6)
	if ok, err := Crypto.TotpValidate(secret, codeFar, 1); err != nil || ok {
		t.Errorf("超出 window 的口令不应通过: %v, %v", ok, err)
	}
	// 过期口令（-90s）不应通过
	codeOld, _ := Crypto.Totp(secret, now.Add(-90*time.Second), 6)
	if ok, err := Crypto.TotpValidate(secret, codeOld, 1); err != nil || ok {
		t.Errorf("过期口令不应通过: %v, %v", ok, err)
	}
	// 8 位口令按其长度校验
	code8, _ := Crypto.Totp(secret, now, 8)
	if ok, err := Crypto.TotpValidate(secret, code8, 1); err != nil || !ok {
		t.Errorf("8 位口令应通过校验: %v, %v", ok, err)
	}
	// 错误路径：非法 Base32
	if _, err := Crypto.TotpValidate("!!", "123456", 1); err == nil {
		t.Error("TotpValidate 非法 Base32 密钥应返回错误")
	}
}

/* ---------------- Google Authenticator 密钥 ---------------- */

func TestCryptoGoogleSecretKey(t *testing.T) {
	secret, uri, err := Crypto.GoogleSecretKey("张三", 20)
	if err != nil {
		t.Fatalf("GoogleSecretKey 失败: %v", err)
	}
	// uri 前缀与 secret 组装
	if !strings.HasPrefix(uri, "otpauth://totp/张三?secret=") {
		t.Errorf("uri 前缀错误: %q", uri)
	}
	if !strings.HasSuffix(uri, secret) {
		t.Errorf("uri 应以 secret 结尾: uri=%q secret=%q", uri, secret)
	}
	// 20 字节 → 32 个 Base32 字符，且可被标准 Base32 解码
	if len(secret) != 32 {
		t.Errorf("secret 长度 = %d, want 32", len(secret))
	}
	if _, err := base32.StdEncoding.DecodeString(secret); err != nil {
		t.Errorf("secret 应可被标准 Base32 解码: %v", err)
	}
	// numBytes<=0 按常用 20 字节处理
	if s0, _, err := Crypto.GoogleSecretKey("acct", 0); err != nil || len(s0) != 32 {
		t.Errorf("numBytes=0 应按 20 字节生成 32 字符: len=%d, %v", len(s0), err)
	}
	// 10 字节 → 16 字符
	if s10, _, err := Crypto.GoogleSecretKey("acct", 10); err != nil || len(s10) != 16 {
		t.Errorf("numBytes=10 应生成 16 字符: len=%d, %v", len(s10), err)
	}
}

/* ---------------- PBKDF2 ---------------- */

func TestCryptoPbkdf2(t *testing.T) {
	// RFC 6070 官方向量（HmacSHA1，P="password"，S="salt"，dkLen=20）
	cases := []struct {
		iter int
		want string
	}{
		{1, "0c60c80f961f0e71f3a9b524af6012062fe037a6"},
		{2, "ea6c014dc72d6f8ccd1ed92ace1d41f0d8de8957"},
		{4096, "4b007901b765489abead49d926f721d065a429c1"},
	}
	for _, c := range cases {
		key, err := Crypto.Pbkdf2Key("password", []byte("salt"), c.iter, 20, sha1.New)
		if err != nil {
			t.Fatalf("Pbkdf2Key(iter=%d) 失败: %v", c.iter, err)
		}
		if got := hex.EncodeToString(key); got != c.want {
			t.Errorf("Pbkdf2Key(iter=%d) = %s, want %s", c.iter, got, c.want)
		}
	}

	// Pbkdf2Hex 固定参数（HmacSHA1/1000 次/64 字节）与 Pbkdf2Key 一致
	hexGot := Crypto.Pbkdf2Hex([]byte("password"), []byte("salt"))
	if len(hexGot) != 128 {
		t.Errorf("Pbkdf2Hex 长度 = %d, want 128", len(hexGot))
	}
	key, err := Crypto.Pbkdf2Key("password", []byte("salt"), 1000, 64, sha1.New)
	if err != nil {
		t.Fatalf("Pbkdf2Key 失败: %v", err)
	}
	if hexGot != hex.EncodeToString(key) {
		t.Errorf("Pbkdf2Hex = %s, want %s", hexGot, hex.EncodeToString(key))
	}

	// 不同迭代次数结果不同
	k1, _ := Crypto.Pbkdf2Key("password", []byte("salt"), 1, 32, sha1.New)
	k2, _ := Crypto.Pbkdf2Key("password", []byte("salt"), 2, 32, sha1.New)
	if hex.EncodeToString(k1) == hex.EncodeToString(k2) {
		t.Error("不同迭代次数的派生密钥不应相同")
	}

	// 错误路径：派生长度非法
	if _, err := Crypto.Pbkdf2Key("password", []byte("salt"), 1, 0, sha1.New); err == nil {
		t.Error("keyLength=0 应返回错误")
	}
}
