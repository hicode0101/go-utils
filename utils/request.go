package utils

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"
)

// HttpRequest 链式 HTTP 请求构建器（hutool HttpRequest 的 Go 移植）。
// 用法：
//
//	resp := utils.NewGetRequest("https://api.example.com/user").
//	    Header("Authorization", "Bearer xxx").
//	    Timeout(10*time.Second).
//	    Execute()
//
// 各构建方法均返回自身以便链式调用；请求体优先级：显式 Body/JsonBody/TextBody 高于 Form。
type HttpRequest struct {
	httpMethod  string            // 请求方法（GET/POST/...），统一转大写
	reqUrl      string            // 请求地址
	headers     map[string]string // 自定义请求头
	formVals    map[string]string // 表单参数（Execute 时编码为 x-www-form-urlencoded 请求体）
	reqBody     []byte            // 显式请求体（Body/JsonBody/TextBody 设置，优先于表单）
	bodyType    string            // 显式请求体的 Content-Type
	queryParams map[string]string // 追加到 URL 的查询参数（值做 URL 编码）
	timeout     time.Duration     // 请求超时（<=0 时用 DefaultTimeout）
	follow      bool              // 是否跟随重定向（默认 true，与标准 client 行为一致）
}

// NewHttpRequest 创建指定方法与地址的请求构建器，method 为空时默认 GET。
// 场景：方法运行时才确定的通用请求。
func NewHttpRequest(method, reqUrl string) *HttpRequest {
	if String.IsEmpty(method) {
		method = http.MethodGet
	}
	return &HttpRequest{
		httpMethod: strings.ToUpper(method),
		reqUrl:     reqUrl,
		follow:     true,
	}
}

// NewGetRequest 创建 GET 请求构建器。场景：查询接口调用。
func NewGetRequest(reqUrl string) *HttpRequest {
	return NewHttpRequest(http.MethodGet, reqUrl)
}

// NewPostRequest 创建 POST 请求构建器。场景：提交表单或 JSON 数据。
func NewPostRequest(reqUrl string) *HttpRequest {
	return NewHttpRequest(http.MethodPost, reqUrl)
}

// Method 设置（覆盖）请求方法，空值忽略。
func (r *HttpRequest) Method(method string) *HttpRequest {
	if String.IsNotEmpty(method) {
		r.httpMethod = strings.ToUpper(method)
	}
	return r
}

// Header 设置单个请求头（同名覆盖）。场景：设置 Authorization、User-Agent 等。
func (r *HttpRequest) Header(k, v string) *HttpRequest {
	if r.headers == nil {
		r.headers = make(map[string]string)
	}
	r.headers[k] = v
	return r
}

// Headers 批量设置请求头（同名覆盖）。场景：整体传入调用方准备好的头集合。
func (r *HttpRequest) Headers(headers map[string]string) *HttpRequest {
	for k, v := range headers {
		r.Header(k, v)
	}
	return r
}

// BasicAuth 设置 HTTP Basic 认证头。
// 场景：调用需要 Basic 认证的接口（复用 utils.Http.BuildBasicAuth）。
func (r *HttpRequest) BasicAuth(user, pass string) *HttpRequest {
	return r.Header("Authorization", Http.BuildBasicAuth(user, pass))
}

// BearerAuth 设置 HTTP Bearer 认证头。
// 场景：调用 OAuth2/JWT 风格接口（复用 utils.Http.BuildBearerAuth）。
func (r *HttpRequest) BearerAuth(token string) *HttpRequest {
	return r.Header("Authorization", Http.BuildBearerAuth(token))
}

// Form 设置单个表单参数，Execute 时按 application/x-www-form-urlencoded 编码为请求体。
// 场景：提交传统表单风格接口（对应 hutool form）。
func (r *HttpRequest) Form(k, v string) *HttpRequest {
	if r.formVals == nil {
		r.formVals = make(map[string]string)
	}
	r.formVals[k] = v
	return r
}

// FormMap 批量设置表单参数（同名覆盖）。
func (r *HttpRequest) FormMap(m map[string]string) *HttpRequest {
	for k, v := range m {
		r.Form(k, v)
	}
	return r
}

// Body 设置显式请求体与 Content-Type（优先于 Form 表单参数）。
// 场景：上传二进制、发送自定义格式报文。
func (r *HttpRequest) Body(contentType string, data []byte) *HttpRequest {
	r.reqBody = data
	r.bodyType = contentType
	return r
}

// JsonBody 设置 JSON 字符串请求体（自动 application/json）。
// 场景：调用主流 REST 接口（jsonStr 可配合 utils.Json.ToJsonString 生成）。
func (r *HttpRequest) JsonBody(jsonStr string) *HttpRequest {
	return r.Body("application/json", []byte(jsonStr))
}

// TextBody 设置纯文本请求体（Content-Type: text/plain; charset=utf-8）。
// 场景：Webhook、日志上报等文本协议。
func (r *HttpRequest) TextBody(text string) *HttpRequest {
	return r.Body("text/plain; charset=utf-8", []byte(text))
}

// Timeout 设置请求超时时间，d<=0 时使用 DefaultTimeout。
func (r *HttpRequest) Timeout(d time.Duration) *HttpRequest {
	if d <= 0 {
		d = DefaultTimeout
	}
	r.timeout = d
	return r
}

// FollowRedirects 设置是否跟随重定向，默认跟随（与标准 client 行为一致）。
// 传 false 时返回第一个重定向响应（3xx）而不发起二次请求。
// 场景：需要检查 Location 或拦截跳转的爬虫逻辑。
func (r *HttpRequest) FollowRedirects(follow bool) *HttpRequest {
	r.follow = follow
	return r
}

// Cookie 追加单个 Cookie（多次调用拼入同一 Cookie 头，分号分隔）。
// 场景：携带会话 Cookie 请求（对应 hutool cookie）。
func (r *HttpRequest) Cookie(name, value string) *HttpRequest {
	pair := name + "=" + value
	if exist, ok := r.headers["Cookie"]; ok && exist != "" {
		r.headers["Cookie"] = exist + "; " + pair
	} else {
		r.Header("Cookie", pair)
	}
	return r
}

// Query 追加单个 URL 查询参数（值做 URL 编码）。
// 场景：链式拼 GET 查询条件。
func (r *HttpRequest) Query(k, v string) *HttpRequest {
	if r.queryParams == nil {
		r.queryParams = make(map[string]string)
	}
	r.queryParams[k] = v
	return r
}

// QueryMap 批量追加 URL 查询参数（同名覆盖）。
func (r *HttpRequest) QueryMap(m map[string]string) *HttpRequest {
	for k, v := range m {
		r.Query(k, v)
	}
	return r
}

// Execute 执行请求并返回统一响应。复用 utils.Http 的共享连接池；
// FollowRedirects(false) 时改用带 CheckRedirect 的独立 client，不污染共享客户端。
func (r *HttpRequest) Execute() *HttpResp {
	// 1. 拼接 URL 查询参数（键排序、值 URL 编码）
	reqUrl := r.reqUrl
	if len(r.queryParams) > 0 {
		reqUrl = Http.UrlWithForm(reqUrl, r.queryParams)
	}

	// 2. 复制请求头，避免污染构建器状态
	headers := make(map[string]string, len(r.headers)+1)
	for k, v := range r.headers {
		headers[k] = v
	}

	// 3. 组装请求体：显式 Body 优先，其次表单编码
	var body io.Reader
	if len(r.reqBody) > 0 {
		body = strings.NewReader(string(r.reqBody))
		if _, ok := headers["Content-Type"]; !ok && r.bodyType != "" {
			headers["Content-Type"] = r.bodyType
		}
	} else if len(r.formVals) > 0 {
		body = strings.NewReader(Http.ToParams(r.formVals))
		if _, ok := headers["Content-Type"]; !ok {
			headers["Content-Type"] = "application/x-www-form-urlencoded"
		}
	}

	// 4. 执行：默认走共享 client；禁用重定向时用独立 client
	if r.follow {
		return Http.do(r.httpMethod, reqUrl, body, headers, r.timeout)
	}
	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		// ErrUseLastResponse：返回第一个重定向响应，不再发起二次请求
		return http.ErrUseLastResponse
	}}
	return reqbuilderDo(client, r.httpMethod, reqUrl, body, headers, r.timeout)
}

// reqbuilderDo 用指定 client 执行请求（与 utilHttp.do 逻辑一致，仅 client 可替换），
// 供禁用重定向场景使用。timeout<=0 时使用 DefaultTimeout。
func reqbuilderDo(client *http.Client, method, reqUrl string, body io.Reader, headers map[string]string, timeout time.Duration) *HttpResp {
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

	resp, err := client.Do(req)
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
