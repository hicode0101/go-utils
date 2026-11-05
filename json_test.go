package utils

import (
	"strings"
	"testing"
)

type demoUser struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type jsonDept struct {
	Name    string     `json:"name"`
	Members []demoUser `json:"members"`
	Tags    []string   `json:"tags"`
}

/* ---------------- 原有基础断言（自 utils_test.go 原样搬入） ---------------- */

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

/* ---------------- ToJson / FromJson 往返 ---------------- */

func TestJsonRoundTrip(t *testing.T) {
	u := demoUser{Name: "李雷", Age: 20}
	data, err := Json.ToJson(u)
	if err != nil {
		t.Fatalf("ToJson 失败: %v", err)
	}
	var back demoUser
	if err := Json.FromJson(data, &back); err != nil {
		t.Fatalf("FromJson 失败: %v", err)
	}
	if back != u {
		t.Errorf("往返不一致: %+v, want %+v", back, u)
	}
}

/* ---------------- ToJsonString / FromJson（字符串入参） ---------------- */

func TestJsonStringForms(t *testing.T) {
	s := Json.ToJsonString(demoUser{Name: "Tom", Age: 18})
	if s != `{"name":"Tom","age":18}` {
		t.Errorf("ToJsonString = %q", s)
	}
	var back demoUser
	if err := Json.FromJson([]byte(s), &back); err != nil {
		t.Fatalf("FromJson(字符串序列化结果) 失败: %v", err)
	}
	if back.Name != "Tom" || back.Age != 18 {
		t.Errorf("FromJson 结果错误: %+v", back)
	}
}

/* ---------------- 带缩进序列化 ---------------- */

func TestJsonPretty(t *testing.T) {
	u := demoUser{Name: "Tom", Age: 18}
	want := "{\n\t\"name\": \"Tom\",\n\t\"age\": 18\n}"
	pretty, err := Json.ToPrettyJson(u)
	if err != nil {
		t.Fatalf("ToPrettyJson 失败: %v", err)
	}
	if string(pretty) != want {
		t.Errorf("ToPrettyJson = %q, want %q", pretty, want)
	}
	got := Json.ToPrettyJsonString(u)
	if got != want {
		t.Errorf("ToPrettyJsonString = %q, want %q", got, want)
	}
	if !strings.Contains(got, "\n\t") {
		t.Error("ToPrettyJsonString 应使用制表符缩进换行")
	}
}

/* ---------------- 错误路径 ---------------- */

func TestJsonInvalid(t *testing.T) {
	var u demoUser
	if err := Json.FromJson([]byte(`{bad json`), &u); err == nil {
		t.Error("FromJson 非法输入应返回错误")
	}
	if u.Name != "" || u.Age != 0 {
		t.Errorf("非法输入不应写入目标对象: %+v", u)
	}
	// 不支持的类型（channel 无法序列化）
	if _, err := Json.ToJson(make(chan int)); err == nil {
		t.Error("ToJson 不支持类型应返回错误")
	}
	if _, err := Json.ToPrettyJson(make(chan int)); err == nil {
		t.Error("ToPrettyJson 不支持类型应返回错误")
	}
	if got := Json.ToJsonString(make(chan int)); got != "" {
		t.Errorf("ToJsonString 失败应返回空串, got %q", got)
	}
	if got := Json.ToPrettyJsonString(make(chan int)); got != "" {
		t.Errorf("ToPrettyJsonString 失败应返回空串, got %q", got)
	}
}

/* ---------------- Valid ---------------- */

func TestJsonValid(t *testing.T) {
	if !Json.Valid([]byte(`{"a":[1,2],"b":"x"}`)) {
		t.Error("合法 JSON 应通过 Valid")
	}
	if !Json.Valid([]byte(`null`)) {
		t.Error("null 应为合法 JSON")
	}
	if !Json.Valid([]byte(`[1,2,3]`)) {
		t.Error("顶层数组应为合法 JSON")
	}
	if Json.Valid([]byte(`{a:}`)) {
		t.Error("非法 JSON 不应通过 Valid")
	}
	if Json.Valid([]byte(``)) {
		t.Error("空串不应通过 Valid")
	}
	if Json.Valid([]byte(`{"a":1,}`)) {
		t.Error("尾逗号不应通过 Valid")
	}
}

/* ---------------- 嵌套结构体与切片 ---------------- */

func TestJsonNested(t *testing.T) {
	d := jsonDept{
		Name:    "研发部",
		Members: []demoUser{{Name: "A", Age: 1}, {Name: "B", Age: 2}},
		Tags:    []string{"go", "util"},
	}
	data, err := Json.ToJson(d)
	if err != nil {
		t.Fatalf("ToJson 嵌套结构失败: %v", err)
	}
	var back jsonDept
	if err := Json.FromJson(data, &back); err != nil {
		t.Fatalf("FromJson 嵌套结构失败: %v", err)
	}
	if back.Name != "研发部" || len(back.Members) != 2 || back.Members[1].Name != "B" || len(back.Tags) != 2 || back.Tags[1] != "util" {
		t.Errorf("嵌套往返不一致: %+v", back)
	}
	// 切片元素级往返
	users := []demoUser{{Name: "X", Age: 9}}
	data2, err := Json.ToJson(users)
	if err != nil {
		t.Fatalf("ToJson 切片失败: %v", err)
	}
	var users2 []demoUser
	if err := Json.FromJson(data2, &users2); err != nil {
		t.Fatalf("FromJson 切片失败: %v", err)
	}
	if len(users2) != 1 || users2[0] != users[0] {
		t.Errorf("切片往返不一致: %+v", users2)
	}
}
