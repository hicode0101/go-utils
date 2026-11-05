package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

/* ---------------- 表达式域定义（对照 hutool Part 枚举） ---------------- */

// cron 表达式各域在字段数组中的下标（对照 hutool Part 枚举的 ordinal）。
// 字段数组恒为 {second, minute, hour, dayOfMonth, month, dayOfWeek, year}。
const (
	cronPartSecond     = 0
	cronPartMinute     = 1
	cronPartHour       = 2
	cronPartDayOfMonth = 3
	cronPartMonth      = 4
	cronPartDayOfWeek  = 5
	cronPartYear       = 6
)

// cronLastDay 日域哨兵值：用户书写 "L"（月末）时内部以 32 表示，
// 匹配时特判为当月最后一天（感知大小月与闰年）。
const cronLastDay = 32

// cronPartDef 表达式单个域的定义。
type cronPartDef struct {
	name string // 域中文名，用于错误提示
	min  int    // 允许的最小值（含）
	max  int    // 允许的最大值（含）
}

// cronPartDefs 各域的名称与取值范围，下标即域在表达式中的序号（对照 hutool Part）。
var cronPartDefs = [7]cronPartDef{
	{"秒", 0, 59},
	{"分", 0, 59},
	{"时", 0, 23},
	{"日", 1, cronLastDay}, // 32 为 "L"（月末）哨兵值
	{"月", 1, 12},
	{"周", 0, 6}, // 0 表示周日
	{"年", 1970, 2099},
}

/* ---------------- 域匹配器（对照 hutool pattern.matcher 包） ---------------- */

// cronPartMatcher 单个域的匹配器接口（对照 hutool PartMatcher）。
type cronPartMatcher interface {
	// match 判断给定值是否匹配此域表达式
	match(value int) bool
	// nextAfter 返回 value 及之后最近的匹配值；value 已越过全部匹配值时回绕到最小匹配值
	nextAfter(value int) int
}

// cronAlwaysTrueMatcher 全匹配器，对应 "*" 或 "?"，恒匹配（对照 hutool AlwaysTrueMatcher）。
type cronAlwaysTrueMatcher struct{}

func (m *cronAlwaysTrueMatcher) match(value int) bool    { return true }
func (m *cronAlwaysTrueMatcher) nextAfter(value int) int { return value }

// cronBoolArrayMatcher 布尔数组匹配器：把域的有效值列表展开为布尔数组按位匹配
// （对照 hutool BoolArrayMatcher），并记录用户定义的最小/最大值供进位推算使用。
type cronBoolArrayMatcher struct {
	bValues  []bool // 下标即域值，true 表示匹配
	minValue int    // 用户定义此域的最小值
	maxValue int    // 用户定义此域的最大值
}

// newCronBoolArrayMatcher 按有效值列表构造布尔数组匹配器
func newCronBoolArrayMatcher(values []int) *cronBoolArrayMatcher {
	max := 0
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	bValues := make([]bool, max+1)
	min := max
	for _, v := range values {
		if v < min {
			min = v
		}
		bValues[v] = true
	}
	return &cronBoolArrayMatcher{bValues: bValues, minValue: min, maxValue: max}
}

// match 值落在用户定义的 [min, max] 区间内时按数组位匹配，区间外恒不匹配
func (m *cronBoolArrayMatcher) match(value int) bool {
	if value >= m.minValue && value <= m.maxValue {
		return m.bValues[value]
	}
	return false
}

// nextAfter 返回 value 及之后的最近匹配值；已越过全部匹配值时回绕到最小值
func (m *cronBoolArrayMatcher) nextAfter(value int) int {
	if value == m.maxValue {
		return value
	}
	if value > m.minValue && value < m.maxValue {
		// 最大值恒小于数组长度，只需判断最大值边界
		for value <= m.maxValue {
			if value == m.maxValue || m.bValues[value] {
				// 达到最大值或到达第一个匹配值
				return value
			}
			value++
		}
	}
	// 给定值小于最小值时下一个匹配即最小值；越过最大值时回绕到下一轮最小值
	return m.minValue
}

func (m *cronBoolArrayMatcher) getMinValue() int { return m.minValue }
func (m *cronBoolArrayMatcher) getMaxValue() int { return m.maxValue }

// cronDayOfMonthMatcher 日域匹配器（对照 hutool DayOfMonthMatcher）。
// 月的实际天数不同且存在闰年，日域单独处理；值 32（cronLastDay）表示用户书写了 "L"。
type cronDayOfMonthMatcher struct {
	cronBoolArrayMatcher
}

// matchDay 判断给定日是否匹配，同时考虑普通日期值与 "L" 月末语义。
func (m *cronDayOfMonthMatcher) matchDay(dayValue, month int, isLeapYear bool) bool {
	return m.cronBoolArrayMatcher.match(dayValue) || // 命中用户定义的某一天
		m.matchLastDay(dayValue, month, isLeapYear) // 或命中 "L" 指定的月末
}

// matchLastDay 检查 dayValue 是否为表达式含 "L" 时当月的最后一天：
// 闰年 2 月匹配 29，其它月份匹配实际最后一天（28/30/31）。
func (m *cronDayOfMonthMatcher) matchLastDay(dayValue, month int, isLeapYear bool) bool {
	return dayValue > 27 && // 月末只可能是 28 及以上，先做粗筛
		m.cronBoolArrayMatcher.match(cronLastDay) && // 表达式中定义了 "L"
		dayValue == cronGetLastDay(month, isLeapYear) // 给定日正好是当月最后一天
}

// nextAfterDay 返回指定日之后（含当天）的最近匹配日，需结合月份天数与闰年计算。
func (m *cronDayOfMonthMatcher) nextAfterDay(dayValue, month int, isLeapYear bool) int {
	maxValue := m.getMaxDay(month, isLeapYear)
	minValue := m.getMinDay(month, isLeapYear)
	if dayValue > minValue {
		for dayValue <= maxValue {
			// 命中有效日；或当天是当月最后一天且表达式含 "L"（最后一天可能不在有效值列表中）
			if m.bValues[dayValue] || (dayValue == maxValue && m.cronBoolArrayMatcher.match(cronLastDay)) {
				return dayValue
			}
			dayValue++
		}
	}
	// 给定日小于最小匹配日，或已越过当月最大匹配日，均回绕到本月最小匹配日
	return minValue
}

// getMinDay 返回表达式中指定月的最小匹配日；表达式含 "L" 时为当月最后一天。
func (m *cronDayOfMonthMatcher) getMinDay(month int, isLeapYear bool) int {
	if m.minValue == cronLastDay {
		// 用户指定了 L 表示最后一天
		return cronGetLastDay(month, isLeapYear)
	}
	return m.minValue
}

// getMaxDay 返回表达式中指定月的最大匹配日：用户定义最大值与当月最后一天取较小者。
// 注意最后一天可能不在表达式的有效值中（靠 matchLastDay 特判命中）。
func (m *cronDayOfMonthMatcher) getMaxDay(month int, isLeapYear bool) int {
	max := m.maxValue
	if last := cronGetLastDay(month, isLeapYear); last < max {
		max = last
	}
	return max
}

// cronYearValueMatcher 年域匹配器（对照 hutool YearValueMatcher）。
// 年份数值过大不适合布尔数组，使用有序去重切片匹配。
type cronYearValueMatcher struct {
	values []int // 升序去重
}

// newCronYearValueMatcher 构造年匹配器，排序去重以保证回绕语义确定
func newCronYearValueMatcher(values []int) *cronYearValueMatcher {
	sort.Ints(values)
	uniq := values[:0]
	for i, v := range values {
		if i == 0 || v != values[i-1] {
			uniq = append(uniq, v)
		}
	}
	return &cronYearValueMatcher{values: uniq}
}

func (m *cronYearValueMatcher) match(value int) bool {
	idx := sort.SearchInts(m.values, value)
	return idx < len(m.values) && m.values[idx] == value
}

// nextAfter 返回不早于 value 的最近年份；年域已无可用值时返回 -1 表示此表达式整体无效
func (m *cronYearValueMatcher) nextAfter(value int) int {
	idx := sort.SearchInts(m.values, value)
	if idx < len(m.values) {
		return m.values[idx]
	}
	return -1
}

// cronGetLastDay 返回指定月份的最后一天（month 从 1 开始，isLeapYear 表示是否闰年）
func cronGetLastDay(month int, isLeapYear bool) int {
	switch month {
	case 2:
		if isLeapYear {
			return 29
		}
		return 28
	case 4, 6, 9, 11:
		return 30
	default:
		return 31
	}
}

// cronIsLeapYear 判断是否闰年（四年一闰，百年不闰，四百年再闰）
func cronIsLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// cronPatternMatcher 单一完整表达式的匹配器，由 7 个域匹配器组成
// （对照 hutool PatternMatcher），顺序：秒 分 时 日 月 周 年。
type cronPatternMatcher struct {
	matchers []cronPartMatcher // 长度恒为 7
}

// matchFields 判断字段值数组是否匹配，dayOfMonth 与 dayOfWeek 为普通 AND 关系
// （hutool 无 Quartz 的 ? 互斥 OR 逻辑，? 完全等价 *）。
func (m *cronPatternMatcher) matchFields(fields *[7]int) bool {
	second := fields[cronPartSecond]
	return (second < 0 || m.matchers[cronPartSecond].match(second)) && // 秒为 -1 哨兵时视为恒匹配
		m.matchers[cronPartMinute].match(fields[cronPartMinute]) && // 分
		m.matchers[cronPartHour].match(fields[cronPartHour]) && // 时
		m.matchDayOfMonth(fields) && // 日（结合月份与闰年）
		m.matchers[cronPartMonth].match(fields[cronPartMonth]) && // 月
		m.matchers[cronPartDayOfWeek].match(fields[cronPartDayOfWeek]) && // 周
		m.matchers[cronPartYear].match(fields[cronPartYear]) // 年
}

// matchDayOfMonth 日域匹配：若为 DayOfMonthMatcher 则结合月份与闰年单独判断
func (m *cronPatternMatcher) matchDayOfMonth(fields *[7]int) bool {
	if dm, ok := m.matchers[cronPartDayOfMonth].(*cronDayOfMonthMatcher); ok {
		return dm.matchDay(fields[cronPartDayOfMonth], fields[cronPartMonth], cronIsLeapYear(fields[cronPartYear]))
	}
	return m.matchers[cronPartDayOfMonth].match(fields[cronPartDayOfMonth])
}

// nextMatchAfter 计算字段值之后的下一个匹配时间并组装为具体时间（对照 hutool
// PatternMatcher.nextMatchAfter）。周域不参与进位，由 CronPattern.NextAfter 的
// 按天回退保证周匹配。ok=false 表示此表达式已不可能再匹配（如年域耗尽）。
func (m *cronPatternMatcher) nextMatchAfter(values *[7]int, loc *time.Location) (time.Time, bool) {
	if !m.nextMatchValuesAfter(values) {
		return time.Time{}, false
	}
	return time.Date(values[cronPartYear], time.Month(values[cronPartMonth]),
		values[cronPartDayOfMonth], values[cronPartHour],
		values[cronPartMinute], values[cronPartSecond], 0, loc), true
}

// nextMatchValuesAfter 逐域推算下一个匹配的字段值（对照 hutool nextMatchValuesAfter）。
// 方法是自年向秒逐域取"下一个匹配值"：
//   - 结果不变：继续检查下一个（更细粒度）域；
//   - 结果变大：此域已进位，以下所有域置最小值，结束；
//   - 结果变小（回绕）：回退到上一个（更粗粒度）域强制 +1 取新值。
func (m *cronPatternMatcher) nextMatchValuesAfter(values *[7]int) bool {
	i := cronPartYear
	nextValue := 0
	rollback := false
	for i >= 0 {
		if i == cronPartDayOfWeek {
			i-- // 周不参与计算
			continue
		}
		nextValue = m.getNextMatch(values, i, 0)
		if nextValue > values[i] {
			// 此域正常获取新值，结束循环，后续的域置最小值
			values[i] = nextValue
			i--
			break
		} else if nextValue < values[i] {
			// 此域产生回退，回到上一个域继续获取新值
			i++
			nextValue = -1 // 标记回退查找
			rollback = true
			break
		}
		// 值不变，检查下一个域
		i--
	}

	if rollback {
		// 值产生回退，向上查找可变更的域（+1 强制前进）
		for i <= cronPartYear {
			if i == cronPartDayOfWeek {
				i++ // 周不参与计算
				continue
			}
			nextValue = m.getNextMatch(values, i, 1)
			if nextValue > values[i] {
				values[i] = nextValue
				i--
				break
			}
			i++
		}
		if i > cronPartYear {
			// 所有域（含年）都无法前进，此表达式整体不可能再匹配
			// （对照 hutool 此处会数组越界，这里优雅返回 false）
			return false
		}
	}

	// 变更域以下的所有域全部归最小值
	m.setToMin(values, i)
	return true
}

// getNextMatch 获取指定域在其当前值偏移 plus 后的下一个匹配值（对照 hutool getNextMatch）。
// 日域需考虑月份天数与闰年，单独处理。
func (m *cronPatternMatcher) getNextMatch(values *[7]int, part, plus int) int {
	if part == cronPartDayOfMonth {
		if dm, ok := m.matchers[part].(*cronDayOfMonthMatcher); ok {
			return dm.nextAfterDay(values[part]+plus, values[cronPartMonth], cronIsLeapYear(values[cronPartYear]))
		}
	}
	return m.matchers[part].nextAfter(values[part] + plus)
}

// setToMin 把从指定域到秒域的全部字段置为各自的最小匹配值（对照 hutool setToMin）
func (m *cronPatternMatcher) setToMin(values *[7]int, toPart int) {
	for i := toPart; i >= 0; i-- {
		if i == cronPartDayOfMonth {
			if dm, ok := m.matchers[i].(*cronDayOfMonthMatcher); ok {
				// "L" 时最小值为当月最后一天
				values[i] = dm.getMinDay(values[cronPartMonth], cronIsLeapYear(values[cronPartYear]))
				continue
			}
		}
		values[i] = m.minOf(i)
	}
}

// minOf 获取域的最小匹配值（对照 hutool PatternMatcher.getMin）
func (m *cronPatternMatcher) minOf(part int) int {
	switch t := m.matchers[part].(type) {
	case *cronAlwaysTrueMatcher:
		// 全匹配器无有效值集合，取域范围的最小值
		return cronPartDefs[part].min
	case interface{ getMinValue() int }:
		return t.getMinValue()
	default:
		return cronPartDefs[part].min
	}
}

/* ---------------- CronPattern：编译与解析（对照 hutool CronPattern / PatternParser / PartParser） ---------------- */

// CronPattern 编译后的定时任务表达式（对照 hutool CronPattern）。
//
// 表达式用空格分隔为 5、6 或 7 个域，按序为：
//
//	5 域：分 时 日 月 周
//	6 域：秒 分 时 日 月 周（第 1 位为秒）
//	7 域：秒 分 时 日 月 周 年（末位为年，1970~2099）
//
// 每个域支持：* 与 ?（两者等价）、数值 n、闭区间 n-m（支持 5-2 反向环绕）、
// 步进 a/b 与 */b、逗号列表（元素可再带范围或步进，/ 优先级高于 - 高于 ,）。
// 月域支持 jan~dec 别名，周域支持 sun~sat 别名（不区分大小写，7 归一化为 0，0 表示周日）；
// 日域支持 "L" 表示当月最后一天（内部以 32 哨兵表示，感知大小月与闰年）。
// 支持 | 分隔多个完整表达式，任一匹配即可。
// 注意：不支持 Quartz 的 W（最近工作日）、#（第 n 个星期）、C 扩展语法，解析到时返回 error。
type CronPattern struct {
	pattern  string                // 原始表达式
	matchers []*cronPatternMatcher // 每个 | 分隔的完整表达式对应一个匹配器
}

// newCronPattern 解析并编译 cron 表达式（utils.NewCronPattern 的实际实现，对照 hutool CronPattern.of）。
// 解析失败时返回包含域位置与原因的描述性 error。
func newCronPattern(pattern string) (*CronPattern, error) {
	matchers := make([]*cronPatternMatcher, 0, 1)
	for _, single := range cronSplitTrim(pattern, '|') {
		m, err := cronParseSingle(single)
		if err != nil {
			return nil, err
		}
		matchers = append(matchers, m)
	}
	if len(matchers) == 0 {
		return nil, fmt.Errorf("cron 表达式 [%s] 无效: 内容为空", pattern)
	}
	return &CronPattern{pattern: pattern, matchers: matchers}, nil
}

// Pattern 返回原始表达式字符串（对照 hutool toString）。
func (p *CronPattern) Pattern() string {
	return p.pattern
}

// Match 判断给定时间是否匹配表达式（对照 hutool CronPattern.match）。
// matchSecond 为 false 时秒域视为恒匹配（-1 哨兵语义），适用于按分钟触发的调度；
// 为 true 时秒域参与匹配。时间按 t 的时区取字段。
func (p *CronPattern) Match(t time.Time, matchSecond bool) bool {
	fields := cronGetFields(t, matchSecond)
	return p.matchFields(&fields)
}

// NextAfter 返回 t 之后的下一次触发时间；若 t 本身匹配则从 t+1s 起算。
// 算法对照 hutool CronPattern.nextMatchAfter / nextMatch：各域自年向秒逐级进位，
// 因周域不参与进位，结果不匹配时回退到次日 0 点重算直至匹配。
// 多个 | 子表达式时返回最早者；再也无匹配（如年域耗尽）时返回零值 time.Time{}。
// 场景：计算任务下次执行时间用于展示或延时调度。
func (p *CronPattern) NextAfter(t time.Time, matchSecond bool) time.Time {
	if p.Match(t, matchSecond) {
		// 给定时间已匹配表达式时增加 1 秒以匹配下一个时间（对照 hutool issue#I9FQUA）
		t = t.Add(time.Second)
	}
	for {
		fields := cronGetFields(t, true)
		next, ok := p.nextMatchAfter(&fields, t.Location())
		if !ok {
			return time.Time{}
		}
		if p.matchTime(next) {
			return next
		}
		// 周域导致不匹配：回退到次日 0 点重新推算（对照 hutool nextMatch 的按天递归）
		t = time.Date(next.Year(), next.Month(), next.Day()+1, 0, 0, 0, 0, next.Location())
		if t.Year() > 2200 {
			// 防御性终止：任何真实表达式在格雷戈里历 400 年周循环内必然存在匹配，
			// 递归至此仍无匹配只可能是异常构造的表达式，视为无下一次
			return time.Time{}
		}
	}
}

// nextMatchAfter 返回多个 | 子表达式中最早的下一次匹配时间（对照 hutool CronPattern.nextMatchAfter）
func (p *CronPattern) nextMatchAfter(values *[7]int, loc *time.Location) (time.Time, bool) {
	var best time.Time
	found := false
	for _, m := range p.matchers {
		vals := *values // 每个子表达式独立推算，避免相互污染
		next, ok := m.nextMatchAfter(&vals, loc)
		if !ok {
			continue
		}
		if !found || next.Before(best) {
			best, found = next, true
		}
	}
	return best, found
}

// matchTime 判断时间是否完整匹配（含秒，对照 hutool nextMatch 内部的 match(next, true)）
func (p *CronPattern) matchTime(t time.Time) bool {
	fields := cronGetFields(t, true)
	return p.matchFields(&fields)
}

// matchFields 任一子表达式匹配即认为整体匹配（对照 hutool CronPattern.match）
func (p *CronPattern) matchFields(fields *[7]int) bool {
	for _, m := range p.matchers {
		if m.matchFields(fields) {
			return true
		}
	}
	return false
}

// cronGetFields 提取时间对应的字段数组 {秒, 分, 时, 日, 月, 周, 年}
// （对照 hutool PatternUtil.getFields）。matchSecond 为 false 时秒以 -1 哨兵表示恒匹配。
// 月从 1 开始；周从 0 开始且 0 表示周日（Go 的 Weekday 与 hutool Week 序数一致）。
func cronGetFields(t time.Time, matchSecond bool) [7]int {
	second := -1
	if matchSecond {
		second = t.Second()
	}
	return [7]int{
		second,
		t.Minute(),
		t.Hour(),
		t.Day(),
		int(t.Month()),
		int(t.Weekday()),
		t.Year(),
	}
}

// cronParseSingle 解析单个完整表达式（不含 |，对照 hutool PatternParser.parseSingle）
func cronParseSingle(pattern string) (*cronPatternMatcher, error) {
	parts := strings.Fields(pattern) // 按连续空白切分（对照 split("\\s+")）
	if len(parts) < 5 || len(parts) > 7 {
		return nil, fmt.Errorf("cron 表达式 [%s] 无效: 域数量必须为 5~7 个（空格分隔），实际为 %d 个", pattern, len(parts))
	}

	// 偏移量用于兼容 Quartz：6/7 位表达式的第 1 位为秒
	offset := 0
	if len(parts) == 6 || len(parts) == 7 {
		offset = 1
	}

	// 秒：5 位表达式不支持秒，固定为 0 表示整分匹配
	secondPart := "0"
	if offset == 1 {
		secondPart = parts[0]
	}

	matchers := make([]cronPartMatcher, 7)
	var err error
	// 秒
	if matchers[cronPartSecond], err = cronParsePart(cronPartSecond, secondPart); err != nil {
		return nil, err
	}
	// 分
	if matchers[cronPartMinute], err = cronParsePart(cronPartMinute, parts[offset]); err != nil {
		return nil, err
	}
	// 时
	if matchers[cronPartHour], err = cronParsePart(cronPartHour, parts[1+offset]); err != nil {
		return nil, err
	}
	// 日
	if matchers[cronPartDayOfMonth], err = cronParsePart(cronPartDayOfMonth, parts[2+offset]); err != nil {
		return nil, err
	}
	// 月
	if matchers[cronPartMonth], err = cronParsePart(cronPartMonth, parts[3+offset]); err != nil {
		return nil, err
	}
	// 周
	if matchers[cronPartDayOfWeek], err = cronParsePart(cronPartDayOfWeek, parts[4+offset]); err != nil {
		return nil, err
	}
	// 年：7 位表达式解析年域，否则全匹配
	if len(parts) == 7 {
		if matchers[cronPartYear], err = cronParsePart(cronPartYear, parts[6]); err != nil {
			return nil, err
		}
	} else {
		matchers[cronPartYear] = &cronAlwaysTrueMatcher{}
	}
	return &cronPatternMatcher{matchers: matchers}, nil
}

// cronParsePart 解析单个域表达式为匹配器（对照 hutool PartParser.parse）：
// * 或 ? 返回全匹配器；日域返回日匹配器；年域返回年匹配器；其余返回布尔数组匹配器。
func cronParsePart(part int, value string) (cronPartMatcher, error) {
	if cronIsMatchAllStr(value) {
		// 兼容 Quartz 的 "?"，与 "*" 作用一致
		return &cronAlwaysTrueMatcher{}, nil
	}

	values, err := cronParseArray(part, value)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("cron 表达式的 %s 域 [%s] 无效: 未解析出有效值", cronPartDefs[part].name, value)
	}

	switch part {
	case cronPartDayOfMonth:
		return &cronDayOfMonthMatcher{cronBoolArrayMatcher: *newCronBoolArrayMatcher(values)}, nil
	case cronPartYear:
		return newCronYearValueMatcher(values), nil
	default:
		return newCronBoolArrayMatcher(values), nil
	}
}

// cronParseArray 处理列表形式表达式（对照 hutool PartParser.parseArray）：
// 按逗号拆分逐项解析并去重（保持首次出现顺序）。
func cronParseArray(part int, value string) ([]int, error) {
	var values []int
	for _, item := range strings.Split(value, ",") {
		results, err := cronParseStep(part, item)
		if err != nil {
			return nil, err
		}
		for _, v := range results {
			if !cronContainsInt(values, v) {
				values = append(values, v)
			}
		}
	}
	return values, nil
}

// cronParseStep 处理步进形式表达式（对照 hutool PartParser.parseStep）：
// 支持 a、*、a/b、*/b、a-b/2 等形式；步进值不做取值范围校验（对照 hutool issue#I7SMP7）。
func cronParseStep(part int, value string) ([]int, error) {
	slash := strings.Index(value, "/")
	if slash < 0 {
		// 普通形式：无步进
		return cronParseRange(part, value, -1)
	}
	head, stepPart := value[:slash], value[slash+1:]
	if strings.Contains(stepPart, "/") {
		return nil, fmt.Errorf("cron 表达式的 %s 域 [%s] 无效: 步进分隔符 / 出现多次", cronPartDefs[part].name, value)
	}
	step, err := cronParseNumber(part, stepPart, false)
	if err != nil {
		return nil, err
	}
	if step < 1 {
		return nil, fmt.Errorf("cron 表达式的 %s 域 [%s] 无效: 步进值必须为正数", cronPartDefs[part].name, value)
	}
	return cronParseRange(part, head, step)
}

// cronParseRange 处理范围表达式（对照 hutool PartParser.parseRange）：
// 支持 *、2、3-8、8-3（反向环绕）、jan、L 等形式；step < 1 表示无步进。
func cronParseRange(part int, value string, step int) ([]int, error) {
	def := &cronPartDefs[part]
	var results []int

	// 短形式：长度 <= 2 的 "*"、"?" 或 1~2 位数字（如 12/3 的起点 12）
	if len(value) <= 2 {
		minValue := def.min
		if !cronIsMatchAllStr(value) {
			n, err := cronParseNumber(part, value, true)
			if err != nil {
				return nil, err
			}
			if n > minValue {
				minValue = n
			}
		} else if step < 1 {
			// 全匹配模式下无步进表示步进为 1
			step = 1
		}
		if step > 0 {
			if minValue > def.max {
				return nil, fmt.Errorf("cron 表达式的 %s 域无效: 起始值 %d 大于最大值 %d", def.name, minValue, def.max)
			}
			for i := minValue; i <= def.max; i += step {
				results = append(results, i)
			}
		} else {
			// 固定时间
			results = append(results, minValue)
		}
		return results, nil
	}

	// 范围模式：按 - 拆分
	dash := strings.Index(value, "-")
	if dash < 0 {
		// 普通值（含别名，如 jan、L）：类似 20/2 的形式从该值步进到域最大值
		v1, err := cronParseNumber(part, value, true)
		if err != nil {
			return nil, err
		}
		if step > 0 {
			cronAppendRange(&results, v1, def.max, step)
		} else {
			results = append(results, v1)
		}
		return results, nil
	}

	v1, err := cronParseNumber(part, value[:dash], true)
	if err != nil {
		return nil, err
	}
	v2, err := cronParseNumber(part, value[dash+1:], true)
	if err != nil {
		return nil, err
	}
	if step < 1 {
		// range 模式下无步进表示步进为 1
		step = 1
	}
	if v1 <= v2 {
		// 正常范围，例如 2-5、3-3
		cronAppendRange(&results, v1, v2, step)
	} else {
		// 逆向范围按环绕反选，例如周域 5-2 表示周五到周二跨周回绕
		cronAppendRange(&results, v1, def.max, step)
		cronAppendRange(&results, def.min, v2, step)
	}
	return results, nil
}

// cronAppendRange 把 [start, stop] 区间按 step 步进的整数追加到 results
// （对照 hutool NumberUtil.appendRange，start 与 stop 均包含）
func cronAppendRange(results *[]int, start, stop, step int) {
	if start == stop {
		*results = append(*results, start)
		return
	}
	if start < stop {
		if step < 0 {
			step = -step
		}
		for i := start; i <= stop; i += step {
			*results = append(*results, i)
		}
	} else {
		if step < 0 {
			step = -step
		}
		for i := start; i >= stop; i -= step {
			*results = append(*results, i)
		}
	}
}

// cronParseNumber 解析单个数值，支持别名（对照 hutool PartParser.parseNumber）。
// checkValue 为 true 时校验值在域的有效范围内。负数按 hutool 规则 v += max；
// 周域的 7 归一化为 0（0 与 7 均表示周日）。
func cronParseNumber(part int, value string, checkValue bool) (int, error) {
	i, err := strconv.Atoi(value)
	if err != nil {
		// 非数字时尝试按别名解析
		i, err = cronParseAlias(part, value)
		if err != nil {
			return 0, err
		}
	}

	// 支持负数：v < 0 则 v += max（对照 hutool 规则）
	if i < 0 {
		i += cronPartDefs[part].max
	}

	// 周日可用 0 或 7 表示，统一归一化为 0（对照 hutool Week.SUNDAY）
	if part == cronPartDayOfWeek && i == 7 {
		i = 0
	}

	if checkValue {
		def := &cronPartDefs[part]
		if i < def.min || i > def.max {
			return 0, fmt.Errorf("cron 表达式的 %s 域值 %d 超出范围 [%d, %d]", def.name, i, def.min, def.max)
		}
	}
	return i, nil
}

// cronParseAlias 解析别名（对照 hutool PartParser.parseAlias）：
// "L" 表示域允许的最大值（日域为 32 即月末哨兵，周域为 6 即周六）；
// 月域支持 jan~dec，周域支持 sun~sat，均不区分大小写。
func cronParseAlias(part int, name string) (int, error) {
	if strings.EqualFold(name, "L") {
		return cronPartDefs[part].max, nil
	}
	switch part {
	case cronPartMonth:
		if v, ok := cronParseMonthAlias(name); ok {
			return v, nil
		}
	case cronPartDayOfWeek:
		if v, ok := cronParseWeekAlias(name); ok {
			return v, nil
		}
	}
	return 0, fmt.Errorf("cron 表达式的 %s 域 [%s] 无效: 无法识别的别名（注意不支持 Quartz 的 W、#、C 扩展语法）",
		cronPartDefs[part].name, name)
}

// cronMonthAliases 月域英文别名，下标 i 对应月份 i+1（对照 hutool Month.ALIASES）
var cronMonthAliases = []string{"jan", "feb", "mar", "apr", "may", "jun", "jul", "aug", "sep", "oct", "nov", "dec"}

// cronWeekAliases 周域英文别名，下标 i 对应星期值 i（0 为周日，对照 hutool Week.ALIASES）
var cronWeekAliases = []string{"sun", "mon", "tue", "wed", "thu", "fri", "sat"}

// cronParseMonthAlias 解析月别名为 1~12
func cronParseMonthAlias(name string) (int, bool) {
	for i, alias := range cronMonthAliases {
		if strings.EqualFold(alias, name) {
			return i + 1, true
		}
	}
	return 0, false
}

// cronParseWeekAlias 解析周别名为 0~6（0 为周日）
func cronParseWeekAlias(name string) (int, bool) {
	for i, alias := range cronWeekAliases {
		if strings.EqualFold(alias, name) {
			return i, true
		}
	}
	return 0, false
}

// cronIsMatchAllStr 是否为全匹配符 * 或 ?（对照 hutool PartParser.isMatchAllStr）
func cronIsMatchAllStr(value string) bool {
	return value == "*" || value == "?"
}

// cronSplitTrim 按分隔符拆分并去除每项首尾空白，丢弃空白项（对照 hutool StrUtil.splitTrim）
func cronSplitTrim(value string, sep rune) []string {
	var result []string
	for _, item := range strings.Split(value, string(sep)) {
		if item = strings.TrimSpace(item); item != "" {
			result = append(result, item)
		}
	}
	return result
}

// cronContainsInt 切片中是否包含指定整数
func cronContainsInt(values []int, target int) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}

/* ---------------- 调度器（对照 hutool Scheduler / TaskTable / CronTimer / TaskExecutor / listener 包） ---------------- */

// TaskListener 定时任务监听器（对照 hutool TaskListener 接口，Go 中以结构体字段代替接口方法）。
// 各回调字段均可为 nil，为 nil 时跳过对应环节的通知。
// 场景：记录任务执行日志、上报执行耗时、告警失败任务。
type TaskListener struct {
	// OnStart 任务开始执行前触发，参数为任务 ID（对照 onStart）
	OnStart func(id string)
	// OnSucceeded 任务执行成功返回后触发，参数为任务 ID（对照 onSucceeded）
	OnSucceeded func(id string)
	// OnFailed 任务执行发生 panic 时触发，参数为任务 ID 与 recover 捕获到的值（对照 onFailed）
	OnFailed func(id string, panicked interface{})
}

// cronTaskEntry 任务表条目：ID 与表达式、任务体一一对应（对照 hutool CronTask）
type cronTaskEntry struct {
	pattern *CronPattern
	task    func()
}

// cronTaskRef 匹配快照中的任务引用，避免持锁执行任务
type cronTaskRef struct {
	id   string
	task func()
}

// cronListenerEntry 监听器与句柄的绑定，句柄用于移除
type cronListenerEntry struct {
	handle   int
	listener TaskListener
}

// Scheduler 定时任务调度器（对照 hutool Scheduler）。
//
// 线程模型：Start 后由一个常驻 goroutine 充当计时器（对照 CronTimer），
// 按分钟或秒对齐边界遍历任务表，表达式匹配的任务各起一个 goroutine 执行
// （对照 TaskLauncher/TaskExecutor 的线程池语义，Go 中直接以 goroutine 等价实现）。
// 任务 panic 会被 recover 并回调 OnFailed，不会拖垮调度器与其它任务。
// 所有操作并发安全。
type Scheduler struct {
	mu          sync.RWMutex
	matchSecond bool // 是否启用秒级匹配
	tasks       map[string]*cronTaskEntry
	started     bool
	stopCh      chan struct{} // 关闭即通知定时 goroutine 退出
	listeners   []cronListenerEntry
	listenerSeq int // 监听器句柄发号器
}

// newScheduler 创建调度器（utils.NewScheduler 的实际实现，对照 hutool new Scheduler）。
// matchSecond 为 true 时按秒级匹配（6/7 位表达式第 1 位为秒），否则按分钟匹配（默认）。
func newScheduler(matchSecond bool) *Scheduler {
	return &Scheduler{
		matchSecond: matchSecond,
		tasks:       make(map[string]*cronTaskEntry),
	}
}

// SetMatchSecond 设置是否启用秒级匹配，返回 this 以支持链式调用
// （对照 hutool Scheduler.setMatchSecond）。
// 仅在 Start 之前调用有效；调度器启动后设置将被忽略。
func (s *Scheduler) SetMatchSecond(matchSecond bool) *Scheduler {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return s
	}
	s.matchSecond = matchSecond
	return s
}

// Schedule 新增定时任务，使用自动生成的 UUID 风格任务 ID（对照 hutool schedule(pattern, task)）。
// 表达式编译失败时返回 error。场景：注册无需外部引用的后台作业。
func (s *Scheduler) Schedule(pattern string, task func()) (id string, err error) {
	compiled, err := newCronPattern(pattern)
	if err != nil {
		return "", err
	}
	id = cronFastUUID()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[id] = &cronTaskEntry{pattern: compiled, task: task}
	return id, nil
}

// ScheduleWithId 以指定 ID 新增定时任务，ID 已存在或表达式非法时返回 error
// （对照 hutool schedule(id, pattern, task)，重复 ID 在 hutool 中抛异常）。
func (s *Scheduler) ScheduleWithId(id, pattern string, task func()) error {
	compiled, err := newCronPattern(pattern)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[id]; ok {
		return fmt.Errorf("定时任务 ID [%s] 已存在", id)
	}
	s.tasks[id] = &cronTaskEntry{pattern: compiled, task: task}
	return nil
}

// Deschedule 移除指定 ID 的任务，返回是否移除成功，false 表示未找到对应 ID 的任务
// （对照 hutool descheduleWithStatus）。正在执行中的任务不受影响，将自然执行完毕。
func (s *Scheduler) Deschedule(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[id]; !ok {
		return false
	}
	delete(s.tasks, id)
	return true
}

// UpdatePattern 更新指定 ID 任务的表达式，表达式编译失败或 ID 不存在时返回 error
// （对照 hutool updatePattern）。适用于运行期动态调整任务的触发规则。
func (s *Scheduler) UpdatePattern(id, pattern string) error {
	compiled, err := newCronPattern(pattern)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.tasks[id]
	if !ok {
		return fmt.Errorf("未找到 ID 为 [%s] 的定时任务", id)
	}
	entry.pattern = compiled
	return nil
}

// GetPattern 返回指定 ID 任务编译后的表达式，ID 不存在时第二个返回值为 false
// （对照 hutool getPattern）。
func (s *Scheduler) GetPattern(id string) (*CronPattern, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.tasks[id]
	if !ok {
		return nil, false
	}
	return entry.pattern, true
}

// Size 返回当前任务数（对照 hutool size）。
func (s *Scheduler) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.tasks)
}

// Clear 清空任务表，不影响调度器运行状态（对照 hutool clear）。
func (s *Scheduler) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks = make(map[string]*cronTaskEntry)
}

// IsStarted 返回调度器是否已启动（对照 hutool isStarted）。
func (s *Scheduler) IsStarted() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.started
}

// Start 启动调度器（对照 hutool Scheduler.start）：立即按当前时间对齐边界触发一次匹配，
// 之后按分钟（matchSecond=true 时按秒）对齐边界循环检查任务表。
// 已启动时为空操作并返回 nil。
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return nil
	}
	stopCh := make(chan struct{})
	s.stopCh = stopCh
	s.started = true
	matchSecond := s.matchSecond
	go s.cronTimerLoop(stopCh, matchSecond)
	return nil
}

// Stop 停止调度器的定时循环；clearTasks 为 true 时同时清空任务表
// （对照 hutool stop(boolean)）。未启动时为空操作。
// 注意：已提交执行的 任务 goroutine 不会被中断，将自然执行完毕。
func (s *Scheduler) Stop(clearTasks bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.started {
		return
	}
	close(s.stopCh)
	s.stopCh = nil
	s.started = false
	if clearTasks {
		s.tasks = make(map[string]*cronTaskEntry)
	}
}

// AddListener 注册任务监听器，返回用于 RemoveListener 的句柄（对照 hutool addListener）。
func (s *Scheduler) AddListener(listener TaskListener) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listenerSeq++
	handle := s.listenerSeq
	s.listeners = append(s.listeners, cronListenerEntry{handle: handle, listener: listener})
	return handle
}

// RemoveListener 按句柄移除监听器，返回是否移除成功（对照 hutool removeListener）。
func (s *Scheduler) RemoveListener(handle int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, le := range s.listeners {
		if le.handle == handle {
			s.listeners = append(s.listeners[:i], s.listeners[i+1:]...)
			return true
		}
	}
	return false
}

// cronTimerLoop 调度器的定时循环（对照 hutool CronTimer.run）：
// 先按当前时间对齐边界触发一次匹配，随后睡眠到下一个边界；
// 处理耗时跨过多个边界时逐边界补触发直到追上（对照 hutool catch-up 逻辑）；
// 睡眠中被 Stop 通知即退出；系统时间被回拨时放弃长睡眠重新对齐（对照 hutool issue#1224）。
func (s *Scheduler) cronTimerLoop(stopCh <-chan struct{}, matchSecond bool) {
	unitSecs := int64(60) // 定时单位（秒）：默认按分钟对齐
	if matchSecond {
		unitSecs = 1 // 秒级模式按秒对齐
	}
	// 对齐到当前边界（除法清零单位以下部分，对照 hutool 的边界计算）
	thisTime := (time.Now().Unix() / unitSecs) * unitSecs
	s.spawnMatch(time.Unix(thisTime, 0), matchSecond)
	nextTime := thisTime + unitSecs

	for {
		now := time.Now().Unix()
		if nextTime > now {
			sleep := nextTime - now
			if sleep > 2*unitSecs {
				// 距下一边界过远，判定系统时间被回拨，按当前时间重新对齐
				thisTime = (now / unitSecs) * unitSecs
				s.spawnMatch(time.Unix(thisTime, 0), matchSecond)
				nextTime = thisTime + unitSecs
				continue
			}
			// 睡眠到下一个边界，期间可被 Stop 通知退出
			timer := time.NewTimer(time.Duration(sleep) * time.Second)
			select {
			case <-stopCh:
				timer.Stop()
				return
			case <-timer.C:
			}
			s.spawnMatch(time.Unix(nextTime, 0), matchSecond)
			nextTime += unitSecs
		} else {
			// 处理耗时跨过了边界（或系统时间被前调），逐边界补触发直到追上当前时间
			for nextTime <= now {
				s.spawnMatch(time.Unix(nextTime, 0), matchSecond)
				nextTime += unitSecs
			}
		}
	}
}

// spawnMatch 遍历任务表，把匹配此时刻的任务各起一个 goroutine 执行
// （对照 hutool TaskLauncher 检查任务表 + TaskExecutorManager.spawnExecutor；
// hutool 使用无界线程池保证任务及时运行，Go 中直接以 goroutine 等价实现）。
func (s *Scheduler) spawnMatch(t time.Time, matchSecond bool) {
	s.mu.RLock()
	matched := make([]cronTaskRef, 0, len(s.tasks))
	for id, entry := range s.tasks {
		if entry.pattern.Match(t, matchSecond) {
			matched = append(matched, cronTaskRef{id: id, task: entry.task})
		}
	}
	listeners := make([]TaskListener, len(s.listeners))
	for i, le := range s.listeners {
		listeners[i] = le.listener
	}
	s.mu.RUnlock()

	for _, ref := range matched {
		go s.runTask(ref.id, ref.task, listeners)
	}
}

// runTask 执行单个任务并通知监听器；recover 任务的 panic 并回调 OnFailed
// （对照 hutool TaskExecutor.run + TaskListenerManager）。
func (s *Scheduler) runTask(id string, task func(), listeners []TaskListener) {
	for _, l := range listeners {
		if l.OnStart != nil {
			l.OnStart(id)
		}
	}

	var panicked interface{}
	func() {
		defer func() { panicked = recover() }()
		task()
	}()

	if panicked != nil {
		// 对照 hutool 无监听器时打印堆栈；Go 侧无日志设施，此处静默忽略
		for _, l := range listeners {
			if l.OnFailed != nil {
				l.OnFailed(id, panicked)
			}
		}
		return
	}
	for _, l := range listeners {
		if l.OnSucceeded != nil {
			l.OnSucceeded(id)
		}
	}
}

// cronFastUUID 生成 UUID v4 风格的任务 ID（8-4-4-4-12 带连字符，对照 hutool IdUtil.fastUUID）
func cronFastUUID() string {
	var uuid [16]byte
	// Go 1.24 起 crypto/rand.Read 保证填满且不返回错误；保留检查以兼容旧版本语义
	if _, err := rand.Read(uuid[:]); err != nil {
		return ""
	}
	uuid[6] = (uuid[6] & 0x0F) | 0x40 // 版本号置为 4（随机 UUID）
	uuid[8] = (uuid[8] & 0x3F) | 0x80 // 变体位置为 RFC 4122
	buf := make([]byte, 36)
	hex.Encode(buf, uuid[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], uuid[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], uuid[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], uuid[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], uuid[10:16])
	return string(buf)
}

/* ---------------- utilCron 门面（对照 hutool CronUtil） ---------------- */

// utilCron 定时任务工具门面，持有包级默认调度器（对照 hutool CronUtil）。
// 通过 utils.Cron 调用；需要相互独立的多套调度时用 utils.NewScheduler 另建实例。
type utilCron struct{}

var (
	cronDefaultOnce     sync.Once  // 保证默认调度器只创建一次
	cronDefaultSchedule *Scheduler // 包级默认调度器（懒创建）
)

// cronGetDefaultScheduler 懒创建并返回包级默认调度器（对照 hutool CronUtil 的静态 scheduler）
func cronGetDefaultScheduler() *Scheduler {
	cronDefaultOnce.Do(func() {
		cronDefaultSchedule = newScheduler(false)
	})
	return cronDefaultSchedule
}

// SetMatchSecond 设置默认调度器是否启用秒级匹配（对照 hutool CronUtil.setMatchSecond）。
// 需在 Start 之前调用；调度器已启动时本设置被忽略。
func (c *utilCron) SetMatchSecond(matchSecond bool) {
	cronGetDefaultScheduler().SetMatchSecond(matchSecond)
}

// Schedule 向默认调度器加入定时任务，返回自动生成的任务 ID（对照 hutool CronUtil.schedule）。
// 表达式非法时返回 error。
func (c *utilCron) Schedule(pattern string, task func()) (id string, err error) {
	return cronGetDefaultScheduler().Schedule(pattern, task)
}

// ScheduleWithId 以指定 ID 向默认调度器加入定时任务（对照 hutool CronUtil.schedule(id, pattern, task)）。
// ID 重复或表达式非法时返回 error。
func (c *utilCron) ScheduleWithId(id, pattern string, task func()) error {
	return cronGetDefaultScheduler().ScheduleWithId(id, pattern, task)
}

// Remove 从默认调度器移除任务，返回是否移除成功，false 表示未找到对应 ID（对照 hutool CronUtil.remove）。
func (c *utilCron) Remove(id string) bool {
	return cronGetDefaultScheduler().Deschedule(id)
}

// UpdatePattern 更新默认调度器中指定任务的表达式（对照 hutool CronUtil.updatePattern）。
// 表达式非法或 ID 不存在时返回 error。
func (c *utilCron) UpdatePattern(id, pattern string) error {
	return cronGetDefaultScheduler().UpdatePattern(id, pattern)
}

// Start 启动默认调度器（对照 hutool CronUtil.start）；已启动时为空操作并返回 nil。
func (c *utilCron) Start() error {
	return cronGetDefaultScheduler().Start()
}

// Stop 停止默认调度器并清空任务表（对照 hutool CronUtil.stop 的 stop(true) 语义）。
func (c *utilCron) Stop() {
	cronGetDefaultScheduler().Stop(true)
}

// Restart 重启默认调度器（对照 hutool CronUtil.restart）：先停止并清空任务表，再重新启动。
func (c *utilCron) Restart() error {
	s := cronGetDefaultScheduler()
	s.Stop(true)
	return s.Start()
}

// Size 返回默认调度器当前任务数。
func (c *utilCron) Size() int {
	return cronGetDefaultScheduler().Size()
}

// IsStarted 返回默认调度器是否已启动。
func (c *utilCron) IsStarted() bool {
	return cronGetDefaultScheduler().IsStarted()
}

// GetScheduler 返回包级默认调度器实例，可进行细粒度操作
// （如添加监听器、按 ID 查询表达式，对照 hutool CronUtil.getScheduler）。
func (c *utilCron) GetScheduler() *Scheduler {
	return cronGetDefaultScheduler()
}
