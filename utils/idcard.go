package utils

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// IdcardInfo 身份证解析信息（对照 hutool IdcardUtil.Idcard）。
type IdcardInfo struct {
	Province string // 省份名称
	City     string // 城市编码（4 位）
	District string // 区县编码（6 位）
	Gender   int    // 1 男 2 女
	Age      int    // 周岁
	Birthday string // yyyyMMdd
}

// utilIdcard 身份证工具（对照 hutool IdcardUtil）。
//
// 支持 15/18 位大陆身份证与 10 位港澳台证件号的校验、互转与信息提取；
// 与 hutool 一致只查省级代码表，不校验 6 位行政区划的区县级合法性。
type utilIdcard struct{}

/* ---------------- 常量表与正则（包初始化时统一预编译） ---------------- */

var (
	// 省级代码表（对照 IdcardUtil.CITY_CODES）
	idcardProvinceCodes = map[string]string{
		"11": "北京", "12": "天津", "13": "河北", "14": "山西", "15": "内蒙古",
		"21": "辽宁", "22": "吉林", "23": "黑龙江",
		"31": "上海", "32": "江苏", "33": "浙江", "34": "安徽", "35": "福建", "36": "江西", "37": "山东",
		"41": "河南", "42": "湖北", "43": "湖南", "44": "广东", "45": "广西", "46": "海南",
		"50": "重庆", "51": "四川", "52": "贵州", "53": "云南", "54": "西藏",
		"61": "陕西", "62": "甘肃", "63": "青海", "64": "宁夏", "65": "新疆",
		"71": "台湾", "81": "香港", "82": "澳门",
		// 台湾身份证号码有以 83 开头的形态，行政区划仍按台湾处理（hutool issue#1277）
		"83": "台湾",
		"91": "国外",
	}
	// 台湾身份证首字母对应数字（对照 IdcardUtil.TW_FIRST_CODE）
	idcardTwFirstCode = map[byte]int{
		'A': 10, 'B': 11, 'C': 12, 'D': 13, 'E': 14, 'F': 15, 'G': 16, 'H': 17,
		'J': 18, 'K': 19, 'L': 20, 'M': 21, 'N': 22, 'P': 23, 'Q': 24, 'R': 25,
		'S': 26, 'T': 27, 'U': 28, 'V': 29, 'X': 30, 'Y': 31, 'W': 32, 'Z': 33,
		'I': 34, 'O': 35,
	}
	// 18 位身份证校验位加权因子（POWER）
	idcardWeight18 = [17]int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	// 纯数字（PatternPool.NUMBERS 整串匹配）
	idcardDigitsRe = regexp.MustCompile(`^\d+$`)
	// 台湾身份证形态：1 位字母 + 9 位数字
	idcardTwFormRe = regexp.MustCompile(`^[a-zA-Z][0-9]{9}$`)
	// 澳门身份证形态：1/5/7 开头 + 6 位数字 + 1 位数字或大写字母（括号可选）
	idcardMoFormRe = regexp.MustCompile(`^[157][0-9]{6}\(?[0-9A-Z]\)?$`)
	// 香港身份证形态：1~2 位大写字母 + 6 位数字 + 校验位 0-9/A（括号可选）
	idcardHkFormRe = regexp.MustCompile(`^[A-Z]{1,2}[0-9]{6}\(?[0-9A]\)?$`)
)

/* ---------------- 综合校验 ---------------- */

// IsValidCard 是否有效身份证号（对照 hutool isValidCard）：自动按长度分派——
// 18 位做完整校验（省码 + 生日 + ISO 7064 MOD 11-2 校验位，末位 x/X 均可），
// 15 位校验省码与出生日期，10 位按港澳台证件算法校验；含空白一律 false。
// 场景：实名认证等不限定证件位数的入口校验。
func (i *utilIdcard) IsValidCard(idcard string) bool {
	if strings.TrimSpace(idcard) == "" {
		return false
	}
	switch len(idcard) {
	case 18:
		return idcardIsValidCard18(idcard)
	case 15:
		return idcardIsValidCard15(idcard)
	case 10:
		recognized, valid := idcardIsValidCard10(idcard)
		return recognized && valid
	default:
		return false
	}
}

// IsValidCard15 验证 15 位身份证是否合法：全数字、省码存在、
// 两位年份按 19XX 补全后出生日期合法（15 位无校验位）。
func (i *utilIdcard) IsValidCard15(idcard string) bool {
	return idcardIsValidCard15(idcard)
}

// IsValidCard18 验证 18 位身份证是否合法：省码、出生日期与
// ISO 7064 MOD 11-2 校验位。与 hutool isValidCard18(idcard)（ignoreCase=true）一致，
// 校验位 x/X 均接受；hutool 的严格大写重载 isValidCard18(idcard, false) 不移植。
func (i *utilIdcard) IsValidCard18(idcard string) bool {
	return idcardIsValidCard18(idcard)
}

// IsValidTWCard 验证台湾身份证号码（10 位，首字母换算 + 加权求和 + 模 10 校验）。
func (i *utilIdcard) IsValidTWCard(idcard string) bool {
	return idcardIsValidTW(idcard)
}

// IsValidHKCard 验证香港身份证号码（可能存在部分特殊证件无法校验，与 hutool 一致）。
func (i *utilIdcard) IsValidHKCard(idcard string) bool {
	return idcardIsValidHK(idcard)
}

/* ---------------- 15/18 位互转 ---------------- */

// Convert15To18 将 15 位身份证号转换为 18 位：出生年份补足 4 位并在末尾
// 追加 ISO 7064 MOD 11-2 校验位。长度非 15、含非数字或出生日期非法时返回 error
// （hutool 对应场景返回 null 或抛异常，Go 端统一以 error 表达）。
func (i *utilIdcard) Convert15To18(idcard string) (string, error) {
	return idcardConvert15To18(idcard)
}

// Convert18To15 将 18 位身份证号转换为 15 位（去掉 2 位年份世纪位与末位校验码）。
// 号码非法时返回 error（hutool 返回原串，Go 端改为显式报错）。
func (i *utilIdcard) Convert18To15(idcard string) (string, error) {
	if strings.TrimSpace(idcard) != "" && idcardIsValidCard18(idcard) {
		return idcard[0:6] + idcard[8:17], nil
	}
	return "", errors.New("非法的 18 位身份证号码")
}

/* ---------------- 信息提取 ---------------- */

// GetBirthByIdCard 根据身份证号获取生日（yyyyMMdd，15 位自动升位后提取）。
// 无法解析返回空串。
func (i *utilIdcard) GetBirthByIdCard(idcard string) string {
	s, ok := idcardNormalize18(idcard)
	if !ok {
		return ""
	}
	return s[6:14]
}

// GetBirthDate 从身份证号中获取出生日期。无法解析时返回零值 time.Time{}。
func (i *utilIdcard) GetBirthDate(idcard string) time.Time {
	birth := i.GetBirthByIdCard(idcard)
	if birth == "" {
		return time.Time{}
	}
	t, err := time.Parse("20060102", birth)
	if err != nil {
		return time.Time{}
	}
	return t
}

// GetAgeByIdCard 根据身份证号计算周岁；now 传零值 time.Time{} 时按当前时间计算。
// 无法解析生日或生日晚于 now 时返回 0（hutool 对后者抛异常，Go 端按 0 处理）。
// 场景：注册时按证件号自动填充年龄。
func (i *utilIdcard) GetAgeByIdCard(idcard string, now time.Time) int {
	birth := i.GetBirthByIdCard(idcard)
	if birth == "" {
		return 0
	}
	b, err := time.Parse("20060102", birth)
	if err != nil {
		return 0
	}
	if now.IsZero() {
		now = time.Now()
	}
	if b.After(now) {
		return 0
	}
	age := now.Year() - b.Year()
	// 未到生日周岁减一
	if now.Month() < b.Month() || (now.Month() == b.Month() && now.Day() < b.Day()) {
		age--
	}
	return age
}

// GetYearByIdCard 获取出生年份（yyyy），无法解析返回 0。
func (i *utilIdcard) GetYearByIdCard(idcard string) int {
	s, ok := idcardNormalize18(idcard)
	if !ok {
		return 0
	}
	year, _ := strconv.Atoi(s[6:10])
	return year
}

// GetMonthByIdCard 获取出生月份（MM），无法解析返回 0。
func (i *utilIdcard) GetMonthByIdCard(idcard string) int {
	s, ok := idcardNormalize18(idcard)
	if !ok {
		return 0
	}
	month, _ := strconv.Atoi(s[10:12])
	return month
}

// GetDayByIdCard 获取出生日（dd），无法解析返回 0。
func (i *utilIdcard) GetDayByIdCard(idcard string) int {
	s, ok := idcardNormalize18(idcard)
	if !ok {
		return 0
	}
	day, _ := strconv.Atoi(s[12:14])
	return day
}

// GetGenderByIdCard 获取性别：1 男 2 女，无法解析（长度非 15/18）返回 0。
// 注意：hutool 返回 1 男 0 女，本移植按更常见的 1/2 约定返回。
func (i *utilIdcard) GetGenderByIdCard(idcard string) int {
	s, ok := idcardNormalize18(idcard)
	if !ok {
		return 0
	}
	// 第 17 位（下标 16）顺序码奇数分配给男性，偶数分配给女性
	if s[16]%2 != 0 {
		return 1
	}
	return 2
}

// GetProvinceCodeByIdCard 获取户籍省份编码：前 2 位；新版外国人永久居留身份证
// 以 9 开头，此时取第 2~3 位受理地代码。仅支持 15/18 位，其它返回空串。
func (i *utilIdcard) GetProvinceCodeByIdCard(idcard string) string {
	if len(idcard) == 15 || len(idcard) == 18 {
		if idcard[0] == '9' {
			return idcard[1:3]
		}
		return idcard[0:2]
	}
	return ""
}

// GetProvinceByIdCard 获取户籍省份名称（如 "广东"），编码未知或长度非法返回空串。
func (i *utilIdcard) GetProvinceByIdCard(idcard string) string {
	code := i.GetProvinceCodeByIdCard(idcard)
	if code == "" {
		return ""
	}
	return idcardProvinceCodes[code]
}

// GetCityCodeByIdCard 获取地市级编码（前 4 位）。仅支持 15/18 位，其它返回空串。
func (i *utilIdcard) GetCityCodeByIdCard(idcard string) string {
	if len(idcard) == 15 || len(idcard) == 18 {
		return idcard[0:4]
	}
	return ""
}

// GetDistrictCodeByIdCard 获取区县级编码（前 6 位）。仅支持 15/18 位，其它返回空串。
func (i *utilIdcard) GetDistrictCodeByIdCard(idcard string) string {
	if len(idcard) == 15 || len(idcard) == 18 {
		return idcard[0:6]
	}
	return ""
}

// Hide 将身份证号 [front, end) 区间的字符替换为 '*'（按字符计，委托 String.Hide，
// 支持负下标语义同 String.Sub）。Hide("110101199003077854", 6, 14) = "110101********7854"。
// 场景：证件号展示打码。
func (i *utilIdcard) Hide(idcard string, front, end int) string {
	return String.Hide(idcard, front, end)
}

// GetIdcardInfo 解析身份证号为结构化信息（省份、城市、区县、性别、周岁、生日）。
// 无法解析（长度非 15/18 或出生日期非法）返回 nil。
func (i *utilIdcard) GetIdcardInfo(idcard string) *IdcardInfo {
	s, ok := idcardNormalize18(idcard)
	if !ok {
		return nil
	}
	if _, err := time.Parse("20060102", s[6:14]); err != nil {
		return nil
	}
	return &IdcardInfo{
		Province: i.GetProvinceByIdCard(idcard),
		City:     i.GetCityCodeByIdCard(idcard),
		District: i.GetDistrictCodeByIdCard(idcard),
		Gender:   i.GetGenderByIdCard(idcard),
		Age:      i.GetAgeByIdCard(idcard, time.Time{}),
		Birthday: s[6:14],
	}
}

/* ---------------- 内部辅助 ---------------- */

// idcardNormalize18 归一化为 18 位：15 位升位后返回，18 位原样返回，
// 其余长度视为无法解析（hutool 对超长串会越界截取，Go 端统一按不支持处理）。
func idcardNormalize18(idcard string) (string, bool) {
	switch len(idcard) {
	case 15:
		s, err := idcardConvert15To18(idcard)
		if err != nil {
			return "", false
		}
		return s, true
	case 18:
		return idcard, true
	}
	return "", false
}

// idcardIsValidCard18 判断 18 位身份证合法性（GB 11643-1999）：
// 六位地址码 + 八位出生日期码 + 三位顺序码 + 一位校验码。
func idcardIsValidCard18(idcard string) bool {
	if len(idcard) != 18 {
		return false
	}
	// 省份代码：新版外国人永久居留身份证以 9 开头，第二三位为受理地代码
	proCode := idcard[0:2]
	if idcard[0] == '9' {
		proCode = idcard[1:3]
	}
	if _, ok := idcardProvinceCodes[proCode]; !ok {
		return false
	}
	// 校验出生日期
	if !validIsBirthdayStr(idcard[6:14]) {
		return false
	}
	code17 := idcard[0:17]
	if !idcardDigitsRe.MatchString(code17) {
		return false
	}
	val := idcardCheckCode18(code17)
	// 校验位 x/X 均接受（等价 hutool ignoreCase=true）
	return val == idcard[17] || (val == 'X' && idcard[17] == 'x')
}

// idcardIsValidCard15 验证 15 位身份编码：全数字、省码存在、
// 两位年份按 19XX 补全后生日合法
func idcardIsValidCard15(idcard string) bool {
	if len(idcard) != 15 || !idcardDigitsRe.MatchString(idcard) {
		return false
	}
	if _, ok := idcardProvinceCodes[idcard[0:2]]; !ok {
		return false
	}
	return validIsBirthdayStr("19" + idcard[6:12])
}

// idcardIsValidCard10 校验 10 位港澳台证件号（对照 IdcardUtil.isValidCard10）。
// recognized 表示证件形态可识别（台湾/澳门/香港），valid 表示算法校验通过。
func idcardIsValidCard10(idcard string) (recognized, valid bool) {
	if strings.TrimSpace(idcard) == "" {
		return false, false
	}
	// 中文括号统一替换为英文括号（hutool issue#IBP6T1）
	idcard = strings.ReplaceAll(idcard, "（", "(")
	idcard = strings.ReplaceAll(idcard, "）", ")")
	card := strings.ReplaceAll(strings.ReplaceAll(idcard, "(", ""), ")", "")
	if len(card) != 8 && len(card) != 9 && len(idcard) != 10 {
		return false, false
	}
	switch {
	case idcardTwFormRe.MatchString(idcard): // 台湾
		if idcard[1] != '1' && idcard[1] != '2' {
			// 第 2 位性别码只能为 1（男）或 2（女）
			return true, false
		}
		return true, idcardIsValidTW(idcard)
	case idcardMoFormRe.MatchString(idcard): // 澳门：无校验位算法，形态合法即视为有效
		return true, true
	case idcardHkFormRe.MatchString(idcard): // 香港
		return true, idcardIsValidHK(idcard)
	default:
		return false, false
	}
}

// idcardIsValidTW 验证台湾身份证号码：首字母换算为两位数（十位 + 个位*9），
// 中间 8 位分别乘 8~1，校验位 = (10 - 和 mod 10) mod 10。
// 与 hutool 的差异：hutool 对非数字中间位抛 NumberFormatException，此处返回 false。
func idcardIsValidTW(idcard string) bool {
	if len(idcard) != 10 {
		return false
	}
	iStart, ok := idcardTwFirstCode[idcard[0]]
	if !ok {
		return false
	}
	sum := iStart/10 + (iStart%10)*9
	iflag := 8
	for k := 1; k <= 8; k++ {
		c := idcard[k]
		if c < '0' || c > '9' {
			return false
		}
		sum += int(c-'0') * iflag
		iflag--
	}
	end := idcard[9]
	if end < '0' || end > '9' {
		return false
	}
	remainder := sum % 10
	if remainder == 0 {
		remainder = 10
	}
	return 10-remainder == int(end-'0')
}

// idcardHkLetterValue 香港身份证字母换算：A-Z 分别对应 10-35（返回 -55 后的值）
func idcardHkLetterValue(c byte) int {
	if c >= 'a' && c <= 'z' {
		c -= 'a' - 'A'
	}
	return int(c) - 55
}

// idcardIsValidHK 验证香港身份证号码（存在部分特殊证件无法校验的已知缺陷，与 hutool 一致）：
// 首两位字母换算为 10-35；单字母视为首位空格（空格对应 58，58*9=522）；
// 各位分别乘 9~1 后求和，总和能被 11 整除则有效（校验位 A/a 代表 10）。
func idcardIsValidHK(idcard string) bool {
	if !idcardHkFormRe.MatchString(idcard) {
		return false
	}
	card := strings.ReplaceAll(strings.ReplaceAll(idcard, "(", ""), ")", "")
	sum := 0
	if len(card) == 9 {
		// 两个字母：换算后分别乘 9、8
		sum = idcardHkLetterValue(card[0])*9 + idcardHkLetterValue(card[1])*8
		card = card[1:9]
	} else {
		// 单字母：首位空格对应 58，58 * 9 = 522
		sum = 522 + idcardHkLetterValue(card[0])*8
	}
	// 中间 6 位数字分别乘 7~2
	iflag := 7
	for k := 1; k <= 6; k++ {
		sum += int(card[k]-'0') * iflag
		iflag--
	}
	// 校验位：A/a 代表 10
	if card[7] == 'A' || card[7] == 'a' {
		sum += 10
	} else {
		sum += int(card[7] - '0')
	}
	return sum%11 == 0
}

// idcardConvert15To18 15 位升 18 位核心实现：解析 yyMMdd 补足 4 位年份
// （2000 年之后不存在 15 位身份证号，误入 21 世纪的年份回退 100 年，
// 对照 hutool convert15To18），并追加 MOD 11-2 校验位。
func idcardConvert15To18(idcard string) (string, error) {
	if len(idcard) != 15 {
		return "", errors.New("身份证长度必须为 15 位")
	}
	if !idcardDigitsRe.MatchString(idcard) {
		return "", errors.New("身份证必须全为数字")
	}
	birth, err := time.Parse("060102", idcard[6:12])
	if err != nil {
		return "", errors.New("身份证出生日期非法: " + err.Error())
	}
	sYear := birth.Year()
	if sYear > 2000 {
		// 2000 年之后不存在 15 位身份证号，用于修复两位年份误判
		sYear -= 100
	}
	id18 := idcard[0:6] + strconv.Itoa(sYear) + idcard[8:]
	return id18 + string(idcardCheckCode18(id18)), nil
}

// idcardCheckCode18 计算 18 位身份证的校验码：前 17 位与加权因子逐位相乘求和，
// 按 mod 11 映射到 "10X98765432"（对照 IdcardUtil.getCheckCode18）。入参须为 17 位数字。
func idcardCheckCode18(code17 string) byte {
	sum := 0
	for k := 0; k < len(code17) && k < len(idcardWeight18); k++ {
		sum += int(code17[k]-'0') * idcardWeight18[k]
	}
	switch sum % 11 {
	case 10:
		return '2'
	case 9:
		return '3'
	case 8:
		return '4'
	case 7:
		return '5'
	case 6:
		return '6'
	case 5:
		return '7'
	case 4:
		return '8'
	case 3:
		return '9'
	case 2:
		return 'X'
	case 1:
		return '0'
	default:
		return '1'
	}
}
