package utils

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// utilSys 系统与环境工具。
type utilSys struct{}

// GetEnv 取环境变量值，不存在返回空串。
// 场景：读取 APP_ENV、DB_HOST 等配置；需要默认值时配合 utils.Sys.If 使用。
func (s *utilSys) GetEnv(key string) string {
	return os.Getenv(key)
}

// SetEnv 设置环境变量。场景：初始化阶段注入子进程需要的变量。
func (s *utilSys) SetEnv(key, value string) error {
	return os.Setenv(key, value)
}

// UnsetEnv 删除环境变量。场景：清理临时注入的变量。
func (s *utilSys) UnsetEnv(key string) error {
	return os.Unsetenv(key)
}

// GetPid 返回当前进程 ID。场景：写 pid 文件配合 systemd/运维脚本。
func (s *utilSys) GetPid() int {
	return os.Getpid()
}

// Hostname 返回主机名，获取失败返回空串。
// 场景：日志/上报数据里标识来源机器。
func (s *utilSys) Hostname() string {
	name, err := os.Hostname()
	if err != nil {
		return ""
	}
	return name
}

// KillPid 结束指定进程。注意：Windows 下只能结束当前进程；
// 目标进程须为本进程或同权限可操作进程。
// 场景：父进程守护中清理残留子进程。
func (s *utilSys) KillPid(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Kill()
}

// IsWindows 当前系统是否为 Windows。场景：按平台选择路径分隔符、命令。
func (s *utilSys) IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsLinux 当前系统是否为 Linux。场景：仅 Linux 启用的 cgroup 采集逻辑。
func (s *utilSys) IsLinux() bool {
	return runtime.GOOS == "linux"
}

// IsMac 当前系统是否为 macOS。场景：开发机特有逻辑。
func (s *utilSys) IsMac() bool {
	return runtime.GOOS == "darwin"
}

// RunCmd 执行外部命令并返回合并的标准输出与标准错误。
// RunCmd("git", "log", "-1")。场景：调用 ffmpeg/git 等外部工具。
// 注意：不要把用户输入直接拼进命令，避免命令注入。
func (s *utilSys) RunCmd(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).CombinedOutput()
	return string(out), err
}

// RunCmdTimeout 带（整体）超时的命令执行，超时后返回错误与已捕获输出。
// 场景：调用不可信/偶发卡死的外部程序时兜底。
func (s *utilSys) RunCmdTimeout(timeout time.Duration, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	return string(out), err
}

/* ---------------- 泛型辅助函数（Go 不支持泛型方法，故定义为包级函数） ---------------- */

// If 泛型三元表达式。Go 没有三元运算符，用它替代臃肿的 if-else：
//
//	max := utils.If(x > y, x, y)
//	name := utils.If(nick != "", nick, "匿名用户")
//
// 注意：trueVal/falseVal 两个表达式都会先求值（与 C 三元一致的限制在于
// 副作用，传函数调用时若副作用敏感请改写为普通 if）。
func If[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}

// Swap 交换两个同类型值：
//
//	a, b := utils.Swap(1, 2)          // a=2, b=1
//	x, y := utils.Swap("Li", "Chen")  // x="Chen", y="Li"
func Swap[T any](a, b T) (T, T) {
	return b, a
}
