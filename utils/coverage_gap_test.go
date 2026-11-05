package utils

import (
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// 本文件补齐覆盖率缺口：针对边界分支与环境无关的错误路径，
// 断言值均先用临时程序实测当前实现后固化（见各用例注释）。

/* ---------------- Snowflake：序列耗尽自旋（tilNextMillis） ---------------- */

// TestGapSnowflakeSeqOverflow 单毫秒内连发超过 4096 个 ID 触发序列耗尽自旋：
// 10000 个 ID 必须严格递增且无重复（覆盖 tilNextMillis 与下一毫秒进位）。
func TestGapSnowflakeSeqOverflow(t *testing.T) {
	sf := NewSnowflake(1, 1)
	prev := sf.NextId()
	seen := make(map[int64]struct{}, 10001)
	seen[prev] = struct{}{}
	for i := 0; i < 10000; i++ {
		id := sf.NextId()
		if id <= prev {
			t.Fatalf("第 %d 个 ID 应严格递增: %d <= %d", i, id, prev)
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("ID 重复: %d", id)
		}
		seen[id] = struct{}{}
		prev = id
	}
}

/* ---------------- Cron：匹配器访问器与区间步进分支 ---------------- */

// TestGapCronMatcherAccessors 白盒覆盖 getMaxValue/getMinValue 访问器
func TestGapCronMatcherAccessors(t *testing.T) {
	m := &cronBoolArrayMatcher{bValues: make([]bool, 60), minValue: 2, maxValue: 7}
	if m.getMinValue() != 2 || m.getMaxValue() != 7 {
		t.Errorf("getMinValue/getMaxValue = %d/%d, want 2/7", m.getMinValue(), m.getMaxValue())
	}
}

// TestGapCronStepRange 覆盖 cronAppendRange 的各分支：start==stop、
// start<stop 正常步进、反向范围（start>stop）按源码语义填充
func TestGapCronStepRange(t *testing.T) {
	// 分钟 5-15/5 → {5,10,15}：start<stop 正向步进
	p1, err := NewCronPattern("0 5-15/5 * * * ?")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if !p1.Match(cronAt(2024, 6, 5, 10, 5, 0), true) ||
		!p1.Match(cronAt(2024, 6, 5, 10, 10, 0), true) ||
		!p1.Match(cronAt(2024, 6, 5, 10, 15, 0), true) {
		t.Error("5-15/5 应匹配 5/10/15 分")
	}
	if p1.Match(cronAt(2024, 6, 5, 10, 7, 0), true) {
		t.Error("7 分不应匹配 5-15/5")
	}
	// 秒 10-10 → 单值（start==stop 分支）
	p2, err := NewCronPattern("10 0 0 * * ?")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if !p2.Match(cronAt(2024, 6, 5, 0, 0, 10), true) {
		t.Error("单值秒 10 应匹配")
	}
	// 反向范围 15-5/2：hutool 环绕语义 = {15,17,...,59} ∪ {0,2,4}
	p3, err := NewCronPattern("0 15-5/2 * * * ?")
	if err != nil {
		t.Fatalf("解析反向范围失败: %v", err)
	}
	if !p3.Match(cronAt(2024, 6, 5, 10, 15, 0), true) || !p3.Match(cronAt(2024, 6, 5, 10, 59, 0), true) {
		t.Error("反向范围前段应含 15 与 59 分")
	}
	if !p3.Match(cronAt(2024, 6, 5, 10, 0, 0), true) || !p3.Match(cronAt(2024, 6, 5, 10, 2, 0), true) {
		t.Error("反向范围环绕段应含 0 与 2 分")
	}
	if p3.Match(cronAt(2024, 6, 5, 10, 5, 0), true) {
		t.Error("5 分步进为奇数不应匹配 15-5/2")
	}
}

// TestGapCronNextAfterDowFallback 覆盖 NextAfter 的周域按天回退：
// 周一 0 点的表达式，从周三出发应推进到下周周一
func TestGapCronNextAfterDowFallback(t *testing.T) {
	p, err := NewCronPattern("0 0 0 * * 1")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	// 2024-06-05 是周三
	next := p.NextAfter(cronAt(2024, 6, 5, 12, 0, 0), true)
	if want := cronAt(2024, 6, 10, 0, 0, 0); !next.Equal(want) {
		t.Errorf("NextAfter = %v, want %v（下周周一）", next.Format(time.DateTime), want.Format(time.DateTime))
	}
}

/* ---------------- Zip：写入失败路径 ---------------- */

// TestGapZipWriteError 目标路径所在目录不存在时写 zip 应返回错误
func TestGapZipWriteError(t *testing.T) {
	dir := t.TempDir()
	badPath := dir + "/no-such-subdir/out.zip"
	if err := Zip.ZipEntryBytes(badPath, "a.txt", []byte("x")); err == nil {
		t.Error("目标目录不存在时应返回错误")
	}
	if err := Zip.ZipFiles(badPath, dir); err == nil {
		t.Error("目标目录不存在时 ZipFiles 应返回错误")
	}
}

/* ---------------- 脱敏：空输入行为（实测固化） ---------------- */

func TestGapDesensitizedEmptyInput(t *testing.T) {
	if got := Desensitized.FixedPhone(""); got != "" {
		t.Errorf("FixedPhone(\"\") = %q", got)
	}
	if got := Desensitized.MobilePhone(""); got != "" {
		t.Errorf("MobilePhone(\"\") = %q", got)
	}
	if got := Desensitized.Address("", 3); got != "" {
		t.Errorf("Address(\"\",3) = %q", got)
	}
	if got := Desensitized.Password(""); got != "" {
		t.Errorf("Password(\"\") = %q", got)
	}
	if got := Desensitized.BankCard(""); got != "" {
		t.Errorf("BankCard(\"\") = %q", got)
	}
	if got := Desensitized.IdCardNum("", 3, 4); got != "" {
		t.Errorf("IdCardNum(\"\") = %q", got)
	}
	if got := Desensitized.CarLicense(""); got != "" {
		t.Errorf("CarLicense(\"\") = %q", got)
	}
	// 空串按段打码的实现语义（实测值）
	if got := Desensitized.Ipv4(""); got != ".*.*.*" {
		t.Errorf("Ipv4(\"\") = %q, want \".*.*.*\"", got)
	}
	if got := Desensitized.Ipv6(""); got != ":*:*:*:*:*:*:*" {
		t.Errorf("Ipv6(\"\") = %q, want \":*:*:*:*:*:*:*\"", got)
	}
	if got := Desensitized.Passport(""); got != "" {
		t.Errorf("Passport(\"\") = %q", got)
	}
	if got := Desensitized.CreditCode(""); got != "" {
		t.Errorf("CreditCode(\"\") = %q", got)
	}
	if got := Desensitized.FirstMask(""); got != "" {
		t.Errorf("FirstMask(\"\") = %q", got)
	}
	// 拒绝采样重试分支：多次生成提高覆盖 validRandInt 的重试路径
	for i := 0; i < 50; i++ {
		if code := Valid.RandomCreditCode(); !Valid.IsCreditCode(code) {
			t.Fatalf("RandomCreditCode 生成的 %q 非法", code)
		}
	}
}

/* ---------------- 正则/字符串/数值：边界分支 ---------------- */

// TestGapRegexInvalidPattern 非法正则返回原文
func TestGapRegexInvalidPattern(t *testing.T) {
	if got := Regex.DelAll("([", "abc"); got != "abc" {
		t.Errorf("DelAll 非法正则应返回原串, got %q", got)
	}
	if got := Regex.Find("([", "abc"); got != "" {
		t.Errorf("Find 非法正则应返回空串, got %q", got)
	}
}

// TestGapStrTruncateNegative 负长度原样返回
func TestGapStrTruncateNegative(t *testing.T) {
	if got := String.Truncate("abc", -1, "..."); got != "abc" {
		t.Errorf("Truncate(-1) = %q, want \"abc\"", got)
	}
}

// TestGapNumberSpecialFloat NaN/Inf 舍入返回 0（big.Rat 无法表示）
func TestGapNumberSpecialFloat(t *testing.T) {
	if got := Number.Round(math.NaN(), 2); got != 0 {
		t.Errorf("Round(NaN) = %v, want 0", got)
	}
	if got := Number.Round(math.Inf(1), 2); got != 0 {
		t.Errorf("Round(+Inf) = %v, want 0", got)
	}
	if got := Number.ToStr(math.NaN()); got != "NaN" {
		t.Errorf("ToStr(NaN) = %q, want \"NaN\"", got)
	}
}

/* ---------------- 版本：字符串 token 比较分支 ---------------- */

func TestGapVersionStringToken(t *testing.T) {
	if got := Version.Compare("1.0.0a", "1.0.0b"); got != -1 {
		t.Errorf("Compare(1.0.0a,1.0.0b) = %d, want -1", got)
	}
	if got := Version.Compare("1a", "1b"); got != -1 {
		t.Errorf("Compare(1a,1b) = %d, want -1", got)
	}
	if got := Version.Compare("2.0-beta", "2.0-alpha"); got != 1 {
		t.Errorf("Compare(2.0-beta,2.0-alpha) = %d, want 1", got)
	}
}

/* ---------------- 身份证：港澳台与未知省份分支 ---------------- */

func TestGapIdcardHkTwBranches(t *testing.T) {
	// hutool 官方测试样本（实测当前实现）
	if !Idcard.IsValidCard("P174468(6)") {
		t.Error("香港证件 P174468(6) 应有效")
	}
	if !Idcard.IsValidCard("B221690311") {
		t.Error("台湾证件 B221690311 应有效")
	}
	if !Valid.IsCitizenId("U193683453") {
		t.Error("10 位证件 U193683453 应有效")
	}
	// 未知省份码 99
	if got := Idcard.GetProvinceByIdCard("990000202001010019"); got != "" {
		t.Errorf("未知省份码应返回空串, got %q", got)
	}
	if got := Idcard.GetIdcardInfo("abc"); got != nil {
		t.Errorf("非法证件 GetIdcardInfo 应返回 nil, got %+v", got)
	}
}

/* ---------------- 哈希：Murmur128 尾部分支 ---------------- */

// TestGapMurmur128Tail 各长度输入覆盖 x64 128 位的尾部处理分支
func TestGapMurmur128Tail(t *testing.T) {
	prev := [2]uint64{}
	for n := 0; n <= 24; n++ {
		in := strings.Repeat("x", n)
		h1 := Hash.Murmur128(in)
		h2 := Hash.Murmur128(in)
		if h1 != h2 {
			t.Fatalf("Murmur128 长度 %d 不确定: %v vs %v", n, h1, h2)
		}
		if n > 0 && h1 == prev {
			// 不同长度允许极小概率相同，仅记录不判失败
			t.Logf("长度 %d 与前一长度哈希相同: %v", n, h1)
		}
		prev = h1
	}
	// 空串哈希为 0（MurmurHash3 规范）
	if h := Hash.Murmur128(""); h != [2]uint64{} {
		t.Errorf("Murmur128(\"\") = %v, want 零值", h)
	}
}

/* ---------------- HTTP：HEAD 空响应体的 BodyString 分支 ---------------- */

// TestGapHttpBodyStringEmpty HEAD 响应体为空时 BodyString 返回空串
func TestGapHttpBodyStringEmpty(t *testing.T) {
	resp := Http.Head(gapTestServerURL(t), nil, 0)
	if got := resp.BodyString(); got != "" {
		t.Errorf("HEAD 响应 BodyString = %q, want 空串", got)
	}
	if !resp.IsOK() {
		t.Errorf("HEAD 应返回 2xx, got %d", resp.Status)
	}
}

// gapTestServerURL 启动一个返回固定内容的本地服务器（本文件专用，避免与其它测试文件的辅助函数重名）
func gapTestServerURL(t *testing.T) string {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv.URL
}

/* ---------------- Hashids：非法字母表（白盒 panic 路径） ---------------- */

func TestGapHashidsInvalidAlphabet(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("含重复字符的字母表应 panic（对照 hutool 构造抛异常）")
		}
	}()
	_ = newHashids("salt", 0, "aab")
}

/* ---------------- 第二批补齐：中文数字/版本混合 token/Hashids 参数/证件分支 ---------------- */

// TestGapNumberChineseFormats 覆盖 numberLongToChinese 的"拾"位、亿分节补零等分支（实测值）
func TestGapNumberChineseFormats(t *testing.T) {
	cases := []struct {
		num  float64
		want string
	}{
		{110.5, "壹佰壹拾元伍角"},
		{20, "贰拾元整"},
		{0.05, "伍分"},
		{100000001, "壹亿零壹元整"},
	}
	for _, c := range cases {
		if got := Number.DigitToChinese(c.num); got != c.want {
			t.Errorf("DigitToChinese(%v) = %q, want %q", c.num, got, c.want)
		}
	}
}

// TestGapVersionMixedToken 覆盖 versionTokenString 的数字分支（混合类型按字符串比较）
func TestGapVersionMixedToken(t *testing.T) {
	if got := Version.Compare("1a", "1"); got != 1 {
		t.Errorf("Compare(1a,1) = %d, want 1", got)
	}
	if got := Version.Compare("1", "1a"); got != -1 {
		t.Errorf("Compare(1,1a) = %d, want -1", got)
	}
}

// TestGapHashidsParams minLength 负值归一为 0 可用；字母表过短（<3）panic
func TestGapHashidsParams(t *testing.T) {
	h := newHashids("salt", -5, "")
	if enc, err := h.Encode(1, 2, 3); err != nil || enc == "" {
		t.Errorf("minLength 为负应按 0 处理, %q, %v", enc, err)
	}
	defer func() {
		if recover() == nil {
			t.Error("字母表长度不足 3 应 panic")
		}
	}()
	_ = newHashids("salt", 0, "ab")
}

// TestGapIdcardMoreBranches 小写 x 校验位、TW 非法样本（hutool errTwCard1）、空省份码
func TestGapIdcardMoreBranches(t *testing.T) {
	if !Idcard.IsValidCard18("33010219200403064x") {
		t.Error("小写 x 校验位应被接受")
	}
	if Idcard.IsValidCard("M517086311") {
		t.Error("台湾证件 M517086311 为 hutool 官方非法样本，应返回 false")
	}
	if got := Idcard.GetProvinceByIdCard(""); got != "" {
		t.Errorf("空串省份应返回空串, got %q", got)
	}
}

// TestGapCronNextAfterDayOfMonth 覆盖 NextAfter 的日域推进（nextAfterDay/getMinDay）
func TestGapCronNextAfterDayOfMonth(t *testing.T) {
	// 日域 20 号：6 月 5 日出发 → 当月 20 日
	p1, err := NewCronPattern("0 0 0 20 * ?")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if next := p1.NextAfter(cronAt(2024, 6, 5, 12, 0, 0), true); !next.Equal(cronAt(2024, 6, 20, 0, 0, 0)) {
		t.Errorf("NextAfter = %v, want 2024-06-20 00:00:00", next.Format(time.DateTime))
	}
	// 已过 20 号 → 推进到下月 20 日
	if next := p1.NextAfter(cronAt(2024, 6, 25, 12, 0, 0), true); !next.Equal(cronAt(2024, 7, 20, 0, 0, 0)) {
		t.Errorf("NextAfter 跨月 = %v, want 2024-07-20 00:00:00", next.Format(time.DateTime))
	}
	// L：6 月 25 日出发 → 6 月 30 日
	p2, err := NewCronPattern("0 0 0 L * ?")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if next := p2.NextAfter(cronAt(2024, 6, 25, 12, 0, 0), true); !next.Equal(cronAt(2024, 6, 30, 0, 0, 0)) {
		t.Errorf("NextAfter(L) = %v, want 2024-06-30 00:00:00", next.Format(time.DateTime))
	}
	// 多表达式 NextAfter 取最早：从 11 点出发，第二个表达式（8 点）胜出（覆盖 minOf 分支）
	p3, err := NewCronPattern("0 0 10 * * ?|0 0 8 * * ?")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if next := p3.NextAfter(cronAt(2024, 6, 5, 11, 0, 0), true); !next.Equal(cronAt(2024, 6, 6, 8, 0, 0)) {
		t.Errorf("多表达式 NextAfter = %v, want 2024-06-06 08:00:00（8 点分支胜出）", next.Format(time.DateTime))
	}
	// 月份别名非法 → 解析错误（覆盖 cronParseMonthAlias 失败分支）
	if _, err := NewCronPattern("0 0 0 * foo ?"); err == nil {
		t.Error("非法月份别名应返回错误")
	}
}
