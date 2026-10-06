package utils

import (
	"math"
)

type utilMatch struct {
}

func (_self *utilMatch) AbsInt(num float64) int {
	//result := math.Abs(float64(num))
	result := math.Abs(num)
	return int(result)
}

func (_self *utilMatch) AbsInt64(num float64) int64 {
	result := math.Abs(num)
	return int64(result)
}

func (_self *utilMatch) CeilInt(num float64) int {
	result := math.Ceil(num)
	return int(result)
}

func (_self *utilMatch) CeilInt64(num float64) int64 {
	//CeilInt64(5.9) = 6
	//CeilInt64(5.3) = 6
	//CeilInt64(5) = 5
	result := math.Ceil(num)
	return int64(result)
}

func (_self *utilMatch) Float64ToInt64(num float64) int64 {
	return int64(num)
}

func (_self *utilMatch) Float64TryToInt64(num interface{}) int64 {
	return int64(num.(float64))
}

// 返回最大值
func (_self *utilMatch) MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// 返回最小值
func (_self *utilMatch) MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (_self *utilMatch) Pages(total, psize int64) int64 {

	pages := float64(total) / float64(psize)
	result := math.Ceil(pages)
	return int64(result)
}
