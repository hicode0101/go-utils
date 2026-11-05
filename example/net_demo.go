package main

import (
	"fmt"

	"github.com/hicode0101/go-utils/utils"
)

// networkDemo 网络与 URL：Net / Url。
func networkDemo() {
	banner("Net 网络工具")
	if long, err := utils.Net.Ipv4ToLong("192.168.1.1"); err == nil {
		fmt.Println(`Ipv4ToLong("192.168.1.1") =`, long, `| LongToIpv4 =`, utils.Net.LongToIpv4(long))
	}
	fmt.Println(`IsInnerIP("10.0.0.5") =`, utils.Net.IsInnerIP("10.0.0.5"),
		`| IsInnerIP("8.8.8.8") =`, utils.Net.IsInnerIP("8.8.8.8"))
	fmt.Println(`IsInRange("192.168.1.100", "192.168.1.0/24") =`,
		utils.Net.IsInRange("192.168.1.100", "192.168.1.0/24"))
	fmt.Println(`HideIpPart("192.168.1.100") =`, utils.Net.HideIpPart("192.168.1.100"))
	fmt.Println("GetLocalIP() =", utils.Net.GetLocalIP(), "| GetMacAddress() =", utils.Net.GetMacAddress())
	if ips := utils.Net.LocalIps(); len(ips) > 0 {
		fmt.Println("LocalIps()[0] =", ips[0], "（共", len(ips), "个）")
	}
	fmt.Println("GetIpByHost(\"localhost\") =", utils.Net.GetIpByHost("localhost"))
	fmt.Println(`IsValidPort(65536) =`, utils.Net.IsValidPort(65536))
	if port, err := utils.Net.UsableLocalPort(9000); err == nil {
		fmt.Println("UsableLocalPort(9000) =", port, "（本机可用端口）")
	}
	fmt.Println(`ParseCookies("a=1; b=2") =`, utils.Net.ParseCookies("a=1; b=2"))

	banner("Url URL 处理")
	if u, err := utils.Url.CompleteUrl("https://a.com/path/", "img/x.png"); err == nil {
		fmt.Println(`CompleteUrl("https://a.com/path/", "img/x.png") =`, u)
	}
	fmt.Println(`Normalize("http://a.com/a/../b?x=1") =`, utils.Url.Normalize("http://a.com/a/../b?x=1"))
	fmt.Println(`BuildQuery({"b":"1","a":"2"}) =`,
		utils.Url.BuildQuery(map[string]string{"b": "1", "a": "2"}), "（按 key 排序，输出稳定）")
	fmt.Println(`EncodeBlank("https://a.com/a b.html") =`, utils.Url.EncodeBlank("https://a.com/a b.html"))
	fmt.Println(`Decode("a%20b+c") =`, utils.Url.Decode("a%20b+c"), "（路径语义：+ 不转空格）")
	fmt.Println(`DataUriText("hi") =`, utils.Url.DataUriText("hi"))
	fmt.Println(`GetHost("https://a.com:8080/x") =`, utils.Url.GetHost("https://a.com:8080/x"))
}
