package utils

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

// 集合（切片）操作：以包级泛型函数提供，直接通过 utils.IndexOf(...) 等方式调用。
//
// 说明：Go 语言不允许方法声明自己的类型参数，因此泛型集合函数无法挂在
// Collection 对象上，统一改为包级函数，这是语言层面的约束而非风格选择。

// IndexOf 返回 value 在切片中首次出现的下标，未找到返回 -1。
// 场景：IndexOf([]int{1, 2, 3}, 2) = 1；替代旧版 IntArrayFind/Int64ArrayFind/StrArrayFind。
func IndexOf[T comparable](slice []T, value T) int {
	for i, v := range slice {
		if v == value {
			return i
		}
	}
	return -1
}

// Contains 判断切片是否包含 value。
// 场景：白名单校验 Contains([]string{"GET","POST"}, method)；替代旧版三个 XxxContain。
func Contains[T comparable](slice []T, value T) bool {
	return IndexOf(slice, value) >= 0
}

// Filter 按谓词 keep 过滤元素，返回新切片（原切片不变）。
// 场景：Filter(nums, func(v int) bool { return v%2 == 0 }) 取偶数。
func Filter[T any](slice []T, keep func(item T) bool) []T {
	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if keep(v) {
			result = append(result, v)
		}
	}
	return result
}

// Map 对每个元素执行 fn 映射，返回新切片（长度与原切片一致）。
// 场景：Map(ids, func(id int64) string { return String.Int64ToStr(id) }) 转字符串。
func Map[T any, R any](slice []T, fn func(item T) R) []R {
	result := make([]R, 0, len(slice))
	for _, v := range slice {
		result = append(result, fn(v))
	}
	return result
}

// Unique 去重并保持首次出现的顺序。
// 场景：Unique([]string{"a","b","a"}) = ["a","b"]。
func Unique[T comparable](slice []T) []T {
	seen := make(map[T]struct{}, len(slice))
	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

// Reverse 原地反转切片并返回自身（便于链式书写）。
// 场景：最新数据倒序展示。
func Reverse[T any](slice []T) []T {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
	return slice
}

// RemoveAt 删除指定下标的元素并返回新切片；下标越界时原样返回。
// 场景：按序号删除列表项。
func RemoveAt[T any](slice []T, index int) []T {
	if index < 0 || index >= len(slice) {
		return slice
	}
	result := make([]T, 0, len(slice)-1)
	result = append(result, slice[:index]...)
	result = append(result, slice[index+1:]...)
	return result
}

// Chunk 把切片按每块 size 个元素分块，最后一块可能不足 size；size<=0 或空切片返回 nil。
// 场景：大批量数据分批写库（每批 500 条）。
func Chunk[T any](slice []T, size int) [][]T {
	if size <= 0 || len(slice) == 0 {
		return nil
	}
	chunks := make([][]T, 0, (len(slice)+size-1)/size)
	for start := 0; start < len(slice); start += size {
		end := start + size
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[start:end])
	}
	return chunks
}

// Union 求多个切片的并集（去重、保持首次出现顺序）。
// 场景：合并多个来源的标签列表。
func Union[T comparable](slices ...[]T) []T {
	seen := make(map[T]struct{})
	result := make([]T, 0)
	for _, slice := range slices {
		for _, v := range slice {
			if _, ok := seen[v]; !ok {
				seen[v] = struct{}{}
				result = append(result, v)
			}
		}
	}
	return result
}

// Intersect 求切片 a 与 b 的交集（按 a 中出现顺序去重）。
// 场景：求两个角色共同拥有的权限。
func Intersect[T comparable](a, b []T) []T {
	set := make(map[T]struct{}, len(b))
	for _, v := range b {
		set[v] = struct{}{}
	}
	seen := make(map[T]struct{}, len(a))
	result := make([]T, 0)
	for _, v := range a {
		if _, ok := set[v]; ok {
			if _, dup := seen[v]; !dup {
				seen[v] = struct{}{}
				result = append(result, v)
			}
		}
	}
	return result
}

// Difference 求差集：返回在 a 中但不在 b 中的元素（保序去重）。
// 场景：计算待删除的数据 = 库里现有的 - 本次提交的。
func Difference[T comparable](a, b []T) []T {
	set := make(map[T]struct{}, len(b))
	for _, v := range b {
		set[v] = struct{}{}
	}
	seen := make(map[T]struct{}, len(a))
	result := make([]T, 0)
	for _, v := range a {
		if _, ok := set[v]; ok {
			continue
		}
		if _, dup := seen[v]; !dup {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

/* ---------------- 追加集合操作（对应 hutool CollUtil 常用方法） ---------------- */

// Empty 切片是否为空（nil 或长度 0）。场景：接口返回集合字段判空。
func Empty[T any](slice []T) bool {
	return len(slice) == 0
}

// NotEmpty 切片是否非空。
func NotEmpty[T any](slice []T) bool {
	return len(slice) > 0
}

// Get 安全取值：下标越界时返回默认值 def（对应 hutool CollUtil.get(coll, i, default)）。
// 场景：按序号取值但不想写 length 判断。
func Get[T any](slice []T, index int, def T) T {
	if index < 0 || index >= len(slice) {
		return def
	}
	return slice[index]
}

// Join 拼接切片元素为字符串（对应 hutool CollUtil.join），元素用 fmt.Sprint 转换。
// Join([]int{1,2,3}, "-") = "1-2-3"。场景：拼日志、拼 IN 条件（字符串切片用 String.Join 更快）。
func Join[T any](slice []T, sep string) string {
	var b strings.Builder
	for i, v := range slice {
		if i > 0 {
			b.WriteString(sep)
		}
		b.WriteString(fmt.Sprint(v))
	}
	return b.String()
}

// Shuffle 原地随机打乱切片并返回自身（对应 hutool CollUtil.shuffle）。
// 场景：抽奖、随机排序展示。
func Shuffle[T any](slice []T) []T {
	rand.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
	return slice
}

// Sub 切片截取 [start, end)，自动收敛越界下标（对应 hutool CollUtil.sub）。
// 场景：手动分页取片段。注意与 Chunk 的区别：Sub 按下标区间，Chunk 按每块大小。
func Sub[T any](slice []T, start, end int) []T {
	n := len(slice)
	if start < 0 {
		start = 0
	}
	if end > n {
		end = n
	}
	if start >= end {
		return nil
	}
	return slice[start:end]
}

// ZipToMap 把 keys 与 values 两个切片合成 map（对应 hutool CollUtil.zip），
// 以较短的一方为准，键冲突时后面的覆盖前面的。
// 场景：两列数据配对成映射。
func ZipToMap[K comparable, V any](keys []K, values []V) map[K]V {
	n := len(keys)
	if len(values) < n {
		n = len(values)
	}
	result := make(map[K]V, n)
	for i := 0; i < n; i++ {
		result[keys[i]] = values[i]
	}
	return result
}
