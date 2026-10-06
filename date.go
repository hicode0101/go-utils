package utils

import (
	"fmt"
	"time"
)

// utilDate 日期时间工具。
//
// 解析与格式化均使用 Go 参考时间记忆锚点：01/02 03:04:05PM '06 = 1234567，
// 如 "2006-01-02 15:04:05" 对应 yyyy-MM-dd HH:mm:ss（Go 1.20+ 可直接使用
// 标准库常量 time.DateTime / time.DateOnly，本工具即基于这两个常量实现）。
type utilDate struct{}

// timeNanoLayout 带纳秒的时间格式
const timeNanoLayout = "2006-01-02 15:04:05.000000000"

/* ---------------- 当前时间 ---------------- */

// Now 返回当前本地时间。场景：记录创建时间。
func (d *utilDate) Now() time.Time {
	return time.Now()
}

// NowIn 返回指定 IANA 时区名对应的当前时间，如 "UTC"、"Local"、"Asia/Shanghai"、
// "America/New_York"；非法时区名回退为本地时间。
// 提示：优先在操作系统层面设置时区（如 Linux 的 timedatectl），
// 代码级转换一般只用于跨时区展示。
func (d *utilDate) NowIn(locName string) time.Time {
	loc, err := time.LoadLocation(locName)
	if err != nil {
		return time.Now()
	}
	return time.Now().In(loc)
}

// NowShanghai 返回北京时间（UTC+8），用固定偏移实现，不依赖系统时区数据。
// 场景：服务器未配置 tzdata 时也要按东八区出报表。
func (d *utilDate) NowShanghai() time.Time {
	return time.Now().In(time.FixedZone("UTC+8", 8*60*60))
}

// GoBirthday 返回 Go 参考时间 2006-01-02 15:04:05（趣味方法，
// 该值是 Go 官方 layout 的记忆锚点，Go 语言实际发布于 2009 年）。
func (d *utilDate) GoBirthday() time.Time {
	return time.Date(2006, time.January, 2, 15, 4, 5, 0, time.Local)
}

/* ---------------- 解析 ---------------- */

// Parse 按 "2006-01-02 15:04:05" 解析为本地时区时间。
// 场景：解析请求参数中的完整时间字符串。
func (d *utilDate) Parse(timeStr string) (time.Time, error) {
	return time.ParseInLocation(time.DateTime, timeStr, time.Local)
}

// MustParse 同 Parse，解析失败返回零值 time.Time{}。
// 场景：格式由调用方保证时的便捷写法；注意用 IsZero() 判断失败。
func (d *utilDate) MustParse(timeStr string) time.Time {
	t, _ := d.Parse(timeStr)
	return t
}

// ParseDate 按 "2006-01-02" 解析为本地时区时间。场景：解析查询区间的日期。
func (d *utilDate) ParseDate(dateStr string) (time.Time, error) {
	return time.ParseInLocation(time.DateOnly, dateStr, time.Local)
}

// ParseWith 按自定义 layout 解析，如 ParseWith("20240601123045", "20060102150405")。
// 场景：解析第三方系统特有的时间格式。
func (d *utilDate) ParseWith(timeStr, layout string) (time.Time, error) {
	return time.ParseInLocation(layout, timeStr, time.Local)
}

/* ---------------- 时间戳互转 ---------------- */

// FromUnix 秒级时间戳转时间。FromUnix(1717225845) -> 2024-06-01 12:30:45（本地时区）。
func (d *utilDate) FromUnix(sec int64) time.Time {
	return time.Unix(sec, 0)
}

// FromUnixMilli 毫秒级时间戳转时间，兼容 Java 的 System.currentTimeMillis()
// 与 JS 的 Date.now()。场景：解析前端/Java 侧传来的 13 位时间戳。
func (d *utilDate) FromUnixMilli(ms int64) time.Time {
	return time.UnixMilli(ms)
}

// ToUnix 取秒级时间戳。场景：写入以秒为精度的存储字段。
func (d *utilDate) ToUnix(t time.Time) int64 {
	return t.Unix()
}

// ToUnixMilli 取毫秒级时间戳。场景：与前端/Java 系统交互。
func (d *utilDate) ToUnixMilli(t time.Time) int64 {
	return t.UnixMilli()
}

/* ---------------- 格式化 ---------------- */

// Format 格式化为 "2006-01-02 15:04:05"。场景：最常用的展示格式。
func (d *utilDate) Format(t time.Time) string {
	return t.Format(time.DateTime)
}

// FormatNano 格式化为带纳秒的完整时间。场景：高精度耗时记录。
func (d *utilDate) FormatNano(t time.Time) string {
	return t.Format(timeNanoLayout)
}

// FormatCompact 格式化为纯数字 "20060102150405"。
// 场景：订单号后缀、导出文件名，保证按时间可排序。
func (d *utilDate) FormatCompact(t time.Time) string {
	return t.Format("20060102150405")
}

// FormatDate 格式化为 "2006-01-02"。场景：日报、账期展示。
func (d *utilDate) FormatDate(t time.Time) string {
	return t.Format(time.DateOnly)
}

// FormatDateCompact 格式化为 "20060102"。场景：按天切分的文件名、分区目录名。
func (d *utilDate) FormatDateCompact(t time.Time) string {
	return t.Format("20060102")
}

// FormatWith 按自定义 layout 格式化。
// 如 FormatWith(t, "2006年01月02日 15时04分")。
func (d *utilDate) FormatWith(t time.Time, layout string) string {
	return t.Format(layout)
}

// CurrentTime 返回当前时间字符串 "2006-01-02 15:04:05"。
// 场景：日志文本、消息时间字段。
func (d *utilDate) CurrentTime() string {
	return time.Now().Format(time.DateTime)
}

// CurrentDate 返回当前日期字符串 "2006-01-02"。
// 场景：按天命名的报表标题。
func (d *utilDate) CurrentDate() string {
	return time.Now().Format(time.DateOnly)
}

/* ---------------- 比较与差值 ---------------- */

// IsBeforeNow 是否早于当前时间。场景：判断优惠券是否已过期。
func (d *utilDate) IsBeforeNow(t time.Time) bool {
	return t.Before(time.Now())
}

// IsAfterNow 是否晚于当前时间。场景：判断活动是否尚未开始。
func (d *utilDate) IsAfterNow(t time.Time) bool {
	return t.After(time.Now())
}

// IsSameDay 两个时间是否为同一年月日（各自按所在时区取日期）。
// 场景：判断两次操作是否发生在同一天（每日限次场景）。
func (d *utilDate) IsSameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

// absDiff 返回两时间差值的绝对值（内部辅助）
func absDiff(a, b time.Time) time.Duration {
	d := b.Sub(a)
	if d < 0 {
		return -d
	}
	return d
}

// DiffDays 两时间相差的整天数（按绝对时长折算，不足一天舍去）。
// 场景：粗略计算相隔天数；跨时区精确"自然日"计算请自行按日界取整后再算。
func (d *utilDate) DiffDays(a, b time.Time) int64 {
	return int64(absDiff(a, b).Hours() / 24)
}

// DiffHours 相差的整小时数（绝对值，不足一小时舍去）。
// 场景：工时统计。
func (d *utilDate) DiffHours(a, b time.Time) int64 {
	return int64(absDiff(a, b).Hours())
}

// DiffMinutes 相差的整分钟数（绝对值）。
// 场景：订单支付剩余分钟提示。
func (d *utilDate) DiffMinutes(a, b time.Time) int64 {
	return int64(absDiff(a, b).Minutes())
}

// DiffSeconds 相差的整秒数（绝对值）。
// 场景：接口耗时、验证码有效期判断。
func (d *utilDate) DiffSeconds(a, b time.Time) int64 {
	return int64(absDiff(a, b).Seconds())
}

// DiffMillis 相差的毫秒数（绝对值）。
// 场景：性能打点。
func (d *utilDate) DiffMillis(a, b time.Time) int64 {
	return absDiff(a, b).Milliseconds()
}

/* ---------------- 时间加减 ---------------- */

// AddSecs 加 n 秒（n 为负即减）。场景：AddSecs(now, 30) 得到 30 秒后的过期点。
func (d *utilDate) AddSecs(t time.Time, n int64) time.Time {
	return t.Add(time.Duration(n) * time.Second)
}

// AddMins 加 n 分钟。AddMins(now, 10) 加 10 分钟，AddMins(now, -5) 减 5 分钟。
func (d *utilDate) AddMins(t time.Time, n int64) time.Time {
	return t.Add(time.Duration(n) * time.Minute)
}

// AddHours 加 n 小时。场景：AddHours(now, -24) 得到 24 小时前。
func (d *utilDate) AddHours(t time.Time, n int64) time.Time {
	return t.Add(time.Duration(n) * time.Hour)
}

// AddDays 加 n 天。场景：AddDays(now, 7) 得到一周后。
// 注意：AddDays 走日历运算，夏令时切换日可能差一小时；跨时区业务请先归一到 UTC。
func (d *utilDate) AddDays(t time.Time, n int) time.Time {
	return t.AddDate(0, 0, n)
}

// AddMonths 加 n 个月。场景：AddMonths(now, 1) 得到下月同日；
// 目标日不存在时 Go 自动归一（1 月 31 日 + 1 月 = 3 月 2/3 日），边界日期请自行处理。
func (d *utilDate) AddMonths(t time.Time, n int) time.Time {
	return t.AddDate(0, n, 0)
}

// AddYears 加 n 年。场景：证书有效期计算。
func (d *utilDate) AddYears(t time.Time, n int) time.Time {
	return t.AddDate(n, 0, 0)
}

/* ---------------- 日界与月界 ---------------- */

// StartOfDay 返回 t 所在时区的当天起点 00:00:00.000000000。
// 场景：StartOfDay(now) 到 now 的数据即"今日已产生"的数据。
func (d *utilDate) StartOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// EndOfDay 返回 t 所在时区的当天终点 23:59:59.999999999。
// 场景：构造 "今日 23:59:59" 的活动截止时间。
func (d *utilDate) EndOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 23, 59, 59, 999999999, t.Location())
}

// StartOfDayIn 返回指定时区 loc 的当天起点。
// 场景：服务器在 UTC、业务按北京时间统计时：StartOfDayIn(now, shanghaiLoc)。
func (d *utilDate) StartOfDayIn(t time.Time, loc *time.Location) time.Time {
	year, month, day := t.In(loc).Date()
	return time.Date(year, month, day, 0, 0, 0, 0, loc)
}

// EndOfDayIn 返回指定时区 loc 的当天终点。
func (d *utilDate) EndOfDayIn(t time.Time, loc *time.Location) time.Time {
	year, month, day := t.In(loc).Date()
	return time.Date(year, month, day, 23, 59, 59, 999999999, loc)
}

// StartOfMonth 返回当月 1 号 0 点。场景：本月账单起始时间。
func (d *utilDate) StartOfMonth(t time.Time) time.Time {
	year, month, _ := t.Date()
	return time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth 返回当月最后一天 23:59:59.999999999（自动处理大小月与闰年）。
// 场景：月度结算截止时间。
func (d *utilDate) EndOfMonth(t time.Time) time.Time {
	year, month, _ := t.Date()
	// 下个月第 0 天 = 本月最后一天（time.Date 的规范化特性）
	return time.Date(year, month+1, 0, 23, 59, 59, 999999999, t.Location())
}

/* ---------------- 其它 ---------------- */

// TimeCost 打印从 start 到现在的耗时，一行统计函数执行时间：
//
//	func handler() {
//		defer utils.Date.TimeCost(time.Now())
//		...
//	}
func (d *utilDate) TimeCost(start time.Time) {
	fmt.Println("TimeCost：", time.Since(start))
}
