package utils

import (
	"crypto/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// utilValid 字段校验器（对照 hutool Validator，正则取自 RegexPool/PatternPool）。
//
// 全部方法返回 bool，可直接参与条件判断；hutool 的 validateXXX 抛异常式校验不移植。
// Java 端 Pattern.matches 为整串匹配，Go 端正则统一以 ^(?:...)$ 锚定等价实现。
type utilValid struct{}

/* ---------------- 正则表（包初始化时统一预编译） ---------------- */

var (
	// 英文字母、数字和下划线（GENERAL）
	validGeneralRe = regexp.MustCompile(`^\w+$`)
	// 中文字、英文字母、数字和下划线（GENERAL_WITH_CHINESE）
	validGeneralWithChineseRe = regexp.MustCompile(`^[\x{4E00}-\x{9FFF}\w]+$`)
	// 单个中文汉字字符类（CHINESE，含扩展 B~F 区，\x{...} 表示增补平面码点）
	validChineseCharClass = `[\x{2E80}-\x{2EFF}\x{2F00}-\x{2FDF}\x{31C0}-\x{31EF}\x{3400}-\x{4DBF}\x{4E00}-\x{9FFF}\x{F900}-\x{FAFF}\x{20000}-\x{2A6DF}\x{2A700}-\x{2B73F}\x{2B740}-\x{2B81F}\x{2B820}-\x{2CEAF}\x{2F800}-\x{2FA1F}]`
	// 全部为汉字（CHINESES）
	validChinesesRe = regexp.MustCompile(`^(?:` + validChineseCharClass + `)+$`)
	// 包含汉字（ReUtil.RE_CHINESES，非锚定查找）
	validHasChineseRe = regexp.MustCompile(validChineseCharClass)
	// 金额（MONEY）
	validMoneyRe = regexp.MustCompile(`^(\d+(?:\.\d+)?)$`)
	// 邮箱（EMAIL，符合 RFC 5322 的宽松规则，正则来自 http://emailregex.com/；
	// 字符集中含反引号，Go 原生字符串无法直接书写，用拼接表达）
	validEmailRe = regexp.MustCompile(`^(?:[a-z0-9!#$%&'*+/=?^_` + "`" + `{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_` + "`" + `{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)]` + ")" + `$`)
	// 大陆手机号（MOBILE，支持可选的 0/86/+86 前缀）
	validMobileRe = regexp.MustCompile(`^(?:0|86|\+86)?1[3-9]\d{9}$`)
	// 生日（BIRTHDAY，分组提取：组 1=年、组 3=月、组 5=日）
	validBirthdayRe = regexp.MustCompile(`^(\d{2,4})([/\-.年]?)(\d{1,2})([/\-.月]?)(\d{1,2})日?$`)
	// 中国邮编（ZIP_CODE，兼容港澳台 999077/999078）
	validZipCodeRe = regexp.MustCompile(`^(?:(0[1-7]|1[0-356]|2[0-7]|3[0-6]|4[0-7]|5[0-7]|6[0-7]|7[0-5]|8[0-9]|9[0-8])\d{4}|99907[78])$`)
	// IPv4（IPV4，每段 0~255）
	validIpv4Re = regexp.MustCompile(`^(25[0-5]|2[0-4]\d|[0-1]?\d?\d)\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)$`)
	// IPv6（IPV6，完整形态、:: 缩写形态与内嵌 IPv4 形态）
	validIpv6Re = regexp.MustCompile(`^(?:(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]+|::(ffff(:0{1,4})?:)?((25[0-5]|(2[0-4]|1?[0-9])?[0-9])\.){3}(25[0-5]|(2[0-4]|1?[0-9])?[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1?[0-9])?[0-9])\.){3}(25[0-5]|(2[0-4]|1?[0-9])?[0-9])))$`)
	// MAC 地址（MAC_ADDRESS：冒号/横线分隔、思科点分、12 位无分隔、0x+ETHER 后缀四种形态）
	validMacRe = regexp.MustCompile(`^(?:((?:[a-fA-F0-9]{1,2}[:-]){5}[a-fA-F0-9]{1,2})|((?:[a-fA-F0-9]{1,4}[.]){2}[a-fA-F0-9]{1,4})|[a-fA-F0-9]{12}|0x(?:\d{12}).+ETHER)$`)
	// 中国车牌号（PLATE_NUMBER，兼容新能源车牌）
	validPlateNumberRe = regexp.MustCompile(`^(([京津沪渝冀豫云辽黑湘皖鲁新苏浙赣鄂桂甘晋蒙陕吉闽贵粤青藏川宁琼使领][A-Z](([0-9]{5}[ABCDEFGHJK])|([ABCDEFGHJKP]([A-HJ-NP-Z0-9])[0-9]{4})))|([京津沪渝冀豫云辽黑湘皖鲁新苏浙赣鄂桂甘晋蒙陕吉闽贵粤青藏川宁琼使领]\d{3}\d{1,3}[领])|([京津沪渝冀豫云辽黑湘皖鲁新苏浙赣鄂桂甘晋蒙陕吉闽贵粤青藏川宁琼使领][A-Z][A-HJ-NP-Z0-9]{4}[A-HJ-NP-Z0-9挂学警港澳使领]))$`)
	// 车架号（CAR_VIN，17 位，排除易混淆的 I/O/Q）
	validCarVinRe = regexp.MustCompile(`^[A-HJ-NPR-Z0-9]{8}[X0-9]([A-HJ-NPR-Z0-9]{3}\d{5}|[A-HJ-NPR-Z0-9]{5}\d{3})$`)
	// 中文姓名（CHINESE_NAME：2~60 位汉字或维吾尔族姓名间隔号 ·，放宽收录生僻姓名用字）
	validChineseNameRe = regexp.MustCompile(`^[\x{3400}-\x{9FFF}·]{2,60}$`)
	// 统一社会信用代码（CREDIT_CODE，18 位，字符集排除易混淆的 I/O/S/V/Z）
	validCreditCodeRe = regexp.MustCompile(`^[0-9A-HJ-NPQRTUWXY]{2}\d{6}[0-9A-HJ-NPQRTUWXY]{10}$`)
)

/* ---------------- 统一社会信用代码校验表（对照 CreditCodeUtil，GB 32100-2015） ---------------- */

var (
	// 校验位加权因子
	validCreditWeight = [17]int{1, 3, 9, 27, 19, 26, 16, 17, 20, 29, 25, 13, 8, 24, 10, 30, 28}
	// 代码字符集（31 个字符，去掉了易混淆的 I/O/S/V/Z）
	validCreditBaseCodes = "0123456789ABCDEFGHJKLMNPQRTUWXY"
)

/* ---------------- 内容与命名格式 ---------------- */

// IsGeneral 验证是否为英文字母、数字和下划线（hutool GENERAL：^\w+$）。
// IsGeneral("abc_123") = true，含中文或中划线为 false。场景：用户名、标识符校验。
func (v *utilValid) IsGeneral(value string) bool {
	return validGeneralRe.MatchString(value)
}

// IsGeneralWithChinese 验证是否为中文字、英文字母、数字和下划线。
// IsGeneralWithChinese("用户_user_01") = true。场景：允许中文的昵称/名称校验。
func (v *utilValid) IsGeneralWithChinese(value string) bool {
	return validGeneralWithChineseRe.MatchString(value)
}

// IsChineseName 验证是否为中文姓名（hutool CHINESE_NAME：2~60 位汉字，
// 允许维吾尔族姓名的间隔号 "·"，如 "霍加阿卜杜拉·麦提喀斯木"；
// 空格、数字、英文句点等均为非法）。场景：实名信息姓名字段校验。
func (v *utilValid) IsChineseName(value string) bool {
	return validChineseNameRe.MatchString(value)
}

// HasChinese 验证是否包含汉字。"abc中国" = true。场景：中文内容粗筛、敏感词预处理。
func (v *utilValid) HasChinese(value string) bool {
	return validHasChineseRe.MatchString(value)
}

// IsChinese 验证是否全部为汉字。"中国" = true，"中国abc" = false。
// 场景：中文姓名/中文名称整体校验。
func (v *utilValid) IsChinese(value string) bool {
	return validChinesesRe.MatchString(value)
}

/* ---------------- 联系方式与号码 ---------------- */

// IsEmail 验证是否为可用邮箱地址（hutool EMAIL 正则，符合 RFC 5322 的宽松规则：
// 允许 jetz_chong@hutool.cn 等形态，域名部分也可为 [1.2.3.4] IP 字面量）。
// 场景：注册/订阅表单校验。
func (v *utilValid) IsEmail(value string) bool {
	return validEmailRe.MatchString(value)
}

// IsMobile 验证是否为中国大陆手机号（hutool MOBILE 正则：1 开头第二位 3~9 共 11 位，
// 支持可选的 0/86/+86 前缀，如 "+8618049531999"）。场景：注册手机号校验。
func (v *utilValid) IsMobile(value string) bool {
	return validMobileRe.MatchString(value)
}

// IsZipCode 验证是否为中国邮政编码（6 位，hutool ZIP_CODE 规则，兼容港澳台
// 特殊编码 999077/999078）。场景：收货地址邮编校验。
func (v *utilValid) IsZipCode(value string) bool {
	return validZipCodeRe.MatchString(value)
}

/* ---------------- 证件与车辆 ---------------- */

// IsCitizenId 验证是否为居民身份证号（对照 hutool Validator.isCitizenId，底层同
// IdcardUtil.isValidCard）：18 位做完整校验（省码 + 出生日期合法性 +
// ISO 7064 MOD 11-2 校验位，末位 x/X 均可），15 位校验省码与出生日期，
// 10 位支持港澳台证件号算法。含空白一律非法。场景：实名认证、证件录入。
func (v *utilValid) IsCitizenId(value string) bool {
	return Idcard.IsValidCard(value)
}

// IsCarVin 验证是否为车架号（车辆识别代号，17 位，形如 LSJA24U62JG269225，
// 排除易混淆的 I/O/Q）。场景：车辆信息录入、VIN 码解析前置校验。
func (v *utilValid) IsCarVin(value string) bool {
	return validCarVinRe.MatchString(value)
}

// IsPlateNumber 验证是否为中国车牌号（hutool PLATE_NUMBER 正则，
// 兼容新能源车牌与挂/学/警/港澳等特殊号牌）。IsPlateNumber("苏ED12345") = true。
// 场景：车辆信息录入、停车系统车牌识别结果校验。
func (v *utilValid) IsPlateNumber(value string) bool {
	return validPlateNumberRe.MatchString(value)
}

// IsCreditCode 是否为有效的统一社会信用代码（GB 32100-2015，对照 CreditCodeUtil）：
// 18 位 = 登记管理部门 1 位 + 机构类别 1 位 + 行政区划码 6 位 + 主体标识码 9 位 + 校验码 1 位，
// 先按正则校验字符集，再用 ISO 7064 MOD 31-2 算法验证校验位。
// 场景：企业资质信息录入、工商数据清洗。
func (v *utilValid) IsCreditCode(creditCode string) bool {
	if strings.TrimSpace(creditCode) == "" || !validCreditCodeRe.MatchString(creditCode) {
		return false
	}
	parityBit := validCreditParityBit(creditCode)
	if parityBit < 0 {
		return false
	}
	return creditCode[17] == validCreditBaseCodes[parityBit]
}

// RandomCreditCode 随机生成一个合法的统一社会信用代码：前 2 位与主体标识码取自
// 完整字符集，行政区划码 6 位为随机数字，末位按 MOD 31-2 算法补上校验位
// （对照 CreditCodeUtil.randomCreditCode）。场景：测试数据构造、演示数据填充。
func (v *utilValid) RandomCreditCode() string {
	var b strings.Builder
	b.Grow(18)
	for i := 0; i < 2; i++ { // 登记管理部门代码 + 机构类别代码
		b.WriteByte(validCreditBaseCodes[validRandInt(len(validCreditBaseCodes))])
	}
	for i := 2; i < 8; i++ { // 登记管理机关行政区划码（6 位数字）
		b.WriteByte(validCreditBaseCodes[validRandInt(10)])
	}
	for i := 8; i < 17; i++ { // 主体标识码（组织机构代码，9 位）
		b.WriteByte(validCreditBaseCodes[validRandInt(len(validCreditBaseCodes))])
	}
	code := b.String()
	return code + string(validCreditBaseCodes[validCreditParityBit(code)])
}

/* ---------------- 网络与数值 ---------------- */

// IsIpv4 验证是否为 IPv4 地址（每段 0~255）。IsIpv4("192.168.1.1") = true。
// 场景：日志分析、访问控制列表校验。
func (v *utilValid) IsIpv4(value string) bool {
	return validIpv4Re.MatchString(value)
}

// IsIpv6 验证是否为 IPv6 地址（完整形态、:: 缩写与内嵌 IPv4 形态均可）。
// IsIpv6("::1") = true。场景：双栈网络环境下的地址校验。
func (v *utilValid) IsIpv6(value string) bool {
	return validIpv6Re.MatchString(value)
}

// IsMac 验证是否为 MAC 地址，支持 "AA:BB:CC:DD:EE:FF"、"AA-BB-CC-DD-EE-FF"、
// 思科点分 "AABB.CCDD.EEFF"、12 位无分隔与 "0x000000000001+ETHER" 形态。
// 场景：设备绑定、ARP 表解析校验。
func (v *utilValid) IsMac(value string) bool {
	return validMacRe.MatchString(value)
}

// IsMoney 验证是否为金额格式（hutool MONEY：整数或小数，如 "123"、"12.34"）。
// 不含千分位与货币符号。场景：表单金额字段格式预校验。
func (v *utilValid) IsMoney(value string) bool {
	return validMoneyRe.MatchString(value)
}

/* ---------------- 生日 ---------------- */

// IsBirthday 验证是否为生日，支持 yyyyMMdd、yyyy-MM-dd、yyyy/MM/dd、yyyy.MM.dd、
// yyyy年M月d日 等写法（分隔符可省略或混用，月/日允许 1~2 位），并校验年月日
// 真实合法性：年份限 1900 年至当前年，含大小月与闰年（2 月闰年最大 29 日）。
// 注意：两位年份（如 "990101"）无法填满 4 位年的分组，按 hutool 行为一律非法。
// 场景：表单生日字段校验、身份证出生日期段校验。
func (v *utilValid) IsBirthday(value string) bool {
	return validIsBirthdayStr(value)
}

// validIsBirthdayStr 按生日正则提取年月日后做真实性校验（对照 Validator.isBirthday(CharSequence)）
func validIsBirthdayStr(value string) bool {
	m := validBirthdayRe.FindStringSubmatch(value)
	if m == nil {
		return false
	}
	year, _ := strconv.Atoi(m[1])
	month, _ := strconv.Atoi(m[3])
	day, _ := strconv.Atoi(m[5])
	return validIsBirthday(year, month, day)
}

// validIsBirthday 校验年月日是否为合法生日（对照 Validator.isBirthday(int, int, int)）
func validIsBirthday(year, month, day int) bool {
	// 验证年：从 1900 年开始，不超过当前年
	if year < 1900 || year > time.Now().Year() {
		return false
	}
	// 验证月
	if month < 1 || month > 12 {
		return false
	}
	// 验证日
	if day < 1 || day > 31 {
		return false
	}
	// 4/6/9/11 月没有 31 日
	if day == 31 && (month == 4 || month == 6 || month == 9 || month == 11) {
		return false
	}
	if month == 2 {
		// 2 月：非闰年最大 28 日，闰年最大 29 日
		return day < 29 || (day == 29 && validIsLeapYear(year))
	}
	return true
}

// validIsLeapYear 公历闰年判断：4 年一闰，百年不闰，四百年再闰
func validIsLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

/* ---------------- 内部辅助 ---------------- */

// validCreditParityBit 计算统一社会信用代码校验位在字符集中的下标
// （ISO 7064 MOD 31-2，对照 CreditCodeUtil.getParityBit），字符非法返回 -1。
func validCreditParityBit(creditCode string) int {
	sum := 0
	for i := 0; i < 17; i++ {
		idx := strings.IndexByte(validCreditBaseCodes, creditCode[i])
		if idx < 0 {
			return -1
		}
		sum += idx * validCreditWeight[i]
	}
	result := 31 - sum%31
	if result == 31 {
		return 0
	}
	return result
}

// validRandInt 返回 [0, n) 内的密码学安全随机整数（拒绝采样保证等概率），n<=1 时返回 0。
func validRandInt(n int) int {
	if n <= 1 {
		return 0
	}
	// 256 内能被 n 整除的最大倍数，超出该值的随机字节直接丢弃
	max := byte(256 - 256%n)
	buf := make([]byte, 1)
	for {
		if _, err := rand.Read(buf); err != nil {
			// 系统熵源不可用（极罕见），返回 0 保证方法总能给出合法结果
			return 0
		}
		if buf[0] < max {
			return int(buf[0]) % n
		}
	}
}
