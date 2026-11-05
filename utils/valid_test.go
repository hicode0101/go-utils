package utils

import "testing"

// valid_test.go：Valid 字段校验器逐方法全覆盖。
// 黄金值来源：hutool 官方测试 ValidatorTest.java、CreditCodeUtilTest.java、
// IdcardUtilTest.java；无 hutool 用例的值已用临时程序逐一验证后固化。

/* ---------------- 内容与命名格式 ---------------- */

func TestValidIsEmail(t *testing.T) {
	cases := []struct {
		name  string
		email string
		want  bool
	}{
		// 正例（hutool ValidatorTest.isEmailTest）
		{"常规邮箱", "abc_cde@163.com", true},
		{"本地部分含 %", "abc_%cde@163.com", true},
		{"单字符域名", "abc_%cde@aaa.c", true},
		{"user@example.com", "user@example.com", true},
		// 反例
		{"双 @@结尾", "bad@@", false},
		{"缺少 @", "abc163.com", false},
		{"双 @", "a@@b.com", false},
	}
	for _, c := range cases {
		if got := Valid.IsEmail(c.email); got != c.want {
			t.Errorf("%s: IsEmail(%q) = %v, want %v", c.name, c.email, got, c.want)
		}
	}
}

func TestValidIsMobile(t *testing.T) {
	cases := []struct {
		name   string
		mobile string
		want   bool
	}{
		// 正例（hutool ValidatorTest.isMobileTest）
		{"138 开头", "13900221432", true},
		{"0 前缀", "015100221432", true},
		{"+86 前缀", "+8618600221432", true},
		{"19 号段", "19312341234", true},
		// 反例
		{"位数多一位", "136123456781", false},
		{"第二位为 2", "12345678901", false},
		{"12 开头", "12900221432", false},
	}
	for _, c := range cases {
		if got := Valid.IsMobile(c.mobile); got != c.want {
			t.Errorf("%s: IsMobile(%q) = %v, want %v", c.name, c.mobile, got, c.want)
		}
	}
}

func TestValidIsCitizenId(t *testing.T) {
	cases := []struct {
		name string
		id   string
		want bool
	}{
		// 正例（hutool ValidatorTest.isCitizenIdTest 与 IdcardUtilTest）
		{"18 位有效", "110101199003074477", true},
		{"18 位有效（江苏）", "321083197812162119", true},
		{"15 位有效", "410001910101123", true},
		{"10 位台湾证件", "U193683453", true},
		// 反例
		{"校验位错误", "321083197812162118", false},
		{"出生日期非法", "360198910283844", false},
		{"空串", "", false},
		{"含空白", " 321083197812162119 ", false},
	}
	for _, c := range cases {
		if got := Valid.IsCitizenId(c.id); got != c.want {
			t.Errorf("%s: IsCitizenId(%q) = %v, want %v", c.name, c.id, got, c.want)
		}
	}
}

func TestValidIsBirthday(t *testing.T) {
	cases := []struct {
		name string
		day  string
		want bool
	}{
		// 正例（hutool ValidatorTest.isBirthdayTest，覆盖源码支持的各分隔符格式）
		{"yyyyMMdd", "20150101", true},
		{"yyyy-MM-dd", "2015-01-01", true},
		{"yyyy/MM/dd", "2015/01/01", true},
		{"yyyy.MM.dd", "2015.01.01", true},
		{"yyyy年MM月dd日", "2015年01月01日", true},
		{"普通日期", "2018-08-15", true},
		{"月日单位数", "2015-1-1", true},
		// 闰年 2 月 29 日有效
		{"闰年 2000-02-29", "20000229", true},
		{"闰年 2024-02-29", "20240229", true},
		// 反例
		{"平年 2 月 29 日", "2015.02.29", false},
		{"平年 2 月 29 日（横线）", "2015-02-29", false},
		{"13 月非法", "2015.13.01", false},
		{"年份超过当前年", "2095.05.01", false},
		{"两位年份", "150101", false},
	}
	for _, c := range cases {
		if got := Valid.IsBirthday(c.day); got != c.want {
			t.Errorf("%s: IsBirthday(%q) = %v, want %v", c.name, c.day, got, c.want)
		}
	}
}

func TestValidIsCarVin(t *testing.T) {
	cases := []struct {
		name string
		vin  string
		want bool
	}{
		// 正例（hutool ValidatorTest.isCarVinTest）
		{"标准分类 1", "LSJA24U62JG269225", true},
		{"标准分类 1（二）", "LDC613P23A1305189", true},
		{"标准分类 1（三）", "LBV5S3102ESJ25655", true},
		{"标准分类 2", "LBV5S3102ESJPE655", true},
		// 反例
		{"含易混淆字母 O", "LOC613P23A1305189", false},
		{"位数不足", "LSJA24U62JG26922", false},
	}
	for _, c := range cases {
		if got := Valid.IsCarVin(c.vin); got != c.want {
			t.Errorf("%s: IsCarVin(%q) = %v, want %v", c.name, c.vin, got, c.want)
		}
	}
}

func TestValidIsPlateNumber(t *testing.T) {
	cases := []struct {
		name  string
		plate string
		want  bool
	}{
		// 正例（hutool ValidatorTest.isPlateNumberTest）
		{"普通车牌", "粤BA03205", true},
		{"领事馆号牌", "闽20401领", true},
		{"新标准普通牌", "沪AE22075", true},
		{"issue#IJNDJR", "粤AP00000", true},
		{"新能源车牌", "苏ED12345", true},
		{"普通 7 位牌", "苏D40000", true},
		// 反例
		{"无省份前缀", "BAD123", false},
		{"位数不足", "粤B1234", false},
	}
	for _, c := range cases {
		if got := Valid.IsPlateNumber(c.plate); got != c.want {
			t.Errorf("%s: IsPlateNumber(%q) = %v, want %v", c.name, c.plate, got, c.want)
		}
	}
}

func TestValidIsMac(t *testing.T) {
	cases := []struct {
		name string
		mac  string
		want bool
	}{
		// 正例：冒号 / 横线 / 思科点分 / 12 位无分隔 / 0x+ETHER 五种形态
		{"冒号分隔", "00:1A:2B:3C:4D:5E", true},
		{"横线分隔", "AA-BB-CC-DD-EE-FF", true},
		{"思科点分", "AABB.CCDD.EEFF", true},
		{"12 位无分隔", "001122334455", true},
		{"0x+ETHER 后缀", "0x000000000001+ETHER", true},
		// 反例
		{"只有 5 组", "00:1A:2B:3C:4D", false},
		{"非十六进制字符", "ZZ:BB:CC:DD:EE:FF", false},
	}
	for _, c := range cases {
		if got := Valid.IsMac(c.mac); got != c.want {
			t.Errorf("%s: IsMac(%q) = %v, want %v", c.name, c.mac, got, c.want)
		}
	}
}

func TestValidIsIpv4(t *testing.T) {
	cases := []struct {
		name string
		ip   string
		want bool
	}{
		{"常规地址", "192.168.1.1", true},
		{"全零", "0.0.0.0", true},
		{"全 255", "255.255.255.255", true},
		{"段超 255", "256.1.1.1", false},
		{"段为 999", "999.1.1.1", false},
		{"只有 3 段", "1.2.3", false},
	}
	for _, c := range cases {
		if got := Valid.IsIpv4(c.ip); got != c.want {
			t.Errorf("%s: IsIpv4(%q) = %v, want %v", c.name, c.ip, got, c.want)
		}
	}
}

func TestValidIsIpv6(t *testing.T) {
	cases := []struct {
		name string
		ip   string
		want bool
	}{
		{"::1 缩写", "::1", true},
		{"fe80::1", "fe80::1", true},
		{"完整形态", "2001:0db8:86a3:08d3:1319:8a2e:0370:7344", true},
		{"内嵌 IPv4", "::ffff:192.168.1.1", true},
		{"段数超 8", "1:2:3:4:5:6:7:8:9", false},
		{"非十六进制字符", "g::1", false},
	}
	for _, c := range cases {
		if got := Valid.IsIpv6(c.ip); got != c.want {
			t.Errorf("%s: IsIpv6(%q) = %v, want %v", c.name, c.ip, got, c.want)
		}
	}
}

func TestValidIsZipCode(t *testing.T) {
	cases := []struct {
		name string
		zip  string
		want bool
	}{
		// 正例（hutool ValidatorTest.isZipCodeTest）
		{"北京", "102629", true},
		{"内蒙", "016063", true},
		{"山西", "045246", true},
		{"河北", "066502", true},
		{"港 999077", "999077", true},
		{"澳 999078", "999078", true},
		{"台 822001", "822001", true},
		// 反例
		{"首两位越界", "099999", false},
		{"位数多一位", "1026290", false},
		{"位数少一位", "12345", false},
	}
	for _, c := range cases {
		if got := Valid.IsZipCode(c.zip); got != c.want {
			t.Errorf("%s: IsZipCode(%q) = %v, want %v", c.name, c.zip, got, c.want)
		}
	}
}

func TestValidIsMoney(t *testing.T) {
	cases := []struct {
		name  string
		money string
		want  bool
	}{
		{"整数", "123", true},
		{"小数", "12.34", true},
		{"零头小数", "0.33323", true},
		{"负数", "-12", false},
		{"双小数点", "12.3.4", false},
		{"千分位", "1,234", false},
		{"空串", "", false},
	}
	for _, c := range cases {
		if got := Valid.IsMoney(c.money); got != c.want {
			t.Errorf("%s: IsMoney(%q) = %v, want %v", c.name, c.money, got, c.want)
		}
	}
}

func TestValidIsChineseName(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		// 正例（hutool ValidatorTest.isChineseNameTest）
		{"三字中文", "段正淳", true},
		{"生僻字 䶮", "刘欣䶮", true},
		{"维吾尔族间隔号 ·", "阿卜杜尼亚孜·毛力尼亚孜", true},
		// 反例
		{"含斜杠句点", "阿卜杜尼亚孜./毛力尼亚孜", false},
		{"含空格", "孟  伟", false},
		{"单字", "李", false},
		{"含数字", "连逍遥0", false},
		{"英文", "SHE", false},
		{"含中文句号", "张三。", false},
	}
	for _, c := range cases {
		if got := Valid.IsChineseName(c.in); got != c.want {
			t.Errorf("%s: IsChineseName(%q) = %v, want %v", c.name, c.in, got, c.want)
		}
	}
}

func TestValidHasChineseAndIsChinese(t *testing.T) {
	// HasChinese（hutool ValidatorTest.hasChineseTest）
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"全中文", "黄单桑米", true},
		{"中英混合", "Kn 四兄弟", true},
		{"增补平面汉字 U+201A3", "\U000201A3", true},
		{"纯英文", "Abc", false},
	}
	for _, c := range cases {
		if got := Valid.HasChinese(c.in); got != c.want {
			t.Errorf("%s: HasChinese(%q) = %v, want %v", c.name, c.in, got, c.want)
		}
	}

	// IsChinese（hutool ValidatorTest.isChineseTest）
	if !Valid.IsChinese("全都是中文") {
		t.Error("IsChinese(全都是中文) 应为 true")
	}
	if !Valid.IsChinese("㐓㐘") {
		t.Error("IsChinese(㐓㐘) 扩展区汉字应为 true")
	}
	if Valid.IsChinese("not全都是中文") {
		t.Error("IsChinese(含英文) 应为 false")
	}
}

func TestValidIsGeneral(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"字母数字下划线", "123_abc_ccc", true},
		{"纯字母", "abc", true},
		{"含中文", "123_abc_ccc中文", false},
		{"含中划线", "abc-123", false},
		{"空串", "", false},
	}
	for _, c := range cases {
		if got := Valid.IsGeneral(c.in); got != c.want {
			t.Errorf("%s: IsGeneral(%q) = %v, want %v", c.name, c.in, got, c.want)
		}
	}
}

func TestValidIsGeneralWithChinese(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"中文+字母数字下划线", "用户_user_01", true},
		{"纯英文数字", "user_01", true},
		{"含中划线", "用户-user", false},
	}
	for _, c := range cases {
		if got := Valid.IsGeneralWithChinese(c.in); got != c.want {
			t.Errorf("%s: IsGeneralWithChinese(%q) = %v, want %v", c.name, c.in, got, c.want)
		}
	}
}

func TestValidIsCreditCode(t *testing.T) {
	cases := []struct {
		name string
		code string
		want bool
	}{
		// 正例（hutool CreditCodeUtilTest，ISO 7064 MOD 31-2 校验通过）
		{"有效代码 T", "91310110666007217T", true},
		{"有效代码 A", "91310115591693856A", true},
		// 反例
		{"试点地区旧码不合规", "91350211M00013FA1N", false},
		{"篡改校验位", "91310110666007217U", false},
		{"空串", "", false},
	}
	for _, c := range cases {
		if got := Valid.IsCreditCode(c.code); got != c.want {
			t.Errorf("%s: IsCreditCode(%q) = %v, want %v", c.name, c.code, got, c.want)
		}
	}
}

func TestValidRandomCreditCode(t *testing.T) {
	// 生成即通过自身校验（hutool CreditCodeUtilTest.randomCreditCode 语义，多次抽样）
	for i := 0; i < 20; i++ {
		code := Valid.RandomCreditCode()
		if len(code) != 18 {
			t.Fatalf("RandomCreditCode 长度 = %d, want 18", len(code))
		}
		if !Valid.IsCreditCode(code) {
			t.Fatalf("RandomCreditCode 生成的 %q 应通过 IsCreditCode 校验", code)
		}
	}
}
