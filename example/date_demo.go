package main

import (
	"fmt"
	"time"

	"github.com/hicode0101/go-utils/utils"
)

// datetimeDemo 日期时间：Date。
func datetimeDemo() {
	banner("Date 日期时间工具")
	now := time.Now()

	// 格式化（默认布局 "2006-01-02 15:04:05"）
	fmt.Println("Format(now) =", utils.Date.Format(now))
	fmt.Println("FormatDate(now) =", utils.Date.FormatDate(now), "| FormatCompact(now) =", utils.Date.FormatCompact(now))
	fmt.Println("CurrentTime() =", utils.Date.CurrentTime(), "| CurrentDate() =", utils.Date.CurrentDate())

	// 解析：自动识别常见布局 / 指定布局
	if t, err := utils.Date.Parse("2024-06-05 12:00:00"); err == nil {
		fmt.Println(`Parse("2024-06-05 12:00:00") =`, utils.Date.Format(t))
	}
	fmt.Println(`MustParse("2024-06-05 08:30:00") =`, utils.Date.Format(utils.Date.MustParse("2024-06-05 08:30:00")))
	if d, err := utils.Date.ParseDate("2024-06-05"); err == nil {
		fmt.Println(`ParseDate("2024-06-05") =`, utils.Date.FormatDate(d))
	}
	if t, err := utils.Date.ParseWith("2024/06/05", "2006/01/02"); err == nil {
		fmt.Println(`ParseWith("2024/06/05", "2006/01/02") =`, utils.Date.FormatDate(t))
	}

	// Unix 时间戳互转
	fmt.Println("FromUnix(1700000000) =", utils.Date.Format(utils.Date.FromUnix(1700000000)),
		"| ToUnix(now) =", utils.Date.ToUnix(now))

	// 加减与边界
	fmt.Println("AddDays(now, 7) =", utils.Date.FormatDate(utils.Date.AddDays(now, 7)),
		"| AddMonths(now, 1) =", utils.Date.FormatDate(utils.Date.AddMonths(now, 1)))
	fmt.Println("StartOfDay(now) =", utils.Date.Format(utils.Date.StartOfDay(now)),
		"| EndOfDay(now) =", utils.Date.Format(utils.Date.EndOfDay(now)))
	fmt.Println("StartOfMonth(now) =", utils.Date.FormatDate(utils.Date.StartOfMonth(now)),
		"| EndOfMonth(now) =", utils.Date.FormatDate(utils.Date.EndOfMonth(now)))
	fmt.Println("BeginOfWeek(now, false) =", utils.Date.FormatDate(utils.Date.BeginOfWeek(now, false)),
		"（false=周日为一周起点）")

	// 比较与差值
	birthday := time.Date(1990, 3, 7, 0, 0, 0, 0, time.Local)
	fmt.Println("IsSameDay(now, now) =", utils.Date.IsSameDay(now, now))
	fmt.Println("DiffDays(birthday, now) =", utils.Date.DiffDays(birthday, now), "天")
	fmt.Println("Age(birthday, now) =", utils.Date.Age(birthday, now), "岁")
	fmt.Println("IsLeapYear(2024) =", utils.Date.IsLeapYear(2024))
	fmt.Println("IsBeforeNow(birthday) =", utils.Date.IsBeforeNow(birthday))
}
