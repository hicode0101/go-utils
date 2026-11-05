package utils

import (
	"encoding/base64"
	"strings"
	"testing"
)

/* ---------------- 补全与规范化（hutool URLUtil） ---------------- */

func TestUrlCompleteUrl(t *testing.T) {
	// 相对路径基于 base 目录补全（原 TestNetUrl 用例）
	if got, err := Url.CompleteUrl("https://a.com/path/", "img/x.png"); err != nil || got != "https://a.com/path/img/x.png" {
		t.Errorf("CompleteUrl = %q, %v, want \"https://a.com/path/img/x.png\"", got, err)
	}
	// 含 ../ 的相对路径：回退到上级
	if got, err := Url.CompleteUrl("https://a.com/x/y", "../z"); err != nil || got != "https://a.com/z" {
		t.Errorf("CompleteUrl(../z) = %q, %v, want \"https://a.com/z\"", got, err)
	}
	// 以 / 开头的根相对路径：挂到域名根
	if got, err := Url.CompleteUrl("https://a.com/a/b", "/x"); err != nil || got != "https://a.com/x" {
		t.Errorf("CompleteUrl(/x) = %q, %v, want \"https://a.com/x\"", got, err)
	}
	// 绝对地址作为 relative：直接采用
	if got, err := Url.CompleteUrl("http://a.com/x", "https://b.com/y"); err != nil || got != "https://b.com/y" {
		t.Errorf("CompleteUrl(绝对地址) = %q, %v, want \"https://b.com/y\"", got, err)
	}
	// baseUrl 先做 Normalize（缺协议自动补 http://）
	if got, err := Url.CompleteUrl("a.com/path/", "x.png"); err != nil || got != "http://a.com/path/x.png" {
		t.Errorf("CompleteUrl(无协议 base) = %q, %v, want \"http://a.com/path/x.png\"", got, err)
	}
	// 错误路径：base 为空、base 解析失败、relative 解析失败
	if _, err := Url.CompleteUrl("", "x"); err == nil {
		t.Error("CompleteUrl(空 base) 应返回错误")
	}
	if _, err := Url.CompleteUrl("http://a.com/\x7f", "x"); err == nil {
		t.Error("CompleteUrl(非法 base) 应返回错误")
	}
	if _, err := Url.CompleteUrl("https://a.com/", "a\x7fb"); err == nil {
		t.Error("CompleteUrl(非法 relative) 应返回错误")
	}
}

func TestUrlNormalize(t *testing.T) {
	// ../ 消解与多斜杠折叠（原 TestNetUrl 用例）
	if got := Url.Normalize("https://a.com/a/b/../c//d"); got != "https://a.com/a/c/d" {
		t.Errorf("Normalize = %q, want \"https://a.com/a/c/d\"", got)
	}
	// ../ 消解且 query 原样保留
	if got := Url.Normalize("http://a.com/a/../b?x=1"); got != "http://a.com/b?x=1" {
		t.Errorf("Normalize(query 保留) = %q, want \"http://a.com/b?x=1\"", got)
	}
	// 缺协议补 http://、开头多余斜杠、. 与 .. 混合（文档黄金值）
	if got := Url.Normalize("//a.com//x/./y/../z?q=1"); got != "http://a.com/x/z?q=1" {
		t.Errorf("Normalize(文档用例) = %q, want \"http://a.com/x/z?q=1\"", got)
	}
	// 无协议主机路径
	if got := Url.Normalize("a.com/x"); got != "http://a.com/x" {
		t.Errorf("Normalize(无协议) = %q, want \"http://a.com/x\"", got)
	}
	// 反斜杠归一化为 /，自定义协议保留
	if got := Url.Normalize(`ftp://a.com\x\\y`); got != "ftp://a.com/x/y" {
		t.Errorf("Normalize(反斜杠) = %q, want \"ftp://a.com/x/y\"", got)
	}
	// .. 不越过根
	if got := Url.Normalize("http://a.com/../x"); got != "http://a.com/x" {
		t.Errorf("Normalize(..越过根) = %q, want \"http://a.com/x\"", got)
	}
	// 末尾 "." 保留尾部斜杠语义
	if got := Url.Normalize("http://a.com/a/."); got != "http://a.com/a/" {
		t.Errorf("Normalize(末尾.) = %q, want \"http://a.com/a/\"", got)
	}
	// query 中含 fragment 一并原样保留
	if got := Url.Normalize("http://a.com/p/../q?x=1#f"); got != "http://a.com/q?x=1#f" {
		t.Errorf("Normalize(fragment) = %q, want \"http://a.com/q?x=1#f\"", got)
	}
	// 空串原样返回
	if got := Url.Normalize(""); got != "" {
		t.Errorf("Normalize(\"\") = %q, want 空串", got)
	}
}

func TestUrlBuildQuery(t *testing.T) {
	// 多键按字典序排序，输出稳定（原 TestNetUrl 用例）
	q := Url.BuildQuery(map[string]string{"b": "1", "a": "2"})
	if q != "a=2&b=1" {
		t.Errorf("BuildQuery = %q, want \"a=2&b=1\"", q)
	}
	if again := Url.BuildQuery(map[string]string{"b": "1", "a": "2"}); again != q {
		t.Errorf("BuildQuery 输出不稳定: %q vs %q", q, again)
	}
	// 值含特殊字符被编码：& = 空格（转 +）中文
	if got := Url.BuildQuery(map[string]string{"b": "x 1", "a": "&"}); got != "a=%26&b=x+1" {
		t.Errorf("BuildQuery(特殊字符) = %q, want \"a=%%26&b=x+1\"", got)
	}
	if got := Url.BuildQuery(map[string]string{"": "忽略空键", "k": "中文"}); got != "k=%E4%B8%AD%E6%96%87" {
		t.Errorf("BuildQuery(空键/中文) = %q, want \"k=%%E4%%B8%%AD%%E6%%96%%87\"", got)
	}
	// 空 map → 空串
	if got := Url.BuildQuery(map[string]string{}); got != "" {
		t.Errorf("BuildQuery(空map) = %q, want 空串", got)
	}
	if got := Url.BuildQuery(nil); got != "" {
		t.Errorf("BuildQuery(nil) = %q, want 空串", got)
	}
}

/* ---------------- 编码与 DataURI ---------------- */

func TestUrlEncodeBlank(t *testing.T) {
	if got := Url.EncodeBlank("https://a.com/a b.html"); got != "https://a.com/a%20b.html" {
		t.Errorf("EncodeBlank = %q, want \"https://a.com/a%%20b.html\"", got)
	}
	// 无空白原样返回
	if got := Url.EncodeBlank("https://a.com/a.html"); got != "https://a.com/a.html" {
		t.Errorf("EncodeBlank(无空格) = %q, want 原样", got)
	}
	// 制表符等空白同样转 %20
	if got := Url.EncodeBlank("a\tb"); got != "a%20b" {
		t.Errorf("EncodeBlank(制表符) = %q, want \"a%%20b\"", got)
	}
	// 已有的 %XX 不做二次编码
	if got := Url.EncodeBlank("a%20b c"); got != "a%20b%20c" {
		t.Errorf("EncodeBlank(已有%%XX) = %q, want \"a%%20b%%20c\"", got)
	}
}

func TestUrlDataUri(t *testing.T) {
	// 前缀断言（原 TestNetUrl 用例）
	got := Url.DataUri("image/png", []byte("x"))
	if !strings.HasPrefix(got, "data:image/png;base64,") {
		t.Errorf("DataUri = %q, 缺少 data:image/png;base64, 前缀", got)
	}
	// mime 为空 → 默认 application/octet-stream；base64 部分可解码回原文
	got2 := Url.DataUri("", []byte("hi"))
	if got2 != "data:application/octet-stream;base64,aGk=" {
		t.Errorf("DataUri(空mime) = %q, want \"data:application/octet-stream;base64,aGk=\"", got2)
	}
	got3 := Url.DataUri("text/plain", []byte("hello 中文"))
	const prefix3 = "data:text/plain;base64,"
	if !strings.HasPrefix(got3, prefix3) {
		t.Fatalf("DataUri(中文) = %q, 缺少 %s 前缀", got3, prefix3)
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(got3, prefix3))
	if err != nil || string(decoded) != "hello 中文" {
		t.Errorf("DataUri base64 解码回原文失败: %q, %v", decoded, err)
	}
}

func TestUrlDataUriText(t *testing.T) {
	const prefix = "data:text/plain;charset=utf-8,"
	// 纯 ASCII 原样
	if got := Url.DataUriText("hi"); got != prefix+"hi" {
		t.Errorf("DataUriText(\"hi\") = %q, want %q", got, prefix+"hi")
	}
	// 中文按路径规则百分号编码，且可经 Url.Decode 还原
	if got := Url.DataUriText("你好"); got != prefix+"%E4%BD%A0%E5%A5%BD" {
		t.Errorf("DataUriText(中文) = %q, want %q", got, prefix+"%E4%BD%A0%E5%A5%BD")
	}
	if back := Url.Decode(strings.TrimPrefix(Url.DataUriText("你好"), prefix)); back != "你好" {
		t.Errorf("DataUriText 中文还原 = %q, want \"你好\"", back)
	}
	// 空格转 %20
	if got := Url.DataUriText("a b"); got != prefix+"a%20b" {
		t.Errorf("DataUriText(空格) = %q, want %q", got, prefix+"a%20b")
	}
}

/* ---------------- 解码与提取 ---------------- */

func TestUrlDecode(t *testing.T) {
	// 路径语义：%20 还原空格，"+" 保留为加号（原 TestNetUrl 用例）
	if got := Url.Decode("a%20b+c"); got != "a b+c" {
		t.Errorf("Url.Decode 路径语义 + 不转空格: %q, want \"a b+c\"", got)
	}
	if got := Url.Decode("%20"); got != " " {
		t.Errorf("Url.Decode(\"%%20\") = %q, want 空格", got)
	}
	if got := Url.Decode("%2B"); got != "+" {
		t.Errorf("Url.Decode(\"%%2B\") = %q, want \"+\"", got)
	}
	if got := Url.Decode("a%41b"); got != "aAb" {
		t.Errorf("Url.Decode(\"a%%41b\") = %q, want \"aAb\"", got)
	}
	// 非法 % 序列原样保留：非十六进制、末尾孤立 %、不完整的 %X
	for _, in := range []string{"%zz", "100%", "a%2"} {
		if got := Url.Decode(in); got != in {
			t.Errorf("Url.Decode(%q) = %q, 非法 %% 序列应原样保留", in, got)
		}
	}
	// 不含 % 的输入走快速路径原样返回
	if got := Url.Decode("abc"); got != "abc" {
		t.Errorf("Url.Decode(无%%) = %q, want 原样", got)
	}
	// 与 Http.UrlEncode 对照往返：查询串语义（空格转 +）→ 归一化为 %20 → 路径解码还原原文
	src := "a b&c=中?x"
	enc := strings.ReplaceAll(Http.UrlEncode(src), "+", "%20")
	if got := Url.Decode(enc); got != src {
		t.Errorf("UrlDecode 往返失败: %q -> %q -> %q", src, enc, got)
	}
	// 与 Http.UrlDecode 的语义对照：Decode 保留 +，UrlDecode 把 + 转空格
	if Url.Decode("a+b") != "a+b" || Http.UrlDecode("a+b") != "a b" {
		t.Error("Url.Decode 与 Http.UrlDecode 对 + 的处理应有路径/查询串语义差异")
	}
}

func TestUrlGetHost(t *testing.T) {
	// 带端口
	if got := Url.GetHost("https://a.com:8080/x"); got != "a.com:8080" {
		t.Errorf("GetHost(带端口) = %q, want \"a.com:8080\"", got)
	}
	// 不带端口
	if got := Url.GetHost("https://a.com/x"); got != "a.com" {
		t.Errorf("GetHost(不带端口) = %q, want \"a.com\"", got)
	}
	if got := Url.GetHost("https://a.com"); got != "a.com" {
		t.Errorf("GetHost(仅域名) = %q, want \"a.com\"", got)
	}
	// 缺协议：补 // 重试解析
	if got := Url.GetHost("a.com/x"); got != "a.com" {
		t.Errorf("GetHost(无协议) = %q, want \"a.com\"", got)
	}
	// 非法输入返回空串：空串、含控制字符两次解析均失败
	if got := Url.GetHost(""); got != "" {
		t.Errorf("GetHost(\"\") = %q, want 空串", got)
	}
	if got := Url.GetHost("\x7f"); got != "" {
		t.Errorf("GetHost(控制字符) = %q, want 空串", got)
	}
}
