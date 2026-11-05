package main

import (
	"fmt"

	"github.com/hicode0101/go-utils/utils"
)

// numberDemo 数值与基础类型：Number / Math / Page / Version / Bool / Char / Obj。
func numberDemo() {
	banner("Number 数值计算（big.Rat 精确四则运算）")
	fmt.Println("Add(0.1, 0.2) =", utils.Number.Add(0.1, 0.2), "| Sub(1, 0.9) =", utils.Number.Sub(1, 0.9))
	fmt.Println("Mul(1.1, 3) =", utils.Number.Mul(1.1, 3), "| Div(10, 3, 4) =", utils.Number.Div(10, 3, 4))
	fmt.Println("Round(3.14159, 2) =", utils.Number.Round(3.14159, 2),
		"| RoundHalfEven(2.5, 0) =", utils.Number.RoundHalfEven(2.5, 0),
		"| RoundDown(3.99, 1) =", utils.Number.RoundDown(3.99, 1))
	fmt.Println("DecimalFormatMoney(1234567.891) =", utils.Number.DecimalFormatMoney(1234567.891))
	fmt.Println("FormatPercent(0.628, 2) =", utils.Number.FormatPercent(0.628, 2))
	fmt.Println(`IsNumber("123") =`, utils.Number.IsNumber("123"), `| IsInteger("1.5") =`, utils.Number.IsInteger("1.5"))
	fmt.Println(`ParseInt("abc", 0)（非法回退默认值）=`, utils.Number.ParseInt("abc", 0))
	fmt.Println("Factorial(10) =", utils.Number.Factorial(10))
	fmt.Println("Divisor(12, 18) =", utils.Number.Divisor(12, 18), "| Multiple(12, 18) =", utils.Number.Multiple(12, 18))
	fmt.Println("IsPrimes(17) =", utils.Number.IsPrimes(17), "| IsOdd(3) =", utils.Number.IsOdd(3),
		"| IsPowerOfTwo(8) =", utils.Number.IsPowerOfTwo(8))
	fmt.Println("CeilDiv(10, 3) =", utils.Number.CeilDiv(10, 3), "| PartValue(101, 10, true) =", utils.Number.PartValue(101, 10, true))
	fmt.Println("DigitToChinese(1024.53) =", utils.Number.DigitToChinese(1024.53))
	fmt.Println("ChineseToNumber(\"壹仟零贰拾肆\") =", utils.Number.ChineseToNumber("壹仟零贰拾肆"))
	fmt.Println("ChineseMoneyToNumber(\"壹仟零贰拾肆元伍角叁分\") =", utils.Number.ChineseMoneyToNumber("壹仟零贰拾肆元伍角叁分"))

	banner("Math 数学计算")
	fmt.Println("Abs(-3.5) =", utils.Math.Abs(-3.5), "| AbsInt(-7) =", utils.Math.AbsInt(-7))
	fmt.Println("CeilToInt(3.2) =", utils.Math.CeilToInt(3.2), "| FloorToInt(3.8) =", utils.Math.FloorToInt(3.8),
		"| RoundToInt(3.5) =", utils.Math.RoundToInt(3.5))
	fmt.Println("RoundFloat(3.14159, 2) =", utils.Math.RoundFloat(3.14159, 2))
	fmt.Println("PageCount(101, 20) =", utils.Math.PageCount(101, 20))

	banner("Page 分页")
	start, end := utils.Page.TransToStartEnd(3, 10)
	fmt.Println("TransToStartEnd(3, 10) -> start =", start, "end =", end)
	fmt.Println("GetStart(3, 10) =", utils.Page.GetStart(3, 10), "| GetEnd(3, 10) =", utils.Page.GetEnd(3, 10))
	fmt.Println("TotalPage(101, 20) =", utils.Page.TotalPage(101, 20))
	fmt.Println("Rainbow(8, 20, 5) =", utils.Page.Rainbow(8, 20, 5))

	banner("Version 版本号比较")
	fmt.Println(`Compare("1.2.0", "1.10.0") =`, utils.Version.Compare("1.2.0", "1.10.0"))
	fmt.Println(`GreaterThan("2.0.0", "1.9.9") =`, utils.Version.GreaterThan("2.0.0", "1.9.9"))
	fmt.Println(`MatchEl("1.5.0", ">1.0,<=2.0", ",") =`, utils.Version.MatchEl("1.5.0", ">1.0,<=2.0", ","))
	fmt.Println(`AnyMatch("2.0.0", "1.x", "2.0.0") =`, utils.Version.AnyMatch("2.0.0", "1.x", "2.0.0"))

	banner("Bool 布尔工具")
	fmt.Println(`ToBoolExt("yes") =`, utils.Bool.ToBoolExt("yes"), `| ToBoolExt("开") =`, utils.Bool.ToBoolExt("开"))
	fmt.Println("And(true, true, false) =", utils.Bool.And(true, true, false),
		"| Or =", utils.Bool.Or(true, false), "| Xor =", utils.Bool.Xor(true, true))
	fmt.Println("ToStringTrueFalse(true) =", utils.Bool.ToStringTrueFalse(true),
		"| ToStringOnOff =", utils.Bool.ToStringOnOff(false), "| ToStringYesNo =", utils.Bool.ToStringYesNo(true))

	banner("Char 字符判断")
	fmt.Println("IsAscii('a') =", utils.Char.IsAscii('a'), "| IsLetterUpper('A') =", utils.Char.IsLetterUpper('A'))
	fmt.Println("IsNumber('9') =", utils.Char.IsNumber('9'), "| IsHexChar('f') =", utils.Char.IsHexChar('f'))
	fmt.Println("EqualsIgnoreCase('A', 'a') =", utils.Char.EqualsIgnoreCase('A', 'a'))
	fmt.Println("IsEmoji('😀') =", utils.Char.IsEmoji('😀'), "| IsBlankChar(' ') =", utils.Char.IsBlankChar(' '))

	banner("Obj 对象工具")
	var p *int
	fmt.Println("IsNil(p) =", utils.Obj.IsNil(p), "| IsEmpty(\"\") =", utils.Obj.IsEmpty(""))
	fmt.Println("Equal([]int{1,2}, []int{1,2}) =", utils.Obj.Equal([]int{1, 2}, []int{1, 2}))
}
