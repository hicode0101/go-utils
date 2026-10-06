# go-utils

只依赖 **Go 标准库**（零第三方依赖）的常用工具库，按功能域划分为 `String` / `Date` / `Crypto` / `Math` / `Bytes` / `Http` / `File` / `Json` / `Sys` 九个工具对象 + 一组**泛型集合函数**，全部方法带中文注释。

## 特性

- **零第三方依赖**：全部基于标准库实现，可放心引入任何项目
- **按功能域分组**：`utils.String`、`utils.Date`…… 一看即知去哪找方法
- **泛型集合操作**：`utils.IndexOf`、`utils.Filter`、`utils.Map` 等一套函数适用所有类型
- **命名约定一致**：`ToInt` / `ToIntE` 成对出现——带 `E` 后缀返回 error（安全转换），不带 `E` 的失败返回零值（便捷转换）

## 安装

```bash
go get github.com/hicode0101/go-utils
```

```go
import "github.com/hicode0101/go-utils"
```

## 快速上手

```go
package main

import (
	"fmt"
	"time"

	"github.com/hicode0101/go-utils"
)

func main() {
	// 字符串
	n := utils.String.ToInt("42")                     // 42
	s := utils.String.Truncate("你好世界欢迎", 5, "...") // "你好世界欢..."

	// 日期
	fmt.Println(utils.Date.CurrentTime())          // 2024-06-01 12:30:45
	next := utils.Date.AddDays(time.Now(), 7)      // 一周后

	// 集合（泛型包级函数）
	nums := []int{1, 2, 3, 2}
	uniq := utils.Unique(nums)                     // [1 2 3]

	// HTTP
	resp := utils.Http.Get("https://httpbin.org/get", nil, 0)
	if resp.IsOK() {
		fmt.Println(resp.BodyString())
	}

	_ = next
	_ = uniq
}
```

## 工具总览

| 成员 | 文件 | 职责 |
|---|---|---|
| `utils.String` | str.go | 类型转换、判断、截取、正则过滤、格式校验、随机串 |
| `utils.Date` | date.go | 解析、格式化、时间戳互转、差值、加减、日界月界 |
| `utils.Crypto` | crypto.go | MD5/SHA/HMAC、Base64、UUID |
| `utils.Math` | math.go | 取整、绝对值、分页数 |
| `utils.Bytes` | bytes.go | 小端序定长编解码、UTF-16、C 风格定长字符串 |
| `utils.Http` | http.go | GET/POST/HEAD/下载、URL 编解码 |
| `utils.File` | file.go | 读写、复制、遍历、路径处理 |
| `utils.Json` | json.go | JSON 编解码与校验 |
| `utils.Sys` | sys.go | 环境变量、进程、平台判断、命令执行 |
| 包级泛型函数 | collection.go | IndexOf / Contains / Filter / Map / Unique / Chunk / 集合运算 |
| 包级泛型函数 | sys.go | `utils.If`、`utils.Swap` |

---

# 一、String 字符串工具（utils.String）

## 1.1 类型转换

| 方法 | 说明 |
|---|---|
| `ToInt(str) int` | 转 int，失败返回 0 |
| `ToIntE(str) (int, error)` | 转 int，返回错误 |
| `ToInt64(str) int64` / `ToInt64E(str)` | 转 int64 |
| `ToFloat64(str) float64` / `ToFloat64E(str)` | 转 float64 |
| `ToBool(str) bool` / `ToBoolE(str)` | 转 bool（支持 1/t/TRUE/0/f/FALSE 等） |
| `IntToStr(num int) string` | int → 字符串 |
| `Int64ToStr(num int64) string` | int64 → 字符串 |
| `Float64ToStr(num float64, prec int) string` | 浮点 → 字符串，prec 为小数位（-1 表示最少位数） |

**使用场景示例**：

```go
utils.String.ToInt("42")          // 42
utils.String.ToInt("abc")         // 0（便捷：解析失败返回零值）
utils.String.ToIntE("abc")        // 0, error（安全：需要区分"值为0"与"失败"时用 E 版本）

utils.String.ToInt64("9007199254740993") // 大数（如 JS 端传来的雪花ID）不丢精度
utils.String.ToFloat64("3.14")    // 3.14
utils.String.ToBool("true")       // true

utils.String.IntToStr(42)         // "42"
utils.String.Int64ToStr(10001)    // "10001"
utils.String.Float64ToStr(3.14159, 2)  // "3.14"（金额展示保留两位）
utils.String.Float64ToStr(3.14159, -1) // "3.14159"（无损还原）
```

## 1.2 二进制与位运算

| 方法 | 说明 |
|---|---|
| `BinToInt(binStr) int64` | 二进制串 → int64，`"1010"` → 10 |
| `IntToBin(num) string` | int64 → 二进制串 |
| `HasBits(value, mask int64) bool` | value 是否包含 mask 的全部位 |
| `BinStrHasBits(binStr, mask) bool` | 二进制串形式上同上判断 |

**使用场景**：权限位/开关位校验（每个 bit 代表一项权限）。

```go
// 权限定义：1=读 2=写 4=删
perm := utils.String.BinToInt("1011")        // 11（读+写+删）
utils.String.HasBits(perm, 1)                // true  有读权限
utils.String.HasBits(perm, 2)                // true  有写权限
utils.String.HasBits(perm, 4)                // true  有删权限
utils.String.HasBits(0b1010, 1)              // false 没有读权限

// 标志位以二进制串存库时的判断
utils.String.BinStrHasBits("1011", 0b0010)   // true
```

## 1.3 判断与检查

| 方法 | 说明 |
|---|---|
| `IsEmpty(str) bool` | 长度为 0 |
| `IsNotEmpty(str) bool` | 非空 |
| `IsBlank(str) bool` | 去空白后为空（`" "`、`"\t\n"` 算空） |
| `IsNotBlank(str) bool` | 非 IsBlank |
| `TrimSpace(str) string` | 去首尾空白 |
| `Len(str) int` | 字节长度（中文按 UTF-8 字节数） |
| `RuneCount(str) int` | 字符个数 |

**使用场景**：入参校验中 `IsEmpty` 判空串、`IsBlank` 判"看起来有值其实全是空格"的用户输入；`Len` 对接数据库字节限制，`RuneCount` 做字数统计。

```go
utils.String.IsEmpty("")       // true
utils.String.IsBlank("   ")    // true
utils.String.IsBlank(" hi ")   // false

utils.String.Len("中国")       // 6（UTF-8 每个汉字 3 字节）
utils.String.RuneCount("中国") // 2

utils.String.TrimSpace("  hello  ") // "hello"
```

## 1.4 查找与处理

| 方法 | 说明 |
|---|---|
| `Contains(str, find) bool` | 是否包含子串 |
| `StartsWith(str, find) bool` | 是否以 find 开头 |
| `EndsWith(str, find) bool` | 是否以 find 结尾 |
| `Count(str, find) int` | 子串出现次数（非重叠） |
| `Index(str, find) int` | 首次出现下标，未找到 -1 |
| `LastIndex(str, find) int` | 最后一次出现下标 |
| `Replace(str, find, to)` | 只替换第一处 |
| `ReplaceAll(str, find, to)` | 替换全部 |
| `Split(str, sep) []string` | 按分隔符拆分 |
| `Join(strs, sep) string` | 用分隔符连接 |
| `SubStr(str, start, end) string` | **按字符**截取 [start, end)，越界自动收敛 |
| `ToLower / ToUpper(str)` | 大小写转换 |
| `Reverse(str) string` | 反转（按字符，中文安全） |
| `Repeat(str, count) string` | 重复 count 次 |
| `Truncate(str, maxLen, suffix) string` | 超长保留前 maxLen 字符加 suffix |
| `PadLeft / PadRight(str, length, pad)` | 左/右填充到定长 |

**使用场景示例**：

```go
utils.String.Contains("hello world", "world")       // true
utils.String.StartsWith("https://a.com", "https://") // true（协议判断）
utils.String.EndsWith("data.csv", ".csv")            // true（扩展名判断）
utils.String.Count("cheese", "e")                    // 3
utils.String.Index("ABC_xyz", "xyz")                 // 4

utils.String.ReplaceAll("a-b-c", "-", "")            // "abc"
utils.String.Split("a,b,c", ",")                     // ["a" "b" "c"]
utils.String.Join([]string{"a", "b"}, ",")           // "a,b"

// 截取：按字符截取，中文不会像字节截取那样出现乱码
utils.String.SubStr("你好世界", 0, 2)                // "你好"
utils.String.SubStr("abc", 2, 100)                   // "c"（越界自动收敛）

utils.String.ToLower("Love GoLang")                  // "love golang"
utils.String.ToUpper("love 中国")                    // "LOVE 中国"
utils.String.Reverse("中国abc")                      // "cba国中"
utils.String.Repeat("ab", 3)                         // "ababab"

// 列表摘要：正文最多显示 5 个字，超出加省略号
utils.String.Truncate("你好世界欢迎", 5, "...")      // "你好世界欢..."

// 定长编号补零
utils.String.PadLeft("7", 3, '0')                    // "007"
utils.String.PadRight("ab", 4, '*')                  // "ab**"
```

## 1.5 拼接与格式化

| 方法 | 说明 |
|---|---|
| `Concat(strs ...string) string` | 高效拼接多个字符串 |
| `ConcatAny(inputs ...interface{}) string` | 拼接任意类型 |
| `FormatBytes(size uint64) string` | 字节数 → 可读单位 |

**使用场景**：

```go
utils.String.Concat("SELECT * FROM t WHERE id=", id)         // 拼接已知字符串
utils.String.ConcatAny("id=", 1001, " ok=", true)            // "id=1001 ok=true"
utils.String.FormatBytes(1536)                               // "1.50 KB"
utils.String.FormatBytes(5 * 1024 * 1024)                    // "5.00 MB"
utils.String.FormatBytes(utils.File.Size("big.bin"))         // 文件大小展示
```

## 1.6 随机字符串

```go
// 场景：验证码、临时文件名、请求 ID（数字+字母，密码学安全随机源）
code := utils.String.RandomString(6)
tmpName := "tmp_" + utils.String.RandomString(8) + ".dat"
```

## 1.7 正则过滤（正则均已预编译，性能好）

| 方法 | 说明 |
|---|---|
| `FilterByRegex(expr, input, placeTo)` | 通用正则替换，表达式非法时原样返回 |
| `FilterStyle(input)` | 删除 `<style>...</style>` |
| `FilterScript(input)` | 删除 `<script>...</script>` |
| `FilterHtml(input)` | 删除全部 HTML 标签，保留文字 |
| `FilterA(input)` | 删除 `<a>`/`</a>` 标签（保留链接文字） |
| `FilterImage(input)` | 删除 `<img>` 标签 |
| `FilterSpecialChar(input)` | 删除 `[+=\|{}':;',]` |
| `FilterUrlPrefix(input)` | 删除协议前缀 `\w+://` |

**使用场景**：富文本正文清洗。

```go
html := `<html><head><style>body{}</style></head><body><p>你好</p><img src="x.png"><a href="/go">跳转</a></body></html>`

utils.String.FilterStyle(html)   // 去样式块
utils.String.FilterScript(html)  // 去脚本块（XSS 缓解的第一道工序，需配合转义使用）
utils.String.FilterImage(html)   // 无图模式
utils.String.FilterA(html)       // 去超链接但保留"跳转"文字
utils.String.FilterHtml(html)    // 全部去标签 → "你好跳转"

utils.String.FilterByRegex(`\d`, "a1b2c3", "") // "abc"（任意正则）
utils.String.FilterUrlPrefix("https://a.com")  // "a.com"（域名归一化）
utils.String.FilterSpecialChar(`a+b|c`)        // "abc"
```

## 1.8 格式校验

| 方法 | 说明 |
|---|---|
| `IsNumber(str) bool` | 非负整数（`^\d+$`） |
| `IsIP(str) bool` | 合法 IP（v4 或 v6，net.ParseIP 校验） |
| `IsIPv4(str) bool` / `IsIPv6(str) bool` | 分别校验 v4 / v6 |
| `IsDomain(str) bool` | 域名格式 |
| `IsDnsDomain(str) bool` | 严格 DNS 域名（每段 ≤63 位，允许根点） |
| `IsEmail(str) bool` | 邮箱格式（宽松规则） |
| `IsMobile(str) bool` | 中国大陆 11 位手机号 |
| `IsIDCard(str) bool` | 18 位身份证格式（末位可 X，仅格式不校验地区码） |
| `IsURL(str) bool` | http/https 链接 |

**使用场景**：API 入参校验、数据清洗前的格式过滤。

```go
utils.String.IsIPv4("192.168.1.1")       // true
utils.String.IsIPv4("999.1.1.1")         // false
utils.String.IsIPv6("fe80::1")           // true
utils.String.IsEmail("a.b-c@example.com")// true
utils.String.IsMobile("13812345678")     // true
utils.String.IsDomain("www.example.com") // true
utils.String.IsURL("https://example.com/a?b=1") // true
utils.String.IsNumber("12345")           // true；"-1"、"1.5" 均为 false
```

---

# 二、Date 日期时间工具（utils.Date）

> Go 格式化记忆锚点：`01/02 03:04:05PM '06` = 1234567，即 `"2006-01-02 15:04:05"` 对应 `yyyy-MM-dd HH:mm:ss`（Go 1.20+ 标准库常量 `time.DateTime` / `time.DateOnly` 即这两个格式）。

## 2.1 当前时间

| 方法 | 说明 |
|---|---|
| `Now() time.Time` | 当前本地时间 |
| `NowIn(locName) time.Time` | 指定 IANA 时区名的当前时间 |
| `NowShanghai() time.Time` | 北京时间（固定 UTC+8，不依赖系统 tzdata） |
| `GoBirthday() time.Time` | Go 参考时间（趣味方法） |

**使用场景**：报表按东八区统计但服务器在 UTC 时，用 `NowShanghai`；跨时区展示用 `NowIn`。

```go
utils.Date.Now()                          // 本地时间
utils.Date.NowIn("America/New_York")      // 纽约当前时间（非法时区名回退本地）
utils.Date.NowShanghai()                  // 始终是北京时间，无需系统时区配置
```

## 2.2 解析

| 方法 | 说明 |
|---|---|
| `Parse(s) (time.Time, error)` | 按 `2006-01-02 15:04:05` 解析（本地时区） |
| `MustParse(s) time.Time` | 同上，失败返回零值（`IsZero()` 判断） |
| `ParseDate(s) (time.Time, error)` | 按 `2006-01-02` 解析 |
| `ParseWith(s, layout) (time.Time, error)` | 按自定义 layout 解析 |

```go
t, err := utils.Date.Parse("2024-06-01 12:30:45")
t2 := utils.Date.MustParse("2024-06-01")        // 确定格式时的便捷写法
t3, _ := utils.Date.ParseWith("20240601123045", "20060102150405")
```

## 2.3 时间戳互转

| 方法 | 说明 |
|---|---|
| `FromUnix(sec) time.Time` | 秒级时间戳 → 时间 |
| `FromUnixMilli(ms) time.Time` | 毫秒级时间戳 → 时间（兼容 Java `currentTimeMillis()` / JS `Date.now()`） |
| `ToUnix(t) int64` | → 秒级时间戳 |
| `ToUnixMilli(t) int64` | → 毫秒级时间戳 |

**使用场景**：与前端/Java 系统交互的 13 位毫秒时间戳。

```go
t := utils.Date.FromUnixMilli(1717225845000) // 前端传来的毫秒时间戳
utils.Date.Format(t)                         // "2024-06-01 12:30:45"（本地时区）
utils.Date.ToUnixMilli(time.Now())           // 回传 13 位时间戳
```

## 2.4 格式化

| 方法 | 输出示例 |
|---|---|
| `Format(t)` | `2024-06-01 12:30:45` |
| `FormatNano(t)` | `2024-06-01 12:30:45.123456789` |
| `FormatCompact(t)` | `20240601123045`（纯数字，用于订单号/文件名） |
| `FormatDate(t)` | `2024-06-01` |
| `FormatDateCompact(t)` | `20240601` |
| `FormatWith(t, layout)` | 任意自定义格式 |
| `CurrentTime()` | 当前时间字符串 |
| `CurrentDate()` | 当前日期字符串 |

```go
now := time.Now()
utils.Date.Format(now)                   // 日志/展示最常用格式
utils.Date.FormatCompact(now)            // "NO" + FormatCompact(now) → 可排序订单号
utils.Date.FormatDateCompact(now)        // 按天切分的日志文件名 app.20240601.log
utils.Date.FormatWith(now, "2006年01月02日 15时04分")
```

## 2.5 比较与差值

| 方法 | 说明 |
|---|---|
| `IsBeforeNow(t) bool` | 是否早于当前时间（如优惠券已过期） |
| `IsAfterNow(t) bool` | 是否晚于当前时间（如活动未开始） |
| `IsSameDay(a, b) bool` | 是否同一年月日 |
| `DiffDays(a, b) int64` | 相差整天数（绝对值） |
| `DiffHours(a, b) int64` | 相差整小时数 |
| `DiffMinutes(a, b) int64` | 相差整分钟数 |
| `DiffSeconds(a, b) int64` | 相差整秒数 |
| `DiffMillis(a, b) int64` | 相差毫秒数 |

```go
utils.Date.IsBeforeNow(coupon.ExpireAt)      // 优惠券过期判断
utils.Date.IsSameDay(lastLogin, time.Now())  // 每日签到限一次

start := utils.Date.MustParse("2024-06-01 10:00:00")
end := utils.Date.MustParse("2024-06-02 10:30:00")
utils.Date.DiffDays(start, end)     // 1
utils.Date.DiffHours(start, end)    // 24
utils.Date.DiffMinutes(start, end)  // 1470
```

## 2.6 时间加减

```go
now := time.Now()
utils.Date.AddSecs(now, 30)     // 30 秒后（验证码过期点）
utils.Date.AddMins(now, -5)     // 5 分钟前
utils.Date.AddHours(now, -24)   // 24 小时前（替代旧版 Before24h）
utils.Date.AddDays(now, 7)      // 一周后（会员到期时间）
utils.Date.AddMonths(now, 1)    // 下月同日（注意 1/31 + 1 月会被 Go 归一）
utils.Date.AddYears(now, 1)     // 一年后（证书有效期）
```

所有 `Add*` 方法参数为负数即减法。

## 2.7 日界与月界

| 方法 | 说明 |
|---|---|
| `StartOfDay(t)` / `EndOfDay(t)` | 当天 00:00:00.000000000 / 23:59:59.999999999（t 的时区） |
| `StartOfDayIn(t, loc)` / `EndOfDayIn(t, loc)` | 指定时区的当天起止 |
| `StartOfMonth(t)` / `EndOfMonth(t)` | 当月 1 号 0 点 / 月末 23:59:59.999999999（自动处理大小月与闰年） |

**使用场景**：统计区间查询、活动截止时间。

```go
now := time.Now()
// "今日已产生的数据"
rows := queryOrder(utils.Date.StartOfDay(now), now)
// 活动截止到今晚 23:59:59
deadline := utils.Date.EndOfDay(now)
// 本月账单区间
monthStart, monthEnd := utils.Date.StartOfMonth(now), utils.Date.EndOfMonth(now) // 6月 → 2024-06-30
```

## 2.8 耗时统计

```go
func handler() {
	defer utils.Date.TimeCost(time.Now()) // 一行统计函数耗时：TimeCost： 128.4ms
	...
}
```

---

# 三、Crypto 哈希与编码（utils.Crypto）

> ⚠️ MD5/SHA1 属弱哈希，仅用于校验/兼容；口令存储请用 bcrypt/argon2，接口签名推荐 `HmacSha256`。

| 方法 | 说明 |
|---|---|
| `MD5(input) string` | 32 位小写 MD5 |
| `MD5WithSalt(input, salt) string` | 拼盐后取 MD5 |
| `MD5File(path) (string, error)` | 流式计算文件 MD5（大文件友好） |
| `SHA1(input) string` | 40 位 SHA1 |
| `SHA256(input) string` | 64 位 SHA256 |
| `HmacSha256(key, data) string` | HMAC-SHA256（接口签名） |
| `Base64Encode / Base64Decode` | 标准 Base64（Decode 返回 error） |
| `Base64UrlEncode / Base64UrlDecode` | URL 安全 Base64（`-` `_` 替代 `+` `/`） |
| `Uuid() string` | UUID v4（36 位，crypto/rand） |

**使用场景示例**：

```go
utils.Crypto.MD5("abc")                    // "900150983cd24fb0d6963f7d28e17f72"（缓存键、校验）
utils.Crypto.MD5WithSalt("password", "s1") // 兼容旧口令体系
utils.Crypto.MD5File("big.bin")            // 上传秒传判断 / 下载完整性校验

utils.Crypto.SHA256(payload)               // 数据指纹
utils.Crypto.HmacSha256(secret, body)      // 开放接口签名（放 Authorization 头）

utils.Crypto.Base64Encode("hello")         // "aGVsbG8="
utils.Crypto.Base64UrlEncode(rawToken)     // 放 URL / Cookie 不会被转义截断

utils.Crypto.Uuid()                        // "a1b2c3d4-e5f6-4xxx-9xxx-xxxxxxxxxxxx"（请求ID、去重键）
```

---

# 四、Math 数学计算（utils.Math）

> Go 1.21+ 内置了 `min` / `max`（支持任意有序类型与多参数），本库不再提供 MaxInt/MinInt。

| 方法 | 说明 |
|---|---|
| `Abs(num float64) float64` | 浮点绝对值 |
| `AbsInt(num int) int` / `AbsInt64` | 整数绝对值 |
| `CeilToInt(num) int` | 向上取整：5.3 → 6 |
| `FloorToInt(num) int` | 向下取整：5.9 → 5 |
| `RoundToInt(num) int` | 四舍五入：5.5 → 6 |
| `RoundFloat(num, prec) float64` | 保留 prec 位小数 |
| `PageCount(total, size) int64` | 总页数（向上取整，size/total 非正返回 0） |

**使用场景示例**：

```go
utils.Math.AbsInt(-7)                 // 7
utils.Math.CeilToInt(10.0 / 3)        // 4（10 个任务每人 3 个需要 4 人）
utils.Math.FloorToInt(5.9)            // 5
utils.Math.RoundToInt(5.5)            // 6
utils.Math.RoundFloat(3.14159, 2)     // 3.14（展示前归一化；精确金额请用整数分或 decimal）

utils.Math.PageCount(101, 20)         // 6（分页接口的 pages 字段）
utils.Math.PageCount(0, 20)           // 0（旧版 Pages 对 size=0 会因除零产生异常大数，已修复）
```

---

# 五、Bytes 字节流编解码（utils.Bytes）

统一**小端序**。写入方法参数为 `*bytes.Buffer`（其 Write 永不报错，故不返回 error）；读取方法**不做越界检查**，长度不足会 panic，解析报文前请先校验长度。典型场景：解析设备/硬件/网络协议的二进制定长报文。

## 5.1 定长数值读写

| 写入 | 读取 | 字节数 |
|---|---|---|
| `WriteUint8` / `WriteInt8` | `ReadUint8` / `ReadInt8` | 1 |
| `WriteUint16` / `WriteInt16` | `ReadUint16` / `ReadInt16` | 2 |
| `WriteUint32` / `WriteInt32` | `ReadUint32` / `ReadInt32` | 4 |
| `WriteUint64` / `WriteInt64` | `ReadUint64` / `ReadInt64` | 8 |
| `WriteFloat32` / `WriteFloat64` | `ReadFloat32` / `ReadFloat64` | 4 / 8 |

**使用场景**：组装/解析定长协议报文。

```go
// 组包：命令字(1B) + 长度(2B) + 毫秒时间戳(8B)
var buf bytes.Buffer
utils.Bytes.WriteUint8(&buf, 0x01)
utils.Bytes.WriteUint16(&buf, uint16(payloadLen))
utils.Bytes.WriteInt64(&buf, utils.Date.ToUnixMilli(time.Now()))

// 解包（先校验总长度！）
cmd := utils.Bytes.ReadUint8(head[0:1])
length := utils.Bytes.ReadUint16(head[1:3])
ts := utils.Bytes.ReadInt64(head[3:11])
```

## 5.2 结构体整体编解码

```go
type msgHead struct {
	Magic uint16
	Len   uint32
}

// 整段报文头一次性编解码（定长结构体）
utils.Bytes.WriteAny(&buf, msgHead{Magic: 0xAA55, Len: 128})
var h msgHead
_ = utils.Bytes.ReadAny(head[0:8], &h)
```

## 5.3 C 风格定长字符串

| 方法 | 说明 |
|---|---|
| `WriteFixedString(buf, size, s)` | 写入定长 `char[size]`：超长截断、不足补 0 |
| `ReadFixedString(b)` | 读取并去掉末尾补位的 0 |
| `WriteUtf16String(buf, size, s)` | 写入 UTF-16LE 宽字符 `WCHAR[size]`（size 为**字符数**，占 size*2 字节） |
| `ReadUtf16String(b)` | 按 UTF-16LE 解码（自动去尾部 0） |

**使用场景**：与 C/C++ 程序或中文工控设备的报文交互。

```go
var buf bytes.Buffer
utils.Bytes.WriteFixedString(&buf, 8, "ab")    // 61 62 00 00 00 00 00 00
utils.Bytes.ReadFixedString(buf.Bytes())       // "ab"

utils.Bytes.WriteUtf16String(&buf, 4, "你好")  // fd 59 7d 4f 00 00 00 00
utils.Bytes.ReadUtf16String(buf.Bytes())       // "你好"（旧版 WriteUnicodeTCHAR 只支持 ASCII，已重写）
```

---

# 六、Http 客户端（utils.Http）

所有请求方法**统一返回 `*HttpResp`**（错误记录在 `Msg` 字段），`timeout` 传 0/负数使用默认 30 秒，底层复用同一连接池。

```go
type HttpResp struct {
	Status  int         // 状态码；请求未发出时为 -1（网络错误）或 -500（本地构造失败）
	Msg     string      // 错误信息，成功为空
	Body    []byte      // 响应体
	Headers http.Header // 响应头
}

func (r *HttpResp) BodyString() string // 响应体转字符串
func (r *HttpResp) IsOK() bool         // 状态码 2xx
```

| 方法 | 说明 |
|---|---|
| `Get(url, headers, timeout) *HttpResp` | GET 请求 |
| `Head(url, headers, timeout) *HttpResp` | HEAD 请求（结果看 Headers） |
| `PostForm(url, form, headers, timeout) *HttpResp` | 表单 POST（自动 URL 编码） |
| `PostJson(url, jsonBody, headers, timeout) *HttpResp` | JSON POST |
| `Download(url, savePath, headers, timeout) error` | 流式下载到文件 |
| `UrlEncode(s) / UrlDecode(s)` | URL 查询参数编解码 |

**使用场景示例**：

```go
// 1. GET：headers 可为 nil；timeout 传 0 用默认 30s
resp := utils.Http.Get("https://api.example.com/v1/user?id=1", nil, 0)
if resp.IsOK() {
	var user User
	_ = utils.Json.FromJson(resp.Body, &user)
} else {
	// 失败时错误信息在 resp.Msg，可配合 go-logger 记录：
	// logger.ErrorWithErr("查询用户失败", errors.New(resp.Msg))
}

// 2. GET 带请求头（如 Token）
resp = utils.Http.Get(url, map[string]string{"Authorization": "Bearer " + token}, 10*time.Second)
// 失败时错误信息在 resp.Msg，可配合 go-logger 记录：
// logger.ErrorWithErr("查询用户失败", errors.New(resp.Msg))

// 3. 表单 POST（键值自动编码，等价手工 url.Values）
resp = utils.Http.PostForm("https://a.com/login",
	map[string]string{"username": "tom", "password": "p@ss word"}, // p@ss 自动编码为 p%40ss
	nil, 0)

// 4. JSON POST（配合 Json 工具生成请求体）
body, _ := utils.Json.ToJsonString(map[string]any{"sku": "A001", "num": 2})
resp = utils.Http.PostJson("https://a.com/order", string(body),
	map[string]string{"X-Token": tk}, 5*time.Second)

// 5. HEAD 探测资源大小
head := utils.Http.Head(fileUrl, nil, 0)
size := head.Headers.Get("Content-Length")

// 6. 流式下载（大文件不占内存；非 2xx 不落盘并返回错误）
if err := utils.Http.Download(installerUrl, "/tmp/app.zip", nil, time.Minute); err != nil {
	// 处理错误
}

// 7. 手工拼查询串时对参数值转义
q := "kw=" + utils.Http.UrlEncode("golang & 泛型") // kw=golang+%26+%E6%B3%9B%E5%9E%8B
```

---

# 七、File 文件目录（utils.File）

路径处理统一使用 `filepath`，Windows/Unix 路径均正确。

## 7.1 打开与读写

| 方法 | 说明 |
|---|---|
| `Create(name) (*os.File, error)` | 创建/截断文件 |
| `Open(name) (*os.File, error)` | 只读打开 |
| `OpenFile(name, flag, perm)` | 指定 flag/权限打开 |
| `Read(name) ([]byte, error)` | 一次性读取全部字节 |
| `ReadText(name) (string, error)` | 一次性读取为字符串 |
| `Write(name, data) error` | 写入（覆盖） |
| `WriteText(name, s) error` | 写入字符串（覆盖） |
| `Append(name, data) error` | 追加（不存在则创建） |
| `AppendText(name, line) error` | 追加一行（自动补换行） |
| `Copy(src, dst) error` | 复制文件（流式，保留权限） |

```go
// 读小配置文件
content, _ := utils.File.ReadText("config.json")

// 导出文本报告
_ = utils.File.WriteText("report.txt", content)

// CSV 逐行追加
_ = utils.File.AppendText("data.csv", "id,name,amount")

// 逐行处理大文件（不一次性载入内存）
f, _ := utils.File.Open("huge.log")
defer f.Close()
scanner := bufio.NewScanner(f)
for scanner.Next() { ... }

// 备份
_ = utils.File.Copy("app.yaml", "app.yaml.bak")
```

## 7.2 存在性与信息

```go
utils.File.Exists("config.yaml")  // 启动自检
info, _ := utils.File.Stat("data.bin")
info.ModTime()                    // 判断文件是否被更新
utils.File.Size("data.bin")       // 字节数（目录或不存在返回 -1）
utils.String.FormatBytes(uint64(utils.File.Size("data.bin"))) // "1.50 KB"
```

## 7.3 目录操作

| 方法 | 说明 |
|---|---|
| `MkdirAll(path, perm) error` | 递归创建目录 |
| `ListFiles(dir, ext) ([]string, error)` | 递归列出文件；ext 如 `.go`（大小写不敏感），空串=全部 |
| `Rename(old, new) error` | 重命名/移动 |
| `Remove(name) error` | 删除文件或空目录 |
| `RemoveAll(path) error` | 递归删除（不可逆，勿拼用户输入） |

```go
_ = utils.File.MkdirAll("logs/2024/06", 0755)

// 批量处理目录下所有 CSV
files, _ := utils.File.ListFiles("./export", ".csv")
for _, fp := range files { process(fp) }

// 临时文件写完后原子改名
_ = utils.File.Rename("report.txt.tmp", "report.txt")
```

## 7.4 路径处理

| 方法 | 说明 | 示例 |
|---|---|---|
| `Join(elems ...string)` | 平台安全拼接 | `Join("logs","06","a.log")` |
| `Dir(fullPath)` | 取目录 | `"/app/src/main.go"` → `"/app/src"` |
| `Name(fullPath)` | 取文件名（含扩展名） | → `"main.go"` |
| `NameWithoutExt(fullPath)` | 文件名去扩展名 | → `"main"` |
| `Ext(fullPath)` | 取扩展名（含点） | → `".go"` |
| `WorkDir()` | 当前工作目录 | |
| `ExePath()` | 当前可执行文件路径 | 定位随程序分发的资源 |

```go
// 以可执行文件所在目录为基准定位配置文件（不受启动目录影响）
exeDir := utils.File.Dir(utils.File.ExePath())
cfgPath := utils.File.Join(exeDir, "config.yaml")
```

---

# 八、Json（utils.Json）

| 方法 | 说明 |
|---|---|
| `ToJson(obj) ([]byte, error)` | 序列化为字节 |
| `ToJsonString(obj) string` | 紧凑字符串（失败返回 ""） |
| `ToPrettyJson(obj) ([]byte, error)` | 带缩进字节 |
| `ToPrettyJsonString(obj) string` | 带缩进字符串 |
| `FromJson(data, obj) error` | 反序列化（obj 传指针） |
| `Valid(data) bool` | 是否合法 JSON |

**使用场景示例**：

```go
type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}
u := User{Name: "Tom", Age: 18}

utils.Json.ToJsonString(u) // {"name":"Tom","age":18}（日志打印参数）
utils.Json.ToPrettyJsonString(u) // 带缩进，写配置/示例报文

var back User
_ = utils.Json.FromJson(resp.Body, &back) // 解析接口响应

utils.Json.Valid([]byte(`{"a":1}`)) // true，解析前预校验
```

---

# 九、Sys 系统与环境（utils.Sys）

| 方法 | 说明 |
|---|---|
| `GetEnv(key) string` | 取环境变量 |
| `SetEnv(key, value) error` | 设置环境变量 |
| `UnsetEnv(key) error` | 删除环境变量 |
| `GetPid() int` | 当前进程 ID |
| `Hostname() string` | 主机名 |
| `KillPid(pid) error` | 结束进程（Windows 仅限当前进程） |
| `IsWindows() / IsLinux() / IsMac() bool` | 平台判断 |
| `RunCmd(name, args...) (string, error)` | 执行命令，返回合并输出 |
| `RunCmdTimeout(timeout, name, args...)` | 带超时的命令执行 |

**使用场景示例**：

```go
env := utils.Sys.GetEnv("APP_ENV")                       // 读取运行环境
_ = utils.Sys.SetEnv("TZ", "Asia/Shanghai")

utils.Sys.IsWindows() // 按平台选择命令：windows 用 "cmd /c"，其它用 "sh -c"

// 调用外部工具
out, err := utils.Sys.RunCmd("git", "log", "-1", "--oneline")
// 偶发卡死的外部程序，兜底超时
out, err = utils.Sys.RunCmdTimeout(10*time.Second, "ffmpeg", "-i", "in.mp4", "out.avi")
```

## 泛型辅助：utils.If / utils.Swap（包级函数）

Go 没有三元运算符，且**泛型不能声明在方法上**（语言限制），因此这两个函数以包级泛型函数提供，不再需要旧版的类型断言：

```go
max := utils.If(x > y, x, y)              // 旧版 utils.Sys.If(x>y, x, y).(int)
name := utils.If(nick != "", nick, "匿名用户")

a, b := utils.Swap(1, 2)                  // a=2, b=1
```

---

# 十、Collection 集合操作（包级泛型函数）

> 同样因为"Go 方法不能带类型参数"，集合函数以**包级函数**提供（`utils.IndexOf(...)` 而非 `utils.Collection.IndexOf(...)`）。一套函数适用所有类型，替代旧版 `IntArrayFind` / `Int64ArrayFind` / `StrArrayContain` 等多份复制粘贴。

| 方法 | 说明 |
|---|---|
| `IndexOf(slice, value) int` | 首次出现下标，未找到 -1 |
| `Contains(slice, value) bool` | 是否包含 |
| `Filter(slice, keep) []T` | 按谓词过滤（新切片） |
| `Map(slice, fn) []R` | 元素映射（新切片） |
| `Unique(slice) []T` | 去重保序 |
| `Reverse(slice) []T` | 原地反转 |
| `RemoveAt(slice, index) []T` | 删除下标元素（越界原样返回） |
| `Chunk(slice, size) [][]T` | 分块 |
| `Union(slices ...[]T) []T` | 并集（去重保序） |
| `Intersect(a, b) []T` | 交集 |
| `Difference(a, b) []T` | 差集（在 a 不在 b） |

**使用场景示例**：

```go
nums := []int{1, 2, 3, 2}
words := []string{"go", "rust", "go"}

// 查找与包含（替代旧版 IntArrayFind/IntArrayContain/StrArrayContain）
utils.IndexOf(nums, 3)                        // 2
utils.Contains([]string{"GET", "POST"}, "PUT") // false
utils.Contains(words, "go")                   // true

// 过滤与映射
utils.Filter(nums, func(v int) bool { return v%2 == 0 })       // [2 2]
ids := []int64{1, 2, 3}
utils.Map(ids, func(id int64) string { return utils.String.Int64ToStr(id) }) // ["1" "2" "3"]

// 去重、反转、删除、分批
utils.Unique(nums)                 // [1 2 3]
utils.Reverse([]int{1, 2, 3})      // [3 2 1]
utils.RemoveAt([]string{"a","b","c"}, 1) // [a c]
utils.Chunk(bigSlice, 500)         // 每批 500 条分批写库

// 集合运算：多角色权限合并 / 共同权限 / 待删除数据
all := utils.Union(roleA, roleB)
common := utils.Intersect(roleA, roleB)
toDelete := utils.Difference(inDb, submitted)
```

---

# 从旧版本升级（破坏性变更对照）

## 命名与参数调整

| 旧版 | 新版 | 说明 |
|---|---|---|
| `utils.Match.*` | `utils.Math.*` | 组名拼写修正（Match→Math） |
| `String.ToIntegral` / `ToLong` | `ToIntE` / `ToInt64E` | 去重并统一 E 后缀约定 |
| `String.ToInt64`（旧版无错版） | 保持 | 新增 `ToInt64E` |
| `String.ToStr / FormatInt / FormatInt64` | `IntToStr` / `Int64ToStr` | 方向明确的命名 |
| `String.FormatFloat64(num)` | `Float64ToStr(num, prec)` | 小数位数参数化 |
| `String.IsEmptyStr` | `IsBlank` | 语义化命名 |
| `String.LinkStrs / LinkInputs` | `Concat / ConcatAny` | 语义化命名 |
| `String.FilterByRegex(expr, input, placeTo)` | 保持 | 表达式非法时不再 panic，改为原样返回 |
| `String.SubStr`（按字节） | `SubStr`（按字符） | 中文不再截出乱码，越界自动收敛 |
| `Collection.*`（IntArrayFind 等） | 包级 `IndexOf` / `Contains` 等 | 泛型化，Go 方法不支持类型参数 |
| `utils.Sys.If(cond, a, b) interface{}` | 包级 `utils.If[T]` | 泛型，去掉类型断言 |
| `utils.Sys.Swap` | 包级 `utils.Swap[T]` | 泛型 |
| `Date.GetTimeNow / GetTimeNowBySH / GetTimeNowByLoc / GetTimeNowByUTC8` | `Now` / `NowShanghai` / `NowIn` | 语义化 |
| `Date.ToTime / ToTimeFromDate / ToTimeByFm / GetTime` | `Parse / ParseDate / ParseWith / MustParse` | 语义化 |
| `Date.JavaLongTime / LongTime / Float64Time*` | `FromUnixMilli` | 三份重复实现合并（内部用标准库 `time.UnixMilli`） |
| `Date.FormatTimens / FormatTimeToNum / FormatTimeToDayNum` | `FormatNano / FormatCompact / FormatDateCompact` | 语义化 |
| `Date.DifferDays/Hour/Min/Sec/Milsec` | `DiffDays/Hours/Minutes/Seconds/Millis` | 拼写修正 |
| `Date.GetBeginTime / GetEndTime` | `StartOfDay / EndOfDay` | 语义化；新增 `*In` 时区版 |
| `Date.Before24h` | `Date.AddHours(now, -24)` | 泛化 |
| `Date.AddSecs/AddMins/AddHours` | 保持 | 实现改为直接 `time.Duration`（不再每次 ParseDuration） |
| `File.FileXxx` 系列（FileCreate/FileRead/FileIsExist…） | 去掉 File 前缀：`Create` / `Read` / `Exists`… | 组内方法无需重复前缀 |
| `File.MakeDir / FileDir / FileFullName / FileExt / GetCurrentDir / GetCurrentExe` | `MkdirAll` / `Dir` / `Name` / `Ext` / `WorkDir` / `ExePath` | 语义化 |
| `File` 的 `path` 包实现 | `filepath` | Windows 反斜杠路径不再出错 |
| `HttpResp.Datas` | `HttpResp.Body` | 命名修正 |
| `HttpResp.ToString` | `HttpResp.BodyString` | 语义化 |
| `Http.SimpleGet(url, t)` | `Http.Get(url, nil, t)` | 合并 API：headers 传 nil 即可 |
| `Http.SimplePost / SimpleHead` | `Http.PostForm(url, nil, nil, t)` / `Http.Head(url, nil, t)` | 合并 API |
| `Http.Post(url, headers, param, t)` | `Http.PostForm(url, form, headers, t)` | **不再自动注入 `curTime` 参数**；表单值自动 URL 编码 |
| `Http.HttpPostJson` | `Http.PostJson` | 统一命名 |
| `Http.Get/Post...` 返回 `(resp, err)` | 统一返回 `*HttpResp` | 错误在 `Msg` 字段 |
| `Bytes.WriteBYTE/WORD/DWORD/Int` | `WriteUint8/Uint16/Uint32/Int32` | 命名规范 |
| `Bytes.ReadWord/ReadDWord/ReadTCHAR` | `ReadUint16/ReadUint32/ReadFixedString` | 去重；`ReadFixedString` 会去掉尾部补位的 0 |
| `Bytes.WriteTCHAR / WriteUnicodeTCHAR` | `WriteFixedString / WriteUtf16String` | UTF-16 实现重写，完整支持中文 |
| `Bytes.BinaryReadAny / BinaryWriteAny` | `ReadAny / WriteAny` | 简化命名 |
| `Crypto.GetMd5 / GetSaltMD5` | `Crypto.MD5 / MD5WithSalt` | Go 缩写词大写惯例 |

## 行为修复

1. **Http 空指针崩溃**：旧版在 `client.Do` 出错时解引用 nil `resp`（`_resp.StatusCode`）直接 panic，新版统一返回 `Status=-1` 与 `Msg`。
2. **PostForm 不再注入 `curTime`**：旧版所有表单请求都被塞入 `curTime=<当前时间>` 参数，属于业务逻辑泄漏进工具库；如业务确需该参数请显式传入。
3. **PageCount 除零保护**：旧版 `Pages(10, 0)` 浮点除零产生异常大数，新版返回 0。
4. **FilterA 误杀修复**：旧版正则 `<.?a(.|\n)*?>` 会误删 `<abbr>`、`<data>` 等标签，新版 `</?a\b[^>]*>` 仅匹配 a 标签。
5. **正则预编译**：过滤类正则从"每次调用重新编译"改为包级预编译，且编译错误不再被忽略。
6. **移除死代码**：`Bytes._writeAny`（内含调试打印）、`Bytes._ReadInt_`、空 `init()` 等已删除。

## 新增方法速览

- **String**：`ToBool(E)`、`Float64ToStr(prec)`、`HasBits`、`IsBlank/IsNotBlank`、`RuneCount`、`LastIndex`、`Reverse`、`Repeat`、`Truncate`、`PadLeft/PadRight`、`RandomString`、`IsMobile`、`IsIDCard`、`IsURL`
- **Collection**：`Filter`、`Map`、`Unique`、`Reverse`、`RemoveAt`、`Chunk`、`Union`、`Intersect`、`Difference`（全部泛型）
- **Date**：`MustParse`、`FromUnix/ToUnix`、`ToUnixMilli`、`IsSameDay`、`AddYears`、`StartOfMonth/EndOfMonth`、`StartOfDayIn/EndOfDayIn`
- **Crypto**：`MD5File`、`SHA1`、`SHA256`、`HmacSha256`、`Base64*`、`Base64Url*`、`Uuid`
- **File**：`ReadText`、`WriteText`、`Append/AppendText`、`Copy`、`Size`、`ListFiles`、`NameWithoutExt`、`Join`
- **Json**：`ToJsonString`、`Valid`
- **Sys**：`UnsetEnv`、`Hostname`、`IsWindows/IsLinux/IsMac`、`RunCmd`、`RunCmdTimeout`

# 设计原则（软件重用）

1. **标准库优先**：全部功能基于标准库，`go.sum` 为空，无供应链风险；
2. **按域分组、单一职责**：一个文件一个功能域，方法通过 `utils.域.方法` 定位；
3. **消除重复**：能用泛型消除的类型特化实现全部泛型化；重复实现合并为单一实现（如毫秒时间戳三合一）；
4. **一致约定**：`E` 后缀 = 返回 error 的安全版本；同名方法在不同域语义一致；错误不静默吞噬；
5. **注释即文档**：每个导出方法的注释包含语义、边界行为（越界、失败、nil）与典型场景。
