package utils

import (
	"math"
	"math/big"
	"testing"
)

/* ---------------- 原有测试（自 utils_test.go 原样迁入） ---------------- */

func TestNumber(t *testing.T) {
	// big.Rat 精确舍入：0.615 的二进制误差不能影响结果
	if got := Number.Round(0.615, 2); got != 0.62 {
		t.Errorf("Round(0.615,2) = %v, want 0.62", got)
	}
	if got := Number.Add(0.1, 0.2); got != 0.3 {
		t.Errorf("Add(0.1,0.2) = %v, want 0.3", got)
	}
	if got := Number.DecimalFormatMoney(355565.115); got != "355,565.12" {
		t.Errorf("DecimalFormatMoney = %q, want \"355,565.12\"", got)
	}
	if got := Number.FormatPercent(0.5, 2); got != "50.00%" {
		t.Errorf("FormatPercent = %q, want \"50.00%%\"", got)
	}
	if !Number.IsNumber("0x1F") || Number.IsNumber("abc") {
		t.Error("IsNumber 判断错误")
	}
	if got := Number.ParseInt("12a", -1); got != -1 {
		t.Errorf("ParseInt 非法应返回默认值, got %d", got)
	}
	if got := Number.Divisor(12, 8); got != 4 {
		t.Errorf("Divisor(12,8) = %d, want 4", got)
	}
	if got := Number.Multiple(4, 6); got != 12 {
		t.Errorf("Multiple(4,6) = %d, want 12", got)
	}
	if !Number.IsPrimes(17) || Number.IsPrimes(18) {
		t.Error("IsPrimes 判断错误")
	}
	if got := Number.CeilDiv(10, 3); got != 4 {
		t.Errorf("CeilDiv(10,3) = %d, want 4", got)
	}
	if got := Number.DigitToChinese(1024.53); got != "壹仟零贰拾肆元伍角叁分" {
		t.Errorf("DigitToChinese(1024.53) = %q", got)
	}
	if got := Number.ChineseToNumber("一百二十三"); got != 123 {
		t.Errorf("ChineseToNumber = %d, want 123", got)
	}
	if got := Number.ChineseMoneyToNumber("壹佰贰拾元伍角"); got != 120.5 {
		t.Errorf("ChineseMoneyToNumber = %v, want 120.5", got)
	}
	if got := Number.ToStr(3.1400); got != "3.14" {
		t.Errorf("ToStr = %q", got)
	}
}

/* ---------------- 精确四则运算 ---------------- */

func TestNumberAddSubMul(t *testing.T) {
	if got := Number.Add(0.1, 0.2, 0.3); got != 0.6 {
		t.Errorf("Add(0.1,0.2,0.3) = %v, want 0.6", got)
	}
	if got := Number.Add(); got != 0 {
		t.Errorf("Add() 无参数应返回 0, got %v", got)
	}
	if got := Number.Sub(1.0, 0.9, 0.1); got != 0 {
		t.Errorf("Sub(1.0,0.9,0.1) = %v, want 0", got)
	}
	if got := Number.Sub(5); got != 5 {
		t.Errorf("Sub(5) 无减数应返回原值, got %v", got)
	}
	if got := Number.Mul(1.1, 3); got != 3.3 {
		t.Errorf("Mul(1.1,3) = %v, want 3.3", got)
	}
	if got := Number.Mul(7); got != 7 {
		t.Errorf("Mul(7) 无乘数应返回原值, got %v", got)
	}
}

func TestNumberDiv(t *testing.T) {
	if got := Number.Div(1, 3, 2); got != 0.33 {
		t.Errorf("Div(1,3,2) = %v, want 0.33", got)
	}
	if got := Number.Div(10, 3, 2); got != 3.33 {
		t.Errorf("Div(10,3,2) = %v, want 3.33", got)
	}
	// scale=0 按 HALF_UP 取整：10/4=2.5 恰为 .5，远离零进一得到 3
	if got := Number.Div(10, 4, 0); got != 3 {
		t.Errorf("Div(10,4,0) = %v, want 3（HALF_UP 取整）", got)
	}
	// 精确商不受舍入影响：scale 足够或保留完整精度时得到 2.5
	if got := Number.Div(10, 4, 2); got != 2.5 {
		t.Errorf("Div(10,4,2) = %v, want 2.5", got)
	}
	if got := Number.Div(10, 4, -1); got != 2.5 {
		t.Errorf("Div(10,4,-1) = %v, want 2.5（scale<0 保留完整精度）", got)
	}
	// 除零返回 0
	if got := Number.Div(10, 0, 2); got != 0 {
		t.Errorf("Div(10,0,2) = %v, want 0（除零保护）", got)
	}
	// scale<0 完整精度等价于 float64 除法
	if got := Number.Div(10, 3, -1); got != 10.0/3.0 {
		t.Errorf("Div(10,3,-1) = %v, want %v", got, 10.0/3.0)
	}
}

func TestNumberCeilDiv(t *testing.T) {
	if got := Number.CeilDiv(7, 2); got != 4 {
		t.Errorf("CeilDiv(7,2) = %d, want 4", got)
	}
	if got := Number.CeilDiv(-7, 2); got != -3 {
		t.Errorf("CeilDiv(-7,2) = %d, want -3（朝正无穷取整）", got)
	}
	if got := Number.CeilDiv(8, 2); got != 4 {
		t.Errorf("CeilDiv(8,2) 整除 = %d, want 4", got)
	}
	if got := Number.CeilDiv(10, 0); got != 0 {
		t.Errorf("CeilDiv(10,0) = %d, want 0（除零保护）", got)
	}
}

/* ---------------- 舍入与格式化 ---------------- */

func TestNumberRound(t *testing.T) {
	// 2.675 的二进制近似不影响 big.Rat 精确舍入：HALF_UP 得 2.68
	if got := Number.Round(2.675, 2); got != 2.68 {
		t.Errorf("Round(2.675,2) = %v, want 2.68", got)
	}
	if got := Number.Round(-2.5, 0); got != -3 {
		t.Errorf("Round(-2.5,0) = %v, want -3（HALF_UP 远离零）", got)
	}
	// scale < 0 按 0 位处理
	if got := Number.Round(2.675, -2); got != 3 {
		t.Errorf("Round(2.675,-2) = %v, want 3（scale<0 按 0 位）", got)
	}
}

func TestNumberRoundStr(t *testing.T) {
	if got := Number.RoundStr(123.456789, 2); got != "123.46" {
		t.Errorf("RoundStr(123.456789,2) = %q, want \"123.46\"", got)
	}
	if got := Number.RoundStr(5, 2); got != "5.00" {
		t.Errorf("RoundStr(5,2) = %q, want \"5.00\"", got)
	}
}

func TestNumberRoundHalfEven(t *testing.T) {
	if got := Number.RoundHalfEven(2.5, 0); got != 2 {
		t.Errorf("RoundHalfEven(2.5,0) = %v, want 2（五前为偶舍去）", got)
	}
	if got := Number.RoundHalfEven(3.5, 0); got != 4 {
		t.Errorf("RoundHalfEven(3.5,0) = %v, want 4（五前为奇进一）", got)
	}
	if got := Number.RoundHalfEven(1.25, 1); got != 1.2 {
		t.Errorf("RoundHalfEven(1.25,1) = %v, want 1.2", got)
	}
	if got := Number.RoundHalfEven(1.35, 1); got != 1.4 {
		t.Errorf("RoundHalfEven(1.35,1) = %v, want 1.4", got)
	}
}

func TestNumberRoundDown(t *testing.T) {
	if got := Number.RoundDown(123.456789, 2); got != 123.45 {
		t.Errorf("RoundDown(123.456789,2) = %v, want 123.45", got)
	}
	if got := Number.RoundDown(-123.456, 2); got != -123.45 {
		t.Errorf("RoundDown(-123.456,2) = %v, want -123.45（朝零截断）", got)
	}
}

func TestNumberDecimalFormatMoney(t *testing.T) {
	if got := Number.DecimalFormatMoney(355565.115); got != "355,565.12" {
		t.Errorf("DecimalFormatMoney(355565.115) = %q", got)
	}
	if got := Number.DecimalFormatMoney(-1234.5); got != "-1,234.50" {
		t.Errorf("DecimalFormatMoney(-1234.5) = %q, want \"-1,234.50\"", got)
	}
	if got := Number.DecimalFormatMoney(0); got != "0.00" {
		t.Errorf("DecimalFormatMoney(0) = %q, want \"0.00\"", got)
	}
}

func TestNumberFormatPercent(t *testing.T) {
	if got := Number.FormatPercent(0.5, 2); got != "50.00%" {
		t.Errorf("FormatPercent(0.5,2) = %q, want \"50.00%%\"", got)
	}
	if got := Number.FormatPercent(0.12345, 1); got != "12.3%" {
		t.Errorf("FormatPercent(0.12345,1) = %q, want \"12.3%%\"", got)
	}
}

func TestNumberToStr(t *testing.T) {
	if got := Number.ToStr(3.14); got != "3.14" {
		t.Errorf("ToStr(3.14) = %q", got)
	}
	if got := Number.ToStr(5.0); got != "5" {
		t.Errorf("ToStr(5.0) = %q, want \"5\"（整数无小数点）", got)
	}
	if got := Number.ToStr(math.NaN()); got != "NaN" {
		t.Errorf("ToStr(NaN) = %q, want \"NaN\"", got)
	}
	if got := Number.ToStr(math.Inf(1)); got != "+Inf" {
		t.Errorf("ToStr(+Inf) = %q, want \"+Inf\"", got)
	}
}

/* ---------------- 校验与解析 ---------------- */

func TestNumberIsNumber(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"123", true}, {"1.5", true}, {"1e5", true}, {"1.2E+3", true},
		{"1e+5", true}, {"-1.2", true}, {"+.5", true}, {"5.", true},
		{"0x1F", true}, {"0XaB", true}, {"123L", true}, {"12f", true},
		{"", false}, {"  ", false}, {"abc", false}, {"12a", false},
		{"1.2.3", false}, {"1e5e6", false}, {"0x", false}, {".", false},
		{"1e", false}, {"1e+", false}, {"123Ld", false},
	}
	for _, c := range cases {
		if got := Number.IsNumber(c.in); got != c.want {
			t.Errorf("IsNumber(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestNumberIsInteger(t *testing.T) {
	for _, s := range []string{"123", "+123", "-123", "0x1A", "-0x10"} {
		if !Number.IsInteger(s) {
			t.Errorf("IsInteger(%q) 应为 true", s)
		}
	}
	for _, s := range []string{"1.5", "", "abc", "0x", "12L"} {
		if Number.IsInteger(s) {
			t.Errorf("IsInteger(%q) 应为 false", s)
		}
	}
}

func TestNumberIsDouble(t *testing.T) {
	for _, s := range []string{"1.5", "3.14", "-0.5", ".5", "1.2e3"} {
		if !Number.IsDouble(s) {
			t.Errorf("IsDouble(%q) 应为 true", s)
		}
	}
	// 无小数点的 "1e3" 按整数对待
	for _, s := range []string{"1e3", "123", "abc", ""} {
		if Number.IsDouble(s) {
			t.Errorf("IsDouble(%q) 应为 false", s)
		}
	}
}

func TestNumberParseInt(t *testing.T) {
	if got := Number.ParseInt("0x1A", -1); got != 26 {
		t.Errorf("ParseInt(\"0x1A\") = %d, want 26（十六进制）", got)
	}
	if got := Number.ParseInt("123.56", -1); got != 123 {
		t.Errorf("ParseInt(\"123.56\") = %d, want 123（截取整数部分）", got)
	}
	if got := Number.ParseInt("abc", 80); got != 80 {
		t.Errorf("ParseInt(\"abc\") = %d, want 80（非法回退默认值）", got)
	}
	if got := Number.ParseInt("  ", 80); got != 80 {
		t.Errorf("ParseInt 空白串 = %d, want 80", got)
	}
}

func TestNumberParseInt64(t *testing.T) {
	if got := Number.ParseInt64("9223372036854775807", 0); got != 9223372036854775807 {
		t.Errorf("ParseInt64 最大 int64 = %d", got)
	}
	if got := Number.ParseInt64("0x10", -1); got != 16 {
		t.Errorf("ParseInt64(\"0x10\") = %d, want 16", got)
	}
	if got := Number.ParseInt64("abc", 7); got != 7 {
		t.Errorf("ParseInt64(\"abc\") = %d, want 7（非法回退默认值）", got)
	}
}

func TestNumberParseFloat64(t *testing.T) {
	if got := Number.ParseFloat64("1.5e2", 0); got != 150 {
		t.Errorf("ParseFloat64(\"1.5e2\") = %v, want 150（科学计数法）", got)
	}
	if got := Number.ParseFloat64("abc", -1); got != -1 {
		t.Errorf("ParseFloat64(\"abc\") = %v, want -1（非法回退默认值）", got)
	}
	if got := Number.ParseFloat64("", 9.9); got != 9.9 {
		t.Errorf("ParseFloat64(\"\") = %v, want 9.9（空串回退默认值）", got)
	}
}

/* ---------------- 数学运算 ---------------- */

func TestNumberFactorial(t *testing.T) {
	if got := Number.Factorial(5); got.Cmp(big.NewInt(120)) != 0 {
		t.Errorf("Factorial(5) = %v, want 120", got)
	}
	if got := Number.Factorial(0); got.Cmp(big.NewInt(1)) != 0 {
		t.Errorf("Factorial(0) = %v, want 1", got)
	}
	if got := Number.Factorial(-1); got.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("Factorial(-1) = %v, want 0（负数约定返回 0）", got)
	}
}

func TestNumberDivisorMultiple(t *testing.T) {
	if got := Number.Divisor(12, 8); got != 4 {
		t.Errorf("Divisor(12,8) = %d, want 4", got)
	}
	if got := Number.Divisor(0, 8); got != 8 {
		t.Errorf("Divisor(0,8) = %d, want 8（gcd(0,n)=n）", got)
	}
	if got := Number.Divisor(8, 0); got != 8 {
		t.Errorf("Divisor(8,0) = %d, want 8", got)
	}
	if got := Number.Divisor(-12, 8); got != 4 {
		t.Errorf("Divisor(-12,8) = %d, want 4（负数取绝对值）", got)
	}
	if got := Number.Multiple(4, 6); got != 12 {
		t.Errorf("Multiple(4,6) = %d, want 12", got)
	}
	if got := Number.Multiple(0, 6); got != 0 {
		t.Errorf("Multiple(0,6) = %d, want 0", got)
	}
}

func TestNumberIsPrimes(t *testing.T) {
	cases := []struct {
		n    int
		want bool
	}{
		{0, false}, {1, false}, {2, true}, {3, true}, {4, false},
		{17, true}, {18, false}, {97, true}, {100, false},
	}
	for _, c := range cases {
		if got := Number.IsPrimes(c.n); got != c.want {
			t.Errorf("IsPrimes(%d) = %v, want %v", c.n, got, c.want)
		}
	}
}

func TestNumberPartValue(t *testing.T) {
	if got := Number.PartValue(10, 3, true); got != 4 {
		t.Errorf("PartValue(10,3,true) = %d, want 4（有余数 plusOne）", got)
	}
	if got := Number.PartValue(10, 3, false); got != 3 {
		t.Errorf("PartValue(10,3,false) = %d, want 3（有余数不进一）", got)
	}
	if got := Number.PartValue(9, 3, true); got != 3 {
		t.Errorf("PartValue(9,3,true) = %d, want 3（整除不进一）", got)
	}
	if got := Number.PartValue(10, 0, true); got != 0 {
		t.Errorf("PartValue(10,0,true) = %d, want 0（partCount<=0）", got)
	}
	if got := Number.PartValue(10, -2, false); got != 0 {
		t.Errorf("PartValue(10,-2,false) = %d, want 0（负份数）", got)
	}
}

func TestNumberIsPowerOfTwo(t *testing.T) {
	for _, n := range []int64{1, 2, 1024} {
		if !Number.IsPowerOfTwo(n) {
			t.Errorf("IsPowerOfTwo(%d) 应为 true", n)
		}
	}
	for _, n := range []int64{1000, 0, -2} {
		if Number.IsPowerOfTwo(n) {
			t.Errorf("IsPowerOfTwo(%d) 应为 false", n)
		}
	}
}

func TestNumberIsOddEven(t *testing.T) {
	cases := []struct {
		n         int64
		odd, even bool
	}{
		{-3, true, false}, {-4, false, true}, {0, false, true},
		{7, true, false}, {8, false, true},
	}
	for _, c := range cases {
		if got := Number.IsOdd(c.n); got != c.odd {
			t.Errorf("IsOdd(%d) = %v, want %v", c.n, got, c.odd)
		}
		if got := Number.IsEven(c.n); got != c.even {
			t.Errorf("IsEven(%d) = %v, want %v", c.n, got, c.even)
		}
	}
}

/* ---------------- 中文数字 ---------------- */

func TestNumberDigitToChinese(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{1024.53, "壹仟零贰拾肆元伍角叁分"},
		{0, "零元整"},
		{-123.45, "负壹佰贰拾叁元肆角伍分"},
		{1024, "壹仟零贰拾肆元整"},                // 无小数以"整"结尾，节内补零
		{12345000000000, "壹拾贰万叁仟肆佰伍拾亿元整"}, // 万亿级
		{100000000000000, ""},             // 超出 ±99999999999999.99 返回空串
		{0.05, "伍分"},                      // 元角均为 0 时仅输出分
		{0.53, "伍角叁分"},
		{0.5, "伍角"},
		{100.5, "壹佰元伍角"},
		{105.1, "壹佰零伍元壹角"},
		{100000000.01, "壹亿元零壹分"},
		{1000000000000, "壹万亿元整"},
	}
	for _, c := range cases {
		if got := Number.DigitToChinese(c.in); got != c.want {
			t.Errorf("DigitToChinese(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNumberChineseToNumber(t *testing.T) {
	cases := []struct {
		in   string
		want int64
	}{
		{"一百二十三", 123},
		{"十", 10},  // 单位开头默认置 1
		{"二十", 20}, // 末尾无数字
		{"零一", 1},  // 含"零"：节结束，未决权位失效
		{"两万二", 22000},
		{"十二亿三千四百万", 1234000000},
		{"一千零一", 1001},
		{"一百", 100},
		{"壹佰贰拾", 120}, // 繁体数字
		{"一二三", 0},    // 连续数字非法
		{"abc", 0},    // 非法字符
		{"", 0},       // 空串
		{"零零", 0},
	}
	for _, c := range cases {
		if got := Number.ChineseToNumber(c.in); got != c.want {
			t.Errorf("ChineseToNumber(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestNumberChineseMoneyToNumber(t *testing.T) {
	cases := []struct {
		in   string
		want float64
	}{
		{"壹佰贰拾元伍角", 120.5},
		{"壹元整", 1},
		{"叁角贰分", 0.32},
		{"陆万柒仟伍佰伍拾陆元叁角贰分", 67556.32},
		{"壹圆伍角", 1.5},   // "圆" 与 "元" 等价
		{"拾元叁分", 10.03}, // 无角有分
		{"", 0},
		{"非法", 0},
	}
	for _, c := range cases {
		if got := Number.ChineseMoneyToNumber(c.in); got != c.want {
			t.Errorf("ChineseMoneyToNumber(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
