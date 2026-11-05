package utils

import "strings"

// utilBool 布尔工具。
//
// 命名约定：ToBoolExt 为扩展真值解析（比 String.ToBool 认识更多"真"的写法），
// ToStringXxx 系列按固定词对输出，And/Or/Xor 支持不定长参数。
type utilBool struct{}

// boolTrueSet 扩展真值集合（对照 hutool BooleanUtil.TRUE_SET，匹配时忽略大小写与首尾空白）
var boolTrueSet = map[string]struct{}{
	"true": {}, "yes": {}, "y": {}, "t": {}, "ok": {}, "correct": {}, "success": {},
	"on": {}, "1": {}, "是": {}, "对": {}, "真": {}, "對": {}, "正确": {},
	"开": {}, "开启": {}, "√": {}, "☑": {},
}

// ToBoolExt 扩展真值解析（对照 hutool BooleanUtil.toBoolean）：去除首尾空白并忽略大小写后，
// 命中真值集合返回 true，其余情况（空串、空白串、以及 "no"、"0"、"off"、"否" 等假值写法）一律 false。
// 真值集合："true"、"yes"、"y"、"t"、"ok"、"correct"、"success"、"on"、"1"、
// "是"、"对"、"真"、"對"、"正确"、"开"、"开启"、"√"、"☑"。
// 场景：解析用户自定义配置、表单/问卷里的中文开关字段。
func (b *utilBool) ToBoolExt(str string) bool {
	str = strings.ToLower(strings.TrimSpace(str))
	if str == "" {
		return false
	}
	_, ok := boolTrueSet[str]
	return ok
}

// Negate 取相反值。Negate(true) = false。场景：开关反转。
func (b *utilBool) Negate(v bool) bool {
	return !v
}

// And 逻辑与：全部为 true 才返回 true（无参返回 false；hutool 对空数组抛异常，Go 便捷处理）。
// And(true, true, false) = false。场景：多项条件全部满足才放行。
func (b *utilBool) And(vals ...bool) bool {
	if len(vals) == 0 {
		return false
	}
	for _, v := range vals {
		if !v {
			return false
		}
	}
	return true
}

// Or 逻辑或：任一为 true 即返回 true（无参返回 false，hutool 对空数组抛异常，Go 便捷处理）。
// Or(false, false, true) = true。场景：多项条件命中其一即通过。
func (b *utilBool) Or(vals ...bool) bool {
	for _, v := range vals {
		if v {
			return true
		}
	}
	return false
}

// Xor 逻辑异或：vals 中 true 的个数为奇数时返回 true（无参返回 false）。
// Xor(true, false) = true，Xor(true, true, true) = true，Xor(true, true) = false。
// 注意：这是"奇偶校验"语义，并非"恰好一个为真"；互斥条件校验请自行统计个数。
func (b *utilBool) Xor(vals ...bool) bool {
	result := false
	for _, v := range vals {
		result = result != v
	}
	return result
}

// ToStringTrueFalse 布尔转字符串：true -> "true"，false -> "false"。
// 场景：JSON 之外的文本协议输出。
func (b *utilBool) ToStringTrueFalse(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

// ToStringOnOff 布尔转字符串：true -> "on"，false -> "off"。场景：配置文件开关项。
func (b *utilBool) ToStringOnOff(v bool) string {
	if v {
		return "on"
	}
	return "off"
}

// ToStringYesNo 布尔转字符串：true -> "yes"，false -> "no"。场景：问卷/报表输出。
func (b *utilBool) ToStringYesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}
