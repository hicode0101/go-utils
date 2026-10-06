package utils

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
