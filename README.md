<div align="center">

# go-utils

**一个只依赖 Go 标准库的 Hutool 风格工具库 —— 34 个工具对象 · 800+ 方法 · 开箱即用**

[![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Coverage](https://img.shields.io/badge/Coverage-94.9%25-brightgreen)](#-测试)
[![Tests](https://img.shields.io/badge/Tests-448%20passed-success)](#-测试)
[![Dependencies](https://img.shields.io/badge/Dependencies-0-8A2BE2)](#-特性)
[![Reference](https://img.shields.io/badge/参考-Hutool%205.8-orange)](https://hutool.cn)

*命名与语义对照 [Hutool](https://hutool.cn) 5.8.x，Java 项目迁移 Go 时可无缝对照查阅*

</div>

---

## 📖 简介

`go-utils` 将 Hutool 最常用的工具能力（core / cron / crypto / http 四大模块）用纯 Go 标准库重新实现，保持「**单例对象 + 方法**」的调用习惯：

```go
utils.String.ToInt("42")                  // 字符串工具
utils.Crypto.AesCbcEncrypt(data, key, iv) // 加解密
utils.Cron.Schedule("0 0 12 * * ?", job)  // 定时任务
```

> **设计原则**
>
> - ✅ **零第三方依赖** —— 仅使用 Go 标准库，`go get` 即用，无供应链风险
> - ✅ **命名对照 hutool** —— 方法语义与命名大体对照 hutool 5.8.x，迁移成本最低
> - ✅ **统一返回值约定** —— `xxxE` 后缀返回 `error`（安全转换），不带后缀失败返回零值（便捷转换）
> - ✅ **泛型化集合操作** —— 集合/随机选取等以包级泛型函数提供，类型安全无断言
> - ✅ **全覆盖测试** —— 448 个测试函数、94.9% 语句覆盖率，关键算法经 RFC 向量与 hutool 官方黄金值校验

## 📦 安装

```bash
go get github.com/hicode0101/go-utils
```

```go
import "github.com/hicode0101/go-utils"
```

要求 **Go 1.25+**（使用了 `crypto/pbkdf2`、`math/rand/v2` 等较新标准库能力）。

## 🚀 快速入门

```go
package main

import (
	"fmt"

	"github.com/hicode0101/go-utils"
)

func main() {
	// 字符串：安全转换 + 截取 + 命名风格
	utils.String.ToInt("42")                 // 42（失败返回 0）
	utils.String.Sub("你好世界", 1, 3)         // "好世"（按字符，中文安全）
	utils.String.ToUnderlineCase("userName") // "user_name"
	utils.String.Hide("13812345678", 3, 7)   // "138****5678"

	// 数值：big.Rat 精确舍入，告别 0.615 二进制误差
	utils.Number.Round(0.615, 2)             // 0.62
	utils.Number.DecimalFormatMoney(355565.115) // "355,565.12"
	utils.Number.DigitToChinese(1024.53)     // "壹仟零贰拾肆元伍角叁分"

	// 加解密
	key, iv := []byte("0123456789abcdef"), []byte("fedcba9876543210")
	enc, _ := utils.Crypto.AesCbcEncrypt([]byte("hello"), key, iv)
	dec, _ := utils.Crypto.AesCbcDecrypt(enc, key, iv)
	fmt.Println(string(dec)) // hello
	utils.Crypto.MD5("abc")  // 900150983cd24fb0d6963f7d28e17f72

	// HTTP：一行调用
	resp := utils.Http.Get("https://httpbin.org/get", nil, 0)
	if resp.IsOK() {
		fmt.Println(resp.BodyString())
	}

	// 集合：泛型函数，类型安全
	nums := []int{3, 1, 2, 3}
	fmt.Println(utils.Unique(nums)) // [3 1 2]
	fmt.Println(utils.Chunk(nums, 2))

	// 定时任务：cron 表达式
	utils.Cron.Schedule("0/30 * * * * ?", func() { /* 每 30 秒 */ })
	utils.Cron.Start()
	defer utils.Cron.Stop()
}
```

## 📚 目录

- [✨ 特性](#-特性)
- [🧰 工具对象总览](#-工具对象总览)
- [📝 字符串 String](#-字符串-string)
- [🔢 数值 Number / Math](#-数值-number--math)
- [📅 日期时间 Date](#-日期时间-date)
- [🔐 加解密 Crypto](#-加解密-crypto)
- [🌐 HTTP 客户端 Http / Request / UserAgent](#-http-客户端-http--request--useragent)
- [🆔 ID 生成 Id / Random](#-id-生成-id--random)
- [✅ 校验 Valid / Idcard / Phone / Desensitized](#-校验-valid--idcard--phone--desensitized)
- [🔤 编解码 Hex / Hash / Codec](#-编解码-hex--hash--codec)
- [🕐 定时任务 Cron](#-定时任务-cron)
- [📡 网络 Net / Url](#-网络-net--url)
- [🗜️ 压缩 Zip](#-压缩-zip)
- [📂 文件 File / 系统 Sys / JSON / 字节流 Bytes](#-文件-file--系统-sys--json--字节流-bytes)
- [🧩 集合与泛型辅助](#-集合与泛型辅助)
- [🌍 更多能力](#-更多能力)
- [🧪 测试](#-测试)
- [⚖️ 与 hutool 的对照说明](#️-与-hutool-的对照说明)
- [⚠️ 注意事项](#️-注意事项)

## ✨ 特性

| | |
|---|---|
| 🧱 **纯标准库** | 无任何第三方依赖，交叉编译零负担 |
| 🇨🇳 **中文场景友好** | 中文数字转换、身份证/手机号校验、脱敏打码、全半角、GCJ-02 坐标系 |
| 🔐 **加密全家桶** | AES/DES/3DES/RC4/ChaCha20/XXTEA、RSA/ECDSA/Ed25519、TOTP/HOTP 动态口令 |
| ⏰ **内置调度器** | cron 表达式解析（含 `L` 月末、`\|` 多表达式）+ 边界对齐的秒/分钟级调度 |
| 🧵 **并发安全** | 全部单例对象无状态或内部加锁（Snowflake/Hashids/Scheduler） |
| 📖 **中文注释** | 每个导出方法附带场景说明与示例值 |

## 🧰 工具对象总览

| 调用入口 | 功能域 | 对照 hutool |
|---|---|---|
| `utils.String` | 字符串：转换/截取/命名风格/模板/打码 | StrUtil · NamingCase · StrFormatter |
| `utils.Number` · `utils.Math` | 精确舍入/千分位/中文数字 · 取整/分页 | NumberUtil · MathUtil |
| `utils.Date` | 解析/格式化/差值/加减/日周月季界/年龄/闰年 | DateUtil |
| `utils.Crypto` | 摘要/HMAC/对称/非对称/签名/OTP/PBKDF2 | SecureUtil · DigestUtil · RSA · Sign |
| `utils.Http` · `utils.NewGetRequest` | 快捷请求/链式构建器/参数编码 | HttpUtil · HttpRequest |
| `utils.UserAgent` | UA 解析（37 项浏览器识别表） | UserAgentUtil |
| `utils.Id` · `utils.Random` | UUID/ObjectId/NanoId/雪花 · 随机数/加权随机 | IdUtil · RandomUtil |
| `utils.Valid` · `utils.Idcard` · `utils.Phone` · `utils.Desensitized` | 校验/身份证/电话/脱敏 | Validator · IdcardUtil · PhoneUtil · DesensitizedUtil |
| `utils.Hex` · `utils.Hash` · `utils.Base32/58/62/85` · `utils.BCD` · `utils.Caesar/Rot/Morse/Punycode/Radix` | 编解码与散列 | HexUtil · HashUtil · codec 包 · RadixUtil |
| `utils.Regex` · `utils.Escape` · `utils.Html` | 正则提取/实体转义/标签清理 | ReUtil · EscapeUtil · HtmlUtil |
| `utils.Net` · `utils.Url` | IP 互转/内网判断/端口探活 · URL 补全/规范化 | NetUtil · URLUtil |
| `utils.Zip` · `utils.Coord` · `utils.Page` · `utils.Version` | 压缩/坐标/分页/版本比较 | ZipUtil · CoordinateUtil · PageUtil · VersionUtil |
| `utils.Cron` · `utils.NewScheduler` | cron 表达式 + 调度器 | hutool-cron 全模块 |
| `utils.File` · `utils.Sys` · `utils.Json` · `utils.Bytes` · `utils.Obj` · `utils.Bool` · `utils.Char` | 文件/系统/JSON/字节流/对象/布尔/字符 | FileUtil · RuntimeUtil · ByteUtil · ObjectUtil · BooleanUtil · CharUtil |

## 📝 字符串 String

```go
s := utils.String

// ── 转换：失败返回零值 / 带 E 返回 error ──
s.ToInt("42")               // 42
s.ToIntE("abc")             // error
s.Int64ToStr(1024)          // "1024"
s.Float64ToStr(3.14159, 2)  // "3.14"（prec=-1 表示最少位数）

// ── 截取：按字符，中文安全，支持负下标 ──
s.Sub("你好世界", 1, 3)           // "好世"
s.Sub("hello", -3, -1)           // "ll"（负数从末尾倒数，start>end 自动交换）
s.SubBefore("abcba", "b", false) // "a"（首个分隔符前）
s.SubAfter("abcba", "b", true)   // "a"（最后一个分隔符后）
s.SubBetween("wx[b]yz", "[", "]") // "b"

// ── 命名风格转换（hutool NamingCase）──
s.ToUnderlineCase("userName") // "user_name"
s.ToCamelCase("user_name")    // "userName"
s.ToPascalCase("user_name")   // "UserName"
s.ToKebabCase("userName")     // "user-name"

// ── 模板 ──
s.Format("{} 是 {} 岁", "张三", 18) // "张三 是 18 岁"
s.FormatWithMap("您好 {name}", map[string]string{"name": "张三"})

// ── 常用处理 ──
s.Hide("13812345678", 3, 7)   // "138****5678"
s.Brief("你好世界欢迎", 5)      // "你好..."（超长缩略）
s.UpperFirst("user")          // "User"
s.SplitTrim("a, b , ,c", ",") // ["a" "b" "c"]
s.ToSBC("aA1 ")               // "ａＡ１　"（半角转全角）
s.RandomString(16)            // 16 位字母数字随机串（密码学安全源）
s.Count("cheese", "e")        // 3
```

## 🔢 数值 Number / Math

```go
n := utils.Number

// big.Rat 精确计算，杜绝二进制浮点误差
n.Round(0.615, 2)        // 0.62（四舍五入）
n.RoundHalfEven(2.5, 0)  // 2（银行家舍入）
n.Add(0.1, 0.2)          // 0.3
n.Div(1, 3, 2)           // 0.33（精确除法+舍入）

// 格式化
n.DecimalFormatMoney(355565.115) // "355,565.12"
n.FormatPercent(0.5, 2)          // "50.00%"

// 中文数字（对照 hutool NumberChineseFormatter）
n.DigitToChinese(1024.53)        // "壹仟零贰拾肆元伍角叁分"
n.ChineseToNumber("一百二十三")     // 123
n.ChineseMoneyToNumber("壹佰贰拾元伍角") // 120.5

// 其它
n.IsPrimes(17)      // true
n.Divisor(12, 8)    // 4（最大公约数）
n.CeilDiv(10, 3)    // 4（向上取整除法）

utils.Math.PageCount(101, 20) // 6（总页数）
```

## 📅 日期时间 Date

```go
d := utils.Date

d.Format(time.Now())        // "2024-06-01 12:30:45"
d.FormatCompact(time.Now()) // "20240601123045"（可排序文件名/订单号）
tm, _ := d.Parse("2024-06-01 08:00:00")

d.AddDays(tm, 7)       // 一周后
d.StartOfDay(tm)       // 当天 00:00:00
d.EndOfMonth(tm)       // 当月最后一天 23:59:59.999...
d.BeginOfWeek(tm, true) // 本周一 00:00（true=周一为一周开始）
d.BeginOfQuarter(tm)   // 本季度第一天
d.DiffDays(tm, tm2)    // 相差天数
d.ToUnixMilli(tm)      // 13 位毫秒时间戳（对接 Java/JS）
d.IsLeapYear(2024)     // true
d.Age(birthday, time.Time{}) // 周岁（零值按当前时间）
```

## 🔐 加解密 Crypto

> ⚠️ MD5/SHA1 属弱哈希，仅用于校验与兼容；密码存储请用专业方案。

```go
c := utils.Crypto

// ── 摘要与 HMAC ──
c.MD5("abc")               // "900150983cd24fb0d6963f7d28e17f72"
c.SHA256("abc")            // 64 位小写 hex
c.HmacSha256("key", "data") // 开放接口签名

// ── 对称加密（PKCS7 填充，与 hutool/Java 互通）──
key, iv := []byte("0123456789abcdef"), []byte("fedcba9876543210")
enc, _ := c.AesCbcEncrypt(data, key, iv)
plain, _ := c.AesCbcDecrypt(enc, key, iv)
c.AesEcbEncrypt(data, key)              // AES/ECB（hutool 默认模式）
c.DesEncrypt(data, []byte("8bytekey"))  // 单 DES
c.Des3Encrypt(data, make([]byte, 24))   // 3DES
c.Rc4Crypt(data, []byte("Key"))         // RC4 对称异或

// ── 非对称（Base64 密钥与 hutool 生成的密钥互通）──
priv, pub, _ := c.GenerateRsaKeyPair(2048)
encRsa, _ := c.RsaEncrypt(data, pub)
plain, _ = c.RsaDecrypt(encRsa, priv)
sign, _ := c.RsaSignSha256(data, priv)
ok, _ := c.RsaVerifySha256(data, sign, pub) // true

// ── 动态口令（Google Authenticator 兼容）──
secret, uri, _ := c.GoogleSecretKey("alice@example.com", 20) // otpauth:// 链接
code, _ := c.Totp(secret, time.Now(), 6)
ok, _ = c.TotpValidate(secret, code, 1) // ±1 步长容错

// ── 接口参数签名（按 key 排序拼接后摘要）──
c.SignParamsMd5(map[string]string{"b": "2", "a": "1"}) // = MD5("a1b2")

// ── 编码 ──
c.Base64Encode("hello")     // "aGVsbG8="
c.Base64UrlEncode("a+b/")   // JWT 场景 → "a-b_"
c.Uuid()                    // 36 位 UUID v4
```

## 🌐 HTTP 客户端 Http / Request / UserAgent

```go
// ── 快捷方法：统一返回 *HttpResp（Status/Msg/Body/Headers）──
resp := utils.Http.Get("https://api.example.com/user", nil, 0)
if resp.IsOK() {
	fmt.Println(resp.BodyString())
}
utils.Http.PostForm(url, map[string]string{"user": "u1"}, nil, 0)
utils.Http.PostJson(url, `{"a":1}`, nil, 0)
utils.Http.Download("https://example.com/a.zip", "a.zip", nil, 0)

// ── 链式构建器（对照 hutool HttpRequest）──
resp = utils.NewGetRequest("https://api.example.com/data").
	Header("Authorization", "Bearer tk").
	Query("page", "1").
	Timeout(10 * time.Second).
	Execute()

resp = utils.NewPostRequest(url).
	FormMap(map[string]string{"name": "张三"}). // 表单
	BasicAuth("user", "pass").
	FollowRedirects(false).
	Execute()

// ── 参数与协议工具 ──
utils.Http.ToParams(map[string]string{"b": "x y", "a": "1"}) // "a=1&b=x+y"
utils.Http.UrlWithForm("http://a.com/api", map[string]string{"k": "v"})
utils.Http.BuildBasicAuth("user", "pass") // "Basic dXNlcjpwYXNz"
utils.Http.SniffContentType(`{"a":1}`)    // "application/json"

// ── UA 解析（37 项浏览器识别表，对照 hutool UserAgentUtil）──
ua := utils.UserAgent.Parse(r.Header.Get("User-Agent"))
if ua != nil {
	ua.Browser        // "Chrome"
	ua.BrowserVersion // "120.0.0.0"
	ua.Os             // "Windows 10 or Windows Server 2016"
	ua.Mobile         // false
}
```

## 🆔 ID 生成 Id / Random

```go
i := utils.Id
i.SimpleUUID()      // 32 位无连字符 UUID
i.ObjectId()        // 24 位 hex（MongoDB 风格，可反推时间）
i.NanoId(21)        // 短 ID，URL 安全
i.SnowflakeNextId() // 包级单例雪花 ID

// 显式分配节点号（多实例部署推荐）
sf := utils.NewSnowflake(workerId, datacenterId)
id := sf.NextId() // 41 位时间戳 + 5 位机房 + 5 位机器 + 12 位序列

r := utils.Random
r.RandomInt(1, 7)   // 模拟骰子 [1,7)
r.RandomNumbers(6)  // "483920"（纯数字验证码）
r.RandomChinese()   // 随机汉字

// 加权随机
wr := utils.NewWeightRandom([]float64{7, 2, 1}, []string{"A", "B", "C"})
prize := wr.Next()
```

## ✅ 校验 Valid / Idcard / Phone / Desensitized

```go
v := utils.Valid
v.IsEmail("user@example.com")        // true
v.IsMobile("+8613812345678")         // true（支持 86/+86 前缀）
v.IsCitizenId("321083197812162119")  // true（省码+生日+ISO 7064 校验位全验证）
v.IsCreditCode("91310110666007217T") // 统一社会信用代码 MOD 31-2 校验
v.RandomCreditCode()                 // 随机生成合法信用代码

ic := utils.Idcard
ic.IsValidCard("321083197812162119")    // 自动分派 15/18 位与港澳台
ic.GetProvinceByIdCard("321083...")     // "江苏"
ic.GetGenderByIdCard(idcard)            // 1 男 / 2 女
ic.GetAgeByIdCard(idcard, time.Time{})  // 周岁
info := ic.GetIdcardInfo(idcard)        // {Province, City, District, Gender, Age, Birthday}

de := utils.Desensitized
de.MobilePhone("13812345678")         // "138****5678"
de.ChineseName("张三丰")                // "张**"
de.Email("zhangsan@example.com")      // "z*******@example.com"
de.BankCard("6222021234567891234")    // 前 4 后 4 保留

p := utils.Phone
p.IsMobile("13812345678")   // true
p.IsTel("010-88888888")     // 座机
```

## 🔤 编解码 Hex / Hash / Codec

```go
// 十六进制
utils.Hex.EncodeStr("你好")      // "e4bda0e5a5bd"
utils.Hex.EncodeColor(255, 0, 0) // "#FF0000"

// 经典字符串哈希（含 Murmur3、Java String.hashCode 对齐版）
utils.Hash.JavaDefaultHash("hello") // 99162322（与 JVM 逐位一致）
utils.Hash.Murmur32("hello")        // 0x248bfa47

// 编码家族
utils.Base58.Encode([]byte("hello world")) // "StV1DL6CwTryKyV"（Bitcoin 字母表）
utils.Base62.Encode(data)                  // 短链接场景
utils.Base32.Encode(data)                  // TOTP 密钥场景
utils.BCD.StrToBcd("12345678")             // 报文 BCD 码
utils.Caesar.Encode("abc", 3)              // 凯撒密码
utils.Morse.Encode("SOS")                  // 摩斯码
utils.Punycode.EncodeDomain("你好.com")     // "xn--6qq79v.com"（中文域名）

// Hashids：把一组数字编成短 ID
h := utils.NewHashids("salt", 10)
enc, _ := h.Encode(1, 2, 3)  // 最短 10 位的混淆 ID
nums, _ := h.Decode(enc)     // [1 2 3]

// 自定义字母表进制
enc, _ = utils.Radix.Encode(utils.Radix34, 123456789)
```

## 🕐 定时任务 Cron

```go
// ── 表达式解析与匹配（5/6/7 字段，支持 * ? , - / L | 与英文别名）──
p, err := utils.NewCronPattern("0 0 12 * * ?") // 每天 12 点
p.Match(time.Date(2024, 6, 5, 12, 0, 0, 0, time.Local), true) // true
next := p.NextAfter(time.Now(), true)          // 下一次触发时间
p2, _ := utils.NewCronPattern("0 0 2 L * ?")   // 每月最后一天 2 点
p3, _ := utils.NewCronPattern("0 0 8 * * ?|0 0 20 * * ?") // 多表达式任一命中

// ── 全局调度器（hutool CronUtil 风格）──
id, _ := utils.Cron.Schedule("0/5 * * * * ?", func() {
	// 每 5 秒执行；任务 panic 会被 recover 并回调监听器
})
utils.Cron.ScheduleWithId("cleanup", "0 0 3 * * ?", dailyCleanup)
utils.Cron.Start() // 启动（幂等）
utils.Cron.Stop()  // 停止

// ── 独立实例（多调度器隔离）──
s := utils.NewScheduler(true) // true=秒级
s.AddListener(utils.TaskListener{
	OnStart:     func(id string) {},
	OnSucceeded: func(id string) {},
	OnFailed:    func(id string, panicked interface{}) {},
})
s.ScheduleWithId("job1", "0/1 * * * * ?", job)
s.Start()
s.Deschedule("job1")
s.Stop(true) // true=同时清空任务表
```

## 📡 网络 Net / Url

```go
n := utils.Net
n.Ipv4ToLong("192.168.1.1")                    // 3232235777
n.LongToIpv4(3232235777)                       // "192.168.1.1"
n.IsInnerIP("10.0.0.5")                        // true（内网判断）
n.IsInRange("192.168.1.10", "192.168.1.0/24")  // true
n.GetLocalIP()                                 // 本机 IPv4
n.IsOpen("127.0.0.1:8080", time.Second)        // TCP 端口探活
n.UsableLocalPort(9000)                        // 找一个空闲端口
n.GetMultistageReverseProxyIp(xffHeader)       // 解多层代理取真实 IP

u := utils.Url
u.CompleteUrl("https://a.com/path/", "img/x.png") // 补全相对地址
u.Normalize("https://a.com/a/b/../c//d")          // "https://a.com/a/c/d"
u.BuildQuery(map[string]string{"b": "1", "a": "2"}) // "a=2&b=1"（键序稳定）
```

## 🗜️ 压缩 Zip

```go
z := utils.Zip
z.ZipDir("./dist", "dist.zip", true)          // 目录打包（true=含根目录层）
z.ZipFiles("out.zip", "a.log", "b.csv")       // 多文件打包
z.ZipEntryBytes("out.zip", "meta.json", data) // 追加内存条目
z.Unzip("dist.zip", "./output")               // 解压（内置 zip-slip 与炸弹防护）
z.ReadEntryBytes("out.zip", "meta.json")      // 读取单个条目
z.Gzip(data)                                  // gzip 压缩/解压（UnGzip）
z.Zlib(data)                                  // zlib 压缩/解压（UnZlib）
```

## 📂 文件 File / 系统 Sys / JSON / 字节流 Bytes

```go
f := utils.File
f.WriteText("conf/app.yaml", content)
f.ReadText("conf/app.yaml")
f.ListFiles("./logs", ".log")        // 递归按扩展名过滤
f.Copy("a.txt", "backup/a.txt")      // 复制（保留权限）
f.NameWithoutExt("/app/main.go")     // "main"
f.Ext("/app/main.go")                // ".go"

s := utils.Sys
s.GetEnv("APP_ENV")                    // 环境变量
s.IsWindows()                          // 平台判断
out, _ := s.RunCmd("git", "log", "-1") // 执行外部命令
s.RunCmdTimeout(5*time.Second, "ffmpeg", "-i", "a.mp4")

j := utils.Json
j.ToJsonString(obj)        // 紧凑 JSON
j.ToPrettyJsonString(obj)  // 带缩进
j.FromJson(data, &target)  // 反序列化
j.Valid([]byte(`{"a":1}`)) // 合法性预检

// 字节流（大小端定长编解码，硬件/协议报文场景）
utils.Bytes.WriteUint16(buf, 0x1234)        // 小端序
utils.Bytes.Int32ToBytes(0x01020304, true)  // 大端序切片
utils.Bytes.WriteFixedString(buf, 8, "ab")  // C 风格定长字符串
utils.Bytes.WriteUtf16String(buf, 4, "你好") // Windows 宽字符
```

## 🧩 集合与泛型辅助

```go
// ── 集合操作：包级泛型函数，类型安全 ──
utils.IndexOf(nums, 3) // 下标，未找到 -1
utils.Contains(tags, "go")
utils.Filter(nums, func(v int) bool { return v%2 == 0 })
utils.Map(ids, func(id int64) string { return strconv.FormatInt(id, 10) })
utils.Unique([]int{3, 1, 3, 2}) // [3 1 2]（保序去重）
utils.Chunk(bigSlice, 500)      // 分批写库
utils.Union(a, b)               // 并集（另有 Intersect / Difference）
utils.Shuffle(list)             // 原地洗牌
utils.Sub(nums, 1, 3)           // 切片截取
utils.ZipToMap(keys, values)    // 两切片 → map
utils.Get(list, 5, "default")   // 越界返回默认值

// ── 泛型辅助 ──
utils.If(x > y, x, y)                  // 三元表达式
utils.Swap(a, b)                       // 交换
utils.DefaultIfNil(cfgPtr, defaultCfg) // 指针判空兜底
utils.Coalesce("", "", "first-valid")  // 取第一个非零值

// ── 对象工具 ──
utils.Obj.IsNil(v)    // 真 nil（含 typed-nil）
utils.Obj.IsEmpty(v)  // 空串/空切片/零值
utils.Obj.Equal(a, b) // 深度相等

// ── 其它便捷 ──
utils.Bool.ToBoolExt("是")         // true（yes/ok/1/对/开/on…）
utils.Char.IsHexChar('a')          // true
utils.Page.TransToStartEnd(2, 10)  // [10, 20) 数据库行区间
utils.Page.Rainbow(5, 20, 7)       // [2 3 4 5 6 7 8] 分页页码
utils.Version.Compare("1.2.0", "1.1.9")           // 1
utils.Version.MatchEl("1.0.2", ">=1.0,<1.1", ",") // true
utils.Coord.Wgs84ToGcj02(116.404, 39.915)         // GPS → 高德/腾讯坐标系
```

## 🌍 更多能力

<details>
<summary><b>坐标系 / 正则 / HTML 清理 / DataURI 等示例（点击展开）</b></summary>

```go
// 坐标系：WGS84 ↔ GCJ-02(火星) ↔ BD-09(百度) ↔ 墨卡托
gcj := utils.Coord.Wgs84ToGcj02(116.404, 39.915)  // GPS → 高德/腾讯
bd := utils.Coord.Gcj02ToBd09(gcj.Lng, gcj.Lat)   // → 百度
wgs := utils.Coord.Gcj02ToWgs84(gcj.Lng, gcj.Lat) // 反纠偏

// 正则工具（hutool ReUtil 风格）
utils.Regex.FindGroup(`(\d+)-(\d+)`, "2024-06", 1)               // "2024"
utils.Regex.ExtractMulti(`(\d+)年(\d+)月`, "2024年6月", "$1/$2") // "2024/6"
utils.Regex.IsMatch(`\d+`, "123") // 全串匹配（Java matches 语义）

// HTML 清理
utils.Html.CleanHtmlTag("<p>你好<b>世界</b></p>") // "你好世界"
utils.Html.RemoveHtmlTag(html, true, "script")    // 连内容删除
utils.Escape.UnescapeHtml4("&lt;p&gt;&amp;nbsp;") // HTML4 全量 251 项实体

// 分页
utils.Page.TotalPage(101, 20) // 6
utils.Page.Rainbow(1, 5, 7)   // [1 2 3 4 5]

// 版本比较
utils.Version.GreaterThanOrEqual("2.3.3", "2.3.3") // true
utils.Version.MatchEl("1.0.2", "1.0.0-1.1.1", ";") // 闭区间

// 网络/IP
utils.Net.HideIpPart("192.168.1.100") // "192.168.1.*"

// DataURI
utils.Url.DataUri("image/png", imgBytes) // "data:image/png;base64,..."
```

</details>

## 🧪 测试

```bash
go test ./...        # 全量测试
go test -race ./...  # 并发竞态检测
go test -cover ./... # 覆盖率
```

| 指标 | 数值 |
|---|---|
| 测试函数 | **448 个**（38 个测试文件） |
| 语句覆盖率 | **94.9%**（无 0% 函数，每个导出方法均被测试触达） |
| 黄金值来源 | hutool 官方测试用例（身份证/中文数字/Base58/Hashids/UA 等） |
| RFC 向量 | RFC 4226/6238（OTP）、RFC 8439（ChaCha20）、RFC 6070（PBKDF2）、RFC 4231（HMAC） |
| 互通验证 | AES/DES/RSA/ECDSA/Ed25519 与 JDK 实测双向互通；`JavaDefaultHash` 与 JVM 逐位一致 |

## ⚖️ 与 hutool 的对照说明

- **命名**：方法语义与命名大体对照 hutool 5.8.x（camelCase → PascalCase），如 `StrUtil.sub` → `String.Sub`、`NumberUtil.decimalFormatMoney` → `Number.DecimalFormatMoney`。
- **语义忠实**：关键算法逐行对照源码移植——cron 表达式（含 `L`、`|`、反向环绕范围、5 字段隐含秒=0）、身份证 15/18 位与港澳台校验、ISO 7064 信用代码校验位、HTML4 全量 251 项实体表等。
- **按需裁剪**：以下依赖 Java 生态或非标准库的能力未移植——SM2/SM3/SM4 国密、BCrypt/Argon2、反射系工具（ReflectUtil/ClassUtil 等）、XML DOM、GBK 编码、图片处理。
- **Go 化改造**：Java 反射类转换（`Convert`/`BeanUtil`）以泛型函数替代；`Optional` 语义由指针 + `DefaultIfNil` 承担。

## ⚠️ 注意事项

1. `utils.Random` 与 `String.RandomString` 使用不同强度的随机源：前者为常规随机（模拟数据），后者为密码学安全源（验证码/令牌）。
2. `Http` 各方法默认 **30 秒超时**（`timeout<=0` 时生效），底层复用同一连接池。
3. AES 的 ECB 模式仅供对接旧系统，新系统请使用 CBC/GCM。
4. `Snowflake` 默认单例的节点号随机生成，多实例部署请用 `NewSnowflake` 显式分配。
5. `Zip.Unzip` 内置 zip-slip 与 2GB 解压上限防护，解压不可信来源的压缩包不会穿越目录。

---

<div align="center">

**如果对你有帮助，欢迎 Star ⭐**

*参考实现：[Hutool](https://hutool.cn) —— 事宜无巨细，尽在 hutool*

</div>
