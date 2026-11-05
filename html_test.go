package utils

import "testing"

/* ---------------- 转义与还原（Escape / Unescape） ---------------- */

// TestHtmlEscape 覆盖 < > & " ' 与 nbsp(160)，及空串、非 Latin-1 透传
func TestHtmlEscape(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"<", "&lt;"},
		{">", "&gt;"},
		{"&", "&amp;"},
		{`"`, "&quot;"},
		{"'", "&#039;"}, // HTML4 规范不支持 &apos;
		{"\u00A0", "&nbsp;"},
	}
	for _, c := range cases {
		if got := Html.Escape(c.in); got != c.want {
			t.Errorf("Html.Escape(%q) = %q, want %q", c.in, got, c.want)
		}
	}
	// 组合串（对照 hutool HtmlUtilTest.escapeTest 黄金值）
	if got := Html.Escape("<html><body>123'123'</body></html>"); got != "&lt;html&gt;&lt;body&gt;123&#039;123&#039;&lt;/body&gt;&lt;/html&gt;" {
		t.Errorf("Html.Escape 组合串 = %q", got)
	}
	// 原 utils_test.go 断言搬入
	if got := Html.Escape(`<a href="x">`); got != "&lt;a href=&quot;x&quot;&gt;" {
		t.Errorf("Html.Escape(<a href=\"x\">) = %q, want &lt;a href=&quot;x&quot;&gt;", got)
	}
	// nbsp（ASCII 160，对照 hutool escapeTest2）
	if got := Html.Escape("<html><body>\u00A0</body></html>"); got != "&lt;html&gt;&lt;body&gt;&nbsp;&lt;/body&gt;&lt;/html&gt;" {
		t.Errorf("Html.Escape nbsp = %q", got)
	}
	if got := Html.Escape(""); got != "" {
		t.Errorf("Html.Escape(\"\") = %q, want 空串", got)
	}
	// 普通 Latin-1 字符与 >255 字符原样保留
	if got := Html.Escape("你好"); got != "你好" {
		t.Errorf("Html.Escape 中文应原样保留, got %q", got)
	}
	if got := Html.Escape("😀"); got != "😀" {
		t.Errorf("Html.Escape 非 Latin-1 应原样保留, got %q", got)
	}
}

// TestHtmlUnescape 委托 HTML4 全量实体解码（含命名、数字实体）
func TestHtmlUnescape(t *testing.T) {
	if got := Html.Unescape("&lt;b&gt;"); got != "<b>" { // 原 utils_test.go 断言搬入
		t.Errorf("Html.Unescape = %q, want \"<b>\"", got)
	}
	if got := Html.Unescape("&apos;"); got != "'" {
		t.Errorf("Html.Unescape(&apos;) = %q, want \"'\"", got)
	}
	if got := Html.Unescape("&nbsp;"); got != "\u00A0" {
		t.Errorf("Html.Unescape(&nbsp;) = %q, want U+00A0", got)
	}
	if got := Html.Unescape("&amp;eacute;&#20013;"); got != "&eacute;中" {
		t.Errorf("Html.Unescape 混合 = %q, want \"&eacute;中\"", got)
	}
	// hutool escapeTest 黄金值往返
	const esc = "&lt;html&gt;&lt;body&gt;123&#039;123&#039;&lt;/body&gt;&lt;/html&gt;"
	if got := Html.Unescape(esc); got != "<html><body>123'123'</body></html>" {
		t.Errorf("Html.Unescape 往返 = %q", got)
	}
	if got := Html.Unescape(""); got != "" {
		t.Errorf("Html.Unescape(\"\") = %q, want 空串", got)
	}
}

/* ---------------- 标签清除与删除 ---------------- */

// TestHtmlCleanHtmlTag 删除全部标签保留文本，含自闭合标签与中文内容
func TestHtmlCleanHtmlTag(t *testing.T) {
	if got := Html.CleanHtmlTag("<p>你好<b>世界</b></p>"); got != "你好世界" { // 原 utils_test.go 断言搬入
		t.Errorf("CleanHtmlTag = %q, want \"你好世界\"", got)
	}
	if got := Html.CleanHtmlTag("纯文本无标签"); got != "纯文本无标签" {
		t.Errorf("CleanHtmlTag 无标签应原样保留, got %q", got)
	}
	if got := Html.CleanHtmlTag("a<br/>b"); got != "ab" {
		t.Errorf("CleanHtmlTag 自闭合 = %q, want \"ab\"", got)
	}
	// 对照 hutool cleanHtmlTagTest 黄金值
	if got := Html.CleanHtmlTag(`pre<img src="xxx/dfdsfds/test.jpg">`); got != "pre" {
		t.Errorf("CleanHtmlTag img = %q, want \"pre\"", got)
	}
	if got := Html.CleanHtmlTag(`pre<div class="test_div">dfdsfdsfdsf</div>`); got != "predfdsfdsfdsf" {
		t.Errorf("CleanHtmlTag div = %q, want \"predfdsfdsfdsf\"", got)
	}
	if got := Html.CleanHtmlTag(""); got != "" {
		t.Errorf("CleanHtmlTag(\"\") = %q, want 空串", got)
	}
}

// TestHtmlRemoveHtmlTag withTagContent 两种语义、多标签名、大小写与无匹配
func TestHtmlRemoveHtmlTag(t *testing.T) {
	// 连同内容删除（原 utils_test.go 断言搬入）
	if got := Html.RemoveHtmlTag("<script>alert(1)</script><p>正文</p>", true, "script"); got != "<p>正文</p>" {
		t.Errorf("RemoveHtmlTag 含内容 = %q, want \"<p>正文</p>\"", got)
	}
	if got := Html.RemoveHtmlTag("<div><script>bad()</script>好</div>", true, "script"); got != "<div>好</div>" {
		t.Errorf("RemoveHtmlTag 含内容 = %q, want \"<div>好</div>\"", got)
	}
	// 大小写不敏感
	if got := Html.RemoveHtmlTag("<SCRIPT>alert(1)</SCRIPT>ok", true, "script"); got != "ok" {
		t.Errorf("RemoveHtmlTag 大写标签 = %q, want \"ok\"", got)
	}
	// 带属性的开标签同样连内容删除
	if got := Html.RemoveHtmlTag(`<div><script type="text/javascript">bad()</script>正文</div>`, true, "script"); got != "<div>正文</div>" {
		t.Errorf("RemoveHtmlTag 带属性 = %q, want \"<div>正文</div>\"", got)
	}
	// 仅删标记保留内容
	if got := Html.RemoveHtmlTag("<b>加粗</b>", false, "b"); got != "加粗" {
		t.Errorf("RemoveHtmlTag 保留内容 = %q, want \"加粗\"", got)
	}
	// 多标签名：避免移除 i 却误删 img（对照 hutool unwrapTest2 黄金值）
	if got := Html.RemoveHtmlTag("<html><img src='aaa'><i>测试文本</i></html>", false, "i", "br"); got != "<html><img src='aaa'>测试文本</html>" {
		t.Errorf("RemoveHtmlTag 多标签 = %q", got)
	}
	// 非闭合/自闭合形式（对照 hutool removeHtmlTagTest）
	if got := Html.RemoveHtmlTag(`pre<img src="xxx/dfdsfds/test.jpg">`, false, "img"); got != "pre" {
		t.Errorf("RemoveHtmlTag 非闭合 img = %q, want \"pre\"", got)
	}
	if got := Html.RemoveHtmlTag(`pre<img src="xxx/dfdsfds/test.jpg" />`, false, "img"); got != "pre" {
		t.Errorf("RemoveHtmlTag 自闭合 img = %q, want \"pre\"", got)
	}
	// 无匹配原样保留
	const noMatch = "<p>正文</p>"
	if got := Html.RemoveHtmlTag(noMatch, true, "script"); got != noMatch {
		t.Errorf("RemoveHtmlTag 无匹配应原样保留, got %q", got)
	}
}

// TestHtmlUnwrapHtmlTag 仅去标记保留内容：单标签、多标签、属性含引号
func TestHtmlUnwrapHtmlTag(t *testing.T) {
	if got := Html.UnwrapHtmlTag("<a href='x'>链接</a>", "a"); got != "链接" { // 原 utils_test.go 断言搬入
		t.Errorf("UnwrapHtmlTag = %q, want \"链接\"", got)
	}
	if got := Html.UnwrapHtmlTag(`<a href="x" target="_blank">链接</a>`, "a"); got != "链接" {
		t.Errorf("UnwrapHtmlTag 双引号属性 = %q, want \"链接\"", got)
	}
	if got := Html.UnwrapHtmlTag("<b>加粗</b><i>斜体</i>", "b", "i"); got != "加粗斜体" {
		t.Errorf("UnwrapHtmlTag 多标签 = %q, want \"加粗斜体\"", got)
	}
	// 保留标签间空白（对照 hutool unwrapHtmlTagTest 黄金值）
	if got := Html.UnwrapHtmlTag("pre<div class=\"test_div\">\r\n\t\tabc\r\n</div>", "div"); got != "pre\r\n\t\tabc\r\n" {
		t.Errorf("UnwrapHtmlTag 换行 = %q", got)
	}
	const noMatch = "<p>x</p>"
	if got := Html.UnwrapHtmlTag(noMatch, "b"); got != noMatch {
		t.Errorf("UnwrapHtmlTag 无匹配应原样保留, got %q", got)
	}
}

// TestHtmlCleanEmptyTag 空标签对删除（对照源码正则语义与 hutool cleanEmptyTagTest）
func TestHtmlCleanEmptyTag(t *testing.T) {
	cases := []struct {
		in   string
		want string
		note string
	}{
		{"<p></p><div></div>", "", "连续空对"},
		{"<p>TEXT</p><div></div>", "<p>TEXT</p>", "空 div"},
		{"<p></p><div>TEXT</div>", "<div>TEXT</div>", "空 p"},
		{"<p>TEXT</p><div>TEXT</div>", "<p>TEXT</p><div>TEXT</div>", "非空对应保留"},
		{"TEXT<p></p><div></div>TEXT", "TEXTTEXT", "夹在文本间"},
		{"<p> </p>", "", "含空白空对"},
		{"<div>\t\n</div>", "", "含制表换行空对"},
	}
	for _, c := range cases {
		if got := Html.CleanEmptyTag(c.in); got != c.want {
			t.Errorf("CleanEmptyTag %s(%q) = %q, want %q", c.note, c.in, got, c.want)
		}
	}
	// 开/闭标签名不一致（RE2 无反向引用，运行时校验）：整体不删除
	if got := Html.CleanEmptyTag("<p></div>"); got != "<p></div>" {
		t.Errorf("CleanEmptyTag 标签名不匹配应原样保留, got %q", got)
	}
	// 嵌套空对：只剥最内层，外层因文本被占而不再是空对
	if got := Html.CleanEmptyTag("<div><p></p></div>"); got != "<div></div>" {
		t.Errorf("CleanEmptyTag 嵌套 = %q, want \"<div></div>\"", got)
	}
}

/* ---------------- 属性删除 ---------------- */

// TestHtmlRemoveHtmlAttr 单属性、多属性、引号形式与无匹配
func TestHtmlRemoveHtmlAttr(t *testing.T) {
	// 双引号包裹（对照 hutool removeHtmlAttrTest）
	if got := Html.RemoveHtmlAttr(`<div class="test_div"></div><span class="test_div"></span>`, "class"); got != "<div></div><span></span>" {
		t.Errorf("RemoveHtmlAttr 双引号 = %q", got)
	}
	// 裸值 / 单引号 / 属性后跟空格（对照 hutool 黄金值）
	if got := Html.RemoveHtmlAttr(`<div class=test_div></div><span Class='test_div' ></span>`, "class"); got != "<div></div><span></span>" {
		t.Errorf("RemoveHtmlAttr 裸值单引号 = %q", got)
	}
	// 属性位于其它属性之后 / 之前（对照 hutool 黄金值）
	if got := Html.RemoveHtmlAttr(`<div style="margin:100%" class=test_div></div><span Class='test_div' width=100></span>`, "class"); got != `<div style="margin:100%"></div><span width=100></span>` {
		t.Errorf("RemoveHtmlAttr 位置混合 = %q", got)
	}
	// 属性名与等号间有空格（对照 hutool 黄金值）
	if got := Html.RemoveHtmlAttr(`<div style = "margin:100%" class = test_div></div><span Class = 'test_div' width=100></span>`, "class"); got != `<div style = "margin:100%"></div><span width=100></span>` {
		t.Errorf("RemoveHtmlAttr 等号空格 = %q", got)
	}
	// 空属性值
	if got := Html.RemoveHtmlAttr(`<img alt="" src="a.png">`, "alt"); got != `<img src="a.png">` {
		t.Errorf("RemoveHtmlAttr 空值 = %q", got)
	}
	// 一次删多个属性
	if got := Html.RemoveHtmlAttr(`<div class="a" style="b" onclick="c()">x</div>`, "class", "onclick"); got != `<div style="b">x</div>` {
		t.Errorf("RemoveHtmlAttr 多属性 = %q", got)
	}
	// 无该属性原样保留
	if got := Html.RemoveHtmlAttr(`<img src="a.png">`, "onclick"); got != `<img src="a.png">` {
		t.Errorf("RemoveHtmlAttr 无匹配应原样保留, got %q", got)
	}
	// 属性值含空格（引号内，对照 hutool issueI6YNTFTest）
	if got := Html.RemoveHtmlAttr(`<html><body><div class="a1 a2">hello world</div></body></html>`, "class"); got != "<html><body><div>hello world</div></body></html>" {
		t.Errorf("RemoveHtmlAttr 值含空格 = %q", got)
	}
}

// TestHtmlRemoveAllHtmlAttr 指定标签属性全删、其它标签不受影响
func TestHtmlRemoveAllHtmlAttr(t *testing.T) {
	// 对照 hutool removeAllHtmlAttrTest 黄金值
	if got := Html.RemoveAllHtmlAttr(`<div class="test_div" width="120"></div>`, "div"); got != "<div></div>" {
		t.Errorf("RemoveAllHtmlAttr div = %q, want \"<div></div>\"", got)
	}
	if got := Html.RemoveAllHtmlAttr(`<a href="x" target="_blank">链接</a>`, "a"); got != "<a>链接</a>" {
		t.Errorf("RemoveAllHtmlAttr a = %q, want \"<a>链接</a>\"", got)
	}
	// 其它标签不受影响
	if got := Html.RemoveAllHtmlAttr(`<a href="x"><img src="y"></a>`, "a"); got != `<a><img src="y"></a>` {
		t.Errorf("RemoveAllHtmlAttr 其它标签不应受影响, got %q", got)
	}
	// 多标签名
	if got := Html.RemoveAllHtmlAttr(`<a href="1"><b class="x">t</b></a>`, "a", "b"); got != "<a><b>t</b></a>" {
		t.Errorf("RemoveAllHtmlAttr 多标签 = %q, want \"<a><b>t</b></a>\"", got)
	}
	if got := Html.RemoveAllHtmlAttr("", "div"); got != "" {
		t.Errorf("RemoveAllHtmlAttr(\"\") = %q, want 空串", got)
	}
}

/* ---------------- 中文内容混合标签综合场景 ---------------- */

// TestHtmlChineseMixed 场景组合：转义、清洗、去属性、解转义往返
func TestHtmlChineseMixed(t *testing.T) {
	// 提取正文（CleanHtmlTag 只剥标签，不解读实体）
	if got := Html.CleanHtmlTag(`<div class="article">标题</div><p>正文 &amp;要点</p>`); got != "标题正文 &amp;要点" {
		t.Errorf("中文混合提取 = %q", got)
	}
	// 去 script 保正文，再清标签
	dirty := `<div><script>恶意代码</script><b>安全</b>内容</div>`
	cleaned := Html.CleanHtmlTag(Html.RemoveHtmlTag(dirty, true, "script"))
	if cleaned != "安全内容" {
		t.Errorf("中文混合去 script = %q, want \"安全内容\"", cleaned)
	}
	// 转义后还原恒等
	const text = `<p class="t">"引号" & '单引号' 中文</p>`
	if got := Html.Unescape(Html.Escape(text)); got != text {
		t.Errorf("Html.Escape→Unescape 往返应恒等, got %q", got)
	}
}
