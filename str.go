package utils

import (
	"crypto/rand"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// utilStr 字符串工具。
//
// 命名约定：ToInt/ToIntE 成对出现，带 E 后缀的返回 error（安全转换），
// 不带 E 的失败时返回零值（便捷转换），按需选用。
type utilStr struct{}

/* ---------------- 类型转换 ---------------- */

// ToInt 将字符串转为 int，转换失败返回 0。
// 场景：解析确定格式或允许默认值的数字字符串；需要感知失败时用 ToIntE。
func (s *utilStr) ToInt(str string) int {
	n, _ := strconv.Atoi(str)
	return n
}

// ToIntE 同 ToInt，但返回错误，适合需要区分"值为 0"和"解析失败"的场景。
func (s *utilStr) ToIntE(str string) (int, error) {
	return strconv.Atoi(str)
}

// ToInt64 将字符串转为 int64，转换失败返回 0。
// 场景：解析大数值，如订单金额（单位分）、雪花 ID。
func (s *utilStr) ToInt64(str string) int64 {
	n, _ := strconv.ParseInt(str, 10, 64)
	return n
}

// ToInt64E 同 ToInt64，但返回错误。
func (s *utilStr) ToInt64E(str string) (int64, error) {
	return strconv.ParseInt(str, 10, 64)
}

// ToFloat64 将字符串转为 float64，转换失败返回 0。
// 场景：解析接口返回的价格、百分比等浮点数字符串。
func (s *utilStr) ToFloat64(str string) float64 {
	n, _ := strconv.ParseFloat(str, 64)
	return n
}

// ToFloat64E 同 ToFloat64，但返回错误。
func (s *utilStr) ToFloat64E(str string) (float64, error) {
	return strconv.ParseFloat(str, 64)
}

// ToBool 将字符串转为 bool，转换失败返回 false。
// 支持 strconv.ParseBool 的全部取值：1/t/T/TRUE/true/True/0/f/F/FALSE/false/False。
// 场景：解析配置项、URL 参数中的开关。
func (s *utilStr) ToBool(str string) bool {
	b, _ := strconv.ParseBool(str)
	return b
}

// ToBoolE 同 ToBool，但返回错误。
func (s *utilStr) ToBoolE(str string) (bool, error) {
	return strconv.ParseBool(str)
}

// IntToStr int 转字符串。场景：拼接 URL 参数、组装报文字段。
func (s *utilStr) IntToStr(num int) string {
	return strconv.Itoa(num)
}

// Int64ToStr int64 转字符串。场景：把 int64 主键 ID 转成字符串响应给前端。
func (s *utilStr) Int64ToStr(num int64) string {
	return strconv.FormatInt(num, 10)
}

// Float64ToStr 浮点数转字符串。
// prec 为小数位数：2 表示保留两位（3.141 -> "3.14"）；-1 表示用能完整还原数值的最少位数。
// 场景：展示金额时 prec=2；无损序列化时 prec=-1。
func (s *utilStr) Float64ToStr(num float64, prec int) string {
	return strconv.FormatFloat(num, 'f', prec, 64)
}

/* ---------------- 二进制与位运算 ---------------- */

// BinToInt 二进制字符串转 int64，如 "1010" -> 10，非法输入返回 0。
// 场景：解析协议报文或数据库中以二进制串存储的标志位。
func (s *utilStr) BinToInt(binStr string) int64 {
	n, _ := strconv.ParseInt(binStr, 2, 64)
	return n
}

// IntToBin int64 转二进制字符串，如 10 -> "1010"（负数带 - 前缀）。
// 场景：把权限位/状态位以可读形式落库或打印。
func (s *utilStr) IntToBin(num int64) string {
	return strconv.FormatInt(num, 2)
}

// HasBits 判断 value 的二进制位是否包含 mask 的全部位，即 (value & mask) == mask。
// 场景：权限校验，如 value=0b1011、mask=0b0011 时返回 true。
func (s *utilStr) HasBits(value, mask int64) bool {
	return value&mask == mask
}

// BinStrHasBits 先把二进制字符串转为数值，再判断是否包含 mask 的全部位。
// 场景：标志位以二进制字符串形式存储时的权限校验。
func (s *utilStr) BinStrHasBits(binStr string, mask int64) bool {
	return s.HasBits(s.BinToInt(binStr), mask)
}

/* ---------------- 判断与检查 ---------------- */

// IsEmpty 长度为 0（仅判断空串 ""）。
// 场景：入参校验。
func (s *utilStr) IsEmpty(str string) bool {
	return len(str) == 0
}

// IsNotEmpty 非 IsEmpty。
func (s *utilStr) IsNotEmpty(str string) bool {
	return len(str) > 0
}

// IsBlank 去除首尾空白后为空：" "、"\t\n" 均视为空。
// 场景：校验用户输入的表单字段是否真正有内容。
func (s *utilStr) IsBlank(str string) bool {
	return strings.TrimSpace(str) == ""
}

// IsNotBlank 非 IsBlank。
func (s *utilStr) IsNotBlank(str string) bool {
	return !s.IsBlank(str)
}

// TrimSpace 去除首尾空白字符。场景：清洗用户输入、粘贴内容。
func (s *utilStr) TrimSpace(str string) string {
	return strings.TrimSpace(str)
}

// Len 字节长度（中文等多字节字符按 UTF-8 字节数计）。
// 场景：限制数据库字段字节长度、报文定长校验；需要字符个数用 RuneCount。
func (s *utilStr) Len(str string) int {
	return len(str)
}

// RuneCount 字符（rune）个数，"中国" -> 2。
// 场景：昵称长度限制、按字数截断展示。
func (s *utilStr) RuneCount(str string) int {
	return utf8.RuneCountInString(str)
}

/* ---------------- 查找与处理 ---------------- */

// Contains 是否包含子串。场景：敏感词粗筛、关键字过滤。
func (s *utilStr) Contains(str, find string) bool {
	return strings.Contains(str, find)
}

// StartsWith 是否以 find 开头。场景：判断协议前缀，如 "https://"。
func (s *utilStr) StartsWith(str, find string) bool {
	return strings.HasPrefix(str, find)
}

// EndsWith 是否以 find 结尾。场景：判断文件扩展名。
func (s *utilStr) EndsWith(str, find string) bool {
	return strings.HasSuffix(str, find)
}

// Count 子串 find 在 str 中出现的次数（非重叠计数）。
// strings.Count("cheese", "e") = 3。场景：统计分隔符个数推算字段数。
func (s *utilStr) Count(str, find string) int {
	return strings.Count(str, find)
}

// Index 返回 find 首次出现的下标，未找到返回 -1。
// strings.Index("ABC_xyz", "xyz") = 4。
func (s *utilStr) Index(str, find string) int {
	return strings.Index(str, find)
}

// LastIndex 返回 find 最后一次出现的下标，未找到返回 -1。
// 场景：取路径最后一个分隔符位置，提取文件名。
func (s *utilStr) LastIndex(str, find string) int {
	return strings.LastIndex(str, find)
}

// Replace 只替换第一处 find。场景：仅修正首个错误标记。
func (s *utilStr) Replace(str, find, to string) string {
	return strings.Replace(str, find, to, 1)
}

// ReplaceAll 替换全部 find。
// 场景：批量清洗数据中的分隔符、去除全部换行符。
func (s *utilStr) ReplaceAll(str, find, to string) string {
	return strings.ReplaceAll(str, find, to)
}

// Split 按分隔符拆分为切片。注意：对空串会返回 [""]。
// 场景：解析 "a,b,c" 逗号分隔参数。
func (s *utilStr) Split(str, sep string) []string {
	return strings.Split(str, sep)
}

// Join 用分隔符连接字符串切片。strings.Join([]string{"a","b"}, ",") = "a,b"。
// 场景：拼 SQL IN 条件、日志输出数组。
func (s *utilStr) Join(strs []string, sep string) string {
	return strings.Join(strs, sep)
}

// SubStr 按"字符"截取子串 [start, end)，自动收敛越界下标，start>=end 返回空串。
// 场景：标题摘要展示。注意与旧版的区别：本方法按字符而非字节截取，
// "你好世界" 截取 [0,2) 得 "你好"（旧版按字节会截出乱码）。
func (s *utilStr) SubStr(str string, start, end int) string {
	runes := []rune(str)
	n := len(runes)
	if start < 0 {
		start = 0
	}
	if end > n {
		end = n
	}
	if start >= end {
		return ""
	}
	return string(runes[start:end])
}

// ToLower 字母转小写。"Love GoLang" -> "love golang"。
// 场景：统一邮箱大小写后再比对。
func (s *utilStr) ToLower(str string) string {
	return strings.ToLower(str)
}

// ToUpper 字母转大写。"love 中国" -> "LOVE 中国"。
// 场景：统一状态码、货币代码大小写。
func (s *utilStr) ToUpper(str string) string {
	return strings.ToUpper(str)
}

// Reverse 反转字符串（按字符处理，中文安全）。
// "中国abc" -> "cba国中"。场景：简单混淆/校验回文。
func (s *utilStr) Reverse(str string) string {
	runes := []rune(str)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Repeat 重复 str 共 count 次。场景：生成分隔线 strings.Repeat("=", 30)。
func (s *utilStr) Repeat(str string, count int) string {
	if count <= 0 {
		return ""
	}
	return strings.Repeat(str, count)
}

// Truncate 按字符数截断：超长时保留前 maxLen 个字符并追加 suffix 结尾，未超长原样返回。
// Truncate("你好世界欢迎", 5, "...") = "你好世..."。
// 场景：列表页摘要、日志脱敏截断。
func (s *utilStr) Truncate(str string, maxLen int, suffix string) string {
	if maxLen < 0 || utf8.RuneCountInString(str) <= maxLen {
		return str
	}
	runes := []rune(str)
	return string(runes[:maxLen]) + suffix
}

// PadLeft 左侧填充 pad 到总长 length，超长原样返回。
// PadLeft("7", 3, '0') = "007"。场景：补齐定长编号。
func (s *utilStr) PadLeft(str string, length int, pad rune) string {
	n := utf8.RuneCountInString(str)
	if n >= length {
		return str
	}
	return strings.Repeat(string(pad), length-n) + str
}

// PadRight 右侧填充 pad 到总长 length。
// PadRight("ab", 4, '*') = "ab**"。场景：表格对齐输出。
func (s *utilStr) PadRight(str string, length int, pad rune) string {
	n := utf8.RuneCountInString(str)
	if n >= length {
		return str
	}
	return str + strings.Repeat(string(pad), length-n)
}

/* ---------------- 拼接 ---------------- */

// Concat 高效拼接多个字符串（strings.Builder 实现）。
// 场景：已知全部为字符串时的拼接，优于 + 号反复分配。
func (s *utilStr) Concat(strs ...string) string {
	var b strings.Builder
	for _, v := range strs {
		b.WriteString(v)
	}
	return b.String()
}

// ConcatAny 拼接任意类型（字符串原样写入，其它类型 fmt.Sprint 转换）。
// ConcatAny("id=", 1001, " ok=", true) = "id=1001 ok=true"。
func (s *utilStr) ConcatAny(inputs ...interface{}) string {
	var b strings.Builder
	for _, v := range inputs {
		if str, ok := v.(string); ok {
			b.WriteString(str)
		} else {
			b.WriteString(fmt.Sprint(v))
		}
	}
	return b.String()
}

// FormatBytes 把字节数格式化为人类可读单位。
// FormatBytes(1536) = "1.50 KB"，FormatBytes(5*1024*1024) = "5.00 MB"。
// 场景：展示文件大小、内存占用。
func (s *utilStr) FormatBytes(size uint64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

/* ---------------- 随机字符串 ---------------- */

// 随机字符串字符集：数字 + 大小写字母
const randomCharset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// RandomString 生成长度为 length 的随机字符串（数字 + 字母，密码学安全随机源，
// 用拒绝采样保证每个字符等概率）。length<=0 或随机源异常时返回空串。
// 场景：验证码、临时文件名、请求 ID。注意：用于高安全场景请配合过期与一次性校验。
func (s *utilStr) RandomString(length int) string {
	if length <= 0 {
		return ""
	}
	// 256 中最大的、能被字符集长度整除的倍数，超出该值的随机字节直接丢弃（拒绝采样）
	max := byte(256 - 256%len(randomCharset))
	b := make([]byte, 0, length)
	buf := make([]byte, length)
	for len(b) < length {
		if _, err := rand.Read(buf); err != nil {
			// 系统熵源不可用（极罕见），返回已生成部分之前的空串更安全
			return ""
		}
		for _, v := range buf {
			if v >= max {
				continue
			}
			b = append(b, randomCharset[int(v)%len(randomCharset)])
			if len(b) == length {
				break
			}
		}
	}
	return string(b)
}

/* ---------------- 正则过滤（正则均预编译，包初始化时完成） ---------------- */

var (
	regStyle   = regexp.MustCompile(`(?s)<style.*?</style>`)
	regScript  = regexp.MustCompile(`(?s)<script.*?</script>`)
	regHtmlTag = regexp.MustCompile(`<[^>]*>`)
	regATag    = regexp.MustCompile(`(?si)</?a\b[^>]*>`)
	regImgTag  = regexp.MustCompile(`(?si)<img\b[^>]*>`)
	regSpecial = regexp.MustCompile(`[+=|{}':;',]`)
	regUrlHead = regexp.MustCompile(`\w+://`)
)

// FilterByRegex 用正则表达式 expr 把匹配内容替换为 placeTo，表达式非法时原样返回。
// 场景：通用文本清洗，如 FilterByRegex(`\d`, "a1b2", "") = "ab"。
func (s *utilStr) FilterByRegex(expr, input, placeTo string) string {
	reg, err := regexp.Compile(expr)
	if err != nil {
		return input
	}
	return reg.ReplaceAllString(input, placeTo)
}

// FilterStyle 删除 HTML 中的 <style>...</style> 块。
// 场景：富文本正文提取纯内容前的预处理。
func (s *utilStr) FilterStyle(input string) string {
	return regStyle.ReplaceAllString(input, "")
}

// FilterScript 删除 HTML 中的 <script>...</script> 块。
// 场景：富文本入库前的 XSS 缓解（需配合其它转义手段，不能只依赖此过滤）。
func (s *utilStr) FilterScript(input string) string {
	return regScript.ReplaceAllString(input, "")
}

// FilterHtml 删除全部 HTML 标签，保留标签间文本。
// FilterHtml("<p>你好<b>世界</b></p>") = "你好世界"。
// 场景：生成摘要、邮件正文转纯文本。
func (s *utilStr) FilterHtml(input string) string {
	return regHtmlTag.ReplaceAllString(input, "")
}

// FilterA 删除 <a> 开标签与 </a> 闭标签（保留链接文字），不影响 <abbr> 等其它标签。
// 场景：保留富文本内容但去掉超链接跳转。
func (s *utilStr) FilterA(input string) string {
	return regATag.ReplaceAllString(input, "")
}

// FilterImage 删除 <img> 标签。场景：无图模式、正文纯文本化。
func (s *utilStr) FilterImage(input string) string {
	return regImgTag.ReplaceAllString(input, "")
}

// FilterSpecialChar 删除特殊字符 [+=|{}':;',]。
// 场景：生成文件名、导出标题时去除易干扰字符。
func (s *utilStr) FilterSpecialChar(input string) string {
	return regSpecial.ReplaceAllString(input, "")
}

// FilterUrlPrefix 删除字符串开头的协议前缀，"https://a.com" -> "a.com"。
// 场景：域名归一化比对。
func (s *utilStr) FilterUrlPrefix(input string) string {
	return regUrlHead.ReplaceAllString(input, "")
}

/* ---------------- 格式校验 ---------------- */

// IsNumber 是否为非负整数字符串（^\d+$，负数与小数均不匹配）。
// 场景：入参类型预判。
func (s *utilStr) IsNumber(input string) bool {
	match, _ := regexp.MatchString(`^\d+$`, input)
	return match
}

// IsIP 是否为合法 IP 地址（IPv4 或 IPv6 均可）。
// 场景：日志分析、访问控制列表校验。
func (s *utilStr) IsIP(input string) bool {
	return net.ParseIP(input) != nil
}

// IsIPv4 是否为合法 IPv4 地址。场景：区分 v4/v6 分别处理。
func (s *utilStr) IsIPv4(ip string) bool {
	parsed := net.ParseIP(ip)
	return parsed != nil && parsed.To4() != nil
}

// IsIPv6 是否为合法 IPv6 地址。场景：区分 v4/v6 分别处理。
func (s *utilStr) IsIPv6(ip string) bool {
	parsed := net.ParseIP(ip)
	return parsed != nil && parsed.To4() == nil
}

// IsDomain 是否为域名格式（多级点分，不含协议前缀）。
// IsDomain("www.example.com") = true，IsDomain("https://a.com") = false。
func (s *utilStr) IsDomain(domain string) bool {
	match, _ := regexp.MatchString(`^[a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)+$`, domain)
	return match
}

// IsDnsDomain 严格校验 DNS 域名：每段 1~63 位、允许末尾根点。
// 场景：DNS 记录、证书 CSR 的域名合法性校验。
func (s *utilStr) IsDnsDomain(domain string) bool {
	match, _ := regexp.MatchString(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}\.?$`, domain)
	return match
}

// IsEmail 是否为邮箱格式（常用宽松规则）。场景：注册/订阅表单校验。
func (s *utilStr) IsEmail(input string) bool {
	match, _ := regexp.MatchString(`^([a-z0-9A-Z]+[-|\.]?)+[a-z0-9A-Z]@([a-z0-9A-Z]+(-[a-z0-9A-Z]+)?\.)+[a-zA-Z]{2,}$`, input)
	return match
}

// IsMobile 是否为中国大陆 11 位手机号（1 开头，第二位 3-9）。场景：手机号入参校验。
func (s *utilStr) IsMobile(input string) bool {
	match, _ := regexp.MatchString(`^1[3-9]\d{9}$`, input)
	return match
}

// IsIDCard 是否为 18 位身份证号格式（末位可为 X/x）。
// 只校验格式，不校验地区码与校验位。场景：实名信息录入校验。
func (s *utilStr) IsIDCard(input string) bool {
	match, _ := regexp.MatchString(`^\d{17}[0-9Xx]$`, input)
	return match
}

// IsURL 是否为 http/https 链接。场景：外链白名单校验前的初筛。
func (s *utilStr) IsURL(input string) bool {
	match, _ := regexp.MatchString(`^https?://\S+$`, input)
	return match
}
