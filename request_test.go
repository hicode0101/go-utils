package utils

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newReqBuilderSrv 按需创建本地测试服务器（httptest，监听 127.0.0.1），
// 测试结束自动关闭，不访问外网。仅供 request_test.go 使用。
func newReqBuilderSrv(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

// reqEchoHandler 回显请求要素：method/path/query/关注请求头/原始请求体，
// 供各构建器用例断言请求内容。仅供 request_test.go 使用。
func reqEchoHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, "method=%s\n", r.Method)
	_, _ = fmt.Fprintf(w, "path=%s\n", r.URL.Path)
	_, _ = fmt.Fprintf(w, "query=%s\n", r.URL.RawQuery)
	for _, h := range []string{"X-Custom", "X-Other", "Authorization", "Cookie", "Content-Type"} {
		_, _ = fmt.Fprintf(w, "H[%s]=%s\n", h, r.Header.Get(h))
	}
	b, _ := io.ReadAll(r.Body)
	_, _ = fmt.Fprintf(w, "body=%s\n", b)
}

// TestRequestBuilderCtor 覆盖 NewGetRequest/NewPostRequest/NewHttpRequest 的 method 与 url。
func TestRequestBuilderCtor(t *testing.T) {
	srv := newReqBuilderSrv(t, reqEchoHandler)

	// NewGetRequest：method=GET 且 url 路径正确
	resp := NewGetRequest(srv.URL + "/get").Execute()
	if resp.Status != 200 || !resp.IsOK() {
		t.Fatalf("NewGetRequest status = %d, msg = %q", resp.Status, resp.Msg)
	}
	body := resp.BodyString()
	if !strings.Contains(body, "method=GET") || !strings.Contains(body, "path=/get") {
		t.Errorf("GET 构建器 method/url 错误: %q", body)
	}

	// NewPostRequest
	resp = NewPostRequest(srv.URL + "/post").Execute()
	body = resp.BodyString()
	if !strings.Contains(body, "method=POST") || !strings.Contains(body, "path=/post") {
		t.Errorf("POST 构建器 method/url 错误: %q", body)
	}

	// NewHttpRequest：method 统一转大写
	resp = NewHttpRequest("patch", srv.URL+"/any").Execute()
	if !strings.Contains(resp.BodyString(), "method=PATCH") {
		t.Errorf("NewHttpRequest 应将 method 转大写: %q", resp.BodyString())
	}

	// NewHttpRequest：空 method 默认 GET
	resp = NewHttpRequest("", srv.URL+"/def").Execute()
	if !strings.Contains(resp.BodyString(), "method=GET") {
		t.Errorf("空 method 应默认 GET: %q", resp.BodyString())
	}
}

// TestRequestMethod 覆盖 Method：链中覆盖请求方法、空值忽略。
func TestRequestMethod(t *testing.T) {
	srv := newReqBuilderSrv(t, reqEchoHandler)

	resp := NewGetRequest(srv.URL + "/m").Method("delete").Execute()
	if !strings.Contains(resp.BodyString(), "method=DELETE") {
		t.Errorf("Method 应覆盖请求方法并转大写: %q", resp.BodyString())
	}

	resp = NewGetRequest(srv.URL + "/m").Method("").Execute()
	if !strings.Contains(resp.BodyString(), "method=GET") {
		t.Errorf("Method 空值应被忽略: %q", resp.BodyString())
	}
}

// TestRequestHeader 覆盖 Header/Headers：服务端回显断言。
func TestRequestHeader(t *testing.T) {
	srv := newReqBuilderSrv(t, reqEchoHandler)

	// 单个 Header
	resp := NewGetRequest(srv.URL+"/h").Header("X-Custom", "v1").Execute()
	if !strings.Contains(resp.BodyString(), "H[X-Custom]=v1") {
		t.Errorf("Header 未送达服务器: %q", resp.BodyString())
	}

	// Headers 批量设置（同名覆盖）
	resp = NewGetRequest(srv.URL+"/h").
		Header("X-Custom", "old").
		Headers(map[string]string{"X-Custom": "v2", "X-Other": "o1"}).
		Execute()
	body := resp.BodyString()
	if !strings.Contains(body, "H[X-Custom]=v2") || !strings.Contains(body, "H[X-Other]=o1") {
		t.Errorf("Headers 批量设置错误: %q", body)
	}
}

// TestRequestAuth 覆盖 BasicAuth/BearerAuth：服务端收到 Authorization 头。
func TestRequestAuth(t *testing.T) {
	srv := newReqBuilderSrv(t, reqEchoHandler)

	resp := NewGetRequest(srv.URL+"/auth").BasicAuth("user", "pass").Execute()
	if !strings.Contains(resp.BodyString(), "H[Authorization]=Basic dXNlcjpwYXNz") {
		t.Errorf("BasicAuth 头错误: %q", resp.BodyString())
	}

	resp = NewGetRequest(srv.URL + "/auth").BearerAuth("tk-1").Execute()
	if !strings.Contains(resp.BodyString(), "H[Authorization]=Bearer tk-1") {
		t.Errorf("BearerAuth 头错误: %q", resp.BodyString())
	}
}

// TestRequestCookie 覆盖 Cookie：多次调用拼入同一 Cookie 头（分号分隔），服务端收到。
func TestRequestCookie(t *testing.T) {
	srv := newReqBuilderSrv(t, reqEchoHandler)

	resp := NewGetRequest(srv.URL+"/c").Cookie("a", "1").Cookie("b", "2").Execute()
	if !strings.Contains(resp.BodyString(), "H[Cookie]=a=1; b=2") {
		t.Errorf("Cookie 拼接错误: %q", resp.BodyString())
	}
}

// TestRequestFormAndQuery 覆盖 Form/FormMap（POST 表单服务端 ParseForm 比对）
// 与 Query/QueryMap（GET 查询串服务端解析比对）。
func TestRequestFormAndQuery(t *testing.T) {
	srv := newReqBuilderSrv(t, func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		_, _ = fmt.Fprintf(w, "CT=%s\n", r.Header.Get("Content-Type"))
		_, _ = fmt.Fprintf(w, "postform=%s\n", r.PostForm.Encode())
		_, _ = fmt.Fprintf(w, "query=%s\n", r.URL.RawQuery)
		_, _ = fmt.Fprintf(w, "q_name=%s\n", r.URL.Query().Get("name"))
		_, _ = fmt.Fprintf(w, "q_cn=%s\n", r.URL.Query().Get("cn"))
	})

	// POST 表单：Form + FormMap，服务端 r.ParseForm 后比对
	resp := NewPostRequest(srv.URL+"/submit").
		Form("name", "张三").
		FormMap(map[string]string{"x": "a b", "y": "1"}).
		Execute()
	if resp.Status != 200 {
		t.Fatalf("Form POST status = %d, msg = %q", resp.Status, resp.Msg)
	}
	body := resp.BodyString()
	if !strings.Contains(body, "CT=application/x-www-form-urlencoded") {
		t.Errorf("表单默认 Content-Type 错误: %q", body)
	}
	if !strings.Contains(body, "name=%E5%BC%A0%E4%B8%89") ||
		!strings.Contains(body, "x=a+b") || !strings.Contains(body, "y=1") {
		t.Errorf("POST 表单服务端解析回显错误: %q", body)
	}

	// GET 查询参数：Query + QueryMap，键排序且值 URL 编码，服务端解码比对
	resp = NewGetRequest(srv.URL+"/search").
		Query("name", "v 1").
		QueryMap(map[string]string{"cn": "中文"}).
		Execute()
	if resp.Status != 200 {
		t.Fatalf("Query GET status = %d, msg = %q", resp.Status, resp.Msg)
	}
	body = resp.BodyString()
	if !strings.Contains(body, "q_name=v 1") || !strings.Contains(body, "q_cn=中文") {
		t.Errorf("GET 查询参数服务端解析错误: %q", body)
	}
	if !strings.Contains(body, "query=cn=%E4%B8%AD%E6%96%87&name=v+1") {
		t.Errorf("查询串编码错误: %q", body)
	}
}

// TestRequestBodyTypes 覆盖 Body/JsonBody/TextBody：Content-Type 与原始 body 回显、
// 显式 Body 优先于 Form。
func TestRequestBodyTypes(t *testing.T) {
	srv := newReqBuilderSrv(t, reqEchoHandler)

	// Body：自定义 Content-Type 与原始 body
	resp := NewPostRequest(srv.URL+"/body").
		Body("text/custom", []byte("raw-body-123")).
		Execute()
	body := resp.BodyString()
	if !strings.Contains(body, "H[Content-Type]=text/custom") || !strings.Contains(body, "body=raw-body-123") {
		t.Errorf("Body 回显错误: %q", body)
	}

	// JsonBody：自动 application/json
	resp = NewPostRequest(srv.URL + "/body").JsonBody(`{"k":1}`).Execute()
	body = resp.BodyString()
	if !strings.Contains(body, "H[Content-Type]=application/json") || !strings.Contains(body, `body={"k":1}`) {
		t.Errorf("JsonBody 回显错误: %q", body)
	}

	// TextBody：text/plain; charset=utf-8
	resp = NewPostRequest(srv.URL + "/body").TextBody("hello text").Execute()
	body = resp.BodyString()
	if !strings.Contains(body, "H[Content-Type]=text/plain; charset=utf-8") || !strings.Contains(body, "body=hello text") {
		t.Errorf("TextBody 回显错误: %q", body)
	}

	// 显式 Body 优先于 Form 表单参数
	resp = NewPostRequest(srv.URL+"/body").Form("a", "1").JsonBody(`{"b":2}`).Execute()
	body = resp.BodyString()
	if !strings.Contains(body, `body={"b":2}`) || strings.Contains(body, "a=1") {
		t.Errorf("显式 Body 应优先于 Form: %q", body)
	}
}

// TestRequestTimeout 覆盖 Timeout：极短超时 + 慢 handler → 失败返回；
// Timeout(0) 回退默认超时，正常请求应成功。
func TestRequestTimeout(t *testing.T) {
	srv := newReqBuilderSrv(t, func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		_, _ = w.Write([]byte("slow"))
	})

	resp := NewGetRequest(srv.URL + "/slow").Timeout(50 * time.Millisecond).Execute()
	if resp.Status >= 0 || resp.Msg == "" {
		t.Errorf("超时应失败返回, got status=%d msg=%q", resp.Status, resp.Msg)
	}

	resp = NewGetRequest(srv.URL + "/slow").Timeout(0).Execute()
	if resp.Status != 200 {
		t.Errorf("Timeout(0) 应使用默认超时并成功, got %d %q", resp.Status, resp.Msg)
	}
}

// TestRequestFollowRedirects 覆盖 FollowRedirects：false 返回 302，true（含默认）跟随到 200。
func TestRequestFollowRedirects(t *testing.T) {
	srv := newReqBuilderSrv(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/echo?from=redirect", http.StatusFound)
			return
		}
		reqEchoHandler(w, r)
	})

	// 默认行为（不显式调用）：跟随重定向 → 最终 200
	resp := NewGetRequest(srv.URL + "/redirect").Execute()
	if resp.Status != 200 {
		t.Errorf("默认应跟随重定向到 200, got %d %q", resp.Status, resp.Msg)
	}

	// 显式 FollowRedirects(true)：跟随到最终 200
	resp = NewGetRequest(srv.URL + "/redirect").FollowRedirects(true).Execute()
	if resp.Status != 200 || !strings.Contains(resp.BodyString(), "path=/echo") {
		t.Errorf("FollowRedirects(true) 应跟随到最终 200: %d %q", resp.Status, resp.BodyString())
	}

	// FollowRedirects(false)：返回第一个 302 响应，不再二次请求
	resp = NewGetRequest(srv.URL + "/redirect").FollowRedirects(false).Execute()
	if resp.Status != 302 {
		t.Errorf("FollowRedirects(false) 应返回 302, got %d", resp.Status)
	}
	if loc := resp.Headers.Get("Location"); !strings.HasSuffix(loc, "/echo?from=redirect") {
		t.Errorf("Location 头错误: %q", loc)
	}
}

// TestRequestExecuteInvalidUrl 覆盖 Execute 对非法 URL 返回 Status<0。
func TestRequestExecuteInvalidUrl(t *testing.T) {
	resp := NewGetRequest("://bad").Execute()
	if resp.Status != -500 {
		t.Errorf("非法 URL 应返回 Status=-500, got %d", resp.Status)
	}
	if resp.Msg == "" {
		t.Error("非法 URL 应有错误信息")
	}

	// 拼接 Query 后仍为非法 URL
	resp = NewGetRequest("://bad").Query("k", "v").Execute()
	if resp.Status != -500 {
		t.Errorf("拼接查询后仍应返回 -500, got %d", resp.Status)
	}
}
