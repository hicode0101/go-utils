package utils

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// newHttpUtilSrv 按需创建本地测试服务器（httptest，监听 127.0.0.1），
// 测试结束自动关闭，不访问外网。仅供 http_test.go 使用。
func newHttpUtilSrv(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

// TestHttpGet 覆盖 utilHttp.Get：状态码、Body、自定义头回显、headers nil、超时、非法 URL。
func TestHttpGet(t *testing.T) {
	srv := newHttpUtilSrv(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/slow" {
			time.Sleep(200 * time.Millisecond)
		}
		w.Header().Set("Echo-X-Test", r.Header.Get("X-Test"))
		_, _ = w.Write([]byte("get-ok"))
	})

	// 常规 GET：状态码 + Body + 自定义请求头回显
	resp := Http.Get(srv.URL+"/data", map[string]string{"X-Test": "abc"}, 0)
	if resp.Status != 200 || !resp.IsOK() {
		t.Fatalf("Get status = %d, msg = %q", resp.Status, resp.Msg)
	}
	if resp.BodyString() != "get-ok" {
		t.Errorf("Get body = %q", resp.BodyString())
	}
	if resp.Headers.Get("Echo-X-Test") != "abc" {
		t.Errorf("自定义头未送达服务器: %q", resp.Headers.Get("Echo-X-Test"))
	}

	// headers 为 nil 路径
	if resp := Http.Get(srv.URL+"/data", nil, 0); resp.Status != 200 {
		t.Errorf("Get(nil headers) status = %d, msg = %q", resp.Status, resp.Msg)
	}

	// 超时路径：handler 睡 200ms，timeout 50ms → Status == -1 且 Msg 非空
	resp = Http.Get(srv.URL+"/slow", nil, 50*time.Millisecond)
	if resp.Status != -1 || resp.Msg == "" {
		t.Errorf("超时应返回 Status=-1 且 Msg 非空, got %d, %q", resp.Status, resp.Msg)
	}

	// 非法 URL → Status == -500
	resp = Http.Get("://bad", nil, 0)
	if resp.Status != -500 || resp.Msg == "" {
		t.Errorf("非法 URL 应返回 Status=-500, got %d, %q", resp.Status, resp.Msg)
	}
}

// TestHttpHead 覆盖 utilHttp.Head：仅取响应头、响应体为空。
func TestHttpHead(t *testing.T) {
	srv := newHttpUtilSrv(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "123")
		w.Header().Set("X-Probe", "ok")
		w.WriteHeader(http.StatusOK)
	})

	resp := Http.Head(srv.URL+"/file.bin", nil, 0)
	if resp.Status != 200 {
		t.Fatalf("Head status = %d, msg = %q", resp.Status, resp.Msg)
	}
	if resp.Headers.Get("X-Probe") != "ok" {
		t.Errorf("Head 应取到响应头: %v", resp.Headers)
	}
	if len(resp.Body) != 0 {
		t.Errorf("HEAD 响应体应为空, got %d bytes", len(resp.Body))
	}

	// 带自定义头路径
	if resp := Http.Head(srv.URL+"/file.bin", map[string]string{"X-Test": "h"}, 0); resp.Status != 200 {
		t.Errorf("Head(带头) status = %d", resp.Status)
	}
}

// TestHttpPostForm 覆盖 utilHttp.PostForm：默认 Content-Type、表单编码回显、
// 显式 Content-Type 优先、form nil。
func TestHttpPostForm(t *testing.T) {
	srv := newHttpUtilSrv(t, func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		_ = r.ParseForm()
		_, _ = fmt.Fprintf(w, "CT=%s|FORM=%s", ct, r.PostForm.Encode())
	})
	form := map[string]string{"name": "张三", "x": "a b&c"}

	resp := Http.PostForm(srv.URL+"/login", form, nil, 0)
	if resp.Status != 200 {
		t.Fatalf("PostForm status = %d, msg = %q", resp.Status, resp.Msg)
	}
	body := resp.BodyString()
	if !strings.Contains(body, "CT=application/x-www-form-urlencoded") {
		t.Errorf("默认 Content-Type 错误: %q", body)
	}
	if !strings.Contains(body, "name=%E5%BC%A0%E4%B8%89") || !strings.Contains(body, "x=a+b%26c") {
		t.Errorf("表单 URL 编码/服务端解析回显错误: %q", body)
	}

	// 调用方显式传入 Content-Type 时以调用方为准
	resp = Http.PostForm(srv.URL+"/login", form, map[string]string{"Content-Type": "application/json"}, 0)
	if !strings.Contains(resp.BodyString(), "CT=application/json") {
		t.Errorf("显式 Content-Type 未生效: %q", resp.BodyString())
	}

	// form 为 nil
	if resp := Http.PostForm(srv.URL+"/login", nil, nil, 0); resp.Status != 200 {
		t.Errorf("PostForm(nil form) status = %d", resp.Status)
	}
}

// TestHttpPostJson 覆盖 utilHttp.PostJson：默认 Content-Type、原始 body 回显、显式头优先。
func TestHttpPostJson(t *testing.T) {
	srv := newHttpUtilSrv(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_, _ = fmt.Fprintf(w, "CT=%s|BODY=%s", r.Header.Get("Content-Type"), b)
	})

	resp := Http.PostJson(srv.URL+"/api", `{"a":1}`, nil, 0)
	if resp.Status != 200 {
		t.Fatalf("PostJson status = %d, msg = %q", resp.Status, resp.Msg)
	}
	body := resp.BodyString()
	if !strings.Contains(body, "CT=application/json") || !strings.Contains(body, `BODY={"a":1}`) {
		t.Errorf("PostJson 回显错误: %q", body)
	}

	// 显式 Content-Type 优先
	resp = Http.PostJson(srv.URL+"/api", "plain", map[string]string{"Content-Type": "text/plain"}, 0)
	if !strings.Contains(resp.BodyString(), "CT=text/plain") {
		t.Errorf("显式 Content-Type 未生效: %q", resp.BodyString())
	}

	// 带自定义头
	if resp := Http.PostJson(srv.URL+"/api", `{"a":1}`, map[string]string{"X-Token": "t1"}, 0); resp.Status != 200 {
		t.Errorf("PostJson(带头) status = %d", resp.Status)
	}
}

// TestHttpDownload 覆盖 utilHttp.Download：成功落盘内容一致、404 报错、非法 URL 报错。
func TestHttpDownload(t *testing.T) {
	content := "download-file-content-12345"
	srv := newHttpUtilSrv(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/missing" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(content))
	})

	// 成功：文件内容一致
	savePath := filepath.Join(t.TempDir(), "a.bin")
	if err := Http.Download(srv.URL+"/file.bin", savePath, nil, 0); err != nil {
		t.Fatalf("Download 失败: %v", err)
	}
	got, err := os.ReadFile(savePath)
	if err != nil || string(got) != content {
		t.Errorf("下载文件内容不一致: %q, %v", got, err)
	}

	// 404 → 返回错误
	if err := Http.Download(srv.URL+"/missing", filepath.Join(t.TempDir(), "b.bin"), nil, 0); err == nil {
		t.Error("404 下载应返回错误")
	}

	// 非法 URL → 返回错误
	if err := Http.Download("://bad", filepath.Join(t.TempDir(), "c.bin"), nil, 0); err == nil {
		t.Error("非法 URL 下载应返回错误")
	}

	// 带自定义头下载
	if err := Http.Download(srv.URL+"/file.bin", filepath.Join(t.TempDir(), "d.bin"), map[string]string{"X-Test": "d"}, 0); err != nil {
		t.Errorf("Download(带头) 失败: %v", err)
	}
}

// TestHttpUrlEncodeDecode 覆盖 UrlEncode/UrlDecode（含空串、非法编码）。
func TestHttpUrlEncodeDecode(t *testing.T) {
	if got := Http.UrlEncode("a b&c"); got != "a+b%26c" {
		t.Errorf("UrlEncode = %q, want a+b%%26c", got)
	}
	if got := Http.UrlEncode(""); got != "" {
		t.Errorf("UrlEncode 空串应返回空串, got %q", got)
	}
	if got := Http.UrlDecode("a+b%26c"); got != "a b&c" {
		t.Errorf("UrlDecode = %q, want %q", got, "a b&c")
	}
	if got := Http.UrlDecode(""); got != "" {
		t.Errorf("UrlDecode 空串应返回空串, got %q", got)
	}
	// 非法编码原样返回
	if got := Http.UrlDecode("%zz"); got != "%zz" {
		t.Errorf("UrlDecode 非法编码应原样返回, got %q", got)
	}
}

// TestHttpGetPostForStr 覆盖 GetForStr/PostForStr：成功取 body、404 返回空串。
func TestHttpGetPostForStr(t *testing.T) {
	srv := newHttpUtilSrv(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/missing" {
			http.NotFound(w, r)
			return
		}
		if r.Method == http.MethodPost {
			_ = r.ParseForm()
			_, _ = fmt.Fprintf(w, "ok:%s", r.PostForm.Get("k"))
			return
		}
		_, _ = w.Write([]byte("plain-text-body"))
	})

	if got := Http.GetForStr(srv.URL + "/text"); got != "plain-text-body" {
		t.Errorf("GetForStr = %q", got)
	}
	if got := Http.GetForStr(srv.URL + "/missing"); got != "" {
		t.Errorf("GetForStr 404 应返回空串, got %q", got)
	}
	if got := Http.PostForStr(srv.URL+"/submit", map[string]string{"k": "v1"}); got != "ok:v1" {
		t.Errorf("PostForStr = %q", got)
	}
	if got := Http.PostForStr(srv.URL+"/missing", map[string]string{"k": "v1"}); got != "" {
		t.Errorf("PostForStr 404 应返回空串, got %q", got)
	}
}

// TestHttpDownloadBytes 覆盖 DownloadBytes：字节数与内容一致、非 2xx 与非法 URL 报错。
func TestHttpDownloadBytes(t *testing.T) {
	content := []byte("bytes-content-0123456789")
	srv := newHttpUtilSrv(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/missing" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(content)
	})

	got, err := Http.DownloadBytes(srv.URL + "/data")
	if err != nil {
		t.Fatalf("DownloadBytes 失败: %v", err)
	}
	if len(got) != len(content) || string(got) != string(content) {
		t.Errorf("DownloadBytes 字节不一致: %d bytes %q, want %d bytes", len(got), got, len(content))
	}

	if b, err := Http.DownloadBytes(srv.URL + "/missing"); err == nil || b != nil {
		t.Errorf("404 应返回错误, got %q, %v", b, err)
	}
	if b, err := Http.DownloadBytes("://bad"); err == nil || b != nil {
		t.Errorf("非法 URL 应返回错误, got %q, %v", b, err)
	}
}

// TestHttpAuthBuilders 覆盖 BuildBasicAuth/BuildBearerAuth 黄金值。
func TestHttpAuthBuilders(t *testing.T) {
	if got := Http.BuildBasicAuth("user", "pass"); got != "Basic dXNlcjpwYXNz" {
		t.Errorf("BuildBasicAuth = %q", got)
	}
	if got := Http.BuildBasicAuth("admin", "123"); got != "Basic YWRtaW46MTIz" {
		t.Errorf("BuildBasicAuth(admin) = %q", got)
	}
	if got := Http.BuildBearerAuth("token123"); got != "Bearer token123" {
		t.Errorf("BuildBearerAuth = %q", got)
	}
}

// TestHttpProtocolChecks 覆盖 IsHttp/IsHttps：大小写混合与其它协议。
func TestHttpProtocolChecks(t *testing.T) {
	if !Http.IsHttp("http://a.com") || !Http.IsHttp("HTTP://A.COM") || !Http.IsHttp("Http://a.com") {
		t.Error("IsHttp 应忽略大小写识别 http:")
	}
	if Http.IsHttp("https://a.com") || Http.IsHttp("ftp://a.com") {
		t.Error("IsHttp 不应匹配其它协议")
	}
	if !Http.IsHttps("https://a.com") || !Http.IsHttps("HTTPS://a.com") || !Http.IsHttps("HttpS://a.com") {
		t.Error("IsHttps 应忽略大小写识别 https:")
	}
	if Http.IsHttps("http://a.com") || Http.IsHttps("ftp://a.com") {
		t.Error("IsHttps 不应匹配其它协议")
	}
}

// TestHttpParams 覆盖 ToParams（排序+编码）、UrlWithForm（含已有 ? 与末尾 &）、
// DecodeParamMap（同键取后值、? 前缀容错、空串）。
func TestHttpParams(t *testing.T) {
	if got := Http.ToParams(map[string]string{"b": "x y", "a": "1"}); got != "a=1&b=x+y" {
		t.Errorf("ToParams = %q, want a=1&b=x+y", got)
	}
	if got := Http.ToParams(nil); got != "" {
		t.Errorf("ToParams(nil) 应为空串, got %q", got)
	}

	if got := Http.UrlWithForm("http://a.com/api", map[string]string{"k": "v"}); got != "http://a.com/api?k=v" {
		t.Errorf("UrlWithForm = %q", got)
	}
	if got := Http.UrlWithForm("http://a.com/api?x=1", map[string]string{"k": "v"}); got != "http://a.com/api?x=1&k=v" {
		t.Errorf("UrlWithForm(已有 ?) = %q", got)
	}
	if got := Http.UrlWithForm("http://a.com/api?x=1&", map[string]string{"k": "v"}); got != "http://a.com/api?x=1&k=v" {
		t.Errorf("UrlWithForm(末尾 &) = %q", got)
	}
	if got := Http.UrlWithForm("http://a.com/api", nil); got != "http://a.com/api" {
		t.Errorf("UrlWithForm(空表单) = %q", got)
	}

	m := Http.DecodeParamMap("a=1&b=2&a=3")
	if m["a"] != "3" || m["b"] != "2" {
		t.Errorf("DecodeParamMap 同键应取后值: %v", m)
	}
	m = Http.DecodeParamMap("?k=%E4%B8%AD")
	if m["k"] != "中" {
		t.Errorf("DecodeParamMap(? 前缀容错+URL 解码) = %v", m)
	}
	if m = Http.DecodeParamMap(""); len(m) != 0 {
		t.Errorf("DecodeParamMap(空串) 应为空 map: %v", m)
	}
}

// TestHttpCharsetSniff 覆盖 GetCharsetFromContentType（有/无 charset/空串）
// 与 SniffContentType（{、[、<、其它、空白）。
func TestHttpCharsetSniff(t *testing.T) {
	if got := Http.GetCharsetFromContentType("text/html; charset=UTF-8"); got != "UTF-8" {
		t.Errorf("GetCharsetFromContentType = %q, want UTF-8", got)
	}
	if got := Http.GetCharsetFromContentType("application/json;charset=utf-8"); got != "utf-8" {
		t.Errorf("GetCharsetFromContentType(无空格) = %q, want utf-8", got)
	}
	if got := Http.GetCharsetFromContentType("text/html"); got != "" {
		t.Errorf("无 charset 应返回空串, got %q", got)
	}
	if got := Http.GetCharsetFromContentType(""); got != "" {
		t.Errorf("空 Content-Type 应返回空串, got %q", got)
	}

	if got := Http.SniffContentType(`{"a":1}`); got != "application/json" {
		t.Errorf("SniffContentType({) = %q", got)
	}
	if got := Http.SniffContentType(`[1,2]`); got != "application/json" {
		t.Errorf("SniffContentType([) = %q", got)
	}
	if got := Http.SniffContentType("<xml/>"); got != "application/xml" {
		t.Errorf("SniffContentType(<) = %q", got)
	}
	if got := Http.SniffContentType("a=1&b=2"); got != "application/x-www-form-urlencoded" {
		t.Errorf("SniffContentType(表单) = %q", got)
	}
	if got := Http.SniffContentType("   "); got != "application/x-www-form-urlencoded" {
		t.Errorf("SniffContentType(空白) = %q", got)
	}
}
