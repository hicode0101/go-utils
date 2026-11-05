// example 目录：go-utils 各工具类的可运行用法示例。
//
// 运行方式（在仓库根目录执行）：
//
//	go run ./example
//
// 统一通过 utils 子包调用（包名与调用方式跟文档一致，仅 import 路径带 /utils）：
//
//	import "github.com/hicode0101/go-utils/utils"
//
// 各文件按工具类分组，覆盖每个单例对象的常用方法；完整方法清单见 README.md
// 或通过 `go doc github.com/hicode0101/go-utils/utils UtilStr` 等命令查阅。
// HTTP/外网相关示例失败时只打印错误，不影响其余演示。
package main

import (
	"fmt"

	"github.com/hicode0101/go-utils/utils"
)

func main() {

	fmt.Println(utils.Date.CurrentTime())

	stringDemo()     // String / Escape / Html
	numberDemo()     // Number / Math / Page / Version / Bool / Char / Obj
	datetimeDemo()   // Date
	cryptoDemo()     // Crypto / Hex / Hash
	httpDemo()       // Http / Request / UserAgent
	idgenDemo()      // Id / Random / 构造器（Snowflake、Hashids、WeightRandom）
	codecDemo()      // Base32/58/62/85 / BCD / Caesar / Rot / Morse / Punycode / Radix
	validDemo()      // Valid / Idcard / Phone / Desensitized
	cronDemo()       // Cron / CronPattern / Scheduler
	networkDemo()    // Net / Url
	storageDemo()    // File / Zip
	basicDemo()      // Sys / Json / Bytes / Coord
	collectionDemo() // 包级泛型函数（集合、随机选元素、If/Swap 等）
	fmt.Println("\n=== 全部示例执行完毕 ===")
}

// banner 打印分组标题，便于在输出中定位某个工具类的示例。
func banner(title string) {
	fmt.Printf("\n======== %s ========\n", title)
}
