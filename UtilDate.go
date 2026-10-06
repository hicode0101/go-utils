package utils

import (
	"fmt"
	"math"
	"time"
)

type utilDate struct {
}

// GO的诞辰
const timeLayout = "2006-01-02 15:04:05"
const dateLayout = "2006-01-02"

// 取当前系统时间
func (_self *utilDate) GetTimeNow() time.Time {
	return time.Now()
}

func (_self *utilDate) GetTimeNowBySH() time.Time {
	return _self.GetTimeNowByLoc("Asia/Shanghai")
}

func (_self *utilDate) GetTimeNowByLoc(locName string) time.Time {
	// Local, UTC, Asia/Shanghai, America/New_York, America/Los_Angeles
	//建议使用 timedatectl 命令在服务器上设置时区，最好不要使用代码强制转换获取
	loc, _ := time.LoadLocation(locName)
	return time.Now().In(loc)
}

func (_self *utilDate) GetTimeNowByUTC8() time.Time {
	loc := time.FixedZone("UTC+8", 8*60*60)
	return time.Now().In(loc)
}

func (_self *utilDate) GoBirthday() time.Time {
	return _self.GetTime("2006-01-02 15:04:05")
}

func (_self *utilDate) GetTime(timeStr string) time.Time {
	toTime, _ := _self.ToTime(timeStr)
	return toTime
}

func (_self *utilDate) JavaLongTime(javaLong int64) time.Time {
	//1492566520958	-> 2017-04-19 09:48:40
	//fmt.Println(time.Unix(1492566520958/1000, 0))
	//fmt.Println(time.Unix(0, 1492566520958*1000000))
	return time.Unix(0, javaLong*1000000)
}

func (_self *utilDate) LongTime(lt int64) time.Time {
	//1492566520958	-> 2017-04-19 09:48:40
	//1623819567485
	return time.Unix(0, lt*1000000)
}

func (_self *utilDate) Float64Time(lt float64) time.Time {
	//1492566520958	-> 2017-04-19 09:48:40
	return time.Unix(0, int64(lt)*1000000)
}

func (_self *utilDate) Float64TimeLocal(lt float64) time.Time {
	//1492566520958	-> 2017-04-19 09:48:40
	return time.Unix(0, int64(lt)*1000000).Local()
}

func (_self *utilDate) ToTime(timeStr string) (time.Time, error) {
	loc, _ := time.LoadLocation("Local")
	toTime, err := time.ParseInLocation(timeLayout, timeStr, loc)
	//toTime, err := time.Parse(timeLayout, timeStr)
	return toTime, err

}

func (_self *utilDate) ToTimeFromDate(timeStr string) (time.Time, error) {
	loc, _ := time.LoadLocation("Local")
	toTime, err := time.ParseInLocation(dateLayout, timeStr, loc)
	//toTime, err := time.Parse(timeLayout, timeStr)
	return toTime, err

}

func (_self *utilDate) ToTimeByFm(timeStr string, format string) (time.Time, error) {
	loc, _ := time.LoadLocation("Local")
	toTime, err := time.ParseInLocation(format, timeStr, loc)
	//toTime, err := time.Parse(timeLayout, timeStr)
	return toTime, err

}

// 要想格式化为：yyyyMMddHHmmss
// 则 format = "20060102150405"
// 要想格式化为：yyyy-MM-dd HH:mm:ss
// 则 format = "2006-01-02 15:04:05"
// 要想格式化为：yyyy-MM-dd
// 则 format = "2006-01-02"
// 2006-01-02 15:04:05.000000000
func (_self *utilDate) FormatTimeByFm(t time.Time, format string) string {

	return t.Format(format)
}

func (_self *utilDate) GetCurrentTime() string {
	return _self.FormatTime(time.Now())
}

func (_self *utilDate) GetCurrentDay() string {
	return _self.FormatTimeByFm(time.Now(), "2006-01-02")
}

func (_self *utilDate) FormatTime(t time.Time) string {
	//
	return _self.FormatTimeByFm(t, "2006-01-02 15:04:05")
}

func (_self *utilDate) FormatTimens(t time.Time) string {
	//
	return _self.FormatTimeByFm(t, "2006-01-02 15:04:05.000000000")
}

func (_self *utilDate) FormatTimeToNum(t time.Time) string {
	//
	return _self.FormatTimeByFm(t, "20060102150405")
}

func (_self *utilDate) FormatTimeToDayNum(t time.Time) string {
	//
	return _self.FormatTimeByFm(t, "20060102")
}

// 在当前时间之前
func (_self *utilDate) IsBeforeNow(t time.Time) (result bool) {
	result = false
	if &t != nil && t.Before(time.Now()) {
		result = true
	}
	return
}

// 在当前时间之后
func (_self *utilDate) IsAfterNow(t time.Time) (result bool) {
	result = false
	if &t != nil && t.After(time.Now()) {
		result = true
	}
	return
}

func (_self *utilDate) SubDateTime(firstTime time.Time, secondTime time.Time) (result time.Duration) {
	result = time.Duration(0)
	if &firstTime != nil && &secondTime != nil {
		result = secondTime.Sub(firstTime)
	}
	return
}

func (_self *utilDate) DifferDays(firstTime time.Time, secondTime time.Time) int64 {
	result := _self.SubDateTime(firstTime, secondTime).Hours()
	return int64(math.Abs(result) / 24)
}

func (_self *utilDate) DifferHour(firstTime time.Time, secondTime time.Time) int64 {
	result := _self.SubDateTime(firstTime, secondTime).Hours()
	//return int64(result) 两个时间的先后顺序不一样，可能出现负数
	return int64(math.Abs(result))
}

func (_self *utilDate) DifferMin(firstTime time.Time, secondTime time.Time) int64 {
	result := _self.SubDateTime(firstTime, secondTime).Minutes()
	return int64(math.Abs(result))
}

func (_self *utilDate) DifferSec(firstTime time.Time, secondTime time.Time) int64 {
	result := _self.SubDateTime(firstTime, secondTime).Seconds()
	return int64(math.Abs(result))
}

func (_self *utilDate) DifferMilsec(firstTime time.Time, secondTime time.Time) int64 {
	result := _self.SubDateTime(firstTime, secondTime).Milliseconds()
	return result
}

// 24小时前的时间
func (_self *utilDate) Before24h() time.Time {
	t, _ := time.ParseDuration("-24h")
	return time.Now().Add(t)
}

func (_self *utilDate) AddSecs(_time time.Time, secs int64) time.Time {
	t, _ := time.ParseDuration("1s")
	return _time.Add(t * time.Duration(secs))
}

/*
增加10分钟：utils.AddMins(time.Now(), 10)
减少5分钟：utils.AddMins(time.Now(), -5)
*/
func (_self *utilDate) AddMins(_time time.Time, mins int64) time.Time {
	t, _ := time.ParseDuration("1m")
	return _time.Add(t * time.Duration(mins))
}

func (_self *utilDate) AddHours(_time time.Time, hours int64) time.Time {
	t, _ := time.ParseDuration("1h")
	return _time.Add(t * time.Duration(hours))
}

func (_self *utilDate) AddDays(_time time.Time, days int) time.Time {
	return _time.AddDate(0, 0, days)
}

func (_self *utilDate) AddMonths(_time time.Time, months int) time.Time {
	return _time.AddDate(0, months, 0)
}

func (_self *utilDate) GetBeginTime(_time time.Time) time.Time {
	//2017-06-28 00:00:00 +0800 CST
	return _self.GetBeginTimeByLoc(_time, time.Local)
	//return GetBeginTimeByLoc(_time, time.UTC)

}

func (_self *utilDate) GetEndTime(_time time.Time) time.Time {
	//2017-06-28 23:59:59.999999999 +0800 CST
	return _self.GetEndTimeByLoc(_time, time.Local)
}

func (_self *utilDate) GetBeginTimeByLoc(_time time.Time, loc *time.Location) time.Time {
	year, month, day := _time.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, loc)

}

func (_self *utilDate) GetEndTimeByLoc(_time time.Time, loc *time.Location) time.Time {
	year, month, day := _time.Date()
	return time.Date(year, month, day, 23, 59, 59, 999999999, loc)
}

// 一行代码计算代码执行时间
// defer utils.TimeCost(time.Now())
func (_self *utilDate) TimeCost(start time.Time) {
	terminal := time.Since(start)
	fmt.Println("TimeCost：", terminal)
}
