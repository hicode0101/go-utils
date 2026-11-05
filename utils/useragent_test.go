package utils

import "testing"

// uaCase 单条 UA 黄金用例。期望值摘自 hutool 官方测试
// hutool-http/src/test/java/cn/hutool/http/useragent/UserAgentUtilTest.java，
// 其余按本模块 useragent.go 识别表语义推导。
type uaCase struct {
	name           string
	ua             string
	browser        string
	browserVersion string
	engine         string
	engineVersion  string
	os             string
	osVersion      string
	platform       string
	mobile         bool
}

// TestUAParseTable 表驱动覆盖 hutool 识别表主要分支：
// 浏览器（Chrome/Firefox/Safari/Edge/IE/微信/企业微信/QQ/UC/小米/安卓原生/Opera/未知）、
// 版本号提取、引擎（Webkit/Gecko/Trident/Presto）、操作系统、平台与移动端判定。
func TestUAParseTable(t *testing.T) {
	cases := []uaCase{
		{
			name:    "Chrome_Windows10",
			ua:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.102 Safari/537.36",
			browser: "Chrome", browserVersion: "70.0.3538.102",
			engine: "Webkit", engineVersion: "537.36",
			os: "Windows 10 or Windows Server 2016", osVersion: "10.0",
			platform: "Windows", mobile: false,
		},
		{
			name:    "Chrome_120_Windows10_版本提取",
			ua:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			browser: "Chrome", browserVersion: "120.0.0.0",
			engine: "Webkit", engineVersion: "537.36",
			os: "Windows 10 or Windows Server 2016", osVersion: "10.0",
			platform: "Windows", mobile: false,
		},
		{
			name:    "Chrome_Windows7_parseDesktopTest",
			ua:      "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/535.1 (KHTML, like Gecko) Chrome/14.0.835.163 Safari/535.1",
			browser: "Chrome", browserVersion: "14.0.835.163",
			engine: "Webkit", engineVersion: "535.1",
			os: "Windows 7 or Windows Server 2008R2", osVersion: "6.1",
			platform: "Windows", mobile: false,
		},
		{
			name:    "Firefox_Windows10_Gecko引擎",
			ua:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:100.0) Gecko/20100101 Firefox/100.0",
			browser: "Firefox", browserVersion: "100.0",
			engine: "Gecko", engineVersion: "20100101",
			os: "Windows 10 or Windows Server 2016", osVersion: "10.0",
			platform: "Windows", mobile: false,
		},
		{
			name:    "Safari_Mac",
			ua:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Safari/605.1.15",
			browser: "Safari", browserVersion: "16.1",
			engine: "Webkit", engineVersion: "605.1.15",
			os: "OSX", osVersion: "10_15_7",
			platform: "Mac", mobile: false,
		},
		{
			name:    "Edge_Edg_Chromium",
			ua:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.69 Safari/537.36 Edg/81.0.416.34",
			browser: "MSEdge", browserVersion: "81.0.416.34",
			engine: "Webkit", engineVersion: "537.36",
			os: "Windows 10 or Windows Server 2016", osVersion: "10.0",
			platform: "Windows", mobile: false,
		},
		{
			name:    "IE8_MSIE_Trident",
			ua:      "Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.1; Trident/4.0)",
			browser: "MSIE", browserVersion: "8.0",
			engine: "Trident", engineVersion: "4.0",
			os: "Windows 7 or Windows Server 2008R2", osVersion: "6.1",
			platform: "Windows", mobile: false,
		},
		{
			name:    "IE11_rv11_Trident",
			ua:      "Mozilla/5.0 (Windows NT 10.0; WOW64; Trident/7.0; rv:11.0) like Gecko",
			browser: "MSIE11", browserVersion: "11.0",
			engine: "Trident", engineVersion: "7.0",
			os: "Windows 10 or Windows Server 2016", osVersion: "10.0",
			platform: "Windows", mobile: false,
		},
		{
			name:    "MicroMessenger_iPhone",
			ua:      "Mozilla/5.0 (iPhone; CPU iPhone OS 11_0 like Mac OS X) AppleWebKit/604.1.38 (KHTML, like Gecko) Mobile/15A372 MicroMessenger/7.0.17(0x17001127) NetType/WIFI Language/zh_CN",
			browser: "MicroMessenger", browserVersion: "7.0.17",
			engine: "Webkit", engineVersion: "604.1.38",
			os: "iPhone", osVersion: "11_0",
			platform: "iPhone", mobile: true,
		},
		{
			name:    "MicroMessenger_iPhone_utils黄金值",
			ua:      "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 MicroMessenger/8.0.20",
			browser: "MicroMessenger", browserVersion: "8.0.20",
			engine: "Webkit", engineVersion: "605.1.15",
			os: "iPhone", osVersion: "16_0",
			platform: "iPhone", mobile: true,
		},
		{
			name:    "wxwork_iPhone_企业微信移动端",
			ua:      "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 wxwork/3.0.31 MicroMessenger/7.0.1 Language/zh",
			browser: "wxwork", browserVersion: "3.0.31",
			engine: "Webkit", engineVersion: "605.1.15",
			os: "iPhone", osVersion: "14_0",
			platform: "iPhone", mobile: true,
		},
		{
			name:    "wxwork_Windows_桌面版",
			ua:      "Mozilla/5.0 (Windows NT 6.2; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.116 Safari/537.36 QBCore/4.0.1326.400 QQBrowser/9.0.2524.400 Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.116 Safari/537.36 wxwork/3.1.10 (MicroMessenger/6.2) WindowsWechat",
			browser: "wxwork", browserVersion: "3.1.10",
			engine: "Webkit", engineVersion: "537.36",
			os: "Windows 10 or Windows Server 2016", osVersion: "10.0",
			platform: "Windows", mobile: false,
		},
		{
			name:    "WindowsWechat_微信电脑端",
			ua:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36 MicroMessenger/8.0.30 WindowsWechat/8.0.30(0x1400003e)",
			browser: "WindowsWechat", browserVersion: "8.0.30",
			engine: "Webkit", engineVersion: "537.36",
			os: "Windows 10 or Windows Server 2016", osVersion: "10.0",
			platform: "Windows", mobile: false,
		},
		{
			name:    "MacWechat_macOS例外非移动端",
			ua:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) MicroMessenger/6.8.0(0x16080000) MacWechat/3.8.7(0x13080710) Safari/605.1.15 NetType/WIFI",
			browser: "MicroMessenger", browserVersion: "6.8.0",
			engine: "Webkit", engineVersion: "605.1.15",
			os: "OSX", osVersion: "10_15_7",
			platform: "Mac", mobile: false,
		},
		{
			name:    "QQBrowser_Android",
			ua:      "User-Agent: MQQBrowser/26 Mozilla/5.0 (Linux; U; Android 2.3.7; zh-cn; MB200 Build/GRJ22; CyanogenMod-7) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1",
			browser: "QQBrowser", browserVersion: "26",
			engine: "Webkit", engineVersion: "533.1",
			os: "Android", osVersion: "2.3.7",
			platform: "Android", mobile: true,
		},
		{
			name:    "QQBrowser_Windows10",
			ua:      "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.71 Safari/537.36 Core/1.94.218.400 QQBrowser/12.1.5496.400",
			browser: "QQBrowser", browserVersion: "12.1.5496.400",
			engine: "Webkit", engineVersion: "537.36",
			os: "Windows 10 or Windows Server 2016", osVersion: "10.0",
			platform: "Windows", mobile: false,
		},
		{
			name:    "UCBrowser_Windows_UBrowser",
			ua:      "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/38.0.2125.122 UBrowser/4.0.3214.0 Safari/537.36",
			browser: "UCBrowser", browserVersion: "4.0.3214.0",
			engine: "Webkit", engineVersion: "537.36",
			os: "Windows 7 or Windows Server 2008R2", osVersion: "6.1",
			platform: "Windows", mobile: false,
		},
		{
			name:    "UCBrowser_iPhone",
			ua:      "Mozilla/5.0 (iPhone; CPU iPhone OS 12_4_1 like Mac OS X; zh-CN) AppleWebKit/537.51.1 (KHTML, like Gecko) Mobile/16G102 UCBrowser/12.7.6.1251 Mobile AliApp(TUnionSDK/0.1.20.3)",
			browser: "UCBrowser", browserVersion: "12.7.6.1251",
			engine: "Webkit", engineVersion: "537.51.1",
			os: "iPhone", osVersion: "12_4_1",
			platform: "iPhone", mobile: true,
		},
		{
			name:    "MobileSafari_iPhone_parseMobileTest",
			ua:      "User-Agent:Mozilla/5.0 (iPhone; U; CPU iPhone OS 4_3_3 like Mac OS X; en-us) AppleWebKit/533.17.9 (KHTML, like Gecko) Version/5.0.2 Mobile/8J2 Safari/6533.18.5",
			browser: "Safari", browserVersion: "5.0.2",
			engine: "Webkit", engineVersion: "533.17.9",
			os: "iPhone", osVersion: "4_3_3",
			platform: "iPhone", mobile: true,
		},
		{
			name:    "AndroidBrowser_Dalvik_无版本",
			ua:      "Dalvik/2.1.0 (Linux; U; Android 9; SM-G950U Build/PPR1.180610.011)",
			browser: "Android Browser", browserVersion: "",
			engine: "Unknown", engineVersion: "",
			os: "Android", osVersion: "9",
			platform: "Android", mobile: true,
		},
		{
			name:    "AndroidBrowser_华为原生",
			ua:      "Mozilla/5.0 (Linux; Android 10; EML-AL00 Build/HUAWEIEML-AL00) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Mobile Safari/537.36",
			browser: "Android Browser", browserVersion: "4.0",
			engine: "Webkit", engineVersion: "537.36",
			os: "Android", osVersion: "10",
			platform: "Android", mobile: true,
		},
		{
			name:    "Opera_经典Presto",
			ua:      "Opera/9.80 (Windows NT 6.1) Presto/2.12.388 Version/12.14",
			browser: "Opera", browserVersion: "9.80",
			engine: "Opera", engineVersion: "9.80",
			os: "Windows 7 or Windows Server 2008R2", osVersion: "6.1",
			platform: "Windows", mobile: false,
		},
		{
			name:    "Presto引擎",
			ua:      "Mozilla/5.0 (Windows NT 6.1; U; en-US) Presto/2.10.229 Version/11.60",
			browser: "Unknown", browserVersion: "",
			engine: "Presto", engineVersion: "2.10.229",
			os: "Windows 7 or Windows Server 2008R2", osVersion: "6.1",
			platform: "Windows", mobile: false,
		},
		{
			name:    "MiuiBrowser_XiaoMi_Android",
			ua:      "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/534.24 (KHTML, like Gecko) Chrome/89.0.4389.116 Safari/534.24 XiaoMi/MiuiBrowser/16.0.18 swan-mibrowser",
			browser: "MiuiBrowser", browserVersion: "16.0.18",
			engine: "Webkit", engineVersion: "534.24",
			os: "Android", osVersion: "11",
			platform: "Android", mobile: true,
		},
		{
			name:    "Chrome_CriOS_iPhone",
			ua:      "Mozilla/5.0 (iPhone; CPU iPhone OS 10_3 like Mac OS X) AppleWebKit/602.1.50 (KHTML, like Gecko) CriOS/56.0.2924.75 Mobile/14E5239e Safari/602.1",
			browser: "Chrome", browserVersion: "56.0.2924.75",
			engine: "Webkit", engineVersion: "602.1.50",
			os: "iPhone", osVersion: "10_3",
			platform: "iPhone", mobile: true,
		},
		{
			name:    "OpenHarmony",
			ua:      "Mozilla/5.0 (Phone; OpenHarmony 4.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36 ArkWeb/4.1.6.1 Mobile",
			browser: "Chrome", browserVersion: "114.0.0.0",
			engine: "Webkit", engineVersion: "537.36",
			os: "Harmony", osVersion: "4.1",
			platform: "Harmony", mobile: true,
		},
		{
			name:    "爬虫全未知",
			ua:      "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
			browser: "Unknown", browserVersion: "",
			engine: "Unknown", engineVersion: "",
			os: "Unknown", osVersion: "",
			platform: "Unknown", mobile: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			info := UserAgent.Parse(tc.ua)
			if info == nil {
				t.Fatalf("解析结果为 nil: %s", tc.ua)
			}
			if info.Browser != tc.browser {
				t.Errorf("Browser = %q, want %q", info.Browser, tc.browser)
			}
			if info.BrowserVersion != tc.browserVersion {
				t.Errorf("BrowserVersion = %q, want %q", info.BrowserVersion, tc.browserVersion)
			}
			if info.Engine != tc.engine {
				t.Errorf("Engine = %q, want %q", info.Engine, tc.engine)
			}
			if info.EngineVersion != tc.engineVersion {
				t.Errorf("EngineVersion = %q, want %q", info.EngineVersion, tc.engineVersion)
			}
			if info.Os != tc.os {
				t.Errorf("Os = %q, want %q", info.Os, tc.os)
			}
			if info.OsVersion != tc.osVersion {
				t.Errorf("OsVersion = %q, want %q", info.OsVersion, tc.osVersion)
			}
			if info.Platform != tc.platform {
				t.Errorf("Platform = %q, want %q", info.Platform, tc.platform)
			}
			if info.Mobile != tc.mobile {
				t.Errorf("Mobile = %v, want %v", info.Mobile, tc.mobile)
			}
		})
	}
}

// TestUAParseEmpty 覆盖 Parse("")→nil（含纯空白）。
func TestUAParseEmpty(t *testing.T) {
	if info := UserAgent.Parse(""); info != nil {
		t.Errorf(`Parse("") 应返回 nil, got %+v`, info)
	}
	if info := UserAgent.Parse("   \t "); info != nil {
		t.Errorf(`Parse(纯空白) 应返回 nil, got %+v`, info)
	}
}

// TestUAIsMobile 覆盖 IsMobile 快捷方法：iPhone 正例、桌面 Chrome/爬虫/空串反例。
func TestUAIsMobile(t *testing.T) {
	if !UserAgent.IsMobile("Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148") {
		t.Error("iPhone UA 应判定为移动端")
	}
	if UserAgent.IsMobile("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36") {
		t.Error("桌面 Chrome UA 不应判定为移动端")
	}
	if UserAgent.IsMobile("Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)") {
		t.Error("爬虫 UA 不应判定为移动端")
	}
	if UserAgent.IsMobile("") {
		t.Error("空 UA 应返回 false")
	}
}
