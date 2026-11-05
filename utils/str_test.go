package utils

import (
	"strings"
	"testing"
)

// str_test.go 字符串工具（str.go，utilStr 经单例 String 暴露）的全覆盖测试。
//
// 测试范围：
//   - 类型转换：ToInt/ToIntE/ToInt64/ToInt64E/ToFloat64/ToFloat64E/ToBool/ToBoolE/
//     IntToStr/Int64ToStr/Float64ToStr
//   - 二进制与位运算：BinToInt/IntToBin/HasBits/BinStrHasBits
//   - 判断与检查：IsEmpty/IsNotEmpty/IsBlank/IsNotBlank/Trim/TrimStart/TrimEnd/
//     TrimToNull/Len/RuneCount
//   - 查找与处理：Contains/StartWith/EndWith/ContainsIgnoreCase/Count/Index/LastIndex/
//     Replace/ReplaceAll/Split/Join/Sub/SubBefore/SubAfter/SubBetween/ToLower/ToUpper/
//     Reverse/Repeat/Truncate/PadPre/PadAfter/Center
//   - 拼接与格式化：Concat/ConcatAny/FormatBytes/Format/FormatWithMap
//   - 随机串：RandomString
//   - 正则过滤：FilterByRegex/FilterSpecialChar/FilterUrlPrefix
//   - 格式校验：IsNumber/IsIP/IsIPv4/IsIPv6/IsDomain/IsDnsDomain/IsEmail/IsMobile/
//     IsIDCard/IsURL
//   - 默认值：NullToDefault/EmptyToDefault/BlankToDefault
//   - 前后缀与包裹：RemovePrefix/RemoveSuffix/AddPrefixIfNot/AddSuffixIfNot/
//     Wrap/UnWrap/IsWrap
//   - 命名风格：UpperFirst/LowerFirst/SwapCase/ToUnderlineCase/ToKebabCase/
//     ToSymbolCase/ToCamelCase/ToPascalCase
//   - 拆分：SplitTrim/SplitByLength
//   - 相似度与摘要：Similar/SimilarPercent/LongestCommonSubstring/Brief/Hide
//   - 全角半角：ToSBC/ToDBC
//   - 公共前后缀：CommonPrefix/CommonSuffix
//
// 其中 TestStrConversions/TestStrProcess/TestStrValidation/TestStrHutoolExt
// 四个函数自 utils_test.go 原样迁入，其余函数补齐上述方法的边界与失败路径用例。
// 所有断言值均经临时程序实测核对（含 hutool 专有名词保留大写的真实行为）。

/* ---------------- 以下四个测试函数自 utils_test.go 原样迁入 ---------------- */

func TestStrConversions(t *testing.T) {
	if got := String.ToInt("42"); got != 42 {
		t.Errorf("ToInt(\"42\") = %d, want 42", got)
	}
	if got := String.ToInt("abc"); got != 0 {
		t.Errorf("ToInt(\"abc\") = %d, want 0", got)
	}
	if _, err := String.ToIntE("abc"); err == nil {
		t.Error("ToIntE(\"abc\") 应返回错误")
	}
	if got := String.IntToStr(42); got != "42" {
		t.Errorf("IntToStr(42) = %q, want \"42\"", got)
	}
	if got := String.Float64ToStr(3.14159, 2); got != "3.14" {
		t.Errorf("Float64ToStr(3.14159, 2) = %q, want \"3.14\"", got)
	}
	if got := String.ToBool("true"); !got {
		t.Error("ToBool(\"true\") 应为 true")
	}
}

func TestStrProcess(t *testing.T) {
	// Sub 按字符截取，中文安全（对应 hutool sub）
	if got := String.Sub("你好世界", 1, 3); got != "好世" {
		t.Errorf("Sub(中文,1,3) = %q, want \"好世\"", got)
	}
	// 越界自动收敛
	if got := String.Sub("abc", 2, 100); got != "c" {
		t.Errorf("Sub 越界 = %q, want \"c\"", got)
	}
	// 负下标从末尾倒数
	if got := String.Sub("hello", -3, -1); got != "ll" {
		t.Errorf("Sub(-3,-1) = %q, want \"ll\"", got)
	}
	if got := String.Reverse("中国abc"); got != "cba国中" {
		t.Errorf("Reverse = %q, want \"cba国中\"", got)
	}
	if got := String.Truncate("你好世界欢迎", 5, "..."); got != "你好世界欢..." {
		t.Errorf("Truncate = %q, want \"你好世界欢...\"", got)
	}
	if got := String.PadPre("7", 3, '0'); got != "007" {
		t.Errorf("PadPre = %q, want \"007\"", got)
	}
	if got := String.ConcatAny("id=", 1001, " ok=", true); got != "id=1001 ok=true" {
		t.Errorf("ConcatAny = %q", got)
	}
	if got := String.FormatBytes(1536); got != "1.50 KB" {
		t.Errorf("FormatBytes(1536) = %q, want \"1.50 KB\"", got)
	}
	if got := String.RandomString(16); len(got) != 16 {
		t.Errorf("RandomString(16) 长度 = %d, want 16", len(got))
	}
}

func TestStrValidation(t *testing.T) {
	cases := []struct {
		name string
		got  bool
		want bool
	}{
		{"IsIPv4 合法", String.IsIPv4("192.168.1.1"), true},
		{"IsIPv4 非法", String.IsIPv4("999.1.1.1"), false},
		{"IsIPv6 合法", String.IsIPv6("fe80::1"), true},
		{"IsEmail 合法", String.IsEmail("a.b-c@example.com"), true},
		{"IsEmail 非法", String.IsEmail("a@@b.com"), false},
		{"IsMobile 合法", String.IsMobile("13812345678"), true},
		{"IsMobile 非法", String.IsMobile("12345678901"), false},
		{"IsNumber 合法", String.IsNumber("12345"), true},
		{"IsNumber 负数不匹配", String.IsNumber("-1"), false},
		{"IsURL 合法", String.IsURL("https://example.com/a?b=1"), true},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, c.got, c.want)
		}
	}
}

func TestStrHutoolExt(t *testing.T) {
	if got := String.SubBefore("abcba", "b", false); got != "a" {
		t.Errorf("SubBefore = %q, want \"a\"", got)
	}
	if got := String.SubAfter("abcba", "b", false); got != "cba" {
		t.Errorf("SubAfter = %q, want \"cba\"", got)
	}
	if got := String.SubAfter("abcba", "b", true); got != "a" {
		t.Errorf("SubAfter(isLast) = %q, want \"a\"", got)
	}
	if got := String.SubBefore("abc", "x", false); got != "abc" {
		t.Errorf("SubBefore 未找到应返回原串, got %q", got)
	}
	if got := String.SubBetween("wx[b]yz", "[", "]"); got != "b" {
		t.Errorf("SubBetween = %q, want \"b\"", got)
	}
	if got := String.ToUnderlineCase("PNLabel"); got != "PN_label" {
		t.Errorf("ToUnderlineCase(PNLabel) = %q, want \"PN_label\"", got)
	}
	if got := String.ToCamelCase("user_name"); got != "userName" {
		t.Errorf("ToCamelCase = %q, want \"userName\"", got)
	}
	if got := String.ToPascalCase("user_name"); got != "UserName" {
		t.Errorf("ToPascalCase = %q, want \"UserName\"", got)
	}
	if got := String.ToKebabCase("userName"); got != "user-name" {
		t.Errorf("ToKebabCase = %q, want \"user-name\"", got)
	}
	if got := String.UpperFirst("user"); got != "User" {
		t.Errorf("UpperFirst = %q", got)
	}
	if got := String.SwapCase("aBc"); got != "AbC" {
		t.Errorf("SwapCase = %q", got)
	}
	if got := String.Brief("你好世界欢迎", 5); got != "你好..." {
		t.Errorf("Brief = %q, want \"你好...\"", got)
	}
	if got := String.Hide("13812345678", 3, 7); got != "138****5678" {
		t.Errorf("Hide = %q, want \"138****5678\"", got)
	}
	if got := String.Format("{} 是 {} 岁", "张三", 18); got != "张三 是 18 岁" {
		t.Errorf("Format = %q", got)
	}
	if got := String.FormatWithMap("您好 {name}", map[string]string{"name": "张三"}); got != "您好 张三" {
		t.Errorf("FormatWithMap = %q", got)
	}
	if got := String.FormatWithMap("您好 {nope}", map[string]string{"name": "张三"}); got != "您好 {nope}" {
		t.Errorf("FormatWithMap 未知 key 应原样保留, got %q", got)
	}
	if got := String.Center("ab", 5, '*'); got != "*ab**" {
		t.Errorf("Center = %q, want \"*ab**\"", got)
	}
	if got := String.Wrap("id", "${", "}"); got != "${id}" {
		t.Errorf("Wrap = %q", got)
	}
	if got := String.UnWrap("${id}", "${", "}"); got != "id" {
		t.Errorf("UnWrap = %q", got)
	}
	if got := String.BlankToDefault("  ", "匿名"); got != "匿名" {
		t.Errorf("BlankToDefault = %q", got)
	}
	if got := String.RemovePrefix("https://a.com", "https://"); got != "a.com" {
		t.Errorf("RemovePrefix = %q", got)
	}
	if got := String.AddSuffixIfNot("logs", "/"); got != "logs/" {
		t.Errorf("AddSuffixIfNot = %q", got)
	}
	if !String.ContainsIgnoreCase("Hello World", "world") {
		t.Error("ContainsIgnoreCase 应为 true")
	}
	if got := String.SplitTrim("a, b , ,c", ","); len(got) != 3 || got[1] != "b" {
		t.Errorf("SplitTrim = %v", got)
	}
	if got := String.SplitByLength("abcdef", 2); len(got) != 3 || got[0] != "ab" || got[2] != "ef" {
		t.Errorf("SplitByLength = %v", got)
	}
	if got := String.ToSBC("aA1 "); got != "ａＡ１　" {
		t.Errorf("ToSBC = %q", got)
	}
	if got := String.ToDBC("ａＡ１　"); got != "aA1 " {
		t.Errorf("ToDBC = %q", got)
	}
	if got := String.Similar("hello", "hello"); got != 1 {
		t.Errorf("Similar 相同串应为 1, got %v", got)
	}
	if got := String.CommonPrefix([]string{"abc123", "abcd"}...); got != "abc" {
		t.Errorf("CommonPrefix = %q, want \"abc\"", got)
	}
	if got := String.LongestCommonSubstring("abcdef", "acef"); got != "acef" {
		t.Errorf("LongestCommonSubstring = %q, want \"acef\"", got)
	}
}

/* ---------------- 类型转换补充（含全部 error 路径） ---------------- */

func TestStrNumericConvertEdges(t *testing.T) {
	// 不带 E 的便捷转换：失败返回零值
	if got := String.ToInt("-7"); got != -7 {
		t.Errorf("ToInt(\"-7\") = %d, want -7", got)
	}
	if got := String.ToInt64("9223372036854775807"); got != 9223372036854775807 {
		t.Errorf("ToInt64(MaxInt64) = %d", got)
	}
	if got := String.ToInt64("abc"); got != 0 {
		t.Errorf("ToInt64(\"abc\") = %d, want 0", got)
	}
	if got := String.ToFloat64("3.14"); got != 3.14 {
		t.Errorf("ToFloat64(\"3.14\") = %v, want 3.14", got)
	}
	if got := String.ToFloat64("abc"); got != 0 {
		t.Errorf("ToFloat64(\"abc\") = %v, want 0", got)
	}
	// ToBool 仅接受 strconv.ParseBool 取值，"yes" 不在内
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"ToBool \"1\"", "1", true},
		{"ToBool \"TRUE\"", "TRUE", true},
		{"ToBool \"0\"", "0", false},
		{"ToBool \"yes\" 不支持", "yes", false},
	}
	for _, c := range cases {
		if got := String.ToBool(c.in); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
	// 带 E 的安全转换：成功与失败路径
	if n, err := String.ToInt64E("123"); err != nil || n != 123 {
		t.Errorf("ToInt64E(\"123\") = %d, %v; want 123, nil", n, err)
	}
	if _, err := String.ToInt64E("abc"); err == nil {
		t.Error("ToInt64E(\"abc\") 应返回错误")
	}
	if f, err := String.ToFloat64E("2.5"); err != nil || f != 2.5 {
		t.Errorf("ToFloat64E(\"2.5\") = %v, %v; want 2.5, nil", f, err)
	}
	if _, err := String.ToFloat64E("abc"); err == nil {
		t.Error("ToFloat64E(\"abc\") 应返回错误")
	}
	if b, err := String.ToBoolE("true"); err != nil || !b {
		t.Errorf("ToBoolE(\"true\") = %v, %v; want true, nil", b, err)
	}
	if _, err := String.ToBoolE("yes"); err == nil {
		t.Error("ToBoolE(\"yes\") 应返回错误")
	}
	// 整数转字符串
	if got := String.Int64ToStr(-99); got != "-99" {
		t.Errorf("Int64ToStr(-99) = %q, want \"-99\"", got)
	}
	if got := String.Int64ToStr(0); got != "0" {
		t.Errorf("Int64ToStr(0) = %q, want \"0\"", got)
	}
	// prec 边界：-1 最少位数、0 保留整数
	if got := String.Float64ToStr(3.14159, -1); got != "3.14159" {
		t.Errorf("Float64ToStr(prec=-1) = %q, want \"3.14159\"", got)
	}
	if got := String.Float64ToStr(5, 0); got != "5" {
		t.Errorf("Float64ToStr(5, 0) = %q, want \"5\"", got)
	}
}

/* ---------------- 二进制与位运算 ---------------- */

func TestStrBinaryAndBits(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want int64
	}{
		{"BinToInt 1010", "1010", 10},
		{"BinToInt 0", "0", 0},
		{"BinToInt 空串", "", 0},
		{"BinToInt 非法字符", "2", 0},
		{"BinToInt 字母", "abc", 0},
		{"BinToInt 负数", "-101", -5},
	}
	for _, c := range cases {
		if got := String.BinToInt(c.in); got != c.want {
			t.Errorf("%s: got %d, want %d", c.name, got, c.want)
		}
	}
	binCases := []struct {
		name string
		in   int64
		want string
	}{
		{"IntToBin 10", 10, "1010"},
		{"IntToBin 0", 0, "0"},
		{"IntToBin -5", -5, "-101"},
	}
	for _, c := range binCases {
		if got := String.IntToBin(c.in); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
	if !String.HasBits(0b1011, 0b0011) {
		t.Error("HasBits(0b1011, 0b0011) 应为 true")
	}
	if String.HasBits(0b1000, 0b0011) {
		t.Error("HasBits(0b1000, 0b0011) 应为 false")
	}
	if !String.HasBits(0, 0) {
		t.Error("HasBits(0, 0) 应为 true（空掩码恒成立）")
	}
	if !String.BinStrHasBits("1011", 0b0011) {
		t.Error("BinStrHasBits(\"1011\", 0b0011) 应为 true")
	}
	if String.BinStrHasBits("1000", 0b0011) {
		t.Error("BinStrHasBits(\"1000\", 0b0011) 应为 false")
	}
	if String.BinStrHasBits("abc", 0b0011) {
		t.Error("BinStrHasBits 非法二进制串应返回 false")
	}
}

/* ---------------- 判断与检查 ---------------- */

func TestStrEmptyBlankTrim(t *testing.T) {
	cases := []struct {
		name string
		got  bool
		want bool
	}{
		{"IsEmpty 空串", String.IsEmpty(""), true},
		{"IsEmpty 空格不算空", String.IsEmpty(" "), false},
		{"IsNotEmpty 空串", String.IsNotEmpty(""), false},
		{"IsNotEmpty 空格非空", String.IsNotEmpty(" "), true},
		{"IsBlank 纯空格", String.IsBlank(" "), true},
		{"IsBlank 制表换行", String.IsBlank("\t\n"), true},
		{"IsBlank 有内容", String.IsBlank("a"), false},
		{"IsNotBlank 有内容", String.IsNotBlank("a"), true},
		{"IsNotBlank 纯空格", String.IsNotBlank(" "), false},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, c.got, c.want)
		}
	}
	if got := String.Trim("  a b  "); got != "a b" {
		t.Errorf("Trim = %q, want \"a b\"", got)
	}
	if got := String.TrimStart("  \tx"); got != "x" {
		t.Errorf("TrimStart = %q, want \"x\"", got)
	}
	if got := String.TrimStart("x "); got != "x " {
		t.Errorf("TrimStart 只处理头部, got %q, want \"x \"", got)
	}
	if got := String.TrimEnd("x\n "); got != "x" {
		t.Errorf("TrimEnd = %q, want \"x\"", got)
	}
	if got := String.TrimEnd(" x"); got != " x" {
		t.Errorf("TrimEnd 只处理尾部, got %q, want \" x\"", got)
	}
	if got := String.TrimToNull("  "); got != "" {
		t.Errorf("TrimToNull(\"  \") = %q, want 空串", got)
	}
	if got := String.TrimToNull(" a "); got != "a" {
		t.Errorf("TrimToNull(\" a \") = %q, want \"a\"", got)
	}
	if got := String.TrimToNull("abc"); got != "abc" {
		t.Errorf("TrimToNull(\"abc\") = %q, want \"abc\"", got)
	}
}

func TestStrLenVsRuneCount(t *testing.T) {
	if got := String.Len("中国"); got != 6 {
		t.Errorf("Len(\"中国\") = %d, want 6（UTF-8 字节数）", got)
	}
	if got := String.Len("abc"); got != 3 {
		t.Errorf("Len(\"abc\") = %d, want 3", got)
	}
	if got := String.RuneCount("中国"); got != 2 {
		t.Errorf("RuneCount(\"中国\") = %d, want 2", got)
	}
	if got := String.RuneCount("abc"); got != 3 {
		t.Errorf("RuneCount(\"abc\") = %d, want 3", got)
	}
	if got := String.RuneCount(""); got != 0 {
		t.Errorf("RuneCount(\"\") = %d, want 0", got)
	}
}

/* ---------------- 查找 ---------------- */

func TestStrFindChars(t *testing.T) {
	cases := []struct {
		name string
		got  bool
		want bool
	}{
		{"Contains 命中", String.Contains("hello", "ell"), true},
		{"Contains 未命中", String.Contains("hello", "xyz"), false},
		{"Contains 空查找串恒真", String.Contains("hello", ""), true},
		{"StartWith 协议头", String.StartWith("https://a.com", "https://"), true},
		{"StartWith 不在前", String.StartWith("abc", "b"), false},
		{"EndWith 扩展名", String.EndWith("main.go", ".go"), true},
		{"EndWith 大小写敏感", String.EndWith("main.go", ".GO"), false},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, c.got, c.want)
		}
	}
	if got := String.Count("cheese", "e"); got != 3 {
		t.Errorf("Count(\"cheese\",\"e\") = %d, want 3", got)
	}
	if got := String.Count("aaa", "aa"); got != 1 {
		t.Errorf("Count 非重叠计数 = %d, want 1", got)
	}
	if got := String.Index("ABC_xyz", "xyz"); got != 4 {
		t.Errorf("Index = %d, want 4", got)
	}
	if got := String.Index("abc", "x"); got != -1 {
		t.Errorf("Index 未找到 = %d, want -1", got)
	}
	if got := String.LastIndex("a/b/c", "/"); got != 3 {
		t.Errorf("LastIndex = %d, want 3", got)
	}
	if got := String.LastIndex("abc", "/"); got != -1 {
		t.Errorf("LastIndex 未找到 = %d, want -1", got)
	}
}

/* ---------------- 替换 ---------------- */

func TestStrReplaceFirstAndAll(t *testing.T) {
	if got := String.Replace("aaa", "a", "b"); got != "baa" {
		t.Errorf("Replace 只替换首个 = %q, want \"baa\"", got)
	}
	if got := String.Replace("abc", "x", "y"); got != "abc" {
		t.Errorf("Replace 未找到应原样, got %q", got)
	}
	if got := String.ReplaceAll("aaa", "a", "b"); got != "bbb" {
		t.Errorf("ReplaceAll = %q, want \"bbb\"", got)
	}
	if got := String.ReplaceAll("a.b.c", ".", "-"); got != "a-b-c" {
		t.Errorf("ReplaceAll = %q, want \"a-b-c\"", got)
	}
	if got := String.ReplaceAll("abc", "x", "y"); got != "abc" {
		t.Errorf("ReplaceAll 未找到应原样, got %q", got)
	}
}

/* ---------------- 拆分与拼接 ---------------- */

func TestStrSplitAndJoin(t *testing.T) {
	if got := String.Split("a,b,c", ","); len(got) != 3 || got[0] != "a" || got[2] != "c" {
		t.Errorf("Split = %v", got)
	}
	// 空串返回 [""]（长度 1）
	if got := String.Split("", ","); len(got) != 1 || got[0] != "" {
		t.Errorf("Split(\"\") = %v, want [\"\"]", got)
	}
	// 中间空项保留
	if got := String.Split("a,,c", ","); len(got) != 3 || got[1] != "" {
		t.Errorf("Split 含空项 = %v", got)
	}
	if got := String.Join([]string{"a", "b", "c"}, "-"); got != "a-b-c" {
		t.Errorf("Join = %q, want \"a-b-c\"", got)
	}
	if got := String.Join(nil, ","); got != "" {
		t.Errorf("Join(nil) = %q, want 空串", got)
	}
	if got := String.Join([]string{"a"}, ","); got != "a" {
		t.Errorf("Join 单元素 = %q, want \"a\"", got)
	}
}

/* ---------------- Sub 边界 ---------------- */

func TestStrSubBoundary(t *testing.T) {
	cases := []struct {
		name  string
		str   string
		start int
		end   int
		want  string
	}{
		{"start>end 自动交换", "abcdef", 4, 2, "cd"},
		{"负下标中文", "你好世界", -2, -1, "世"},
		{"负下标越下界收敛", "abc", -10, 1, "a"},
		{"双双越上界收敛为空", "abc", 5, 6, ""},
		{"start==end 为空", "abc", 3, 3, ""},
		{"end 越界收敛到末尾", "你好世界", 0, 100, "你好世界"},
		{"空串", "", 0, 1, ""},
	}
	for _, c := range cases {
		if got := String.Sub(c.str, c.start, c.end); got != c.want {
			t.Errorf("Sub %s: Sub(%q,%d,%d) = %q, want %q", c.name, c.str, c.start, c.end, got, c.want)
		}
	}
}

/* ---------------- SubBefore/SubAfter/SubBetween 边界 ---------------- */

func TestStrSubBeforeAfterBoundary(t *testing.T) {
	if got := String.SubBefore("abcba", "b", true); got != "abc" {
		t.Errorf("SubBefore(isLast) = %q, want \"abc\"", got)
	}
	if got := String.SubBefore("abc", "a", false); got != "" {
		t.Errorf("SubBefore 分隔符在开头应返回空串, got %q", got)
	}
	if got := String.SubBefore("abc", "", false); got != "" {
		t.Errorf("SubBefore 空分隔符应返回空串, got %q", got)
	}
	if got := String.SubBefore("", "a", false); got != "" {
		t.Errorf("SubBefore 空串应返回空串, got %q", got)
	}
	if got := String.SubAfter("abc", "c", false); got != "" {
		t.Errorf("SubAfter 分隔符在末尾应返回空串, got %q", got)
	}
	if got := String.SubAfter("abc", "x", false); got != "" {
		t.Errorf("SubAfter 未找到应返回空串, got %q", got)
	}
	if got := String.SubAfter("abc", "", false); got != "" {
		t.Errorf("SubAfter 空分隔符应返回空串, got %q", got)
	}
	if got := String.SubAfter("", "a", false); got != "" {
		t.Errorf("SubAfter 空串应返回空串, got %q", got)
	}
	if got := String.SubBetween("[abc]", "[", "]"); got != "abc" {
		t.Errorf("SubBetween 起始标识 = %q, want \"abc\"", got)
	}
	if got := String.SubBetween("abc", "[", "]"); got != "" {
		t.Errorf("SubBetween 左标识未找到应返回空串, got %q", got)
	}
	if got := String.SubBetween("a[b", "[", "]"); got != "" {
		t.Errorf("SubBetween 右标识未找到应返回空串, got %q", got)
	}
	if got := String.SubBetween("abc", "", "]"); got != "" {
		t.Errorf("SubBetween 空标识应返回空串, got %q", got)
	}
	if got := String.SubBetween("", "[", "]"); got != "" {
		t.Errorf("SubBetween 空串应返回空串, got %q", got)
	}
}

/* ---------------- 大小写转换 ---------------- */

func TestStrLetterCase(t *testing.T) {
	if got := String.ToLower("Love GoLang"); got != "love golang" {
		t.Errorf("ToLower = %q, want \"love golang\"", got)
	}
	if got := String.ToUpper("love 中国"); got != "LOVE 中国" {
		t.Errorf("ToUpper 保留中文 = %q, want \"LOVE 中国\"", got)
	}
	if got := String.LowerFirst("User"); got != "user" {
		t.Errorf("LowerFirst = %q, want \"user\"", got)
	}
	if got := String.LowerFirst("USER"); got != "uSER" {
		t.Errorf("LowerFirst 只动首字母 = %q, want \"uSER\"", got)
	}
	if got := String.LowerFirst(""); got != "" {
		t.Errorf("LowerFirst(\"\") = %q, want 空串", got)
	}
	if got := String.LowerFirst("你好"); got != "你好" {
		t.Errorf("LowerFirst 非字母首字符应原样, got %q", got)
	}
	if got := String.UpperFirst(""); got != "" {
		t.Errorf("UpperFirst(\"\") = %q, want 空串", got)
	}
	if got := String.UpperFirst("你好"); got != "你好" {
		t.Errorf("UpperFirst 非字母首字符应原样, got %q", got)
	}
	if got := String.UpperFirst("aBc"); got != "ABc" {
		t.Errorf("UpperFirst 只动首字母 = %q, want \"ABc\"", got)
	}
	if got := String.SwapCase("Love Go"); got != "lOVE gO" {
		t.Errorf("SwapCase = %q, want \"lOVE gO\"", got)
	}
	if got := String.SwapCase("中文aB"); got != "中文Ab" {
		t.Errorf("SwapCase 中文不变 = %q, want \"中文Ab\"", got)
	}
}

/* ---------------- 重复与填充 ---------------- */

func TestStrRepeatZero(t *testing.T) {
	if got := String.Repeat("ab", 3); got != "ababab" {
		t.Errorf("Repeat = %q, want \"ababab\"", got)
	}
	if got := String.Repeat("x", 0); got != "" {
		t.Errorf("Repeat(count=0) = %q, want 空串", got)
	}
	if got := String.Repeat("x", -2); got != "" {
		t.Errorf("Repeat(count<0) = %q, want 空串", got)
	}
}

func TestStrPadAndCenter(t *testing.T) {
	if got := String.PadAfter("ab", 4, '*'); got != "ab**" {
		t.Errorf("PadAfter = %q, want \"ab**\"", got)
	}
	if got := String.PadAfter("abc", 2, '*'); got != "abc" {
		t.Errorf("PadAfter 超长应原样, got %q", got)
	}
	if got := String.PadPre("1234", 3, '0'); got != "1234" {
		t.Errorf("PadPre 超长应原样, got %q", got)
	}
	if got := String.PadPre("你", 3, '0'); got != "00你" {
		t.Errorf("PadPre 按字符计长 = %q, want \"00你\"", got)
	}
	// 差值为偶数时两侧均分
	if got := String.Center("ab", 6, '*'); got != "**ab**" {
		t.Errorf("Center 偶数差值 = %q, want \"**ab**\"", got)
	}
	if got := String.Center("abc", 2, '*'); got != "abc" {
		t.Errorf("Center 超长应原样, got %q", got)
	}
}

/* ---------------- 拼接 ---------------- */

func TestStrConcatVariants(t *testing.T) {
	if got := String.Concat("a", "b", "c"); got != "abc" {
		t.Errorf("Concat = %q, want \"abc\"", got)
	}
	if got := String.Concat(); got != "" {
		t.Errorf("Concat 无参 = %q, want 空串", got)
	}
	if got := String.ConcatAny(3.5); got != "3.5" {
		t.Errorf("ConcatAny 非字符串经 fmt 转换 = %q, want \"3.5\"", got)
	}
	if got := String.ConcatAny("a", 1, "b", 2.5, "c", true); got != "a1b2.5ctrue" {
		t.Errorf("ConcatAny 混合类型 = %q, want \"a1b2.5ctrue\"", got)
	}
}

/* ---------------- FormatBytes 单位边界 ---------------- */

func TestStrFormatBytesUnits(t *testing.T) {
	cases := []struct {
		name string
		in   uint64
		want string
	}{
		{"不足 1KB", 512, "512 B"},
		{"恰好 1KB", 1024, "1.00 KB"},
		{"恰好 1MB", 5 * 1024 * 1024, "5.00 MB"},
		{"恰好 3GB", 3 * 1024 * 1024 * 1024, "3.00 GB"},
		{"1 字节", 1, "1 B"},
	}
	for _, c := range cases {
		if got := String.FormatBytes(c.in); got != c.want {
			t.Errorf("%s: FormatBytes(%d) = %q, want %q", c.name, c.in, got, c.want)
		}
	}
}

/* ---------------- RandomString ---------------- */

func TestStrRandomStringCharset(t *testing.T) {
	if got := String.RandomString(0); got != "" {
		t.Errorf("RandomString(0) = %q, want 空串", got)
	}
	if got := String.RandomString(-1); got != "" {
		t.Errorf("RandomString(-1) = %q, want 空串", got)
	}
	got := String.RandomString(64)
	if len(got) != 64 {
		t.Errorf("RandomString(64) 长度 = %d, want 64", len(got))
	}
	for _, r := range got {
		if !strings.ContainsRune(randomCharset, r) {
			t.Errorf("RandomString 产生字符集外的字符 %q", r)
		}
	}
}

/* ---------------- 正则过滤 ---------------- */

func TestStrRegexFilter(t *testing.T) {
	if got := String.FilterByRegex(`\d`, "a1b2", ""); got != "ab" {
		t.Errorf("FilterByRegex = %q, want \"ab\"", got)
	}
	if got := String.FilterByRegex("o", "foo", "0"); got != "f00" {
		t.Errorf("FilterByRegex 替换全部匹配 = %q, want \"f00\"", got)
	}
	// 非法正则原样返回
	if got := String.FilterByRegex("[", "abc", "x"); got != "abc" {
		t.Errorf("FilterByRegex 非法表达式应原样返回, got %q", got)
	}
	if got := String.FilterSpecialChar("a+b|c"); got != "abc" {
		t.Errorf("FilterSpecialChar = %q, want \"abc\"", got)
	}
	if got := String.FilterSpecialChar("你好=世界"); got != "你好世界" {
		t.Errorf("FilterSpecialChar 中文保留 = %q, want \"你好世界\"", got)
	}
	if got := String.FilterSpecialChar("abc"); got != "abc" {
		t.Errorf("FilterSpecialChar 无特殊字符应原样, got %q", got)
	}
	if got := String.FilterUrlPrefix("https://a.com"); got != "a.com" {
		t.Errorf("FilterUrlPrefix = %q, want \"a.com\"", got)
	}
	// \\w+:// 不限于 http(s)
	if got := String.FilterUrlPrefix("ftp://a.com"); got != "a.com" {
		t.Errorf("FilterUrlPrefix 任意协议 = %q, want \"a.com\"", got)
	}
	if got := String.FilterUrlPrefix("www.example.com"); got != "www.example.com" {
		t.Errorf("FilterUrlPrefix 无协议应原样, got %q", got)
	}
}

/* ---------------- IP 校验 ---------------- */

func TestStrIPFamilyCheck(t *testing.T) {
	cases := []struct {
		name string
		got  bool
		want bool
	}{
		{"IsIP IPv4", String.IsIP("192.168.1.1"), true},
		{"IsIP IPv6", String.IsIP("fe80::1"), true},
		{"IsIP 非法", String.IsIP("abc"), false},
		{"IsIP 超段", String.IsIP("999.1.1.1"), false},
		{"IsIPv4 回环 v6 不是 v4", String.IsIPv4("::1"), false},
		{"IsIPv6 回环 v6", String.IsIPv6("::1"), true},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, c.got, c.want)
		}
	}
}

/* ---------------- 域名校验 ---------------- */

func TestStrDomainCheck(t *testing.T) {
	domainCases := []struct {
		name string
		in   string
		want bool
	}{
		{"IsDomain 多级", "www.example.com", true},
		{"IsDomain 含中划线", "sub.domain-x.co", true},
		{"IsDomain 带协议", "https://a.com", false},
		{"IsDomain 无点单标签", "localhost", false},
		{"IsDomain 连续点", "a..b", false},
	}
	for _, c := range domainCases {
		if got := String.IsDomain(c.in); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
	dnsCases := []struct {
		name string
		in   string
		want bool
	}{
		{"IsDnsDomain 标准", "www.example.com", true},
		{"IsDnsDomain 允许根点", "example.com.", true},
		{"IsDnsDomain 顶级域过短", "abc.c", false},
		{"IsDnsDomain 段首中划线", "-abc.com", false},
		{"IsDnsDomain 无点", "localhost", false},
	}
	for _, c := range dnsCases {
		if got := String.IsDnsDomain(c.in); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

/* ---------------- 身份证与 URL 校验 ---------------- */

func TestStrIDCardURLCheck(t *testing.T) {
	if !String.IsIDCard("321083197812162119") {
		t.Error("IsIDCard 18 位数字应为 true")
	}
	if !String.IsIDCard("32108319781216211X") {
		t.Error("IsIDCard 末位 X 应为 true")
	}
	if String.IsIDCard("3210831978121621199") {
		t.Error("IsIDCard 19 位应为 false")
	}
	if String.IsIDCard("123") {
		t.Error("IsIDCard 位数不足应为 false")
	}
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"IsURL https", "https://a.com", true},
		{"IsURL 带路径参数", "http://a.com/x?y=1", true},
		{"IsURL ftp 不匹配", "ftp://a.com", false},
		{"IsURL 无协议", "example.com", false},
		{"IsURL 协议后无内容", "https://", false},
	}
	for _, c := range cases {
		if got := String.IsURL(c.in); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

/* ---------------- 默认值 ---------------- */

func TestStrDefaultValuePick(t *testing.T) {
	if got := String.NullToDefault("", "d"); got != "d" {
		t.Errorf("NullToDefault 空串 = %q, want \"d\"", got)
	}
	if got := String.NullToDefault("x", "d"); got != "x" {
		t.Errorf("NullToDefault 非空 = %q, want \"x\"", got)
	}
	if got := String.EmptyToDefault("", "d"); got != "d" {
		t.Errorf("EmptyToDefault 空串 = %q, want \"d\"", got)
	}
	if got := String.EmptyToDefault("x", "d"); got != "x" {
		t.Errorf("EmptyToDefault 非空 = %q, want \"x\"", got)
	}
	// BlankToDefault 命中时返回原串（不裁剪）
	if got := String.BlankToDefault(" a ", "d"); got != " a " {
		t.Errorf("BlankToDefault 非空白应返回原串, got %q", got)
	}
}

/* ---------------- 前后缀处理 ---------------- */

func TestStrPrefixSuffixOps(t *testing.T) {
	if got := String.RemovePrefix("abc", "x"); got != "abc" {
		t.Errorf("RemovePrefix 不匹配应原样, got %q", got)
	}
	if got := String.RemoveSuffix("main.go", ".go"); got != "main" {
		t.Errorf("RemoveSuffix = %q, want \"main\"", got)
	}
	if got := String.RemoveSuffix("abc", "x"); got != "abc" {
		t.Errorf("RemoveSuffix 不匹配应原样, got %q", got)
	}
	if got := String.AddPrefixIfNot("a.com", "https://"); got != "https://a.com" {
		t.Errorf("AddPrefixIfNot 补前缀 = %q", got)
	}
	if got := String.AddPrefixIfNot("https://a.com", "https://"); got != "https://a.com" {
		t.Errorf("AddPrefixIfNot 已有前缀应原样, got %q", got)
	}
	if got := String.AddPrefixIfNot("", "p"); got != "" {
		t.Errorf("AddPrefixIfNot 空串应原样, got %q", got)
	}
	if got := String.AddSuffixIfNot("logs/", "/"); got != "logs/" {
		t.Errorf("AddSuffixIfNot 已有后缀应原样, got %q", got)
	}
}

/* ---------------- 包裹判断 ---------------- */

func TestStrWrapOps(t *testing.T) {
	if got := String.Wrap("${id}", "${", "}"); got != "${id}" {
		t.Errorf("Wrap 已包裹不应重复包裹, got %q", got)
	}
	if got := String.UnWrap("id", "${", "}"); got != "id" {
		t.Errorf("UnWrap 未包裹应原样, got %q", got)
	}
	cases := []struct {
		name string
		got  bool
		want bool
	}{
		{"IsWrap 完整包裹", String.IsWrap("${name}", "${", "}"), true},
		{"IsWrap 缺后缀", String.IsWrap("${name", "${", "}"), false},
		{"IsWrap 缺前缀", String.IsWrap("name}", "${", "}"), false},
		{"IsWrap 长度恰达下限", String.IsWrap("${}", "${", "}"), true},
		{"IsWrap 长度不足", String.IsWrap("${", "${", "$"), false},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, c.got, c.want)
		}
	}
}

/* ---------------- 命名风格：ToSymbolCase 专有名词语义 ---------------- */

func TestStrNamingCaseExt(t *testing.T) {
	// 连续大写按专有名词保留原大写（hutool 完整移植的真实行为）
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"含数字尾部大写", "customerNickV2", "customer_nick_v2"},
		{"连续大写开头", "HTTPServer", "HTTP_server"},
		{"专有名词保留", "PNLabel", "PN_label"},
		{"前导下划线后大写", "_Abb", "_abb"},
		{"纯小写不变", "username", "username"},
	}
	for _, c := range cases {
		if got := String.ToSymbolCase(c.in, '_'); got != c.want {
			t.Errorf("ToSymbolCase %s: got %q, want %q", c.name, got, c.want)
		}
	}
	if got := String.ToSymbolCase("userName", '-'); got != "user-name" {
		t.Errorf("ToSymbolCase 自定义连接符 = %q, want \"user-name\"", got)
	}
	// ToCamelCase 边界
	if got := String.ToCamelCase("USER_NAME"); got != "userName" {
		t.Errorf("ToCamelCase 全大写段 = %q, want \"userName\"", got)
	}
	if got := String.ToCamelCase("user_name_extra"); got != "userNameExtra" {
		t.Errorf("ToCamelCase 多段 = %q, want \"userNameExtra\"", got)
	}
	if got := String.ToCamelCase("HTTPServer"); got != "HTTPServer" {
		t.Errorf("ToCamelCase 无连接符应原样返回, got %q", got)
	}
	if got := String.ToPascalCase("USER_NAME"); got != "UserName" {
		t.Errorf("ToPascalCase 首字母大写 = %q, want \"UserName\"", got)
	}
	if got := String.ToUnderlineCase("userName"); got != "user_name" {
		t.Errorf("ToUnderlineCase = %q, want \"user_name\"", got)
	}
	if got := String.ToKebabCase("HTTPServer"); got != "HTTP-server" {
		t.Errorf("ToKebabCase 专有名词 = %q, want \"HTTP-server\"", got)
	}
}

/* ---------------- 模板格式化边界 ---------------- */

func TestStrTemplateFormat(t *testing.T) {
	if got := String.Format("{}, {}", "a"); got != "a, {}" {
		t.Errorf("Format 占位符多于参数应保留 = %q, want \"a, {}\"", got)
	}
	if got := String.Format("{} {}", true, 3.5); got != "true 3.5" {
		t.Errorf("Format 非字符串参数 = %q, want \"true 3.5\"", got)
	}
	if got := String.Format("no args"); got != "no args" {
		t.Errorf("Format 无参数应原样, got %q", got)
	}
	if got := String.Format(""); got != "" {
		t.Errorf("Format 空模板 = %q, want 空串", got)
	}
	if got := String.FormatWithMap("{a}-{b}", map[string]string{"a": "1", "b": "2"}); got != "1-2" {
		t.Errorf("FormatWithMap 多占位符 = %q, want \"1-2\"", got)
	}
	if got := String.FormatWithMap("abc", nil); got != "abc" {
		t.Errorf("FormatWithMap 空 map 应原样, got %q", got)
	}
}

/* ---------------- 相似度 ---------------- */

func TestStrSimilarMetrics(t *testing.T) {
	cases := []struct {
		name string
		a    string
		b    string
		want float64
	}{
		{"两空串相似度为 1", "", "", 1},
		{"无公共字符为 0", "abc", "xyz", 0},
		{"符号剔除后相同", "a b!c", "abc", 1},
		{"部分相似", "abcde", "abc", 0.6},
	}
	for _, c := range cases {
		if got := String.Similar(c.a, c.b); got != c.want {
			t.Errorf("Similar %s: Similar(%q,%q) = %v, want %v", c.name, c.a, c.b, got, c.want)
		}
	}
	if got := String.SimilarPercent("abc", "abd", 2); got != "66.67%" {
		t.Errorf("SimilarPercent = %q, want \"66.67%%\"", got)
	}
	if got := String.SimilarPercent("hello", "hello", 0); got != "100%" {
		t.Errorf("SimilarPercent 整数百分比 = %q, want \"100%%\"", got)
	}
	if got := String.LongestCommonSubstring("你好世界", "你界"); got != "你界" {
		t.Errorf("LongestCommonSubstring 中文 = %q, want \"你界\"", got)
	}
	if got := String.LongestCommonSubstring("", "abc"); got != "" {
		t.Errorf("LongestCommonSubstring 空串 = %q, want 空串", got)
	}
	if got := String.LongestCommonSubstring("abc", "abc"); got != "abc" {
		t.Errorf("LongestCommonSubstring 相同串 = %q, want \"abc\"", got)
	}
	if got := String.LongestCommonSubstring("abc", "xyz"); got != "" {
		t.Errorf("LongestCommonSubstring 无公共 = %q, want 空串", got)
	}
}

/* ---------------- Brief/Hide 边界 ---------------- */

func TestStrBriefHideEdges(t *testing.T) {
	cases := []struct {
		name string
		in   string
		n    int
		want string
	}{
		{"未超长原样", "abc", 5, "abc"},
		{"maxLength 等于长度原样", "abc", 3, "abc"},
		{"maxLength<=3 直接截断", "abcdef", 3, "abc"},
		{"maxLength<=3 中文", "你好世界", 3, "你好世"},
		{"maxLength=0 截为空", "abc", 0, ""},
	}
	for _, c := range cases {
		if got := String.Brief(c.in, c.n); got != c.want {
			t.Errorf("Brief %s: Brief(%q,%d) = %q, want %q", c.name, c.in, c.n, got, c.want)
		}
	}
	hideCases := []struct {
		name  string
		str   string
		start int
		end   int
		want  string
	}{
		{"负下标从末尾倒数", "13812345678", -4, -1, "1381234***8"},
		{"end 越界收敛", "abc", 0, 100, "***"},
		{"start>=end 原样", "abc", 2, 2, "abc"},
		{"空串原样", "", 0, 1, ""},
	}
	for _, c := range hideCases {
		if got := String.Hide(c.str, c.start, c.end); got != c.want {
			t.Errorf("Hide %s: Hide(%q,%d,%d) = %q, want %q", c.name, c.str, c.start, c.end, got, c.want)
		}
	}
}

/* ---------------- 全角/半角往返 ---------------- */

func TestStrSBCDBCRoundTrip(t *testing.T) {
	if got := String.ToSBC("中国"); got != "中国" {
		t.Errorf("ToSBC 中文应原样, got %q", got)
	}
	if got := String.ToSBC("!"); got != "！" {
		t.Errorf("ToSBC 半角感叹号 = %q, want \"！\"", got)
	}
	if got := String.ToDBC("！"); got != "!" {
		t.Errorf("ToDBC 全角感叹号 = %q, want \"!\"", got)
	}
	// ASCII 可见字符与空格应能完整往返
	const sample = "Ab12!~ "
	if got := String.ToDBC(String.ToSBC(sample)); got != sample {
		t.Errorf("ToDBC(ToSBC(%q)) 往返 = %q", sample, got)
	}
}

/* ---------------- 公共前后缀 ---------------- */

func TestStrCommonAffix(t *testing.T) {
	if got := String.CommonPrefix("abc", "xyz"); got != "" {
		t.Errorf("CommonPrefix 无公共 = %q, want 空串", got)
	}
	if got := String.CommonPrefix("abc", "ab", "abcd"); got != "ab" {
		t.Errorf("CommonPrefix 多参数 = %q, want \"ab\"", got)
	}
	if got := String.CommonPrefix("abcd"); got != "abcd" {
		t.Errorf("CommonPrefix 单参数 = %q, want \"abcd\"", got)
	}
	if got := String.CommonPrefix(); got != "" {
		t.Errorf("CommonPrefix 无参 = %q, want 空串", got)
	}
	if got := String.CommonSuffix("a.txt", "b.txt"); got != ".txt" {
		t.Errorf("CommonSuffix = %q, want \".txt\"", got)
	}
	if got := String.CommonSuffix("abc", "xyz"); got != "" {
		t.Errorf("CommonSuffix 无公共 = %q, want 空串", got)
	}
	if got := String.CommonSuffix("abcd", "abcd"); got != "abcd" {
		t.Errorf("CommonSuffix 相同串 = %q, want \"abcd\"", got)
	}
	if got := String.CommonSuffix("abcd"); got != "abcd" {
		t.Errorf("CommonSuffix 单参数 = %q, want \"abcd\"", got)
	}
	if got := String.CommonSuffix(); got != "" {
		t.Errorf("CommonSuffix 无参 = %q, want 空串", got)
	}
}
