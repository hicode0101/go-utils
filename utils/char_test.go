package utils

import "testing"

/* ---------------- 迁移自 utils_test.go 的 TestBoolChar（Char 部分，原样保留） ---------------- */

func TestCharLegacy(t *testing.T) {
	if !Char.IsHexChar('a') || Char.IsHexChar('g') {
		t.Error("IsHexChar 判断错误")
	}
	if !Char.IsNumber('5') || Char.IsNumber('x') {
		t.Error("Char.IsNumber 判断错误")
	}
	if !Char.IsBlankChar(' ') || !Char.IsBlankChar(12288) {
		t.Error("IsBlankChar 应含空格与全角空格")
	}
}

/* ---------------- 全部 rune 判断方法（表驱动） ---------------- */

func TestCharClassify(t *testing.T) {
	cases := []struct {
		name string
		got  bool
		want bool
	}{
		// IsAscii：0x00~0x7F
		{"IsAscii a", Char.IsAscii('a'), true},
		{"IsAscii 0x00", Char.IsAscii(0), true},
		{"IsAscii 0x7F 上边界", Char.IsAscii(0x7F), true},
		{"IsAscii 0x80 出界", Char.IsAscii(0x80), false},
		{"IsAscii 中文", Char.IsAscii('中'), false},
		{"IsAscii 负值", Char.IsAscii(-1), false},
		// IsAsciiPrintable：0x20~0x7E
		{"IsAsciiPrintable a", Char.IsAsciiPrintable('a'), true},
		{"IsAsciiPrintable 0x20 下边界", Char.IsAsciiPrintable(0x20), true},
		{"IsAsciiPrintable 0x7E 上边界", Char.IsAsciiPrintable(0x7E), true},
		{"IsAsciiPrintable 换行", Char.IsAsciiPrintable('\n'), false},
		{"IsAsciiPrintable 0x1F", Char.IsAsciiPrintable(0x1F), false},
		{"IsAsciiPrintable 0x7F", Char.IsAsciiPrintable(0x7F), false},
		// IsAsciiControl：0x00~0x1F 与 0x7F
		{"IsAsciiControl 换行", Char.IsAsciiControl('\n'), true},
		{"IsAsciiControl 制表符", Char.IsAsciiControl('\t'), true},
		{"IsAsciiControl 0x00", Char.IsAsciiControl(0), true},
		{"IsAsciiControl 0x1F", Char.IsAsciiControl(0x1F), true},
		{"IsAsciiControl 0x7F", Char.IsAsciiControl(0x7F), true},
		{"IsAsciiControl a", Char.IsAsciiControl('a'), false},
		{"IsAsciiControl 空格", Char.IsAsciiControl(' '), false},
		// IsLetter：Unicode 字母（含中文）
		{"IsLetter 小写", Char.IsLetter('a'), true},
		{"IsLetter 大写", Char.IsLetter('Z'), true},
		{"IsLetter 中文", Char.IsLetter('中'), true},
		{"IsLetter 数字", Char.IsLetter('1'), false},
		{"IsLetter 符号", Char.IsLetter('-'), false},
		// IsLetterUpper：'A'~'Z'
		{"IsLetterUpper A", Char.IsLetterUpper('A'), true},
		{"IsLetterUpper Z", Char.IsLetterUpper('Z'), true},
		{"IsLetterUpper a", Char.IsLetterUpper('a'), false},
		{"IsLetterUpper 中文", Char.IsLetterUpper('中'), false},
		// IsLetterLower：'a'~'z'
		{"IsLetterLower a", Char.IsLetterLower('a'), true},
		{"IsLetterLower z", Char.IsLetterLower('z'), true},
		{"IsLetterLower A", Char.IsLetterLower('A'), false},
		{"IsLetterLower 数字", Char.IsLetterLower('9'), false},
		// IsNumber：'0'~'9'
		{"IsNumber 0", Char.IsNumber('0'), true},
		{"IsNumber 9", Char.IsNumber('9'), true},
		{"IsNumber x", Char.IsNumber('x'), false},
		{"IsNumber 全角数字", Char.IsNumber('１'), false},
		// IsLetterOrNumber：字母或数字
		{"IsLetterOrNumber a", Char.IsLetterOrNumber('a'), true},
		{"IsLetterOrNumber Z", Char.IsLetterOrNumber('Z'), true},
		{"IsLetterOrNumber 5", Char.IsLetterOrNumber('5'), true},
		{"IsLetterOrNumber 中文", Char.IsLetterOrNumber('国'), true},
		{"IsLetterOrNumber 连字符", Char.IsLetterOrNumber('-'), false},
		{"IsLetterOrNumber 空格", Char.IsLetterOrNumber(' '), false},
		// IsHexChar：0-9、a-f、A-F
		{"IsHexChar 0", Char.IsHexChar('0'), true},
		{"IsHexChar 9", Char.IsHexChar('9'), true},
		{"IsHexChar a", Char.IsHexChar('a'), true},
		{"IsHexChar f", Char.IsHexChar('f'), true},
		{"IsHexChar A", Char.IsHexChar('A'), true},
		{"IsHexChar F", Char.IsHexChar('F'), true},
		{"IsHexChar g", Char.IsHexChar('g'), false},
		{"IsHexChar G", Char.IsHexChar('G'), false},
		{"IsHexChar 中文", Char.IsHexChar('中'), false},
		// IsEmoji：XML/文本安全区段外视为 emoji（含负值与代理区）
		{"IsEmoji 😀 U+1F600", Char.IsEmoji('😀'), true},
		{"IsEmoji 代理区 0xD800", Char.IsEmoji(0xD800), true},
		{"IsEmoji 负值", Char.IsEmoji(-1), true},
		{"IsEmoji a", Char.IsEmoji('a'), false},
		{"IsEmoji 空格", Char.IsEmoji(' '), false},
		{"IsEmoji 换行", Char.IsEmoji('\n'), false},
		{"IsEmoji 中文", Char.IsEmoji('中'), false},
		{"IsEmoji 0xE000", Char.IsEmoji(0xE000), false},
		{"IsEmoji 0x10000", Char.IsEmoji(0x10000), true},
		{"IsEmoji 0x100000", Char.IsEmoji(0x100000), false},
		// IsBlankChar：Unicode 空白 + hutool 特判的零宽/特殊字符
		{"IsBlankChar 空格", Char.IsBlankChar(' '), true},
		{"IsBlankChar 制表符", Char.IsBlankChar('\t'), true},
		{"IsBlankChar 换行", Char.IsBlankChar('\n'), true},
		{"IsBlankChar 全角空格", Char.IsBlankChar(0x3000), true},
		{"IsBlankChar 不间断空格", Char.IsBlankChar(0x00A0), true},
		{"IsBlankChar BOM", Char.IsBlankChar(0xFEFF), true},
		{"IsBlankChar 0x202A", Char.IsBlankChar(0x202A), true},
		{"IsBlankChar 0x0000", Char.IsBlankChar(0), true},
		{"IsBlankChar 韩文填充符", Char.IsBlankChar(0x3164), true},
		{"IsBlankChar 盲文空白", Char.IsBlankChar(0x2800), true},
		{"IsBlankChar 零宽不连字", Char.IsBlankChar(0x200C), true},
		{"IsBlankChar 蒙古文元音分隔符", Char.IsBlankChar(0x180E), true},
		{"IsBlankChar a", Char.IsBlankChar('a'), false},
		{"IsBlankChar 中文", Char.IsBlankChar('中'), false},
		// IsFileSeparator：'/' 或 '\\'
		{"IsFileSeparator 正斜杠", Char.IsFileSeparator('/'), true},
		{"IsFileSeparator 反斜杠", Char.IsFileSeparator('\\'), true},
		{"IsFileSeparator 字母", Char.IsFileSeparator('a'), false},
		{"IsFileSeparator 冒号", Char.IsFileSeparator(':'), false},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, c.got, c.want)
		}
	}
}

/* ---------------- EqualsIgnoreCase ---------------- */

func TestCharEqualsIgnoreCase(t *testing.T) {
	if !Char.EqualsIgnoreCase('A', 'a') {
		t.Error("EqualsIgnoreCase('A','a') 应为 true")
	}
	if !Char.EqualsIgnoreCase('a', 'A') {
		t.Error("EqualsIgnoreCase('a','A') 应为 true")
	}
	if !Char.EqualsIgnoreCase('k', 'k') {
		t.Error("EqualsIgnoreCase 相同字符应为 true")
	}
	if !Char.EqualsIgnoreCase('Z', 'z') {
		t.Error("EqualsIgnoreCase('Z','z') 应为 true")
	}
	if Char.EqualsIgnoreCase('x', 'Y') {
		t.Error("EqualsIgnoreCase 不同字符应为 false")
	}
	if !Char.EqualsIgnoreCase('中', '中') {
		t.Error("EqualsIgnoreCase 中文相同字符应为 true")
	}
	if Char.EqualsIgnoreCase('中', '文') {
		t.Error("EqualsIgnoreCase 中文不同字符应为 false")
	}
}
