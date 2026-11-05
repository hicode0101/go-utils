package main

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/hicode0101/go-utils/utils"
)

// cronDemo 定时任务：CronPattern 表达式 / Cron 调度门面 / Scheduler 独立调度器。
func cronDemo() {
	banner("CronPattern 表达式解析与匹配")
	// 6 域：秒 分 时 日 月 周；支持 * ? , - / L 与 | 多表达式
	p, err := utils.NewCronPattern("0 0 12 * * ?")
	if err != nil {
		fmt.Println("解析失败:", err)
		return
	}
	noon := time.Date(2024, 6, 5, 12, 0, 0, 0, time.Local)
	fmt.Printf("Pattern() = %s | Match(12:00:00) = %v | Match(13:00:00) = %v\n",
		p.Pattern(), p.Match(noon, true), p.Match(noon.Add(time.Hour), true))
	next := p.NextAfter(noon.Add(-time.Hour), true)
	fmt.Println("NextAfter(11:00) = 下一次触发 →", next.Format(time.DateTime))

	// | 分隔多个表达式，任一匹配即可
	multi, _ := utils.NewCronPattern("0 0 10 * * ?|0 0 8 * * ?")
	fmt.Println("多表达式 10点|8点，8:30 出发的下一次触发 →",
		multi.NextAfter(time.Date(2024, 6, 5, 8, 30, 0, 0, time.Local), true).Format(time.DateTime))

	banner("Cron 调度门面（全局单例）")
	utils.Cron.SetMatchSecond(true) // 启用秒级匹配（默认分钟级）
	var fired int32
	id, err := utils.Cron.Schedule("0/1 * * * * ?", func() { // 每秒触发
		atomic.AddInt32(&fired, 1)
	})
	if err != nil {
		fmt.Println("注册失败:", err)
		return
	}
	if err := utils.Cron.Start(); err != nil {
		fmt.Println("启动失败:", err)
		return
	}
	fmt.Println("已注册任务数 =", utils.Cron.Size(), "| IsStarted =", utils.Cron.IsStarted())
	time.Sleep(3 * time.Second) // 运行 3 秒（调度按秒边界对齐触发）
	utils.Cron.Stop()
	fmt.Println("Stop() 后累计触发 =", atomic.LoadInt32(&fired), "次 | IsStarted =", utils.Cron.IsStarted())
	utils.Cron.Remove(id) // 移除任务
	// 其它方法：ScheduleWithId(自定义ID) / UpdatePattern / Restart / GetScheduler

	banner("Scheduler 独立调度器（不启动也能管理任务表）")
	sch := utils.NewScheduler(true) // true = 启用秒级匹配
	err2 := sch.ScheduleWithId("clean", "0 0 3 * * ?", func() { /* 每天 3 点清理 */ })
	if err2 == nil {
		if cp, ok := sch.GetPattern("clean"); ok {
			fmt.Printf("ScheduleWithId(\"clean\", ...) -> GetPattern = %s\n", cp.Pattern())
		}
	}
	_ = sch.UpdatePattern("clean", "0 0 4 * * ?") // 动态改期
	fmt.Println("UpdatePattern 后 Size =", sch.Size())
	handle := sch.AddListener(utils.TaskListener{OnStart: func(id string) { fmt.Println("[listener] 开始:", id) }})
	fmt.Println("AddListener 句柄 =", handle)
	fmt.Println("Deschedule(\"clean\") =", sch.Deschedule("clean"), "| 清空后 Size =", sch.Size())
}
