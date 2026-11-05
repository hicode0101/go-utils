package utils

import "testing"

/* ---------------- 查找与提取 ---------------- */

// TestRegexFind Find：首个整体匹配、无匹配、非法表达式、DOTALL 语义。
func TestRegexFind(t *testing.T) {
	if got := Regex.Find(`\d+年`, "2013年5月"); got != "2013年" {
		t.Errorf("Find = %q, want \"2013年\"", got)
	}
	if got := Regex.Find(`\d+`, "abc"); got != "" {
		t.Errorf("Find 无匹配应返回空串, got %q", got)
	}
	if got := Regex.Find(`(`, "abc"); got != "" {
		t.Errorf("Find 非法表达式应返回空串, got %q", got)
	}
	// 统一 (?s) 编译：. 可匹配换行符（对应 Java Pattern.DOTALL）
	if got := Regex.Find(`a.b`, "a\nb"); got != "a\nb" {
		t.Errorf("Find DOTALL = %q, want \"a\\nb\"", got)
	}
}

// TestRegexFindGroup FindGroup：分组 0、指定分组、无匹配与分组不存在。
func TestRegexFindGroup(t *testing.T) {
	// 原 utils_test.go TestRegex 断言迁入
	if got := Regex.FindGroup(`(\d+)-(\d+)`, "日期 2024-06-01", 1); got != "2024" {
		t.Errorf("FindGroup = %q, want \"2024\"", got)
	}
	if got := Regex.FindGroup(`(\d+)年(\d+)月`, "2013年5月", 0); got != "2013年5月" {
		t.Errorf("FindGroup 分组 0 应为整体匹配, got %q", got)
	}
	if got := Regex.FindGroup(`(\d+)年(\d+)月`, "2013年5月", 2); got != "5" {
		t.Errorf("FindGroup = %q, want \"5\"", got)
	}
	if got := Regex.FindGroup(`(\d+)年`, "abc", 1); got != "" {
		t.Errorf("FindGroup 无匹配应返回空串, got %q", got)
	}
	if got := Regex.FindGroup(`(\d+)年`, "2013年", 5); got != "" {
		t.Errorf("FindGroup 分组不存在应返回空串, got %q", got)
	}
}

// TestRegexGetAllGroups GetAllGroups：withGroup0 两种取值 + 无匹配 nil。
func TestRegexGetAllGroups(t *testing.T) {
	// withGroup0=false：不含整体分组
	groups := Regex.GetAllGroups(`(\d+)年(\d+)月`, "2013年5月", false)
	if len(groups) != 2 || groups[0] != "2013" || groups[1] != "5" {
		t.Errorf("GetAllGroups(false) = %v, want [2013 5]", groups)
	}
	// withGroup0=true：首个元素为整体匹配
	groups = Regex.GetAllGroups(`(\d+)年(\d+)月`, "2013年5月", true)
	if len(groups) != 3 || groups[0] != "2013年5月" || groups[1] != "2013" || groups[2] != "5" {
		t.Errorf("GetAllGroups(true) = %v, want [2013年5月 2013 5]", groups)
	}
	// 无匹配返回 nil
	if got := Regex.GetAllGroups(`\d+年`, "abc", true); got != nil {
		t.Errorf("GetAllGroups 无匹配应返回 nil, got %v", got)
	}
	// 非法表达式返回 nil
	if got := Regex.GetAllGroups(`(`, "abc", true); got != nil {
		t.Errorf("GetAllGroups 非法表达式应返回 nil, got %v", got)
	}
}

// TestRegexFindAllAndIndex FindAll 多匹配 + FindIndex 命中与未命中。
func TestRegexFindAllAndIndex(t *testing.T) {
	// 多匹配（分组 0）
	all := Regex.FindAll(`\d+`, "a1b22c333", 0)
	if len(all) != 3 || all[0] != "1" || all[1] != "22" || all[2] != "333" {
		t.Errorf("FindAll = %v, want [1 22 333]", all)
	}
	// 批量提取指定分组
	all = Regex.FindAll(`(\d+)x(\d+)`, "1x2y3x4", 2)
	if len(all) != 2 || all[0] != "2" || all[1] != "4" {
		t.Errorf("FindAll 分组 2 = %v, want [2 4]", all)
	}
	// 无匹配返回 nil
	if got := Regex.FindAll(`\d+`, "abc", 0); got != nil {
		t.Errorf("FindAll 无匹配应返回 nil, got %v", got)
	}
	// FindIndex 命中：字节区间 [start, end)
	start, end, found := Regex.FindIndex(`\d+`, "ab123cd")
	if !found || start != 2 || end != 5 {
		t.Errorf("FindIndex 命中 = %d, %d, %v; want 2, 5, true", start, end, found)
	}
	// FindIndex 未命中
	start, end, found = Regex.FindIndex(`\d+`, "abcd")
	if found || start != 0 || end != 0 {
		t.Errorf("FindIndex 未命中 = %d, %d, %v; want 0, 0, false", start, end, found)
	}
	// 非法表达式
	if _, _, found := Regex.FindIndex(`(`, "abc"); found {
		t.Error("FindIndex 非法表达式应 found=false")
	}
}

/* ---------------- 删除与替换 ---------------- */

// TestRegexDelete DelFirst/DelLast/DelAll/DelPre 四种删除语义。
func TestRegexDelete(t *testing.T) {
	if got := Regex.DelFirst(`\d+`, "a1b22"); got != "ab22" {
		t.Errorf("DelFirst = %q, want \"ab22\"", got)
	}
	if got := Regex.DelFirst(`\d+`, "abc"); got != "abc" {
		t.Errorf("DelFirst 无匹配应返回原文, got %q", got)
	}
	if got := Regex.DelLast(`\d+`, "a1b22"); got != "a1b" {
		t.Errorf("DelLast = %q, want \"a1b\"", got)
	}
	if got := Regex.DelLast(`\d+`, "abc"); got != "abc" {
		t.Errorf("DelLast 无匹配应返回原文, got %q", got)
	}
	// 原 utils_test.go TestRegex 断言迁入
	if got := Regex.DelAll(`\d`, "a1b2c3"); got != "abc" {
		t.Errorf("DelAll = %q", got)
	}
	if got := Regex.DelAll(`<[^>]+>`, "<b>标题</b>"); got != "标题" {
		t.Errorf("DelAll 标签 = %q, want \"标题\"", got)
	}
	// DelPre：返回首个匹配结束之后的尾串（丢弃头部含匹配本身）
	if got := Regex.DelPre(`\d+`, "编号：123 正文"); got != " 正文" {
		t.Errorf("DelPre = %q, want \" 正文\"", got)
	}
	if got := Regex.DelPre(`\d+`, "ab12cd34ef"); got != "cd34ef" {
		t.Errorf("DelPre 多匹配应只删首个匹配之前, got %q, want \"cd34ef\"", got)
	}
	if got := Regex.DelPre(`\d+`, "abc"); got != "abc" {
		t.Errorf("DelPre 无匹配应返回原文, got %q", got)
	}
	// 非法表达式返回原文
	if got := Regex.DelFirst(`(`, "abc"); got != "abc" {
		t.Errorf("DelFirst 非法表达式应返回原文, got %q", got)
	}
}

// TestRegexExtractAndReplace ExtractMulti（$1/$2/$0 模板）与 ReplaceAll（$N 模板）。
func TestRegexExtractAndReplace(t *testing.T) {
	// 原 utils_test.go TestRegex 断言迁入
	if got := Regex.ExtractMulti(`(\d+)年(\d+)月`, "2024年6月", "$1/$2"); got != "2024/6" {
		t.Errorf("ExtractMulti = %q", got)
	}
	if got := Regex.ExtractMulti(`(\d+)年(\d+)月`, "2013年5月", "$1-$2"); got != "2013-5" {
		t.Errorf("ExtractMulti = %q, want \"2013-5\"", got)
	}
	// $0 为整体匹配
	if got := Regex.ExtractMulti(`(\d+)年`, "2013年", "[$0]"); got != "[2013年]" {
		t.Errorf("ExtractMulti $0 = %q, want \"[2013年]\"", got)
	}
	// 无匹配返回空串
	if got := Regex.ExtractMulti(`(\d+)年`, "abc", "$1"); got != "" {
		t.Errorf("ExtractMulti 无匹配应返回空串, got %q", got)
	}
	// 非法表达式返回空串
	if got := Regex.ExtractMulti(`(`, "abc", "$1"); got != "" {
		t.Errorf("ExtractMulti 非法表达式应返回空串, got %q", got)
	}
	// 分组不存在的占位符原样保留
	if got := Regex.ExtractMulti(`(\d+)年`, "2013年", "$9"); got != "$9" {
		t.Errorf("ExtractMulti 分组不存在应原样保留, got %q", got)
	}
	// ReplaceAll：$N 模板替换
	if got := Regex.ReplaceAll("中文1234", `(\d+)`, "($1)"); got != "中文(1234)" {
		t.Errorf("ReplaceAll = %q, want \"中文(1234)\"", got)
	}
	if got := Regex.ReplaceAll("ab12", `\d+`, "[$0]"); got != "ab[12]" {
		t.Errorf("ReplaceAll $0 = %q, want \"ab[12]\"", got)
	}
	if got := Regex.ReplaceAll("abc", `(`, "x"); got != "abc" {
		t.Errorf("ReplaceAll 非法表达式应原样返回, got %q", got)
	}
}

/* ---------------- 判断与统计 ---------------- */

// TestRegexMatchAndCount Contains/IsMatch（全串语义）/Count。
func TestRegexMatchAndCount(t *testing.T) {
	if !Regex.Contains(`\d+`, "a1b") {
		t.Error("Contains 应为 true")
	}
	if Regex.Contains(`\d+`, "abc") {
		t.Error("Contains 不命中应为 false")
	}
	if Regex.Contains(`(`, "abc") {
		t.Error("Contains 非法表达式应为 false")
	}
	// 原 utils_test.go TestRegex 断言迁入：IsMatch 为全串匹配（Java matches 语义）
	if !Regex.IsMatch(`\d+`, "123") || Regex.IsMatch(`\d+`, "a123b") {
		t.Error("IsMatch 应为全串匹配语义")
	}
	if !Regex.IsMatch(`\d{4}`, "1234") {
		t.Error("IsMatch 全串覆盖应为 true")
	}
	// 多一位即不匹配（区别于 Go 标准库 MatchString 的包含语义）
	if Regex.IsMatch(`\d{4}`, "12345") {
		t.Error("IsMatch 超长串应为 false")
	}
	if !Regex.IsMatch("", "任意内容") {
		t.Error("IsMatch 空正则应直接返回 true")
	}
	// 原 utils_test.go TestRegex 断言迁入
	if got := Regex.Count(`\d`, "a1b22"); got != 3 {
		t.Errorf("Count = %d, want 3", got)
	}
	if got := Regex.Count(`\d+`, "a1b22c333"); got != 3 {
		t.Errorf("Count = %d, want 3", got)
	}
	if got := Regex.Count(`\d+`, "abc"); got != 0 {
		t.Errorf("Count 无匹配应为 0, got %d", got)
	}
	if got := Regex.Count(`(`, "abc"); got != 0 {
		t.Errorf("Count 非法表达式应为 0, got %d", got)
	}
}

/* ---------------- 模板填充与其它 ---------------- */

// TestRegexEscapeAndNumber Escape 转义与 FirstNumber 提取。
func TestRegexEscapeAndNumber(t *testing.T) {
	// 原 utils_test.go TestRegex 断言迁入
	if got := Regex.Escape("a.b*c"); got != `a\.b\*c` {
		t.Errorf("Escape = %q", got)
	}
	if got := Regex.Escape("1+1=2"); got != "1\\+1=2" {
		t.Errorf("Escape = %q, want \"1\\\\+1=2\"", got)
	}
	// 转义后可安全作为字面量参与匹配
	if got := Regex.Find(Regex.Escape("a.b"), "xayb a.b"); got != "a.b" {
		t.Errorf("转义后应按字面量命中, got %q", got)
	}
	// 对照：未转义时按元字符语义命中别处
	if got := Regex.Find("a.b", "xayb a.b"); got != "ayb" {
		t.Errorf("未转义应按元字符命中, got %q, want \"ayb\"", got)
	}
	// FirstNumber：正例（原 utils_test.go TestRegex 断言迁入）
	if n, err := Regex.FirstNumber("第 42 名"); err != nil || n != 42 {
		t.Errorf("FirstNumber = %d, %v", n, err)
	}
	if n, err := Regex.FirstNumber("abc123def456"); err != nil || n != 123 {
		t.Errorf("FirstNumber = %d, %v; want 123", n, err)
	}
	// 负号不在 \d+ 匹配范围，仅提取数字部分
	if n, err := Regex.FirstNumber("欠款-88 元"); err != nil || n != 88 {
		t.Errorf("FirstNumber = %d, %v; want 88", n, err)
	}
	// 错误路径：无数字
	if _, err := Regex.FirstNumber("abc"); err == nil {
		t.Error("FirstNumber 无数字应返回错误")
	}
	// 错误路径：数值超出 int64 范围
	if _, err := Regex.FirstNumber("99999999999999999999999"); err == nil {
		t.Error("FirstNumber 超出 int64 范围应返回错误")
	}
}
