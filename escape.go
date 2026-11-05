package utils

import (
	"strconv"
	"strings"
	"unicode/utf16"
)

// utilEscape 转义与反转义工具（对照 hutool EscapeUtil / UnicodeUtil）。
//
// 覆盖三类转义：
//   - XML：& < > " ' 五个特殊字符；
//   - HTML4：251 项命名实体 + &#123; / &#x1F; 数字实体；
//   - Unicode：非 ASCII 字符与不可见字符转 \uXXXX 形式。
type utilEscape struct{}

/* ---------------- XML 转义（对照 hutool XmlEscape / XmlUnescape） ---------------- */

// EscapeXml 转义 XML 五个特殊字符：& < > " ' 分别替换为 &amp; &lt; &gt; &quot; &apos;。
// 场景：拼接 XML 报文、生成 RSS/配置片段时防止特殊字符破坏标签结构。
func (e *utilEscape) EscapeXml(xml string) string {
	var b strings.Builder
	b.Grow(len(xml))
	for _, c := range xml {
		switch c {
		case '\'':
			b.WriteString("&apos;")
		case '"':
			b.WriteString("&quot;")
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		default:
			b.WriteRune(c)
		}
	}
	return b.String()
}

// UnescapeXml 还原 EscapeXml 转义过的五个实体，同时支持数字实体
// （&#39;、&#x27; 等，对照 hutool XmlUnescape 内置的 NumericEntityUnescaper）。
// 非法实体原样保留。场景：解析外部 XML 前的预处理、报文字段还原。
func (e *utilEscape) UnescapeXml(xml string) string {
	return escapeXmlUnescapeTable.decode(xml)
}

/* ---------------- HTML4 实体转义（对照 hutool Html4Escape / Html4Unescape） ---------------- */

// escapeHtml4Entities HTML4 命名实体转义表：字符 -> 实体。
// 完整移植自 hutool Html4Escape（BASIC_ESCAPE + ISO8859_1_ESCAPE + HTML40_EXTENDED_ESCAPE），
// 共 251 项。注意：与 XML 不同，HTML4 转义不处理单引号。
var escapeHtml4Entities = map[rune]string{
	// 基础实体（对照 hutool BASIC_ESCAPE）
	0x22: "&quot;", // " - 双引号
	0x26: "&amp;",  // & - 和号
	0x3C: "&lt;",   // < - 小于号
	0x3E: "&gt;",   // > - 大于号

	// ISO8859-1 区段（对照 hutool ISO8859_1_ESCAPE）
	0xA0: "&nbsp;",
	0xA1: "&iexcl;",
	0xA2: "&cent;",
	0xA3: "&pound;",
	0xA4: "&curren;",
	0xA5: "&yen;",
	0xA6: "&brvbar;",
	0xA7: "&sect;",
	0xA8: "&uml;",
	0xA9: "&copy;",
	0xAA: "&ordf;",
	0xAB: "&laquo;",
	0xAC: "&not;",
	0xAD: "&shy;",
	0xAE: "&reg;",
	0xAF: "&macr;",
	0xB0: "&deg;",
	0xB1: "&plusmn;",
	0xB2: "&sup2;",
	0xB3: "&sup3;",
	0xB4: "&acute;",
	0xB5: "&micro;",
	0xB6: "&para;",
	0xB7: "&middot;",
	0xB8: "&cedil;",
	0xB9: "&sup1;",
	0xBA: "&ordm;",
	0xBB: "&raquo;",
	0xBC: "&frac14;",
	0xBD: "&frac12;",
	0xBE: "&frac34;",
	0xBF: "&iquest;",
	0xC0: "&Agrave;",
	0xC1: "&Aacute;",
	0xC2: "&Acirc;",
	0xC3: "&Atilde;",
	0xC4: "&Auml;",
	0xC5: "&Aring;",
	0xC6: "&AElig;",
	0xC7: "&Ccedil;",
	0xC8: "&Egrave;",
	0xC9: "&Eacute;",
	0xCA: "&Ecirc;",
	0xCB: "&Euml;",
	0xCC: "&Igrave;",
	0xCD: "&Iacute;",
	0xCE: "&Icirc;",
	0xCF: "&Iuml;",
	0xD0: "&ETH;",
	0xD1: "&Ntilde;",
	0xD2: "&Ograve;",
	0xD3: "&Oacute;",
	0xD4: "&Ocirc;",
	0xD5: "&Otilde;",
	0xD6: "&Ouml;",
	0xD7: "&times;",
	0xD8: "&Oslash;",
	0xD9: "&Ugrave;",
	0xDA: "&Uacute;",
	0xDB: "&Ucirc;",
	0xDC: "&Uuml;",
	0xDD: "&Yacute;",
	0xDE: "&THORN;",
	0xDF: "&szlig;",
	0xE0: "&agrave;",
	0xE1: "&aacute;",
	0xE2: "&acirc;",
	0xE3: "&atilde;",
	0xE4: "&auml;",
	0xE5: "&aring;",
	0xE6: "&aelig;",
	0xE7: "&ccedil;",
	0xE8: "&egrave;",
	0xE9: "&eacute;",
	0xEA: "&ecirc;",
	0xEB: "&euml;",
	0xEC: "&igrave;",
	0xED: "&iacute;",
	0xEE: "&icirc;",
	0xEF: "&iuml;",
	0xF0: "&eth;",
	0xF1: "&ntilde;",
	0xF2: "&ograve;",
	0xF3: "&oacute;",
	0xF4: "&ocirc;",
	0xF5: "&otilde;",
	0xF6: "&ouml;",
	0xF7: "&divide;",
	0xF8: "&oslash;",
	0xF9: "&ugrave;",
	0xFA: "&uacute;",
	0xFB: "&ucirc;",
	0xFC: "&uuml;",
	0xFD: "&yacute;",
	0xFE: "&thorn;",
	0xFF: "&yuml;",

	// HTML4 扩展实体（对照 hutool HTML40_EXTENDED_ESCAPE）
	0x0192: "&fnof;", // 拉丁小写 f 带钩（函数符号）
	// 希腊字母
	0x0391: "&Alpha;",
	0x0392: "&Beta;",
	0x0393: "&Gamma;",
	0x0394: "&Delta;",
	0x0395: "&Epsilon;",
	0x0396: "&Zeta;",
	0x0397: "&Eta;",
	0x0398: "&Theta;",
	0x0399: "&Iota;",
	0x039A: "&Kappa;",
	0x039B: "&Lambda;",
	0x039C: "&Mu;",
	0x039D: "&Nu;",
	0x039E: "&Xi;",
	0x039F: "&Omicron;",
	0x03A0: "&Pi;",
	0x03A1: "&Rho;",
	0x03A3: "&Sigma;",
	0x03A4: "&Tau;",
	0x03A5: "&Upsilon;",
	0x03A6: "&Phi;",
	0x03A7: "&Chi;",
	0x03A8: "&Psi;",
	0x03A9: "&Omega;",
	0x03B1: "&alpha;",
	0x03B2: "&beta;",
	0x03B3: "&gamma;",
	0x03B4: "&delta;",
	0x03B5: "&epsilon;",
	0x03B6: "&zeta;",
	0x03B7: "&eta;",
	0x03B8: "&theta;",
	0x03B9: "&iota;",
	0x03BA: "&kappa;",
	0x03BB: "&lambda;",
	0x03BC: "&mu;",
	0x03BD: "&nu;",
	0x03BE: "&xi;",
	0x03BF: "&omicron;",
	0x03C0: "&pi;",
	0x03C1: "&rho;",
	0x03C2: "&sigmaf;", // 词尾 sigma
	0x03C3: "&sigma;",
	0x03C4: "&tau;",
	0x03C5: "&upsilon;",
	0x03C6: "&phi;",
	0x03C7: "&chi;",
	0x03C8: "&psi;",
	0x03C9: "&omega;",
	0x03D1: "&thetasym;",
	0x03D2: "&upsih;",
	0x03D6: "&piv;",
	// 通用标点
	0x2022: "&bull;",
	0x2026: "&hellip;",
	0x2032: "&prime;",
	0x2033: "&Prime;",
	0x203E: "&oline;",
	0x2044: "&frasl;",
	// 字母式符号
	0x2118: "&weierp;",
	0x2111: "&image;",
	0x211C: "&real;",
	0x2122: "&trade;",
	0x2135: "&alefsym;",
	// 箭头
	0x2190: "&larr;",
	0x2191: "&uarr;",
	0x2192: "&rarr;",
	0x2193: "&darr;",
	0x2194: "&harr;",
	0x21B5: "&crarr;",
	0x21D0: "&lArr;",
	0x21D1: "&uArr;",
	0x21D2: "&rArr;",
	0x21D3: "&dArr;",
	0x21D4: "&hArr;",
	// 数学运算符
	0x2200: "&forall;",
	0x2202: "&part;",
	0x2203: "&exist;",
	0x2205: "&empty;",
	0x2207: "&nabla;",
	0x2208: "&isin;",
	0x2209: "&notin;",
	0x220B: "&ni;",
	0x220F: "&prod;",
	0x2211: "&sum;",
	0x2212: "&minus;",
	0x2217: "&lowast;",
	0x221A: "&radic;",
	0x221D: "&prop;",
	0x221E: "&infin;",
	0x2220: "&ang;",
	0x2227: "&and;",
	0x2228: "&or;",
	0x2229: "&cap;",
	0x222A: "&cup;",
	0x222B: "&int;",
	0x2234: "&there4;",
	0x223C: "&sim;",
	0x2245: "&cong;",
	0x2248: "&asymp;",
	0x2260: "&ne;",
	0x2261: "&equiv;",
	0x2264: "&le;",
	0x2265: "&ge;",
	0x2282: "&sub;",
	0x2283: "&sup;",
	0x2286: "&sube;",
	0x2287: "&supe;",
	0x2295: "&oplus;",
	0x2297: "&otimes;",
	0x22A5: "&perp;",
	0x22C5: "&sdot;",
	// 技术符号
	0x2308: "&lceil;",
	0x2309: "&rceil;",
	0x230A: "&lfloor;",
	0x230B: "&rfloor;",
	0x2329: "&lang;",
	0x232A: "&rang;",
	// 几何形状
	0x25CA: "&loz;",
	// 扑克花色
	0x2660: "&spades;",
	0x2663: "&clubs;",
	0x2665: "&hearts;",
	0x2666: "&diams;",
	// 拉丁扩展
	0x0152: "&OElig;",
	0x0153: "&oelig;",
	0x0160: "&Scaron;",
	0x0161: "&scaron;",
	0x0178: "&Yuml;",
	// 变音修饰符
	0x02C6: "&circ;",
	0x02DC: "&tilde;",
	// 空白与引号
	0x2002: "&ensp;",
	0x2003: "&emsp;",
	0x2009: "&thinsp;",
	0x200C: "&zwnj;",
	0x200D: "&zwj;",
	0x200E: "&lrm;",
	0x200F: "&rlm;",
	0x2013: "&ndash;",
	0x2014: "&mdash;",
	0x2018: "&lsquo;",
	0x2019: "&rsquo;",
	0x201A: "&sbquo;",
	0x201C: "&ldquo;",
	0x201D: "&rdquo;",
	0x201E: "&bdquo;",
	0x2020: "&dagger;",
	0x2021: "&Dagger;",
	0x2030: "&permil;",
	0x2039: "&lsaquo;",
	0x203A: "&rsaquo;",
	0x20AC: "&euro;",
}

/* ---------------- 命名实体解码器（对照 hutool LookupReplacer + NumericEntityUnescaper） ---------------- */

// escapeEntityTable 命名实体反转义表：实体 -> 字符。
// minLen/maxLen 为最短/最长实体名长度，解码时按从长到短匹配（与 hutool
// LookupReplacer 一致，避免短实体名误吞长实体名的前缀）。
type escapeEntityTable struct {
	entities map[string]rune
	minLen   int
	maxLen   int
}

// 由转义表生成反转义表的包装（对照 hutool InternalEscapeUtil.invert）
func newEscapeEntityTable(entities map[string]rune) *escapeEntityTable {
	t := &escapeEntityTable{entities: entities, minLen: 1 << 30}
	for name := range entities {
		if l := len(name); l > t.maxLen {
			t.maxLen = l
		} else if l < t.minLen {
			t.minLen = l
		}
	}
	return t
}

// decode 扫描并解码字符串：优先匹配命名实体（从最长 key 开始），
// 其次解析数字实体 &#123; / &#x1F;，均无法匹配的字符原样保留。
func (t *escapeEntityTable) decode(s string) string {
	if s == "" {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		if s[i] == '&' {
			// 命名实体：从最长候选开始尝试
			max := len(s) - i
			if max > t.maxLen {
				max = t.maxLen
			}
			matched := false
			for l := max; l >= t.minLen; l-- {
				if r, ok := t.entities[s[i:i+l]]; ok {
					b.WriteRune(r)
					i += l
					matched = true
					break
				}
			}
			if matched {
				continue
			}
			// 数字实体：&#123; 或 &#x1F;
			if r, size, ok := escapeNumericEntity(s[i:]); ok {
				b.WriteRune(r)
				i += size
				continue
			}
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// escapeNumericEntity 解析以 '&' 开头的数字实体（十进制 &#123; / 十六进制 &#x1F;，
// x 大小写均可），返回解码字符与消耗的字节数。数字后必须紧跟 ';' 才算合法，
// 值溢出或格式非法返回 ok=false（对照 hutool NumericEntityUnescaper；
// 值超出 16 位时截断为低 16 位，对应 Java 的 char 强转）。
func escapeNumericEntity(s string) (rune, int, bool) {
	if len(s) < 3 || s[1] != '#' {
		return 0, 0, false
	}
	start, base := 2, 10
	if s[2] == 'x' || s[2] == 'X' {
		start, base = 3, 16
	}
	end := start
	for end < len(s) && escapeIsHexDigit(s[end]) {
		end++
	}
	if end == start || end >= len(s) || s[end] != ';' {
		return 0, 0, false
	}
	value, err := strconv.ParseInt(s[start:end], base, 32)
	if err != nil {
		return 0, 0, false
	}
	return rune(value & 0xFFFF), end + 1, true
}

// escapeIsHexDigit 是否为十六进制数字字符
func escapeIsHexDigit(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}

// XML 反转义表：五个基础命名实体 + 数字实体
var escapeXmlUnescapeTable *escapeEntityTable

// HTML4 反转义表：251 项命名实体反转 + &apos;（共 252 项）+ 数字实体
var escapeHtml4UnescapeTable *escapeEntityTable

func init() {
	escapeXmlUnescapeTable = newEscapeEntityTable(map[string]rune{
		"&apos;": '\'',
		"&quot;": '"',
		"&amp;":  '&',
		"&lt;":   '<',
		"&gt;":   '>',
	})

	html4Named := make(map[string]rune, len(escapeHtml4Entities)+1)
	for r, name := range escapeHtml4Entities {
		html4Named[name] = r
	}
	// &apos; 仅出现在反转义方向：HTML4 转义不转义单引号，但解码需支持
	html4Named["&apos;"] = '\''
	escapeHtml4UnescapeTable = newEscapeEntityTable(html4Named)
}

// EscapeHtml4 按 HTML4 命名实体表转义特殊字符（对照 hutool escapeHtml4，
// 共 251 项：&quot; &amp; &lt; &gt;、&nbsp;、希腊字母、数学符号、箭头等）。
// 注意 HTML4 转义不处理单引号。表外字符（含控制字符）原样保留。
// 场景：把富文本/非常用符号（如 ©、€、∑）安全输出到网页。
func (e *utilEscape) EscapeHtml4(html string) string {
	var b strings.Builder
	b.Grow(len(html))
	for _, c := range html {
		if entity, ok := escapeHtml4Entities[c]; ok {
			b.WriteString(entity)
		} else {
			b.WriteRune(c)
		}
	}
	return b.String()
}

// UnescapeHtml4 反转义 HTML4 实体：支持全量命名实体（含 &apos;，共 252 项）
// 与数字实体 &#123;（十进制）、&#x1F;（十六进制）。无法识别的实体原样保留。
// 场景：爬虫页面解码、富文本编辑器内容还原。
func (e *utilEscape) UnescapeHtml4(html string) string {
	return escapeHtml4UnescapeTable.decode(html)
}

/* ---------------- Unicode 转义（对照 hutool UnicodeUtil.toUnicode / toString） ---------------- */

// EscapeUnicode 将非 ASCII 字符与不可见字符转义为 \uXXXX 形式（对照 hutool
// UnicodeUtil.toUnicode(str) 默认行为：ASCII 可见字符 0x20~0x7E 原样保留）。
// 增补平面字符（如 emoji）按 UTF-16 代理对拆成两个 \uXXXX，与 Java char 行为一致。
// EscapeUnicode("中文") = "\u4e2d\u6587"。
// 场景：生成 JS/properties 资源文件内容、日志中转义不可见字符。
func (e *utilEscape) EscapeUnicode(str string) string {
	if str == "" {
		return str
	}
	var b strings.Builder
	b.Grow(len(str) * 6)
	for _, c := range str {
		if c >= 0x20 && c <= 0x7E {
			// ASCII 可见字符原样保留（对照 hutool isSkipAscii=true）
			b.WriteRune(c)
			continue
		}
		if c < 0x10000 {
			escapeWriteUnicodeHex(&b, c)
		} else {
			hi, lo := utf16.EncodeRune(c)
			escapeWriteUnicodeHex(&b, hi)
			escapeWriteUnicodeHex(&b, lo)
		}
	}
	return b.String()
}

// UnescapeUnicode 将 \uXXXX 转义还原为原文字符（对照 hutool UnicodeUtil.toString）。
// 标识 \u 不区分大小写（\U4e2d 也可解码）；连续的高/低代理对自动合并回一个字符；
// 非法序列（十六进制位不足或非十六进制字符）原样保留。
// UnescapeUnicode("\u4e2d\u6587") = "中文"。
func (e *utilEscape) UnescapeUnicode(str string) string {
	if str == "" {
		return str
	}
	var b strings.Builder
	b.Grow(len(str))
	pos := 0
	for {
		i := escapeIndexOfUnicodeMark(str, pos)
		if i < 0 {
			break
		}
		b.WriteString(str[pos:i]) // 写入 \u 之前的部分
		pos = i
		if i+6 > len(str) {
			// 剩余不足 4 位十六进制，原样保留
			break
		}
		v, err := strconv.ParseInt(str[i+2:i+6], 16, 32)
		if err != nil {
			// 非法 Unicode 序列，"\u" 原样保留后继续扫描
			b.WriteString(str[i : i+2])
			pos = i + 2
			continue
		}
		r := rune(v)
		pos = i + 6
		// 高代理 + \uXXXX 低代理 => 合并为一个字符（对应 Java 连续两个 char）
		if r >= 0xD800 && r <= 0xDBFF && pos+6 <= len(str) &&
			str[pos] == '\\' && (str[pos+1] == 'u' || str[pos+1] == 'U') {
			if lo, err := strconv.ParseInt(str[pos+2:pos+6], 16, 32); err == nil &&
				lo >= 0xDC00 && lo <= 0xDFFF {
				r = 0x10000 + (r-0xD800)<<10 + rune(lo-0xDC00)
				pos += 6
			}
		}
		b.WriteRune(r) // 孤立代理项无法表示，WriteRune 自动写成 U+FFFD
	}
	if pos < len(str) {
		b.WriteString(str[pos:])
	}
	return b.String()
}

// escapeWriteUnicodeHex 写出一个 \uXXXX（小写十六进制，不足 4 位补 0，
// 对照 hutool HexUtil.toUnicodeHex）："中" -> "\u4e2d"
func escapeWriteUnicodeHex(b *strings.Builder, r rune) {
	const digits = "0123456789abcdef"
	var buf [4]byte
	buf[0] = digits[(r>>12)&0xF]
	buf[1] = digits[(r>>8)&0xF]
	buf[2] = digits[(r>>4)&0xF]
	buf[3] = digits[r&0xF]
	b.WriteString("\\u")
	b.Write(buf[:])
}

// escapeIndexOfUnicodeMark 从 from 开始查找 Unicode 转义标识 "\u"（不区分大小写），
// 返回 '\' 的下标，未找到返回 -1（对照 hutool StrUtil.indexOfIgnoreCase）
func escapeIndexOfUnicodeMark(s string, from int) int {
	for i := from; i+1 < len(s); i++ {
		if s[i] == '\\' && (s[i+1] == 'u' || s[i+1] == 'U') {
			return i
		}
	}
	return -1
}
