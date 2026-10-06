package utils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// DefaultTimeout HTTP 请求默认超时时间。调用方传 0 或负数时使用该值。
const DefaultTimeout = 30 * time.Second

// utilHttp HTTP 客户端工具。
//
// 所有方法统一返回 *HttpResp（错误信息记录在 Msg 字段，不为 nil），
// 请求超时 time<=0 时使用 DefaultTimeout；底层复用同一连接池。
type utilHttp struct{}

// HttpResp HTTP 响应统一封装。
type HttpResp struct {
	Status  int         // HTTP 状态码；请求未能发出时为 -1（网络错误）或 -500（本地构造请求失败）
	Msg     string      // 错误信息，成功时为空串
	Body    []byte      // 响应体
	Headers http.Header // 响应头（HEAD 请求的结果主要取这里）
}

// BodyString 把响应体转为字符串（空响应体返回空串）。
// 场景：接口返回 JSON/文本时直接取字符串。
func (r *HttpResp) BodyString() string {
	if len(r.Body) == 0 {
		return ""
	}
	return string(r.Body)
}

// IsOK 状态码是否为 2xx 成功。场景：快速判断请求是否业务可用。
func (r *HttpResp) IsOK() bool {
	return r.Status >= 200 && r.Status < 300
}

// httpClient 包级共享客户端，复用 TCP/TLS 连接池，避免每次请求重建连接。
var httpClient = &http.Client{}

// do 统一执行请求：设置超时（context 实现，不破坏连接复用）、
// 附加请求头、读取响应体并关闭。
func (h *utilHttp) do(method, reqUrl string, body io.Reader, headers map[string]string, timeout time.Duration) *HttpResp {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, reqUrl, body)
	if err != nil {
		return &HttpResp{Status: -500, Msg: "构造请求失败: " + err.Error()}
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return &HttpResp{Status: -1, Msg: "请求失败: " + err.Error()}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &HttpResp{Status: resp.StatusCode, Msg: "读取响应失败: " + err.Error(), Headers: resp.Header}
	}
	return &HttpResp{Status: resp.StatusCode, Body: respBody, Headers: resp.Header}
}

// Get 发送 GET 请求。headers 可为 nil（不带自定义头）。
// 场景：调用第三方查询接口，响应体在 resp.Body，文本用 resp.BodyString()。
func (h *utilHttp) Get(reqUrl string, headers map[string]string, timeout time.Duration) *HttpResp {
	return h.do(http.MethodGet, reqUrl, nil, headers, timeout)
}

// Head 发送 HEAD 请求，仅取响应头（资源大小、是否存在等），响应体为空。
// 场景：下载前探测文件大小：resp.Headers.Get("Content-Length")。
func (h *utilHttp) Head(reqUrl string, headers map[string]string, timeout time.Duration) *HttpResp {
	return h.do(http.MethodHead, reqUrl, nil, headers, timeout)
}

// PostForm 发送表单 POST 请求（application/x-www-form-urlencoded，
// 键值自动做 URL 编码）。form、headers 均可为 nil。
// 场景：提交登录表单、调用传统表单风格接口。
func (h *utilHttp) PostForm(reqUrl string, form map[string]string, headers map[string]string, timeout time.Duration) *HttpResp {
	vals := url.Values{}
	for k, v := range form {
		vals.Set(k, v)
	}
	// 复制 headers 并补充默认 Content-Type，调用方显式传入时以调用方为准
	merged := make(map[string]string, len(headers)+1)
	for k, v := range headers {
		merged[k] = v
	}
	if _, ok := merged["Content-Type"]; !ok {
		merged["Content-Type"] = "application/x-www-form-urlencoded"
	}
	return h.do(http.MethodPost, reqUrl, strings.NewReader(vals.Encode()), merged, timeout)
}

// PostJson 发送 JSON POST 请求（Content-Type: application/json）。
// jsonBody 为已序列化的 JSON 字符串（可配合 utils.Json.ToJsonString 生成）。
// 场景：调用主流 REST 接口。
func (h *utilHttp) PostJson(reqUrl, jsonBody string, headers map[string]string, timeout time.Duration) *HttpResp {
	merged := make(map[string]string, len(headers)+1)
	for k, v := range headers {
		merged[k] = v
	}
	if _, ok := merged["Content-Type"]; !ok {
		merged["Content-Type"] = "application/json"
	}
	return h.do(http.MethodPost, reqUrl, strings.NewReader(jsonBody), merged, timeout)
}

// Download 流式下载文件到 savePath（不占用大量内存），headers 可为 nil。
// 状态码非 2xx 时返回错误且不落盘。场景：下载安装包、抓取图片素材。
func (h *utilHttp) Download(reqUrl, savePath string, headers map[string]string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl, nil)
	if err != nil {
		return fmt.Errorf("构造请求失败: %w", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("下载 %s 失败，状态码 %d", reqUrl, resp.StatusCode)
	}

	out, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer out.Close()

	if _, err = io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}
	return nil
}

// UrlEncode 对字符串做 URL 查询参数编码（空格转 +）。
// UrlEncode("a b&c") = "a+b%26c"。场景：手工拼查询串时对参数值转义。
func (h *utilHttp) UrlEncode(input string) string {
	if String.IsEmpty(input) {
		return ""
	}
	return url.QueryEscape(input)
}

// UrlDecode 解码 URL 编码字符串，非法编码原样返回。
// UrlDecode("a+b%26c") = "a b&c"。
func (h *utilHttp) UrlDecode(input string) string {
	if String.IsEmpty(input) {
		return ""
	}
	result, err := url.QueryUnescape(input)
	if err != nil {
		return input
	}
	return result
}
