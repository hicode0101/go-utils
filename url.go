package utils

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// utilUrl URL 工具。
//
// 覆盖相对地址补全、规范化、查询串构建、DataURI 生成与路径级解码
// （对应 hutool URLUtil 的可移植子集）。
//
// 与 utils.Http.UrlEncode/UrlDecode 的分工：那两个面向查询串
// （application/x-www-form-urlencoded 语义，空格转 "+"）；
// 本文件的 Decode 面向 URL 路径（"+" 保留为加号，仅还原 %XX）。
type utilUrl struct{}

/* ---------------- 补全与规范化 ---------------- */

// CompleteUrl 相对地址补全为绝对地址（对应 hutool completeUrl）。
// baseUrl 先做 Normalize 规范化，再基于其解析相对地址：
// CompleteUrl("https://a.com/x/y", "../z") = "https://a.com/z"。
// baseUrl 为空或任一地址无法解析时返回 error。场景：爬虫把页面里的相对链接转绝对链接。
func (u *utilUrl) CompleteUrl(baseUrl, relative string) (string, error) {
	baseUrl = u.Normalize(baseUrl)
	if strings.TrimSpace(baseUrl) == "" {
		return "", fmt.Errorf("基准 URL 为空")
	}
	base, err := url.Parse(baseUrl)
	if err != nil {
		return "", fmt.Errorf("解析基准 URL %q 失败: %w", baseUrl, err)
	}
	ref, err := url.Parse(relative)
	if err != nil {
		return "", fmt.Errorf("解析相对地址 %q 失败: %w", relative, err)
	}
	return base.ResolveReference(ref).String(), nil
}

// Normalize 规范化 URL 字符串（对照 hutool normalize(url, false, false) 默认行为）：
//  1. 缺协议时自动补 "http://" 头；
//  2. 去除开头的 "\" 或 "/"，并把 "\" 替换为 "/"；
//  3. 折叠路径中的连续斜杠，并消解 "." 与 ".." 段（绝对路径不越过根）；
//  4. query 部分（首个 "?" 起的全部内容，含可能的 "#fragment"）原样保留。
//
// Normalize("//a.com//x/./y/../z?q=1") = "http://a.com/x/z?q=1"。
// 场景：清洗用户手填网址、爬虫入口 URL 预处理。
func (u *utilUrl) Normalize(urlStr string) string {
	if strings.TrimSpace(urlStr) == "" {
		return urlStr
	}
	// 拆协议：以首个 "://"（下标须大于 0）为界，无协议默认补 http://（同 hutool）
	protocol := "http://"
	body := urlStr
	if sepIndex := strings.Index(urlStr, "://"); sepIndex > 0 {
		protocol = urlStr[:sepIndex+3]
		body = urlStr[sepIndex+3:]
	}
	// 拆 query：首个 "?"（下标须大于 0）之后的内容原样保留
	params := ""
	if qIndex := strings.Index(body, "?"); qIndex > 0 {
		params = body[qIndex:]
		body = body[:qIndex]
	}
	// 去除开头的 \ 或 /，再把 \ 替换为 /
	body = strings.TrimLeft(body, `\/`)
	body = strings.ReplaceAll(body, `\`, `/`)
	// 拆 host 与 path，仅对 path 折叠斜杠并消解 . / .. 段
	domain := body
	path := ""
	if pathSepIndex := strings.Index(body, "/"); pathSepIndex >= 0 {
		domain = body[:pathSepIndex]
		path = body[pathSepIndex:]
		for strings.Contains(path, "//") {
			path = strings.ReplaceAll(path, "//", "/")
		}
		path = urlutilResolveDotSegments(path)
	}
	return protocol + domain + path + params
}

// BuildQuery 把 map 拼接为查询串（对应 hutool buildQuery）。
// 键按字典序排序保证同一 map 输出稳定；键与值均做查询串编码（空格转 "+"，
// 同 utils.Http.UrlEncode）；空键（""）的项忽略。
// BuildQuery(map{"b":"2","a":"x 1"}) = "a=x+1&b=2"。
// 场景：组装 GET 查询串、签名前参数规范化。
func (u *utilUrl) BuildQuery(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "" {
			// 对应 hutool：空键忽略
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, url.QueryEscape(k)+"="+url.QueryEscape(params[k]))
	}
	return strings.Join(pairs, "&")
}

/* ---------------- 编码与 DataURI ---------------- */

// EncodeBlank 单独编码 URL 中的空白符为 %20（对应 hutool encodeBlank），
// 其余字符（含中文、已有的 %XX）原样保留，不做二次编码。
// EncodeBlank("https://a.com/a b.html") = "https://a.com/a%20b.html"。
// 场景：URL 中空格导致请求异常时的预清洗（hutool toUrlForHttp 即先做此步）。
func (u *utilUrl) EncodeBlank(urlStr string) string {
	if !strings.ContainsAny(urlStr, " \t\n\v\f\r") {
		return urlStr
	}
	var b strings.Builder
	b.Grow(len(urlStr))
	for i := 0; i < len(urlStr); i++ {
		switch urlStr[i] {
		case ' ', '\t', '\n', '\v', '\f', '\r':
			b.WriteString("%20")
		default:
			b.WriteByte(urlStr[i])
		}
	}
	return b.String()
}

// DataUri 生成 base64 编码的 data URI，格式 "data:<mime>;base64,<数据>"
// （对应 hutool getDataUriBase64）；mimeType 为空时默认 application/octet-stream。
// DataUri("image/png", pngBytes) = "data:image/png;base64,..."。
// 场景：把小图片/小文件内联进 HTML、CSS，省一次外部请求。
func (u *utilUrl) DataUri(mimeType string, data []byte) string {
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	return "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(data)
}

// DataUriText 生成文本 data URI："data:text/plain;charset=utf-8,<URL 编码后的文本>"，
// 文本按路径规则做百分号编码（空格转 %20，中文转 %XX）。
// 场景：内联短文本、把一段话变成可点击的 data 链接。
func (u *utilUrl) DataUriText(text string) string {
	return "data:text/plain;charset=utf-8," + url.PathEscape(text)
}

/* ---------------- 解码与提取 ---------------- */

// Decode URL 路径解码：还原 %XX（按字节还原，配合 UTF-8 文本使用），"+" 保留为加号
// （对照 hutool URLUtil.decode 默认行为）。非法的 % 序列（如 "%zz"、末尾孤立 "%"）原样保留。
// 注意与 utils.Http.UrlDecode 的区别：UrlDecode 面向查询串，会把 "+" 转为空格；
// Decode("a%2Bb%20c+d") = "a+b c+d"（加号未被转成空格）。
// 场景：解码 URL 路径、还原爬虫抓到的资源地址。
func (u *utilUrl) Decode(content string) string {
	if !strings.Contains(content, "%") {
		return content
	}
	var b strings.Builder
	b.Grow(len(content))
	for i := 0; i < len(content); i++ {
		c := content[i]
		if c == '%' && i+2 < len(content) {
			hi := urlutilHexDigit(content[i+1])
			if hi >= 0 {
				lo := urlutilHexDigit(content[i+2])
				if lo >= 0 {
					b.WriteByte(byte(hi<<4 | lo))
					i += 2
					continue
				}
			}
		}
		// 普通字符与不符合规范的 % 形式原样输出（对应 hutool"跳过不符合规范的%形式"）
		b.WriteByte(c)
	}
	return b.String()
}

// GetHost 取 URL 的 host 部分，带端口时连同端口一起返回：
// GetHost("https://a.com:8080/x") = "a.com:8080"。
// 输入缺协议时（如 "a.com/x"）补 "//" 重试解析；解析失败返回空串。
// 场景：提取接口域名做白名单校验、按站点归并统计。
func (u *utilUrl) GetHost(urlStr string) string {
	urlStr = strings.TrimSpace(urlStr)
	parsed, err := url.Parse(urlStr)
	if err != nil {
		parsed = nil
	}
	if parsed == nil || (parsed.Host == "" && parsed.Scheme == "") {
		// "a.com/x"、"a.com:8080/x" 这类无协议输入直接解析会失败或被当作路径，补 "//" 重试
		parsed, err = url.Parse("//" + urlStr)
		if err != nil {
			return ""
		}
	}
	return parsed.Host
}

/* ---------------- 内部辅助（urlutil 前缀，避免与其它文件冲突） ---------------- */

// urlutilResolveDotSegments 消解路径中的 "." 与 ".." 段（RFC 3986 remove_dot_segments 简化实现）：
// 绝对路径不越过根（"/../a" -> "/a"），末尾的 "." / ".." 保留尾部斜杠语义（"/a/." -> "/a/"）
func urlutilResolveDotSegments(path string) string {
	// 允许 ".." 弹出的下界：绝对路径的首段是根的前导空段，不可被弹出
	base := 0
	if strings.HasPrefix(path, "/") {
		base = 1
	}
	segs := strings.Split(path, "/")
	stack := make([]string, 0, len(segs))
	for i, seg := range segs {
		switch seg {
		case ".":
			if i == len(segs)-1 {
				stack = append(stack, "")
			}
		case "..":
			if len(stack) > base {
				stack = stack[:len(stack)-1]
			}
			if i == len(segs)-1 {
				stack = append(stack, "")
			}
		default:
			stack = append(stack, seg)
		}
	}
	return strings.Join(stack, "/")
}

// urlutilHexDigit 十六进制字符转数值，非法字符返回 -1
func urlutilHexDigit(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	default:
		return -1
	}
}
