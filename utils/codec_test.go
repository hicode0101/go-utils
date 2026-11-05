package utils

import (
	"bytes"
	"strings"
	"testing"
)

/* ---------------- Base32（RFC 4648） ---------------- */

func TestBase32EncodeDecode(t *testing.T) {
	// RFC 4648 标准向量（5 字节一组、各种尾部残缺长度）
	rfcVectors := []struct {
		plain string
		want  string
	}{
		{"f", "MY======"},
		{"fo", "MZXQ===="},
		{"foo", "MZXW6==="},
		{"foob", "MZXW6YQ="},
		{"fooba", "MZXW6YTB"},
		{"foobar", "MZXW6YTBOI======"},
	}
	for _, v := range rfcVectors {
		if got := Base32.Encode([]byte(v.plain)); got != v.want {
			t.Errorf("Base32.Encode(%q) = %q, want %q", v.plain, got, v.want)
		}
		back, err := Base32.Decode(v.want)
		if err != nil || string(back) != v.plain {
			t.Errorf("Base32.Decode(%q) = %q, %v; want %q", v.want, back, err, v.plain)
		}
	}
	// hutool Base32Test 黄金值：UTF-8 中文串
	const cn = "伦家是一个非常长的字符串"
	const cnEncoded = "4S6KNZNOW3TJRL7EXCAOJOFK5GOZ5ZNYXDUZLP7HTKCOLLMX46WKNZFYWI======"
	if got := Base32.Encode([]byte(cn)); got != cnEncoded {
		t.Errorf("Base32.Encode(中文) = %q, want %q（hutool 黄金值）", got, cnEncoded)
	}
	if back, err := Base32.Decode(cnEncoded); err != nil || string(back) != cn {
		t.Errorf("Base32.Decode 往返失败: %q, %v", back, err)
	}
	// 空输入
	if got := Base32.Encode(nil); got != "" {
		t.Errorf("Base32.Encode(空) = %q, want \"\"", got)
	}
	if back, err := Base32.Decode(""); err != nil || len(back) != 0 {
		t.Errorf("Base32.Decode(\"\") = %v, %v; want 空切片无错误", back, err)
	}
	// 非法输入（'!' 不在 A-Z2-7 内）
	if _, err := Base32.Decode("!!!!"); err == nil {
		t.Error("Base32.Decode 非法字符应返回 error")
	}
}

func TestBase32EncodeHexDecodeHex(t *testing.T) {
	// hutool Base32Test 黄金值：Extended Hex 字母表 0-9A-V
	const cn = "伦家是一个非常长的字符串"
	const cnEncodedHex = "SIUADPDEMRJ9HBV4N20E9E5AT6EPTPDON3KPBFV7JA2EBBCNSUMADP5OM8======"
	if got := Base32.EncodeHex([]byte(cn)); got != cnEncodedHex {
		t.Errorf("Base32.EncodeHex(中文) = %q, want %q（hutool 黄金值）", got, cnEncodedHex)
	}
	if back, err := Base32.DecodeHex(cnEncodedHex); err != nil || string(back) != cn {
		t.Errorf("Base32.DecodeHex 往返失败: %q, %v", back, err)
	}
	if got := Base32.EncodeHex([]byte("foobar")); got != "CPNMUOJ1E8======" {
		t.Errorf("Base32.EncodeHex(foobar) = %q, want \"CPNMUOJ1E8======\"", got)
	}
	// 非法输入（Extended Hex 字母表为 0-9A-V，'W' 越界）
	if _, err := Base32.DecodeHex("!!!!!"); err == nil {
		t.Error("Base32.DecodeHex 非法字符应返回 error")
	}
}

/* ---------------- Base58（Bitcoin 字母表） ---------------- */

func TestBase58EncodeDecode(t *testing.T) {
	// Bitcoin 社区标准向量（原 utils_test.go 已验证）
	if got := Base58.Encode([]byte("hello world")); got != "StV1DL6CwTryKyV" {
		t.Errorf("Base58.Encode(hello world) = %q, want \"StV1DL6CwTryKyV\"", got)
	}
	if back, err := Base58.Decode("StV1DL6CwTryKyV"); err != nil || string(back) != "hello world" {
		t.Errorf("Base58.Decode 往返失败: %q, %v", back, err)
	}
	// 前导 0x00 逐个映射为字母表首字符 '1'
	if got := Base58.Encode([]byte{0, 0, 1}); got != "112" {
		t.Errorf("Base58.Encode({0,0,1}) = %q, want \"112\"", got)
	}
	if back, err := Base58.Decode("112"); err != nil || !bytes.Equal(back, []byte{0, 0, 1}) {
		t.Errorf("Base58.Decode(\"112\") = %v, %v; want [0 0 1]", back, err)
	}
	if got := Base58.Encode([]byte{0}); got != "1" {
		t.Errorf("Base58.Encode({0}) = %q, want \"1\"", got)
	}
	// 空输入
	if got := Base58.Encode(nil); got != "" {
		t.Errorf("Base58.Encode(空) = %q, want \"\"", got)
	}
	if back, err := Base58.Decode(""); err != nil || len(back) != 0 {
		t.Errorf("Base58.Decode(\"\") = %v, %v; want 空切片无错误", back, err)
	}
	// 非法字符（0、O、I、l 均被 Bitcoin 字母表剔除）
	if _, err := Base58.Decode("0OIl"); err == nil {
		t.Error("Base58.Decode 非法字符 0OIl 应返回 error")
	}
}

func TestBase58Checked(t *testing.T) {
	// hutool Base58Test 黄金值：version=0 时 version 字节 0x00 编码为前导 '1'
	if got := Base58.EncodeChecked(0, []byte("hello world")); got != "13vQB7B6MrGQZaxCuFg4oh" {
		t.Errorf("EncodeChecked(0, hello world) = %q, want \"13vQB7B6MrGQZaxCuFg4oh\"", got)
	}
	// hutool decodeCheckedTest 黄金值：无版本串（versionless）
	if back, err := Base58.DecodeChecked("3vQB7B6MrGQZaxCuFg4oh", false); err != nil || string(back) != "hello world" {
		t.Errorf("DecodeChecked(versionless) = %q, %v; want \"hello world\"", back, err)
	}
	// 带版本往返（withVersion=true 保留 version 字节语义下的载荷一致性）
	enc := Base58.EncodeChecked(1, []byte("hello world"))
	if back, err := Base58.DecodeChecked(enc, true); err != nil || string(back) != "hello world" {
		t.Errorf("DecodeChecked(withVersion) 往返失败: %q, %v", back, err)
	}
	// 篡改末位（校验和不匹配）应失败
	if _, err := Base58.DecodeChecked("3vQB7B6MrGQZaxCuFg4oi", false); err == nil {
		t.Error("DecodeChecked 篡改后的校验和应返回 error")
	}
	// 解码结果不足 4 字节（缺少校验和）应失败
	if _, err := Base58.DecodeChecked("2s", false); err == nil {
		t.Error("DecodeChecked 数据过短应返回 error")
	}
}

/* ---------------- Base62（GMP 标准与反转字母表） ---------------- */

func TestBase62EncodeDecode(t *testing.T) {
	// hutool Base62Test 黄金值（标准字母表 0-9A-Za-z）
	const cn = "伦家是一个非常长的字符串66"
	const stdEncoded = "17vKU8W4JMG8dQF8lk9VNnkdMOeWn4rJMva6F0XsLrrT53iKBnqo"
	if got := Base62.Encode([]byte(cn)); got != stdEncoded {
		t.Errorf("Base62.Encode(中文) = %q, want %q（hutool 黄金值）", got, stdEncoded)
	}
	if back, err := Base62.Decode(stdEncoded); err != nil || string(back) != cn {
		t.Errorf("Base62.Decode 往返失败: %q, %v", back, err)
	}
	// 原始数据前导 0x00 映射为前导 '0' 字符且可逆
	if got := Base62.Encode([]byte{0, 0, 5}); got != "005" {
		t.Errorf("Base62.Encode({0,0,5}) = %q, want \"005\"", got)
	}
	if back, err := Base62.Decode("005"); err != nil || !bytes.Equal(back, []byte{0, 0, 5}) {
		t.Errorf("Base62.Decode(\"005\") = %v, %v; want [0 0 5]", back, err)
	}
	// 非法字符
	if _, err := Base62.Decode("!!"); err == nil {
		t.Error("Base62.Decode 非法字符应返回 error")
	}
}

func TestBase62EncodeInvertedDecodeInverted(t *testing.T) {
	// hutool Base62Test 黄金值（反转字母表 0-9a-zA-Z）
	const cn = "伦家是一个非常长的字符串66"
	const invEncoded = "17Vku8w4jmg8Dqf8LK9vnNKDmoEwN4RjmVA6f0xSlRRt53IkbNQO"
	if got := Base62.EncodeInverted([]byte(cn)); got != invEncoded {
		t.Errorf("Base62.EncodeInverted(中文) = %q, want %q（hutool 黄金值）", got, invEncoded)
	}
	if back, err := Base62.DecodeInverted(invEncoded); err != nil || string(back) != cn {
		t.Errorf("Base62.DecodeInverted 往返失败: %q, %v", back, err)
	}
	// 两种字母表数字段之后的第 1 个字符不同：值 10 标准表为 'A'、反转表为 'a'
	if got := Base62.Encode([]byte{10}); got != "A" {
		t.Errorf("Base62.Encode({10}) = %q, want \"A\"", got)
	}
	if got := Base62.EncodeInverted([]byte{10}); got != "a" {
		t.Errorf("Base62.EncodeInverted({10}) = %q, want \"a\"", got)
	}
	if Base62.Encode([]byte(cn)) == Base62.EncodeInverted([]byte(cn)) {
		t.Error("两种字母表的编码结果不应相同")
	}
	// 中文多字节输入往返
	for _, s := range []string{"中文编码测试", "go-utils 工具库"} {
		if back, err := Base62.DecodeInverted(Base62.EncodeInverted([]byte(s))); err != nil || string(back) != s {
			t.Errorf("Base62 反转字母表中文往返失败: %q, %v", back, err)
		}
	}
	// 非法字符
	if _, err := Base62.DecodeInverted("!!"); err == nil {
		t.Error("Base62.DecodeInverted 非法字符应返回 error")
	}
}

/* ---------------- Base85（Ascii85 与 Z85） ---------------- */

func TestBase85EncodeDecode(t *testing.T) {
	// ascii85 经典向量（Wikipedia 样例首分组）
	if got := Base85.Encode([]byte("Man ")); got != "9jqo^" {
		t.Errorf("Base85.Encode(\"Man \") = %q, want \"9jqo^\"", got)
	}
	// 全 0 的 4 字节组压缩为 'z'（Ascii85 规范行为，Go 标准库语义）
	if got := Base85.Encode([]byte{0, 0, 0, 0}); got != "z" {
		t.Errorf("Base85.Encode({0,0,0,0}) = %q, want \"z\"", got)
	}
	if got := Base85.Encode([]byte("hello")); got != "BOu!rDZ" {
		t.Errorf("Base85.Encode(\"hello\") = %q, want \"BOu!rDZ\"", got)
	}
	// 解码容忍空白字符
	if back, err := Base85.Decode("9jq o^"); err != nil || string(back) != "Man " {
		t.Errorf("Base85.Decode 带空白 = %q, %v; want \"Man \"", back, err)
	}
	// 各种尾部残缺长度（1/2/3/4/5 字节）往返
	for n := 1; n <= 5; n++ {
		data := make([]byte, n)
		for i := range data {
			data[i] = byte(0x40 + i)
		}
		back, err := Base85.Decode(Base85.Encode(data))
		if err != nil || !bytes.Equal(back, data) {
			t.Errorf("Base85 %d 字节往返失败: %v, %v", n, back, err)
		}
	}
	// 空输入
	if got := Base85.Encode(nil); got != "" {
		t.Errorf("Base85.Encode(空) = %q, want \"\"", got)
	}
	if back, err := Base85.Decode(""); err != nil || len(back) != 0 {
		t.Errorf("Base85.Decode(\"\") = %v, %v; want 空切片无错误", back, err)
	}
	// 非法字符（'~' 超出 Ascii85 的 '!'~'u' 范围）
	if _, err := Base85.Decode("~~~~"); err == nil {
		t.Error("Base85.Decode 非法字符应返回 error")
	}
}

func TestBase85EncodeZ85DecodeZ85(t *testing.T) {
	// Z85 规范（ZeroMQ RFC 32）标准向量
	z85Bytes := []byte{0x86, 0x4F, 0xD2, 0x6F, 0xB5, 0x59, 0xF7, 0x5B}
	if got, err := Base85.EncodeZ85(z85Bytes); err != nil || got != "HelloWorld" {
		t.Errorf("Base85.EncodeZ85 规范向量 = %q, %v; want \"HelloWorld\"", got, err)
	}
	if back, err := Base85.DecodeZ85("HelloWorld"); err != nil || !bytes.Equal(back, z85Bytes) {
		t.Errorf("Base85.DecodeZ85(\"HelloWorld\") = % X, %v", back, err)
	}
	// 长度为 4 倍数的输入往返
	if got, err := Base85.EncodeZ85([]byte{1, 2, 3, 4}); err != nil {
		t.Fatalf("Base85.EncodeZ85 失败: %v", err)
	} else if back, err := Base85.DecodeZ85(got); err != nil || !bytes.Equal(back, []byte{1, 2, 3, 4}) {
		t.Errorf("Base85 Z85 往返失败: %v, %v", back, err)
	}
	// 空输入
	if got, err := Base85.EncodeZ85(nil); err != nil || got != "" {
		t.Errorf("Base85.EncodeZ85(空) = %q, %v; want \"\"", got, err)
	}
	if back, err := Base85.DecodeZ85(""); err != nil || len(back) != 0 {
		t.Errorf("Base85.DecodeZ85(\"\") = %v, %v; want 空切片无错误", back, err)
	}
	// 长度非法：编码要求数据为 4 的倍数、解码要求字符串为 5 的倍数
	if _, err := Base85.EncodeZ85([]byte{1, 2, 3}); err == nil {
		t.Error("EncodeZ85 长度非 4 倍数应返回 error")
	}
	if _, err := Base85.DecodeZ85("1234"); err == nil {
		t.Error("DecodeZ85 长度非 5 倍数应返回 error")
	}
	// 非法字符（'`' 不在 Z85 字母表内）
	if _, err := Base85.DecodeZ85("12`34"); err == nil {
		t.Error("DecodeZ85 非法字符应返回 error")
	}
}

/* ---------------- BCD 码 ---------------- */

func TestBCDStrToBcd(t *testing.T) {
	// hutool BCDTest 向量：数字与十六进制字母混合
	bcd, err := BCD.StrToBcd("123456ABCDEF")
	if err != nil {
		t.Fatalf("StrToBcd 失败: %v", err)
	}
	if want := []byte{0x12, 0x34, 0x56, 0xAB, 0xCD, 0xEF}; !bytes.Equal(bcd, want) {
		t.Errorf("StrToBcd(123456ABCDEF) = % X, want % X", bcd, want)
	}
	// 奇数位长度按 hutool 规则左侧补 '0'
	bcdOdd, err := BCD.StrToBcd("123")
	if err != nil {
		t.Fatalf("StrToBcd(123) 失败: %v", err)
	}
	if want := []byte{0x01, 0x23}; !bytes.Equal(bcdOdd, want) {
		t.Errorf("StrToBcd(123) = % X, want % X（左补 0）", bcdOdd, want)
	}
	// 小写十六进制字母同样支持
	bcdLower, err := BCD.StrToBcd("abcdef")
	if err != nil {
		t.Fatalf("StrToBcd(abcdef) 失败: %v", err)
	}
	if want := []byte{0xAB, 0xCD, 0xEF}; !bytes.Equal(bcdLower, want) {
		t.Errorf("StrToBcd(abcdef) = % X, want % X", bcdLower, want)
	}
	// 非法字符（'g' 超出 0-9 与 a-f/A-F）
	if _, err := BCD.StrToBcd("12g"); err == nil {
		t.Error("StrToBcd 非法字符应返回 error")
	}
}

func TestBCDBcdToStr(t *testing.T) {
	if got := BCD.BcdToStr([]byte{0x12, 0x34}); got != "1234" {
		t.Errorf("BcdToStr({0x12,0x34}) = %q, want \"1234\"", got)
	}
	// hutool BCDTest 黄金值：含十六进制字母时以大写还原
	if got := BCD.BcdToStr([]byte{0xAB, 0xCD, 0xEF}); got != "ABCDEF" {
		t.Errorf("BcdToStr({0xAB,0xCD,0xEF}) = %q, want \"ABCDEF\"", got)
	}
	// 与 StrToBcd 互逆（原 utils_test.go 已验证向量）
	bcd, err := BCD.StrToBcd("12345678")
	if err != nil {
		t.Fatalf("StrToBcd 失败: %v", err)
	}
	if got := BCD.BcdToStr(bcd); got != "12345678" {
		t.Errorf("BCD 往返失败: %q", got)
	}
}

/* ---------------- Caesar 凯撒密码 ---------------- */

func TestCaesarEncodeDecode(t *testing.T) {
	// hutool CaesarTest 黄金值
	const hutoolPlain = "1f2e9df6131b480b9fdddc633cf24996"
	if got := Caesar.Encode(hutoolPlain, 3); got != "1H2G9FH6131D480D9HFFFE633EH24996" {
		t.Errorf("Caesar.Encode(hutool 向量, 3) = %q", got)
	}
	if got := Caesar.Decode("1H2G9FH6131D480D9HFFFE633EH24996", 3); got != hutoolPlain {
		t.Errorf("Caesar.Decode 往返失败: %q", got)
	}
	// 负偏移量（hutool encodeNegativeOffsetTest 黄金值）
	if got := Caesar.Encode("abc", -3); got != "ZAB" {
		t.Errorf("Caesar.Encode(abc, -3) = %q, want \"ZAB\"", got)
	}
	if got := Caesar.Decode("abc", 3); got != "ZAB" {
		t.Errorf("Caesar.Decode(abc, 3) = %q, want \"ZAB\"", got)
	}
	if got := Caesar.Encode("AttackAtDawn", -1); got != "zTTACKzTcAWN" {
		t.Errorf("Caesar.Encode(AttackAtDawn, -1) = %q, want \"zTTACKzTcAWN\"", got)
	}
	// 偏移量按 52 循环（正、负、超过 52）
	if got := Caesar.Encode("abc", 3); got != "CDE" {
		t.Errorf("Caesar.Encode(abc, 3) = %q, want \"CDE\"", got)
	}
	for _, tc := range []struct {
		offset int
		want   string
	}{
		{0, "abc"}, {52, "abc"}, {-52, "abc"}, {55, "CDE"}, {-55, "ZAB"},
	} {
		if got := Caesar.Encode("abc", tc.offset); got != tc.want {
			t.Errorf("Caesar.Encode(abc, %d) = %q, want %q", tc.offset, got, tc.want)
		}
	}
	// 表外字符（中文、音标字母、数字、符号）原样保留
	if got := Caesar.Encode("中文", 3); got != "中文" {
		t.Errorf("Caesar.Encode(中文, 3) = %q, want 原样", got)
	}
	if got := Caesar.Encode("café", 2); got != "dbgé" {
		t.Errorf("Caesar.Encode(café, 2) = %q, want \"dbgé\"", got)
	}
	if got := Caesar.Encode("你好, Hutool! 123", 5); got != "你好, jXWRRO! 123" {
		t.Errorf("Caesar.Encode 混合文本 = %q, want \"你好, jXWRRO! 123\"", got)
	}
	if got := Caesar.Decode("你好, jXWRRO! 123", 5); got != "你好, Hutool! 123" {
		t.Errorf("Caesar.Decode 混合文本 = %q", got)
	}
	// 空串与往返（各偏移量，含负数与超过 52 的值）
	if got := Caesar.Encode("", 3); got != "" {
		t.Errorf("Caesar.Encode(\"\") = %q, want \"\"", got)
	}
	const msg = "AttackAtDawn, Hutool 5.8!"
	for _, offset := range []int{0, 1, 3, 25, 26, 51, 52, 53, -1, -3, -26, -52, -53, -100} {
		if got := Caesar.Decode(Caesar.Encode(msg, offset), offset); got != msg {
			t.Errorf("Caesar 往返失败（offset=%d）: %q", offset, got)
		}
	}
}

/* ---------------- Rot ROT-N ---------------- */

func TestRotEncodeDecode(t *testing.T) {
	// ROT13 经典向量（原 utils_test.go 已验证）
	if got := Rot.Encode13("hello", false); got != "uryyb" {
		t.Errorf("Rot.Encode13(hello) = %q, want \"uryyb\"", got)
	}
	if got := Rot.Decode13("uryyb", false); got != "hello" {
		t.Errorf("Rot.Decode13(uryyb) = %q, want \"hello\"", got)
	}
	// hutool RotTest 黄金值：encode13 默认同时位移数字（数字走 ROT5，即 +13 模 10）
	const hutoolPlain = "1f2e9df6131b480b9fdddc633cf24996"
	if got := Rot.Encode13(hutoolPlain, true); got != "4s5r2qs9464o713o2sqqqp966ps57229" {
		t.Errorf("Rot.Encode13(hutool 向量, true) = %q, want \"4s5r2qs9464o713o2sqqqp966ps57229\"", got)
	}
	if got := Rot.Decode13("4s5r2qs9464o713o2sqqqp966ps57229", true); got != hutoolPlain {
		t.Errorf("Rot.Decode13 往返失败: %q", got)
	}
	// 通用 Encode：字母与数字（encodeNumber=true 时按模 10 位移）
	if got := Rot.Encode("hello", 1, false); got != "ifmmp" {
		t.Errorf("Rot.Encode(hello, 1, false) = %q, want \"ifmmp\"", got)
	}
	if got := Rot.Encode("0123456789", 13, true); got != "3456789012" {
		t.Errorf("Rot.Encode 数字位移 = %q, want \"3456789012\"", got)
	}
	// 按源码语义：Encode 负位移不做负数修正（'a'-1 越出字母区），Decode 负位移有修正
	if got := Rot.Encode("abc", -1, false); got != "`ab" {
		t.Errorf("Rot.Encode(abc, -1, false) = %q, want \"`ab\"（源码语义）", got)
	}
	if got := Rot.Decode("abc", -1, false); got != "bcd" {
		t.Errorf("Rot.Decode(abc, -1, false) = %q, want \"bcd\"（源码语义）", got)
	}
	// 表外字符原样、通用偏移往返
	if got := Rot.Encode("a-b_c!", 7, false); got != "h-i_j!" {
		t.Errorf("Rot.Encode 表外字符 = %q, want \"h-i_j!\"", got)
	}
	const msg = "Attack At Dawn, 123!"
	for _, offset := range []int{1, 5, 13, 25, 26} {
		if got := Rot.Decode(Rot.Encode(msg, offset, true), offset, true); got != msg {
			t.Errorf("Rot 往返失败（offset=%d）: %q", offset, got)
		}
	}
}

/* ---------------- Morse 摩斯密码 ---------------- */

func TestMorseEncodeDecode(t *testing.T) {
	// 黄金值（对照 hutool MorseTest 码表：S="..."、O="---"，点划分别为 . -）
	if got := Morse.Encode("SOS"); got != "... --- ..." {
		t.Errorf("Morse.Encode(SOS) = %q, want \"... --- ...\"", got)
	}
	if got := Morse.Decode("... --- ..."); got != "SOS" {
		t.Errorf("Morse.Decode(\"... --- ...\") = %q, want \"SOS\"", got)
	}
	// 单词分隔符 " /"（原 utils_test.go 已验证向量）
	if got := Morse.Encode("HI YOU"); got != ".... .. / -.-- --- ..-" {
		t.Errorf("Morse.Encode(HI YOU) = %q, want \".... .. / -.-- --- ..-\"", got)
	}
	if got := Morse.Decode(".... .. / -.-- --- ..-"); got != "HI YOU" {
		t.Errorf("Morse.Decode 单词分隔 = %q, want \"HI YOU\"", got)
	}
	// 大小写统一按大写处理、数字参与编码，整体往返
	const msg = "Hello World 123"
	if got := Morse.Decode(Morse.Encode(msg)); got != strings.ToUpper(msg) {
		t.Errorf("Morse 往返失败: %q, want %q", got, strings.ToUpper(msg))
	}
	// 未知字符跳过（'-' 编码为划，注意与分隔符区分）
	if got := Morse.Encode("A#B"); got != ".- -..." {
		t.Errorf("Morse.Encode(A#B) = %q, want \".- -...\"（# 跳过）", got)
	}
	if got := Morse.Encode("#"); got != "" {
		t.Errorf("Morse.Encode(#全未知) = %q, want \"\"", got)
	}
	// 未知电码解码时跳过
	if got := Morse.Decode("--------"); got != "" {
		t.Errorf("Morse.Decode(未知电码) = %q, want \"\"", got)
	}
}

/* ---------------- Punycode（RFC 3492） ---------------- */

func TestPunycodeEncodeDecode(t *testing.T) {
	// RFC 3492 样例（原 utils_test.go 已验证）
	if got := Punycode.Encode("bücher", true); got != "xn--bcher-kva" {
		t.Errorf("Punycode.Encode(bücher, true) = %q, want \"xn--bcher-kva\"", got)
	}
	if got := Punycode.Encode("bücher", false); got != "bcher-kva" {
		t.Errorf("Punycode.Encode(bücher, false) = %q, want \"bcher-kva\"", got)
	}
	// hutool PunyCodeTest 黄金值：混合基础码点 + 非基础码点
	if got := Punycode.Encode("Hutool编码器", false); got != "Hutool-ux9js33tgln" {
		t.Errorf("Punycode.Encode(Hutool编码器) = %q, want \"Hutool-ux9js33tgln\"", got)
	}
	if got := Punycode.Encode("Hutool编码器", true); got != "xn--Hutool-ux9js33tgln" {
		t.Errorf("Punycode.Encode 带前缀 = %q, want \"xn--Hutool-ux9js33tgln\"", got)
	}
	// 解码自动去掉 xn-- 前缀（忽略大小写，两种入参）
	if got, err := Punycode.Decode("Hutool-ux9js33tgln"); err != nil || got != "Hutool编码器" {
		t.Errorf("Punycode.Decode 无前缀 = %q, %v", got, err)
	}
	if got, err := Punycode.Decode("xn--Hutool-ux9js33tgln"); err != nil || got != "Hutool编码器" {
		t.Errorf("Punycode.Decode 带前缀 = %q, %v", got, err)
	}
	// 纯 ASCII 输入直接原样返回（不加前缀）
	if got := Punycode.Encode("Hutool", true); got != "Hutool" {
		t.Errorf("Punycode.Encode(纯 ASCII) = %q, want \"Hutool\"", got)
	}
	// 中文样例（经临时程序验证）
	if got := Punycode.Encode("你好", true); got != "xn--6qq79v" {
		t.Errorf("Punycode.Encode(你好, true) = %q, want \"xn--6qq79v\"", got)
	}
	// 非法输入：含非法编码字符
	if _, err := Punycode.Decode("xn--!!!"); err == nil {
		t.Error("Punycode.Decode 非法字符应返回 error")
	}
	if _, err := Punycode.Decode("bücher"); err == nil {
		t.Error("Punycode.Decode 非 ASCII 输入应返回 error")
	}
}

func TestPunycodeEncodeDomainDecodeDomain(t *testing.T) {
	// hutool PunyCodeTest 黄金值（含大写 TLD 原样保留）
	if got := Punycode.EncodeDomain("赵新虎.com"); got != "xn--efvz93e52e.com" {
		t.Errorf("EncodeDomain(赵新虎.com) = %q, want \"xn--efvz93e52e.com\"", got)
	}
	if got := Punycode.EncodeDomain("赵新虎.COM"); got != "xn--efvz93e52e.COM" {
		t.Errorf("EncodeDomain(赵新虎.COM) = %q, want \"xn--efvz93e52e.COM\"", got)
	}
	// 纯 ASCII 域名原样保留
	if got := Punycode.EncodeDomain("example.com"); got != "example.com" {
		t.Errorf("EncodeDomain(纯 ASCII) = %q, want \"example.com\"", got)
	}
	// 多标签混合：仅 xn-- 标签参与解码
	if got := Punycode.EncodeDomain("www.你好.com"); got != "www.xn--6qq79v.com" {
		t.Errorf("EncodeDomain 多标签 = %q, want \"www.xn--6qq79v.com\"", got)
	}
	if got, err := Punycode.DecodeDomain("www.xn--6qq79v.com"); err != nil || got != "www.你好.com" {
		t.Errorf("DecodeDomain 多标签 = %q, %v", got, err)
	}
	// hutool 黄金值与原 utils_test.go 已验证向量
	if got, err := Punycode.DecodeDomain("xn--efvz93e52e.com"); err != nil || got != "赵新虎.com" {
		t.Errorf("DecodeDomain(赵新虎.com) = %q, %v", got, err)
	}
	if got, err := Punycode.DecodeDomain("xn--bcher-kva.com"); err != nil || got != "bücher.com" {
		t.Errorf("DecodeDomain(bcher) = %q, %v", got, err)
	}
	// 域名整体往返（多标签 + 全中文 TLD）
	const domain = "赵新虎.中国"
	if got, err := Punycode.DecodeDomain(Punycode.EncodeDomain(domain)); err != nil || got != domain {
		t.Errorf("DecodeDomain 往返失败: %q, %v", got, err)
	}
	// 非法标签返回 error
	if _, err := Punycode.DecodeDomain("xn--!!.com"); err == nil {
		t.Error("DecodeDomain 非法标签应返回 error")
	}
}

/* ---------------- Radix 自定义进制 ---------------- */

func TestRadixConstants(t *testing.T) {
	// 常量长度与无重复字符
	if len(Radix34) != 34 {
		t.Errorf("len(Radix34) = %d, want 34", len(Radix34))
	}
	if len(Radix59) != 59 {
		t.Errorf("len(Radix59) = %d, want 59", len(Radix59))
	}
	for name, alphabet := range map[string]string{"Radix34": Radix34, "Radix59": Radix59} {
		seen := make(map[rune]bool, len(alphabet))
		for _, c := range alphabet {
			if seen[c] {
				t.Errorf("%s 含重复字符 %q", name, c)
			}
			seen[c] = true
		}
	}
}

func TestRadixEncodeDecodeToInt(t *testing.T) {
	// 固定值（经临时程序验证）
	if got, err := Radix.Encode(Radix34, 100); err != nil || got != "U7" {
		t.Errorf("Radix.Encode(Radix34, 100) = %q, %v; want \"U7\"", got, err)
	}
	if got, err := Radix.Encode(Radix59, 100); err != nil || got != "hL" {
		t.Errorf("Radix.Encode(Radix59, 100) = %q, %v; want \"hL\"", got, err)
	}
	if got, _ := Radix.Encode(Radix34, 0); got != "H" {
		t.Errorf("Radix.Encode(Radix34, 0) = %q, want \"H\"", got)
	}
	if got, _ := Radix.Encode(Radix59, 1); got != "h" {
		t.Errorf("Radix.Encode(Radix59, 1) = %q, want \"h\"", got)
	}
	// 往返：0、1、大数、int64 上界
	for _, alphabet := range []string{Radix34, Radix59} {
		for _, num := range []int64{0, 1, 42, 123456789, 9223372036854775807} {
			enc, err := Radix.Encode(alphabet, num)
			if err != nil {
				t.Fatalf("Radix.Encode(%d) 失败: %v", num, err)
			}
			back, err := Radix.DecodeToInt(alphabet, enc)
			if err != nil || back != num {
				t.Errorf("Radix 往返失败: %d -> %q -> %d, %v", num, enc, back, err)
			}
		}
	}
	if got, _ := Radix.Encode(Radix34, 123456789); got != "U4JU23" {
		t.Errorf("Radix.Encode(Radix34, 123456789) = %q, want \"U4JU23\"", got)
	}
	// 非法字母表：长度不足、含重复字符
	if _, err := Radix.Encode("A", 5); err == nil {
		t.Error("Encode 字母表长度不足 2 应返回 error")
	}
	if _, err := Radix.Encode("AA", 5); err == nil {
		t.Error("Encode 字母表含重复字符应返回 error")
	}
	// 负数
	if _, err := Radix.Encode(Radix34, -1); err == nil {
		t.Error("Encode 负数应返回 error")
	}
	// DecodeToInt 失败路径
	if _, err := Radix.DecodeToInt(Radix34, ""); err == nil {
		t.Error("DecodeToInt 空编码串应返回 error")
	}
	if _, err := Radix.DecodeToInt("A", "11"); err == nil {
		t.Error("DecodeToInt 字母表过短应返回 error")
	}
	// hutool RadixUtilTest 黄金场景：编码串含字母表之外字符
	if _, err := Radix.DecodeToInt("0123456789ABC", "1X3"); err == nil {
		t.Error("DecodeToInt 非法字符 'X' 应返回 error")
	}
}

/* ---------------- Hashids ---------------- */

func TestHashidsEncodeDecode(t *testing.T) {
	// 默认字母表（盐为空）：hashids v1 经典向量
	h := NewHashids("", 0)
	enc, err := h.Encode(1, 2, 3)
	if err != nil {
		t.Fatalf("Hashids.Encode 失败: %v", err)
	}
	if enc != "o2fXhV" {
		t.Errorf("Hashids.Encode(1,2,3) = %q, want \"o2fXhV\"", enc)
	}
	if back, err := h.Decode(enc); err != nil || len(back) != 3 || back[0] != 1 || back[1] != 2 || back[2] != 3 {
		t.Errorf("Hashids.Decode 往返失败: %v, %v", back, err)
	}
	// 带盐（原 utils_test.go 已验证往返）
	hs := NewHashids("this is my salt", 0)
	encS, err := hs.Encode(1, 2, 3)
	if err != nil {
		t.Fatalf("Hashids.Encode 带盐失败: %v", err)
	}
	if encS != "laHquq" {
		t.Errorf("Hashids.Encode 带盐 = %q, want \"laHquq\"", encS)
	}
	if back, err := hs.Decode(encS); err != nil || len(back) != 3 || back[0] != 1 || back[2] != 3 {
		t.Errorf("Hashids 带盐往返失败: %v, %v", back, err)
	}
	// 单数、0、大数
	h2 := NewHashids("", 0)
	if enc347, err := h2.Encode(347); err != nil {
		t.Fatalf("Hashids.Encode(347) 失败: %v", err)
	} else if back, err := h2.Decode(enc347); err != nil || len(back) != 1 || back[0] != 347 {
		t.Errorf("Hashids 单数往返失败: %v, %v", back, err)
	}
	if enc0, err := h2.Encode(0); err != nil {
		t.Fatalf("Hashids.Encode(0) 失败: %v", err)
	} else if back, err := h2.Decode(enc0); err != nil || len(back) != 1 || back[0] != 0 {
		t.Errorf("Hashids 0 往返失败: %v, %v", back, err)
	}
	encMax, err := h2.Encode(9223372036854775807)
	if err != nil {
		t.Fatalf("Hashids.Encode(MaxInt64) 失败: %v", err)
	}
	if back, err := h2.Decode(encMax); err != nil || len(back) != 1 || back[0] != 9223372036854775807 {
		t.Errorf("Hashids MaxInt64 往返失败: %v, %v", back, err)
	}
	// 失败路径：空参数、负数、非法 hash
	if _, err := h2.Encode(); err == nil {
		t.Error("Hashids.Encode 空参数应返回 error")
	}
	if _, err := h2.Encode(1, -2); err == nil {
		t.Error("Hashids.Encode 负数应返回 error")
	}
	if _, err := h2.Decode("gYX"); err == nil {
		t.Error("Hashids.Decode 非法 hash 应返回 error")
	}
	if _, err := h2.Decode("！"); err == nil {
		t.Error("Hashids.Decode 字母表外字符应返回 error")
	}
}

func TestHashidsMinLength(t *testing.T) {
	// minLength 生效：输出长度不小于最小长度（经临时程序验证固定值）
	h := NewHashids("salt", 10)
	enc, err := h.Encode(1, 2, 3)
	if err != nil {
		t.Fatalf("Hashids.Encode 失败: %v", err)
	}
	if len(enc) < 10 {
		t.Errorf("Hashids minLength=10 输出长度 %d < 10: %q", len(enc), enc)
	}
	if enc != "a8JWiouVk1" {
		t.Errorf("Hashids minLength 输出 = %q, want \"a8JWiouVk1\"", enc)
	}
	if back, err := h.Decode(enc); err != nil || len(back) != 3 || back[0] != 1 || back[2] != 3 {
		t.Errorf("Hashids minLength 往返失败: %v, %v", back, err)
	}
	// minLength 为负时按 0 处理，不 panic 且可逆
	hn := NewHashids("s", -5)
	if enc, err := hn.Encode(7); err != nil {
		t.Fatalf("Hashids.Encode 失败: %v", err)
	} else if back, err := hn.Decode(enc); err != nil || len(back) != 1 || back[0] != 7 {
		t.Errorf("Hashids 负 minLength 往返失败: %v, %v", back, err)
	}
}

func TestHashidsEncodeHexDecodeToHex(t *testing.T) {
	// hutool HashidsTest 黄金值：0x/0X 前缀等价
	h := NewHashids("my awesome salt", 0)
	const hexID = "507f1f77bcf86cd799439011"
	enc1, err := h.EncodeHex(hexID)
	if err != nil {
		t.Fatalf("EncodeHex 失败: %v", err)
	}
	if enc1 != "R2qnd2vkOJTXm7XV7yq4" {
		t.Errorf("EncodeHex = %q, want \"R2qnd2vkOJTXm7XV7yq4\"（hutool 黄金值）", enc1)
	}
	enc2, err := h.EncodeHex("0x" + hexID)
	if err != nil || enc2 != enc1 {
		t.Errorf("EncodeHex(0x 前缀) = %q, %v; want %q", enc2, err, enc1)
	}
	enc3, err := h.EncodeHex("0X" + hexID)
	if err != nil || enc3 != enc1 {
		t.Errorf("EncodeHex(0X 前缀) = %q, %v; want %q", enc3, err, enc1)
	}
	// DecodeToHex 往返（含奇数长度 hex）
	if back, err := h.DecodeToHex(enc1); err != nil || back != hexID {
		t.Errorf("DecodeToHex = %q, %v; want %q", back, err, hexID)
	}
	h2 := NewHashids("", 0)
	if enc, err := h2.EncodeHex("abc"); err != nil {
		t.Fatalf("EncodeHex(abc) 失败: %v", err)
	} else if back, err := h2.DecodeToHex(enc); err != nil || back != "abc" {
		t.Errorf("DecodeToHex 奇数长度往返失败: %q, %v", back, err)
	}
	// 失败路径：空串、非法 hex、非法 hash
	if _, err := h2.EncodeHex(""); err == nil {
		t.Error("EncodeHex 空串应返回 error")
	}
	if _, err := h2.EncodeHex("zzzz"); err == nil {
		t.Error("EncodeHex 非法 hex 应返回 error")
	}
	if _, err := h2.DecodeToHex("!!!!"); err == nil {
		t.Error("DecodeToHex 非法 hash 应返回 error")
	}
}

func TestHashidsNewHashidsAlphabet(t *testing.T) {
	// 默认字母表（空串入参走 hutool 默认字母表）：构造成功且可逆
	h := newHashids("", 0, "")
	if h == nil {
		t.Fatal("newHashids 默认字母表构造失败")
	}
	enc, err := h.Encode(1, 2, 3)
	if err != nil {
		t.Fatalf("newHashids 默认字母表 Encode 失败: %v", err)
	}
	if back, err := h.Decode(enc); err != nil || len(back) != 3 || back[0] != 1 || back[2] != 3 {
		t.Errorf("newHashids 默认字母表往返失败: %v, %v", back, err)
	}
	// 自定义字母表非法（去重后少于 16 个字符）应 panic
	assertPanic := func(name string, fn func()) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("newHashids %s 应 panic", name)
			}
		}()
		fn()
	}
	assertPanic("过短字母表", func() { _ = newHashids("salt", 0, "abc") })
	// 自定义字母表含空格应 panic（对照 hutool IllegalArgumentException）
	assertPanic("含空格字母表", func() { _ = newHashids("salt", 0, "abcdefghijklmnop qrs") })
}
