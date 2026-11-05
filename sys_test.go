package utils

import (
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"
)

/* ---------------- 原有基础断言（自 utils_test.go 原样搬入） ---------------- */

func TestSys(t *testing.T) {
	if Sys.GetPid() <= 0 {
		t.Error("GetPid 应为正数")
	}
	if Sys.Hostname() == "" {
		t.Error("Hostname 不应为空")
	}
	if Sys.SetEnv("GO_UTILS_TEST", "1") == nil && Sys.GetEnv("GO_UTILS_TEST") != "1" {
		t.Error("环境变量设置后应可读取")
	}
	_ = Sys.UnsetEnv("GO_UTILS_TEST")
}

/* ---------------- 泛型 If / Swap（自 utils_test.go 的 TestIfAndSwap 更名搬入并扩展） ---------------- */

func TestSysIfSwap(t *testing.T) {
	if got := If(1 > 2, "yes", "no"); got != "no" {
		t.Errorf("If = %q, want \"no\"", got)
	}
	if got := If(2 > 1, "yes", "no"); got != "yes" {
		t.Errorf("If = %q, want \"yes\"", got)
	}
	if got := If(true, 1, -1); got != 1 {
		t.Errorf("If(int) = %d, want 1", got)
	}
	if got := If(false, 1, -1); got != -1 {
		t.Errorf("If(int) = %d, want -1", got)
	}
	a, b := Swap(1, 2)
	if a != 2 || b != 1 {
		t.Errorf("Swap(int) = %d, %d; want 2, 1", a, b)
	}
	x, y := Swap("Li", "Chen")
	if x != "Chen" || y != "Li" {
		t.Errorf("Swap(string) = %q, %q; want \"Chen\", \"Li\"", x, y)
	}
}

/* ---------------- 环境变量 ---------------- */

func TestSysEnv(t *testing.T) {
	const key = "GO_UTILS_TEST_ENV_ROUNDTRIP"
	if err := Sys.SetEnv(key, "v1"); err != nil {
		t.Fatalf("SetEnv 失败: %v", err)
	}
	if got := Sys.GetEnv(key); got != "v1" {
		t.Errorf("GetEnv = %q, want \"v1\"", got)
	}
	// 覆盖写入
	if err := Sys.SetEnv(key, "v2"); err != nil {
		t.Fatalf("SetEnv 覆盖失败: %v", err)
	}
	if got := Sys.GetEnv(key); got != "v2" {
		t.Errorf("GetEnv 覆盖后 = %q, want \"v2\"", got)
	}
	// 删除后为空
	if err := Sys.UnsetEnv(key); err != nil {
		t.Fatalf("UnsetEnv 失败: %v", err)
	}
	if got := Sys.GetEnv(key); got != "" {
		t.Errorf("UnsetEnv 后 GetEnv = %q, want \"\"", got)
	}
	// 不存在的 key 返回空串
	if got := Sys.GetEnv("GO_UTILS_NO_SUCH_ENV_XYZ_12345"); got != "" {
		t.Errorf("GetEnv 不存在的 key = %q, want \"\"", got)
	}
}

/* ---------------- 进程与主机信息 ---------------- */

func TestSysProcessInfo(t *testing.T) {
	if Sys.GetPid() <= 0 {
		t.Errorf("GetPid = %d, 应为正数", Sys.GetPid())
	}
	if Sys.Hostname() == "" {
		t.Error("Hostname 不应为空")
	}
}

/* ---------------- 平台判断 ---------------- */

func TestSysPlatformCheck(t *testing.T) {
	// 三者恰好有一个为 true
	count := 0
	if Sys.IsWindows() {
		count++
	}
	if Sys.IsLinux() {
		count++
	}
	if Sys.IsMac() {
		count++
	}
	if count != 1 {
		t.Errorf("IsWindows/IsLinux/IsMac 应恰好一个为 true, 实际 %d 个", count)
	}
	// 与 runtime.GOOS 一致
	if Sys.IsWindows() != (runtime.GOOS == "windows") {
		t.Error("IsWindows 应与 runtime.GOOS 一致")
	}
	if Sys.IsLinux() != (runtime.GOOS == "linux") {
		t.Error("IsLinux 应与 runtime.GOOS 一致")
	}
	if Sys.IsMac() != (runtime.GOOS == "darwin") {
		t.Error("IsMac 应与 runtime.GOOS 一致")
	}
}

/* ---------------- 外部命令 ---------------- */

func TestSysRunCmd(t *testing.T) {
	var out string
	var err error
	if runtime.GOOS == "windows" {
		out, err = Sys.RunCmd("cmd", "/c", "echo", "hi")
	} else {
		out, err = Sys.RunCmd("echo", "hi")
	}
	if err != nil {
		t.Fatalf("RunCmd 失败: %v", err)
	}
	if !strings.Contains(out, "hi") {
		t.Errorf("RunCmd 输出应包含 hi, got %q", out)
	}
	// 错误路径：不存在的命令
	if _, err := Sys.RunCmd("go_utils_no_such_cmd_xyz"); err == nil {
		t.Error("执行不存在的命令应返回错误")
	}
}

func TestSysRunCmdTimeout(t *testing.T) {
	// 正常路径：5 秒超时足够执行完 echo
	var out string
	var err error
	if runtime.GOOS == "windows" {
		out, err = Sys.RunCmdTimeout(5*time.Second, "cmd", "/c", "echo", "ok")
	} else {
		out, err = Sys.RunCmdTimeout(5*time.Second, "echo", "ok")
	}
	if err != nil {
		t.Fatalf("RunCmdTimeout 正常路径失败: %v", err)
	}
	if !strings.Contains(out, "ok") {
		t.Errorf("RunCmdTimeout 输出应包含 ok, got %q", out)
	}
	// 超时路径：1ns 超时必然失败
	if runtime.GOOS == "windows" {
		_, err = Sys.RunCmdTimeout(1*time.Nanosecond, "cmd", "/c", "echo", "never")
	} else {
		_, err = Sys.RunCmdTimeout(1*time.Nanosecond, "sleep", "1")
	}
	if err == nil {
		t.Error("超时后应返回错误")
	}
}

/* ---------------- 结束进程（错误路径，不杀真进程） ---------------- */

func TestSysKillPid(t *testing.T) {
	if runtime.GOOS == "windows" {
		// Windows 下 os.FindProcess(-1) 直接返回错误，不会触及任何真实进程
		if err := Sys.KillPid(-1); err == nil {
			t.Error("KillPid(-1) 应返回错误")
		}
		return
	}
	// Unix：启动后立即结束并完全回收的子进程，其 pid 已不存在，Kill 应返回错误
	cmd := exec.Command("sleep", "30")
	if err := cmd.Start(); err != nil {
		t.Skipf("无法启动测试子进程: %v", err)
	}
	pid := cmd.Process.Pid
	_ = cmd.Process.Kill()
	_ = cmd.Wait() // 完全回收，避免僵尸进程
	if err := Sys.KillPid(pid); err == nil {
		t.Errorf("KillPid(%d)（已退出的进程）应返回错误", pid)
	}
}
