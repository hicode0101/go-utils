package main

import (
	"fmt"
	"time"

	"github.com/hicode0101/go-utils/utils"
)

// httpDemo HTTP 客户端与请求构建：Http / Request / UserAgent。
func httpDemo() {
	banner("Http HTTP 客户端（URL 编解码与工具方法，无网络依赖）")
	fmt.Println(`UrlEncode("a b&c=中") =`, utils.Http.UrlEncode("a b&c=中"),
		"| UrlDecode =", utils.Http.UrlDecode(utils.Http.UrlEncode("a b&c=中")))
	fmt.Println(`BuildBasicAuth("user", "pass") =`, utils.Http.BuildBasicAuth("user", "pass"))
	fmt.Println(`BuildBearerAuth("my-token") =`, utils.Http.BuildBearerAuth("my-token"))
	fmt.Println(`IsHttp("http://a.com") =`, utils.Http.IsHttp("http://a.com"),
		`| IsHttps("https://a.com") =`, utils.Http.IsHttps("https://a.com"))
	fmt.Println(`ToParams({"name":"张三","page":"1"}) =`,
		utils.Http.ToParams(map[string]string{"name": "张三", "page": "1"}))
	fmt.Println(`UrlWithForm("http://api.example.com/users", {"name":"李四"}) =`,
		utils.Http.UrlWithForm("http://api.example.com/users", map[string]string{"name": "李四"}))
	fmt.Println(`DecodeParamMap("a=1&b=中文") =`, utils.Http.DecodeParamMap("a=1&b=中文"))
	fmt.Println(`GetCharsetFromContentType("application/json; charset=utf-8") =`,
		utils.Http.GetCharsetFromContentType("application/json; charset=utf-8"))
	fmt.Println(`SniffContentType(`+"`{\"a\":1}`"+`) =`, utils.Http.SniffContentType(`{"a":1}`))

	// 网络请求：失败只打印错误，不影响其它示例
	banner("Http 发起请求（需外网）")
	resp := utils.Http.Get("https://example.com", nil, 10*time.Second)
	if resp.IsOK() {
		fmt.Println("Get(example.com) Status =", resp.Status, "| Body 前 48 字符:",
			utils.String.Brief(resp.BodyString(), 48))
	} else {
		fmt.Println("Get(example.com) 失败: Status =", resp.Status, "Msg =", resp.Msg)
	}
	// 其余请求方式：
	//   utils.Http.PostForm(url, map[string]string{"name": "张三"}, nil, 0)   // 表单提交
	//   utils.Http.PostJson(url, `{"name":"张三"}`, nil, 0)                   // JSON 提交
	//   utils.Http.GetForStr(url)                                            // 直接取响应体字符串
	//   utils.Http.Download(url, "save.dat", nil, 0)                         // 下载到文件
	//   utils.Http.DownloadBytes(url)                                        // 下载到内存

	banner("Request 链式请求构建器")
	resp2 := utils.NewGetRequest("https://example.com").
		Header("Accept", "text/html").
		Query("q", "go-utils").
		Timeout(10 * time.Second).
		Execute()
	if resp2.IsOK() {
		fmt.Println("NewGetRequest(...).Execute() Status =", resp2.Status, "| IsOK =", resp2.IsOK())
	} else {
		fmt.Println("Request Execute 失败: Status =", resp2.Status, "Msg =", resp2.Msg)
	}
	// 更多链式方法：Method / Headers / BasicAuth / BearerAuth / Form / FormMap /
	//              Body / JsonBody / TextBody / Cookie / QueryMap / FollowRedirects

	banner("UserAgent UA 解析")
	uaStr := "Mozilla/5.0 (iPhone; CPU iPhone OS 16_6 like Mac OS X) AppleWebKit/605.1.15 " +
		"(KHTML, like Gecko) Version/16.6 Mobile/15E148 Safari/604.1"
	ua := utils.UserAgent.Parse(uaStr)
	fmt.Printf("Parse(iphone-ua) -> Browser=%s %s | Engine=%s | Os=%s | Platform=%s | Mobile=%v\n",
		ua.Browser, ua.BrowserVersion, ua.Engine, ua.Os, ua.Platform, ua.Mobile)
	fmt.Println("IsMobile(iphone-ua) =", utils.UserAgent.IsMobile(uaStr))
}
