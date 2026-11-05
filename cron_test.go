package utils

import (
	"testing"
	"time"
)

/* ---------------- 定时任务（hutool-cron）：CronPattern / Scheduler / TaskListener / utilCron ---------------- */

// cronAt 以本地时区构造固定的测试时间，保证表达式匹配断言不依赖运行时刻
func cronAt(year int, month time.Month, day, hour, min, sec int) time.Time {
	return time.Date(year, month, day, hour, min, sec, 0, time.Local)
}

// cronResetFacade 复位包级默认调度器：未启动时 Stop 为空操作，需再显式清空任务表
func cronResetFacade() {
	Cron.Stop()
	Cron.GetScheduler().Clear()
}

// TestCronPattern 原有基础用例（自 utils_test.go 迁入）：6 字段解析、L 月末、步进、W 报错、NextAfter
func TestCronPattern(t *testing.T) {
	p, err := NewCronPattern("0 0 12 * * ?")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	noon := time.Date(2024, 6, 5, 12, 0, 0, 0, time.Local)
	if !p.Match(noon, true) {
		t.Error("12:00:00 应匹配 0 0 12 * * ?")
	}
	if p.Match(noon.Add(time.Second), true) {
		t.Error("12:00:01 不应匹配")
	}
	if p.Match(noon.Add(-time.Hour), true) {
		t.Error("11:00:00 不应匹配")
	}

	// L 月末
	pl, err := NewCronPattern("0 0 2 L * ?")
	if err != nil {
		t.Fatalf("解析 L 失败: %v", err)
	}
	if !pl.Match(time.Date(2024, 6, 30, 2, 0, 0, 0, time.Local), true) {
		t.Error("6 月 30 日（月末）应匹配")
	}
	if pl.Match(time.Date(2024, 6, 29, 2, 0, 0, 0, time.Local), true) {
		t.Error("6 月 29 日不应匹配")
	}
	if !pl.Match(time.Date(2024, 2, 29, 2, 0, 0, 0, time.Local), true) {
		t.Error("闰年 2 月 29 日应匹配")
	}

	// 步进与多表达式
	ps, err := NewCronPattern("0/30 * * * * ?")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if !ps.Match(time.Date(2024, 6, 5, 10, 30, 30, 0, time.Local), true) {
		t.Error("30 秒应匹配 0/30")
	}
	if _, err := NewCronPattern("0 0 12 * * W"); err == nil {
		t.Error("不支持的表达式 W 应返回错误")
	}

	// 下一触发时间
	next := p.NextAfter(time.Date(2024, 6, 5, 12, 0, 1, 0, time.Local), true)
	if want := time.Date(2024, 6, 6, 12, 0, 0, 0, time.Local); !next.Equal(want) {
		t.Errorf("NextAfter = %v, want %v", next, want)
	}
}

// TestCronPatternMatchFixedTimes 固定时间正反例矩阵（覆盖 5/6/7 字段、忽略秒域、反向环绕、
// 日与周 AND、月末、周 7 归一化、月别名、列表复合项、多表达式）
func TestCronPatternMatchFixedTimes(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		at          time.Time
		matchSecond bool
		want        bool
	}{
		{"6字段秒域命中", "0 0 12 * * ?", cronAt(2024, 6, 5, 12, 0, 0), true, true},
		{"6字段秒不匹配", "0 0 12 * * ?", cronAt(2024, 6, 5, 12, 0, 1), true, false},
		{"matchSecond为false秒恒匹配", "0 0 12 * * ?", cronAt(2024, 6, 5, 12, 0, 37), false, true},
		{"5字段分钟不匹配", "30 4 1 1 *", cronAt(2024, 1, 1, 4, 31, 0), true, false},
		{"7字段年份命中", "0 0 8 * * ? 2024", cronAt(2024, 6, 5, 8, 0, 0), true, true},
		{"7字段年份不匹配", "0 0 8 * * ? 2024", cronAt(2025, 6, 5, 8, 0, 0), true, false},
		{"反向环绕周五命中", "0 0 0 * * 5-2", cronAt(2024, 6, 7, 0, 0, 0), true, true},
		{"反向环绕周三不匹配", "0 0 0 * * 5-2", cronAt(2024, 6, 5, 0, 0, 0), true, false},
		{"日周AND命中", "0 0 0 1 * 1", cronAt(2024, 7, 1, 0, 0, 0), true, true},
		{"日周AND仅周一不匹配", "0 0 0 1 * 1", cronAt(2024, 6, 3, 0, 0, 0), true, false},
		{"日周AND仅1号不匹配", "0 0 0 1 * 1", cronAt(2024, 6, 1, 0, 0, 0), true, false},
		{"月末闰年2月29", "0 0 0 L 2 ?", cronAt(2024, 2, 29, 0, 0, 0), true, true},
		{"月末平年2月28", "0 0 0 L 2 ?", cronAt(2023, 2, 28, 0, 0, 0), true, true},
		{"非月末不匹配", "0 0 0 L 2 ?", cronAt(2024, 2, 28, 0, 0, 0), true, false},
		{"周7归一化为周日", "0 0 0 * * 7", cronAt(2024, 6, 2, 0, 0, 0), true, true},
		{"周7非周日不匹配", "0 0 0 * * 7", cronAt(2024, 6, 3, 0, 0, 0), true, false},
		{"月别名DEC命中", "0 0 0 25 DEC ?", cronAt(2024, 12, 25, 0, 0, 0), true, true},
		{"列表复合项命中", "0 1,5,10-12/2 * * * ?", cronAt(2024, 6, 5, 10, 12, 0), true, true},
		{"列表复合项不匹配", "0 1,5,10-12/2 * * * ?", cronAt(2024, 6, 5, 10, 11, 0), true, false},
		{"多表达式任一命中", "0 0 12 * * ?|0 30 18 * * ?", cronAt(2024, 6, 5, 18, 30, 0), true, true},
	}
	for _, tt := range tests {
		p, err := NewCronPattern(tt.pattern)
		if err != nil {
			t.Errorf("%s: 解析 [%s] 失败: %v", tt.name, tt.pattern, err)
			continue
		}
		if got := p.Match(tt.at, tt.matchSecond); got != tt.want {
			t.Errorf("%s: [%s] Match(%v, matchSecond=%v) = %v, want %v",
				tt.name, tt.pattern, tt.at, tt.matchSecond, got, tt.want)
		}
	}
}

// TestCronPatternFieldCounts 5/6/7 字段表达式：5 字段隐含秒 0，6 字段首位为秒，7 字段末位为年
func TestCronPatternFieldCounts(t *testing.T) {
	// 5 字段：分 时 日 月 周，隐含秒 0
	p5, err := NewCronPattern("30 4 1 1 *")
	if err != nil {
		t.Fatalf("解析 5 字段失败: %v", err)
	}
	if !p5.Match(cronAt(2024, 1, 1, 4, 30, 0), true) {
		t.Error("5 字段表达式应匹配 2024-01-01 04:30:00")
	}
	if p5.Match(cronAt(2024, 1, 1, 4, 30, 15), true) {
		t.Error("5 字段表达式隐含秒 0，04:30:15 不应匹配")
	}
	if !p5.Match(cronAt(2024, 1, 1, 4, 30, 15), false) {
		t.Error("matchSecond=false 时秒域应恒匹配")
	}

	// 6 字段：秒 分 时 日 月 周
	p6, err := NewCronPattern("30 0 0 1 1 ?")
	if err != nil {
		t.Fatalf("解析 6 字段失败: %v", err)
	}
	if !p6.Match(cronAt(2024, 1, 1, 0, 0, 30), true) {
		t.Error("6 字段表达式应匹配 00:00:30")
	}
	if p6.Match(cronAt(2024, 1, 1, 0, 0, 31), true) {
		t.Error("00:00:31 不应匹配秒域 30")
	}

	// 7 字段：末位年份，命中与不匹配
	p7, err := NewCronPattern("0 0 8 * * ? 2024")
	if err != nil {
		t.Fatalf("解析 7 字段失败: %v", err)
	}
	if !p7.Match(cronAt(2024, 6, 5, 8, 0, 0), true) {
		t.Error("2024 年应匹配")
	}
	if p7.Match(cronAt(2025, 6, 5, 8, 0, 0), true) {
		t.Error("2025 年不应匹配年份 2024")
	}

	// 7 字段年份范围
	pr, err := NewCronPattern("0 0 0 1 1 ? 2024-2026")
	if err != nil {
		t.Fatalf("解析年份范围失败: %v", err)
	}
	if !pr.Match(cronAt(2024, 1, 1, 0, 0, 0), true) || !pr.Match(cronAt(2026, 1, 1, 0, 0, 0), true) {
		t.Error("2024 与 2026 年应匹配 2024-2026")
	}
	if pr.Match(cronAt(2023, 1, 1, 0, 0, 0), true) || pr.Match(cronAt(2027, 1, 1, 0, 0, 0), true) {
		t.Error("2023 与 2027 年不应匹配 2024-2026")
	}
}

// TestCronPatternWildcards * 与 ? 等价的全匹配语义
func TestCronPatternWildcards(t *testing.T) {
	anyTime := cronAt(2024, 6, 5, 10, 20, 30)
	p1, err := NewCronPattern("* * * * * ?")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if !p1.Match(anyTime, true) {
		t.Error("* * * * * ? 应匹配任意时间")
	}
	p2, err := NewCronPattern("? ? ? ? ? ?")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if !p2.Match(anyTime, true) {
		t.Error("? ? ? ? ? ? 应与 * 等价匹配任意时间")
	}
	p3, err := NewCronPattern("* * * * *")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	// 5 字段表达式隐含秒=0（hutool 语义）：只匹配秒为 0 的时刻，而非任意时间
	if !p3.Match(cronAt(2024, 6, 5, 10, 20, 0), true) {
		t.Error("5 字段全 * 应匹配秒为 0 的任意时刻")
	}
	if p3.Match(anyTime, true) {
		t.Error("5 字段全 * 不应匹配秒为 30 的时刻（隐含秒=0）")
	}
	// 部分域用 ?
	p4, err := NewCronPattern("0 0 12 ? * *")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if !p4.Match(cronAt(2024, 6, 5, 12, 0, 0), true) {
		t.Error("日域 ? 应等价 *")
	}
	if p4.Match(cronAt(2024, 6, 5, 11, 0, 0), true) {
		t.Error("11 点不应匹配时域 12")
	}
}

// TestCronPatternSingleValue 单值域
func TestCronPatternSingleValue(t *testing.T) {
	p, err := NewCronPattern("0 5 * * * ?")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if !p.Match(cronAt(2024, 6, 5, 12, 5, 0), true) {
		t.Error("第 5 分应匹配")
	}
	if p.Match(cronAt(2024, 6, 5, 12, 6, 0), true) {
		t.Error("第 6 分不应匹配")
	}
	ph, err := NewCronPattern("0 0 6 * * ?")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if !ph.Match(cronAt(2024, 6, 5, 6, 0, 0), true) {
		t.Error("6 点应匹配")
	}
	if ph.Match(cronAt(2024, 6, 5, 7, 0, 0), true) {
		t.Error("7 点不应匹配")
	}
}

// TestCronPatternRange 闭区间范围 n-m（两端包含）
func TestCronPatternRange(t *testing.T) {
	p, err := NewCronPattern("0 0 10-12 * * ?")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	for _, h := range []int{10, 11, 12} {
		if !p.Match(cronAt(2024, 6, 5, h, 0, 0), true) {
			t.Errorf("%d 点应匹配 10-12", h)
		}
	}
	for _, h := range []int{9, 13} {
		if p.Match(cronAt(2024, 6, 5, h, 0, 0), true) {
			t.Errorf("%d 点不应匹配 10-12", h)
		}
	}
}

// TestCronPatternReverseRange 反向环绕范围（如周域 5-2 表示周五到周二，分钟域 55-5 跨小时环绕）
func TestCronPatternReverseRange(t *testing.T) {
	p, err := NewCronPattern("0 0 0 * * 5-2")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	// 2024-06-07 周五、06-04 周二、06-08 周六均在 5-2（五、六、日、一、二）内
	for _, d := range []int{4, 7, 8} {
		if !p.Match(cronAt(2024, 6, d, 0, 0, 0), true) {
			t.Errorf("6 月 %d 日应匹配周域 5-2", d)
		}
	}
	// 06-05 周三、06-06 周四不在范围内
	for _, d := range []int{5, 6} {
		if p.Match(cronAt(2024, 6, d, 0, 0, 0), true) {
			t.Errorf("6 月 %d 日不应匹配周域 5-2", d)
		}
	}
	// 分钟域反向环绕 55-5 = {55,56,57,58,59,0,1,2,3,4,5}
	pm, err := NewCronPattern("0 55-5 * * * ?")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if !pm.Match(cronAt(2024, 6, 5, 10, 2, 0), true) || !pm.Match(cronAt(2024, 6, 5, 10, 57, 0), true) {
		t.Error("第 2 分与第 57 分应匹配 55-5")
	}
	if pm.Match(cronAt(2024, 6, 5, 10, 6, 0), true) || pm.Match(cronAt(2024, 6, 5, 10, 54, 0), true) {
		t.Error("第 6 分与第 54 分不应匹配 55-5")
	}
}

// TestCronPatternStep 步进 a/b 与 */b
func TestCronPatternStep(t *testing.T) {
	p, err := NewCronPattern("0 10/5 * * * ?")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	// 10/5 = {10,15,20,...,55}
	if !p.Match(cronAt(2024, 6, 5, 12, 10, 0), true) || !p.Match(cronAt(2024, 6, 5, 12, 55, 0), true) {
		t.Error("第 10 分与第 55 分应匹配 10/5")
	}
	if p.Match(cronAt(2024, 6, 5, 12, 9, 0), true) || p.Match(cronAt(2024, 6, 5, 12, 14, 0), true) {
		t.Error("第 9 分与第 14 分不应匹配 10/5")
	}
	// */20 = {0,20,40}
	pa, err := NewCronPattern("0 */20 * * * ?")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if !pa.Match(cronAt(2024, 6, 5, 12, 0, 0), true) || !pa.Match(cronAt(2024, 6, 5, 12, 20, 0), true) {
		t.Error("第 0 分与第 20 分应匹配 */20")
	}
	if pa.Match(cronAt(2024, 6, 5, 12, 41, 0), true) {
		t.Error("第 41 分不应匹配 */20")
	}
}

// TestCronPatternList 逗号列表且元素可为复合项（范围+步进）
func TestCronPatternList(t *testing.T) {
	p, err := NewCronPattern("0 1,5,10-12/2 * * * ?")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	// 1,5,10-12/2 = {1,5,10,12}
	for _, m := range []int{1, 5, 10, 12} {
		if !p.Match(cronAt(2024, 6, 5, 10, m, 0), true) {
			t.Errorf("第 %d 分应匹配 1,5,10-12/2", m)
		}
	}
	for _, m := range []int{2, 6, 11, 13} {
		if p.Match(cronAt(2024, 6, 5, 10, m, 0), true) {
			t.Errorf("第 %d 分不应匹配 1,5,10-12/2", m)
		}
	}
}

// TestCronPatternMonthAlias 月域英文别名不区分大小写
func TestCronPatternMonthAlias(t *testing.T) {
	p, err := NewCronPattern("0 0 0 1 jan ?")
	if err != nil {
		t.Fatalf("解析 jan 失败: %v", err)
	}
	if !p.Match(cronAt(2024, 1, 1, 0, 0, 0), true) {
		t.Error("1 月 1 日应匹配 jan")
	}
	if p.Match(cronAt(2024, 2, 1, 0, 0, 0), true) {
		t.Error("2 月 1 日不应匹配 jan")
	}
	pd, err := NewCronPattern("0 0 0 25 DEC ?")
	if err != nil {
		t.Fatalf("解析 DEC 失败: %v", err)
	}
	if !pd.Match(cronAt(2024, 12, 25, 0, 0, 0), true) {
		t.Error("12 月 25 日应匹配 DEC")
	}
	if pd.Match(cronAt(2024, 12, 24, 0, 0, 0), true) {
		t.Error("12 月 24 日不应匹配日域 25")
	}
	pm, err := NewCronPattern("0 0 0 25 Dec ?")
	if err != nil {
		t.Fatalf("解析 Dec 失败: %v", err)
	}
	if !pm.Match(cronAt(2024, 12, 25, 0, 0, 0), true) {
		t.Error("大小写混合 Dec 应等价 DEC")
	}
}

// TestCronPatternWeekAlias 周域英文别名不区分大小写
func TestCronPatternWeekAlias(t *testing.T) {
	p, err := NewCronPattern("0 0 0 * * mon")
	if err != nil {
		t.Fatalf("解析 mon 失败: %v", err)
	}
	if !p.Match(cronAt(2024, 6, 3, 0, 0, 0), true) {
		t.Error("2024-06-03 周一应匹配 mon")
	}
	if p.Match(cronAt(2024, 6, 4, 0, 0, 0), true) {
		t.Error("2024-06-04 周二不应匹配 mon")
	}
	// 别名用于范围 sun-sat 覆盖全周
	pa, err := NewCronPattern("0 0 0 * * sun-sat")
	if err != nil {
		t.Fatalf("解析 sun-sat 失败: %v", err)
	}
	if !pa.Match(cronAt(2024, 6, 5, 0, 0, 0), true) {
		t.Error("任意星期应匹配 sun-sat")
	}
	// 大写别名
	ps, err := NewCronPattern("0 0 0 * * SAT")
	if err != nil {
		t.Fatalf("解析 SAT 失败: %v", err)
	}
	if !ps.Match(cronAt(2024, 6, 8, 0, 0, 0), true) {
		t.Error("2024-06-08 周六应匹配 SAT")
	}
	if ps.Match(cronAt(2024, 6, 7, 0, 0, 0), true) {
		t.Error("2024-06-07 周五不应匹配 SAT")
	}
}

// TestCronPatternWeek7AsSunday 周域 7 归一化为 0，与 0 同为周日
func TestCronPatternWeek7AsSunday(t *testing.T) {
	p7, err := NewCronPattern("0 0 0 * * 7")
	if err != nil {
		t.Fatalf("解析周域 7 失败: %v", err)
	}
	if !p7.Match(cronAt(2024, 6, 2, 0, 0, 0), true) {
		t.Error("2024-06-02 周日应匹配周域 7")
	}
	if p7.Match(cronAt(2024, 6, 1, 0, 0, 0), true) {
		t.Error("2024-06-01 周六不应匹配周域 7")
	}
	if p7.Match(cronAt(2024, 6, 3, 0, 0, 0), true) {
		t.Error("2024-06-03 周一不应匹配周域 7")
	}
	p0, err := NewCronPattern("0 0 0 * * 0")
	if err != nil {
		t.Fatalf("解析周域 0 失败: %v", err)
	}
	if !p0.Match(cronAt(2024, 6, 2, 0, 0, 0), true) {
		t.Error("周域 0 与 7 应同为周日")
	}
}

// TestCronPatternLastDayOfMonth 日域 L 月末：大小月、平年 2/28、闰年 2/29、31 天月、30 天月
func TestCronPatternLastDayOfMonth(t *testing.T) {
	feb, err := NewCronPattern("0 0 0 L 2 ?")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if !feb.Match(cronAt(2023, 2, 28, 0, 0, 0), true) {
		t.Error("平年 2023-02-28 应匹配 L")
	}
	if feb.Match(cronAt(2023, 2, 27, 0, 0, 0), true) {
		t.Error("2023-02-27 不应匹配 L")
	}
	if !feb.Match(cronAt(2024, 2, 29, 0, 0, 0), true) {
		t.Error("闰年 2024-02-29 应匹配 L")
	}
	if feb.Match(cronAt(2024, 2, 28, 0, 0, 0), true) {
		t.Error("闰年 2024-02-28 不应匹配 L")
	}
	// 31 天月
	jan, err := NewCronPattern("0 0 0 L 1 ?")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if !jan.Match(cronAt(2024, 1, 31, 0, 0, 0), true) {
		t.Error("2024-01-31 应匹配 L")
	}
	if jan.Match(cronAt(2024, 1, 30, 0, 0, 0), true) {
		t.Error("2024-01-30 不应匹配 L")
	}
	// 30 天月
	apr, err := NewCronPattern("0 0 0 L 4 ?")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if !apr.Match(cronAt(2024, 4, 30, 0, 0, 0), true) {
		t.Error("2024-04-30 应匹配 L")
	}
	if apr.Match(cronAt(2024, 4, 29, 0, 0, 0), true) {
		t.Error("2024-04-29 不应匹配 L")
	}
	// 小写 l 等价 L
	lower, err := NewCronPattern("0 0 0 l 1 ?")
	if err != nil {
		t.Fatalf("解析小写 l 失败: %v", err)
	}
	if !lower.Match(cronAt(2024, 1, 31, 0, 0, 0), true) {
		t.Error("小写 l 应等价 L")
	}
}

// TestCronPatternLastDayOfWeek 周域 L 表示周六（值 6）
func TestCronPatternLastDayOfWeek(t *testing.T) {
	p, err := NewCronPattern("0 0 0 * * L")
	if err != nil {
		t.Fatalf("解析周域 L 失败: %v", err)
	}
	if !p.Match(cronAt(2024, 6, 8, 0, 0, 0), true) {
		t.Error("2024-06-08 周六应匹配周域 L")
	}
	if !p.Match(cronAt(2024, 6, 15, 0, 0, 0), true) {
		t.Error("2024-06-15 周六应匹配周域 L")
	}
	if p.Match(cronAt(2024, 6, 7, 0, 0, 0), true) {
		t.Error("2024-06-07 周五不应匹配周域 L")
	}
}

// TestCronPatternNegativeValue 负数值按源码语义 v += max：
// 月 -1→11、秒 -1→58、日 -1→31（日域 max 为哨兵 32）、周 -6→0（周日）
func TestCronPatternNegativeValue(t *testing.T) {
	pm, err := NewCronPattern("0 0 0 1 -1 ?")
	if err != nil {
		t.Fatalf("解析月域 -1 失败: %v", err)
	}
	if !pm.Match(cronAt(2024, 11, 1, 0, 0, 0), true) {
		t.Error("月域 -1 应加 12 得 11 月")
	}
	if pm.Match(cronAt(2024, 12, 1, 0, 0, 0), true) {
		t.Error("12 月不应匹配月域 -1")
	}
	ps, err := NewCronPattern("-1 0 0 1 1 ?")
	if err != nil {
		t.Fatalf("解析秒域 -1 失败: %v", err)
	}
	if !ps.Match(cronAt(2024, 1, 1, 0, 0, 58), true) {
		t.Error("秒域 -1 应加 59 得第 58 秒")
	}
	if ps.Match(cronAt(2024, 1, 1, 0, 0, 57), true) {
		t.Error("第 57 秒不应匹配秒域 -1")
	}
	pd, err := NewCronPattern("0 0 0 -1 1 ?")
	if err != nil {
		t.Fatalf("解析日域 -1 失败: %v", err)
	}
	if !pd.Match(cronAt(2024, 1, 31, 0, 0, 0), true) {
		t.Error("日域 -1 应加 32 得 31 日")
	}
	if pd.Match(cronAt(2024, 1, 30, 0, 0, 0), true) {
		t.Error("30 日不应匹配日域 -1")
	}
	pw, err := NewCronPattern("0 0 0 * * -6")
	if err != nil {
		t.Fatalf("解析周域 -6 失败: %v", err)
	}
	if !pw.Match(cronAt(2024, 6, 2, 0, 0, 0), true) {
		t.Error("周域 -6 应加 6 得周日")
	}
	if pw.Match(cronAt(2024, 6, 3, 0, 0, 0), true) {
		t.Error("周一不应匹配周域 -6")
	}
}

// TestCronPatternMultiExpr | 分隔多表达式，任一命中即匹配，NextAfter 取最早者
func TestCronPatternMultiExpr(t *testing.T) {
	p, err := NewCronPattern("0 0 12 * * ?|0 30 18 * * ?")
	if err != nil {
		t.Fatalf("解析多表达式失败: %v", err)
	}
	if !p.Match(cronAt(2024, 6, 5, 12, 0, 0), true) {
		t.Error("12:00 应命中第一个表达式")
	}
	if !p.Match(cronAt(2024, 6, 5, 18, 30, 0), true) {
		t.Error("18:30 应命中第二个表达式")
	}
	if p.Match(cronAt(2024, 6, 5, 12, 30, 0), true) {
		t.Error("12:30 不应匹配")
	}
	if p.Match(cronAt(2024, 6, 5, 18, 0, 0), true) {
		t.Error("18:00 不应匹配")
	}
	// | 两侧允许空白（按 trim 处理）
	p2, err := NewCronPattern("0 0 12 * * ? | 0 30 18 * * ?")
	if err != nil {
		t.Fatalf("解析带空白的多表达式失败: %v", err)
	}
	if !p2.Match(cronAt(2024, 6, 5, 18, 30, 0), true) {
		t.Error("带空白的 | 分隔应正常匹配")
	}
	// 多表达式 NextAfter 取最早的下一次
	next := p.NextAfter(cronAt(2024, 6, 5, 13, 0, 0), true)
	if want := cronAt(2024, 6, 5, 18, 30, 0); !next.Equal(want) {
		t.Errorf("多表达式 NextAfter = %v, want %v", next, want)
	}
}

// TestCronPatternDayAndWeekAnd 日域与周域为 AND 语义（非 Quartz OR）：需同时满足
func TestCronPatternDayAndWeekAnd(t *testing.T) {
	p, err := NewCronPattern("0 0 0 1 * 1")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	// 2024-07-01 为周一且是 1 号
	if !p.Match(cronAt(2024, 7, 1, 0, 0, 0), true) {
		t.Error("1 号且周一应匹配")
	}
	// 2024-06-03 周一但非 1 号
	if p.Match(cronAt(2024, 6, 3, 0, 0, 0), true) {
		t.Error("仅周一（非 1 号）不应匹配")
	}
	// 2024-06-01 是 1 号但是周六
	if p.Match(cronAt(2024, 6, 1, 0, 0, 0), true) {
		t.Error("仅 1 号（非周一）不应匹配")
	}
}

// TestCronPatternParseErrors 非法表达式解析报错
func TestCronPatternParseErrors(t *testing.T) {
	badPatterns := []struct {
		pattern string
		reason  string
	}{
		{"", "空表达式"},
		{"|", "无有效子表达式"},
		{"0 0 12 *", "字段数不足"},
		{"0 0 12 * * ? 2024 2025", "字段数超出"},
		{"0 0 12 * * W", "不支持 W 语法"},
		{"0 0 12 ? * 1#2", "不支持 # 语法"},
		{"0 0 12 C * ?", "不支持 C 语法"},
		{"0/0 0 12 * * ?", "步进为 0"},
		{"0/abc 0 12 * * ?", "步进非数字"},
		{"0/1/2 0 12 * * ?", "步进分隔符多次"},
		{"0 0 0 1 13 ?", "月 13 越界"},
		{"60 0 0 1 1 ?", "秒 60 越界"},
		{"0 0 25 * * ?", "时 25 越界"},
		{"0 0 0 * * 8", "周 8 越界"},
		{"0 0 0 1 1 ? 2100", "年 2100 越界"},
		{"0 0 0 1 1 ? 1969", "年 1969 越界"},
		{"abc 0 0 1 1 ?", "无法识别的别名"},
		{"0 0 1x * * ?", "非法字符"},
	}
	for _, tt := range badPatterns {
		if _, err := NewCronPattern(tt.pattern); err == nil {
			t.Errorf("表达式 [%s]（%s）应返回错误", tt.pattern, tt.reason)
		}
	}
}

// TestCronPatternReturnsPattern Pattern 返回原始表达式串
func TestCronPatternReturnsPattern(t *testing.T) {
	raw := "  0 0 12 * * ?  "
	p, err := NewCronPattern(raw)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if p.Pattern() != raw {
		t.Errorf("Pattern() = %q, want 原始串 %q", p.Pattern(), raw)
	}
	plain := "0/5 * * * * ?"
	p2, err := NewCronPattern(plain)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if p2.Pattern() != plain {
		t.Errorf("Pattern() = %q, want %q", p2.Pattern(), plain)
	}
}

// TestCronPatternNextAfter 下一次触发时间：命中 +1 秒、跨日、跨月、周域按天回退、年域耗尽零值
func TestCronPatternNextAfter(t *testing.T) {
	p, err := NewCronPattern("0 0 12 * * ?")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	// 给定时间本身命中：从 +1 秒起算得次日同时刻
	next := p.NextAfter(cronAt(2024, 6, 5, 12, 0, 0), true)
	if want := cronAt(2024, 6, 6, 12, 0, 0); !next.Equal(want) {
		t.Errorf("命中时间的 NextAfter = %v, want %v", next, want)
	}
	// 给定时间不命中：取下一边界
	next = p.NextAfter(cronAt(2024, 6, 5, 12, 0, 1), true)
	if want := cronAt(2024, 6, 6, 12, 0, 0); !next.Equal(want) {
		t.Errorf("未命中时间的 NextAfter = %v, want %v", next, want)
	}
	// 秒级步进的下一边界
	ps, err := NewCronPattern("0/20 * * * * ?")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	next = ps.NextAfter(cronAt(2024, 6, 5, 10, 0, 5), true)
	if want := cronAt(2024, 6, 5, 10, 0, 20); !next.Equal(want) {
		t.Errorf("步进表达式 NextAfter = %v, want %v", next, want)
	}
	// 跨月：6 月 30 日正午后下一次为 7 月 31 日（月末）0 点
	pl, err := NewCronPattern("0 0 0 L * ?")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	next = pl.NextAfter(cronAt(2024, 6, 30, 12, 0, 0), true)
	if want := cronAt(2024, 7, 31, 0, 0, 0); !next.Equal(want) {
		t.Errorf("跨月 NextAfter = %v, want %v", next, want)
	}
	// 周域不参与进位：周五正午后按天回退至下周一 0 点
	pw, err := NewCronPattern("0 0 0 * * 1")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	next = pw.NextAfter(cronAt(2024, 6, 7, 12, 0, 0), true)
	if want := cronAt(2024, 6, 10, 0, 0, 0); !next.Equal(want) {
		t.Errorf("周域回退 NextAfter = %v, want %v", next, want)
	}
	// 年域耗尽：无下一次，返回零值
	py, err := NewCronPattern("0 0 0 1 1 ? 2024")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	next = py.NextAfter(cronAt(2025, 6, 1, 0, 0, 0), true)
	if !next.IsZero() {
		t.Errorf("年域耗尽应返回零值, got %v", next)
	}
}

// TestCronSchedulerBasic 原有调度器冒烟用例（自 utils_test.go 迁入，按 TestCron 前缀更名）
func TestCronSchedulerBasic(t *testing.T) {
	s := NewScheduler(true)
	fired := make(chan string, 2)
	id, err := s.Schedule("0/1 * * * * ?", func() { fired <- "tick" })
	if err != nil {
		t.Fatalf("Schedule 失败: %v", err)
	}
	if s.Size() != 1 || s.IsStarted() {
		t.Error("启动前状态错误")
	}
	if err := s.Start(); err != nil {
		t.Fatalf("Start 失败: %v", err)
	}
	select {
	case <-fired:
		// 触发成功
	case <-time.After(4 * time.Second):
		t.Error("4 秒内未触发任务")
	}
	if !s.Deschedule(id) {
		t.Error("Deschedule 应返回 true")
	}
	s.Stop(true)
	if s.IsStarted() {
		t.Error("Stop 后不应处于启动状态")
	}
	if _, err := s.Schedule("bad pattern", func() {}); err == nil {
		t.Error("非法表达式应返回错误")
	}
}

// TestCronSchedulerNew 创建调度器与 SetMatchSecond 链式调用
func TestCronSchedulerNew(t *testing.T) {
	s := NewScheduler(true)
	if s == nil {
		t.Fatal("NewScheduler 返回 nil")
	}
	if s.Size() != 0 || s.IsStarted() {
		t.Error("新建调度器应为空且未启动")
	}
	if got := s.SetMatchSecond(false); got != s {
		t.Error("SetMatchSecond 应返回调度器自身以支持链式调用")
	}
	if got := s.SetMatchSecond(true); got != s {
		t.Error("SetMatchSecond 应返回调度器自身以支持链式调用")
	}
	s.Stop(false)
}

// TestCronSchedulerSchedule 自动生成任务 ID 与任务计数
func TestCronSchedulerSchedule(t *testing.T) {
	s := NewScheduler(true)
	defer s.Stop(true)
	id1, err := s.Schedule("0 0 0 1 1 ?", func() {})
	if err != nil {
		t.Fatalf("Schedule 失败: %v", err)
	}
	if id1 == "" {
		t.Error("自动生成的任务 ID 不应为空")
	}
	if s.Size() != 1 {
		t.Errorf("任务数 = %d, want 1", s.Size())
	}
	id2, err := s.Schedule("0 0 0 1 1 ?", func() {})
	if err != nil {
		t.Fatalf("Schedule 失败: %v", err)
	}
	if id2 == id1 {
		t.Error("两次自动生成的任务 ID 不应相同")
	}
	if s.Size() != 2 {
		t.Errorf("任务数 = %d, want 2", s.Size())
	}
	// 非法表达式返回 error 且不计数
	if _, err := s.Schedule("bad pattern", func() {}); err == nil {
		t.Error("非法表达式应返回错误")
	}
	if s.Size() != 2 {
		t.Errorf("非法表达式不应入表, 任务数 = %d", s.Size())
	}
}

// TestCronSchedulerScheduleWithId 指定 ID 注册：重复 ID 与非法表达式报错
func TestCronSchedulerScheduleWithId(t *testing.T) {
	s := NewScheduler(true)
	defer s.Stop(true)
	if err := s.ScheduleWithId("task-a", "0 0 0 1 1 ?", func() {}); err != nil {
		t.Fatalf("ScheduleWithId 失败: %v", err)
	}
	if s.Size() != 1 {
		t.Errorf("任务数 = %d, want 1", s.Size())
	}
	if err := s.ScheduleWithId("task-a", "0 0 0 1 1 ?", func() {}); err == nil {
		t.Error("重复 ID 应返回错误")
	}
	if err := s.ScheduleWithId("task-b", "bad pattern", func() {}); err == nil {
		t.Error("非法表达式应返回错误")
	}
	if s.Size() != 1 {
		t.Errorf("失败注册不应入表, 任务数 = %d", s.Size())
	}
}

// TestCronSchedulerDeschedule 移除任务：存在返回 true，不存在返回 false
func TestCronSchedulerDeschedule(t *testing.T) {
	s := NewScheduler(true)
	defer s.Stop(true)
	if s.Deschedule("never-exists") {
		t.Error("移除不存在的任务应返回 false")
	}
	if err := s.ScheduleWithId("task-a", "0 0 0 1 1 ?", func() {}); err != nil {
		t.Fatalf("ScheduleWithId 失败: %v", err)
	}
	if !s.Deschedule("task-a") {
		t.Error("移除存在的任务应返回 true")
	}
	if s.Size() != 0 {
		t.Errorf("移除后任务数 = %d, want 0", s.Size())
	}
	if s.Deschedule("task-a") {
		t.Error("重复移除应返回 false")
	}
}

// TestCronSchedulerUpdatePattern 更新表达式：成功、非法表达式、ID 不存在
func TestCronSchedulerUpdatePattern(t *testing.T) {
	s := NewScheduler(true)
	defer s.Stop(true)
	if err := s.ScheduleWithId("task-u", "0 0 0 1 1 ?", func() {}); err != nil {
		t.Fatalf("ScheduleWithId 失败: %v", err)
	}
	// 成功更新
	if err := s.UpdatePattern("task-u", "0 0 12 * * ?"); err != nil {
		t.Fatalf("UpdatePattern 失败: %v", err)
	}
	if p, ok := s.GetPattern("task-u"); !ok || p == nil || p.Pattern() != "0 0 12 * * ?" {
		t.Errorf("更新后的表达式 = %v, want 0 0 12 * * ?", p)
	}
	// 非法表达式报错且保留原表达式
	if err := s.UpdatePattern("task-u", "bad pattern"); err == nil {
		t.Error("非法表达式应返回错误")
	}
	if p, _ := s.GetPattern("task-u"); p == nil || p.Pattern() != "0 0 12 * * ?" {
		t.Errorf("失败更新不应改动原表达式, got %v", p)
	}
	// ID 不存在报错
	if err := s.UpdatePattern("no-such-id", "0 0 0 1 1 ?"); err == nil {
		t.Error("更新不存在的任务应返回错误")
	}
}

// TestCronSchedulerGetPattern 查询任务表达式：命中与未命中
func TestCronSchedulerGetPattern(t *testing.T) {
	s := NewScheduler(true)
	defer s.Stop(true)
	if err := s.ScheduleWithId("task-g", "0 0 0 1 1 ?", func() {}); err != nil {
		t.Fatalf("ScheduleWithId 失败: %v", err)
	}
	p, ok := s.GetPattern("task-g")
	if !ok || p == nil {
		t.Fatal("已注册任务应能查到表达式")
	}
	if p.Pattern() != "0 0 0 1 1 ?" {
		t.Errorf("GetPattern = %q, want 0 0 0 1 1 ?", p.Pattern())
	}
	if p2, ok := s.GetPattern("no-such-id"); ok || p2 != nil {
		t.Error("未注册任务应返回 nil 与 false")
	}
}

// TestCronSchedulerClear 清空任务表
func TestCronSchedulerClear(t *testing.T) {
	s := NewScheduler(true)
	defer s.Stop(true)
	if _, err := s.Schedule("0 0 0 1 1 ?", func() {}); err != nil {
		t.Fatalf("Schedule 失败: %v", err)
	}
	if _, err := s.Schedule("0 0 0 1 1 ?", func() {}); err != nil {
		t.Fatalf("Schedule 失败: %v", err)
	}
	s.Clear()
	if s.Size() != 0 {
		t.Errorf("Clear 后任务数 = %d, want 0", s.Size())
	}
	// 重复 Clear 不 panic
	s.Clear()
	if s.Size() != 0 {
		t.Errorf("重复 Clear 后任务数 = %d, want 0", s.Size())
	}
}

// TestCronSchedulerLifecycle 启停生命周期：IsStarted、重复 Start 幂等、未启动 Stop 不 panic、
// Stop(false) 保留任务表、Stop(true) 清空任务表、停止后可重启
func TestCronSchedulerLifecycle(t *testing.T) {
	s := NewScheduler(true)
	// 未启动时 Stop 为空操作，不应 panic
	s.Stop(true)
	s.Stop(false)
	if s.IsStarted() {
		t.Error("新建调度器不应处于启动状态")
	}
	if err := s.ScheduleWithId("l1", "0 0 0 1 1 ?", func() {}); err != nil {
		t.Fatalf("ScheduleWithId 失败: %v", err)
	}
	if err := s.ScheduleWithId("l2", "0 0 0 1 1 ?", func() {}); err != nil {
		t.Fatalf("ScheduleWithId 失败: %v", err)
	}
	if err := s.Start(); err != nil {
		t.Fatalf("Start 失败: %v", err)
	}
	if !s.IsStarted() {
		t.Error("Start 后应处于启动状态")
	}
	// 重复 Start 幂等
	if err := s.Start(); err != nil {
		t.Errorf("重复 Start 应为空操作并返回 nil: %v", err)
	}
	if !s.IsStarted() {
		t.Error("重复 Start 后仍应处于启动状态")
	}
	// Stop(false)：停止但保留任务表
	s.Stop(false)
	if s.IsStarted() {
		t.Error("Stop 后不应处于启动状态")
	}
	if s.Size() != 2 {
		t.Errorf("Stop(false) 应保留任务表, 任务数 = %d", s.Size())
	}
	// 停止后可重新启动
	if err := s.Start(); err != nil {
		t.Fatalf("停止后重新 Start 失败: %v", err)
	}
	// Stop(true)：清空任务表
	s.Stop(true)
	if s.IsStarted() {
		t.Error("Stop(true) 后不应处于启动状态")
	}
	if s.Size() != 0 {
		t.Errorf("Stop(true) 应清空任务表, 任务数 = %d", s.Size())
	}
	// 已停止后重复 Stop 不 panic
	s.Stop(true)
}

// TestCronSchedulerFire Start 后秒级表达式 2 秒内触发
func TestCronSchedulerFire(t *testing.T) {
	s := NewScheduler(true)
	defer s.Stop(false)
	fired := make(chan struct{}, 1)
	if _, err := s.Schedule("0/1 * * * * ?", func() { fired <- struct{}{} }); err != nil {
		t.Fatalf("Schedule 失败: %v", err)
	}
	if err := s.Start(); err != nil {
		t.Fatalf("Start 失败: %v", err)
	}
	select {
	case <-fired:
		// 秒级任务已触发
	case <-time.After(2 * time.Second):
		t.Error("Start 后 2 秒内未触发秒级任务")
	}
}

// TestCronSchedulerPanicRecover 任务 panic 被 recover 并回调 OnFailed，不影响调度器与其它任务
func TestCronSchedulerPanicRecover(t *testing.T) {
	s := NewScheduler(true)
	defer s.Stop(false)
	started := make(chan string, 16)
	succeeded := make(chan string, 16)
	failedID := make(chan string, 16)
	failedVal := make(chan interface{}, 16)
	s.AddListener(TaskListener{
		OnStart:     func(id string) { started <- id },
		OnSucceeded: func(id string) { succeeded <- id },
		OnFailed: func(id string, panicked interface{}) {
			failedID <- id
			failedVal <- panicked
		},
	})
	if err := s.ScheduleWithId("panic-task", "0/1 * * * * ?", func() { panic("boom") }); err != nil {
		t.Fatalf("ScheduleWithId 失败: %v", err)
	}
	healthy := make(chan struct{}, 1)
	if _, err := s.Schedule("0/1 * * * * ?", func() { healthy <- struct{}{} }); err != nil {
		t.Fatalf("Schedule 失败: %v", err)
	}
	if err := s.Start(); err != nil {
		t.Fatalf("Start 失败: %v", err)
	}
	// panic 任务的 OnStart 先到
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("2 秒内未收到 OnStart 回调")
	}
	// 任务 panic 被 recover 并回调 OnFailed
	select {
	case id := <-failedID:
		if id != "panic-task" {
			t.Errorf("OnFailed 任务 ID = %s, want panic-task", id)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("2 秒内未收到 OnFailed 回调")
	}
	if val := <-failedVal; val != "boom" {
		t.Errorf("OnFailed 捕获的 panic 值 = %v, want boom", val)
	}
	// panic 任务不应回调 OnSucceeded
	select {
	case id := <-succeeded:
		if id == "panic-task" {
			t.Error("panic 任务不应收到 OnSucceeded 回调")
		}
	case <-time.After(1200 * time.Millisecond):
		// 窗口内无任何 OnSucceeded 即符合预期
	}
	// panic 不拖垮调度器，其它任务继续触发
	if !s.IsStarted() {
		t.Error("任务 panic 后调度器不应停止")
	}
	select {
	case <-healthy:
	case <-time.After(2 * time.Second):
		t.Error("任务 panic 后其它任务应正常触发")
	}
}

// TestCronSchedulerListenerOrder OnStart → OnSucceeded 回调顺序，RemoveListener 后不再回调
func TestCronSchedulerListenerOrder(t *testing.T) {
	s := NewScheduler(true)
	defer s.Stop(false)
	events := make(chan string, 8)
	handle := s.AddListener(TaskListener{
		OnStart:     func(id string) { events <- "start" },
		OnSucceeded: func(id string) { events <- "succeeded" },
	})
	if handle == 0 {
		t.Error("AddListener 应返回非零句柄")
	}
	if err := s.ScheduleWithId("order-task", "0/1 * * * * ?", func() {}); err != nil {
		t.Fatalf("ScheduleWithId 失败: %v", err)
	}
	if err := s.Start(); err != nil {
		t.Fatalf("Start 失败: %v", err)
	}
	// 回调顺序：OnStart → 任务体 → OnSucceeded
	var got []string
	deadline := time.After(2 * time.Second)
	for len(got) < 2 {
		select {
		case ev := <-events:
			got = append(got, ev)
		case <-deadline:
			t.Fatalf("2 秒内回调不足: %v", got)
		}
	}
	if got[0] != "start" || got[1] != "succeeded" {
		t.Errorf("回调顺序 = %v, want [start succeeded]", got)
	}
	// 移除监听器
	if !s.RemoveListener(handle) {
		t.Error("RemoveListener 应返回 true")
	}
	if s.RemoveListener(handle) {
		t.Error("重复 RemoveListener 应返回 false")
	}
	if s.RemoveListener(handle + 1000) {
		t.Error("移除不存在的句柄应返回 false")
	}
	// 跨越至少一个秒级边界，确认不再有任何回调
	select {
	case ev := <-events:
		t.Errorf("移除监听器后仍收到回调: %s", ev)
	case <-time.After(1300 * time.Millisecond):
		// 无回调，符合预期
	}
}

// TestCronSchedulerDescheduleStopsFire 触发一次后 Deschedule，确认后续不再触发
func TestCronSchedulerDescheduleStopsFire(t *testing.T) {
	s := NewScheduler(true)
	defer s.Stop(false)
	fired := make(chan struct{}, 2)
	id, err := s.Schedule("0/1 * * * * ?", func() { fired <- struct{}{} })
	if err != nil {
		t.Fatalf("Schedule 失败: %v", err)
	}
	if err := s.Start(); err != nil {
		t.Fatalf("Start 失败: %v", err)
	}
	// 等待首次触发
	select {
	case <-fired:
	case <-time.After(2 * time.Second):
		t.Fatal("2 秒内未触发任务")
	}
	// 立即移除任务
	if !s.Deschedule(id) {
		t.Fatal("Deschedule 应返回 true")
	}
	// 跨越至少一个秒级边界，确认第二次不触发
	select {
	case <-fired:
		t.Error("Deschedule 后任务不应再触发")
	case <-time.After(1500 * time.Millisecond):
		// 无触发，符合预期
	}
}

// TestCronFacadeTable utilCron 门面：SetMatchSecond、Schedule、ScheduleWithId、Remove、
// UpdatePattern、Size、GetScheduler
func TestCronFacadeTable(t *testing.T) {
	cronResetFacade()
	Cron.SetMatchSecond(true)
	if Cron.GetScheduler() == nil {
		t.Fatal("GetScheduler 应返回非 nil 的默认调度器")
	}
	if Cron.IsStarted() {
		t.Error("门面默认不应处于启动状态")
	}
	if Cron.Size() != 0 {
		t.Errorf("复位后任务数 = %d, want 0", Cron.Size())
	}
	// Schedule 自动 ID
	id, err := Cron.Schedule("0 0 0 1 1 ?", func() {})
	if err != nil {
		t.Fatalf("Cron.Schedule 失败: %v", err)
	}
	if id == "" {
		t.Error("自动生成的任务 ID 不应为空")
	}
	if Cron.Size() != 1 {
		t.Errorf("任务数 = %d, want 1", Cron.Size())
	}
	// ScheduleWithId 与重复 ID
	if err := Cron.ScheduleWithId("facade-fixed", "0 0 0 2 1 ?", func() {}); err != nil {
		t.Fatalf("Cron.ScheduleWithId 失败: %v", err)
	}
	if err := Cron.ScheduleWithId("facade-fixed", "0 0 0 2 1 ?", func() {}); err == nil {
		t.Error("重复 ID 应返回错误")
	}
	if Cron.Size() != 2 {
		t.Errorf("任务数 = %d, want 2", Cron.Size())
	}
	// UpdatePattern 成功、非法表达式、ID 不存在
	if err := Cron.UpdatePattern("facade-fixed", "0 0 0 3 1 ?"); err != nil {
		t.Fatalf("Cron.UpdatePattern 失败: %v", err)
	}
	if p, ok := Cron.GetScheduler().GetPattern("facade-fixed"); !ok || p == nil || p.Pattern() != "0 0 0 3 1 ?" {
		t.Errorf("更新后的表达式 = %v, want 0 0 0 3 1 ?", p)
	}
	if err := Cron.UpdatePattern("facade-fixed", "bad pattern"); err == nil {
		t.Error("非法表达式应返回错误")
	}
	if err := Cron.UpdatePattern("no-such-id", "0 0 0 1 1 ?"); err == nil {
		t.Error("更新不存在的任务应返回错误")
	}
	// Remove 命中与未命中
	if !Cron.Remove(id) {
		t.Error("Remove 存在的任务应返回 true")
	}
	if Cron.Remove(id) {
		t.Error("重复 Remove 应返回 false")
	}
	if Cron.Size() != 1 {
		t.Errorf("移除后任务数 = %d, want 1", Cron.Size())
	}
	// 复位，避免影响后续用例
	cronResetFacade()
}

// TestCronFacadeLifecycle utilCron 门面 Start/Stop 全流程（触发 1 次即停）与 Restart
func TestCronFacadeLifecycle(t *testing.T) {
	cronResetFacade()
	Cron.SetMatchSecond(true)
	fired := make(chan struct{}, 1)
	if err := Cron.ScheduleWithId("facade-fire", "0/1 * * * * ?", func() { fired <- struct{}{} }); err != nil {
		t.Fatalf("Cron.ScheduleWithId 失败: %v", err)
	}
	// Start 后 2 秒内触发
	if err := Cron.Start(); err != nil {
		t.Fatalf("Cron.Start 失败: %v", err)
	}
	if !Cron.IsStarted() {
		t.Error("Start 后应处于启动状态")
	}
	select {
	case <-fired:
		// 触发成功
	case <-time.After(2 * time.Second):
		t.Error("Start 后 2 秒内未触发任务")
	}
	// Stop 清空任务表并停止
	Cron.Stop()
	if Cron.IsStarted() {
		t.Error("Stop 后不应处于启动状态")
	}
	if Cron.Size() != 0 {
		t.Errorf("Stop 后任务数 = %d, want 0", Cron.Size())
	}
	// Restart：重新注册任务后重启并再次触发
	if err := Cron.ScheduleWithId("facade-restart", "0/1 * * * * ?", func() { fired <- struct{}{} }); err != nil {
		t.Fatalf("Cron.ScheduleWithId 失败: %v", err)
	}
	if err := Cron.Restart(); err != nil {
		t.Fatalf("Cron.Restart 失败: %v", err)
	}
	if !Cron.IsStarted() {
		t.Error("Restart 后应处于启动状态")
	}
	select {
	case <-fired:
		// 触发成功
	case <-time.After(2 * time.Second):
		t.Error("Restart 后 2 秒内未触发任务")
	}
	// 收尾与重复 Stop 不 panic
	Cron.Stop()
	Cron.Stop()
	if Cron.IsStarted() {
		t.Error("收尾 Stop 后不应处于启动状态")
	}
}
