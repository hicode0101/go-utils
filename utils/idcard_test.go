package utils

import (
	"testing"
	"time"
)

// idcard_test.go：Idcard 身份证工具逐方法全覆盖。
// 黄金值来源：hutool 官方测试 IdcardUtilTest.java（ID_18/ID_15/FOREIGN_ID_18 等）
// 与 ValidatorTest.java（isCitizenIdTest）；无 hutool 用例的值已用临时程序验证后固化。

const (
	testID18      = "321083197812162119" // hutool ID_18（男，江苏）
	testID15      = "150102880730303"    // hutool ID_15（内蒙古）
	testForeign18 = "932682198501010017" // hutool FOREIGN_ID_18（新版外国人永久居留证）
)

/* ---------------- 综合校验 ---------------- */

func TestIdcardIsValidCard(t *testing.T) {
	cases := []struct {
		name string
		id   string
		want bool
	}{
		{"18 位", testID18, true},
		{"15 位", testID15, true},
		{"外国人永久居留证", testForeign18, true},
		{"香港证件", "P174468(6)", true},
		{"台湾证件", "B221690311", true},
		{"台湾证件（isCitizenId 用例）", "U193683453", true},
		{"15 位全 1（issue#I88YKM）", "111111111111111", true},
		// 非法
		{"香港形态不完整（issue#IAFOLI）", "H01487002", false},
		{"15 位出生日期非法", "360198910283844", false},
		{"18 位年份早于 1900", "201511221897205960", false},
		{"15 位日期非法", "815727834224151", false},
		{"空串", "", false},
		{"空白串", "   ", false},
	}
	for _, c := range cases {
		if got := Idcard.IsValidCard(c.id); got != c.want {
			t.Errorf("%s: IsValidCard(%q) = %v, want %v", c.name, c.id, got, c.want)
		}
	}
}

func TestIdcardIsValidCard15(t *testing.T) {
	if !Idcard.IsValidCard15(testID15) {
		t.Error("IsValidCard15(150102880730303) 应为 true")
	}
	if !Idcard.IsValidCard15("111111111111111") {
		t.Error("IsValidCard15(111111111111111) 应为 true（issue#I88YKM）")
	}
	if Idcard.IsValidCard15("815727834224151") {
		t.Error("IsValidCard15(815727834224151) 出生日期非法应为 false")
	}
	if Idcard.IsValidCard15("15010288073030") {
		t.Error("IsValidCard15(14 位) 应为 false")
	}
}

func TestIdcardIsValidCard18(t *testing.T) {
	cases := []struct {
		name string
		id   string
		want bool
	}{
		{"有效", testID18, true},
		{"顺序码含字母 D", "3301022011022000D6", false},
		// 校验位 x/X 均接受（ignoreCase=true）
		{"校验位小写 x", "33010219200403064x", true},
		{"校验位大写 X", "33010219200403064X", true},
		{"香港人在大陆", "81000019980902013X", true},
		{"澳门人在大陆", "820000200009100032", true},
		{"台湾人在大陆", "830000200209060065", true},
		{"外国人永久居留证", testForeign18, true},
	}
	for _, c := range cases {
		if got := Idcard.IsValidCard18(c.id); got != c.want {
			t.Errorf("%s: IsValidCard18(%q) = %v, want %v", c.name, c.id, got, c.want)
		}
	}
}

func TestIdcardIsValidHKCard(t *testing.T) {
	// hutool IdcardUtilTest.isValidHKCardIdTest / issueIAFOLITest
	if !Idcard.IsValidHKCard("P174468(6)") {
		t.Error("IsValidHKCard(P174468(6)) 应为 true")
	}
	if Idcard.IsValidHKCard("H01487002") {
		t.Error("IsValidHKCard(H01487002) 应为 false")
	}
}

func TestIdcardIsValidTWCard(t *testing.T) {
	// hutool IdcardUtilTest.isValidTWCardIdTest
	if !Idcard.IsValidTWCard("B221690311") {
		t.Error("IsValidTWCard(B221690311) 应为 true")
	}
	if Idcard.IsValidTWCard("M517086311") {
		t.Error("IsValidTWCard(M517086311) 应为 false")
	}
	if Idcard.IsValidTWCard("B2216903112") {
		t.Error("IsValidTWCard(B2216903112) 11 位应为 false")
	}
}

/* ---------------- 15/18 位互转 ---------------- */

func TestIdcardConvert15To18(t *testing.T) {
	// hutool IdcardUtilTest.convert15To18Test
	got, err := Idcard.Convert15To18(testID15)
	if err != nil || got != "150102198807303035" {
		t.Errorf("Convert15To18(ID_15) = %q, %v; want 150102198807303035", got, err)
	}
	// 两位年份 20 误入 21 世纪时回退 100 年
	got, err = Idcard.Convert15To18("330102200403064")
	if err != nil || got != "33010219200403064X" {
		t.Errorf("Convert15To18(330102200403064) = %q, %v; want 33010219200403064X", got, err)
	}

	// 非法输入返回 error
	for _, bad := range []string{"1501028807303033", "1501028a0730303", "150102880732303", ""} {
		if _, err := Idcard.Convert15To18(bad); err == nil {
			t.Errorf("Convert15To18(%q) 应返回 error", bad)
		}
	}
}

func TestIdcardConvert18To15(t *testing.T) {
	// hutool IdcardUtilTest.convert18To15Test
	got, err := Idcard.Convert18To15("150102198807303035")
	if err != nil || got != testID15 {
		t.Errorf("Convert18To15 往返 = %q, %v; want %q", got, err, testID15)
	}
	got, err = Idcard.Convert18To15(testID18)
	if err != nil || got != "321083781216211" {
		t.Errorf("Convert18To15(ID_18) = %q, %v; want 321083781216211", got, err)
	}

	// 非法输入返回 error
	if _, err := Idcard.Convert18To15("321083197812162118"); err == nil {
		t.Error("Convert18To15(校验位错误) 应返回 error")
	}
	if _, err := Idcard.Convert18To15(""); err == nil {
		t.Error("Convert18To15(空串) 应返回 error")
	}
}

/* ---------------- 信息提取 ---------------- */

func TestIdcardGetBirth(t *testing.T) {
	// hutool IdcardUtilTest.getBirthByIdCardTest
	if got := Idcard.GetBirthByIdCard(testID18); got != "19781216" {
		t.Errorf("GetBirthByIdCard(ID_18) = %q, want 19781216", got)
	}
	if got := Idcard.GetBirthByIdCard(testID15); got != "19880730" {
		t.Errorf("GetBirthByIdCard(ID_15) = %q, want 19880730", got)
	}
	if got := Idcard.GetBirthByIdCard("abc"); got != "" {
		t.Errorf("GetBirthByIdCard(非法) = %q, want 空串", got)
	}

	// GetBirthDate 按年月日断言
	bd := Idcard.GetBirthDate(testID18)
	if bd.Year() != 1978 || bd.Month() != time.December || bd.Day() != 16 {
		t.Errorf("GetBirthDate(ID_18) = %v, want 1978-12-16", bd)
	}
	if got := Idcard.GetBirthDate(testID15); got.Year() != 1988 || got.Month() != time.July || got.Day() != 30 {
		t.Errorf("GetBirthDate(ID_15) = %v, want 1988-07-30", got)
	}
	if !Idcard.GetBirthDate("abc").IsZero() {
		t.Error("GetBirthDate(非法) 应返回零值 time.Time")
	}
}

func TestIdcardGetAgeByIdCard(t *testing.T) {
	now := time.Date(2017, 4, 10, 0, 0, 0, 0, time.Local) // hutool getAgeByIdCardTest
	cases := []struct {
		name string
		id   string
		now  time.Time
		want int
	}{
		{"ID_18 未到生日", testID18, now, 38},
		{"外国人证", testForeign18, now, 32},
		{"ID_15 未到生日", testID15, now, 28},
		{"生日前一天", "321083200807112111", time.Date(2014, 7, 10, 0, 0, 0, 0, time.Local), 5},
		{"生日当天", "321083200807112111", time.Date(2014, 7, 11, 0, 0, 0, 0, time.Local), 6},
		{"生日次日", "321083200807112111", time.Date(2014, 7, 12, 0, 0, 0, 0, time.Local), 6},
		{"生日晚于 now 返回 0", testID18, time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local), 0},
		{"无法解析返回 0", "abc", now, 0},
	}
	for _, c := range cases {
		if got := Idcard.GetAgeByIdCard(c.id, c.now); got != c.want {
			t.Errorf("%s: GetAgeByIdCard(%q) = %d, want %d", c.name, c.id, got, c.want)
		}
	}

	// now 传零值时按当前时间计算（能正常解析出正周岁即可）
	if got := Idcard.GetAgeByIdCard(testID18, time.Time{}); got <= 0 {
		t.Errorf("GetAgeByIdCard(零值 now) = %d, want > 0", got)
	}
}

func TestIdcardGetYearMonthDay(t *testing.T) {
	if got := Idcard.GetYearByIdCard(testID18); got != 1978 {
		t.Errorf("GetYearByIdCard = %d, want 1978", got)
	}
	if got := Idcard.GetMonthByIdCard(testID18); got != 12 {
		t.Errorf("GetMonthByIdCard = %d, want 12", got)
	}
	if got := Idcard.GetDayByIdCard(testID18); got != 16 {
		t.Errorf("GetDayByIdCard = %d, want 16", got)
	}
	// 无法解析返回 0
	if got := Idcard.GetYearByIdCard("12345"); got != 0 {
		t.Errorf("GetYearByIdCard(非法) = %d, want 0", got)
	}
	if got := Idcard.GetMonthByIdCard("12345"); got != 0 {
		t.Errorf("GetMonthByIdCard(非法) = %d, want 0", got)
	}
	if got := Idcard.GetDayByIdCard("12345"); got != 0 {
		t.Errorf("GetDayByIdCard(非法) = %d, want 0", got)
	}
}

func TestIdcardGetGenderByIdCard(t *testing.T) {
	// 1 男 2 女：第 17 位顺序码奇数男、偶数女
	if got := Idcard.GetGenderByIdCard(testID18); got != 1 {
		t.Errorf("GetGenderByIdCard(ID_18) = %d, want 1（男）", got)
	}
	if got := Idcard.GetGenderByIdCard("33010219200403064X"); got != 2 {
		t.Errorf("GetGenderByIdCard(33010219200403064X) = %d, want 2（女）", got)
	}
	if got := Idcard.GetGenderByIdCard("830000200209060065"); got != 2 {
		t.Errorf("GetGenderByIdCard(830000200209060065) = %d, want 2（女）", got)
	}
	if got := Idcard.GetGenderByIdCard("12345"); got != 0 {
		t.Errorf("GetGenderByIdCard(非法) = %d, want 0", got)
	}
}

func TestIdcardGetProvince(t *testing.T) {
	// 编码：GetProvinceCodeByIdCard
	if got := Idcard.GetProvinceCodeByIdCard(testID18); got != "32" {
		t.Errorf("GetProvinceCodeByIdCard(ID_18) = %q, want 32", got)
	}
	// 外国人永久居留证以 9 开头时取第 2~3 位受理地代码
	if got := Idcard.GetProvinceCodeByIdCard(testForeign18); got != "32" {
		t.Errorf("GetProvinceCodeByIdCard(FOREIGN) = %q, want 32", got)
	}
	if got := Idcard.GetProvinceCodeByIdCard("12345"); got != "" {
		t.Errorf("GetProvinceCodeByIdCard(非法长度) = %q, want 空串", got)
	}

	// 名称：GetProvinceByIdCard（hutool getProvinceByIdCardTest）
	if got := Idcard.GetProvinceByIdCard(testID18); got != "江苏" {
		t.Errorf("GetProvinceByIdCard(ID_18) = %q, want 江苏", got)
	}
	if got := Idcard.GetProvinceByIdCard(testID15); got != "内蒙古" {
		t.Errorf("GetProvinceByIdCard(ID_15) = %q, want 内蒙古", got)
	}
	// 未知省份编码返回空串
	if got := Idcard.GetProvinceByIdCard("031234190001010011"); got != "" {
		t.Errorf("GetProvinceByIdCard(未知编码 03) = %q, want 空串", got)
	}
}

func TestIdcardGetCityDistrictCode(t *testing.T) {
	// hutool getCityCodeByIdCardTest / getDistrictCodeByIdCardTest
	if got := Idcard.GetCityCodeByIdCard(testID18); got != "3210" {
		t.Errorf("GetCityCodeByIdCard = %q, want 3210", got)
	}
	if got := Idcard.GetDistrictCodeByIdCard(testID18); got != "321083" {
		t.Errorf("GetDistrictCodeByIdCard = %q, want 321083", got)
	}
	if got := Idcard.GetCityCodeByIdCard("12345"); got != "" {
		t.Errorf("GetCityCodeByIdCard(非法长度) = %q, want 空串", got)
	}
	if got := Idcard.GetDistrictCodeByIdCard("12345"); got != "" {
		t.Errorf("GetDistrictCodeByIdCard(非法长度) = %q, want 空串", got)
	}
}

func TestIdcardHide(t *testing.T) {
	if got := Idcard.Hide("110101199003077854", 6, 14); got != "110101********7854" {
		t.Errorf("Hide = %q, want 110101********7854", got)
	}
}

func TestIdcardGetIdcardInfo(t *testing.T) {
	info := Idcard.GetIdcardInfo(testID18)
	if info == nil {
		t.Fatal("GetIdcardInfo(ID_18) 不应返回 nil")
	}
	if info.Province != "江苏" {
		t.Errorf("Province = %q, want 江苏", info.Province)
	}
	if info.City != "3210" {
		t.Errorf("City = %q, want 3210", info.City)
	}
	if info.District != "321083" {
		t.Errorf("District = %q, want 321083", info.District)
	}
	if info.Gender != 1 {
		t.Errorf("Gender = %d, want 1", info.Gender)
	}
	if info.Birthday != "19781216" {
		t.Errorf("Birthday = %q, want 19781216", info.Birthday)
	}
	if info.Age != Idcard.GetAgeByIdCard(testID18, time.Time{}) {
		t.Errorf("Age = %d, 与 GetAgeByIdCard 当前周岁不一致", info.Age)
	}

	// 15 位同样可解析
	if Idcard.GetIdcardInfo(testID15) == nil {
		t.Error("GetIdcardInfo(ID_15) 不应返回 nil")
	}
	// 非法输入返回 nil
	if got := Idcard.GetIdcardInfo("360198910283844"); got != nil {
		t.Errorf("GetIdcardInfo(15 位日期非法) = %+v, want nil", got)
	}
	if got := Idcard.GetIdcardInfo("12345"); got != nil {
		t.Errorf("GetIdcardInfo(非法长度) = %+v, want nil", got)
	}
}
