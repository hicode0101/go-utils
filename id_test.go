package utils

import (
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

/* ---------------- 迁移自 utils_test.go 的 TestIdAndRandom（Id 部分，原样保留） ---------------- */

func TestIdAndRandomLegacy(t *testing.T) {
	su := Id.SimpleUUID()
	if len(su) != 32 || strings.Contains(su, "-") {
		t.Errorf("SimpleUUID 格式错误: %q", su)
	}
	oid := Id.ObjectId()
	if len(oid) != 24 {
		t.Errorf("ObjectId 长度 = %d, want 24", len(oid))
	}
	nid := Id.NanoId(21)
	if len(nid) != 21 {
		t.Errorf("NanoId 长度错误: %q", nid)
	}
	a, b := Id.SnowflakeNextId(), Id.SnowflakeNextId()
	if a >= b {
		t.Errorf("雪花 ID 应单调递增: %d >= %d", a, b)
	}
}

/* ---------------- SimpleUUID ---------------- */

func TestIdSimpleUUID(t *testing.T) {
	su := Id.SimpleUUID()
	// 32 位十六进制、无连字符
	if len(su) != 32 {
		t.Fatalf("SimpleUUID 长度 = %d, want 32", len(su))
	}
	if strings.Contains(su, "-") {
		t.Errorf("SimpleUUID 不应含连字符: %q", su)
	}
	for _, c := range su {
		if !Char.IsHexChar(c) {
			t.Errorf("SimpleUUID 含非 hex 字符 %q: %q", c, su)
		}
		if c >= 'A' && c <= 'F' {
			t.Errorf("SimpleUUID 应为小写: %q", su)
		}
	}
	// UUID v4：版本号半字节为 4，变体半字节为 8/9/a/b
	if su[12] != '4' {
		t.Errorf("SimpleUUID 版本号应为 4: %q", su)
	}
	if !strings.ContainsRune("89ab", rune(su[16])) {
		t.Errorf("SimpleUUID 变体位应为 8/9/a/b: %q", su)
	}
	// 两次调用结果不同
	if su == Id.SimpleUUID() {
		t.Error("两次 SimpleUUID 不应相同")
	}
}

/* ---------------- ObjectId ---------------- */

func TestIdObjectId(t *testing.T) {
	oid := Id.ObjectId()
	// 24 位十六进制
	if len(oid) != 24 {
		t.Fatalf("ObjectId 长度 = %d, want 24", len(oid))
	}
	for _, c := range oid {
		if !Char.IsHexChar(c) {
			t.Fatalf("ObjectId 含非 hex 字符 %q: %q", c, oid)
		}
	}
	// 同进程连续 100 次单调不减（同长 hex 字典序即数值序）
	prev := Id.ObjectId()
	for i := 0; i < 100; i++ {
		cur := Id.ObjectId()
		if cur < prev {
			t.Fatalf("ObjectId 应单调不减: %q < %q", cur, prev)
		}
		prev = cur
	}
	// 可从 ID 反推时间：前 8 位 hex = 秒级时间戳，应在当前时间 ±5s 内
	ts, err := strconv.ParseInt(oid[:8], 16, 64)
	if err != nil {
		t.Fatalf("ObjectId 时间戳解析失败: %v", err)
	}
	if diff := time.Now().Unix() - ts; diff < -5 || diff > 5 {
		t.Errorf("ObjectId 时间戳与当前时间相差 %d 秒, 应在 ±5s 内", diff)
	}
}

/* ---------------- NanoId ---------------- */

func TestIdNanoId(t *testing.T) {
	// size<=0 使用默认长度 21
	for _, size := range []int{0, -1, -100} {
		if got := Id.NanoId(size); len(got) != 21 {
			t.Errorf("NanoId(%d) 长度 = %d, want 默认 21", size, len(got))
		}
	}
	// 自定义长度
	if got := Id.NanoId(10); len(got) != 10 {
		t.Errorf("NanoId(10) 长度 = %d, want 10", len(got))
	}
	// 字符都在 URL 安全字母表内
	got := Id.NanoId(64)
	if len(got) != 64 {
		t.Fatalf("NanoId(64) 长度 = %d, want 64", len(got))
	}
	for _, c := range got {
		if !strings.ContainsRune(idNanoAlphabet, c) {
			t.Fatalf("NanoId 字符 %q 不在字母表内: %q", c, got)
		}
	}
}

/* ---------------- Snowflake ---------------- */

func TestIdSnowflake(t *testing.T) {
	sf := NewSnowflake(1, 1)
	id := sf.NextId()
	if id <= 0 {
		t.Fatalf("雪花 ID 应为正数, got %d", id)
	}
	// 位结构：机器/数据中心位与构造参数一致
	if got := (id >> sfWorkerIdShift) & sfMaxWorkerId; got != 1 {
		t.Errorf("ID 中 workerId 位 = %d, want 1", got)
	}
	if got := (id >> sfDataCenterIdShift) & sfMaxDataCenterId; got != 1 {
		t.Errorf("ID 中 datacenterId 位 = %d, want 1", got)
	}
	// 连续调用严格递增
	prev := id
	for i := 0; i < 1000; i++ {
		cur := sf.NextId()
		if cur <= prev {
			t.Fatalf("雪花 ID 应严格递增: %d <= %d", cur, prev)
		}
		prev = cur
	}
	// NextIdStr 等于十进制串
	idStr := sf.NextIdStr()
	parsed, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		t.Fatalf("NextIdStr 应为合法十进制串: %q, %v", idStr, err)
	}
	if parsed <= prev {
		t.Errorf("NextIdStr 生成的 ID = %d 应大于前一个 %d", parsed, prev)
	}

	// 并发安全：10 个 goroutine 各取 100 个 ID，收集后无重复
	const goroutines, perGoroutine = 10, 100
	ids := make(chan int64, goroutines*perGoroutine)
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perGoroutine; i++ {
				ids <- sf.NextId()
			}
		}()
	}
	wg.Wait()
	close(ids)
	seen := make(map[int64]struct{}, goroutines*perGoroutine)
	for gid := range ids {
		if _, dup := seen[gid]; dup {
			t.Fatalf("并发雪花 ID 重复: %d", gid)
		}
		seen[gid] = struct{}{}
	}
	if len(seen) != goroutines*perGoroutine {
		t.Errorf("并发共收集 %d 个 ID, want %d", len(seen), goroutines*perGoroutine)
	}

	// 越界参数收敛：workerId 33 -> 1，datacenterId -1 -> 31，不 panic
	if got := sfClampId(33); got != 1 {
		t.Errorf("sfClampId(33) = %d, want 1", got)
	}
	if got := sfClampId(-1); got != 31 {
		t.Errorf("sfClampId(-1) = %d, want 31", got)
	}
	sf2 := NewSnowflake(33, -1) // 不得 panic
	id2 := sf2.NextId()
	if id2 <= 0 {
		t.Fatalf("越界参数创建的雪花 ID 应仍为正数, got %d", id2)
	}
	if got := (id2 >> sfWorkerIdShift) & sfMaxWorkerId; got != 1 {
		t.Errorf("越界参数收敛后 workerId 位 = %d, want 1", got)
	}
	if got := (id2 >> sfDataCenterIdShift) & sfMaxDataCenterId; got != 31 {
		t.Errorf("越界参数收敛后 datacenterId 位 = %d, want 31", got)
	}
}

/* ---------------- 包级雪花单例 ---------------- */

func TestIdSnowflakeNextIdSingleton(t *testing.T) {
	if id := Id.SnowflakeNextId(); id <= 0 {
		t.Errorf("包级 SnowflakeNextId 应为正数, got %d", id)
	}
	s := Id.SnowflakeNextIdStr()
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n <= 0 {
		t.Errorf("SnowflakeNextIdStr 应为正数十进制串: %q, %v", s, err)
	}
}
