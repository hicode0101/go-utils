package utils

import "unicode"

// utilChar 字符工具（rune 级判断，对照 hutool CharUtil）。
//
// 说明：Go 用 rune（int32）承载字符，负值视为非法输入，
// 各判断方法对负值统一返回 false（IsEmoji 除外，见其注释）。
type utilChar struct{}

// IsAscii 判断是否为 ASCII 字符（0x00~0x7F）。
// IsAscii('a') = true，IsAscii('中') = false。
func (c *utilChar) IsAscii(r rune) bool {
	return r >= 0 && r <= 0x7F
}

// IsAsciiPrintable 判断是否为 ASCII 可见字符（0x20~0x7E，含空格）。
// IsAsciiPrintable('a') = true，IsAsciiPrintable('\n') = false。
// 场景：过滤不可打印字符、判断文本是否可安全展示。
func (c *utilChar) IsAsciiPrintable(r rune) bool {
	return r >= 0x20 && r <= 0x7E
}

// IsAsciiControl 判断是否为 ASCII 控制符（0x00~0x1F 与 0x7F，不可见字符）。
// IsAsciiControl('\n') = true，IsAsciiControl('a') = false。
// 场景：清洗终端转义字符、防控制字符注入。
func (c *utilChar) IsAsciiControl(r rune) bool {
	return (r >= 0 && r < 0x20) || r == 0x7F
}

// IsLetter 判断是否为 Unicode 字母（对照 Character.isLetter，含中文、日文等各语言字母；
// hutool CharUtil.isLetter 仅覆盖 A~Z/a~z，Go 版本按 Unicode 语义放宽，
// 需要严格英文字母时用 IsLetterUpper/IsLetterLower 组合）。
// 场景：标识符宽松校验、文本分类。
func (c *utilChar) IsLetter(r rune) bool {
	return unicode.IsLetter(r)
}

// IsLetterUpper 判断是否为大写英文字母（'A'~'Z'）。场景：常量命名检查。
func (c *utilChar) IsLetterUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

// IsLetterLower 判断是否为小写英文字母（'a'~'z'）。场景：变量命名检查。
func (c *utilChar) IsLetterLower(r rune) bool {
	return r >= 'a' && r <= 'z'
}

// IsNumber 判断是否为数字字符（'0'~'9'，对照 hutool CharUtil.isNumber）。
// 场景：单字符数字判断；整串判断用 String.IsNumber。
func (c *utilChar) IsNumber(r rune) bool {
	return r >= '0' && r <= '9'
}

// IsLetterOrNumber 判断是否为字母或数字（A~Z、a~z、0~9，以及 IsLetter 认可的 Unicode 字母）。
// 场景：昵称/编号合法性校验。
func (c *utilChar) IsLetterOrNumber(r rune) bool {
	return c.IsLetter(r) || c.IsNumber(r)
}

// IsHexChar 判断是否为十六进制字符（0-9、a-f、A-F）。
// 场景：颜色值、哈希串校验的单字符版本。
func (c *utilChar) IsHexChar(r rune) bool {
	return c.IsNumber(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

// IsEmoji 判断是否为 emoji 表情符（对照 hutool isEmoji 的区间表）。
// 判定规则：0x0、0x9、0xA、0xD、[0x20,0xD7FF]、[0xE000,0xFFFD]、[0x100000,0x10FFFF]
// 这些 XML/文本安全区段内的字符返回 false，区段之外的（含负值与代理区）均视为 emoji/非法展示字符。
// 场景：昵称、评论内容的 emoji 与异常字符过滤。
func (c *utilChar) IsEmoji(r rune) bool {
	return !(r == 0x0 ||
		r == 0x9 ||
		r == 0xA ||
		r == 0xD ||
		(r >= 0x20 && r <= 0xD7FF) ||
		(r >= 0xE000 && r <= 0xFFFD) ||
		(r >= 0x100000 && r <= 0x10FFFF))
}

// IsBlankChar 判断是否为空白符（对照 hutool isBlankChar）：空格、制表符、换行、
// 全角空格（U+3000）、不间断空格（U+00A0）等 Unicode 空白（Go 侧用 unicode.IsSpace
// 覆盖 Java isWhitespace + isSpaceChar 的并集），以及 hutool 特判的零宽/特殊字符：
// BOM（U+FEFF）、U+202A、U+0000、韩文填充符（U+3164）、盲文空白（U+2800）、
// 零宽不连字（U+200C）、蒙古文元音分隔符（U+180E）。
// 场景：清洗不可见字符、trim 的单字符判断。
func (c *utilChar) IsBlankChar(r rune) bool {
	return unicode.IsSpace(r) ||
		r == 0xFEFF ||
		r == 0x202A ||
		r == 0 ||
		r == 0x3164 ||
		r == 0x2800 ||
		r == 0x200C ||
		r == 0x180E
}

// IsFileSeparator 判断是否为文件路径分隔符：'/'（Unix/Linux）或 '\\'（Windows）。
// 场景：路径解析中定位分隔符位置。
func (c *utilChar) IsFileSeparator(r rune) bool {
	return r == '/' || r == '\\'
}

// EqualsIgnoreCase 忽略大小写比较两个字符是否相同（对照 hutool equals(c1, c2, caseInsensitive=true)，
// 按 Unicode 简单大小写折叠）。EqualsIgnoreCase('A', 'a') = true。
func (c *utilChar) EqualsIgnoreCase(a, b rune) bool {
	return unicode.ToLower(a) == unicode.ToLower(b)
}
