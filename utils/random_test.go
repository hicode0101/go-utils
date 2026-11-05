package utils

import (
	"strings"
	"testing"
	"unicode"
)

/* ---------------- 迁移自 utils_test.go 的 TestIdAndRandom（Random 部分，原样保留） ---------------- */

func TestRandomLegacy(t *testing.T) {
	if got := Random.RandomInt(5, 10); got < 5 || got >= 10 {
		t.Errorf("RandomInt 越界: %d", got)
	}
	if got := RandomEle([]string{"x"}); got != "x" {
		t.Errorf("RandomEle = %q", got)
	}
	if got := Random.RandomNumbers(6); len(got) != 6 {
		t.Errorf("RandomNumbers 长度错误: %q", got)
	}
	wr := NewWeightRandom([]float64{0, 1}, []string{"a", "b"})
	for i := 0; i < 20; i++ {
		if got := wr.Next(); got != "b" { // a 权重为 0，永远不应选中
			t.Fatalf("WeightRandom 选中了 0 权重元素: %q", got)
		}
	}
}

/* ---------------- RandomInt / RandomInt64 / RandomFloat ---------------- */

func TestRandomInt(t *testing.T) {
	for i := 0; i < 200; i++ {
		if got := Random.RandomInt(5, 10); got < 5 || got >= 10 {
			t.Fatalf("RandomInt(5,10) 越界: %d", got)
		}
	}
	// min>=max 返回 min
	if got := Random.RandomInt(5, 5); got != 5 {
		t.Errorf("RandomInt(5,5) = %d, want 5", got)
	}
	if got := Random.RandomInt(7, 3); got != 7 {
		t.Errorf("RandomInt(7,3) = %d, want 7", got)
	}
}

func TestRandomInt64(t *testing.T) {
	const min, max = int64(-5), int64(1) << 62
	for i := 0; i < 200; i++ {
		if got := Random.RandomInt64(min, max); got < min || got >= max {
			t.Fatalf("RandomInt64 越界: %d", got)
		}
	}
	if got := Random.RandomInt64(9, 9); got != 9 {
		t.Errorf("RandomInt64(9,9) = %d, want 9", got)
	}
	if got := Random.RandomInt64(9, 2); got != 9 {
		t.Errorf("RandomInt64(9,2) = %d, want 9", got)
	}
}

func TestRandomFloat(t *testing.T) {
	for i := 0; i < 200; i++ {
		if got := Random.RandomFloat(2.5, 5.5); got < 2.5 || got >= 5.5 {
			t.Fatalf("RandomFloat(2.5,5.5) 越界: %v", got)
		}
	}
	if got := Random.RandomFloat(3.3, 3.3); got != 3.3 {
		t.Errorf("RandomFloat(3.3,3.3) = %v, want 3.3", got)
	}
	if got := Random.RandomFloat(4.4, 1.1); got != 4.4 {
		t.Errorf("RandomFloat(4.4,1.1) = %v, want 4.4", got)
	}
}

/* ---------------- RandomBool / RandomBytes ---------------- */

func TestRandomBool(t *testing.T) {
	seenTrue, seenFalse := false, false
	for i := 0; i < 200; i++ {
		if Random.RandomBool() {
			seenTrue = true
		} else {
			seenFalse = true
		}
	}
	if !seenTrue || !seenFalse {
		t.Errorf("RandomBool 200 次采样应覆盖 true 与 false, true=%v false=%v", seenTrue, seenFalse)
	}
}

func TestRandomBytes(t *testing.T) {
	if got := Random.RandomBytes(16); len(got) != 16 {
		t.Errorf("RandomBytes(16) 长度 = %d, want 16", len(got))
	}
	if got := Random.RandomBytes(0); len(got) != 0 {
		t.Errorf("RandomBytes(0) 长度 = %d, want 0", len(got))
	}
	if got := Random.RandomBytes(-3); len(got) != 0 {
		t.Errorf("RandomBytes(-3) 长度 = %d, want 0", len(got))
	}
}

/* ---------------- RandomEle / RandomEles / RandomEleSet ---------------- */

// randomDemoStruct 自定义结构体，验证空切片返回可比较的零值
type randomDemoStruct struct {
	N int
}

func TestRandomEle(t *testing.T) {
	if got := RandomEle([]string{"x"}); got != "x" {
		t.Errorf("RandomEle = %q, want \"x\"", got)
	}
	if got := RandomEle([]int{7, 7, 7}); got != 7 {
		t.Errorf("RandomEle = %d, want 7", got)
	}
	// 空切片 / nil 返回零值
	if got := RandomEle([]randomDemoStruct{}); got != (randomDemoStruct{}) {
		t.Errorf("RandomEle 空切片应返回零值, got %+v", got)
	}
	if got := RandomEle[randomDemoStruct](nil); got != (randomDemoStruct{}) {
		t.Errorf("RandomEle nil 应返回零值, got %+v", got)
	}
}

func TestRandomEles(t *testing.T) {
	src := []int{1, 2, 3}
	got := RandomEles(src, 50)
	if len(got) != 50 {
		t.Fatalf("RandomEles 长度 = %d, want 50", len(got))
	}
	distinct := map[int]struct{}{}
	for _, v := range got {
		if v != 1 && v != 2 && v != 3 {
			t.Fatalf("RandomEles 产出集合外元素: %d", v)
		}
		distinct[v] = struct{}{}
	}
	if len(distinct) > len(src) {
		t.Fatalf("RandomEles 产出源切片外的值: %v", got)
	}
	// 有放回抽样允许重复：3 个元素抽 50 次必有重复（鸽笼原理）
	if len(distinct) >= len(got) {
		t.Errorf("RandomEles 3 元素有放回抽 50 次应出现重复, 实际 %d 个不同值", len(distinct))
	}
	// 边界：count<=0 或空源返回空切片
	if got := RandomEles(src, 0); len(got) != 0 {
		t.Errorf("RandomEles count=0 应返回空切片, got %v", got)
	}
	if got := RandomEles(src, -1); len(got) != 0 {
		t.Errorf("RandomEles count<0 应返回空切片, got %v", got)
	}
	if got := RandomEles([]int{}, 3); len(got) != 0 {
		t.Errorf("RandomEles 空源应返回空切片, got %v", got)
	}
}

func TestRandomEleSet(t *testing.T) {
	src := []int{1, 1, 2, 3, 3}
	got := RandomEleSet(src, 2)
	if len(got) != 2 {
		t.Fatalf("RandomEleSet 长度 = %d, want 2", len(got))
	}
	seen := map[int]struct{}{}
	for _, v := range got {
		if _, dup := seen[v]; dup {
			t.Fatalf("RandomEleSet 不应含重复元素: %v", got)
		}
		seen[v] = struct{}{}
		if v != 1 && v != 2 && v != 3 {
			t.Fatalf("RandomEleSet 产出集合外元素: %d", v)
		}
	}
	// count 超过去重后总数：返回全部
	if all := RandomEleSet(src, 100); len(all) != 3 {
		t.Errorf("count>len 应返回全部去重元素, got %v", all)
	}
	// count<=0 返回空切片
	if empty := RandomEleSet(src, 0); len(empty) != 0 {
		t.Errorf("RandomEleSet count=0 应返回空切片, got %v", empty)
	}
}

/* ---------------- RandomString / RandomNumbers / RandomStringWithoutStr ---------------- */

func TestRandomString(t *testing.T) {
	got := Random.RandomString(32)
	if len(got) != 32 {
		t.Fatalf("RandomString(32) 长度 = %d, want 32", len(got))
	}
	// 委托 String.RandomString：字符集为字母 + 数字
	for _, c := range got {
		if !unicode.IsDigit(c) && !unicode.IsLetter(c) {
			t.Fatalf("RandomString 产出非字母数字字符 %q: %q", c, got)
		}
	}
	if got := Random.RandomString(0); got != "" {
		t.Errorf("RandomString(0) 应返回空串, got %q", got)
	}
}

func TestRandomNumbers(t *testing.T) {
	got := Random.RandomNumbers(12)
	if len(got) != 12 {
		t.Fatalf("RandomNumbers(12) 长度 = %d, want 12", len(got))
	}
	for _, c := range got {
		if c < '0' || c > '9' {
			t.Fatalf("RandomNumbers 应为纯数字: %q", got)
		}
	}
	if got := Random.RandomNumbers(0); got != "" {
		t.Errorf("RandomNumbers(0) 应返回空串, got %q", got)
	}
}

func TestRandomStringWithoutStr(t *testing.T) {
	got := Random.RandomStringWithoutStr(64, "oO0lL1")
	if len(got) != 64 {
		t.Fatalf("RandomStringWithoutStr 长度 = %d, want 64", len(got))
	}
	for _, c := range got {
		if strings.ContainsRune("oO0lL1", c) {
			t.Fatalf("RandomStringWithoutStr 仍含排除字符 %q: %q", c, got)
		}
	}
	// 全部字符被剔除返回空串
	if got := Random.RandomStringWithoutStr(10, randomBaseCharNumber); got != "" {
		t.Errorf("全剔除应返回空串, got %q", got)
	}
	// length<=0 返回空串
	if got := Random.RandomStringWithoutStr(0, ""); got != "" {
		t.Errorf("RandomStringWithoutStr(0) 应返回空串, got %q", got)
	}
}

/* ---------------- RandomChar / RandomChinese ---------------- */

func TestRandomChar(t *testing.T) {
	const base = "abZ9中"
	for i := 0; i < 100; i++ {
		if r := Random.RandomChar(base); !strings.ContainsRune(base, r) {
			t.Fatalf("RandomChar 返回 base 之外的字符: %q", r)
		}
	}
	// 空 base 返回 0
	if got := Random.RandomChar(""); got != 0 {
		t.Errorf("RandomChar 空 base 应返回 0, got %d", got)
	}
}

func TestRandomChinese(t *testing.T) {
	for i := 0; i < 100; i++ {
		got := Random.RandomChinese()
		runes := []rune(got)
		if len(runes) != 1 {
			t.Fatalf("RandomChinese 应返回单个汉字, got %q", got)
		}
		if runes[0] < 0x4E00 || runes[0] > 0x9FA5 {
			t.Fatalf("RandomChinese 不在 CJK 区间: %q (U+%04X)", got, runes[0])
		}
	}
}

/* ---------------- WeightRandom ---------------- */

func TestRandomWeightRandom(t *testing.T) {
	// 正常权重：多次采样只返回集合内元素
	wr := NewWeightRandom([]float64{1, 2, 3, 4}, []string{"A", "B", "C", "D"})
	for i := 0; i < 300; i++ {
		if got := wr.Next(); got != "A" && got != "B" && got != "C" && got != "D" {
			t.Fatalf("WeightRandom 产出集合外元素: %q", got)
		}
	}
	// 0 权重永不被选中
	wr2 := NewWeightRandom([]float64{0, 0, 5}, []string{"a", "b", "c"})
	for i := 0; i < 200; i++ {
		if got := wr2.Next(); got != "c" {
			t.Fatalf("WeightRandom 选中了 0 权重元素: %q", got)
		}
	}
	// 全零权重退化为等概率：只断言能返回元素且属于集合
	wr3 := NewWeightRandom([]float64{0, 0, 0}, []string{"x", "y", "z"})
	for i := 0; i < 50; i++ {
		if got := wr3.Next(); got != "x" && got != "y" && got != "z" {
			t.Fatalf("WeightRandom 全零权重应返回集合内元素, got %q", got)
		}
	}
	// 空构造返回零值
	if got := NewWeightRandom[string](nil, nil).Next(); got != "" {
		t.Errorf("WeightRandom 空构造应返回零值, got %q", got)
	}
	// 长度不一致时以较短者为准
	wr4 := NewWeightRandom([]float64{1, 2, 3}, []string{"only"})
	if got := wr4.Next(); got != "only" {
		t.Errorf("WeightRandom 长度不一致应截断为较短者, got %q", got)
	}
}
