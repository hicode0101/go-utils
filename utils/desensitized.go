package utils

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// utilDesensitized 脱敏工具（对照 hutool DesensitizedUtil）：
// 姓名、证件号、电话、地址、邮箱、密码、车牌、银行卡、IP、护照、信用代码等打码。
//
// 所有打码以 '*' 填充，长度按 Unicode 字符（rune）计，中文安全；
// 输入为空白串时统一返回空串。
type utilDesensitized struct{}

/* ---------------- 专项脱敏 ---------------- */

// ChineseName 中文姓名脱敏：仅保留第一个字符，其余打码。
// ChineseName("张三") = "张*"，ChineseName("段正淳") = "段**"。场景：用户昵称/姓名展示。
func (d *utilDesensitized) ChineseName(fullName string) string {
	return d.FirstMask(fullName)
}

// FirstMask 首字符保留、其余全部打码（hutool first_mask 规则）；
// 单个字符原样返回，空白串返回空串。FirstMask("123456789") = "1********"。
func (d *utilDesensitized) FirstMask(s string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	return String.Hide(s, 1, utf8.RuneCountInString(s))
}

// IdCardNum 身份证号脱敏：保留前 front 位与后 end 位，中间打码（位数从 1 起计）。
// front+end 超过号码长度或任一为负时返回空串。
// IdCardNum("51343620000320711X", 1, 2) = "5***************1X"。
func (d *utilDesensitized) IdCardNum(num string, front, end int) string {
	if strings.TrimSpace(num) == "" {
		return ""
	}
	n := utf8.RuneCountInString(num)
	if front < 0 || end < 0 || front+end > n {
		return ""
	}
	return String.Hide(num, front, n-end)
}

// FixedPhone 固定电话脱敏：保留前 4 位与后 2 位。
// FixedPhone("09157518479") = "0915*****79"。
func (d *utilDesensitized) FixedPhone(num string) string {
	if strings.TrimSpace(num) == "" {
		return ""
	}
	n := utf8.RuneCountInString(num)
	return String.Hide(num, 4, n-2)
}

// MobilePhone 手机号脱敏：保留前 3 位与后 4 位。
// MobilePhone("13812341234") = "138****1234"。
func (d *utilDesensitized) MobilePhone(num string) string {
	if strings.TrimSpace(num) == "" {
		return ""
	}
	n := utf8.RuneCountInString(num)
	return String.Hide(num, 3, n-4)
}

// Address 地址脱敏：尾部打码 sensitiveSize 个字符，只显示到地区不显示详细地址。
// Address("北京市海淀区马连洼街道289号", 8) = "北京市海淀区马********"。
func (d *utilDesensitized) Address(addr string, sensitiveSize int) string {
	if strings.TrimSpace(addr) == "" {
		return ""
	}
	n := utf8.RuneCountInString(addr)
	return String.Hide(addr, n-sensitiveSize, n)
}

// Email 邮箱脱敏：@ 前仅保留第一个字符，@ 及域名原样保留；
// 无 @ 或前缀不超过 1 个字符时原样返回。
// Email("duandazhi-jack@gmail.com.cn") = "d*************@gmail.com.cn"。
func (d *utilDesensitized) Email(email string) string {
	if strings.TrimSpace(email) == "" {
		return ""
	}
	// 按字符定位 @（本地部分可能含中文等多字节字符）
	index := -1
	for i, r := range []rune(email) {
		if r == '@' {
			index = i
			break
		}
	}
	if index <= 1 {
		return email
	}
	return String.Hide(email, 1, index)
}

// Password 密码脱敏：每个字符替换为 '*'，星号数量与原密码长度一致
// （对照 hutool password：StrUtil.repeat('*', length)，并非固定 8 个）。
// Password("123456") = "******"。场景：日志中隐藏明文口令。
func (d *utilDesensitized) Password(password string) string {
	if strings.TrimSpace(password) == "" {
		return ""
	}
	return strings.Repeat("*", utf8.RuneCountInString(password))
}

// CarLicense 中国车牌脱敏：中间打码（普通 7 位牌打第 4~6 位，
// 新能源 8 位牌打第 4~7 位），其它长度视为非法车牌原样返回。
// CarLicense("苏D40000") = "苏D4***0"，CarLicense("陕A12345D") = "陕A1****D"。
func (d *utilDesensitized) CarLicense(license string) string {
	if strings.TrimSpace(license) == "" {
		return ""
	}
	switch utf8.RuneCountInString(license) {
	case 7: // 普通车牌
		return String.Hide(license, 3, 6)
	case 8: // 新能源车牌
		return String.Hide(license, 3, 7)
	}
	return license
}

// BankCard 银行卡号脱敏：先去除全部空白字符，保留前 4 位与末尾 1~4 位，
// 中间按每 4 位一组打码并以空格分隔；去除空白后不足 9 位原样返回。
// BankCard("11011111222233333256") = "1101 **** **** **** 3256"。
func (d *utilDesensitized) BankCard(no string) string {
	if strings.TrimSpace(no) == "" {
		return no
	}
	no = desensitizedCleanBlank(no)
	runes := []rune(no)
	length := len(runes)
	if length < 9 {
		return no
	}
	endLength := length % 4 // 末尾保留 1~4 位（能整除时保留 4 位）
	if endLength == 0 {
		endLength = 4
	}
	midLength := length - 4 - endLength
	var b strings.Builder
	b.Grow(length + midLength/4 + 2)
	b.WriteString(string(runes[:4]))
	for i := 0; i < midLength; i++ {
		if i%4 == 0 {
			b.WriteByte(' ')
		}
		b.WriteByte('*')
	}
	b.WriteByte(' ')
	b.WriteString(string(runes[length-endLength:]))
	return b.String()
}

// Ipv4 IPv4 脱敏：保留第一段，其余段打码。
// Ipv4("192.168.1.1") = "192.*.*.*"。无分隔符时原样追加打码段。
func (d *utilDesensitized) Ipv4(ip string) string {
	return desensitizedSubBefore(ip, '.') + ".*.*.*"
}

// Ipv6 IPv6 脱敏：保留第一段，其余段打码。
// Ipv6("2001:0db8:86a3:08d3:1319:8a2e:0370:7344") = "2001:*:*:*:*:*:*:*"。
func (d *utilDesensitized) Ipv6(ip string) string {
	return desensitizedSubBefore(ip, ':') + ":*:*:*:*:*:*:*"
}

// Passport 护照号脱敏：保留前 2 位与后 2 位，中间打码；长度不足等于 2 位时整体打码。
// Passport("PJ1234567") = "PJ*****67"。
func (d *utilDesensitized) Passport(passport string) string {
	if strings.TrimSpace(passport) == "" {
		return passport
	}
	n := utf8.RuneCountInString(passport)
	if n <= 2 {
		return String.Hide(passport, 0, n)
	}
	return String.Hide(passport, 2, n-2)
}

// CreditCode 统一社会信用代码脱敏：保留前 4 位与后 4 位，中间打码；
// 长度不足等于 4 位时整体打码。
// CreditCode("91110108MA01ABCDE7") = "9111**********CDE7"。
func (d *utilDesensitized) CreditCode(code string) string {
	if strings.TrimSpace(code) == "" {
		return code
	}
	n := utf8.RuneCountInString(code)
	if n <= 4 {
		return String.Hide(code, 0, n)
	}
	return String.Hide(code, 4, n-4)
}

/* ---------------- 占位与清空 ---------------- */

// UserId 用户 ID 脱敏占位：不对外提供 userId，固定返回 0（对照 hutool userId()）。
func (d *utilDesensitized) UserId() int64 {
	return 0
}

// Clear 清空：恒返回空串（对照 hutool clearToEmpty）。场景：高敏字段整体不可见。
func (d *utilDesensitized) Clear(s string) string {
	return ""
}

/* ---------------- 类型分发的统一入口 ---------------- */

// Desensitized 按 typeName 分发到对应脱敏策略的统一入口（对照 hutool
// DesensitizedUtil.desensitized(str, DesensitizedType)，类型名大小写不敏感）：
// CHINESE_NAME / ID_CARD（保留前 1 后 2）/ FIXED_PHONE / MOBILE_PHONE /
// ADDRESS（打码 8 位）/ EMAIL / PASSWORD / CAR_LICENSE / BANK_CARD / IPV4 /
// IPV6 / PASSPORT / CREDIT_CODE / USER_ID（返回 "0"）/ FIRST_MASK /
// CLEAR（含 CLEAR_TO_EMPTY、CLEAR_TO_NULL，返回空串）。
// 未知类型返回原串；输入为空白串时统一返回空串。
// 场景：注解/配置驱动的字段脱敏，按字段上声明的类型名动态选择策略。
func (d *utilDesensitized) Desensitized(s, typeName string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	switch strings.ToUpper(strings.TrimSpace(typeName)) {
	case "USER_ID":
		return strconv.FormatInt(d.UserId(), 10)
	case "CHINESE_NAME":
		return d.ChineseName(s)
	case "ID_CARD":
		return d.IdCardNum(s, 1, 2)
	case "FIXED_PHONE":
		return d.FixedPhone(s)
	case "MOBILE_PHONE":
		return d.MobilePhone(s)
	case "ADDRESS":
		return d.Address(s, 8)
	case "EMAIL":
		return d.Email(s)
	case "PASSWORD":
		return d.Password(s)
	case "CAR_LICENSE":
		return d.CarLicense(s)
	case "BANK_CARD":
		return d.BankCard(s)
	case "IPV4":
		return d.Ipv4(s)
	case "IPV6":
		return d.Ipv6(s)
	case "PASSPORT":
		return d.Passport(s)
	case "CREDIT_CODE":
		return d.CreditCode(s)
	case "FIRST_MASK":
		return d.FirstMask(s)
	case "CLEAR", "CLEAR_TO_EMPTY", "CLEAR_TO_NULL":
		return d.Clear(s)
	default:
		return s
	}
}

/* ---------------- 内部辅助 ---------------- */

// desensitizedCleanBlank 去除全部空白字符（对照 hutool StrUtil.cleanBlank，
// 含半角/全角空格与 BOM）
func desensitizedCleanBlank(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\t', '\n', '\r', '\v', '\f', 12288, 65279:
			return -1
		}
		return r
	}, s)
}

// desensitizedSubBefore 取首个 sep 之前的部分（对照 hutool subBefore：
// 未找到分隔符返回原串，分隔符位于开头返回空串）
func desensitizedSubBefore(s string, sep rune) string {
	for i, r := range []rune(s) {
		if r == sep {
			return string([]rune(s)[:i])
		}
	}
	return s
}
