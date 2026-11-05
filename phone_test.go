package utils

import "testing"

// phone_test.go：Phone 电话工具逐方法全覆盖。
// 黄金值来源：hutool 官方测试 PhoneUtilTest.java；港/台/澳号码按源码正则构造，
// 并已用临时程序验证后固化。

/* ---------------- 号码校验 ---------------- */

func TestPhoneIsMobile(t *testing.T) {
	// hutool PhoneUtilTest.testCheck
	if !Phone.IsMobile("13612345678") {
		t.Error("IsMobile(13612345678) 应为 true")
	}
	if Phone.IsMobile("136123456781") {
		t.Error("IsMobile(136123456781) 位数多一位应为 false")
	}
	if Phone.IsMobile("12345678901") {
		t.Error("IsMobile(12345678901) 应为 false")
	}
}

func TestPhoneIsMobileHk(t *testing.T) {
	if !Phone.IsMobileHk("51004810") {
		t.Error("IsMobileHk(51004810) 应为 true")
	}
	if !Phone.IsMobileHk("+85251004810") {
		t.Error("IsMobileHk(+852 前缀) 应为 true")
	}
	if Phone.IsMobileHk("1234567") {
		t.Error("IsMobileHk(7 位) 应为 false")
	}
}

func TestPhoneIsMobileTw(t *testing.T) {
	if !Phone.IsMobileTw("0960000000") {
		t.Error("IsMobileTw(0960000000) 应为 true")
	}
	if !Phone.IsMobileTw("886-0960000000") {
		t.Error("IsMobileTw(886- 前缀) 应为 true")
	}
	if Phone.IsMobileTw("096000000") {
		t.Error("IsMobileTw(9 位) 应为 false")
	}
}

func TestPhoneIsMobileMo(t *testing.T) {
	if !Phone.IsMobileMo("68000000") {
		t.Error("IsMobileMo(68000000) 应为 true")
	}
	if !Phone.IsMobileMo("+853-68000000") {
		t.Error("IsMobileMo(+853- 前缀) 应为 true")
	}
	if Phone.IsMobileMo("58000000") {
		t.Error("IsMobileMo(5 开头) 应为 false")
	}
}

func TestPhoneIsTel(t *testing.T) {
	// 正例（hutool PhoneUtilTest.testTel / testNewTel）
	for _, tel := range []string{
		"010-88993108", "010-12345678", "01088889999", "01012345678",
		"020-9999999", "0209999999", "0571-8888888", "0755-7654321", "07557654321",
	} {
		if !Phone.IsTel(tel) {
			t.Errorf("IsTel(%q) 应为 true", tel)
		}
	}
	// 反例
	for _, tel := range []string{
		"010 12345678", "A20-9999999", "0755-7654.321", "13619887123", "010-889931081",
	} {
		if Phone.IsTel(tel) {
			t.Errorf("IsTel(%q) 应为 false", tel)
		}
	}
}

func TestPhoneIsTel400800(t *testing.T) {
	// hutool PhoneUtilTest.isTel400800Test
	if !Phone.IsTel400800("400-860-8608") {
		t.Error("IsTel400800(400-860-8608) 应为 true")
	}
	if !Phone.IsTel400800("400-8608608") {
		t.Error("IsTel400800(400-8608608) 应为 true")
	}
	if !Phone.IsTel400800("8009996555") {
		t.Error("IsTel400800(8009996555) 应为 true")
	}
	if !Phone.IsTel400800("800-830-3811") {
		t.Error("IsTel400800(800-830-3811) 应为 true")
	}
	if Phone.IsTel400800("500-8608608") {
		t.Error("IsTel400800(500 开头) 应为 false")
	}
}

func TestPhoneIsPhone(t *testing.T) {
	// 手机号与座机、400/800、港澳台任一匹配即合法（hutool isPhone）
	for _, s := range []string{"13612345678", "010-88993108", "51004810", "0960000000", "68000000", "400-8608608"} {
		if !Phone.IsPhone(s) {
			t.Errorf("IsPhone(%q) 应为 true", s)
		}
	}
	for _, s := range []string{"136123456781", "010-889931081"} {
		if Phone.IsPhone(s) {
			t.Errorf("IsPhone(%q) 应为 false", s)
		}
	}
}

/* ---------------- 号码打码 ---------------- */

func TestPhoneHide(t *testing.T) {
	// hutool PhoneUtilTest.testHide
	const mobile = "13612345678"
	if got := Phone.HideBefore(mobile); got != "*******5678" {
		t.Errorf("HideBefore = %q, want *******5678", got)
	}
	if got := Phone.HideBetween(mobile); got != "136****5678" {
		t.Errorf("HideBetween = %q, want 136****5678", got)
	}
	if got := Phone.HideAfter(mobile); got != "1361234****" {
		t.Errorf("HideAfter = %q, want 1361234****", got)
	}
}

/* ---------------- 号段截取 ---------------- */

func TestPhoneSub(t *testing.T) {
	// hutool PhoneUtilTest.testSubString
	const mobile = "13612345678"
	if got := Phone.SubBefore(mobile); got != "136" {
		t.Errorf("SubBefore = %q, want 136", got)
	}
	if got := Phone.SubBetween(mobile); got != "1234" {
		t.Errorf("SubBetween = %q, want 1234", got)
	}
	if got := Phone.SubAfter(mobile); got != "5678" {
		t.Errorf("SubAfter = %q, want 5678", got)
	}
}

func TestPhoneSubTel(t *testing.T) {
	// hutool PhoneUtilTest.testNewTel
	cases := []struct {
		tel        string
		wantBefore string
		wantAfter  string
	}{
		{"010-12345678", "010", "12345678"},
		{"01012345678", "010", "12345678"},
		{"0755-7654321", "0755", "7654321"},
		{"07557654321", "0755", "7654321"},
		{"0571-88888888", "0571", "88888888"},
	}
	for _, c := range cases {
		if got := Phone.SubTelBefore(c.tel); got != c.wantBefore {
			t.Errorf("SubTelBefore(%q) = %q, want %q", c.tel, got, c.wantBefore)
		}
		if got := Phone.SubTelAfter(c.tel); got != c.wantAfter {
			t.Errorf("SubTelAfter(%q) = %q, want %q", c.tel, got, c.wantAfter)
		}
	}
	// 非座机格式返回空串
	if got := Phone.SubTelBefore("13612345678"); got != "" {
		t.Errorf("SubTelBefore(手机号) = %q, want 空串", got)
	}
	if got := Phone.SubTelAfter("13612345678"); got != "" {
		t.Errorf("SubTelAfter(手机号) = %q, want 空串", got)
	}
}
