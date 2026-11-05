package main

import (
	"fmt"

	"github.com/hicode0101/go-utils/utils"
)

// codecDemo 编解码与古典密码：Base32 / Base58 / Base62 / Base85 / BCD / Caesar / Rot / Morse / Punycode / Radix。
func codecDemo() {
	banner("Base32")
	if s, err := utils.Base32.Decode(utils.Base32.Encode([]byte("go-utils"))); err == nil {
		fmt.Printf("Encode([]byte(\"go-utils\")) = %s -> Decode = %q\n", utils.Base32.Encode([]byte("go-utils")), string(s))
	}
	fmt.Println(`EncodeHex([]byte("go")) =`, utils.Base32.EncodeHex([]byte("go")))

	banner("Base58（Bitcoin 字母表）")
	enc58 := utils.Base58.Encode([]byte("hello world")) // 对照黄金值 StV1DL6CwTryKyV
	if back, err := utils.Base58.Decode(enc58); err == nil {
		fmt.Printf("Encode([]byte(\"hello world\")) = %s -> Decode = %q\n", enc58, string(back))
	}
	checked := utils.Base58.EncodeChecked(0x00, []byte("hello"))
	if back, err := utils.Base58.DecodeChecked(checked, true); err == nil {
		fmt.Printf("EncodeChecked(0x00, \"hello\") = %s -> DecodeChecked = %q\n", checked, string(back))
	}

	banner("Base62")
	enc62 := utils.Base62.Encode([]byte("go-utils"))
	if back, err := utils.Base62.Decode(enc62); err == nil {
		fmt.Printf("Encode([]byte(\"go-utils\")) = %s -> Decode = %q\n", enc62, string(back))
	}
	fmt.Println(`EncodeInverted([]byte("go")) =`, utils.Base62.EncodeInverted([]byte("go")), "（反转字母表）")

	banner("Base85（Ascii85 / Z85）")
	// 黄金值：Encode([]byte("Man ")) 应为 "9jqo^"，Encode([]byte("hello")) 应为 "BOu!rDZ"
	fmt.Printf("Encode([]byte(\"Man \")) = %s | Encode([]byte(\"hello\")) = %s\n",
		utils.Base85.Encode([]byte("Man ")), utils.Base85.Encode([]byte("hello")))
	z85, err := utils.Base85.EncodeZ85([]byte("12345678"))
	if err == nil {
		back, _ := utils.Base85.DecodeZ85(z85)
		fmt.Printf("EncodeZ85([]byte(\"12345678\")) = %s -> DecodeZ85 = %q\n", z85, string(back))
	}

	banner("BCD 码")
	bcd, err := utils.BCD.StrToBcd("13800138000")
	if err == nil {
		fmt.Printf("StrToBcd(\"13800138000\") = % x -> BcdToStr = %s\n", bcd, utils.BCD.BcdToStr(bcd))
	}

	banner("Caesar 凯撒密码")
	fmt.Println(`Encode("Hello, World", 3) =`, utils.Caesar.Encode("Hello, World", 3),
		"| Decode =", utils.Caesar.Decode(utils.Caesar.Encode("Hello, World", 3), 3))

	banner("Rot 移位密码")
	fmt.Println(`Encode13("GoLand") =`, utils.Rot.Encode13("GoLand", false),
		"| Decode13 =", utils.Rot.Decode13(utils.Rot.Encode13("GoLand", false), false))
	fmt.Println(`Encode("abc123", 5, true)（数字同移）=`, utils.Rot.Encode("abc123", 5, true))

	banner("Morse 摩斯密码")
	morse := utils.Morse.Encode("SOS")
	fmt.Println(`Encode("SOS") =`, morse, "| Decode =", utils.Morse.Decode(morse))

	banner("Punycode 中文域名")
	encPuny := utils.Punycode.EncodeDomain("赵新虎.com") // 对照黄金值 xn--efvz93e52e.com
	if back, err := utils.Punycode.DecodeDomain(encPuny); err == nil {
		fmt.Printf("EncodeDomain(\"赵新虎.com\") = %s -> DecodeDomain = %q\n", encPuny, back)
	}
	fmt.Println(`Encode("中文", true) =`, utils.Punycode.Encode("中文", true))

	banner("Radix 自定义字母表进制转换")
	code34, _ := utils.Radix.Encode(utils.Radix34, 20260101)
	back34, _ := utils.Radix.DecodeToInt(utils.Radix34, code34)
	fmt.Printf("Encode(Radix34, 20260101) = %s -> DecodeToInt = %d\n", code34, back34)
	code59, _ := utils.Radix.Encode(utils.Radix59, 20260101)
	fmt.Printf("Encode(Radix59, 20260101) = %s（也可用自定义字母表）\n", code59)
}
