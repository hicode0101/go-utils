package utils

import (
	"regexp"
)

// utilUserAgent User-Agent 解析工具（utils.go 已注册单例 utils.UserAgent）。
type utilUserAgent struct{}

// UserAgentInfo User-Agent 解析结果。
// 场景：统计访问端分布、按设备类型做服务端适配。
type UserAgentInfo struct {
	Browser        string // 浏览器名称，如 Chrome、MicroMessenger
	BrowserVersion string // 浏览器版本
	Engine         string // 浏览器内核（引擎）名称，如 Webkit、Gecko
	EngineVersion  string // 引擎版本
	Os             string // 操作系统名称，如 Windows 10 or Windows Server 2016、Android
	OsVersion      string // 操作系统版本
	Platform       string // 平台名称，如 iPhone、Windows、Android
	Mobile         bool   // 是否移动端（平台或浏览器为移动端且系统非 macOS）
}

// uaEntry UA 识别表的单个条目。
type uaEntry struct {
	name    string         // 识别后的名称（如 Chrome、Windows）
	pattern *regexp.Regexp // 命中正则（忽略大小写）
	version *regexp.Regexp // 版本提取正则（取第 1 个分组），nil 表示无版本信息
	mobile  bool           // 是否移动平台（仅平台表使用）
}

// uaOtherVersion hutool Browser.Other_Version：通用版本正则（使用时前面拼上名称）。
const uaOtherVersion = `[\/ ]([\d\w\.\-]+)`

// uaMustCompile 编译忽略大小写的正则，表均为常量，编译失败直接 panic 暴露问题。
func uaMustCompile(expr string) *regexp.Regexp {
	return regexp.MustCompile(`(?i)` + expr)
}

// uaNewEntry 构造识别条目：命中正则 + 可选的版本正则。
func uaNewEntry(name, expr, versionExpr string) uaEntry {
	e := uaEntry{name: name, pattern: uaMustCompile(expr)}
	if versionExpr != "" {
		e.version = uaMustCompile(versionExpr)
	}
	return e
}

// uaNewBrowser 构造浏览器条目：版本正则为通用正则时前面拼上名称（对照 hutool Browser 构造器）。
func uaNewBrowser(name, expr, versionExpr string) uaEntry {
	if versionExpr == uaOtherVersion {
		versionExpr = name + uaOtherVersion
	}
	return uaNewEntry(name, expr, versionExpr)
}

// uaNewEngine 构造引擎条目：命中正则为名称本身，版本正则为 名称 + [/- 空格] + 版本
// （对照 hutool Engine 构造器）。
func uaNewEngine(name string) uaEntry {
	return uaNewEntry(name, name, name+`[/\- ]([\d\w.\-]+)`)
}

// uaNewPlatform 构造平台条目（无版本），mobile 标记是否移动平台。
func uaNewPlatform(name, expr string, mobile bool) uaEntry {
	e := uaNewEntry(name, expr, "")
	e.mobile = mobile
	return e
}

// uaBrowsers 浏览器识别表（顺序与 hutool Browser.java 完全一致，特殊浏览器优先判断）。
var uaBrowsers = []uaEntry{
	uaNewBrowser("wxwork", "wxwork", `wxwork/([\d\w.\-]+)`),                                   // 企业微信（含 MicroMessenger，须先于微信判断）
	uaNewBrowser("WindowsWechat", "WindowsWechat", "MicroMessenger"+uaOtherVersion),           // 微信电脑端
	uaNewBrowser("MicroMessenger", "MicroMessenger", uaOtherVersion),                          // 微信
	uaNewBrowser("miniProgram", "miniProgram", uaOtherVersion),                                // 微信小程序
	uaNewBrowser("QQBrowser", "QQBrowser", `QQBrowser/([\d\w.\-]+)`),                          // QQ浏览器
	uaNewBrowser("DingTalk-win", "dingtalk-win", `DingTalk\(([\d\w.\-]+)\)`),                  // 钉钉PC端浏览器
	uaNewBrowser("DingTalk", "DingTalk", `AliApp\(DingTalk/([\d\w.\-]+)\)`),                   // 钉钉内置浏览器
	uaNewBrowser("Alipay", "AlipayClient", `AliApp\(AP/([\d\w.\-]+)\)`),                       // 支付宝内置浏览器
	uaNewBrowser("Taobao", "taobao", `AliApp\(TB/([\d\w.\-]+)\)`),                             // 淘宝内置浏览器
	uaNewBrowser("UCBrowser", "UC?Browser", `UC?Browser/([\d\w.\-]+)`),                        // UC浏览器
	uaNewBrowser("MiuiBrowser", "MiuiBrowser|mibrowser", `MiuiBrowser/([\d\w.\-]+)`),          // 小米浏览器
	uaNewBrowser("Quark", "Quark", uaOtherVersion),                                            // 夸克浏览器
	uaNewBrowser("Lenovo", "SLBrowser", `SLBrowser/([\d\w.\-]+)`),                             // 联想浏览器
	uaNewBrowser("MSEdge", "Edge|Edg", `(?:edge|Edg|EdgA)/([\d\w.\-]+)`),                      // Edge
	uaNewBrowser("Chrome", `chrome|(iphone.*crios.*safari)`, `(?:Chrome|CriOS)/([\d\w.\-]+)`), // Chrome（含 iOS CriOS）
	uaNewBrowser("Firefox", "firefox", uaOtherVersion),                                        // Firefox
	uaNewBrowser("IEMobile", "iemobile", uaOtherVersion),                                      // IE 移动版
	uaNewBrowser("Android Browser", "android", `version/([\d\w.\-]+)`),                        // 安卓原生浏览器
	uaNewBrowser("Safari", "safari", `version/([\d\w.\-]+)`),                                  // Safari
	uaNewBrowser("Opera", "opera", uaOtherVersion),                                            // Opera
	uaNewBrowser("Konqueror", "konqueror", uaOtherVersion),                                    // Konqueror
	uaNewBrowser("PS3", "playstation 3", `([\d\w.\-]+)\)\s*$`),                                // PS3
	uaNewBrowser("PSP", "playstation portable", `([\d\w.\-]+)\)?\s*$`),                        // PSP
	uaNewBrowser("Lotus", "lotus.notes", `Lotus-Notes/([\w.]+)`),                              // Lotus Notes
	uaNewBrowser("Thunderbird", "thunderbird", uaOtherVersion),                                // Thunderbird 邮件客户端
	uaNewBrowser("Netscape", "netscape", uaOtherVersion),                                      // Netscape
	uaNewBrowser("Seamonkey", "seamonkey", uaOtherVersion),                                    // Seamonkey
	uaNewBrowser("Outlook", "microsoft.outlook", uaOtherVersion),                              // Outlook 邮件客户端
	uaNewBrowser("Evolution", "evolution", uaOtherVersion),                                    // Evolution 邮件客户端
	uaNewBrowser("MSIE", "msie", `msie ([\d\w.\-]+)`),                                         // IE 10 及以下
	uaNewBrowser("MSIE11", "rv:11", `rv:([\d\w.\-]+)`),                                        // IE 11
	uaNewBrowser("Gabble", "Gabble", uaOtherVersion),                                          // Gabble（XMPP 客户端）
	uaNewBrowser("Yammer Desktop", "AdobeAir", `([\d\w.\-]+)/Yammer`),                         // Yammer 桌面端
	uaNewBrowser("Yammer Mobile", `Yammer[\s]+([\d\w.\-]+)`, `Yammer[\s]+([\d\w.\-]+)`),       // Yammer 移动端
	// 注：与 hutool 原正则保持一致（match 侧的 \\- 即要求字面反斜杠，为 hutool 原始行为）
	uaNewBrowser("Apache HTTP Client", `Apache\\-HttpClient`, `Apache\-HttpClient/([\d\w.\-]+)`), // Apache HttpClient
	uaNewBrowser("BlackBerry", "BlackBerry", `BlackBerry[\d]+/([\d\w.\-]+)`),                     // 黑莓
	uaNewBrowser("Baidu", "Baidu", `baiduboxapp/([\d\w.\-]+)`),                                   // 百度
}

// uaEngines 引擎识别表（顺序与 hutool Engine.java 一致）。
var uaEngines = []uaEntry{
	uaNewEngine("Trident"),
	uaNewEngine("Webkit"),
	uaNewEngine("Chrome"),
	uaNewEngine("Opera"),
	uaNewEngine("Presto"),
	uaNewEngine("Gecko"),
	uaNewEngine("KHTML"),
	uaNewEngine("Konqueror"),
	uaNewEngine("MIDP"),
}

// uaOSes 操作系统识别表（顺序与 hutool OS.java 一致，含 Windows NT 版本映射）。
var uaOSes = []uaEntry{
	uaNewEntry("Windows 10 or Windows Server 2016", `windows nt 10\.0`, `windows nt (10\.0)`),
	uaNewEntry("Windows 8.1 or Windows Server 2012R2", `windows nt 6\.3`, `windows nt (6\.3)`),
	uaNewEntry("Windows 8 or Windows Server 2012", `windows nt 6\.2`, `windows nt (6\.2)`),
	uaNewEntry("Windows Vista", `windows nt 6\.0`, `windows nt (6\.0)`),
	uaNewEntry("Windows 7 or Windows Server 2008R2", `windows nt 6\.1`, `windows nt (6\.1)`),
	uaNewEntry("Windows 2003", `windows nt 5\.2`, `windows nt (5\.2)`),
	uaNewEntry("Windows XP", `windows nt 5\.1`, `windows nt (5\.1)`),
	uaNewEntry("Windows 2000", `windows nt 5\.0`, `windows nt (5\.0)`),
	uaNewEntry("Windows Phone", `windows (ce|phone|mobile)( os)?`, `windows (?:ce|phone|mobile) (\d+([._]\d+)*)`),
	uaNewEntry("Windows", "windows", ""),
	uaNewEntry("OSX", `os x (\d+)[._](\d+)`, `os x (\d+([._]\d+)*)`),
	uaNewEntry("Android", "Android", `Android (\d+([._]\d+)*)`),
	uaNewEntry("Harmony", "OpenHarmony", `OpenHarmony (\d+([._]\d+)*)`),
	uaNewEntry("Android", `XiaoMi|MI\s+`, `\(X(\d+([._]\d+)*)`),
	uaNewEntry("Linux", "linux", ""),
	uaNewEntry("Wii", "wii", `wii libnup/(\d+([._]\d+)*)`),
	uaNewEntry("PS3", "playstation 3", `playstation 3; (\d+([._]\d+)*)`),
	uaNewEntry("PSP", "playstation portable", `Portable\); (\d+([._]\d+)*)`),
	uaNewEntry("iPad", `\(iPad.*os (\d+)[._](\d+)`, `\(iPad.*os (\d+([._]\d+)*)`),
	uaNewEntry("iPhone", `\(iPhone.*os (\d+)[._](\d+)`, `\(iPhone.*os (\d+([._]\d+)*)`),
	uaNewEntry("YPod", `iPod touch[\s\;]+iPhone.*os (\d+)[._](\d+)`, `iPod touch[\s\;]+iPhone.*os (\d+([._]\d+)*)`),
	uaNewEntry("YPad", `iPad[\s\;]+iPhone.*os (\d+)[._](\d+)`, `iPad[\s\;]+iPhone.*os (\d+([._]\d+)*)`),
	uaNewEntry("YPhone", `iPhone[\s\;]+iPhone.*os (\d+)[._](\d+)`, `iPhone[\s\;]+iPhone.*os (\d+([._]\d+)*)`),
	uaNewEntry("Symbian", `symbian(os)?`, ""),
	uaNewEntry("Darwin", `Darwin/([\d\w.\-]+)`, `Darwin/([\d\w.\-]+)`),
	uaNewEntry("Adobe Air", `AdobeAir/([\d\w.\-]+)`, `AdobeAir/([\d\w.\-]+)`),
	uaNewEntry("Java", `Java[\s]+([\d\w.\-]+)`, `Java[\s]+([\d\w.\-]+)`),
}

// uaPlatforms 平台识别表（顺序与 hutool Platform.java 一致：先移动平台后桌面平台）。
var uaPlatforms = []uaEntry{
	// 移动平台
	uaNewPlatform("Windows Phone", `windows (ce|phone|mobile)( os)?`, true),
	uaNewPlatform("iPad", "ipad", true),
	uaNewPlatform("iPod", "ipod", true),
	uaNewPlatform("iPhone", "iphone", true),
	uaNewPlatform("Android", `XiaoMi|MI\s+`, true),
	uaNewPlatform("Android", "android", true),
	uaNewPlatform("GoogleTV", "googletv", true),
	uaNewPlatform("htcFlyer", "htc_flyer", true),
	uaNewPlatform("Symbian", `symbian(os)?`, true),
	uaNewPlatform("Blackberry", "blackberry", true),
	uaNewPlatform("Harmony", "OpenHarmony", true),
	// 桌面平台
	uaNewPlatform("Windows", "windows", false),
	uaNewPlatform("Mac", `(macintosh|darwin)`, false),
	uaNewPlatform("Linux", "linux", false),
	uaNewPlatform("Wii", "wii", false),
	uaNewPlatform("Playstation", "playstation", false),
	uaNewPlatform("Java", "java", false),
}

// uaMobileBrowsers 移动浏览器名称集合（对照 hutool Browser.isMobile）。
var uaMobileBrowsers = map[string]bool{
	"PSP":             true,
	"Yammer Mobile":   true,
	"Android Browser": true,
	"IEMobile":        true,
	"MicroMessenger":  true,
	"miniProgram":     true,
	"DingTalk":        true,
}

// uaMatchTable 依序遍历识别表返回第一个命中的条目，无命中返回 nil。
func uaMatchTable(table []uaEntry, content string) *uaEntry {
	for i := range table {
		if table[i].pattern.MatchString(content) {
			return &table[i]
		}
	}
	return nil
}

// uaGetVersion 用版本正则提取第 1 个分组，无版本正则或未命中返回空串
// （对照 hutool ReUtil.getGroup1）。
func uaGetVersion(e *uaEntry, content string) string {
	if e.version == nil {
		return ""
	}
	m := e.version.FindStringSubmatch(content)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

// Parse 解析 User-Agent 字符串，空串（或纯空白）返回 nil。
// 识别顺序：浏览器 → 引擎 → 操作系统 → 平台 → 移动端判定；
// 未识别的字段值为 "Unknown"（对照 hutool UserAgentParser）。
func (u *utilUserAgent) Parse(uaStr string) *UserAgentInfo {
	if String.IsBlank(uaStr) {
		return nil
	}
	info := &UserAgentInfo{
		Browser:        "Unknown",
		BrowserVersion: "",
		Engine:         "Unknown",
		EngineVersion:  "",
		Os:             "Unknown",
		OsVersion:      "",
		Platform:       "Unknown",
	}

	// 浏览器
	if e := uaMatchTable(uaBrowsers, uaStr); e != nil {
		info.Browser = e.name
		info.BrowserVersion = uaGetVersion(e, uaStr)
	}

	// 浏览器引擎
	if e := uaMatchTable(uaEngines, uaStr); e != nil {
		info.Engine = e.name
		info.EngineVersion = uaGetVersion(e, uaStr)
	}

	// 操作系统
	if e := uaMatchTable(uaOSes, uaStr); e != nil {
		info.Os = e.name
		info.OsVersion = uaGetVersion(e, uaStr)
	}

	// 平台
	platformMobile := false
	if e := uaMatchTable(uaPlatforms, uaStr); e != nil {
		info.Platform = e.name
		platformMobile = e.mobile
	}

	// 移动端判定：平台或浏览器为移动端，且 OS 非 macOS
	// （macOS 下的微信等客户端不属于移动平台，对照 hutool UserAgentParser 的例外规则）
	if (platformMobile || uaMobileBrowsers[info.Browser]) && info.Os != "OSX" {
		info.Mobile = true
	}
	return info
}

// IsMobile 快捷判断 UA 是否为移动端（空串返回 false）。
// 场景：只需移动端二分判断、不关心完整信息。
func (u *utilUserAgent) IsMobile(uaStr string) bool {
	info := u.Parse(uaStr)
	return info != nil && info.Mobile
}
