package utils

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// utilNumber 数值计算工具（对应 hutool NumberUtil）。
//
// 精度约定：浮点四则运算、除法与舍入内部统一用 big.Rat（从 float64 的最短十进制
// 字符串构造，避免 0.1 之类的二进制近似值被放大），算完再转回 float64，
// 因此 Add(0.1, 0.2)=0.3、Round(0.615, 2)=0.62 与 hutool 的 BigDecimal 行为一致。
type utilNumber struct{}

/* ---------------- 精确四则运算 ---------------- */

// Add 精确加法，返回多个数之和（对应 hutool add(double...)）。
// 内部用 big.Rat 累加，Add(0.1, 0.2) = 0.3 而非 0.30000000000000004；无参数返回 0。
// 场景：金额合计，如订单多个商品价格求和。
func (t *utilNumber) Add(nums ...float64) float64 {
	if len(nums) == 0 {
		return 0
	}
	sum := new(big.Rat)
	for _, v := range nums {
		sum.Add(sum, numberToRat(v))
	}
	f, _ := sum.Float64()
	return f
}

// Sub 精确减法，num 依次减去 others 中的每个数（对应 hutool sub(double...)）。
// Sub(1.0, 0.9, 0.1) = 0 而非 2.7755575615628914e-17。others 为空时返回 num 本身。
// 场景：余额扣减、优惠金额分摊后的差值核算。
func (t *utilNumber) Sub(num float64, others ...float64) float64 {
	result := numberToRat(num)
	for _, v := range others {
		result.Sub(result, numberToRat(v))
	}
	f, _ := result.Float64()
	return f
}

// Mul 精确乘法，返回 num1 与 others 中所有数的连乘积（对应 hutool mul(double...)）。
// Mul(1.1, 3) = 3.3 而非 3.3000000000000003；others 为空时返回 num1 本身。
// 场景：单价 × 数量、汇率换算。
func (t *utilNumber) Mul(num1 float64, others ...float64) float64 {
	result := numberToRat(num1)
	for _, v := range others {
		result.Mul(result, numberToRat(v))
	}
	f, _ := result.Float64()
	return f
}

// Div 精确除法（对应 hutool div(double, double, scale)，HALF_UP 四舍五入）。
// scale >= 0 时结果四舍五入到 scale 位小数（hutool 默认除法即保留 10 位）；
// scale < 0 时保留完整精度（此处与 hutool 取绝对值的语义不同，按完整精度处理）。
// divisor == 0 时返回 0：hutool 底层 BigDecimal.divide 会抛 ArithmeticException，
// Go 侧按惯例返回零值，由调用方自行判断除数。
// 场景：均值计算、比例分摊，如 Div(10, 3, 2) = 3.33。
func (t *utilNumber) Div(dividend, divisor float64, scale int) float64 {
	if divisor == 0 {
		return 0
	}
	quotient := new(big.Rat).Quo(numberToRat(dividend), numberToRat(divisor))
	if scale >= 0 {
		quotient = numberRoundRat(quotient, scale, numberRoundHalfUp)
	}
	f, _ := quotient.Float64()
	return f
}

// CeilDiv 向上取整除法（对应 hutool ceilDiv，补 JDK8 只有 floorDiv 没有 ceilDiv 的空缺）。
// CeilDiv(7, 2) = 4，CeilDiv(-7, 2) = -3；divisor == 0 返回 0。
// 这里用纯整数运算实现（hutool 借助 double），大数值无精度损失。
// 场景：分页计算总页数、按批次拆分任务数。
func (t *utilNumber) CeilDiv(dividend, divisor int64) int64 {
	if divisor == 0 {
		return 0
	}
	q := dividend / divisor
	// 余数非零且符号与除数一致时商加一（即商朝正无穷方向取整）
	if dividend%divisor != 0 && (dividend > 0) == (divisor > 0) {
		q++
	}
	return q
}

/* ---------------- 舍入与格式化 ---------------- */

// Round 保留 scale 位小数，四舍五入 HALF_UP（对应 hutool round(double, scale)）。
// scale < 0 按 0 处理。内部用 big.Rat 从 float64 的最短十进制串构造后舍入，
// 因此 Round(0.615, 2) = 0.62（纯 math.Round*pow 方案会因 0.615 的二进制近似得到 0.61）。
// 场景：金额展示位数规整、评分计算。
func (t *utilNumber) Round(num float64, scale int) float64 {
	f, _ := numberRoundRat(numberToRat(num), scale, numberRoundHalfUp).Float64()
	return f
}

// RoundStr 同 Round，但返回十进制字符串，不经过 float64 再转换，无尾数噪音。
// RoundStr(123.456789, 2) = "123.46"，RoundStr(5, 2) = "5.00"。
// 场景：直接输出到报文、页面或落库的定长小数字符串。
func (t *utilNumber) RoundStr(num float64, scale int) string {
	return numberRoundRat(numberToRat(num), scale, numberRoundHalfUp).FloatString(scale)
}

// RoundHalfEven 四舍六入五成双（银行家舍入，对应 hutool roundHalfEven）：
// 五后非零就进一，五后皆零看奇偶，五前为偶舍去、五前为奇进一。
// RoundHalfEven(2.5, 0) = 2，RoundHalfEven(3.5, 0) = 4。
// 场景：统计学/财务修约，避免大量 .5 数据单向偏差。
func (t *utilNumber) RoundHalfEven(num float64, scale int) float64 {
	f, _ := numberRoundRat(numberToRat(num), scale, numberRoundHalfEven).Float64()
	return f
}

// RoundDown 保留 scale 位小数，多余位数直接舍去（对应 hutool roundDown，朝零截断）。
// RoundDown(123.456789, 2) = 123.45，RoundDown(-123.456, 2) = -123.45。
// 场景：佣金、手续费计算宁少勿多。
func (t *utilNumber) RoundDown(num float64, scale int) float64 {
	f, _ := numberRoundRat(numberToRat(num), scale, numberRoundDown).Float64()
	return f
}

// DecimalFormatMoney 金额千分位格式化，每三位逗号分隔并保留两位小数，HALF_UP
// （对应 hutool decimalFormatMoney，格式 "#,##0.00"）。
// DecimalFormatMoney(355565.115) = "355,565.12"，DecimalFormatMoney(-1234.5) = "-1,234.50"。
// 场景：报表、页面展示金额。
func (t *utilNumber) DecimalFormatMoney(num float64) string {
	rounded := numberRoundRat(numberToRat(num), 2, numberRoundHalfUp)
	// 舍入结果可能被约分（如 35556512/100 会约成 8889128/25），
	// 因此乘 100 后取整得到以"分"计的带符号整数，再拆分整数与小数部分
	cents := new(big.Rat).Mul(rounded, new(big.Rat).SetInt64(100))
	k := new(big.Int).Quo(cents.Num(), cents.Denom())
	neg := k.Sign() < 0
	abs := new(big.Int).Abs(k)
	frac := new(big.Int).Mod(abs, big.NewInt(100)).Int64()
	intPart := new(big.Int).Div(abs, big.NewInt(100))
	s := numberGroupThousands(intPart) + "." + fmt.Sprintf("%02d", frac)
	if neg {
		s = "-" + s
	}
	return s
}

// FormatPercent 百分比格式化：ratio 先乘 100，再按 scale 位小数（HALF_UP）输出并以 % 结尾。
// FormatPercent(0.5, 2) = "50.00%"，FormatPercent(0.12345, 1) = "12.3%"。
// 场景：进度、占比展示。
func (t *utilNumber) FormatPercent(ratio float64, scale int) string {
	r := new(big.Rat).Mul(numberToRat(ratio), big.NewRat(100, 1))
	return numberRoundRat(r, scale, numberRoundHalfUp).FloatString(scale) + "%"
}

// ToStr 浮点数转字符串并去除小数末尾多余的 0（对应 hutool toStr）。
// ToStr(3.14) = "3.14"，ToStr(5.0) = "5"（整数无小数点），ToStr(3.1400) = "3.14"。
// 非法值原样输出："NaN"、"+Inf"、"-Inf"。场景：对外输出可读数值，避免科学计数法。
func (t *utilNumber) ToStr(num float64) string {
	return strconv.FormatFloat(num, 'f', -1, 64)
}

/* ---------------- 校验与解析 ---------------- */

// IsNumber 是否为合法数字字符串（对应 hutool isNumber），支持：
// 十进制、0x 十六进制、科学计数法（1234E3）、正负号、Java 风格类型后缀（d/f/l 等）。
// 空串与纯空白返回 false。IsNumber("1.2E+3") = true，IsNumber("12a") = false。
// 场景：外部输入数值合法性初筛。
func (t *utilNumber) IsNumber(input string) bool {
	if strings.TrimSpace(input) == "" {
		return false
	}
	chars := []byte(input)
	sz := len(chars)
	hasExp := false
	hasDecPoint := false
	allowSigns := false
	foundDigit := false
	// 处理开头的正负号
	start := 0
	if chars[0] == '-' || chars[0] == '+' {
		start = 1
	}
	// 0x 开头按十六进制校验（不可能再是其它形式）
	if sz > start+1 && chars[start] == '0' && (chars[start+1] == 'x' || chars[start+1] == 'X') {
		if start+2 == sz {
			return false // 仅 "0x"
		}
		for i := start + 2; i < sz; i++ {
			c := chars[i]
			if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
				return false
			}
		}
		return true
	}
	sz-- // 最后一个字符可能是类型后缀（d/D/f/F/l/L），循环完单独检查
	i := start
	for i < sz || (i < sz+1 && allowSigns && !foundDigit) {
		c := chars[i]
		switch {
		case c >= '0' && c <= '9':
			foundDigit = true
			allowSigns = false
		case c == '.':
			if hasDecPoint || hasExp {
				return false // 两个小数点，或指数部分出现小数点
			}
			hasDecPoint = true
		case c == 'e' || c == 'E':
			if hasExp {
				return false // 两个 E
			}
			if !foundDigit {
				return false // E 前必须有数字
			}
			hasExp = true
			allowSigns = true
		case c == '+' || c == '-':
			if !allowSigns {
				return false // 只允许紧跟 E 之后
			}
			allowSigns = false
			foundDigit = false // E 后必须再有数字
		default:
			return false
		}
		i++
	}
	if i < len(chars) {
		c := chars[i]
		if c >= '0' && c <= '9' {
			return true // 无类型后缀，合法
		}
		if c == 'e' || c == 'E' {
			return false // E 不能是最后一个字符
		}
		if c == '.' {
			if hasDecPoint || hasExp {
				return false
			}
			return foundDigit // 结尾单独一个小数点，前面有数字即合法
		}
		if !allowSigns && (c == 'd' || c == 'D' || c == 'f' || c == 'F') {
			return foundDigit
		}
		if c == 'l' || c == 'L' {
			return foundDigit && !hasExp && !hasDecPoint // L 不允许与指数/小数同用
		}
		return false
	}
	// 走到这里说明以 E 或 E+/- 结尾，或没有任何数字
	return !allowSigns && foundDigit
}

// IsInteger 是否为合法整数字符串，支持正负号与 0x 十六进制（对应 hutool isInteger）。
// IsInteger("+123") = true，IsInteger("0x1A") = true，IsInteger("1.5") = false。
// 场景：入参类型预判、配置项校验。
func (t *utilNumber) IsInteger(input string) bool {
	if strings.TrimSpace(input) == "" {
		return false
	}
	s := input
	if s[0] == '+' || s[0] == '-' {
		s = s[1:]
	}
	if len(s) > 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X') {
		_, err := strconv.ParseInt(s[2:], 16, 0)
		return err == nil
	}
	_, err := strconv.ParseInt(input, 10, 0)
	return err == nil
}

// IsDouble 是否为合法小数/浮点数字符串（对应 hutool isDouble：可解析为浮点且必须含小数点）。
// IsDouble("1.5") = true，IsDouble("1e3") = false（无小数点，按整数对待），IsDouble("abc") = false。
// 场景：区分整数与小数入参。
func (t *utilNumber) IsDouble(input string) bool {
	if strings.TrimSpace(input) == "" {
		return false
	}
	if _, err := strconv.ParseFloat(input, 64); err != nil {
		return false
	}
	return strings.Contains(input, ".")
}

// ParseInt 解析字符串为 int，解析失败返回 defaultValue（对应 hutool parseInt(str, default)）。
// 支持 0x 十六进制与普通十进制；小数字符串按 hutool 语义截取整数部分（"123.56" -> 123）。
// 场景：表单/配置取值并兜底，如 ParseInt(str, 80) 端口默认 80。
func (t *utilNumber) ParseInt(input string, defaultValue int) int {
	v, ok := numberParseInt64(input)
	if !ok {
		return defaultValue
	}
	return int(v)
}

// ParseInt64 解析字符串为 int64，解析失败返回 defaultValue（对应 hutool parseLong(str, default)）。
// 规则同 ParseInt，适合雪花 ID、分（金额单位）等大数值。
func (t *utilNumber) ParseInt64(input string, defaultValue int64) int64 {
	v, ok := numberParseInt64(input)
	if !ok {
		return defaultValue
	}
	return v
}

// ParseFloat64 解析字符串为 float64，解析失败返回 defaultValue（对应 hutool parseDouble(str, default)）。
// 支持科学计数法。场景：接口返回的价格解析兜底。
func (t *utilNumber) ParseFloat64(input string, defaultValue float64) float64 {
	if strings.TrimSpace(input) == "" {
		return defaultValue
	}
	if v, err := strconv.ParseFloat(input, 64); err == nil {
		return v
	}
	return defaultValue
}

/* ---------------- 数学运算 ---------------- */

// Factorial 计算阶乘 n! = n * (n-1) * ... * 2 * 1（对应 hutool factorial）。
// 返回 *big.Int，n 很大也不会溢出；n < 0 返回 0（hutool 此处抛异常），Factorial(0) = 1。
// 场景：排列组合计数、算法题辅助。
func (t *utilNumber) Factorial(num int64) *big.Int {
	if num < 0 {
		return big.NewInt(0)
	}
	result := big.NewInt(1)
	for i := int64(2); i <= num; i++ {
		result.Mul(result, big.NewInt(i))
	}
	return result
}

// Divisor 计算最大公约数（对应 hutool divisor）。
// Divisor(12, 8) = 4；任一参数为 0 时返回另一个值（gcd(m,0)=m 约定，hutool 此处会除零异常）。
// 场景：约分、按份数取公共步长。
func (t *utilNumber) Divisor(m, n int64) int64 {
	if m == 0 {
		return n
	}
	if n == 0 {
		return m
	}
	if m < 0 {
		m = -m
	}
	if n < 0 {
		n = -n
	}
	for n != 0 {
		m, n = n, m%n
	}
	return m
}

// Multiple 计算最小公倍数（对应 hutool multiple），m 或 n 为 0 返回 0。
// Multiple(4, 6) = 12。场景：周期任务对齐、分数通分。
func (t *utilNumber) Multiple(m, n int64) int64 {
	if m == 0 || n == 0 {
		return 0
	}
	return m / t.Divisor(m, n) * n
}

// IsPrimes 判断是否为质数（素数）。
// n <= 1 返回 false（hutool 要求 n > 1 否则断言异常，Go 侧统一返回 false 不 panic）。
// 场景：哈希表容量选取、数学计算。
func (t *utilNumber) IsPrimes(num int) bool {
	if num <= 1 {
		return false
	}
	if num <= 3 {
		return true
	}
	if num&1 == 0 || num%3 == 0 {
		return false
	}
	// 6k±1 试除，i*i <= num 用除法写法避免乘法溢出
	for i := 5; i <= num/i; i += 6 {
		if num%i == 0 || num%(i+2) == 0 {
			return false
		}
	}
	return true
}

// PartValue 把 total 平均分成 partCount 份，返回每份的个数（对应 hutool partValue）。
// plusOneWhenHasRem 为 true 且除不尽时每份 +1；partCount <= 0 返回 0。
// 注意：+1 后每份之和可能超过 total（余数不够分），实际是最后几份小于该值。
// 场景：批量任务按机器数分配，宁可多分配保证覆盖，如 PartValue(10, 3, true) = 4。
func (t *utilNumber) PartValue(total, partCount int64, plusOneWhenHasRem bool) int64 {
	if partCount <= 0 {
		return 0
	}
	part := total / partCount
	if plusOneWhenHasRem && total%partCount > 0 {
		part++
	}
	return part
}

// IsPowerOfTwo 判断 n 是否为 2 的幂（对应 hutool isPowerOfTwo）。
// n <= 0 返回 false。IsPowerOfTwo(1024) = true，IsPowerOfTwo(1000) = false。
// 场景：缓冲区/分片大小校验，位运算优化前置判断。
func (t *utilNumber) IsPowerOfTwo(n int64) bool {
	return n > 0 && n&(n-1) == 0
}

// IsOdd 是否为奇数（对应 hutool isOdd）。负数同样适用：IsOdd(-3) = true。
// 场景：奇偶分流、隔行变色。
func (t *utilNumber) IsOdd(n int64) bool {
	return n&1 == 1
}

// IsEven 是否为偶数（对应 hutool isEven）。IsEven(0) = true。
func (t *utilNumber) IsEven(n int64) bool {
	return !t.IsOdd(n)
}

/* ---------------- 中文数字 ---------------- */

// DigitToChinese 金额转中文大写（对应 hutool Convert.digitToChinese）。
// DigitToChinese(67556.32) = "陆万柒仟伍佰伍拾陆元叁角贰分"，
// DigitToChinese(1024.00) = "壹仟零贰拾肆元整"（节内补零规则与 hutool 一致），
// DigitToChinese(0) = "零元整"，负数以"负"开头。
// 支持范围 ±99999999999999.99（万亿级），超出返回空串（hutool 此处抛异常）。
// 实现用 big.Rat 从字符串构造再放大 100 倍取整到"分"，避免浮点乘法误差。
// 场景：发票、收据、合同金额大写。
func (t *utilNumber) DigitToChinese(num float64) string {
	if num == 0 {
		return "零元整"
	}
	if math.Abs(num) > 99_9999_9999_9999.99 {
		return ""
	}
	neg := num < 0
	if neg {
		num = -num
	}
	// 放大 100 倍取整得到总"分"数
	r := new(big.Rat).Mul(numberToRat(num), big.NewRat(100, 1))
	total := numberRoundRat(r, 0, numberRoundHalfUp).Num().Int64()
	fen := int(total % 10)
	total /= 10
	jiao := int(total % 10)
	total /= 10
	yuan := total

	var b strings.Builder
	if neg {
		b.WriteString("负")
	}
	// 金额模式下整数部分为 0 时无需"零元"
	if yuan != 0 {
		b.WriteString(numberLongToChinese(yuan))
		b.WriteString("元")
	}
	if jiao == 0 && fen == 0 {
		b.WriteString("整") // 无小数部分的金额以"整"结尾
		return b.String()
	}
	// 角：元与角都为 0 时（如 0.03）跳过角位，仅输出分
	if yuan != 0 || jiao != 0 {
		b.WriteString(numberChineseDigits[1][jiao])
		if jiao != 0 {
			b.WriteString("角")
		}
	}
	// 分
	if fen != 0 {
		b.WriteString(numberChineseDigits[1][fen])
		b.WriteString("分")
	}
	return b.String()
}

// ChineseToNumber 中文数字转 int64（对应 hutool chineseToNumber）。
// 支持 零一二三四五六七八九十百千万亿 与繁体 壹贰叁肆伍陆柒捌玖拾佰仟 及"两"，
// 如 "一百二十三" -> 123，"十二亿三千四百万" -> 1203400000，"两万二" -> 22000。
// 含非法字符或连续数字时返回 0（hutool 此处抛异常）。
// 场景：用户口述金额/数量录入、语音识别结果转换。
func (t *utilNumber) ChineseToNumber(chinese string) int64 {
	if strings.TrimSpace(chinese) == "" {
		return 0
	}
	return numberChineseToNumber([]rune(chinese))
}

// ChineseMoneyToNumber 中文大写金额转元（对应 hutool chineseMoneyToNumber）。
// 支持 元/圆、角、分 单位："陆万柒仟伍佰伍拾陆元叁角贰分" -> 67556.32，
// "壹佰贰拾元伍角" -> 120.5，"叁角贰分" -> 0.32。空串或非法输入返回 0。
// 场景：发票 OCR 结果还原为数值、财务对账。
func (t *utilNumber) ChineseMoneyToNumber(money string) float64 {
	if strings.TrimSpace(money) == "" {
		return 0
	}
	rs := []rune(money)
	idxOf := func(ch rune) int {
		for i, r := range rs {
			if r == ch {
				return i
			}
		}
		return -1
	}
	yi := idxOf('元')
	if yi < 0 {
		yi = idxOf('圆')
	}
	ji := idxOf('角')
	fi := idxOf('分')

	// 分别截取 元、角、分 对应的中文数字段
	var yStr, jStr, fStr []rune
	if yi > 0 {
		yStr = rs[:yi]
	}
	if ji > 0 {
		if yi >= 0 {
			if ji > yi { // 有元时角必在元之后
				jStr = rs[yi+1 : ji]
			}
		} else {
			jStr = rs[:ji] // 没有元，只有角
		}
	}
	if fi > 0 {
		if ji >= 0 {
			if fi > ji { // 有角时分必在角之后
				fStr = rs[ji+1 : fi]
			}
		} else if yi > 0 {
			if fi > yi { // 没有角有元，分在元之后
				fStr = rs[yi+1 : fi]
			}
		} else {
			fStr = rs[:fi] // 没有元、角，只有分
		}
	}

	var y, j, f int64
	if len(yStr) > 0 {
		y = numberChineseToNumber(yStr)
	}
	if len(jStr) > 0 {
		j = numberChineseToNumber(jStr)
	}
	if len(fStr) > 0 {
		f = numberChineseToNumber(fStr)
	}
	return float64(y) + float64(j)/10 + float64(f)/100
}

/* ---------------- 未导出辅助（number 前缀，避免与其它文件冲突） ---------------- */

// 舍入模式
const (
	numberRoundHalfUp   = iota // 四舍五入：舍弃部分 >= 0.5 时远离零进一
	numberRoundHalfEven        // 银行家舍入：五后皆零看奇偶，五前为偶舍去、为奇进一
	numberRoundDown            // 截断：直接舍弃多余小数位
)

// numberChineseDigits 单个数字的中文写法：[0] 简体（普通数字），[1] 繁体（金额大写），
// 下标 0~9 对应 零~玖。
var numberChineseDigits = [2][10]string{
	{"零", "一", "二", "三", "四", "五", "六", "七", "八", "九"},
	{"零", "壹", "贰", "叁", "肆", "伍", "陆", "柒", "捌", "玖"},
}

// numberChineseUnits 节内权位（个十百千）：[0] 简体，[1] 繁体（金额大写）。
var numberChineseUnits = [2][4]string{
	{"", "十", "百", "千"},
	{"", "拾", "佰", "仟"},
}

// numberChineseDigitMap 中文数字字符 -> 数值（含简体、繁体与口语"两"）
var numberChineseDigitMap = map[rune]int64{
	'零': 0, '一': 1, '壹': 1, '二': 2, '贰': 2, '两': 2, '三': 3, '叁': 3,
	'四': 4, '肆': 4, '五': 5, '伍': 5, '六': 6, '陆': 6, '七': 7, '柒': 7,
	'八': 8, '捌': 8, '九': 9, '玖': 9,
}

// numberChineseUnitMap 中文权位字符 -> 值与是否为节权位（万、亿与整个小节关联，与相邻数字无关）
var numberChineseUnitMap = map[rune]struct {
	value   int64
	secUnit bool
}{
	'十': {10, false}, '拾': {10, false},
	'百': {100, false}, '佰': {100, false},
	'千': {1000, false}, '仟': {1000, false},
	'万': {10000, true}, '亿': {1_0000_0000, true},
}

// numberToRat 把 float64 转为 big.Rat：先取其最短十进制表示（strconv 'f' 格式），
// 再按字符串构造，保证 0.1 得到 1/10 而非二进制近似值（等价 hutool new BigDecimal(String)）。
func numberToRat(v float64) *big.Rat {
	r, ok := new(big.Rat).SetString(strconv.FormatFloat(v, 'f', -1, 64))
	if !ok {
		// NaN/Inf 等不可解析值按 0 处理，避免向上传播异常
		return new(big.Rat)
	}
	return r
}

// numberPow10 返回 10^scale（scale >= 0）
func numberPow10(scale int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
}

// numberRoundRat 把有理数 r 舍入到 scale 位小数（scale < 0 按 0 处理），返回精确结果。
// 原理：r * 10^scale 后按指定模式取整（QuoRem 截断 + 余数判定），再除回 10^scale。
func numberRoundRat(r *big.Rat, scale int, mode int) *big.Rat {
	if scale < 0 {
		scale = 0
	}
	pow := numberPow10(scale)
	scaled := new(big.Rat).Mul(r, new(big.Rat).SetInt(pow))
	quo, rem := new(big.Int).QuoRem(scaled.Num(), scaled.Denom(), new(big.Int))
	// 比较 |rem| * 2 与分母：大于则进位（down 模式除外，无论余数多大都直接截断），
	// 等于则触发 HALF_UP / HALF_EVEN 的"五"判定
	twice := new(big.Int).Lsh(new(big.Int).Abs(rem), 1)
	switch cmp := twice.Cmp(scaled.Denom()); {
	case cmp > 0 && mode != numberRoundDown:
		quo = numberRoundStep(quo, scaled.Num().Sign())
	case cmp == 0 && mode == numberRoundHalfUp:
		quo = numberRoundStep(quo, scaled.Num().Sign())
	case cmp == 0 && mode == numberRoundHalfEven && quo.Bit(0) == 1:
		quo = numberRoundStep(quo, scaled.Num().Sign())
	}
	// numberRoundDown 走默认截断值
	return new(big.Rat).SetFrac(quo, pow)
}

// numberRoundStep 进位一步：正数 +1，负数 -1（HALF_UP 语义为远离零）
func numberRoundStep(quo *big.Int, sign int) *big.Int {
	if sign < 0 {
		return quo.Sub(quo, big.NewInt(1))
	}
	return quo.Add(quo, big.NewInt(1))
}

// numberGroupThousands 整数千分位分组：355565 -> "355,565"
func numberGroupThousands(intPart *big.Int) string {
	s := intPart.String()
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	b.Grow(len(s) + len(s)/3)
	head := len(s) % 3
	if head > 0 {
		b.WriteString(s[:head])
		b.WriteByte(',')
	}
	for i := head; i < len(s); i += 3 {
		b.WriteString(s[i : i+3])
		if i+3 < len(s) {
			b.WriteByte(',')
		}
	}
	return b.String()
}

// numberParseInt64 解析整数字符串：支持 0x 十六进制与十进制，小数截取整数部分
// （hutool 语义 "123.56" -> 123），失败返回 ok=false。
func numberParseInt64(input string) (int64, bool) {
	if strings.TrimSpace(input) == "" {
		return 0, false
	}
	lower := strings.ToLower(input)
	if strings.HasPrefix(lower, "0x") {
		if v, err := strconv.ParseInt(input[2:], 16, 64); err == nil {
			return v, true
		}
		return 0, false
	}
	if v, err := strconv.ParseInt(input, 10, 64); err == nil {
		return v, true
	}
	// 小数字符串截取整数部分
	if v, err := strconv.ParseFloat(input, 64); err == nil {
		return int64(v), true
	}
	return 0, false
}

// numberThousandToChinese 把 0~9999 的一节整数转为中文大写（金额用），0 返回"零"。
// 如 1024 -> "壹仟零贰拾肆"。
func numberThousandToChinese(part int64) string {
	if part == 0 {
		return "零"
	}
	segments := make([]string, 0, 8) // 从低位到高位逐位生成
	lastIsZero := true
	for i := 0; part > 0; i++ {
		digit := part % 10
		if digit == 0 {
			if !lastIsZero {
				// 前一位不是 0，则补一个"零"，如 1024 中间的 0
				segments = append(segments, numberChineseDigits[1][0])
			}
			lastIsZero = true
		} else {
			segments = append(segments, numberChineseDigits[1][digit]+numberChineseUnits[1][i])
			lastIsZero = false
		}
		part /= 10
	}
	var b strings.Builder
	for i := len(segments) - 1; i >= 0; i-- { // 反向拼接得到高位在前
		b.WriteString(segments[i])
	}
	return b.String()
}

// numberLongToChinese 把非负整数整数部分转为中文大写，支持到万亿级。
// 按"万"分节处理后逐节拼接，节间缺位补"零"（如 一亿零三百万），最终去掉开头多余的"零"。
func numberLongToChinese(amount int64) string {
	if amount == 0 {
		return "零"
	}
	var parts [4]int64 // 依次为 个节、万节、亿节、万亿节
	for i := 0; amount != 0; i++ {
		parts[i] = amount % 10000
		amount /= 10000
	}
	var s string
	addPreZero := func() { s = numberAddPreZero(s) }

	// 千（最低一节）
	if parts[0] > 0 {
		s = numberThousandToChinese(parts[0]) + s
		if parts[0] < 1000 {
			addPreZero() // 与万位之间隔 0，如 一万零三百
		}
	}
	// 万
	if parts[1] > 0 {
		if parts[1]%10 == 0 && parts[0] > 0 {
			addPreZero() // 万节个位为 0，如 十万零八千
		}
		s = numberThousandToChinese(parts[1]) + "万" + s
		if parts[1] < 1000 {
			addPreZero() // 与亿位之间隔 0，如 一亿零三百万
		}
	} else {
		addPreZero()
	}
	// 亿
	if parts[2] > 0 {
		if parts[2]%10 == 0 && parts[1] > 0 {
			addPreZero()
		}
		s = numberThousandToChinese(parts[2]) + "亿" + s
		if parts[2] < 1000 {
			addPreZero() // 与万亿位之间隔 0，如 一万亿零三百亿
		}
	} else {
		addPreZero()
	}
	// 万亿
	if parts[3] > 0 {
		if parts[2] == 0 {
			s = "亿" + s
		}
		s = numberThousandToChinese(parts[3]) + "万" + s
	}
	return strings.TrimPrefix(s, "零")
}

// numberAddPreZero 在字符串前补"零"（已以"零"开头或为空串时不补）
func numberAddPreZero(s string) string {
	if s == "" || strings.HasPrefix(s, "零") {
		return s
	}
	return "零" + s
}

// numberChineseToNumber 中文数字核心解析（hutool chineseToNumber 算法）。
// 按"节"（万/亿）累加：普通权位（十百千）与本节数字相乘累加，节权位（万亿）把
// 本节合计乘以权值并入结果；非法字符或连续数字返回 0。
func numberChineseToNumber(rs []rune) int64 {
	var result, section, number int64
	var unitValue int64
	hasUnit := false
	for i, c := range rs {
		if num, isDigit := numberChineseDigitMap[c]; isDigit {
			if num == 0 {
				// 遇到零：节结束，未决权位失效（如 "两百零一十" 中零后的处理）
				if number > 0 && hasUnit {
					section += number * (unitValue / 10)
				}
				hasUnit = false
			} else if number > 0 {
				return 0 // 连续两个数字，非法
			}
			number = num
		} else if u, isUnit := numberChineseUnitMap[c]; isUnit {
			if u.secUnit {
				// 节权位（万/亿）：整节求和后升级
				section = (section + number) * u.value
				result += section
				section = 0
			} else {
				unitNumber := number
				if number == 0 && i == 0 {
					// 单位开头默认置 1："十二" -> 一十二，"百二" -> 一百二
					unitNumber = 1
				}
				section += unitNumber * u.value
			}
			hasUnit = true
			unitValue = u.value
			number = 0
		} else {
			return 0 // 非法字符
		}
	}
	if number > 0 && hasUnit {
		// 末尾数字挂在前一个权位上："二十三" 中 3 挂在"十"上
		number *= unitValue / 10
	}
	return result + section + number
}
