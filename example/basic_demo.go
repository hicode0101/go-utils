package main

import (
	"bytes"
	"fmt"

	"github.com/hicode0101/go-utils/utils"
)

// basicDemo 系统与基础能力：Sys / Json / Bytes / Coord。
func basicDemo() {
	banner("Sys 系统与环境")
	if err := utils.Sys.SetEnv("GO_UTILS_DEMO", "hello"); err == nil {
		fmt.Println(`SetEnv/GetEnv("GO_UTILS_DEMO") =`, utils.Sys.GetEnv("GO_UTILS_DEMO"))
		_ = utils.Sys.UnsetEnv("GO_UTILS_DEMO")
	}
	fmt.Println("GetPid() =", utils.Sys.GetPid(), "| Hostname() =", utils.Sys.Hostname())
	fmt.Println("IsWindows() =", utils.Sys.IsWindows(), "| IsLinux() =", utils.Sys.IsLinux(),
		"| IsMac() =", utils.Sys.IsMac())
	if out, err := utils.Sys.RunCmd("go", "version"); err == nil {
		fmt.Println("RunCmd(\"go\", \"version\") =", utils.String.Brief(out, 40))
	}
	// 超时控制：utils.Sys.RunCmdTimeout(5*time.Second, "go", "version")

	banner("Json JSON 编解码")
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	u := User{Name: "张三", Age: 25}
	fmt.Println("ToJsonString(u) =", utils.Json.ToJsonString(u))
	if pretty, err := utils.Json.ToPrettyJson(u); err == nil {
		fmt.Println("ToPrettyJson(u) =", string(pretty))
	}
	var back User
	if err := utils.Json.FromJson([]byte(`{"name":"李四","age":30}`), &back); err == nil {
		fmt.Printf("FromJson -> %+v\n", back)
	}
	fmt.Println(`Valid([]byte("{\"a\":1}")) =`, utils.Json.Valid([]byte(`{"a":1}`)))

	banner("Bytes 字节流（大端/小端定长编解码）")
	be := utils.Bytes.Int32ToBytes(20260101, true) // true=大端
	fmt.Printf("Int32ToBytes(20260101, true) = % x -> BytesToInt32 = %d\n", be, utils.Bytes.BytesToInt32(be, true))
	var buf bytes.Buffer
	utils.Bytes.WriteUint16(&buf, 0x1234)
	fmt.Printf("WriteUint16(0x1234) -> ReadUint16 = %#x\n", utils.Bytes.ReadUint16(buf.Bytes()))
	var sbuf bytes.Buffer
	utils.Bytes.WriteFixedString(&sbuf, 8, "go") // C 风格定长字符串，不足补 0
	fmt.Printf("WriteFixedString(8, \"go\") -> ReadFixedString = %q\n", utils.Bytes.ReadFixedString(sbuf.Bytes()))
	fmt.Printf("NumberToBytes(3.14, true) = % x\n", utils.Bytes.NumberToBytes(3.14, true))

	banner("Coord 坐标系转换")
	fmt.Println("OutOfChina(0, 0) =", utils.Coord.OutOfChina(0, 0))
	gcj := utils.Coord.Wgs84ToGcj02(116.404, 39.915) // 北京 WGS84 -> 火星坐标
	fmt.Printf("Wgs84ToGcj02(116.404, 39.915) = (%.6f, %.6f)\n", gcj.Lng, gcj.Lat)
	bd := utils.Coord.Gcj02ToBd09(gcj.Lng, gcj.Lat) // 火星 -> 百度
	fmt.Printf("Gcj02ToBd09(gcj) = (%.6f, %.6f)\n", bd.Lng, bd.Lat)
	mt := utils.Coord.Wgs84ToMercator(116.404, 39.915) // Web 墨卡托
	fmt.Printf("Wgs84ToMercator(116.404, 39.915) = (%.2f, %.2f)\n", mt.Lng, mt.Lat)
}
