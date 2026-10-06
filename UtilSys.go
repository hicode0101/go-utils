package utils

import (
	"fmt"
	"os"
)

type utilSys struct {
}

// go不支持三元表达式，可以使用自定义的函数实现
// 例如：max := utils.If(x > y, x, y).(int)
func (_self *utilSys) If(condition bool, trueVal, falseVal interface{}) interface{} {

	if condition {
		return trueVal
	}
	return falseVal
}

/*
交换int数据：a, b := utils.Swap(2, 9)
交换字符串数据：A, B := utils.Swap("Li", "Chen")
*/
func (_self *utilSys) Swap(x, y interface{}) (interface{}, interface{}) {
	return y, x
}

// 设置环境变量
func (_self *utilSys) SetEnv(key, value string) error {

	return os.Setenv(key, value)
}

// 取环境变量的值
func (_self *utilSys) GetEnv(key string) string {

	return os.Getenv(key)
}

// 取进程ID
func (_self *utilSys) GetPid() int {
	return os.Getpid()
}

func (_self *utilSys) KillByPid(pid int) {
	p, _ := os.FindProcess(pid)
	fmt.Println("KillByPid", p)
	p.Kill()
}
