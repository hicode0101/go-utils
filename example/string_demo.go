package main

import (
	"fmt"

	"github.com/hicode0101/go-utils/utils"
)

// stringDemo 字符串与文本处理：String / Escape / Html。
func stringDemo() {
	banner("String 字符串工具")

	// 类型转换（xxxE 后缀返回 error，不带后缀失败时返回零值）
	fmt.Println(`ToInt("42") =`, utils.String.ToInt("42"))
	if n, err := utils.String.ToIntE("abc"); err != nil {
		fmt.Println(`ToIntE("abc") ->`, n, "err:", err)
	}
	fmt.Println(`ToFloat64("3.14") =`, utils.String.ToFloat64("3.14"))
	fmt.Println(`ToBool("true") =`, utils.String.ToBool("true"))
	fmt.Println(`IntToStr(42) =`, utils.String.IntToStr(42))
	fmt.Println(`Float64ToStr(3.14159, 2) =`, utils.String.Float64ToStr(3.14159, 2))

	// 判断与修剪
	fmt.Println(`IsEmpty("") =`, utils.String.IsEmpty(""), `| IsBlank("  ") =`, utils.String.IsBlank("  "))
	fmt.Println(`Trim("  hi  ") =`, utils.String.Trim("  hi  "))
	fmt.Println(`Contains("hello", "ell") =`, utils.String.Contains("hello", "ell"),
		`| StartWith =`, utils.String.StartWith("hello", "he"),
		`| EndWith =`, utils.String.EndWith("hello", "lo"))
	fmt.Println(`IsNumber("123") =`, utils.String.IsNumber("123"),
		`| IsEmail =`, utils.String.IsEmail("user@example.com"),
		`| IsMobile =`, utils.String.IsMobile("13612345678"))

	// 截取与拼接
	fmt.Println(`Sub("hello", 1, 3) =`, utils.String.Sub("hello", 1, 3))
	fmt.Println(`SubBefore("a=b", "=", false) =`, utils.String.SubBefore("a=b", "=", false),
		`| SubAfter =`, utils.String.SubAfter("a=b", "=", false))
	fmt.Println(`SubBetween("a[xx]b", "[", "]") =`, utils.String.SubBetween("a[xx]b", "[", "]"))
	fmt.Println(`Split("a,b,c", ",") =`, utils.String.Split("a,b,c", ","),
		`| SplitTrim("a, b , c", ",") =`, utils.String.SplitTrim("a, b , c", ","))
	fmt.Println(`Join([a b c], "-") =`, utils.String.Join([]string{"a", "b", "c"}, "-"))
	fmt.Println(`Concat("go", "-", "utils") =`, utils.String.Concat("go", "-", "utils"))
	fmt.Println(`ConcatAny(1, "-", "a", "-", true) =`, utils.String.ConcatAny(1, "-", "a", "-", true))

	// 替换与大小写
	fmt.Println(`ReplaceAll("a-b-c", "-", "+") =`, utils.String.ReplaceAll("a-b-c", "-", "+"))
	fmt.Println(`ToUpper("go") / UpperFirst("go") =`, utils.String.ToUpper("go"), "/", utils.String.UpperFirst("go"))
	fmt.Println(`ToCamelCase("user_name") =`, utils.String.ToCamelCase("user_name"))
	fmt.Println(`ToUnderlineCase("userName") =`, utils.String.ToUnderlineCase("userName"))
	fmt.Println(`ToKebabCase("userName") =`, utils.String.ToKebabCase("userName"))

	// 补位 / 缩略 / 打码
	fmt.Println(`PadPre("7", 3, '0') =`, utils.String.PadPre("7", 3, '0'))
	fmt.Println(`Center("id", 8, '*') =`, utils.String.Center("id", 8, '*'))
	fmt.Println(`Truncate("hello world", 8, "...") =`, utils.String.Truncate("hello world", 8, "..."))
	fmt.Println(`Brief("一句话很长要缩略", 6) =`, utils.String.Brief("一句话很长要缩略", 6))
	fmt.Println(`Hide("13812345678", 3, 7) =`, utils.String.Hide("13812345678", 3, 7))
	fmt.Println(`Repeat("ab", 3) =`, utils.String.Repeat("ab", 3))

	// 模板与相似度
	fmt.Println(`Format("name={}, age={}", "张三", 25) =`, utils.String.Format("name={}, age={}", "张三", 25))
	fmt.Println(`FormatWithMap("hi {name}", {name:李四}) =`,
		utils.String.FormatWithMap("hi {name}", map[string]string{"name": "李四"}))
	fmt.Println(`Similar("hello", "hallo") =`, utils.String.Similar("hello", "hallo"))
	fmt.Println(`CommonPrefix("flower", "flow") =`, utils.String.CommonPrefix("flower", "flow"))

	// 其它常用
	fmt.Println(`RandomString(8) =`, utils.String.RandomString(8))
	fmt.Println(`FormatBytes(1536000000) =`, utils.String.FormatBytes(1536000000))
	fmt.Println(`ToSBC("abc123") =`, utils.String.ToSBC("abc123"))
	fmt.Println(`Reverse("hello") =`, utils.String.Reverse("hello"))
	fmt.Println(`BlankToDefault("  ", "默认值") =`, utils.String.BlankToDefault("  ", "默认值"))

	banner("Escape 转义工具")
	fmt.Println(`EscapeXml(<a href="x">&</a>) =`, utils.Escape.EscapeXml(`<a href="x">&</a>`))
	fmt.Println(`UnescapeXml("&lt;a&gt;") =`, utils.Escape.UnescapeXml("&lt;a&gt;"))
	fmt.Println(`EscapeHtml4(<div>"go"</div>) =`, utils.Escape.EscapeHtml4(`<div>"go"</div>`))
	fmt.Println(`UnescapeHtml4("&lt;div&gt;&quot;go&quot;&lt;/div&gt;") =`,
		utils.Escape.UnescapeHtml4("&lt;div&gt;&quot;go&quot;&lt;/div&gt;"))
	fmt.Println(`EscapeUnicode("中文") =`, utils.Escape.EscapeUnicode("中文"),
		`| UnescapeUnicode =`, utils.Escape.UnescapeUnicode(utils.Escape.EscapeUnicode("中文")))

	banner("Html HTML 处理")
	fmt.Println(`Escape(<a href="#">x</a>) =`, utils.Html.Escape(`<a href="#">x</a>`))
	fmt.Println(`CleanHtmlTag("<p>hello</p><b>world</b>") =`, utils.Html.CleanHtmlTag("<p>hello</p><b>world</b>"))
	fmt.Println(`RemoveHtmlTag(带内容删 script) =`,
		utils.Html.RemoveHtmlTag("<p>a</p><script>bad()</script>", true, "script"))
	fmt.Println(`RemoveHtmlAttr('<p class="x">y</p>', "class") =`,
		utils.Html.RemoveHtmlAttr(`<p class="x">y</p>`, "class"))
	fmt.Println(`CleanEmptyTag("<p></p>keep<i></i>") =`, utils.Html.CleanEmptyTag("<p></p>keep<i></i>"))
}
