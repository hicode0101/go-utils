package utils

import "testing"

/* ---------------- 原有测试（自 utils_test.go 原样迁入） ---------------- */

func TestObj(t *testing.T) {
	if !Obj.IsNil(nil) || Obj.IsNil(1) {
		t.Error("IsNil 判断错误")
	}
	var np *int
	if !Obj.IsNil(np) {
		t.Error("包裹 nil 指针的 any 应判定为 nil")
	}
	if !Obj.IsEmpty("") || !Obj.IsEmpty(0) || !Obj.IsEmpty([]int{}) || Obj.IsEmpty("x") {
		t.Error("IsEmpty 判断错误")
	}
	if !Obj.Equal([]int{1, 2}, []int{1, 2}) {
		t.Error("Equal 深度比较错误")
	}
	n := 5
	if DefaultIfNil(&n, 0) != 5 {
		t.Error("DefaultIfNil 非 nil 应返回原值")
	}
	if DefaultIfNil[int](nil, 9) != 9 {
		t.Error("DefaultIfNil nil 应回退默认值")
	}
	if got := Coalesce("", "", "x"); got != "x" {
		t.Errorf("Coalesce = %q", got)
	}
}

/* ---------------- 补充覆盖 ---------------- */

func TestObjIsNil(t *testing.T) {
	if !Obj.IsNil(nil) {
		t.Error("IsNil(nil) 应为 true")
	}
	var np *int
	if !Obj.IsNil(np) {
		t.Error("typed-nil 指针应判定为 nil")
	}
	var nm map[string]int
	if !Obj.IsNil(nm) {
		t.Error("nil map 应判定为 nil")
	}
	var ns []int
	if !Obj.IsNil(ns) {
		t.Error("nil slice 应判定为 nil")
	}
	x := 5
	if Obj.IsNil(&x) {
		t.Error("非 nil 指针不应判定为 nil")
	}
	if Obj.IsNil(1) || Obj.IsNil("") {
		t.Error("非指针零值不应判定为 nil")
	}
}

func TestObjIsEmpty(t *testing.T) {
	cases := []struct {
		name string
		v    interface{}
		want bool
	}{
		{"nil", nil, true},
		{"空串", "", true},
		{"非空串", "x", false},
		{"空切片", []int{}, true},
		{"非空切片", []int{1}, false},
		{"空 map", map[string]int{}, true},
		{"非空 map", map[string]int{"a": 1}, false},
		{"int 0", 0, true},
		{"int 1", 1, false},
		{"uint 0", uint(0), true},
		{"uint 1", uint(1), false},
		{"float 0", 0.0, true},
		{"float 0.5", 0.5, false},
		{"bool false", false, true},
		{"bool true", true, false},
	}
	for _, c := range cases {
		if got := Obj.IsEmpty(c.v); got != c.want {
			t.Errorf("IsEmpty(%s) = %v, want %v", c.name, got, c.want)
		}
	}
	type demo struct{ A int }
	// 结构体非零值不为空（结构体类型不在判定范围，统一返回 false）
	if Obj.IsEmpty(demo{A: 1}) {
		t.Error("IsEmpty(非空结构体) 应为 false")
	}
}

func TestObjIsNotEmpty(t *testing.T) {
	if !Obj.IsNotEmpty("x") {
		t.Error("IsNotEmpty(\"x\") 应为 true")
	}
	if !Obj.IsNotEmpty(1) {
		t.Error("IsNotEmpty(1) 应为 true")
	}
	if Obj.IsNotEmpty("") || Obj.IsNotEmpty(nil) || Obj.IsNotEmpty(0) {
		t.Error("IsNotEmpty 空值应为 false")
	}
}

func TestObjEqual(t *testing.T) {
	if !Obj.Equal([]int{1, 2}, []int{1, 2}) {
		t.Error("Equal 切片深比较应相等")
	}
	if Obj.Equal([]int{1, 2}, []int{1, 3}) {
		t.Error("Equal 切片元素不同应不相等")
	}
	if !Obj.Equal(map[string]int{"a": 1}, map[string]int{"a": 1}) {
		t.Error("Equal map 深比较应相等")
	}
	if Obj.Equal(map[string]int{"a": 1}, map[string]int{"a": 2}) {
		t.Error("Equal map 值不同应不相等")
	}
	if !Obj.Equal("a", "a") {
		t.Error("Equal 相同字符串应相等")
	}
	if Obj.Equal("a", "b") {
		t.Error("Equal 不同字符串应不相等")
	}
	if Obj.Equal(nil, 0) {
		t.Error("Equal(nil, 0) 应不相等")
	}
}

func TestObjDefaultIfNil(t *testing.T) {
	n := 5
	if got := DefaultIfNil(&n, 0); got != 5 {
		t.Errorf("DefaultIfNil 非 nil = %d, want 5", got)
	}
	if got := DefaultIfNil[int](nil, 9); got != 9 {
		t.Errorf("DefaultIfNil nil = %d, want 9", got)
	}
	s := "real"
	if got := DefaultIfNil(&s, "def"); got != "real" {
		t.Errorf("DefaultIfNil 字符串非 nil = %q, want \"real\"", got)
	}
	if got := DefaultIfNil[string](nil, "def"); got != "def" {
		t.Errorf("DefaultIfNil 字符串 nil = %q, want \"def\"", got)
	}
}

func TestObjCoalesce(t *testing.T) {
	if got := Coalesce("", "", "x"); got != "x" {
		t.Errorf("Coalesce 多值 = %q, want \"x\"", got)
	}
	if got := Coalesce(0, 5, 7); got != 5 {
		t.Errorf("Coalesce(0,5,7) = %d, want 5", got)
	}
	if got := Coalesce("first", "second"); got != "first" {
		t.Errorf("Coalesce 首个非零 = %q, want \"first\"", got)
	}
	if got := Coalesce("", ""); got != "" {
		t.Errorf("Coalesce 全零值 = %q, want \"\"", got)
	}
	if got := Coalesce[int](); got != 0 {
		t.Errorf("Coalesce 无参数 = %d, want 0", got)
	}
}
