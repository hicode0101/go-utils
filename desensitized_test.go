package utils

import (
	"strings"
	"testing"
)

// desensitized_test.go：Desensitized 脱敏工具逐方法全覆盖。
// 黄金值来源：hutool 官方测试 DesensitizedUtilTest.java；无 hutool 用例的值
// 已用临时程序验证后固化。

/* ---------------- 专项脱敏 ---------------- */

func TestDesensitizedChineseName(t *testing.T) {
	// hutool DesensitizedUtilTest.chineseNameTest / desensitizedTest
	if got := Desensitized.ChineseName("段正淳"); got != "段**" {
		t.Errorf("ChineseName(段正淳) = %q, want 段**", got)
	}
	// 2 字姓名
	if got := Desensitized.ChineseName("张三"); got != "张*" {
		t.Errorf("ChineseName(张三) = %q, want 张*", got)
	}
}

func TestDesensitizedFirstMask(t *testing.T) {
	// hutool first_mask：FirstMask("123456789") = "1********"
	if got := Desensitized.FirstMask("123456789"); got != "1********" {
		t.Errorf("FirstMask(123456789) = %q, want 1********", got)
	}
	if got := Desensitized.FirstMask("a"); got != "a" {
		t.Errorf("FirstMask(单字符) = %q, want 原样", got)
	}
	if got := Desensitized.FirstMask(""); got != "" {
		t.Errorf("FirstMask(空串) = %q, want 空串", got)
	}
}

func TestDesensitizedIdCardNum(t *testing.T) {
	// hutool DesensitizedUtilTest.idCardNumTest：保留前 1 后 2
	if got := Desensitized.IdCardNum("51343620000320711X", 1, 2); got != "5***************1X" {
		t.Errorf("IdCardNum = %q, want 5***************1X", got)
	}
	// front+end 超过长度返回空串
	if got := Desensitized.IdCardNum("12345", 3, 3); got != "" {
		t.Errorf("IdCardNum(超长保留) = %q, want 空串", got)
	}
	// 负数保留位返回空串
	if got := Desensitized.IdCardNum("12345", -1, 2); got != "" {
		t.Errorf("IdCardNum(负数) = %q, want 空串", got)
	}
}

func TestDesensitizedFixedPhone(t *testing.T) {
	// hutool DesensitizedUtilTest.fixedPhoneTest
	if got := Desensitized.FixedPhone("09157518479"); got != "0915*****79" {
		t.Errorf("FixedPhone = %q, want 0915*****79", got)
	}
}

func TestDesensitizedMobilePhone(t *testing.T) {
	// hutool DesensitizedUtilTest.mobilePhoneTest
	if got := Desensitized.MobilePhone("18049531999"); got != "180****1999" {
		t.Errorf("MobilePhone(18049531999) = %q, want 180****1999", got)
	}
	if got := Desensitized.MobilePhone("13812345678"); got != "138****5678" {
		t.Errorf("MobilePhone(13812345678) = %q, want 138****5678", got)
	}
}

func TestDesensitizedAddress(t *testing.T) {
	// hutool DesensitizedUtilTest.addressTest
	const addr = "北京市海淀区马连洼街道289号"
	cases := []struct {
		size int
		want string
	}{
		{8, "北京市海淀区马********"},
		{5, "北京市海淀区马连洼街*****"},
		{50, "***************"},
		{0, addr},
		{-1, addr}, // 打码区间非法时原样返回
	}
	for _, c := range cases {
		if got := Desensitized.Address(addr, c.size); got != c.want {
			t.Errorf("Address(sensitiveSize=%d) = %q, want %q", c.size, got, c.want)
		}
	}
}

func TestDesensitizedEmail(t *testing.T) {
	// hutool DesensitizedUtilTest.emailTest
	cases := []struct {
		email string
		want  string
	}{
		{"duandazhi@126.com", "d********@126.com"},
		{"duandazhi@gmail.com.cn", "d********@gmail.com.cn"},
		{"duandazhi-jack@gmail.com.cn", "d*************@gmail.com.cn"},
		{"zhangsan@example.com", "z*******@example.com"},
	}
	for _, c := range cases {
		if got := Desensitized.Email(c.email); got != c.want {
			t.Errorf("Email(%q) = %q, want %q", c.email, got, c.want)
		}
	}
	// 结构校验：首字符保留 + @ 前星号数 = 本地部分长度-1，@ 及域名原样
	const email = "duandazhi-jack@gmail.com.cn"
	got := Desensitized.Email(email)
	at := strings.Index(got, "@")
	if at != strings.Index(email, "@") ||
		got[0] != email[0] ||
		strings.Count(got[:at], "*") != at-1 ||
		got[at+1:] != email[at+1:] {
		t.Errorf("Email(%q) = %q 不满足首字符+星号数=长度-1", email, got)
	}
	// 无 @ 时原样返回
	if got := Desensitized.Email("abcdef"); got != "abcdef" {
		t.Errorf("Email(无 @) = %q, want 原样", got)
	}
}

func TestDesensitizedPassword(t *testing.T) {
	// hutool DesensitizedUtilTest.passwordTest：星号数与原长一致
	if got := Desensitized.Password("1234567890"); got != "**********" {
		t.Errorf("Password(1234567890) = %q, want 10 个星号", got)
	}
	if got := Desensitized.Password("123456"); got != "******" {
		t.Errorf("Password(123456) = %q, want 6 个星号", got)
	}
	if n := strings.Count(Desensitized.Password("abc123!@"), "*"); n != 8 {
		t.Errorf("Password 星号数 = %d, want 8（等于原长）", n)
	}
}

func TestDesensitizedCarLicense(t *testing.T) {
	// hutool DesensitizedUtilTest.carLicenseTest
	cases := []struct {
		license string
		want    string
	}{
		{"苏D40000", "苏D4***0"},   // 普通 7 位牌
		{"陕A12345D", "陕A1****D"}, // 新能源 8 位牌
		{"京A123", "京A123"},       // 非法长度原样返回
		{"", ""},                 // 空串返回空串
	}
	for _, c := range cases {
		if got := Desensitized.CarLicense(c.license); got != c.want {
			t.Errorf("CarLicense(%q) = %q, want %q", c.license, got, c.want)
		}
	}
}

func TestDesensitizedBankCard(t *testing.T) {
	// hutool DesensitizedUtilTest.bankCardTest
	cases := []struct {
		no   string
		want string
	}{
		{"11011111222233333256", "1101 **** **** **** 3256"},
		{"6227880100100105123", "6227 **** **** **** 123"},
		{"1234 2222 3333 4444 6789 9", "1234 **** **** **** **** 9"},
		{"1234 2222 3333 4444 6789", "1234 **** **** **** 6789"},
		{"1234 2222 3333 4444 678", "1234 **** **** **** 678"},
		{"12345678", "12345678"}, // 不足 9 位原样返回
		{"", ""},
	}
	for _, c := range cases {
		if got := Desensitized.BankCard(c.no); got != c.want {
			t.Errorf("BankCard(%q) = %q, want %q", c.no, got, c.want)
		}
	}
}

func TestDesensitizedIp(t *testing.T) {
	// hutool DesensitizedUtilTest.desensitizedTest（IPV4/IPV6）
	if got := Desensitized.Ipv4("192.168.1.1"); got != "192.*.*.*" {
		t.Errorf("Ipv4 = %q, want 192.*.*.*", got)
	}
	if got := Desensitized.Ipv6("2001:0db8:86a3:08d3:1319:8a2e:0370:7344"); got != "2001:*:*:*:*:*:*:*" {
		t.Errorf("Ipv6 = %q, want 2001:*:*:*:*:*:*:*", got)
	}
}

func TestDesensitizedPassport(t *testing.T) {
	// hutool DesensitizedUtilTest.passportTest
	cases := []struct {
		passport string
		want     string
	}{
		{"EM1234567", "EM*****67"},
		{"PJ1234567", "PJ*****67"},
		{"3", "*"}, // 长度不足等于 2 位时整体打码
		{"", ""},
	}
	for _, c := range cases {
		if got := Desensitized.Passport(c.passport); got != c.want {
			t.Errorf("Passport(%q) = %q, want %q", c.passport, got, c.want)
		}
	}
}

func TestDesensitizedCreditCode(t *testing.T) {
	// hutool DesensitizedUtilTest.creditCodeTest
	if got := Desensitized.CreditCode("91110108MA01ABCDE7"); got != "9111**********CDE7" {
		t.Errorf("CreditCode = %q, want 9111**********CDE7", got)
	}
	// 长度不足等于 4 位时整体打码
	if got := Desensitized.CreditCode("abc"); got != "***" {
		t.Errorf("CreditCode(短串) = %q, want ***", got)
	}
	if got := Desensitized.CreditCode(""); got != "" {
		t.Errorf("CreditCode(空串) = %q, want 空串", got)
	}
}

/* ---------------- 占位与清空 ---------------- */

func TestDesensitizedUserIdAndClear(t *testing.T) {
	// hutool DesensitizedUtilTest.userIdTest
	if got := Desensitized.UserId(); got != 0 {
		t.Errorf("UserId = %d, want 0", got)
	}
	if got := Desensitized.Clear("secret"); got != "" {
		t.Errorf("Clear = %q, want 空串", got)
	}
}

/* ---------------- 类型分发的统一入口 ---------------- */

func TestDesensitizedDispatch(t *testing.T) {
	cases := []struct {
		name     string
		in       string
		typeName string
		want     string
	}{
		// 全类型分发（黄金值同 hutool DesensitizedUtilTest.desensitizedTest）
		{"USER_ID", "100", "USER_ID", "0"},
		{"CHINESE_NAME", "段正淳", "CHINESE_NAME", "段**"},
		{"ID_CARD", "51343620000320711X", "ID_CARD", "5***************1X"},
		{"FIXED_PHONE", "09157518479", "FIXED_PHONE", "0915*****79"},
		{"MOBILE_PHONE", "18049531999", "MOBILE_PHONE", "180****1999"},
		{"ADDRESS", "北京市海淀区马连洼街道289号", "ADDRESS", "北京市海淀区马********"},
		{"EMAIL", "duandazhi-jack@gmail.com.cn", "EMAIL", "d*************@gmail.com.cn"},
		{"PASSWORD", "1234567890", "PASSWORD", "**********"},
		{"BANK_CARD", "11011111222233333256", "BANK_CARD", "1101 **** **** **** 3256"},
		{"IPV4", "192.168.1.1", "IPV4", "192.*.*.*"},
		{"IPV6", "2001:0db8:86a3:08d3:1319:8a2e:0370:7344", "IPV6", "2001:*:*:*:*:*:*:*"},
		{"PASSPORT", "EM1234567", "PASSPORT", "EM*****67"},
		{"CREDIT_CODE", "91110108MA01ABCDE7", "CREDIT_CODE", "9111**********CDE7"},
		{"CAR_LICENSE", "苏D40000", "CAR_LICENSE", "苏D4***0"},
		{"FIRST_MASK", "123456789", "FIRST_MASK", "1********"},
		{"CLEAR", "100", "CLEAR", ""},
		{"CLEAR_TO_EMPTY", "100", "CLEAR_TO_EMPTY", ""},
		{"CLEAR_TO_NULL", "100", "CLEAR_TO_NULL", ""},
		// 类型名大小写不敏感
		{"小写类型名", "18049531999", "mobile_phone", "180****1999"},
		// 未知类型原样返回
		{"未知类型", "13812345678", "UNKNOWN_TYPE", "13812345678"},
		// 空白输入统一返回空串
		{"空输入", "", "MOBILE_PHONE", ""},
	}
	for _, c := range cases {
		if got := Desensitized.Desensitized(c.in, c.typeName); got != c.want {
			t.Errorf("%s: Desensitized(%q, %q) = %q, want %q", c.name, c.in, c.typeName, got, c.want)
		}
	}
}
