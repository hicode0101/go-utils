package utils

import "testing"

/* ---------------- 原有测试（自 utils_test.go 原样迁入） ---------------- */

func TestVersion(t *testing.T) {
	if Version.Compare("1.2.0", "1.1.9") != 1 {
		t.Error("1.2.0 应大于 1.1.9")
	}
	if Version.Compare("1.0", "1.0.0") != 0 {
		t.Error("尾部 0 应视为相等")
	}
	if Version.Compare("1.0.0-alpha", "1.0.0") != -1 {
		t.Error("预发布版本应小于正式版本")
	}
	if !Version.GreaterThanOrEqual("2.3.3", "2.3.3") {
		t.Error("相等应满足 >=")
	}
	if !Version.MatchEl("1.0.2", ">=1.0,<1.1", ",") {
		t.Error("范围表达式应匹配")
	}
	if !Version.MatchEl("1.0.2", "1.0.0-1.1.1", ";") {
		t.Error("闭区间应匹配")
	}
	if !Version.MatchEl("1.0.2", "<1.0.1;1.0.2", ";") {
		t.Error("任一命中即应返回 true")
	}
	if Version.MatchEl("1.0.2", "<1.0.2", ";") {
		t.Error("小于自身不应匹配")
	}
	if !Version.AnyMatch("1.0.2", "1.0.0", "1.0.2") {
		t.Error("AnyMatch 应命中第二个")
	}
}

/* ---------------- 补充覆盖 ---------------- */

func TestVersionCompareEdge(t *testing.T) {
	// 完全相同走快速路径
	if got := Version.Compare("1.2.3", "1.2.3"); got != 0 {
		t.Errorf("Compare(相同串) = %d, want 0", got)
	}
	// "2.0" 与 "2" 尾部 0 段可省略
	if got := Version.Compare("2.0", "2"); got != 0 {
		t.Errorf("Compare(2.0,2) = %d, want 0", got)
	}
	// 构建段：数字按数值比较
	if got := Version.Compare("1.0.0+build.1", "1.0.0+build.2"); got != -1 {
		t.Errorf("Compare(build.1,build.2) = %d, want -1", got)
	}
	// 带构建段大于不带构建段
	if got := Version.Compare("1.0.0+build.1", "1.0.0"); got <= 0 {
		t.Errorf("Compare(1.0.0+build.1,1.0.0) = %d, 应 > 0", got)
	}
	// 空串视为最小版本
	if got := Version.Compare("", "1.0"); got != -1 {
		t.Errorf("Compare(\"\",1.0) = %d, want -1", got)
	}
	if got := Version.Compare("1.0", ""); got != 1 {
		t.Errorf("Compare(1.0,\"\") = %d, want 1", got)
	}
	if got := Version.Compare("", ""); got != 0 {
		t.Errorf("Compare(\"\",\"\") = %d, want 0", got)
	}
	// 数字段按数值而非字典序：10 > 9
	if got := Version.Compare("1.10.0", "1.9.0"); got != 1 {
		t.Errorf("Compare(1.10.0,1.9.0) = %d, want 1（按数值比较）", got)
	}
	// 预发布段数字同样按数值比较：2 < 10
	if got := Version.Compare("1.0.0-alpha.2", "1.0.0-alpha.10"); got >= 0 {
		t.Errorf("Compare(alpha.2,alpha.10) = %d, 应 < 0（按数值比较）", got)
	}
	// 预发布段按字典序：beta > alpha
	if got := Version.Compare("1.0.0-beta", "1.0.0-alpha"); got != 1 {
		t.Errorf("Compare(beta,alpha) = %d, want 1", got)
	}
}

func TestVersionCompareWrappers(t *testing.T) {
	if !Version.GreaterThan("1.2.0", "1.1.9") {
		t.Error("GreaterThan(1.2.0,1.1.9) 应为 true")
	}
	if Version.GreaterThan("1.0", "1.0") {
		t.Error("GreaterThan 相等应为 false")
	}
	if !Version.GreaterThanOrEqual("2.3.3", "2.3.3") {
		t.Error("GreaterThanOrEqual 相等应为 true")
	}
	if Version.GreaterThanOrEqual("1.0.0-alpha", "1.0.0") {
		t.Error("GreaterThanOrEqual 预发布 < 正式应为 false")
	}
	if !Version.LessThan("1.0.0-alpha", "1.0.0") {
		t.Error("LessThan(1.0.0-alpha,1.0.0) 应为 true")
	}
	if Version.LessThan("2.0", "1.0") {
		t.Error("LessThan(2.0,1.0) 应为 false")
	}
	if !Version.LessThanOrEqual("1.0", "1.0.0") {
		t.Error("LessThanOrEqual 尾部 0 相等应为 true")
	}
	if Version.LessThanOrEqual("1.1", "1.0") {
		t.Error("LessThanOrEqual(1.1,1.0) 应为 false")
	}
}

func TestVersionMatchElOperators(t *testing.T) {
	cases := []struct {
		current, el string
		want        bool
	}{
		{"1.0.2", ">1.0", true},
		{"1.0.0", ">1.0", false},
		{"1.0.2", "<1.1", true},
		{"1.1.0", "<1.1", false},
		{"2.0.0", ">=2.0", true},  // 边界含等号
		{"1.9.9", ">=2.0", false}, // 反例
		{"1.9.9", "<=2.0", true},
		{"2.0.1", "<=2.0", false}, // 反例
		{"1.5", "≥1.0", true},     // 全角大于等于
		{"0.9", "≥1.0", false},    // 反例
		{"1.5", "≤2.0", true},     // 全角小于等于
		{"2.1", "≤2.0", false},    // 反例
	}
	for _, c := range cases {
		if got := Version.MatchEl(c.current, c.el, ";"); got != c.want {
			t.Errorf("MatchEl(%q, %q) = %v, want %v", c.current, c.el, got, c.want)
		}
	}
}

func TestVersionMatchElRange(t *testing.T) {
	// 闭区间
	if !Version.MatchEl("1.0.2", "1.0.0-1.1.1", ";") {
		t.Error("闭区间内应匹配")
	}
	if Version.MatchEl("1.2", "1.0.0-1.1.1", ";") {
		t.Error("闭区间外（大于右端）不应匹配")
	}
	if Version.MatchEl("0.9", "1.0.0-1.1.1", ";") {
		t.Error("闭区间外（小于左端）不应匹配")
	}
	// 开区间 "-" 左空表示 <= 右值
	if !Version.MatchEl("1.2", "-1.5", ";") {
		t.Error("\"-1.5\" 形式应等价 <=1.5")
	}
	if Version.MatchEl("1.6", "-1.5", ";") {
		t.Error("\"-1.5\" 形式下 1.6 不应匹配")
	}
	// 开区间 "-" 右空表示 >= 左值
	if !Version.MatchEl("1.6", "1.5-", ";") {
		t.Error("\"1.5-\" 形式应等价 >=1.5")
	}
	if Version.MatchEl("1.2", "1.5-", ";") {
		t.Error("\"1.5-\" 形式下 1.2 不应匹配")
	}
}

func TestVersionMatchElPreciseAndMulti(t *testing.T) {
	// 精确匹配
	if !Version.MatchEl("1.0.2", "1.0.2", ";") {
		t.Error("精确匹配应命中")
	}
	if Version.MatchEl("1.0.2", "1.0.3", ";") {
		t.Error("版本不同精确匹配不应命中")
	}
	// 多条件任一命中即 true
	if !Version.MatchEl("1.0.2", "<1.0.1;1.0.2", ";") {
		t.Error("任一命中应返回 true")
	}
	if Version.MatchEl("1.0.5", "<1.0.1;1.0.2", ";") {
		t.Error("全部未命中应返回 false")
	}
	// 非法分隔符返回 false
	if Version.MatchEl("1.0", "1.0", "-") {
		t.Error("分隔符 \"-\" 非法应返回 false")
	}
	if Version.MatchEl("1.0", "1.0", "") {
		t.Error("分隔符空串非法应返回 false")
	}
}

func TestVersionAnyMatch(t *testing.T) {
	if !Version.AnyMatch("1.0.2", "<1.0.1", "1.0.2") {
		t.Error("AnyMatch 第二个条件命中应返回 true")
	}
	if Version.AnyMatch("1.0.5", "<1.0.1", "1.0.2") {
		t.Error("AnyMatch 全部未命中应返回 false")
	}
	if !Version.AnyMatch("1.0.2", ">=1.0.2") {
		t.Error("AnyMatch 单条件命中应返回 true")
	}
}
