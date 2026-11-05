package utils

import "testing"

/* ---------------- XML 转义（EscapeXml / UnescapeXml） ---------------- */

// TestEscapeXmlFiveChars 逐一验证五个 XML 特殊字符的转义
func TestEscapeXmlFiveChars(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"&", "&amp;"},
		{"<", "&lt;"},
		{">", "&gt;"},
		{`"`, "&quot;"},
		{"'", "&apos;"},
	}
	for _, c := range cases {
		if got := Escape.EscapeXml(c.in); got != c.want {
			t.Errorf("EscapeXml(%q) = %q, want %q", c.in, got, c.want)
		}
	}
	// 表外字符原样保留（对照 hutool escapeXmlTest：中文弯引号不是 XML 特殊字符）
	if got := Escape.EscapeXml("中文“双引号”"); got != "中文“双引号”" {
		t.Errorf("EscapeXml 中文弯引号应原样保留, got %q", got)
	}
}

// TestEscapeXmlRoundTrip 组合串转义与往返恒等（原 utils_test.go TestEscapeHtml 断言搬入）
func TestEscapeXmlRoundTrip(t *testing.T) {
	const in = `<a & "'>`
	want := "&lt;a &amp; &quot;&apos;&gt;"
	if got := Escape.EscapeXml(in); got != want {
		t.Errorf("EscapeXml(%q) = %q, want %q", in, got, want)
	}
	// 对照 hutool escapeXmlTest
	if got := Escape.EscapeXml("<>"); got != "&lt;&gt;" {
		t.Errorf("EscapeXml(<>) = %q, want &lt;&gt;", got)
	}
	if got := Escape.UnescapeXml("&lt;a&gt;&amp;&quot;&apos;&lt;/a&gt;"); got != `<a>&"'</a>` {
		t.Errorf("UnescapeXml = %q, want `<a>&\"'</a>`", got)
	}
	if got := Escape.UnescapeXml(Escape.EscapeXml(`<x a="1" b='2'>-&</x>`)); got != `<x a="1" b='2'>-&</x>` {
		t.Errorf("EscapeXml→UnescapeXml 往返应恒等, got %q", got)
	}
}

// TestUnescapeXmlNumericEntities 数字实体解码与非法实体保留（error 路径）
func TestUnescapeXmlNumericEntities(t *testing.T) {
	// 十进制与十六进制（x 小写）
	if got := Escape.UnescapeXml("&#39;&#x27;"); got != "''" {
		t.Errorf("UnescapeXml 数字实体 = %q, want \"''\"", got)
	}
	if got := Escape.UnescapeXml("&#65;"); got != "A" {
		t.Errorf("UnescapeXml(&#65;) = %q, want \"A\"", got)
	}
	// 非法实体一律原样保留
	cases := []struct {
		in   string
		note string
	}{
		{"&unknown;", "未知命名实体"},
		{"&#123", "缺右分号"},
		{"&#;", "无数字"},
		{"&#xZZ;", "非十六进制"},
	}
	for _, c := range cases {
		if got := Escape.UnescapeXml(c.in); got != c.in {
			t.Errorf("UnescapeXml %s(%q) = %q, 应原样保留", c.note, c.in, got)
		}
	}
	if got := Escape.UnescapeXml(""); got != "" {
		t.Errorf("UnescapeXml(\"\") = %q, want 空串", got)
	}
}

/* ---------------- HTML4 实体转义（EscapeHtml4 / UnescapeHtml4） ---------------- */

// TestEscapeHtml4Basics 基础实体与黄金值（对照 hutool EscapeUtilTest.escapeHtml4Test）
func TestEscapeHtml4Basics(t *testing.T) {
	if got := Escape.EscapeHtml4("&<>\""); got != "&amp;&lt;&gt;&quot;" {
		t.Errorf("EscapeHtml4 基础实体 = %q", got)
	}
	// hutool 黄金值
	if got := Escape.EscapeHtml4("<a>你好</a>"); got != "&lt;a&gt;你好&lt;/a&gt;" {
		t.Errorf("EscapeHtml4 = %q, want &lt;a&gt;你好&lt;/a&gt;", got)
	}
	if got := Escape.EscapeHtml4("*@-_+./(123你好)"); got != "*@-_+./(123你好)" {
		t.Errorf("EscapeHtml4 表外字符应原样保留, got %q", got)
	}
	// HTML4 转义不处理单引号（hutool escapeSingleQuotesTest）
	const sq = "'some text with single quotes'"
	if got := Escape.EscapeHtml4(sq); got != sq {
		t.Errorf("EscapeHtml4 单引号不应转义, got %q", got)
	}
	if got := Escape.EscapeHtml4(""); got != "" {
		t.Errorf("EscapeHtml4 空串应返回空串")
	}
}

// TestEscapeHtml4NamedEntities 非 ASCII 命名实体，逐一按 251 项表抽验
func TestEscapeHtml4NamedEntities(t *testing.T) {
	cases := []struct {
		in   rune
		want string
	}{
		{0xE9, "&eacute;"},  // é ISO8859-1 区段
		{0xA9, "&copy;"},    // ©
		{0xA0, "&nbsp;"},    // 不间断空格
		{0x20AC, "&euro;"},  // € 扩展区
		{0x2211, "&sum;"},   // ∑ 数学符号
		{0x3B1, "&alpha;"},  // α 希腊字母
		{0x2014, "&mdash;"}, // — 长破折号
		{0x2122, "&trade;"}, // ™
		{0x4E2D, "中"},       // 汉字不在表内，原样保留
	}
	for _, c := range cases {
		if got := Escape.EscapeHtml4(string(c.in)); got != c.want {
			t.Errorf("EscapeHtml4(%q U+%04X) = %q, want %q", string(c.in), c.in, got, c.want)
		}
	}
}

// TestUnescapeHtml4Entities 命名实体与数字实体反转义
func TestUnescapeHtml4Entities(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"&eacute;", "é"}, // 命名实体
		{"&copy;", "©"},   // 版权符号
		{"&alpha;", "α"},  // 希腊字母
		{"&apos;", "'"},   // 仅反转义方向支持（hutool unescapeSingleQuotesTest）
		{"&#20013;", "中"}, // 十进制
		{"&#x56FD;", "国"}, // 十六进制
		{"&#x56fd;", "国"}, // 小写十六进制位
		{"&#25391;&#33633;&#22120;&#31867;&#22411;", "振荡器类型"}, // hutool 黄金值
	}
	for _, c := range cases {
		if got := Escape.UnescapeHtml4(c.in); got != c.want {
			t.Errorf("UnescapeHtml4(%q) = %q, want %q", c.in, got, c.want)
		}
	}
	// 混合串（原 utils_test.go 断言搬入）
	if got := Escape.UnescapeHtml4("&lt;p&gt;&amp;nbsp;&lt;/p&gt;"); got != "<p>&nbsp;</p>" {
		t.Errorf("UnescapeHtml4 混合串 = %q, want \"<p>&nbsp;</p>\"", got)
	}
	if got := Escape.UnescapeHtml4("&#20013;&#x56FD;"); got != "中国" {
		t.Errorf("UnescapeHtml4 混合数字实体 = %q, want \"中国\"", got)
	}
	// 单次解码不递归：&amp;eacute; 只解出 &，保留 eacute; 原文
	if got := Escape.UnescapeHtml4("&amp;eacute;"); got != "&eacute;" {
		t.Errorf("UnescapeHtml4 单次解码 = %q, want \"&eacute;\"", got)
	}
	if got := Escape.UnescapeHtml4(""); got != "" {
		t.Errorf("UnescapeHtml4 空串应返回空串")
	}
}

// TestUnescapeHtml4UnknownEntity 未知实体原样保留（error 路径）
func TestUnescapeHtml4UnknownEntity(t *testing.T) {
	cases := []string{"&foobar;", "&nope", "&#", "&#x", "&notreal;"}
	for _, in := range cases {
		if got := Escape.UnescapeHtml4(in); got != in {
			t.Errorf("UnescapeHtml4(%q) = %q, 未知实体应原样保留", in, got)
		}
	}
	// 数字实体值超出 16 位时截断为低 16 位（对应 Java char 强转）：128512 & 0xFFFF = 0xF600
	if got := Escape.UnescapeHtml4("&#128512;"); got != "\uf600" {
		t.Errorf("UnescapeHtml4 溢出截断 = %q, want U+F600", got)
	}
}

// TestEscapeHtml4RoundTrip 往返恒等：UnescapeHtml4(EscapeHtml4(x)) == x
func TestEscapeHtml4RoundTrip(t *testing.T) {
	const in = "© € ∑ α é — ™ <a>你好</a> \"x\" &y"
	if got := Escape.UnescapeHtml4(Escape.EscapeHtml4(in)); got != in {
		t.Errorf("EscapeHtml4→UnescapeHtml4 往返应恒等, got %q", got)
	}
}

/* ---------------- Unicode 转义（EscapeUnicode / UnescapeUnicode） ---------------- */

// TestEscapeUnicode 中英混合、空串、纯 ASCII、控制字符与 emoji 代理对
func TestEscapeUnicode(t *testing.T) {
	if got := Escape.EscapeUnicode("中a"); got != "\\u4e2da" { // 原 utils_test.go 断言搬入
		t.Errorf("EscapeUnicode(\"中a\") = %q, want \"\\\\u4e2da\"", got)
	}
	if got := Escape.EscapeUnicode("中文"); got != "\\u4e2d\\u6587" {
		t.Errorf("EscapeUnicode(\"中文\") = %q, want \"\\\\u4e2d\\\\u6587\"", got)
	}
	if got := Escape.EscapeUnicode(""); got != "" {
		t.Errorf("EscapeUnicode(\"\") = %q, want 空串", got)
	}
	const ascii = "abc XYZ 123!~"
	if got := Escape.EscapeUnicode(ascii); got != ascii {
		t.Errorf("EscapeUnicode 纯 ASCII 应原样保留, got %q", got)
	}
	if got := Escape.EscapeUnicode("\t"); got != "\\u0009" {
		t.Errorf("EscapeUnicode 控制字符 = %q, want \"\\\\u0009\"", got)
	}
	// 增补平面字符按 UTF-16 代理对拆成两个 \uXXXX（😀 = U+1F600）
	if got := Escape.EscapeUnicode("😀"); got != "\\ud83d\\ude00" {
		t.Errorf("EscapeUnicode emoji = %q, want \"\\\\ud83d\\\\ude00\"", got)
	}
}

// TestUnescapeUnicode 解码、大小写标识、非法序列保留与代理对合并
func TestUnescapeUnicode(t *testing.T) {
	if got := Escape.UnescapeUnicode("\\u4e2d\\u6587"); got != "中文" { // 原 utils_test.go 断言搬入
		t.Errorf("UnescapeUnicode = %q, want \"中文\"", got)
	}
	// 标识 \u 不区分大小写，十六进制位亦不区分大小写
	if got := Escape.UnescapeUnicode("\\U4E2D"); got != "中" {
		t.Errorf("UnescapeUnicode 大写 \\U = %q, want \"中\"", got)
	}
	if got := Escape.UnescapeUnicode("\\u4E2D"); got != "中" {
		t.Errorf("UnescapeUnicode 大写十六进制 = %q, want \"中\"", got)
	}
	// 非法序列原样保留
	cases := []struct {
		in   string
		note string
	}{
		{"\\u12g4", "非十六进制位"},
		{"ab\\u12", "不足 4 位"},
		{"ab\\u123", "恰好 3 位"},
	}
	for _, c := range cases {
		if got := Escape.UnescapeUnicode(c.in); got != c.in {
			t.Errorf("UnescapeUnicode %s(%q) = %q, 应原样保留", c.note, c.in, got)
		}
	}
	if got := Escape.UnescapeUnicode(""); got != "" {
		t.Errorf("UnescapeUnicode(\"\") = %q, want 空串", got)
	}
	// 高/低代理对合并回一个字符
	if got := Escape.UnescapeUnicode("\\ud83d\\ude00"); got != "😀" {
		t.Errorf("UnescapeUnicode emoji = %q, want 😀", got)
	}
	// 孤立高代理无法单独表示，写成 U+FFFD
	if got := Escape.UnescapeUnicode("\\ud83d"); got != "\uFFFD" {
		t.Errorf("UnescapeUnicode 孤立代理 = %q, want U+FFFD", got)
	}
	// 混合内容
	if got := Escape.UnescapeUnicode("Hello\\u4e16\\u754C!"); got != "Hello世界!" {
		t.Errorf("UnescapeUnicode 混合 = %q, want \"Hello世界!\"", got)
	}
}

// TestEscapeUnicodeRoundTrip 往返恒等
func TestEscapeUnicodeRoundTrip(t *testing.T) {
	const in = "中a😀\t 世界 English 123"
	if got := Escape.UnescapeUnicode(Escape.EscapeUnicode(in)); got != in {
		t.Errorf("EscapeUnicode→UnescapeUnicode 往返应恒等, got %q", got)
	}
}
