package utils

import "reflect"

// utilObj 对象与通用值工具（hutool ObjectUtil 的 Go 精简版）。
//
// Go 无继承与统一 Object 概念，仅保留跨类型有意义的部分；
// 需要泛型三元/交换时使用包级函数 utils.If / utils.Swap / utils.DefaultIfNil。
type utilObj struct{}

// IsNil 判断值是否为"真 nil"（含接口包裹 nil 指针的情况）。
// 注意：v == nil 对"接口变量装着 nil 指针"会返回 false，本方法用 reflect 兜底。
// 场景：any 类型参数的空值校验。
func (o *utilObj) IsNil(v interface{}) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func, reflect.Interface, reflect.UnsafePointer:
		return rv.IsNil()
	}
	return false
}

// IsEmpty 判断值是否为空：nil、空串、空切片、空 map、空通道、零值数值/布尔。
// （对应 hutool ObjectUtil.isEmpty 的 Go 化改造，数值零值视为空）
// 场景：合并多个来源的可选字段后判断是否有效。
func (o *utilObj) IsEmpty(v interface{}) bool {
	if o.IsNil(v) {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array, reflect.Chan:
		return rv.Len() == 0
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.Ptr, reflect.Interface:
		return rv.IsNil()
	}
	return false
}

// IsNotEmpty 非 IsEmpty。
func (o *utilObj) IsNotEmpty(v interface{}) bool {
	return !o.IsEmpty(v)
}

// Equal 深度比较两个值是否相等（reflect.DeepEqual 封装）。
// 场景：测试断言、配置变更检测。
func (o *utilObj) Equal(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

// DefaultIfNil 泛型空指针兜底（hutool ObjectUtil.defaultIfNull 的指针版）：
// ptr 为 nil 时返回 def，否则返回 *ptr。
//
//	cfg := utils.DefaultIfNil(conf, DefaultConf())
func DefaultIfNil[T any](ptr *T, def T) T {
	if ptr == nil {
		return def
	}
	return *ptr
}

// Coalesce 返回第一个非零值的参数（全为零值时返回零值）。
// name := utils.Coalesce(nick, userName, "匿名用户")
// 场景：多来源取第一个有效值。注意与 DefaultIfNil 的区别：本方法按零值判断，支持非指针。
func Coalesce[T comparable](vals ...T) T {
	var zero T
	for _, v := range vals {
		if v != zero {
			return v
		}
	}
	return zero
}
