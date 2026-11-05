package utils

import "testing"

/* ---------------- 迁移自 utils_test.go 的 TestBoolChar（Bool 部分，原样保留） ---------------- */

func TestBoolLegacy(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"yes", true}, {"是", true}, {"ok", true}, {"1", true}, {"ON", true},
		{"no", false}, {"", false}, {"随便", false},
	}
	for _, c := range cases {
		if got := Bool.ToBoolExt(c.in); got != c.want {
			t.Errorf("ToBoolExt(%q) = %v, want %v", c.in, got, c.want)
		}
	}
	if !Bool.Xor(true, false, false) || Bool.Xor(true, true) {
		t.Error("Xor 奇偶语义错误")
	}
	if got := Bool.ToStringYesNo(true); got != "yes" {
		t.Errorf("ToStringYesNo = %q", got)
	}
}

/* ---------------- ToBoolExt（表驱动全覆盖） ---------------- */

func TestBoolToBoolExt(t *testing.T) {
	trueCases := []string{
		// 全部真值写法
		"true", "yes", "y", "t", "ok", "correct", "success", "on", "1",
		"是", "对", "真", "對", "正确", "开", "开启", "√", "☑",
		// 大写变体（忽略大小写）
		"TRUE", "Yes", "OK", "On", "Success", "CORRECT",
		// 带首尾空白变体
		"  true  ", "\t是\n", " yes", "on ", "\n 1 ",
	}
	for _, in := range trueCases {
		if got := Bool.ToBoolExt(in); !got {
			t.Errorf("ToBoolExt(%q) = false, want true", in)
		}
	}
	falseCases := []string{
		"no", "", "   ", "0", "off", "否", "false", "随便", "乱串#%&",
		"yes1", "truetype", "ok1", "\u00a0", "对错",
	}
	for _, in := range falseCases {
		if got := Bool.ToBoolExt(in); got {
			t.Errorf("ToBoolExt(%q) = true, want false", in)
		}
	}
}

/* ---------------- Negate ---------------- */

func TestBoolNegate(t *testing.T) {
	if Bool.Negate(true) {
		t.Error("Negate(true) 应为 false")
	}
	if !Bool.Negate(false) {
		t.Error("Negate(false) 应为 true")
	}
}

/* ---------------- And / Or / Xor（含无参、全 true、奇偶） ---------------- */

func TestBoolAndOrXor(t *testing.T) {
	// And：全部为 true 才为 true
	if Bool.And() {
		t.Error("And 无参应为 false")
	}
	if !Bool.And(true) || !Bool.And(true, true, true) {
		t.Error("And 全 true 应为 true")
	}
	if Bool.And(true, false, true) || Bool.And(false) {
		t.Error("And 含 false 应为 false")
	}
	// Or：任一为 true 即为 true
	if Bool.Or() {
		t.Error("Or 无参应为 false")
	}
	if Bool.Or(false, false) {
		t.Error("Or 全 false 应为 false")
	}
	if !Bool.Or(false, true) || !Bool.Or(true, true, true) {
		t.Error("Or 含 true 应为 true")
	}
	// Xor：true 的个数为奇数时为 true
	if Bool.Xor() {
		t.Error("Xor 无参应为 false")
	}
	if !Bool.Xor(true) || !Bool.Xor(true, false, false) || !Bool.Xor(true, true, true) {
		t.Error("Xor 奇数个 true 应为 true")
	}
	if Bool.Xor(true, true) || Bool.Xor(false, false, false) || Bool.Xor(true, true, false, false) {
		t.Error("Xor 偶数个 true 应为 false")
	}
}

/* ---------------- ToStringTrueFalse / ToStringOnOff / ToStringYesNo ---------------- */

func TestBoolToString(t *testing.T) {
	if got := Bool.ToStringTrueFalse(true); got != "true" {
		t.Errorf("ToStringTrueFalse(true) = %q, want \"true\"", got)
	}
	if got := Bool.ToStringTrueFalse(false); got != "false" {
		t.Errorf("ToStringTrueFalse(false) = %q, want \"false\"", got)
	}
	if got := Bool.ToStringOnOff(true); got != "on" {
		t.Errorf("ToStringOnOff(true) = %q, want \"on\"", got)
	}
	if got := Bool.ToStringOnOff(false); got != "off" {
		t.Errorf("ToStringOnOff(false) = %q, want \"off\"", got)
	}
	if got := Bool.ToStringYesNo(true); got != "yes" {
		t.Errorf("ToStringYesNo(true) = %q, want \"yes\"", got)
	}
	if got := Bool.ToStringYesNo(false); got != "no" {
		t.Errorf("ToStringYesNo(false) = %q, want \"no\"", got)
	}
}
