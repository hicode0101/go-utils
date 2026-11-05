package utils

import (
	"regexp"
	"testing"
	"time"
)

/* ---------------- 原有基础断言（自 utils_test.go 原样搬入） ---------------- */

func TestDate(t *testing.T) {
	tm := Date.MustParse("2024-06-01 12:30:45")
	if got := Date.Format(tm); got != "2024-06-01 12:30:45" {
		t.Errorf("Format = %q", got)
	}
	if got := Date.FormatCompact(tm); got != "20240601123045" {
		t.Errorf("FormatCompact = %q", got)
	}
	if got := Date.FormatDate(tm); got != "2024-06-01" {
		t.Errorf("FormatDate = %q", got)
	}
	// 时间戳往返
	if !Date.FromUnixMilli(Date.ToUnixMilli(tm)).Equal(tm) {
		t.Error("毫秒时间戳往返应相等")
	}
	if got := Date.DiffDays(Date.MustParse("2024-06-01 10:00:00"), Date.MustParse("2024-06-02 10:00:00")); got != 1 {
		t.Errorf("DiffDays = %d, want 1", got)
	}
	if !Date.IsSameDay(tm, Date.MustParse("2024-06-01 23:59:59")) {
		t.Error("IsSameDay 同日应为 true")
	}
	if got := Date.FormatDate(Date.EndOfMonth(tm)); got != "2024-06-30" {
		t.Errorf("EndOfMonth = %q, want \"2024-06-30\"", got)
	}
	if got := Date.FormatDate(Date.StartOfMonth(tm)); got != "2024-06-01" {
		t.Errorf("StartOfMonth = %q", got)
	}
	if _, err := Date.Parse("bad"); err == nil {
		t.Error("Parse 非法输入应返回错误")
	}
	if !Date.MustParse("bad").IsZero() {
		t.Error("MustParse 失败应返回零值时间")
	}
}

/* ---------------- 日期扩展（hutool DateUtil，原样搬入） ---------------- */

func TestDateHutoolExt(t *testing.T) {
	if !Date.IsLeapYear(2024) || Date.IsLeapYear(2023) || Date.IsLeapYear(1900) || !Date.IsLeapYear(2000) {
		t.Error("IsLeapYear 判断错误")
	}
	// 2024-06-05 是周三：周一为一周第一天时，本周起点为 06-03
	tm := Date.MustParse("2024-06-05 12:00:00")
	if got := Date.FormatDate(Date.BeginOfWeek(tm, true)); got != "2024-06-03" {
		t.Errorf("BeginOfWeek(周一起) = %q, want \"2024-06-03\"", got)
	}
	if got := Date.FormatDate(Date.BeginOfWeek(tm, false)); got != "2024-06-02" {
		t.Errorf("BeginOfWeek(周日起) = %q, want \"2024-06-02\"", got)
	}
	if got := Date.FormatDate(Date.BeginOfQuarter(tm)); got != "2024-04-01" {
		t.Errorf("BeginOfQuarter = %q, want \"2024-04-01\"", got)
	}
	if got := Date.FormatDate(Date.EndOfQuarter(tm)); got != "2024-06-30" {
		t.Errorf("EndOfQuarter = %q, want \"2024-06-30\"", got)
	}
	birth := Date.MustParse("1990-06-01 00:00:00")
	if got := Date.Age(birth, Date.MustParse("2024-06-01 08:00:00")); got != 34 {
		t.Errorf("Age 生日当天 = %d, want 34", got)
	}
	if got := Date.Age(birth, Date.MustParse("2024-05-31 08:00:00")); got != 33 {
		t.Errorf("Age 生日前 = %d, want 33", got)
	}
}

/* ---------------- 当前时间 ---------------- */

func TestDateNowFamily(t *testing.T) {
	if Date.Now().IsZero() {
		t.Error("Now 不应返回零值")
	}
	// 指定时区
	utc := Date.NowIn("UTC")
	if zone, off := utc.Zone(); zone != "UTC" || off != 0 {
		t.Errorf("NowIn(UTC) 时区 = %s(%d), want UTC(0)", zone, off)
	}
	// 非法时区名回退本地时间
	if got := Date.NowIn("NoSuch/Zone"); got.Location() != time.Local {
		t.Errorf("NowIn 非法时区名应回退本地时区, got %v", got.Location())
	}
	// 固定东八区偏移
	sh := Date.NowShanghai()
	if zone, off := sh.Zone(); zone != "UTC+8" || off != 8*60*60 {
		t.Errorf("NowShanghai 时区 = %s(%d), want UTC+8(%d)", zone, off, 8*60*60)
	}
	// Go 参考时间固定值
	if got := Date.Format(Date.GoBirthday()); got != "2006-01-02 15:04:05" {
		t.Errorf("GoBirthday = %q, want \"2006-01-02 15:04:05\"", got)
	}
}

/* ---------------- 解析 ---------------- */

func TestDateParseFamily(t *testing.T) {
	tm, err := Date.Parse("2024-06-01 12:30:45")
	if err != nil || Date.Format(tm) != "2024-06-01 12:30:45" {
		t.Errorf("Parse = %v, %v", tm, err)
	}
	if _, err := Date.Parse("bad"); err == nil {
		t.Error("Parse 非法输入应返回错误")
	}
	if !Date.MustParse("bad").IsZero() {
		t.Error("MustParse 失败应返回零值时间")
	}
	if good := Date.MustParse("2024-06-01 12:30:45"); good.IsZero() {
		t.Error("MustParse 合法输入不应返回零值")
	}
	d, err := Date.ParseDate("2024-06-01")
	if err != nil || Date.FormatDate(d) != "2024-06-01" {
		t.Errorf("ParseDate = %v, %v", d, err)
	}
	if _, err := Date.ParseDate("2024/06/01"); err == nil {
		t.Error("ParseDate 非法格式应返回错误")
	}
	tw, err := Date.ParseWith("20240601123045", "20060102150405")
	if err != nil || Date.FormatCompact(tw) != "20240601123045" {
		t.Errorf("ParseWith = %v, %v", tw, err)
	}
	if _, err := Date.ParseWith("bad", "20060102150405"); err == nil {
		t.Error("ParseWith 非法输入应返回错误")
	}
}

/* ---------------- 时间戳互转 ---------------- */

func TestDateUnixRoundTrip(t *testing.T) {
	const sec = int64(1717225845)
	if got := Date.ToUnix(Date.FromUnix(sec)); got != sec {
		t.Errorf("秒级往返 = %d, want %d", got, sec)
	}
	if got := Date.ToUnixMilli(Date.FromUnixMilli(sec * 1000)); got != sec*1000 {
		t.Errorf("毫秒级往返 = %d, want %d", got, sec*1000)
	}
	if !Date.FromUnixMilli(sec * 1000).Equal(Date.FromUnix(sec)) {
		t.Error("同一时刻的秒/毫秒转换应相等")
	}
	if got := Date.ToUnixMilli(Date.FromUnix(sec)); got != sec*1000 {
		t.Errorf("ToUnixMilli = %d, want %d", got, sec*1000)
	}
}

/* ---------------- 格式化 ---------------- */

func TestDateFormatSeries(t *testing.T) {
	tm := time.Date(2024, 6, 1, 12, 30, 45, 123456789, time.Local)
	if got := Date.Format(tm); got != "2024-06-01 12:30:45" {
		t.Errorf("Format = %q", got)
	}
	if got := Date.FormatNano(tm); got != "2024-06-01 12:30:45.123456789" {
		t.Errorf("FormatNano = %q", got)
	}
	if got := Date.FormatCompact(tm); got != "20240601123045" {
		t.Errorf("FormatCompact = %q", got)
	}
	if got := Date.FormatDate(tm); got != "2024-06-01" {
		t.Errorf("FormatDate = %q", got)
	}
	if got := Date.FormatDateCompact(tm); got != "20240601" {
		t.Errorf("FormatDateCompact = %q", got)
	}
	if got := Date.FormatWith(tm, "2006年01月02日 15时04分"); got != "2024年06月01日 12时30分" {
		t.Errorf("FormatWith = %q", got)
	}
	timeRe := regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`)
	if !timeRe.MatchString(Date.CurrentTime()) {
		t.Errorf("CurrentTime 格式错误: %q", Date.CurrentTime())
	}
	dateRe := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if !dateRe.MatchString(Date.CurrentDate()) {
		t.Errorf("CurrentDate 格式错误: %q", Date.CurrentDate())
	}
}

/* ---------------- 与当前时间比较 ---------------- */

func TestDateNowCompare(t *testing.T) {
	past := Date.AddHours(Date.Now(), -1)
	future := Date.AddHours(Date.Now(), 1)
	if !Date.IsBeforeNow(past) || Date.IsBeforeNow(future) {
		t.Error("IsBeforeNow 判断错误")
	}
	if !Date.IsAfterNow(future) || Date.IsAfterNow(past) {
		t.Error("IsAfterNow 判断错误")
	}
}

/* ---------------- 同日判断 ---------------- */

func TestDateIsSameDay(t *testing.T) {
	base := Date.MustParse("2024-06-01 08:00:00")
	if !Date.IsSameDay(base, Date.MustParse("2024-06-01 23:59:59")) {
		t.Error("同日不同时刻应为 true")
	}
	if Date.IsSameDay(base, Date.MustParse("2024-06-02 00:00:01")) {
		t.Error("跨日应为 false")
	}
	if Date.IsSameDay(base, Date.MustParse("2025-06-01 08:00:00")) {
		t.Error("跨年同月同日应为 false")
	}
}

/* ---------------- 差值（固定时间对，含反向绝对值） ---------------- */

func TestDateDiff(t *testing.T) {
	base := Date.MustParse("2024-06-01 10:00:00")
	later := base.Add(50*time.Hour + 30*time.Minute + 45*time.Second + 500*time.Millisecond)
	if got := Date.DiffDays(base, later); got != 2 {
		t.Errorf("DiffDays = %d, want 2", got)
	}
	if got := Date.DiffHours(base, later); got != 50 {
		t.Errorf("DiffHours = %d, want 50", got)
	}
	if got := Date.DiffMinutes(base, later); got != 3030 {
		t.Errorf("DiffMinutes = %d, want 3030", got)
	}
	if got := Date.DiffSeconds(base, later); got != 181845 {
		t.Errorf("DiffSeconds = %d, want 181845", got)
	}
	if got := Date.DiffMillis(base, later); got != 181845500 {
		t.Errorf("DiffMillis = %d, want 181845500", got)
	}
	// 参数互换结果不变（绝对值语义）
	if got := Date.DiffDays(later, base); got != 2 {
		t.Errorf("DiffDays 反向 = %d, want 2", got)
	}
	if got := Date.DiffMillis(later, base); got != 181845500 {
		t.Errorf("DiffMillis 反向 = %d, want 181845500", got)
	}
	if got := Date.DiffSeconds(base, base); got != 0 {
		t.Errorf("相同时间 DiffSeconds = %d, want 0", got)
	}
}

/* ---------------- 时间加减（含负数与月末归一） ---------------- */

func TestDateAdd(t *testing.T) {
	base := Date.MustParse("2024-06-01 12:30:45")
	if got := Date.Format(Date.AddSecs(base, 30)); got != "2024-06-01 12:31:15" {
		t.Errorf("AddSecs(30) = %q", got)
	}
	if got := Date.Format(Date.AddSecs(base, -45)); got != "2024-06-01 12:30:00" {
		t.Errorf("AddSecs(-45) = %q", got)
	}
	if got := Date.Format(Date.AddMins(base, 10)); got != "2024-06-01 12:40:45" {
		t.Errorf("AddMins(10) = %q", got)
	}
	if got := Date.Format(Date.AddMins(base, -5)); got != "2024-06-01 12:25:45" {
		t.Errorf("AddMins(-5) = %q", got)
	}
	if got := Date.FormatDate(Date.AddHours(base, 24)); got != "2024-06-02" {
		t.Errorf("AddHours(24) = %q", got)
	}
	if got := Date.FormatDate(Date.AddHours(base, -24)); got != "2024-05-31" {
		t.Errorf("AddHours(-24) = %q", got)
	}
	if got := Date.FormatDate(Date.AddDays(base, 7)); got != "2024-06-08" {
		t.Errorf("AddDays(7) = %q", got)
	}
	if got := Date.FormatDate(Date.AddDays(base, -1)); got != "2024-05-31" {
		t.Errorf("AddDays(-1) = %q", got)
	}
	// 月末归一：Go 的 AddDate 把不存在的目标日顺延（1 月 31 日 + 1 月）。
	// 2024 闰年 2 月 29 天 → 3 月 2 日；2023 平年 2 月 28 天 → 3 月 3 日。
	if got := Date.FormatDate(Date.AddMonths(Date.MustParse("2024-01-31 00:00:00"), 1)); got != "2024-03-02" {
		t.Errorf("AddMonths 闰年月末归一 = %q, want \"2024-03-02\"", got)
	}
	if got := Date.FormatDate(Date.AddMonths(Date.MustParse("2023-01-31 00:00:00"), 1)); got != "2023-03-03" {
		t.Errorf("AddMonths 平年月末归一 = %q, want \"2023-03-03\"", got)
	}
	if got := Date.FormatDate(Date.AddMonths(base, -1)); got != "2024-05-01" {
		t.Errorf("AddMonths(-1) = %q", got)
	}
	if got := Date.FormatDate(Date.AddYears(base, 1)); got != "2025-06-01" {
		t.Errorf("AddYears(1) = %q", got)
	}
	if got := Date.FormatDate(Date.AddYears(base, -2)); got != "2022-06-01" {
		t.Errorf("AddYears(-2) = %q", got)
	}
	// 闰年 2 月 29 日 + 1 年归一到 3 月 1 日（实际值断言）
	if got := Date.FormatDate(Date.AddYears(Date.MustParse("2024-02-29 00:00:00"), 1)); got != "2025-03-01" {
		t.Errorf("AddYears 闰日归一 = %q, want \"2025-03-01\"", got)
	}
}

/* ---------------- 日界（含指定时区） ---------------- */

func TestDateDayBoundary(t *testing.T) {
	tm := Date.MustParse("2024-06-01 12:30:45")
	if got := Date.FormatNano(Date.StartOfDay(tm)); got != "2024-06-01 00:00:00.000000000" {
		t.Errorf("StartOfDay = %q", got)
	}
	if got := Date.FormatNano(Date.EndOfDay(tm)); got != "2024-06-01 23:59:59.999999999" {
		t.Errorf("EndOfDay = %q", got)
	}
	// 固定时区：UTC 20 点在北京时间已是次日 04:00，日界应落在 06-02
	sh := time.FixedZone("UTC+8", 8*60*60)
	utcNight := time.Date(2024, 6, 1, 20, 0, 0, 0, time.UTC)
	if got := Date.StartOfDayIn(utcNight, sh); !got.Equal(time.Date(2024, 6, 2, 0, 0, 0, 0, sh)) {
		t.Errorf("StartOfDayIn = %v, want 2024-06-02 00:00:00 +0800", got)
	}
	if got := Date.EndOfDayIn(utcNight, sh); !got.Equal(time.Date(2024, 6, 2, 23, 59, 59, 999999999, sh)) {
		t.Errorf("EndOfDayIn = %v, want 2024-06-02 23:59:59.999999999 +0800", got)
	}
	if got := Date.FormatWith(Date.StartOfDayIn(utcNight, sh), "2006-01-02 15:04:05 -0700"); got != "2024-06-02 00:00:00 +0800" {
		t.Errorf("StartOfDayIn 格式化 = %q", got)
	}
}

/* ---------------- 月界（含闰年 2 月） ---------------- */

func TestDateMonthBoundary(t *testing.T) {
	leapFeb := Date.MustParse("2024-02-10 08:00:00")
	if got := Date.FormatDate(Date.StartOfMonth(leapFeb)); got != "2024-02-01" {
		t.Errorf("StartOfMonth = %q, want \"2024-02-01\"", got)
	}
	if got := Date.FormatDate(Date.EndOfMonth(leapFeb)); got != "2024-02-29" {
		t.Errorf("EndOfMonth 闰年 2 月 = %q, want \"2024-02-29\"", got)
	}
	if got := Date.FormatNano(Date.EndOfMonth(leapFeb)); got != "2024-02-29 23:59:59.999999999" {
		t.Errorf("EndOfMonth 精确值 = %q", got)
	}
	if got := Date.FormatDate(Date.EndOfMonth(Date.MustParse("2023-02-10 08:00:00"))); got != "2023-02-28" {
		t.Errorf("EndOfMonth 平年 2 月 = %q, want \"2023-02-28\"", got)
	}
	// 12 月边界不跨年
	if got := Date.FormatDate(Date.EndOfMonth(Date.MustParse("2024-12-15 00:00:00"))); got != "2024-12-31" {
		t.Errorf("EndOfMonth 12 月 = %q, want \"2024-12-31\"", got)
	}
	if got := Date.FormatDate(Date.EndOfMonth(Date.MustParse("2024-04-18 00:00:00"))); got != "2024-04-30" {
		t.Errorf("EndOfMonth 小月 = %q, want \"2024-04-30\"", got)
	}
}

/* ---------------- 闰年 ---------------- */

func TestDateIsLeapYear(t *testing.T) {
	if !Date.IsLeapYear(2024) {
		t.Error("2024 应为闰年")
	}
	if !Date.IsLeapYear(2000) {
		t.Error("2000（能被 400 整除）应为闰年")
	}
	if Date.IsLeapYear(1900) {
		t.Error("1900（能被 100 整除但不能被 400 整除）不应为闰年")
	}
	if Date.IsLeapYear(2023) {
		t.Error("2023 不应为闰年")
	}
}

/* ---------------- 周界（周一起与周日起两种） ---------------- */

func TestDateWeek(t *testing.T) {
	// 2024-06-05 为周三
	tm := Date.MustParse("2024-06-05 12:00:00")
	if got := Date.FormatNano(Date.EndOfWeek(tm, true)); got != "2024-06-09 23:59:59.999999999" {
		t.Errorf("EndOfWeek(周一起) = %q", got)
	}
	if got := Date.FormatNano(Date.EndOfWeek(tm, false)); got != "2024-06-08 23:59:59.999999999" {
		t.Errorf("EndOfWeek(周日起) = %q", got)
	}
	// 周日当天：周一起算应回退到周一，周日起算即当天
	sunday := Date.MustParse("2024-06-09 10:00:00")
	if got := Date.FormatDate(Date.BeginOfWeek(sunday, true)); got != "2024-06-03" {
		t.Errorf("BeginOfWeek(周日,周一起) = %q, want \"2024-06-03\"", got)
	}
	if got := Date.FormatDate(Date.BeginOfWeek(sunday, false)); got != "2024-06-09" {
		t.Errorf("BeginOfWeek(周日,周日起) = %q, want \"2024-06-09\"", got)
	}
	if got := Date.FormatDate(Date.BeginOfWeek(tm, true)); got != "2024-06-03" {
		t.Errorf("BeginOfWeek(周一起) = %q, want \"2024-06-03\"", got)
	}
}

/* ---------------- 季度界（Q1~Q4 与跨年） ---------------- */

func TestDateQuarter(t *testing.T) {
	cases := []struct {
		date    string
		wantBeg string
		wantEnd string
	}{
		{"2024-02-15 08:00:00", "2024-01-01", "2024-03-31"}, // Q1（闰年）
		{"2024-06-05 08:00:00", "2024-04-01", "2024-06-30"}, // Q2
		{"2024-08-20 08:00:00", "2024-07-01", "2024-09-30"}, // Q3
		{"2024-11-25 08:00:00", "2024-10-01", "2024-12-31"}, // Q4（年末跨年边界）
	}
	for _, c := range cases {
		tm := Date.MustParse(c.date)
		if got := Date.FormatDate(Date.BeginOfQuarter(tm)); got != c.wantBeg {
			t.Errorf("BeginOfQuarter(%s) = %q, want %q", c.date, got, c.wantBeg)
		}
		if got := Date.FormatDate(Date.EndOfQuarter(tm)); got != c.wantEnd {
			t.Errorf("EndOfQuarter(%s) = %q, want %q", c.date, got, c.wantEnd)
		}
	}
	// Q4 季度终点精确值：次年 1 月 1 日 0 点的前一纳秒
	if got := Date.FormatNano(Date.EndOfQuarter(Date.MustParse("2024-11-25 08:00:00"))); got != "2024-12-31 23:59:59.999999999" {
		t.Errorf("EndOfQuarter(Q4) 精确值 = %q", got)
	}
}

/* ---------------- 年龄 ---------------- */

func TestDateAge(t *testing.T) {
	birth := Date.MustParse("1990-06-01 00:00:00")
	if got := Date.Age(birth, Date.MustParse("2024-06-01 08:00:00")); got != 34 {
		t.Errorf("Age 生日当天 = %d, want 34", got)
	}
	if got := Date.Age(birth, Date.MustParse("2024-06-02 08:00:00")); got != 34 {
		t.Errorf("Age 生日次日 = %d, want 34", got)
	}
	if got := Date.Age(birth, Date.MustParse("2024-05-31 08:00:00")); got != 33 {
		t.Errorf("Age 生日前一天 = %d, want 33", got)
	}
	// 未满 1 周岁按 0 计
	if got := Date.Age(Date.MustParse("2024-01-01 00:00:00"), Date.MustParse("2024-06-01 00:00:00")); got != 0 {
		t.Errorf("Age 未满周岁 = %d, want 0", got)
	}
	// 出生晚于比较时间返回 0
	if got := Date.Age(Date.MustParse("2030-01-01 00:00:00"), Date.MustParse("2024-06-01 00:00:00")); got != 0 {
		t.Errorf("Age 出生晚于比较日 = %d, want 0", got)
	}
	// compare 为零值时按当前时间（1990 年出生至今至少 34 岁）
	if got := Date.Age(birth, time.Time{}); got < 34 {
		t.Errorf("Age 零值比较时间 = %d, 应 >= 34", got)
	}
}

/* ---------------- 耗时打印 ---------------- */

func TestDateTimeCost(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("TimeCost 不应 panic: %v", r)
		}
	}()
	Date.TimeCost(time.Now())
}
