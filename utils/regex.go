package utils

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// utilRegex 正则工具。
//
// 方法对照 hutool ReUtil 移植：所有正则统一加 (?s) 标志编译（对应 Java Pattern.DOTALL，
// 使 . 可以匹配换行符）；表达式编译失败时不抛异常，按各方法语义返回零值或原文。
type utilRegex struct{}

// regexGroupVar 匹配模板中 $N 分组引用的正则（对应 hutool PatternPool.GROUP_VAR）
var regexGroupVar = regexp.MustCompile(`\$(\d+)`)

// regexNumber 匹配整数的正则（对应 hutool PatternPool.NUMBERS）
var regexNumber = regexp.MustCompile(`\d+`)

// regexCompile 编译正则，统一加 (?s) 标志（对应 hutool 各处的 Pattern.DOTALL）
func regexCompile(regex string) (*regexp.Regexp, error) {
	return regexp.Compile(`(?s)` + regex)
}

/* ---------------- 查找与提取 ---------------- */

// Find 返回首个匹配的整体内容（分组 0），无匹配或表达式非法返回空串（hutool getGroup0）。
// Find("\\d+年", "2013年5月") = "2013年"。场景：提取第一个命中片段。
func (r *utilRegex) Find(regex, content string) string {
	re, err := regexCompile(regex)
	if err != nil {
		return ""
	}
	return re.FindString(content)
}

// FindGroup 返回首个匹配的指定分组内容（0 为整体，1 起为括号分组），无匹配或分组不存在返回空串
// （hutool get(regex, content, groupIndex)）。
// FindGroup("(\\d+)年(\\d+)月", "2013年5月", 2) = "5"。场景：从结构化文本中抽取字段。
func (r *utilRegex) FindGroup(regex, content string, groupIndex int) string {
	re, err := regexCompile(regex)
	if err != nil {
		return ""
	}
	subs := re.FindStringSubmatch(content)
	if subs == nil || groupIndex < 0 || groupIndex >= len(subs) {
		return ""
	}
	return subs[groupIndex]
}

// GetAllGroups 返回首个匹配的全部分组，withGroup0=false 时不包含整体分组（分组 0），
// 无匹配返回 nil（hutool getAllGroups）。
// GetAllGroups("(\\d+)年(\\d+)月", "2013年5月", false) = ["2013", "5"]。
// 场景：一次拿到一个匹配的全部字段。
func (r *utilRegex) GetAllGroups(regex, content string, withGroup0 bool) []string {
	re, err := regexCompile(regex)
	if err != nil {
		return nil
	}
	subs := re.FindStringSubmatch(content)
	if subs == nil {
		return nil
	}
	if withGroup0 {
		return subs
	}
	return subs[1:]
}

// FindAll 返回所有匹配中指定分组的内容列表，无匹配返回 nil（hutool findAll）。
// FindAll("\\d+", "a1b22c333", 0) = ["1", "22", "333"]；分组下标越界的匹配项跳过。
// 场景：批量提取，如抓取全文中所有数字、链接。
func (r *utilRegex) FindAll(regex, content string, groupIndex int) []string {
	re, err := regexCompile(regex)
	if err != nil {
		return nil
	}
	all := re.FindAllStringSubmatch(content, -1)
	if all == nil {
		return nil
	}
	result := make([]string, 0, len(all))
	for _, subs := range all {
		if groupIndex >= 0 && groupIndex < len(subs) {
			result = append(result, subs[groupIndex])
		}
	}
	return result
}

// FindIndex 返回首个匹配的字节区间 [start, end)，未匹配或表达式非法时 found=false（hutool indexOf）。
// 场景：需要自行处理原文时先定位，如找到位置后做高亮、替换。
func (r *utilRegex) FindIndex(regex, content string) (start, end int, found bool) {
	re, err := regexCompile(regex)
	if err != nil {
		return 0, 0, false
	}
	loc := re.FindStringIndex(content)
	if loc == nil {
		return 0, 0, false
	}
	return loc[0], loc[1], true
}

/* ---------------- 删除与替换 ---------------- */

// DelFirst 删除首个匹配的内容，无匹配或表达式非法时返回原文（hutool delFirst）。
// DelFirst("\\d+", "a1b22") = "ab22"。场景：只清理第一处脏字符。
func (r *utilRegex) DelFirst(regex, content string) string {
	re, err := regexCompile(regex)
	if err != nil {
		return content
	}
	loc := re.FindStringIndex(content)
	if loc == nil {
		return content
	}
	return content[:loc[0]] + content[loc[1]:]
}

// DelLast 删除最后一个匹配的内容，无匹配或表达式非法时返回原文（hutool delLast）。
// DelLast("\\d+", "a1b22") = "a1b"。场景：去掉结尾的序号、单位。
func (r *utilRegex) DelLast(regex, content string) string {
	re, err := regexCompile(regex)
	if err != nil {
		return content
	}
	all := re.FindAllStringIndex(content, -1)
	if all == nil {
		return content
	}
	last := all[len(all)-1]
	return content[:last[0]] + content[last[1]:]
}

// DelAll 删除全部匹配的内容，表达式非法时返回原文（hutool delAll）。
// DelAll("<[^>]+>", "<b>标题</b>") = "标题"。场景：清洗富文本标签、批量去除噪声。
func (r *utilRegex) DelAll(regex, content string) string {
	re, err := regexCompile(regex)
	if err != nil {
		return content
	}
	return re.ReplaceAllString(content, "")
}

// DelPre 删除首个匹配之前的内容（含匹配本身），即返回首个匹配结束位置之后的尾部内容；
// 无匹配或表达式非法时返回原文（对应 hutool delPre，源码取第一个 find() 结果的 end 截取）。
// DelPre("编号：123 正文", "\\d+") = " 正文"。场景：丢弃报文头部，保留首个标记之后的正文。
func (r *utilRegex) DelPre(regex, content string) string {
	re, err := regexCompile(regex)
	if err != nil {
		return content
	}
	loc := re.FindStringIndex(content)
	if loc == nil {
		return content
	}
	return content[loc[1]:]
}

// ReplaceAll 正则替换：每个匹配按 replTemplate 模板替换，$1/$2 引用对应分组、$0 为整体匹配
// （直接使用 Go regexp 原生 $ 语法，等价 hutool replaceAll 的模板能力）。
// ReplaceAll("中文1234", "(\\d+)", "($1)") = "中文(1234)"。场景：批量格式化、改写提取内容。
// 表达式非法时原样返回 content。
func (r *utilRegex) ReplaceAll(content, regex, replTemplate string) string {
	re, err := regexCompile(regex)
	if err != nil {
		return content
	}
	return re.ReplaceAllString(content, replTemplate)
}

/* ---------------- 判断与统计 ---------------- */

// Count 返回匹配次数（非重叠计数），表达式非法返回 0（hutool count）。
// Count("\\d+", "a1b22c333") = 3。场景：统计命中数量，如敏感词出现次数。
func (r *utilRegex) Count(regex, content string) int {
	re, err := regexCompile(regex)
	if err != nil {
		return 0
	}
	return len(re.FindAllStringIndex(content, -1))
}

// Contains 判断是否存在任意匹配（hutool contains，等价 Java find() 的包含匹配）。
// 场景：敏感词、关键字的快速判断。
func (r *utilRegex) Contains(regex, content string) bool {
	re, err := regexCompile(regex)
	if err != nil {
		return false
	}
	return re.MatchString(content)
}

// IsMatch 判断 content 是否被 regex 完整覆盖（对应 Java Matcher.matches() 的全串匹配语义）。
// 注意与 Go 标准库 regexp.MatchString 的区别：Go 的 MatchString 是包含匹配（子串命中即 true，
// 等价 Java find()），而本方法要求正则覆盖整个字符串，实现上等价于在正则两侧加 ^(?:...)$ 锚点，例如：
// IsMatch("\\d{4}", "12345") = false，而 regexp.MatchString("\\d{4}", "12345") = true。
// regex 为空串时直接返回 true（hutool 语义：无正则即不校验）。
// 场景：入参整体格式校验，如 IsMatch("1[3-9]\\d{9}", mobile) 校验整个串为合法手机号。
func (r *utilRegex) IsMatch(regex, content string) bool {
	if regex == "" {
		return true
	}
	re, err := regexp.Compile(`(?s)^(?:` + regex + `)$`)
	if err != nil {
		return false
	}
	return re.MatchString(content)
}

/* ---------------- 模板填充与其它 ---------------- */

// ExtractMulti 用首个匹配的分组按 template 填充新字符串，$1/$2 引用分组、$0 为整体匹配，
// 无匹配或表达式非法返回空串（hutool extractMulti）。
// ExtractMulti("(\\d+)年(\\d+)月", "2013年5月", "$1-$2") = "2013-5"。
// 模板中的分组编号按降序替换（避免 $1 抢先消费 $10 的前缀）；分组不存在的占位符原样保留。
func (r *utilRegex) ExtractMulti(pattern, content, template string) string {
	re, err := regexCompile(pattern)
	if err != nil {
		return ""
	}
	subs := re.FindStringSubmatch(content)
	if subs == nil {
		return ""
	}
	// 提取模板中的分组编号并按降序替换（对应 hutool TreeSet 降序遍历）
	for _, n := range regexExtractVarNums(template) {
		if n < len(subs) {
			template = strings.ReplaceAll(template, "$"+strconv.Itoa(n), subs[n])
		}
	}
	return template
}

// Escape 转义字符串中的正则元字符（regexp.QuoteMeta，对应 hutool escape），
// 使其可安全作为字面量参与正则匹配或拼接。
// Escape("1+1=2") = "1\\+1=2"。场景：把用户输入嵌入正则表达式。
func (r *utilRegex) Escape(str string) string {
	return regexp.QuoteMeta(str)
}

// FirstNumber 提取字符串中第一个整数；无数字或数值超出 int64 范围时返回错误（hutool getFirstNumber）。
// FirstNumber("abc123def456") = 123。场景：从报错信息、文件名中提取编号。
func (r *utilRegex) FirstNumber(content string) (int64, error) {
	num := regexNumber.FindString(content)
	if num == "" {
		return 0, fmt.Errorf("内容 %q 中未找到数字", content)
	}
	return strconv.ParseInt(num, 10, 64)
}

// regexExtractVarNums 提取模板中 $N 形式的分组编号，去重后按降序返回
func regexExtractVarNums(template string) []int {
	matches := regexGroupVar.FindAllStringSubmatch(template, -1)
	seen := make(map[int]struct{}, len(matches))
	nums := make([]int, 0, len(matches))
	for _, m := range matches {
		n, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		if _, ok := seen[n]; !ok {
			seen[n] = struct{}{}
			nums = append(nums, n)
		}
	}
	sort.Slice(nums, func(i, j int) bool { return nums[i] > nums[j] })
	return nums
}
