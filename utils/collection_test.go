package utils

import (
	"sort"
	"testing"
)

/* ---------------- 原有基础断言（自 utils_test.go 原样搬入） ---------------- */

func TestCollection(t *testing.T) {
	nums := []int{1, 2, 3, 2}
	if IndexOf(nums, 3) != 2 {
		t.Errorf("IndexOf = %d, want 2", IndexOf(nums, 3))
	}
	if IndexOf(nums, 9) != -1 {
		t.Error("IndexOf 未找到应返回 -1")
	}
	if !Contains(nums, 2) {
		t.Error("Contains 应为 true")
	}
	if got := Unique(nums); len(got) != 3 {
		t.Errorf("Unique 长度 = %d, want 3", len(got))
	}
	if got := Filter(nums, func(v int) bool { return v%2 == 0 }); len(got) != 2 {
		t.Errorf("Filter 偶数个数 = %d, want 2", len(got))
	}
	strs := Map(nums, func(v int) string { return String.IntToStr(v) })
	if len(strs) != 4 || strs[0] != "1" {
		t.Errorf("Map 结果错误: %v", strs)
	}
	if got := Chunk([]int{1, 2, 3, 4, 5}, 2); len(got) != 3 || len(got[2]) != 1 {
		t.Errorf("Chunk 分块错误: %v", got)
	}
	if got := Union([]int{1, 2}, []int{2, 3}); len(got) != 3 {
		t.Errorf("Union 长度 = %d, want 3", len(got))
	}
	if got := Intersect([]int{1, 2, 3}, []int{2, 3, 4}); len(got) != 2 {
		t.Errorf("Intersect 长度 = %d, want 2", len(got))
	}
	if got := Difference([]int{1, 2, 3}, []int{2}); len(got) != 2 {
		t.Errorf("Difference 长度 = %d, want 2", len(got))
	}
	if got := RemoveAt([]string{"a", "b", "c"}, 1); len(got) != 2 || got[1] != "c" {
		t.Errorf("RemoveAt 结果错误: %v", got)
	}
}

/* ---------------- 集合扩展（hutool CollUtil，原样搬入） ---------------- */

func TestCollectionHutoolExt(t *testing.T) {
	if !NotEmpty([]int{1}) || !Empty([]int{}) || !Empty[int](nil) || NotEmpty([]int{}) {
		t.Error("Empty/NotEmpty 判断错误")
	}
	if got := Get([]string{"a", "b"}, 5, "def"); got != "def" {
		t.Errorf("Get 越界应返回默认值, got %q", got)
	}
	if got := Get([]string{"a", "b"}, 1, "def"); got != "b" {
		t.Errorf("Get = %q", got)
	}
	if got := Join([]int{1, 2, 3}, "-"); got != "1-2-3" {
		t.Errorf("Join = %q", got)
	}
	nums := []int{1, 2, 3, 4, 5}
	Shuffle(nums)
	if len(nums) != 5 {
		t.Errorf("Shuffle 后长度应为 5, got %d", len(nums))
	}
	if got := Sub([]int{0, 1, 2, 3, 4}, 1, 3); len(got) != 2 || got[0] != 1 {
		t.Errorf("Sub = %v", got)
	}
	m := ZipToMap([]string{"a", "b"}, []int{1, 2})
	if m["a"] != 1 || m["b"] != 2 {
		t.Errorf("ZipToMap = %v", m)
	}
}

/* ---------------- IndexOf / Contains（int 与 string 两种类型参数） ---------------- */

func TestCollectionIndexOfContains(t *testing.T) {
	nums := []int{10, 20, 30}
	if IndexOf(nums, 10) != 0 {
		t.Errorf("IndexOf 首元素 = %d, want 0", IndexOf(nums, 10))
	}
	if IndexOf(nums, 20) != 1 {
		t.Errorf("IndexOf = %d, want 1", IndexOf(nums, 20))
	}
	if IndexOf(nums, 99) != -1 {
		t.Error("IndexOf 未找到应返回 -1")
	}
	if IndexOf([]int{}, 1) != -1 {
		t.Error("IndexOf 空切片应返回 -1")
	}
	if !Contains(nums, 30) || Contains(nums, 99) || Contains([]int{}, 1) {
		t.Error("Contains 判断错误")
	}
	strs := []string{"get", "post"}
	if IndexOf(strs, "post") != 1 || IndexOf(strs, "put") != -1 {
		t.Errorf("IndexOf(string) 错误: %d, %d", IndexOf(strs, "post"), IndexOf(strs, "put"))
	}
	if !Contains(strs, "get") || Contains(strs, "put") {
		t.Error("Contains(string) 判断错误")
	}
}

/* ---------------- Filter / Map ---------------- */

func TestCollectionFilterMap(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5, 6}
	evens := Filter(nums, func(v int) bool { return v%2 == 0 })
	if len(evens) != 3 || evens[0] != 2 || evens[2] != 6 {
		t.Errorf("Filter 偶数 = %v", evens)
	}
	if len(nums) != 6 {
		t.Error("Filter 不应修改原切片")
	}
	if got := Filter([]int{}, func(v int) bool { return true }); len(got) != 0 {
		t.Errorf("Filter 空切片应返回空, got %v", got)
	}
	if got := Filter(nums, func(v int) bool { return false }); len(got) != 0 {
		t.Errorf("Filter 全不命中应返回空, got %v", got)
	}
	strs := Map(nums, func(v int) string { return String.IntToStr(v) })
	if len(strs) != 6 || strs[0] != "1" || strs[5] != "6" {
		t.Errorf("Map(int->string) = %v", strs)
	}
	lens := Map([]string{"a", "bb", "ccc"}, func(s string) int { return len(s) })
	if len(lens) != 3 || lens[2] != 3 {
		t.Errorf("Map(string->int) = %v", lens)
	}
	if got := Map([]int{}, func(v int) string { return "x" }); len(got) != 0 {
		t.Errorf("Map 空切片应返回空, got %v", got)
	}
}

/* ---------------- Unique / Reverse ---------------- */

func TestCollectionUniqueReverse(t *testing.T) {
	strs := Unique([]string{"a", "b", "a", "c", "b"})
	if len(strs) != 3 || strs[0] != "a" || strs[1] != "b" || strs[2] != "c" {
		t.Errorf("Unique 应保序去重, got %v", strs)
	}
	if got := Unique([]int{}); len(got) != 0 {
		t.Errorf("Unique 空切片应返回空, got %v", got)
	}
	nums := []int{1, 2, 3}
	if got := Reverse(nums); len(got) != 3 || got[0] != 3 || got[1] != 2 || got[2] != 1 {
		t.Errorf("Reverse = %v", got)
	}
	if got := Reverse([]int{}); len(got) != 0 {
		t.Errorf("Reverse 空切片应安全, got %v", got)
	}
	single := []string{"x"}
	if got := Reverse(single); len(got) != 1 || got[0] != "x" {
		t.Errorf("Reverse 单元素 = %v", got)
	}
}

/* ---------------- RemoveAt ---------------- */

func TestCollectionRemoveAt(t *testing.T) {
	strs := []string{"a", "b", "c"}
	if got := RemoveAt(strs, 1); len(got) != 2 || got[0] != "a" || got[1] != "c" {
		t.Errorf("RemoveAt = %v", got)
	}
	if len(strs) != 3 {
		t.Error("RemoveAt 不应修改原切片")
	}
	if got := RemoveAt(strs, 99); len(got) != 3 {
		t.Errorf("RemoveAt 越界应原样返回, got %v", got)
	}
	if got := RemoveAt(strs, -1); len(got) != 3 {
		t.Errorf("RemoveAt 负下标应原样返回, got %v", got)
	}
	if got := RemoveAt([]int{1}, 0); len(got) != 0 {
		t.Errorf("RemoveAt 删到空 = %v", got)
	}
}

/* ---------------- Chunk ---------------- */

func TestCollectionChunk(t *testing.T) {
	chunks := Chunk([]int{1, 2, 3, 4, 5}, 2)
	if len(chunks) != 3 || len(chunks[0]) != 2 || len(chunks[1]) != 2 || len(chunks[2]) != 1 || chunks[2][0] != 5 {
		t.Errorf("Chunk 分块错误: %v", chunks)
	}
	// 恰好整除
	if got := Chunk([]int{1, 2, 3, 4}, 2); len(got) != 2 || len(got[1]) != 2 || got[1][1] != 4 {
		t.Errorf("Chunk 整除分块错误: %v", got)
	}
	// size<=0 返回 nil
	if Chunk([]int{1, 2}, 0) != nil {
		t.Error("Chunk size=0 应返回 nil")
	}
	if Chunk([]int{1, 2}, -1) != nil {
		t.Error("Chunk size<0 应返回 nil")
	}
	// 空切片返回 nil
	if Chunk([]int{}, 2) != nil {
		t.Error("Chunk 空切片应返回 nil")
	}
	words := Chunk([]string{"a", "b", "c"}, 2)
	if len(words) != 2 || words[0][0] != "a" || words[1][0] != "c" {
		t.Errorf("Chunk(string) = %v", words)
	}
}

/* ---------------- Union / Intersect / Difference ---------------- */

func TestCollectionUnion(t *testing.T) {
	got := Union([]int{1, 2}, []int{2, 3}, []int{}, []int{3, 4})
	if len(got) != 4 || got[0] != 1 || got[1] != 2 || got[2] != 3 || got[3] != 4 {
		t.Errorf("Union 多切片应去重保序, got %v", got)
	}
	if got := Union[int](); len(got) != 0 {
		t.Errorf("Union 无切片应返回空, got %v", got)
	}
	strs := Union([]string{"a"}, []string{"b", "a"})
	if len(strs) != 2 || strs[0] != "a" || strs[1] != "b" {
		t.Errorf("Union(string) = %v", strs)
	}
}

func TestCollectionIntersect(t *testing.T) {
	got := Intersect([]int{1, 2, 2, 3}, []int{2, 3, 4})
	if len(got) != 2 || got[0] != 2 || got[1] != 3 {
		t.Errorf("Intersect 应按 a 顺序去重, got %v", got)
	}
	if got := Intersect([]int{1, 2}, []int{3, 4}); len(got) != 0 {
		t.Errorf("Intersect 无交集应返回空, got %v", got)
	}
	strs := Intersect([]string{"x", "y"}, []string{"y", "x"})
	if len(strs) != 2 || strs[0] != "x" {
		t.Errorf("Intersect(string) 应按 a 顺序, got %v", strs)
	}
}

func TestCollectionDifference(t *testing.T) {
	got := Difference([]int{1, 2, 3, 2}, []int{2})
	if len(got) != 2 || got[0] != 1 || got[1] != 3 {
		t.Errorf("Difference 应保序去重, got %v", got)
	}
	if got := Difference([]int{1, 2}, []int{}); len(got) != 2 {
		t.Errorf("Difference 空 b 应返回全部 a, got %v", got)
	}
	if got := Difference([]int{1}, []int{1, 2}); len(got) != 0 {
		t.Errorf("Difference 全被排除应返回空, got %v", got)
	}
	strs := Difference([]string{"a", "b"}, []string{"b"})
	if len(strs) != 1 || strs[0] != "a" {
		t.Errorf("Difference(string) = %v", strs)
	}
}

/* ---------------- Empty / NotEmpty / Get ---------------- */

func TestCollectionEmptyGet(t *testing.T) {
	if !Empty[int](nil) || !Empty([]int{}) || Empty([]int{1}) {
		t.Error("Empty 判断错误")
	}
	if !NotEmpty([]int{1}) || NotEmpty([]int{}) || NotEmpty[int](nil) {
		t.Error("NotEmpty 判断错误")
	}
	if got := Get([]string{"a", "b"}, 1, "def"); got != "b" {
		t.Errorf("Get 合法下标 = %q", got)
	}
	if got := Get([]string{"a", "b"}, 5, "def"); got != "def" {
		t.Errorf("Get 越界应返回默认值, got %q", got)
	}
	if got := Get([]string{"a", "b"}, -1, "def"); got != "def" {
		t.Errorf("Get 负下标应返回默认值, got %q", got)
	}
	if got := Get([]int{}, 0, 42); got != 42 {
		t.Errorf("Get 空切片应返回默认值, got %d", got)
	}
}

/* ---------------- Join ---------------- */

func TestCollectionJoin(t *testing.T) {
	if got := Join([]int{1, 2, 3}, "-"); got != "1-2-3" {
		t.Errorf("Join(int) = %q", got)
	}
	if got := Join([]string{"a", "b"}, ","); got != "a,b" {
		t.Errorf("Join(string) = %q", got)
	}
	if got := Join([]int{}, "-"); got != "" {
		t.Errorf("Join 空切片 = %q, want \"\"", got)
	}
	if got := Join([]int{7}, "-"); got != "7" {
		t.Errorf("Join 单元素不应带分隔符, got %q", got)
	}
	if got := Join([]float64{1.5, 2}, "|"); got != "1.5|2" {
		t.Errorf("Join(float64) = %q", got)
	}
}

/* ---------------- Shuffle ---------------- */

func TestCollectionShuffle(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	got := Shuffle(nums)
	if len(got) != 5 {
		t.Errorf("Shuffle 后长度 = %d, want 5", len(got))
	}
	sorted := append([]int{}, got...)
	sort.Ints(sorted)
	for i, want := range []int{1, 2, 3, 4, 5} {
		if sorted[i] != want {
			t.Fatalf("Shuffle 后元素应保持不变: %v", sorted)
		}
	}
	if got := Shuffle([]string{}); len(got) != 0 {
		t.Errorf("Shuffle 空切片应安全, got %v", got)
	}
}

/* ---------------- Sub ---------------- */

func TestCollectionSub(t *testing.T) {
	nums := []int{0, 1, 2, 3, 4}
	if got := Sub(nums, 1, 3); len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Errorf("Sub = %v", got)
	}
	if got := Sub(nums, 3, 100); len(got) != 2 || got[0] != 3 {
		t.Errorf("Sub end 越界应收敛, got %v", got)
	}
	if got := Sub(nums, -10, 2); len(got) != 2 || got[0] != 0 {
		t.Errorf("Sub start<0 应收敛为 0, got %v", got)
	}
	if Sub(nums, 3, 3) != nil {
		t.Error("Sub start>=end 应返回 nil")
	}
	if Sub(nums, 10, 20) != nil {
		t.Error("Sub 完全越界应返回 nil")
	}
	strs := Sub([]string{"a", "b", "c"}, 0, 2)
	if len(strs) != 2 || strs[1] != "b" {
		t.Errorf("Sub(string) = %v", strs)
	}
}

/* ---------------- ZipToMap ---------------- */

func TestCollectionZipToMap(t *testing.T) {
	m := ZipToMap([]string{"a", "b"}, []int{1, 2})
	if len(m) != 2 || m["a"] != 1 || m["b"] != 2 {
		t.Errorf("ZipToMap = %v", m)
	}
	// 以短的一方为准（values 短）
	m2 := ZipToMap([]string{"k1", "k2", "k3"}, []int{1})
	if len(m2) != 1 || m2["k1"] != 1 {
		t.Errorf("ZipToMap values 较短应以短为准, got %v", m2)
	}
	// 以短的一方为准（keys 短）
	m3 := ZipToMap([]string{"k1"}, []int{1, 2, 3})
	if len(m3) != 1 || m3["k1"] != 1 {
		t.Errorf("ZipToMap keys 较短应以短为准, got %v", m3)
	}
	// 键冲突时后面的覆盖前面的
	m4 := ZipToMap([]string{"a", "a"}, []int{1, 2})
	if len(m4) != 1 || m4["a"] != 2 {
		t.Errorf("ZipToMap 键冲突应后者覆盖, got %v", m4)
	}
	// 空输入
	if got := ZipToMap([]string{}, []int{}); len(got) != 0 {
		t.Errorf("ZipToMap 空输入 = %v", got)
	}
}
