// Package utils 一个只依赖 Go 标准库的常用工具库，按功能域划分为若干工具对象。
//
// 通过包级别的单例对象调用各类工具方法：
//
//	utils.String.ToInt("42")           // 字符串工具
//	utils.Date.Format(time.Now())      // 日期时间工具
//	utils.Http.Get(url, nil, 0)        // HTTP 客户端
//	utils.IndexOf([]int{1, 2, 3}, 2)   // 集合操作（包级泛型函数）
//
// 设计原则（软件重用）：
//   - 仅依赖标准库，无任何第三方依赖；
//   - 按功能域分组，单一职责，方法命名保持一致约定
//     （如 xxxE 后缀表示返回 error 的安全转换）；
//   - 能用泛型消除的重复实现全部泛型化（集合操作）；
//   - 全部导出方法附带中文注释与使用场景说明，详见各 readme.md。
package utils

// 各功能域的工具对象。内部类型均为空结构体，无状态、并发安全，直接使用即可。
var (
	Sys    = new(utilSys)    // 系统与环境：环境变量、进程、三元表达式、跨平台判断
	Json   = new(utilJson)   // JSON 编解码
	String = new(utilStr)    // 字符串：转换、截取、过滤、校验、随机串
	Date   = new(utilDate)   // 日期时间：解析、格式化、差值、加减、日界月界
	Crypto = new(utilCrypto) // 哈希与编码：MD5/SHA/HMAC/Base64/UUID
	Math   = new(utilMath)   // 数学计算：取整、绝对值、分页计算
	Bytes  = new(utilBytes)  // 字节流：小端序定长编解码、UTF-16、C 风格定长字符串
	Http   = new(utilHttp)   // HTTP 客户端：GET/POST/HEAD/下载、URL 编解码
	File   = new(utilFile)   // 文件目录：读写、复制、遍历、路径处理
)
