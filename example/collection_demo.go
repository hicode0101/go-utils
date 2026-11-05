package main

import (
	"fmt"

	"github.com/hicode0101/go-utils/utils"
)

// collectionDemo 包级泛型函数：集合操作 / 随机选元素 / If / Swap / DefaultIfNil / Coalesce。
// （Go 方法不能自带类型参数，因此集合能力以包级函数提供。）
func collectionDemo() {
	nums := []int{3, 1, 2, 3}

	banner("集合操作（泛型包级函数）")
	fmt.Println("IndexOf(nums, 2) =", utils.IndexOf(nums, 2), "| Contains(nums, 3) =", utils.Contains(nums, 3))
	fmt.Println("Unique(nums) =", utils.Unique(nums), "（保序去重）")
	fmt.Println("Filter(nums, 偶数) =", utils.Filter(nums, func(v int) bool { return v%2 == 0 }))
	fmt.Println("Map(nums, *10) =", utils.Map(nums, func(v int) int { return v * 10 }))
	fmt.Println("Reverse(nums) =", utils.Reverse(nums), "（原地反转并返回自身）")
	fmt.Println("RemoveAt(nums, 0) =", utils.RemoveAt(nums, 0))
	fmt.Println("Chunk(nums, 2) =", utils.Chunk(nums, 2), "| Sub(nums, 1, 3) =", utils.Sub(nums, 1, 3))
	fmt.Println("Union([1,2], [2,3]) =", utils.Union([]int{1, 2}, []int{2, 3}),
		"| Intersect =", utils.Intersect([]int{1, 2, 3}, []int{2, 3, 4}),
		"| Difference =", utils.Difference([]int{1, 2, 3}, []int{2, 3, 4}))
	fmt.Println("Empty([]int{}) =", utils.Empty([]int{}), "| NotEmpty(nums) =", utils.NotEmpty(nums))
	fmt.Println("Get(nums, 10, -1) =", utils.Get(nums, 10, -1), "（越界返回默认值）")
	fmt.Println("Join(nums, \"-\") =", utils.Join(nums, "-"))
	fmt.Println("ZipToMap([a b], [1 2]) =", utils.ZipToMap([]string{"a", "b"}, []int{1, 2}))
	fmt.Println("Shuffle([1,2,3,4,5]) =", utils.Shuffle([]int{1, 2, 3, 4, 5}))

	banner("随机选元素")
	fmt.Println("RandomEle(nums) =", utils.RandomEle(nums))
	fmt.Println("RandomEles(nums, 3) =", utils.RandomEles(nums, 3), "（可重复）")
	fmt.Println("RandomEleSet(nums, 2) =", utils.RandomEleSet(nums, 2), "（不重复）")

	banner("泛型辅助")
	fmt.Println(`If(3 > 2, "对", "错") =`, utils.If(3 > 2, "对", "错"))
	a, b := utils.Swap(1, 2)
	fmt.Println("Swap(1, 2) =", a, b)
	var ptr *int
	fmt.Println("DefaultIfNil(nil, 100) =", utils.DefaultIfNil(ptr, 100))
	fmt.Println(`Coalesce("", "x", "y") =`, utils.Coalesce("", "x", "y"), "（第一个非零值）")
}
