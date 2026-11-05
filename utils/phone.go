package utils

import (
	"regexp"
)

// utilPhone 电话号码工具（对照 hutool PhoneUtil）：
// 大陆/港澳台手机号与座机、400/800 电话的校验，号段隐藏与截取。
type utilPhone struct{}

/* ---------------- 正则表（包初始化时统一预编译） ---------------- */

var (
	// 大陆手机号（MOBILE，支持可选的 0/86/+86 前缀）
	phoneMobileRe = regexp.MustCompile(`^(?:0|86|\+86)?1[3-9]\d{9}$`)
	// 香港手机号（MOBILE_HK：8 位数字，可带 0/852/+852 前缀）
	phoneMobileHkRe = regexp.MustCompile(`^(?:0|852|\+852)?\d{8}$`)
	// 台湾手机号（MOBILE_TW：09 开头 + 8 位数字，可带前缀与 "-"）
	phoneMobileTwRe = regexp.MustCompile(`^(?:0|886|\+886)?(?:|-)09\d{8}$`)
	// 澳门手机号（MOBILE_MO：6 开头 + 7 位数字，可带前缀与 "-"）
	phoneMobileMoRe = regexp.MustCompile(`^(?:0|853|\+853)?(?:|-)6\d{7}$`)
	// 座机（TEL：区号 010/02x/0[3-9]xx + 6~8 位号码，"-" 可选）
	phoneTelRe = regexp.MustCompile(`^(010|02\d|0[3-9]\d{2})-?(\d{6,8})$`)
	// 座机 + 400/800 电话（TEL_400_800）
	phoneTel400800Re = regexp.MustCompile(`^(?:0\d{2,3}[\- ]?[0-9]\d{6,7}|[48]00[\- ]?[0-9]\d{2}[\- ]?\d{4})$`)
	// 座机分组提取（hutool ReUtil.get 为非锚定 find：组 1=区号、组 2=号码）
	phoneTelGroupRe = regexp.MustCompile(`(010|02\d|0[3-9]\d{2})-?(\d{6,8})`)
)

/* ---------------- 号码校验 ---------------- */

// IsMobile 验证是否为中国大陆手机号（hutool MOBILE 正则：1 开头第二位 3~9 共 11 位，
// 支持可选的 0/86/+86 前缀）。场景：注册手机号校验。
func (p *utilPhone) IsMobile(s string) bool {
	return phoneMobileRe.MatchString(s)
}

// IsMobileHk 验证是否为中国香港手机号（8 位数字，可带 0/852/+852 前缀）。
// IsMobileHk("51004810") = true。场景：港澳用户联系方式校验。
func (p *utilPhone) IsMobileHk(s string) bool {
	return phoneMobileHkRe.MatchString(s)
}

// IsMobileTw 验证是否为中国台湾手机号（09 开头共 10 位，可带 0/886/+886 前缀与 "-"）。
// IsMobileTw("0960000000") = true。
func (p *utilPhone) IsMobileTw(s string) bool {
	return phoneMobileTwRe.MatchString(s)
}

// IsMobileMo 验证是否为中国澳门手机号（6 开头共 8 位，可带 0/853/+853 前缀与 "-"）。
// IsMobileMo("68000000") = true。
func (p *utilPhone) IsMobileMo(s string) bool {
	return phoneMobileMoRe.MatchString(s)
}

// IsTel 验证是否为中国大陆座机号（区号 010/02x/0[3-9]xx + 6~8 位号码，"-" 可选）。
// IsTel("01088889999")、IsTel("0571-8888888") 均为 true。场景：企业联系方式校验。
func (p *utilPhone) IsTel(s string) bool {
	return phoneTelRe.MatchString(s)
}

// IsTel400800 验证是否为座机号或 400/800 电话（分隔符 "-" 或空格均可）。
// IsTel400800("400-8801234")、IsTel400800("8009996555") 均为 true。
func (p *utilPhone) IsTel400800(s string) bool {
	return phoneTel400800Re.MatchString(s)
}

// IsPhone 手机号（大陆/港澳台）或座机、400/800 电话任一匹配即视为合法电话
// （对照 hutool PhoneUtil.isPhone）。场景：不区分号码类型的通用联系字段校验。
func (p *utilPhone) IsPhone(s string) bool {
	return p.IsMobile(s) || p.IsTel400800(s) || p.IsMobileHk(s) || p.IsMobileTw(s) || p.IsMobileMo(s)
}

/* ---------------- 号码打码 ---------------- */

// HideBefore 隐藏手机号前 7 位（对照 PhoneUtil.hideBefore）。
// HideBefore("18049531999") = "*******1999"。场景：手机号脱敏展示。
func (p *utilPhone) HideBefore(s string) string {
	return String.Hide(s, 0, 7)
}

// HideBetween 隐藏手机号中间 4 位（对照 PhoneUtil.hideBetween）。
// HideBetween("18049531999") = "180****1999"。
func (p *utilPhone) HideBetween(s string) string {
	return String.Hide(s, 3, 7)
}

// HideAfter 隐藏手机号后 4 位（对照 PhoneUtil.hideAfter）。
// HideAfter("18049531999") = "1804953****"。
func (p *utilPhone) HideAfter(s string) string {
	return String.Hide(s, 7, 11)
}

/* ---------------- 号段截取 ---------------- */

// SubBefore 取手机号前 3 位（号段，对照 PhoneUtil.subBefore）。
// SubBefore("18049531999") = "180"。场景：按运营商号段统计。
func (p *utilPhone) SubBefore(s string) string {
	return String.Sub(s, 0, 3)
}

// SubBetween 取手机号中间 4 位（对照 PhoneUtil.subBetween）。
// SubBetween("18049531999") = "4953"。
func (p *utilPhone) SubBetween(s string) string {
	return String.Sub(s, 3, 7)
}

// SubAfter 取手机号后 4 位（对照 PhoneUtil.subAfter）。
// SubAfter("18049531999") = "1999"。场景：短信验证码尾号核对。
func (p *utilPhone) SubAfter(s string) string {
	return String.Sub(s, 7, 11)
}

// SubTelBefore 取座机号的区号部分（TEL 正则第 1 组，对照 PhoneUtil.subTelBefore）。
// SubTelBefore("0571-88888888") = "0571"；非座机格式返回空串。
func (p *utilPhone) SubTelBefore(s string) string {
	m := phoneTelGroupRe.FindStringSubmatch(s)
	if len(m) < 3 {
		return ""
	}
	return m[1]
}

// SubTelAfter 取座机号的号码部分（TEL 正则第 2 组，对照 PhoneUtil.subTelAfter）。
// SubTelAfter("0571-88888888") = "88888888"；非座机格式返回空串。
func (p *utilPhone) SubTelAfter(s string) string {
	m := phoneTelGroupRe.FindStringSubmatch(s)
	if len(m) < 3 {
		return ""
	}
	return m[2]
}
