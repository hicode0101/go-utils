// Package utils 一个只依赖 Go 标准库的常用工具库，按功能域划分为若干工具对象。
//
// 通过包级别的单例对象调用各类工具方法（方法语义与命名大体对照 hutool 5.8.x，
// 便于从 Java 项目迁移时对照查阅）：
//
//	utils.String.ToInt("42")           // 字符串工具
//	utils.Number.Round(3.14159, 2)     // 数值工具
//	utils.Date.Format(time.Now())      // 日期时间工具
//	utils.Crypto.MD5("abc")            // 哈希与编码
//	utils.Crypto.AesCbcEncrypt(...)    // 对称加密
//	utils.Http.Get(url, nil, 0)        // HTTP 客户端
//	utils.IndexOf([]int{1, 2, 3}, 2)   // 集合操作（包级泛型函数）
//	utils.If(x > y, x, y)              // 泛型辅助（包级函数）
//
// 设计原则（软件重用）：
//   - 仅依赖标准库，无任何第三方依赖；
//   - 按功能域分组，单一职责，方法命名保持一致约定
//     （如 xxxE 后缀表示返回 error 的安全转换）；
//   - 能用泛型消除的重复实现全部泛型化（集合操作）；
//   - 全部导出方法附带中文注释与使用场景说明。
package utils

// 各功能域的工具对象。内部类型均为空结构体，无状态、并发安全，直接使用即可。
var (
	Sys          = new(utilSys)          // 系统与环境：环境变量、进程、三元表达式、跨平台判断
	Json         = new(utilJson)         // JSON 编解码
	String       = new(utilStr)          // 字符串：转换、截取、过滤、校验、随机串、命名风格
	Date         = new(utilDate)         // 日期时间：解析、格式化、差值、加减、日界月界周界季度
	Crypto       = new(utilCrypto)       // 加解密：哈希/HMAC/对称/AES/DES/RSA/签名/OTP/PBKDF2
	Math         = new(utilMath)         // 数学计算：取整、绝对值、分页计算
	Bytes        = new(utilBytes)        // 字节流：大小端定长编解码、UTF-16、C 风格定长字符串
	Http         = new(utilHttp)         // HTTP 客户端：GET/POST/下载、请求构建、URL 编解码
	File         = new(utilFile)         // 文件目录：读写、复制、遍历、路径处理
	Number       = new(utilNumber)       // 数值计算：精确四则运算、舍入、格式化、中文数字
	Bool         = new(utilBool)         // 布尔：扩展解析、逻辑运算、字符串化
	Char         = new(utilChar)         // 字符：ASCII/字母/数字/十六进制/emoji 判断、全半角
	Hash         = new(utilHash)         // 散列：FNV/ELF/JS 等经典字符串哈希与 Murmur3
	Hex          = new(utilHex)          // 十六进制：编码解码、颜色值转换
	Id           = new(utilId)           // ID 生成：UUID/ObjectId/NanoId/雪花 ID
	Random       = new(utilRandom)       // 随机：整数/字节/元素选取/加权随机/汉字
	Regex        = new(utilRegex)        // 正则：分组提取、删除、计数、全量匹配
	Valid        = new(utilValid)        // 校验：邮箱/手机号/身份证格式/车牌/统一社会信用代码
	Idcard       = new(utilIdcard)       // 身份证：15/18 位互转、校验、生日性别省份提取
	Phone        = new(utilPhone)        // 电话：手机号/座机校验、号段脱敏与截取
	Desensitized = new(utilDesensitized) // 脱敏：姓名/手机号/身份证/邮箱/银行卡等打码
	Escape       = new(utilEscape)       // 转义：XML/HTML4 实体、Unicode 转义
	Html         = new(utilHtml)         // HTML：标签清除/删除、转义反转义
	Net          = new(utilNet)          // 网络：IP 与整数互转、内网判断、本机 IP/MAC、端口探活
	Url          = new(utilUrl)          // URL：补全/规范化/查询串构建/DataURI
	Page         = new(utilPage)         // 分页：行号区间、总页数、彩虹页码
	Version      = new(utilVersion)      // 版本号：比较、表达式匹配
	Zip          = new(utilZip)          // 压缩：zip 打包解压、gzip/zlib 压缩解压
	Obj          = new(utilObj)          // 对象：判空、深度相等、默认值
	UserAgent    = new(utilUserAgent)    // User-Agent 解析：浏览器/引擎/操作系统/平台
	Cron         = new(utilCron)         // 定时任务：cron 表达式解析匹配 + 调度器门面
	Coord        = new(utilCoord)        // 坐标：WGS84/GCJ-02/BD-09/墨卡托互转
	Base32       = new(utilBase32)       // Base32 编解码
	Base58       = new(utilBase58)       // Base58 编解码（Bitcoin 字母表）
	Base62       = new(utilBase62)       // Base62 编解码（标准/反转字母表）
	Base85       = new(utilBase85)       // Base85 编解码（Ascii85/Z85）
	BCD          = new(utilBCD)          // BCD 码：字符串与压缩 BCD 字节互转
	Caesar       = new(utilCaesar)       // 凯撒密码
	Rot          = new(utilRot)          // ROT-N 移位密码
	Morse        = new(utilMorse)        // 摩斯密码
	Punycode     = new(utilPunycode)     // Punycode（RFC 3492）与中文域名 xn-- 转换
	Radix        = new(utilRadix)        // 自定义字母表进制转换
)

// NewSnowflake 创建雪花 ID 生成器（41 位时间戳 + 5 位数据中心 + 5 位机器 + 12 位序列）。
// 常用快捷方式：utils.Id.SnowflakeNextId()（使用默认单例）。
func NewSnowflake(workerId, datacenterId int64) *Snowflake {
	return newSnowflake(workerId, datacenterId)
}

// NewHashids 创建 Hashids 编解码器（把一组数字编码成短 ID，可加盐控制输出）。
func NewHashids(salt string, minLength int) *Hashids {
	return newHashids(salt, minLength, "")
}

// NewScheduler 创建独立的定时任务调度器（全局快捷方式见 utils.Cron）。
func NewScheduler(matchSecond bool) *Scheduler {
	return newScheduler(matchSecond)
}

// NewCronPattern 编译 cron 表达式（5/6/7 字段，支持 * ? , - / L 与 | 多表达式）。
// 出错时返回 error，由调用方决定处理方式。
func NewCronPattern(pattern string) (*CronPattern, error) {
	return newCronPattern(pattern)
}

// NewWeightRandom 创建加权随机选取器（权重可与元素一一对应）。
func NewWeightRandom[T any](weights []float64, items []T) *WeightRandom[T] {
	return newWeightRandom(weights, items)
}
