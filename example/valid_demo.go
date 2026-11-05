package main

import (
	"fmt"
	"time"

	"github.com/hicode0101/go-utils/utils"
)

// validDemo 校验与脱敏：Valid / Idcard / Phone / Desensitized。
func validDemo() {
	banner("Valid 格式校验")
	fmt.Println(`IsEmail("user@example.com") =`, utils.Valid.IsEmail("user@example.com"))
	fmt.Println(`IsMobile("13612345678") =`, utils.Valid.IsMobile("13612345678"))
	fmt.Println(`IsCitizenId("33010219200403064x") =`, utils.Valid.IsCitizenId("33010219200403064x"))
	fmt.Println(`IsZipCode("100000") =`, utils.Valid.IsZipCode("100000"),
		`| IsBirthday("19900307") =`, utils.Valid.IsBirthday("19900307"))
	fmt.Println(`IsPlateNumber("粤A12345") =`, utils.Valid.IsPlateNumber("粤A12345"),
		`| IsCarVin("LFV3A23C8K3123456") =`, utils.Valid.IsCarVin("LFV3A23C8K3123456"))
	fmt.Println(`IsMac("00-16-EA-AE-3C-40") =`, utils.Valid.IsMac("00-16-EA-AE-3C-40"),
		`| IsIpv4("192.168.1.1") =`, utils.Valid.IsIpv4("192.168.1.1"),
		`| IsIpv6("::1") =`, utils.Valid.IsIpv6("::1"))
	fmt.Println(`IsMoney("12.34") =`, utils.Valid.IsMoney("12.34"),
		`| IsChineseName("张三丰") =`, utils.Valid.IsChineseName("张三丰"),
		`| HasChinese("你好 go") =`, utils.Valid.HasChinese("你好 go"))
	code := utils.Valid.RandomCreditCode()
	fmt.Println("RandomCreditCode() =", code, "| IsCreditCode 验证自身 =", utils.Valid.IsCreditCode(code))

	banner("Idcard 身份证")
	id := "33010219200403064x" // hutool 官方测试样本（15 位 330102200403064 升位而来）
	fmt.Println(`IsValidCard(id) =`, utils.Idcard.IsValidCard(id),
		`| IsValidCard15("330102200403064") =`, utils.Idcard.IsValidCard15("330102200403064"))
	if id18, err := utils.Idcard.Convert15To18("330102200403064"); err == nil {
		fmt.Println(`Convert15To18("330102200403064") =`, id18)
	}
	fmt.Println(`GetBirthByIdCard(id) =`, utils.Idcard.GetBirthByIdCard(id),
		"| GetAgeByIdCard(id, now) =", utils.Idcard.GetAgeByIdCard(id, time.Now()))
	fmt.Println("GetGenderByIdCard(id) =", utils.Idcard.GetGenderByIdCard(id), "（1 男 2 女）",
		"| GetProvinceByIdCard(id) =", utils.Idcard.GetProvinceByIdCard(id))
	fmt.Println(`Hide(id, 6, 14) =`, utils.Idcard.Hide(id, 6, 14), "（[start, end) 区间打码）")
	if info := utils.Idcard.GetIdcardInfo(id); info != nil {
		fmt.Printf("GetIdcardInfo(id) -> Province=%s City=%s District=%s Gender=%d\n",
			info.Province, info.City, info.District, info.Gender)
	}

	banner("Phone 电话号码")
	mobile := "13612345678"
	fmt.Println(`IsMobile(mobile) =`, utils.Phone.IsMobile(mobile),
		`| IsTel400800("400-860-8608") =`, utils.Phone.IsTel400800("400-860-8608"))
	fmt.Println(`HideBefore(mobile) =`, utils.Phone.HideBefore(mobile),
		"| HideBetween =", utils.Phone.HideBetween(mobile),
		"| HideAfter =", utils.Phone.HideAfter(mobile))
	fmt.Println(`SubBefore(mobile) =`, utils.Phone.SubBefore(mobile),
		"| SubBetween =", utils.Phone.SubBetween(mobile),
		"| SubAfter =", utils.Phone.SubAfter(mobile))

	banner("Desensitized 脱敏打码")
	fmt.Println(`ChineseName("张三丰") =`, utils.Desensitized.ChineseName("张三丰"))
	fmt.Println(`IdCardNum(id, 6, 4) =`, utils.Desensitized.IdCardNum(id, 6, 4))
	fmt.Println(`MobilePhone("13612345678") =`, utils.Desensitized.MobilePhone("13612345678"))
	fmt.Println(`Email("test@example.com") =`, utils.Desensitized.Email("test@example.com"))
	fmt.Println(`BankCard("6222020200112233445") =`, utils.Desensitized.BankCard("6222020200112233445"))
	fmt.Println(`Address("北京市海淀区西二旗街道100号", 8) =`,
		utils.Desensitized.Address("北京市海淀区西二旗街道100号", 8))
	fmt.Println(`Password("123456789") =`, utils.Desensitized.Password("123456789"))
	fmt.Println(`CarLicense("粤A12345") =`, utils.Desensitized.CarLicense("粤A12345"))
	fmt.Println(`Ipv4("192.168.1.100") =`, utils.Desensitized.Ipv4("192.168.1.100"))
	fmt.Println(`FixedPhone("0571-88888888") =`, utils.Desensitized.FixedPhone("0571-88888888"))
}
