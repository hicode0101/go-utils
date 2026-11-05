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

// Trim 去除首尾空白字符（对应 hutool trim）。场景：清洗用户输入、粘贴内容。
func (s *utilStr) Trim(str string) string {
	return strings.TrimSpace(str)
}

// TrimStart 去除头部空白字符。场景：规整多行文本缩进。
func (s *utilStr) TrimStart(str string) string {
	return strings.TrimLeft(str, " \t\n\r\v\f")
}

// TrimEnd 去除尾部空白字符。场景：去掉结尾换行。
func (s *utilStr) TrimEnd(str string) string {
	return strings.TrimRight(str, " \t\n\r\v\f")
}

// TrimToNull 去除首尾空白后为空串则返回空串（Go 无 null，语义为"归一化到空串"）。
// 场景：可选字段清洗，空值统一为 "" 便于落库。
func (s *utilStr) TrimToNull(str string) string {
	trimmed := strings.TrimSpace(str)
	if trimmed == "" {
		return ""
	}
	return trimmed
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

// StartWith 是否以 find 开头（对应 hutool startWith）。场景：判断协议前缀，如 "https://"。
func (s *utilStr) StartWith(str, find string) bool {
	return strings.HasPrefix(str, find)
}

// EndWith 是否以 find 结尾（对应 hutool endWith）。场景：判断文件扩展名。
func (s *utilStr) EndWith(str, find string) bool {
	return strings.HasSuffix(str, find)
}

// ContainsIgnoreCase 忽略大小写判断是否包含子串。场景：协议头/关键字大小写不敏感匹配。
func (s *utilStr) ContainsIgnoreCase(str, find string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(find))
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

// Sub 按"字符"截取子串 [start, end)（对应 hutool sub，按字符而非字节，中文安全）。
// 支持 hutool 语义：下标为负表示从末尾倒数（-1 为最后一个字符）；start > end 时自动交换；
// 越界下标自动收敛到边界。Sub("你好世界", 1, 3) = "好世"，Sub("hello", -3, -1) = "ll"。
// 场景：标题摘要展示、按位置裁剪文本。
func (s *utilStr) Sub(str string, start, end int) string {
	runes := []rune(str)
	n := len(runes)
	if start < 0 {
		start = n + start
		if start < 0 {
			start = 0
		}
	}
	if start > n {
		start = n
	}
	if end < 0 {
		end = n + end
		if end < 0 {
			end = 0
		}
	}
	if end > n {
		end = n
	}
	if start > end {
		start, end = end, start
	}
	return string(runes[start:end])
}

// SubBefore 截取分隔符 sep 之前的部分（isLast=true 取最后一个分隔符之前）。
// 未找到分隔符返回原串；分隔符位于开头返回空串。SubBefore("abcba", "b", false) = "a"。
// 场景：取文件路径目录部分、取域名主机部分。
func (s *utilStr) SubBefore(str, sep string, isLast bool) string {
	if str == "" {
		return ""
	}
	if sep == "" {
		return ""
	}
	pos := strings.Index(str, sep)
	if isLast {
		pos = strings.LastIndex(str, sep)
	}
	if pos < 0 {
		return str
	}
	if pos == 0 {
		return ""
	}
	return str[:pos]
}

// SubAfter 截取分隔符 sep 之后的部分（isLast=true 取最后一个分隔符之后）。
// 未找到分隔符、或分隔符位于末尾时返回空串。SubAfter("abcba", "b", false) = "cba"。
// 场景：取文件扩展名、取查询参数值。
func (s *utilStr) SubAfter(str, sep string, isLast bool) string {
	if str == "" {
		return ""
	}
	if sep == "" {
		return ""
	}
	pos := strings.Index(str, sep)
	if isLast {
		pos = strings.LastIndex(str, sep)
	}
	if pos < 0 || pos == len(str)-1 {
		return ""
	}
	return str[pos+len(sep):]
}

// SubBetween 截取 left 与 right 标识串之间的内容（取第一个匹配对，不含标识本身）。
// 未成对出现返回空串。SubBetween("wx[b]yz", "[", "]") = "b"。
// 场景：提取括号/标签内容。
func (s *utilStr) SubBetween(str, left, right string) string {
	if str == "" || left == "" || right == "" {
		return ""
	}
	start := strings.Index(str, left)
	if start < 0 {
		return ""
	}
	start += len(left)
	end := strings.Index(str[start:], right)
	if end < 0 {
		return ""
	}
	return str[start : start+end]
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

// PadPre 左侧填充 pad 到总长 length，超长原样返回（对应 hutool padPre）。
// PadPre("7", 3, '0') = "007"。场景：补齐定长编号。
func (s *utilStr) PadPre(str string, length int, pad rune) string {
	n := utf8.RuneCountInString(str)
	if n >= length {
		return str
	}
	return strings.Repeat(string(pad), length-n) + str
}

// PadAfter 右侧填充 pad 到总长 length（对应 hutool padAfter）。
// PadAfter("ab", 4, '*') = "ab**"。场景：表格对齐输出。
func (s *utilStr) PadAfter(str string, length int, pad rune) string {
	n := utf8.RuneCountInString(str)
	if n >= length {
		return str
	}
	return str + strings.Repeat(string(pad), length-n)
}

// Center 居中填充：两侧填充 pad 到总长 length（hutool center），差值为奇数时右侧多填一个。
// Center("ab", 5, '*') = "*ab**"。场景：标题栏居中展示。
func (s *utilStr) Center(str string, length int, pad rune) string {
	n := utf8.RuneCountInString(str)
	if n >= length {
		return str
	}
	pads := length - n
	left := pads / 2
	return strings.Repeat(string(pad), left) + str + strings.Repeat(string(pad), pads-left)
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

/* ---------------- 默认值（对应 hutool nullToDefault 系列扁平化到空值处理） ---------------- */

// NullToDefault Go 无 null，等价于"空串给默认值"：str 为空串时返回 defaultStr。
// 场景：可选参数兜底。
func (s *utilStr) NullToDefault(str, defaultStr string) string {
	if s.IsEmpty(str) {
		return defaultStr
	}
	return str
}

// EmptyToDefault 空串时返回 defaultStr（同 NullToDefault，保留 hutool 命名便于对照）。
func (s *utilStr) EmptyToDefault(str, defaultStr string) string {
	return s.NullToDefault(str, defaultStr)
}

// BlankToDefault 去除首尾空白后为空时返回 defaultStr。
// BlankToDefault("  ", "匿名") = "匿名"。场景：表单字段兜底。
func (s *utilStr) BlankToDefault(str, defaultStr string) string {
	if s.IsBlank(str) {
		return defaultStr
	}
	return str
}

/* ---------------- 前后缀处理 ---------------- */

// RemovePrefix 移除开头匹配的 prefix（不匹配则原样返回）。
// RemovePrefix("https://a.com", "https://") = "a.com"。场景：去协议头。
func (s *utilStr) RemovePrefix(str, prefix string) string {
	return strings.TrimPrefix(str, prefix)
}

// RemoveSuffix 移除结尾匹配的 suffix（不匹配则原样返回）。
// RemoveSuffix("main.go", ".go") = "main"。场景：取不含扩展名的名称。
func (s *utilStr) RemoveSuffix(str, suffix string) string {
	return strings.TrimSuffix(str, suffix)
}

// AddPrefixIfNot 不以 prefix 开头时补上，已开头原样返回。
// 场景：规范路径/URL 前缀。
func (s *utilStr) AddPrefixIfNot(str, prefix string) string {
	if s.IsEmpty(str) || strings.HasPrefix(str, prefix) {
		return str
	}
	return prefix + str
}

// AddSuffixIfNot 不以 suffix 结尾时补上，已结尾原样返回。
// 场景：确保路径以 / 结尾。
func (s *utilStr) AddSuffixIfNot(str, suffix string) string {
	if s.IsEmpty(str) || strings.HasSuffix(str, suffix) {
		return str
	}
	return str + suffix
}

// Wrap 用 prefix 与 suffix 包裹 str（已包裹则不重复包裹，对应 hutool wrapIfMissing 语义）。
// Wrap("id", "${", "}") = "${id}"。场景：模板变量占位。
func (s *utilStr) Wrap(str, prefix, suffix string) string {
	if s.IsWrap(str, prefix, suffix) {
		return str
	}
	return prefix + str + suffix
}

// UnWrap 去除首尾的 prefix 与 suffix 包裹（未包裹则原样返回）。
// UnWrap("${id}", "${", "}") = "id"。
func (s *utilStr) UnWrap(str, prefix, suffix string) string {
	if s.IsWrap(str, prefix, suffix) {
		return str[len(prefix) : len(str)-len(suffix)]
	}
	return str
}

// IsWrap 判断 str 是否被 prefix 与 suffix 包裹。场景：识别模板占位符。
func (s *utilStr) IsWrap(str, prefix, suffix string) bool {
	return strings.HasPrefix(str, prefix) && strings.HasSuffix(str, suffix) && len(str) >= len(prefix)+len(suffix)
}

/* ---------------- 大小写与命名风格（对应 hutool NamingCase） ---------------- */

// UpperFirst 首字母转大写。UpperFirst("user") = "User"。场景：生成导出风格标识。
func (s *utilStr) UpperFirst(str string) string {
	if str == "" {
		return str
	}
	runes := []rune(str)
	first := runes[0]
	if first >= 'a' && first <= 'z' {
		first -= 'a' - 'A'
	}
	return string(first) + string(runes[1:])
}

// LowerFirst 首字母转小写。LowerFirst("User") = "user"。
func (s *utilStr) LowerFirst(str string) string {
	if str == "" {
		return str
	}
	runes := []rune(str)
	first := runes[0]
	if first >= 'A' && first <= 'Z' {
		first += 'a' - 'A'
	}
	return string(first) + string(runes[1:])
}

// SwapCase 大小写互换："Love Go" -> "lOVE gO"。
func (s *utilStr) SwapCase(str string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r - ('a' - 'A')
		case r >= 'A' && r <= 'Z':
			return r + ('a' - 'A')
		}
		return r
	}, str)
}

// ToUnderlineCase 驼峰转下划线（hutool NamingCase.toUnderlineCase）。
// ToUnderlineCase("userName") = "user_name"，"UserName" -> "user_name"。
// 场景：Go 结构体字段映射数据库列名。
func (s *utilStr) ToUnderlineCase(str string) string {
	return s.ToSymbolCase(str, '_')
}

// ToKebabCase 驼峰转短横线（hutool NamingCase.toKebabCase）。
// ToKebabCase("userName") = "user-name"。场景：URL 路径段、CSS 类名。
func (s *utilStr) ToKebabCase(str string) string {
	return s.ToSymbolCase(str, '-')
}

// ToSymbolCase 驼峰转指定连接符（hutool NamingCase.toSymbolCase 完整移植）：
// 连续大写按专有名词对待，"HTTPServer" -> "http_server"。
func (s *utilStr) ToSymbolCase(str string, symbol rune) string {
	runes := []rune(str)
	var b strings.Builder
	b.Grow(len(runes) + 4)
	for i, c := range runes {
		if c >= 'A' && c <= 'Z' {
			var pre, next rune
			hasPre, hasNext := i > 0, i < len(runes)-1
			if hasPre {
				pre = runes[i-1]
			}
			if hasNext {
				next = runes[i+1]
			}
			isLower := func(r rune) bool { return r >= 'a' && r <= 'z' }
			isDigit := func(r rune) bool { return r >= '0' && r <= '9' }
			if hasPre {
				switch {
				case pre == symbol:
					// 前一个为分隔符：_Abb -> _abb；_AB（后一个大写）按专有名词保留
					if !hasNext || isLower(next) {
						c += 'a' - 'A'
					}
				case isLower(pre):
					// 前一个为小写：aBcc -> a_bcc；aBC 按专有名词保留
					b.WriteRune(symbol)
					if !hasNext || isLower(next) || isDigit(next) {
						c += 'a' - 'A'
					}
				default:
					// 前一个为大写：ABcc -> A_bcc；ABC 按专有名词保留
					if hasNext && isLower(next) {
						b.WriteRune(symbol)
						c += 'a' - 'A'
					}
				}
			} else {
				// 首字母：Abc -> abc；ABC 按专有名词保留
				if !hasNext || isLower(next) {
					c += 'a' - 'A'
				}
			}
		}
		b.WriteRune(c)
	}
	return b.String()
}

// ToCamelCase 下划线/连接符命名转驼峰（hutool NamingCase.toCamelCase）。
// ToCamelCase("user_name") = "userName"，ToCamelCase("USER_NAME") = "userName"。
func (s *utilStr) ToCamelCase(str string) string {
	return s.toCamelCaseWithSymbol(str, '_')
}

// ToPascalCase 下划线/连接符命名转帕斯卡（大驼峰）。
// ToPascalCase("user_name") = "UserName"。场景：生成结构体名。
func (s *utilStr) ToPascalCase(str string) string {
	return s.UpperFirst(s.ToCamelCase(str))
}

// toCamelCaseWithSymbol 按指定连接符转驼峰（hutool toCamelCase(name, symbol, otherCharToLower=true)）
func (s *utilStr) toCamelCaseWithSymbol(str string, symbol rune) string {
	if !strings.ContainsRune(str, symbol) {
		return str
	}
	var b strings.Builder
	b.Grow(len(str))
	upperNext := false
	for _, c := range str {
		switch {
		case c == symbol:
			upperNext = true
		case upperNext:
			if c >= 'a' && c <= 'z' {
				c -= 'a' - 'A'
			}
			b.WriteRune(c)
			upperNext = false
		default:
			if c >= 'A' && c <= 'Z' {
				c += 'a' - 'A'
			}
			b.WriteRune(c)
		}
	}
	return b.String()
}

/* ---------------- 拆分与格式化 ---------------- */

// SplitTrim 按分隔符拆分并去除每项首尾空白，丢弃空白项（hutool splitTrim）。
// SplitTrim("a, b , ,c", ",") = ["a","b","c"]。场景：解析用户手填的标签列表。
func (s *utilStr) SplitTrim(str, sep string) []string {
	if str == "" {
		return nil
	}
	parts := strings.Split(str, sep)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			result = append(result, p)
		}
	}
	return result
}

// SplitByLength 按固定长度（字符数）切分为多段，最后一段可能不足。
// SplitByLength("abcdef", 2) = ["ab","cd","ef"]。场景：定长报文逐段解析。
func (s *utilStr) SplitByLength(str string, length int) []string {
	if length <= 0 || str == "" {
		return nil
	}
	runes := []rune(str)
	result := make([]string, 0, (len(runes)+length-1)/length)
	for start := 0; start < len(runes); start += length {
		end := start + length
		if end > len(runes) {
			end = len(runes)
		}
		result = append(result, string(runes[start:end]))
	}
	return result
}

// Format 用 "{}" 占位符按顺序填充模板（hutool StrFormatter.format）。
// Format("{} 是 {} 岁", "张三", 18) = "张三 是 18 岁"；占位符多于参数时多余占位符原样保留。
// 场景：日志拼装、简单模板渲染。
func (s *utilStr) Format(template string, args ...interface{}) string {
	if template == "" || len(args) == 0 {
		return template
	}
	var b strings.Builder
	argIndex := 0
	for i := 0; i < len(template); {
		if template[i] == '{' && i+1 < len(template) && template[i+1] == '}' && argIndex < len(args) {
			b.WriteString(fmt.Sprint(args[argIndex]))
			argIndex++
			i += 2
			continue
		}
		b.WriteByte(template[i])
		i++
	}
	return b.String()
}

// FormatWithMap 用 "{key}" 占位符从 map 取值填充模板（hutool CharSequenceUtil.format）。
// key 不存在时占位符原样保留。FormatWithMap("您好 {name}", map{"name":"张三"}) = "您好 张三"。
// 场景：消息/邮件模板渲染。
func (s *utilStr) FormatWithMap(template string, m map[string]string) string {
	if template == "" || len(m) == 0 {
		return template
	}
	var b strings.Builder
	for i := 0; i < len(template); {
		if template[i] == '{' {
			if end := strings.IndexByte(template[i:], '}'); end > 1 {
				key := template[i+1 : i+end]
				if val, ok := m[key]; ok {
					b.WriteString(val)
					i += end + 1
					continue
				}
			}
		}
		b.WriteByte(template[i])
		i++
	}
	return b.String()
}

/* ---------------- 相似度与摘要（对应 hutool TextSimilarity / brief / hide） ---------------- */

// Similar 计算两个字符串的相似度 [0,1]（hutool TextSimilarity.similar）：
// 仅比较字母/数字/汉字（符号先剔除），最长公共子序列长度 / 较长者长度；两空串相似度为 1。
// 场景：重复内容初筛、模糊匹配评分。
func (s *utilStr) Similar(a, b string) float64 {
	sa, sb := s.removeSign(a), s.removeSign(b)
	longer := len([]rune(sa))
	if lb := len([]rune(sb)); lb > longer {
		longer = lb
	}
	if longer == 0 {
		return 1
	}
	return float64(s.longestCommonSubsequenceLen(sa, sb)) / float64(longer)
}

// SimilarPercent 计算相似度百分比字符串（hutool similar(str, str, scale)）。
// SimilarPercent("abc", "abd", 2) = "66.67%"。场景：报表展示。
func (s *utilStr) SimilarPercent(a, b string, scale int) string {
	pct := s.Similar(a, b) * 100
	return strconv.FormatFloat(pct, 'f', scale, 64) + "%"
}

// LongestCommonSubstring 返回最长公共子序列（hutool longestCommonSubstring，动态规划，
// 不要求连续）。LongestCommonSubstring("abcdef", "acef") = "acef"。
// 场景：diff 类工具、相似度分析的基础算法。
func (s *utilStr) LongestCommonSubstring(a, b string) string {
	ar, br := []rune(a), []rune(b)
	m, n := len(ar), len(br)
	matrix := make([][]int, m+1)
	for i := range matrix {
		matrix[i] = make([]int, n+1)
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if ar[i-1] == br[j-1] {
				matrix[i][j] = matrix[i-1][j-1] + 1
			} else if matrix[i-1][j] >= matrix[i][j-1] {
				matrix[i][j] = matrix[i-1][j]
			} else {
				matrix[i][j] = matrix[i][j-1]
			}
		}
	}
	// 回溯取出子序列
	result := make([]rune, matrix[m][n])
	idx := len(result) - 1
	for m > 0 && n > 0 {
		switch {
		case matrix[m][n] == matrix[m-1][n]:
			m--
		case matrix[m][n] == matrix[m][n-1]:
			n--
		default:
			result[idx] = ar[m-1]
			idx--
			m--
			n--
		}
	}
	return string(result)
}

// removeSign 剔除字母/数字/汉字以外的字符（hutool TextSimilarity.removeSign）
func (s *utilStr) removeSign(str string) string {
	var b strings.Builder
	b.Grow(len(str))
	for _, c := range str {
		if (c >= 0x4E00 && c <= 0x9FFF) || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			b.WriteRune(c)
		}
	}
	return b.String()
}

// longestCommonSubsequenceLen 最长公共子序列长度（动态规划，滚动数组省内存）
func (s *utilStr) longestCommonSubsequenceLen(a, b string) int {
	ar, br := []rune(a), []rune(b)
	m, n := len(ar), len(br)
	prev := make([]int, n+1)
	curr := make([]int, n+1)
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if ar[i-1] == br[j-1] {
				curr[j] = prev[j-1] + 1
			} else if prev[j] >= curr[j-1] {
				curr[j] = prev[j]
			} else {
				curr[j] = curr[j-1]
			}
		}
		prev, curr = curr, prev
	}
	return prev[n]
}

// Brief 超长缩略：超过 maxLength 个字符时保留前 maxLength-3 个字符并以 "..." 结尾
// （hutool brief）；maxLength<=3 时直接截取前 maxLength 个字符。
// Brief("你好世界欢迎", 5) = "你好..."。场景：列表页单行摘要。
func (s *utilStr) Brief(str string, maxLength int) string {
	n := utf8.RuneCountInString(str)
	if n <= maxLength {
		return str
	}
	if maxLength <= 3 {
		return s.Sub(str, 0, maxLength)
	}
	return s.Sub(str, 0, maxLength-3) + "..."
}

// Hide 将 [start, end) 区间的字符替换为 '*'（hutool hide，按字符计，支持负下标语义同 Sub）。
// Hide("13812345678", 3, 7) = "138****5678"。场景：手机号/证件号打码。
func (s *utilStr) Hide(str string, start, end int) string {
	if str == "" {
		return str
	}
	runes := []rune(str)
	n := len(runes)
	if start < 0 {
		start = n + start
		if start < 0 {
			start = 0
		}
	}
	if end < 0 {
		end = n + end
		if end < 0 {
			end = 0
		}
	}
	if end > n {
		end = n
	}
	if start >= end {
		return str
	}
	result := make([]rune, n)
	copy(result, runes)
	for i := start; i < end; i++ {
		result[i] = '*'
	}
	return string(result)
}

/* ---------------- 全角/半角（对应 hutool Convert.toSBC / toDBC） ---------------- */

// ToSBC 半角转全角（hutool Convert.toSBC）：ASCII 可见字符与空格转为全角形态。
// ToSBC("aA1 ") = "ａＡ１　"。场景：统一全半角后比对。
func (s *utilStr) ToSBC(str string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r == ' ':
			return 12288 // 全角空格 U+3000
		case r >= 33 && r <= 126:
			return r + 65248 // 偏移 0xFEE0
		}
		return r
	}, str)
}

// ToDBC 全角转半角（hutool Convert.toDBC）：全角字符与全角空格还原为半角。
// ToDBC("ａＡ１　") = "aA1 "。
func (s *utilStr) ToDBC(str string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r == 12288:
			return ' '
		case r >= 65281 && r <= 65374: // 0xFF01 ~ 0xFF5E
			return r - 65248
		}
		return r
	}, str)
}

/* ---------------- 通用字段提取 ---------------- */

// CommonPrefix 返回多个字符串的最长公共前缀（hutool commonPrefix）。
// CommonPrefix("abc123", "abcd") = "abc"。场景：批量文件名归一化。
func (s *utilStr) CommonPrefix(strs ...string) string {
	if len(strs) == 0 {
		return ""
	}
	prefix := strs[0]
	for _, str := range strs[1:] {
		for strings.Index(str, prefix) != 0 {
			prefix = prefix[:len(prefix)-1]
			if prefix == "" {
				return ""
			}
		}
	}
	return prefix
}

// CommonSuffix 返回多个字符串的最长公共后缀。场景：批量扩展名统计。
func (s *utilStr) CommonSuffix(strs ...string) string {
	if len(strs) == 0 {
		return ""
	}
	suffix := strs[0]
	for _, str := range strs[1:] {
		for len(suffix) > 0 && !strings.HasSuffix(str, suffix) {
			suffix = suffix[1:]
		}
		if suffix == "" {
			return ""
		}
	}
	return suffix
}
