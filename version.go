package utils

import (
	"strconv"
	"strings"
)

// utilVersion 版本号对比工具（hutool VersionUtil + VersionComparator 移植）。
//
// 版本号按"主序列-预发布+构建"三段解析（如 1.2.3-beta.1+build.5），
// 逐 token 比较：数字按数值比、字符串按字典序比、混合时按字符串比；
// 尾部全 0 的段可省略（1.0 与 1.0.0 相等）；预发布版本小于正式版本。
type utilVersion struct{}

// versionTokens 版本号 token 序列：元素为 int（数字段）或 string（字母段）
type versionTokens []interface{}

// versionToken 版本号结构化结果
type versionToken struct {
	sequence versionTokens // 主序列（首个 - 之前）
	pre      versionTokens // 预发布（- 与 + 之间）
	build    versionTokens // 构建（+ 之后）
}

// Compare 比较两个版本号，返回 -1/0/1。空串视为最小版本。
// Version.Compare("1.2.0", "1.1.9") = 1，Version.Compare("1.0", "1.0.0") = 0。
// 场景：依赖版本判断、灰度发布范围匹配。
func (v *utilVersion) Compare(v1, v2 string) int {
	if v1 == v2 {
		return 0
	}
	if v1 == "" {
		return -1
	}
	if v2 == "" {
		return 1
	}
	return compareVersionToken(parseVersionToken(v1), parseVersionToken(v2))
}

// GreaterThan 当前版本是否大于待比较版本。
func (v *utilVersion) GreaterThan(currentVersion, compareVersion string) bool {
	return v.Compare(currentVersion, compareVersion) > 0
}

// GreaterThanOrEqual 当前版本是否大于等于待比较版本。
func (v *utilVersion) GreaterThanOrEqual(currentVersion, compareVersion string) bool {
	return v.Compare(currentVersion, compareVersion) >= 0
}

// LessThan 当前版本是否小于待比较版本。
func (v *utilVersion) LessThan(currentVersion, compareVersion string) bool {
	return v.Compare(currentVersion, compareVersion) < 0
}

// LessThanOrEqual 当前版本是否小于等于待比较版本。
func (v *utilVersion) LessThanOrEqual(currentVersion, compareVersion string) bool {
	return v.Compare(currentVersion, compareVersion) <= 0
}

// AnyMatch 当前版本是否满足任意一个版本表达式（分号分隔，等价于 MatchEl）。
// Version.AnyMatch("1.0.2", "<1.0.1", "1.0.2") = true。
// 场景：兼容性白名单配置。
func (v *utilVersion) AnyMatch(currentVersion string, compareVersions ...string) bool {
	return v.MatchEl(currentVersion, strings.Join(compareVersions, ";"), ";")
}

// MatchEl 当前版本是否满足版本表达式（hutool matchEl）。
// 表达式支持以 delimiter（默认 ";"）分隔的多个条件，任一命中即返回 true：
//   - ">1.0" "<1.1" ">=2.0" "<=2.0"（含全角 ≥ ≤）
//   - "1.0.0-1.1.1" 闭区间；"-" 左空表示 <= 右值，右空表示 >= 左值
//   - "1.0.2" 精确匹配
//
// MatchEl("1.0.2", ">=1.0,<1.1", ",") = true。场景：配置化版本门槛。
func (v *utilVersion) MatchEl(currentVersion, versionEl, delimiter string) bool {
	if delimiter == "" || delimiter == "-" {
		return false // hutool 对非法分隔符抛异常，这里直接返回不匹配
	}
	currentVersion = strings.TrimSpace(currentVersion)
	for _, el := range String.SplitTrim(versionEl, delimiter) {
		// 带比较运算符的条件
		if op, ver, ok := splitVersionOp(el); ok {
			switch op {
			case ">=", "≥":
				if v.Compare(currentVersion, ver) >= 0 {
					return true
				}
			case "<=", "≤":
				if v.Compare(currentVersion, ver) <= 0 {
					return true
				}
			case "<":
				if v.Compare(currentVersion, ver) < 0 {
					return true
				}
			case ">":
				if v.Compare(currentVersion, ver) > 0 {
					return true
				}
			}
			continue
		}
		// 闭区间条件 "-"
		if idx := strings.Index(el, "-"); idx >= 0 {
			left := strings.TrimSpace(el[:idx])
			right := strings.TrimSpace(el[idx+1:])
			leftMatch := left == "" || v.Compare(left, currentVersion) <= 0
			rightMatch := right == "" || v.Compare(right, currentVersion) >= 0
			if leftMatch && rightMatch {
				return true
			}
			continue
		}
		// 精确匹配
		if currentVersion == el {
			return true
		}
	}
	return false
}

// splitVersionOp 拆出比较运算符与版本号（内部辅助），无运算符时 ok=false
func splitVersionOp(el string) (op, ver string, ok bool) {
	for _, candidate := range []string{">=", "<=", "≥", "≤", ">", "<"} {
		if strings.HasPrefix(el, candidate) {
			return candidate, strings.TrimSpace(strings.TrimPrefix(el, candidate)), true
		}
	}
	return "", "", false
}

// parseVersionToken 解析版本号为三段 token（内部辅助，对应 hutool Version.of）
func parseVersionToken(v string) versionToken {
	var result versionToken
	main, pre, build := v, "", ""
	if idx := strings.IndexByte(v, '+'); idx >= 0 {
		build = v[idx+1:]
		v = v[:idx]
	}
	if idx := strings.IndexByte(v, '-'); idx >= 0 {
		pre = v[idx+1:]
		main = v[:idx]
	}
	result.sequence = tokenizeVersion(main)
	result.pre = tokenizeVersion(pre)
	result.build = tokenizeVersion(build)
	return result
}

// tokenizeVersion 把一段版本串拆成数字/字母 token 序列（内部辅助）：
// "1.2.3-beta" -> [1, 2, 3] / ["beta"]；分隔符为 '.' 与 '-'
func tokenizeVersion(part string) versionTokens {
	if part == "" {
		return nil
	}
	var tokens versionTokens
	var b strings.Builder
	flush := func() {
		if b.Len() == 0 {
			return
		}
		s := b.String()
		b.Reset()
		if n, err := strconv.Atoi(s); err == nil {
			tokens = append(tokens, n)
		} else {
			tokens = append(tokens, s)
		}
	}
	for _, c := range part {
		switch {
		case c == '.' || c == '-':
			flush()
		case c >= '0' && c <= '9':
			// 数字与字母边界处切开（如 "3a" -> [3, "a"]）
			if b.Len() > 0 {
				last := rune(b.String()[len(b.String())-1])
				if !(last >= '0' && last <= '9') {
					flush()
				}
			}
			b.WriteRune(c)
		default:
			if b.Len() > 0 {
				last := rune(b.String()[len(b.String())-1])
				if last >= '0' && last <= '9' {
					flush()
				}
			}
			b.WriteRune(c)
		}
	}
	flush()
	return tokens
}

// compareVersionToken 三段依次比较：主序列、预发布（空 pre 大于非空 pre）、构建
func compareVersionToken(a, b versionToken) int {
	if c := compareVersionTokens(a.sequence, b.sequence); c != 0 {
		return c
	}
	if len(a.pre) == 0 && len(b.pre) != 0 {
		return 1
	}
	if len(a.pre) != 0 && len(b.pre) == 0 {
		return -1
	}
	if c := compareVersionTokens(a.pre, b.pre); c != 0 {
		return c
	}
	return compareVersionTokens(a.build, b.build)
}

// compareVersionTokens token 序列逐个比较（对应 hutool Version.compareTokens）：
// 同类型直接比较，混合类型转字符串比较；尾部全 0 的多余 token 不影响结果
func compareVersionTokens(ts1, ts2 versionTokens) int {
	n := len(ts1)
	if len(ts2) < n {
		n = len(ts2)
	}
	for i := 0; i < n; i++ {
		c := compareVersionValue(ts1[i], ts2[i])
		if c != 0 {
			return c
		}
	}
	// 较长一方的剩余 token：全为 0 则忽略，否则长的一方大
	rest := ts1
	if len(ts2) > len(ts1) {
		rest = ts2
	}
	for i := n; i < len(rest); i++ {
		if num, isInt := rest[i].(int); isInt && num == 0 {
			continue
		}
		return len(ts1) - len(ts2)
	}
	return 0
}

// compareVersionValue 单个 token 比较（内部辅助）
func compareVersionValue(o1, o2 interface{}) int {
	n1, i1 := o1.(int)
	n2, i2 := o2.(int)
	if i1 && i2 {
		return n1 - n2
	}
	s1, s2 := versionTokenString(o1), versionTokenString(o2)
	return strings.Compare(s1, s2)
}

// versionTokenString token 转字符串（内部辅助）
func versionTokenString(o interface{}) string {
	if n, ok := o.(int); ok {
		return strconv.Itoa(n)
	}
	return o.(string)
}
