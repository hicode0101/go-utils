package utils

import (
	"math"
	"testing"
)

/* ---------------- 编码与解码 ---------------- */

// TestHexEncodeDecode 字节与 hex 互转：往返（含中文、空切片）、大写编码、错误路径。
func TestHexEncodeDecode(t *testing.T) {
	// 往返：ASCII
	if got := Hex.Encode([]byte("hello")); got != "68656c6c6f" {
		t.Errorf("Encode(hello) = %q, want \"68656c6c6f\"", got)
	}
	back, err := Hex.Decode("68656c6c6f")
	if err != nil || string(back) != "hello" {
		t.Errorf("Decode 往返失败: %q, %v", back, err)
	}
	// 往返：中文 UTF-8
	if got := Hex.Encode([]byte("你好")); got != "e4bda0e5a5bd" {
		t.Errorf("Encode(你好) = %q", got)
	}
	if cn, err := Hex.Decode("e4bda0e5a5bd"); err != nil || string(cn) != "你好" {
		t.Errorf("Decode 中文往返失败: %q, %v", cn, err)
	}
	// 空切片
	if got := Hex.Encode([]byte{}); got != "" {
		t.Errorf("Encode(空切片) = %q, want \"\"", got)
	}
	if empty, err := Hex.Decode(""); err != nil || len(empty) != 0 {
		t.Errorf("Decode(\"\") = %q, %v; want 空切片无错误", empty, err)
	}
	// 大写编码
	if got := Hex.EncodeUpper([]byte{0x01, 0xAB}); got != "01AB" {
		t.Errorf("EncodeUpper = %q, want \"01AB\"", got)
	}
	// 解码大小写均可
	if up, err := Hex.Decode("01AB"); err != nil || len(up) != 2 || up[0] != 0x01 || up[1] != 0xAB {
		t.Errorf("Decode 大写输入失败: %q, %v", up, err)
	}
	// 错误路径：奇数长度（原 utils_test.go TestHashHex 断言迁入）
	if _, err := Hex.Decode("abc"); err == nil {
		t.Error("奇数长度 hex 应返回错误")
	}
	// 错误路径：非法字符
	if _, err := Hex.Decode("zz"); err == nil {
		t.Error("非法字符 hex 应返回错误")
	}
}

// TestHexStrRoundtrip 字符串与 hex 互转：EncodeStr/DecodeStr 往返 + 错误路径。
func TestHexStrRoundtrip(t *testing.T) {
	// 原 utils_test.go TestHashHex 断言迁入
	if got := Hex.EncodeStr("你好"); got != "e4bda0e5a5bd" {
		t.Errorf("Hex.EncodeStr(你好) = %q", got)
	}
	if back, err := Hex.DecodeStr("e4bda0e5a5bd"); err != nil || back != "你好" {
		t.Errorf("Hex.DecodeStr 往返失败: %q, %v", back, err)
	}
	// 中文含标点与 ASCII 混排
	enc := Hex.EncodeStr("中文，test！123")
	if back, err := Hex.DecodeStr(enc); err != nil || back != "中文，test！123" {
		t.Errorf("DecodeStr 混排往返失败: %q, %v", back, err)
	}
	// 错误路径：奇数长度
	if _, err := Hex.DecodeStr("abc"); err == nil {
		t.Error("DecodeStr 奇数长度应返回错误")
	}
	// 错误路径：非法字符
	if _, err := Hex.DecodeStr("zz"); err == nil {
		t.Error("DecodeStr 非法字符应返回错误")
	}
}

// TestHexIsHexNumber 十六进制数判断：前缀、无前缀、大小写与反例。
func TestHexIsHexNumber(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"0x1F", true},  // 0x 前缀
		{"0X1f", true},  // 0X 前缀 + 小写
		{"#FF", true},   // # 前缀
		{"1F", true},    // 无前缀（按源码：body 直接按 hex 字符校验）
		{"002c", true},  // hutool 官方用例
		{"0xG1", false}, // 含非 hex 字符
		{"-11", false},  // 负号开头
		{"", false},     // 空串
		{"0x", false},   // 仅前缀无 body
		{"#", false},    // 仅 # 无 body
	}
	for _, c := range cases {
		if got := Hex.IsHexNumber(c.in); got != c.want {
			t.Errorf("IsHexNumber(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

/* ---------------- 数值转换 ---------------- */

// TestHexNumberConversions 字符/整数与 hex 互转 + HexToInt 前缀兼容与错误路径。
func TestHexNumberConversions(t *testing.T) {
	// 字符转 4 位 hex（不足 4 位前补 0）
	if got := Hex.ToUnicodeHex('中'); got != "4e2d" {
		t.Errorf("ToUnicodeHex('中') = %q, want \"4e2d\"", got)
	}
	if got := Hex.ToUnicodeHex('A'); got != "0041" {
		t.Errorf("ToUnicodeHex('A') = %q, want \"0041\"", got)
	}
	// 整数转 hex
	if got := Hex.ToHex(255); got != "ff" {
		t.Errorf("ToHex(255) = %q, want \"ff\"", got)
	}
	if got := Hex.ToHex(0); got != "0" {
		t.Errorf("ToHex(0) = %q, want \"0\"", got)
	}
	if got := Hex.ToHex(-1); got != "ffffffffffffffff" { // 负数按补码（Java Long.toHexString 语义）
		t.Errorf("ToHex(-1) = %q", got)
	}
	// HexToInt 前缀兼容（hutool 官方用例：FF / 0xFF / #FF 均为 255）
	for _, s := range []string{"FF", "0xFF", "#FF", "0Xff"} {
		if got, err := Hex.HexToInt(s); err != nil || got != 255 {
			t.Errorf("HexToInt(%q) = %d, %v; want 255", s, got, err)
		}
	}
	// 往返
	if got, err := Hex.HexToInt(Hex.ToHex(123456789)); err != nil || got != 123456789 {
		t.Errorf("HexToInt 往返失败: %d, %v", got, err)
	}
	// 错误路径：非法字符
	if _, err := Hex.HexToInt("xyz"); err == nil {
		t.Error("HexToInt 非法输入应返回错误")
	}
	// 错误路径：空串
	if _, err := Hex.HexToInt(""); err == nil {
		t.Error("HexToInt 空串应返回错误")
	}
	// 错误路径：超出 int64 范围
	if _, err := Hex.HexToInt("0x10000000000000000"); err == nil {
		t.Error("HexToInt 超出 int64 范围应返回错误")
	}
}

// TestHexFloat IEEE754 位模式与浮点还原：以 math.Float32bits/Float64bits 构造已知值。
func TestHexFloat(t *testing.T) {
	// float32(1.5) 位模式 0x3FC00000
	if got := Hex.ToHex(int64(math.Float32bits(1.5))); got != "3fc00000" {
		t.Errorf("float32(1.5) 位模式 hex = %q, want \"3fc00000\"", got)
	}
	if got, err := Hex.HexToFloat32("3fc00000"); err != nil || got != 1.5 {
		t.Errorf("HexToFloat32(3fc00000) = %v, %v; want 1.5", got, err)
	}
	// 带 0x / # 前缀（hutool 官方用例）
	if got, err := Hex.HexToFloat32("0x3fc00000"); err != nil || got != 1.5 {
		t.Errorf("HexToFloat32(0x 前缀) = %v, %v", got, err)
	}
	if got, err := Hex.HexToFloat32("#3fc00000"); err != nil || got != 1.5 {
		t.Errorf("HexToFloat32(# 前缀) = %v, %v", got, err)
	}
	// 负数：float32(-1.5) 位模式 0xBFC00000
	if got, err := Hex.HexToFloat32("bfc00000"); err != nil || got != -1.5 {
		t.Errorf("HexToFloat32(bfc00000) = %v, %v; want -1.5", got, err)
	}
	// 任意值往返（位模式还原应精确相等）
	for _, v := range []float32{0.25, -1.5, 123.456} {
		back, err := Hex.HexToFloat32(Hex.ToHex(int64(math.Float32bits(v))))
		if err != nil || back != v {
			t.Errorf("HexToFloat32 往返失败: %v -> %v, %v", v, back, err)
		}
	}
	// float64：1.5 位模式 0x3FF8000000000000
	if got, err := Hex.HexToFloat64("3ff8000000000000"); err != nil || got != 1.5 {
		t.Errorf("HexToFloat64(3ff8000000000000) = %v, %v; want 1.5", got, err)
	}
	if got, err := Hex.HexToFloat64(Hex.ToHex(int64(math.Float64bits(math.Pi)))); err != nil || got != math.Pi {
		t.Errorf("HexToFloat64 往返失败: %v, %v", got, err)
	}
	// 错误路径：非法字符
	if _, err := Hex.HexToFloat32("zz"); err == nil {
		t.Error("HexToFloat32 非法输入应返回错误")
	}
	if _, err := Hex.HexToFloat64("xyz"); err == nil {
		t.Error("HexToFloat64 非法输入应返回错误")
	}
}

/* ---------------- 格式化与颜色 ---------------- */

// TestHexFormat hex 分组格式化：偶数长度、奇数长度前补 0、空串。
func TestHexFormat(t *testing.T) {
	// 原 utils_test.go TestHashHex 断言迁入
	if got := Hex.Format("010203"); got != "01 02 03" {
		t.Errorf("Hex.Format = %q, want \"01 02 03\"", got)
	}
	// 奇数长度前补 0（hutool format 语义）
	if got := Hex.Format("123"); got != "01 23" {
		t.Errorf("Hex.Format(123) = %q, want \"01 23\"", got)
	}
	if got := Hex.Format("1"); got != "01" {
		t.Errorf("Hex.Format(1) = %q, want \"01\"", got)
	}
	// 空串原样
	if got := Hex.Format(""); got != "" {
		t.Errorf("Hex.Format(\"\") = %q, want \"\"", got)
	}
}

// TestHexColor 颜色值编解码：往返、无 # 前缀、错误路径。
func TestHexColor(t *testing.T) {
	// 原 utils_test.go TestHashHex 断言迁入
	if got := Hex.EncodeColor(255, 0, 0); got != "#FF0000" {
		t.Errorf("EncodeColor = %q", got)
	}
	r, g, b, err := Hex.DecodeColor("#FF0000")
	if err != nil || r != 255 || g != 0 || b != 0 {
		t.Errorf("DecodeColor = %d,%d,%d,%v", r, g, b, err)
	}
	if got := Hex.EncodeColor(255, 255, 255); got != "#FFFFFF" {
		t.Errorf("EncodeColor(白) = %q", got)
	}
	// 往返
	cr, cg, cb := uint8(0x12), uint8(0x34), uint8(0xAB)
	rr, gg, bb, err := Hex.DecodeColor(Hex.EncodeColor(cr, cg, cb))
	if err != nil || rr != cr || gg != cg || bb != cb {
		t.Errorf("颜色往返失败: %d,%d,%d,%v", rr, gg, bb, err)
	}
	// 无 # 前缀
	if r, g, b, err := Hex.DecodeColor("FF8000"); err != nil || r != 255 || g != 128 || b != 0 {
		t.Errorf("DecodeColor 无前缀 = %d,%d,%d,%v", r, g, b, err)
	}
	// 错误路径：长度不为 6
	if _, _, _, err := Hex.DecodeColor("#FFF"); err == nil {
		t.Error("DecodeColor 长度非法应返回错误")
	}
	// 错误路径：含非 hex 字符
	if _, _, _, err := Hex.DecodeColor("#GG0000"); err == nil {
		t.Error("DecodeColor 非法字符应返回错误")
	}
}
