package utils

import (
	"bytes"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

type utilStr struct {
}

func (_self *utilStr) ToInt(str string) int {
	_num, _ := strconv.Atoi(str)
	return _num
}

func (_self *utilStr) ToInt64(str string) int64 {
	_num, _ := strconv.ParseInt(str, 10, 64)
	return _num
}

func (_self *utilStr) ToInteger(str string) (int, error) {
	_num, _err := strconv.Atoi(str)
	return _num, _err
}

func (_self *utilStr) ToLong(str string) (int64, error) {
	_num, _err := strconv.ParseInt(str, 10, 64)
	return _num, _err
}

func (_self *utilStr) ToFloat64(str string) (float64, error) {
	_num, _err := strconv.ParseFloat(str, 64)
	return _num, _err
}

func (_self *utilStr) BinaryToInt(str string) (int64, error) {
	_num, _err := strconv.ParseInt(str, 2, 64)
	return _num, _err
}

func (_self *utilStr) IntToBinary(num int64) string {
	bin := strconv.FormatInt(num, 2)
	return bin
}

func (_self *utilStr) IsBinaryOverInt(binStr string, number int64) bool {
	_num, _ := strconv.ParseInt(binStr, 2, 64)
	return (_num & number) == number
}

func (_self *utilStr) IsBinNumOverInt(binNum int64, number int64) bool {

	return (binNum & number) == number
}

func (_self *utilStr) ToStr(_num int) string {
	return strconv.Itoa(_num)
}

func (_self *utilStr) FormatInt(_num int) string {
	return strconv.FormatInt(int64(_num), 10)
}

func (_self *utilStr) FormatInt64(_num int64) string {
	return strconv.FormatInt(_num, 10)
}

func (_self *utilStr) FormatFloat64(_num float64) string {
	return strconv.FormatFloat(_num, 'f', 2, 64)
}

func (_self *utilStr) IsEmpty(str string) bool {

	return _self.Len(str) <= 0
}

func (_self *utilStr) IsNotEmpty(str string) bool {

	return !_self.IsEmpty(str)
}

func (_self *utilStr) IsEmptyStr(str string) bool {

	trimmedStr := strings.TrimSpace(str)
	// 检查去除空格后的字符串长度是否为 0
	return len(trimmedStr) == 0
}

func (_self *utilStr) TrimSpace(str string) string {
	//去掉首尾的空格
	return strings.TrimSpace(str)
}

func (_self *utilStr) Replace(str string, find string, to string) string {

	return strings.Replace(str, find, to, 1)
}

func (_self *utilStr) ReplaceAll(str string, find string, to string) string {

	return strings.Replace(str, find, to, -1)
}

func (_self *utilStr) Split(str string, spChar string) []string {

	return strings.Split(str, spChar)
}

func (_self *utilStr) Contains(str string, find string) bool {

	return strings.Contains(str, find)
}

// strings.HasPrefix("ABC_xyz", "ABC")
func (_self *utilStr) StartsWith(str string, find string) bool {

	return strings.HasPrefix(str, find)
}

// strings.HasSuffix("ABC_xyz", "xyz")
func (_self *utilStr) EndsWith(str string, find string) bool {

	return strings.HasSuffix(str, find)
}

// strings.Count("cheese", "e") = 3
func (_self *utilStr) Count(str string, find string) int {

	return strings.Count(str, find)
}

func (_self *utilStr) SubStr(input string, start, end int) string {

	return input[start:end]
}

// 返回第一个匹配字符的位置，返回-1为未找到
// strings.Index("ABC_xyz", "xyz") = 4
// strings.Index("ABC_xyz", "B") = 1
func (_self *utilStr) Index(str string, find string) int {

	return strings.Index(str, find)
}

// strings.Join(arrays, ",") = "foo, bar, bas"
func (_self *utilStr) Join(strs []string, spChar string) string {

	return strings.Join(strs, spChar)
}

// 字母转为小写
// strings.ToLower("Love GoLang") = "love golang"
func (_self *utilStr) ToLower(str string) string {

	return strings.ToLower(str)
}

// 字母转为大写
// strings.ToTitle("love 中国") = "LOVE 中国"
func (_self *utilStr) ToUpper(str string) string {
	return strings.ToUpper(str)
	//return strings.ToTitle(str)
}

func (_self *utilStr) Len(str string) int {

	return len(str)
}

func (_self *utilStr) Print(str string) {
	//var show = fmt.Println
	//show(str)
	fmt.Println(str)
}

func (_self *utilStr) FilterByRegex(expr, input, placeTo string) string {
	regx, _ := regexp.Compile(expr)
	return regx.ReplaceAllString(input, placeTo)
}

func (_self *utilStr) FilterStyle(input string) string {
	//regx, _ := regexp.Compile("<style((?:.|\\n)*?)</style>")
	regx, _ := regexp.Compile("\\<style[\\S\\s]+?\\</style\\>")
	return regx.ReplaceAllString(input, "")
}

func (_self *utilStr) FilterScript(input string) string {
	//regx, _ := regexp.Compile("<script((?:.|\\n)*?)</script>")
	regx, _ := regexp.Compile("\\<script[\\S\\s]+?\\</script\\>")
	return regx.ReplaceAllString(input, "")
}

func (_self *utilStr) FilterHtml(input string) string {
	regx, _ := regexp.Compile("<.+?>")
	//regx, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
	return regx.ReplaceAllString(input, "")
}

func (_self *utilStr) FilterA(input string) string {

	regx, _ := regexp.Compile("<.?a(.|\n)*?>")
	return regx.ReplaceAllString(input, "")
}

func (_self *utilStr) FilterImage(input string) string {

	regx, _ := regexp.Compile("<img(.|\\n)*?>")
	return regx.ReplaceAllString(input, "")
}

func (_self *utilStr) FilterSpecialChar(input string) string {

	regx, _ := regexp.Compile("[+=|{}':;',]")
	return regx.ReplaceAllString(input, "")
}

func (_self *utilStr) FilterUrlPrefix(input string) string {

	regx, _ := regexp.Compile("\\w+://")
	return regx.ReplaceAllString(input, "")
}

func (_self *utilStr) IsNumber(input string) bool {

	match, _ := regexp.MatchString("^\\d+$", input)
	return match
}

func (_self *utilStr) IsIP(input string) bool {

	match, _ := regexp.MatchString("^((2[0-4]\\d|25[0-5]|[01]?\\d\\d?)\\.){3}(2[0-4]\\d|25[0-5]|[01]?\\d\\d?)$", input)
	return match
}

// IsIPv4 判断是否为 IPv4 地址
func (_self *utilStr) IsIPv4(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.To4() != nil
}

// IsIPv6 判断是否为 IPv6 地址
func (_self *utilStr) IsIPv6(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.To4() == nil
}

// IsDomain 判断是否为域名格式
func (_self *utilStr) IsDomain(domain string) bool {
	// 域名的正则表达式
	regex := `^[a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)+$`
	//regex := `^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`
	match, _ := regexp.MatchString(regex, domain)
	return match
}

func (_self *utilStr) IsDnsDomain(domain string) bool {
	// 域名的正则表达式，允许域名以点号结尾
	regex := `^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}\.?$`
	match, _ := regexp.MatchString(regex, domain)
	return match
}

func (_self *utilStr) IsEMail(input string) bool {

	match, _ := regexp.MatchString("^([a-z0-9A-Z]+[-|\\.]?)+[a-z0-9A-Z]@([a-z0-9A-Z]+(-[a-z0-9A-Z]+)?\\.)+[a-zA-Z]{2,}$", input)
	return match
}

// 高效拼接字符串
func (_self *utilStr) LinkStrs(inputs ...string) string {
	var buf bytes.Buffer
	for _, v := range inputs {
		buf.WriteString(v)
	}
	return buf.String()
}

func (_self *utilStr) LinkInputs(inputs ...interface{}) string {
	var buf bytes.Buffer
	for _, v := range inputs {
		switch t := v.(type) {
		case string:
			buf.WriteString(t)
		default:
			buf.WriteString(fmt.Sprint(t))

		}
	}
	return buf.String()
}

// 格式化字节单位
func (_self *utilStr) FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
