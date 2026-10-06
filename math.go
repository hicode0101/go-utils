package utils

import "math"

// utilMath 数学计算工具。
//
// 提示：Go 1.21+ 内置了 min/max 函数（支持任意可比较有序值与多参数），
// 因此本工具不再提供 MaxInt/MinInt。
type utilMath struct{}

// Abs 返回浮点数绝对值。场景：误差比较 Abs(a-b) < 0.0001。
func (m *utilMath) Abs(num float64) float64 {
	return math.Abs(num)
}

// AbsInt 返回 int 绝对值。场景：坐标偏移量取绝对值。
func (m *utilMath) AbsInt(num int) int {
	if num < 0 {
		return -num
	}
	return num
}

// AbsInt64 返回 int64 绝对值。场景：余额变动差额计算。
func (m *utilMath) AbsInt64(num int64) int64 {
	if num < 0 {
		return -num
	}
	return num
}

// CeilToInt 向上取整并转 int：5.3 -> 6，-5.3 -> -5。
// 场景：计算需要的服务实例数 = 上限(CeilToInt(total/qps))。
func (m *utilMath) CeilToInt(num float64) int {
	return int(math.Ceil(num))
}

// FloorToInt 向下取整并转 int：5.9 -> 5，-5.9 -> -6。
// 场景：整箱数计算。
func (m *utilMath) FloorToInt(num float64) int {
	return int(math.Floor(num))
}

// RoundToInt 四舍五入取整（.5 远离零方向）：5.5 -> 6，-5.5 -> -6。
// 场景：报表数值取整。
func (m *utilMath) RoundToInt(num float64) int {
	return int(math.Round(num))
}

// RoundFloat 保留 prec 位小数（四舍五入）：RoundFloat(3.14159, 2) = 3.14。
// 场景：金额展示前归一化（精确金额运算请使用整数分或 decimal 库）。
func (m *utilMath) RoundFloat(num float64, prec int) float64 {
	p := math.Pow10(prec)
	return math.Round(num*p) / p
}

// PageCount 按总数与每页大小计算总页数（向上取整）。
// PageCount(101, 20) = 6；total 或 size 非正时返回 0（避免除零）。
// 场景：分页接口返回 pages 字段。
func (m *utilMath) PageCount(total, size int64) int64 {
	if total <= 0 || size <= 0 {
		return 0
	}
	return (total + size - 1) / size
}
