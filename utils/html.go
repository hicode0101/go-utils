package utils

import (
	"regexp"
	"strings"
)

// utilHtml HTML 工具（对照 hutool HtmlUtil）。
//
// 面向爬虫与富文本清洗场景：HTML 转义/还原、清除全部标签、按标签名删除
// （可选连同内容）、删除空标签对、删除标签属性等。
type utilHtml struct{}

/* ---------------- 预编译正则（包初始化时完成） ---------------- */

var (
	// htmlTagRe 匹配任意 HTML 标签（对照 hutool RE_HTML_MARK）：
	// 普通开标签 <a href="x">、闭标签 </a>（含 < /a > 形式）、自闭合标签 <br/>
	htmlTagRe = regexp.MustCompile(`(<[^<]*?>)|(<[\s]*?/[^<]*?>)|(<[^<]*?/[\s]*?>)`)

	// htmlEmptyTagRe 匹配空标签对（对照 hutool RE_HTML_EMPTY_MARK "<(\w+)([^>]*)>\s*</\1>"）。
	// Go 正则（RE2）不支持反向引用 \1，故分别捕获开/闭标签名，运行时比较两者是否一致
	htmlEmptyTagRe = regexp.MustCompile(`(?i)<(\w+)([^>]*)>\s*</(\w+)>`)

	// htmlAttrTailSpaceRe 属性删除后标签内残留的尾部空白（对照 hutool
	// removeHtmlAttr 收尾清理正则 "\s+(>|/>)" -> "$1"）
	htmlAttrTailSpaceRe = regexp.MustCompile(`\s+(>|/>)`)
)

/* ---------------- 转义与还原 ---------------- */

// Escape 转义文本中的 HTML 特殊字符为安全的字符实体（对照 hutool HtmlUtil.escape）：
// ' 替换为 &#039;（HTML4 规范不支持 &apos;）、" 替换为 &quot;、& 替换为 &amp;、
// < 替换为 &lt;、> 替换为 &gt;、不间断空格(ASCII 160) 替换为 &nbsp;，其余字符原样保留。
// 场景：用户输入回显到页面，防 XSS 与排版错乱。
func (h *utilHtml) Escape(text string) string {
	var b strings.Builder
	b.Grow(len(text) + len(text)/4)
	for _, c := range text {
		if c > 255 {
			// 非 Latin-1 字符不参与转义，原样保留
			b.WriteRune(c)
			continue
		}
		switch c {
		case '\'':
			b.WriteString("&#039;")
		case '"':
			b.WriteString("&quot;")
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case 160: // 不间断空格 &nbsp;
			b.WriteString("&nbsp;")
		default:
			b.WriteRune(c)
		}
	}
	return b.String()
}

// Unescape 还原被转义的 HTML 实体，即 HTML4 全量实体解码
// （对照 hutool HtmlUtil.unescape，其内部即 EscapeUtil.unescapeHtml4，此处直接委托）。
// 支持全部 252 项命名实体与 &#123; / &#x1F; 数字实体，非法实体原样保留。
func (h *utilHtml) Unescape(html string) string {
	return Escape.UnescapeHtml4(html)
}

/* ---------------- 标签清除与删除 ---------------- */

// CleanHtmlTag 删除全部 HTML 标签但保留标签内的文本（对照 hutool cleanHtmlTag）。
// CleanHtmlTag("<p>你好<b>世界</b></p>") = "你好世界"。
// 场景：爬虫正文提取、生成纯文本摘要。
func (h *utilHtml) CleanHtmlTag(content string) string {
	return htmlTagRe.ReplaceAllString(content, "")
}

// RemoveHtmlTag 删除指定 HTML 标签（忽略大小写，对照 hutool removeHtmlTag）：
//   - withTagContent=true：连同标签包裹的内容一起删除，如删除 <script> 标签及脚本；
//   - withTagContent=false：仅删除标签标记，保留被包裹的内容。
//
// RemoveHtmlTag("<div><script>bad()</script>好</div>", true, "script") = "<div>好</div>"。
// 场景：过滤广告节点、去除 script/style 块。
func (h *utilHtml) RemoveHtmlTag(content string, withTagContent bool, tagNames ...string) string {
	for _, tagName := range tagNames {
		tagName = strings.TrimSpace(tagName)
		if tagName == "" {
			continue
		}
		name := regexp.QuoteMeta(tagName)
		// 对照 hutool：withTagContent 为 "(?i)<{}(\s+[^>]*?)?/?>(.*?</{}>)?"，
		// 否则为 "(?i)<{}(\s+[^>]*?)?/?>|</?{}>"（/?> 兼容自闭合形式）
		var re *regexp.Regexp
		if withTagContent {
			re = regexp.MustCompile(`(?i)<` + name + `(\s+[^>]*?)?/?>(.*?</` + name + `>)?`)
		} else {
			re = regexp.MustCompile(`(?i)<` + name + `(\s+[^>]*?)?/?>|</?` + name + `>`)
		}
		content = re.ReplaceAllString(content, "")
	}
	return content
}

// UnwrapHtmlTag 仅删除指定标签的标记而保留被包裹的内容（对照 hutool
// unwrapHtmlTag，忽略大小写）。UnwrapHtmlTag("<b>加粗</b>", "b") = "加粗"。
// 场景：去除加粗/斜体等排版标签，保留正文。
func (h *utilHtml) UnwrapHtmlTag(content string, tagNames ...string) string {
	return h.RemoveHtmlTag(content, false, tagNames...)
}

// CleanEmptyTag 删除空标签对（忽略大小写，对照 hutool cleanEmptyTag），
// 如 <p></p>、<div> </div>（两标签之间允许空白）。
// 场景：富文本清洗，去掉编辑器产生的空段落。
func (h *utilHtml) CleanEmptyTag(content string) string {
	return htmlEmptyTagRe.ReplaceAllStringFunc(content, func(tag string) string {
		// RE2 不支持反向引用，手工校验开/闭标签名一致（忽略大小写）才删除
		sub := htmlEmptyTagRe.FindStringSubmatch(tag)
		if len(sub) == 4 && strings.EqualFold(sub[1], sub[3]) {
			return ""
		}
		return tag
	})
}

/* ---------------- 属性删除 ---------------- */

// RemoveHtmlAttr 从所有标签上删除指定属性（忽略大小写，对照 hutool removeHtmlAttr）。
// 属性值支持 "双引号包裹" 与裸值两种形式，删除后自动清理标签内残留的尾部空白。
// RemoveHtmlAttr(`<img src="a.png" onclick="x()">`, "onclick") = `<img src="a.png">`。
// 场景：页面净化，去除 onclick/onerror 等事件属性防 XSS。
func (h *utilHtml) RemoveHtmlAttr(content string, attrNames ...string) string {
	for _, attr := range attrNames {
		attr = strings.TrimSpace(attr)
		if attr == "" {
			continue
		}
		// 对照 hutool "(?i)(\s*{}\s*=\s*)(\"[^\"]+?\")|([^>]+?\s*(?=\s|>))"：
		// Go 无前瞻，借由 [^>\s]+ 在空白或 > 处天然截断，语义等价
		re := regexp.MustCompile(`(?i)\s*` + regexp.QuoteMeta(attr) + `\s*=\s*(?:"[^"]*"|[^>\s]+)`)
		content = re.ReplaceAllString(content, "")
	}
	// 清理删除后残留的尾部空白（对照 hutool "\s+(>|/>)" -> "$1"）
	return htmlAttrTailSpaceRe.ReplaceAllString(content, "$1")
}

// RemoveAllHtmlAttr 删除指定标签上的全部属性（忽略大小写，对照 hutool removeAllHtmlAttr）。
// RemoveAllHtmlAttr(`<a href="x" target="_blank">链接</a>`, "a") = "<a>链接</a>"。
// 场景：仅保留标签结构，剥离全部属性。
func (h *utilHtml) RemoveAllHtmlAttr(content string, tagNames ...string) string {
	for _, tagName := range tagNames {
		if tagName == "" {
			continue
		}
		// 对照 hutool "(?i)<{}[^>]*?>" 替换为 "<{}>"
		re := regexp.MustCompile(`(?i)<` + regexp.QuoteMeta(tagName) + `[^>]*?>`)
		content = re.ReplaceAllLiteralString(content, "<"+tagName+">")
	}
	return content
}
