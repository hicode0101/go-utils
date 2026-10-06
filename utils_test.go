package utils

import (
	"bytes"
	"strings"
	"testing"
)

/* ---------------- 字符串 ---------------- */

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
	// SubStr 按字符截取，中文安全
	if got := String.SubStr("你好世界", 1, 3); got != "好世" {
		t.Errorf("SubStr(中文,1,3) = %q, want \"好世\"", got)
	}
	// 越界自动收敛
	if got := String.SubStr("abc", 2, 100); got != "c" {
		t.Errorf("SubStr 越界 = %q, want \"c\"", got)
	}
	if got := String.Reverse("中国abc"); got != "cba国中" {
		t.Errorf("Reverse = %q, want \"cba国中\"", got)
	}
	if got := String.Truncate("你好世界欢迎", 5, "..."); got != "你好世界欢..." {
		t.Errorf("Truncate = %q, want \"你好世界欢...\"", got)
	}
	if got := String.PadLeft("7", 3, '0'); got != "007" {
		t.Errorf("PadLeft = %q, want \"007\"", got)
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

/* ---------------- 集合（泛型函数） ---------------- */

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

/* ---------------- 泛型辅助 ---------------- */

func TestIfAndSwap(t *testing.T) {
	if got := If(1 > 2, "yes", "no"); got != "no" {
		t.Errorf("If = %q, want \"no\"", got)
	}
	a, b := Swap(1, 2)
	if a != 2 || b != 1 {
		t.Errorf("Swap = %d, %d; want 2, 1", a, b)
	}
}

/* ---------------- 数学 ---------------- */

func TestMath(t *testing.T) {
	if got := Math.PageCount(101, 20); got != 6 {
		t.Errorf("PageCount(101,20) = %d, want 6", got)
	}
	if got := Math.PageCount(0, 20); got != 0 {
		t.Errorf("PageCount(0,20) = %d, want 0", got)
	}
	if got := Math.PageCount(10, 0); got != 0 {
		t.Errorf("PageCount(10,0) = %d, want 0（除零保护）", got)
	}
	if got := Math.RoundToInt(5.5); got != 6 {
		t.Errorf("RoundToInt(5.5) = %d, want 6", got)
	}
	if got := Math.CeilToInt(5.3); got != 6 {
		t.Errorf("CeilToInt(5.3) = %d, want 6", got)
	}
	if got := Math.RoundFloat(3.14159, 2); got != 3.14 {
		t.Errorf("RoundFloat(3.14159,2) = %v, want 3.14", got)
	}
	if got := Math.AbsInt(-7); got != 7 {
		t.Errorf("AbsInt(-7) = %d, want 7", got)
	}
}

/* ---------------- 日期 ---------------- */

func TestDate(t *testing.T) {
	tm := Date.MustParse("2024-06-01 12:30:45")
	if got := Date.Format(tm); got != "2024-06-01 12:30:45" {
		t.Errorf("Format = %q", got)
	}
	if got := Date.FormatCompact(tm); got != "20240601123045" {
		t.Errorf("FormatCompact = %q", got)
	}
	if got := Date.FormatDate(tm); got != "2024-06-01" {
		t.Errorf("FormatDate = %q", got)
	}
	// 时间戳往返
	if !Date.FromUnixMilli(Date.ToUnixMilli(tm)).Equal(tm) {
		t.Error("毫秒时间戳往返应相等")
	}
	if got := Date.DiffDays(Date.MustParse("2024-06-01 10:00:00"), Date.MustParse("2024-06-02 10:00:00")); got != 1 {
		t.Errorf("DiffDays = %d, want 1", got)
	}
	if !Date.IsSameDay(tm, Date.MustParse("2024-06-01 23:59:59")) {
		t.Error("IsSameDay 同日应为 true")
	}
	if got := Date.FormatDate(Date.EndOfMonth(tm)); got != "2024-06-30" {
		t.Errorf("EndOfMonth = %q, want \"2024-06-30\"", got)
	}
	if got := Date.FormatDate(Date.StartOfMonth(tm)); got != "2024-06-01" {
		t.Errorf("StartOfMonth = %q", got)
	}
	if _, err := Date.Parse("bad"); err == nil {
		t.Error("Parse 非法输入应返回错误")
	}
	if !Date.MustParse("bad").IsZero() {
		t.Error("MustParse 失败应返回零值时间")
	}
}

/* ---------------- 字节 ---------------- */

func TestBytes(t *testing.T) {
	var buf bytes.Buffer
	Bytes.WriteUint16(&buf, 0x1234)
	if got := Bytes.ReadUint16(buf.Bytes()); got != 0x1234 {
		t.Errorf("Uint16 往返 = %#x", got)
	}

	buf.Reset()
	Bytes.WriteInt64(&buf, -99)
	if got := Bytes.ReadInt64(buf.Bytes()); got != -99 {
		t.Errorf("Int64 往返 = %d", got)
	}

	buf.Reset()
	Bytes.WriteFloat64(&buf, 3.14)
	if got := Bytes.ReadFloat64(buf.Bytes()); got != 3.14 {
		t.Errorf("Float64 往返 = %v", got)
	}

	buf.Reset()
	Bytes.WriteFixedString(&buf, 8, "ab")
	if got := Bytes.ReadFixedString(buf.Bytes()); got != "ab" {
		t.Errorf("FixedString 往返 = %q, want \"ab\"", got)
	}
	if buf.Len() != 8 {
		t.Errorf("定长字符串应占 8 字节，实际 %d", buf.Len())
	}

	buf.Reset()
	Bytes.WriteUtf16String(&buf, 4, "你好")
	if got := Bytes.ReadUtf16String(buf.Bytes()); got != "你好" {
		t.Errorf("Utf16String 往返 = %q, want \"你好\"", got)
	}
	if buf.Len() != 8 { // 4 字符 * 2 字节
		t.Errorf("UTF-16 定长应占 8 字节，实际 %d", buf.Len())
	}

	// 超长截断
	buf.Reset()
	Bytes.WriteFixedString(&buf, 3, "hello")
	if got := Bytes.ReadFixedString(buf.Bytes()); got != "hel" {
		t.Errorf("超长截断 = %q, want \"hel\"", got)
	}
}

/* ---------------- 哈希与编码 ---------------- */

func TestCrypto(t *testing.T) {
	if got := Crypto.MD5("abc"); got != "900150983cd24fb0d6963f7d28e17f72" {
		t.Errorf("MD5(abc) = %q", got)
	}
	if got := Crypto.SHA256("abc"); len(got) != 64 {
		t.Errorf("SHA256 长度 = %d, want 64", len(got))
	}
	u := Crypto.Uuid()
	if len(u) != 36 || strings.Count(u, "-") != 4 || u[14] != '4' {
		t.Errorf("Uuid 格式错误: %q", u)
	}
	enc := Crypto.Base64Encode("hello")
	dec, err := Crypto.Base64Decode(enc)
	if err != nil || dec != "hello" {
		t.Errorf("Base64 往返失败: %q, %v", dec, err)
	}
	if got := Crypto.HmacSha256("key", "data"); len(got) != 64 {
		t.Errorf("HmacSha256 长度错误: %d", len(got))
	}
}

/* ---------------- JSON ---------------- */

type demoUser struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestJson(t *testing.T) {
	u := demoUser{Name: "Tom", Age: 18}
	if got := Json.ToJsonString(u); got != `{"name":"Tom","age":18}` {
		t.Errorf("ToJsonString = %q", got)
	}
	var back demoUser
	if err := Json.FromJson([]byte(`{"name":"Tom","age":18}`), &back); err != nil {
		t.Fatalf("FromJson 失败: %v", err)
	}
	if back.Name != "Tom" || back.Age != 18 {
		t.Errorf("FromJson 结果错误: %+v", back)
	}
	if !Json.Valid([]byte(`{"a":1}`)) || Json.Valid([]byte(`{a:}`)) {
		t.Error("Valid 校验错误")
	}
}

/* ---------------- 文件 ---------------- */

func TestFile(t *testing.T) {
	dir := t.TempDir()
	path := File.Join(dir, "a.txt")

	if err := File.WriteText(path, "hello"); err != nil {
		t.Fatalf("WriteText 失败: %v", err)
	}
	if got, _ := File.ReadText(path); got != "hello" {
		t.Errorf("ReadText = %q", got)
	}
	if err := File.AppendText(path, "world"); err != nil {
		t.Fatalf("AppendText 失败: %v", err)
	}
	if got, _ := File.ReadText(path); got != "helloworld\n" {
		t.Errorf("Append 后内容 = %q", got)
	}
	if got := File.Size(path); got != int64(len("helloworld\n")) {
		t.Errorf("Size = %d", got)
	}
	if !File.Exists(path) {
		t.Error("文件应存在")
	}

	dst := File.Join(dir, "b.txt")
	if err := File.Copy(path, dst); err != nil {
		t.Fatalf("Copy 失败: %v", err)
	}
	if !File.Exists(dst) {
		t.Error("复制后的文件应存在")
	}

	files, err := File.ListFiles(dir, ".txt")
	if err != nil || len(files) != 2 {
		t.Errorf("ListFiles = %v, %v; want 2 个 txt 文件", files, err)
	}

	if got := File.Name("/app/src/main.go"); got != "main.go" {
		t.Errorf("Name = %q", got)
	}
	if got := File.NameWithoutExt("/app/src/main.go"); got != "main" {
		t.Errorf("NameWithoutExt = %q", got)
	}
	if got := File.Dir("/app/src/main.go"); got != File.Join("/app", "src") {
		t.Errorf("Dir = %q", got)
	}
}

/* ---------------- HTTP（仅本地函数，不发网络请求） ---------------- */

func TestHttpHelpers(t *testing.T) {
	if got := Http.UrlEncode("a b&c"); got != "a+b%26c" {
		t.Errorf("UrlEncode = %q", got)
	}
	if got := Http.UrlDecode("a+b%26c"); got != "a b&c" {
		t.Errorf("UrlDecode = %q", got)
	}
	if got := Http.UrlEncode(""); got != "" {
		t.Errorf("UrlEncode 空串应返回空串")
	}
}

/* ---------------- 系统 ---------------- */

func TestSys(t *testing.T) {
	if Sys.GetPid() <= 0 {
		t.Error("GetPid 应为正数")
	}
	if Sys.Hostname() == "" {
		t.Error("Hostname 不应为空")
	}
	if Sys.SetEnv("GO_UTILS_TEST", "1") == nil && Sys.GetEnv("GO_UTILS_TEST") != "1" {
		t.Error("环境变量设置后应可读取")
	}
	_ = Sys.UnsetEnv("GO_UTILS_TEST")
}
