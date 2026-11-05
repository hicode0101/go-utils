package utils

import (
	"strings"
	"testing"
)

/* ---------------- 经典字符串哈希（hutool HashUtil 移植） ---------------- */

// TestHashClassicGolden 覆盖全部 uint32 返回的经典字符串哈希：
// 以 "hello" 为输入对每个算法断言黄金值（值由实现实际运行后固化，
// 其中 JavaDefaultHash/Murmur32 两个锚点已与 Java/hutool 侧对齐验证），
// 并覆盖空串下"初值直接可见"的确定性语义。
func TestHashClassicGolden(t *testing.T) {
	const key = "hello"
	golden := []struct {
		name string
		got  uint32
		want uint32
	}{
		{"FnvHash", Hash.FnvHash(key), 350638217},
		{"ElfHash", Hash.ElfHash(key), 7258927},
		{"BkdrHash", Hash.BkdrHash(key), 792145550},
		{"SdbmHash", Hash.SdbmHash(key), 684824882},
		{"DjbHash", Hash.DjbHash(key), 261238937},
		{"DekHash", Hash.DekHash(key), 213073391},
		{"ApHash", Hash.ApHash(key), 1889857574},
		{"JsHash", Hash.JsHash(key), 166137909},
		{"RsHash", Hash.RsHash(key), 987012754},
		{"PjwHash", Hash.PjwHash(key), 7258927},
		{"RotatingHash", Hash.RotatingHash(key), 4078255},
		{"OneByOneHash", Hash.OneByOneHash(key), 3035920923},
		{"Bernstein", Hash.Bernstein(key), 261238937},
	}
	for _, c := range golden {
		if c.got != c.want {
			t.Errorf("%s(%q) = %d, want %d（黄金值）", c.name, key, c.got, c.want)
		}
	}
	// 原 utils_test.go TestHashHex 断言迁入
	if Hash.FnvHash("hello") == 0 {
		t.Error("FnvHash 不应为 0")
	}

	// 空串语义：初值直接可见（Bernstein/Djb 初值 5381，Js 初值 1315423911，其余为 0）
	empty := []struct {
		name string
		got  uint32
		want uint32
	}{
		{"RotatingHash", Hash.RotatingHash(""), 0},
		{"OneByOneHash", Hash.OneByOneHash(""), 0},
		{"Bernstein", Hash.Bernstein(""), 5381},
		{"FnvHash", Hash.FnvHash(""), 1494218850},
		{"RsHash", Hash.RsHash(""), 0},
		{"JsHash", Hash.JsHash(""), 1315423911},
		{"PjwHash", Hash.PjwHash(""), 0},
		{"ElfHash", Hash.ElfHash(""), 0},
		{"BkdrHash", Hash.BkdrHash(""), 0},
		{"SdbmHash", Hash.SdbmHash(""), 0},
		{"DjbHash", Hash.DjbHash(""), 5381},
		{"DekHash", Hash.DekHash(""), 0},
		{"ApHash", Hash.ApHash(""), 0},
	}
	for _, c := range empty {
		if c.got != c.want {
			t.Errorf("%s(\"\") = %d, want %d", c.name, c.got, c.want)
		}
	}
}

// TestHashDeterministic 所有简单哈希对同一输入两次调用结果一致。
func TestHashDeterministic(t *testing.T) {
	u32Funcs := []struct {
		name string
		fn   func(string) uint32
	}{
		{"FnvHash", Hash.FnvHash},
		{"RotatingHash", Hash.RotatingHash},
		{"OneByOneHash", Hash.OneByOneHash},
		{"Bernstein", Hash.Bernstein},
		{"RsHash", Hash.RsHash},
		{"JsHash", Hash.JsHash},
		{"PjwHash", Hash.PjwHash},
		{"ElfHash", Hash.ElfHash},
		{"BkdrHash", Hash.BkdrHash},
		{"SdbmHash", Hash.SdbmHash},
		{"DjbHash", Hash.DjbHash},
		{"DekHash", Hash.DekHash},
		{"ApHash", Hash.ApHash},
		{"Murmur32", Hash.Murmur32},
	}
	inputs := []string{"hello", "", "a", "中文", "你好，世界"}
	for _, in := range inputs {
		for _, c := range u32Funcs {
			if a, b := c.fn(in), c.fn(in); a != b {
				t.Errorf("%s(%q) 两次调用结果不一致: %d != %d", c.name, in, a, b)
			}
		}
		if a, b := Hash.JavaDefaultHash(in), Hash.JavaDefaultHash(in); a != b {
			t.Errorf("JavaDefaultHash(%q) 两次调用结果不一致", in)
		}
		if a, b := Hash.TianlHash(in), Hash.TianlHash(in); a != b {
			t.Errorf("TianlHash(%q) 两次调用结果不一致", in)
		}
		if a, b := Hash.MixHash(in), Hash.MixHash(in); a != b {
			t.Errorf("MixHash(%q) 两次调用结果不一致", in)
		}
		if a, b := Hash.Murmur64(in), Hash.Murmur64(in); a != b {
			t.Errorf("Murmur64(%q) 两次调用结果不一致", in)
		}
		if a, b := Hash.Murmur128(in), Hash.Murmur128(in); a != b {
			t.Errorf("Murmur128(%q) 两次调用结果不一致", in)
		}
	}
}

// TestHashChineseDeterministic 中文（多字节 UTF-8）与增补平面字符输入下的确定性。
func TestHashChineseDeterministic(t *testing.T) {
	// "𠀀" 为增补平面字符，JavaDefaultHash 按代理对拆成两个 code unit 参与运算
	inputs := []string{"中文", "你好，世界！", "日本語テスト", "𠀀"}
	for _, in := range inputs {
		if a, b := Hash.JavaDefaultHash(in), Hash.JavaDefaultHash(in); a != b {
			t.Errorf("JavaDefaultHash(%q) 中文输入不确定", in)
		}
		if a, b := Hash.Murmur32(in), Hash.Murmur32(in); a != b {
			t.Errorf("Murmur32(%q) 中文输入不确定", in)
		}
		if a, b := Hash.FnvHash(in), Hash.FnvHash(in); a != b {
			t.Errorf("FnvHash(%q) 中文输入不确定", in)
		}
		if a, b := Hash.MixHash(in), Hash.MixHash(in); a != b {
			t.Errorf("MixHash(%q) 中文输入不确定", in)
		}
	}
	// 中文黄金值（固化）
	if got := Hash.JavaDefaultHash("中文"); got != 646394 {
		t.Errorf("JavaDefaultHash(中文) = %d, want 646394", got)
	}
	if got := Hash.Murmur32("中文"); got != 0xdb267197 {
		t.Errorf("Murmur32(中文) = %#x, want 0xdb267197", got)
	}
	if got := Hash.FnvHash("中文"); got != 121895792 {
		t.Errorf("FnvHash(中文) = %d, want 121895792", got)
	}
}

// TestHashAdditive 加法哈希：hash = (串长 + Σ字节) % prime，可手算验证；
// 不同 prime 结果不同。
func TestHashAdditive(t *testing.T) {
	// "hello" 累加和 = 5 + 104+101+108+108+111 = 537
	if got := Hash.AdditiveHash("hello", 31); got != 10 {
		t.Errorf("AdditiveHash(hello,31) = %d, want 10", got)
	}
	if got := Hash.AdditiveHash("hello", 131); got != 13 {
		t.Errorf("AdditiveHash(hello,131) = %d, want 13", got)
	}
	if got := Hash.AdditiveHash("hello", 1000003); got != 537 { // prime 大于累加和时即原和
		t.Errorf("AdditiveHash(hello,1000003) = %d, want 537", got)
	}
	if got := Hash.AdditiveHash("", 31); got != 0 {
		t.Errorf("AdditiveHash(\"\",31) = %d, want 0", got)
	}
	// 不同 prime 结果不同
	if Hash.AdditiveHash("hello", 31) == Hash.AdditiveHash("hello", 131) {
		t.Error("不同 prime 的 AdditiveHash 结果应不同")
	}
}

// assertPanics 断言 fn 执行时发生 panic（用于表过短的数组越界语义）
func assertPanics(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Errorf("%s: 应 panic", name)
		}
	}()
	fn()
}

// TestHashUniversal 通用哈希：固定 tab 断言确定性 + 逐位取 tab[i] 索引语义 + 表过短 panic。
func TestHashUniversal(t *testing.T) {
	key := "ab"
	// 固定随机表：长度 = len(key)*8（逐位取用）
	tab := make([]uint32, len(key)*8)
	for i := range tab {
		tab[i] = uint32(i*i + 1)
	}
	const mask = 0xFFFF
	got := Hash.Universal(key, mask, tab)

	// 期望值按语义独立推算：初始为串长，字节第 j 位为 0 时异或 tab[全局位下标]
	want := uint32(len(key))
	for i := 0; i < len(key); i++ {
		for j := 0; j < 8; j++ {
			if key[i]&(1<<uint(j)) == 0 {
				want ^= tab[i*8+j]
			}
		}
	}
	want &= mask
	if got != want {
		t.Errorf("Universal(%q) = %d, want %d", key, got, want)
	}
	// 固定 tab 下确定性
	if got != Hash.Universal(key, mask, tab) {
		t.Error("Universal 固定 tab 应确定性")
	}
	// tab 过短应 panic（对应 Java 数组越界）
	assertPanics(t, "Universal tab 过短", func() {
		Hash.Universal(key, mask, tab[:3])
	})
}

// TestHashZobrist Zobrist 哈希：固定 tab 断言确定性 + tab[i*256+b] 索引语义
// （与 hutool int[][] tab 按 [字节位置][字节值] 取用对齐）+ 表过短 panic。
func TestHashZobrist(t *testing.T) {
	key := "ab"
	// 扁平化二维随机表：长度 = len(key)*256
	tab := make([]uint32, len(key)*256)
	for i := range tab {
		tab[i] = uint32(i*i + 1)
	}
	const mask = 0xFFFFFF
	got := Hash.Zobrist(key, mask, tab)

	// 期望值按 tab[i*256+b] 索引语义独立推算（i 为字节位置、b 为字节值）
	want := uint32(len(key))
	for i := 0; i < len(key); i++ {
		want ^= tab[i*256+int(key[i])]
	}
	want &= mask
	if got != want {
		t.Errorf("Zobrist(%q) = %d, want %d", key, got, want)
	}
	// 固定 tab 下确定性
	if got != Hash.Zobrist(key, mask, tab) {
		t.Error("Zobrist 固定 tab 应确定性")
	}
	// tab 过短应 panic
	assertPanics(t, "Zobrist tab 过短", func() {
		Hash.Zobrist(key, mask, tab[:255])
	})
}

// TestHashTianl TianL 哈希：黄金值 + 空串 0 + 大小写折叠 + 长串仅取末尾 96 字节。
func TestHashTianl(t *testing.T) {
	if got := Hash.TianlHash("hello"); got != 67646319 {
		t.Errorf("TianlHash(hello) = %d, want 67646319", got)
	}
	if got := Hash.TianlHash(""); got != 0 {
		t.Errorf("TianlHash(\"\") = %d, want 0", got)
	}
	// 大写字母折算小写参与计算
	if Hash.TianlHash("HELLO") != Hash.TianlHash("hello") {
		t.Error("TianlHash 大小写折叠后应一致")
	}
	// 超过 96 字节仅取末尾 96 字节：两者尾段相同，差值只来自基数 16777216*(len-1)
	long97 := strings.Repeat("a", 97)
	long100 := strings.Repeat("a", 100)
	if diff := Hash.TianlHash(long100) - Hash.TianlHash(long97); diff != 16777216*3 {
		t.Errorf("TianlHash 长串尾 96 字节语义错误: diff = %d, want %d", diff, 16777216*3)
	}
}

// TestHashJavaDefaultMix Java 默认哈希与 64 位混合哈希。
func TestHashJavaDefaultMix(t *testing.T) {
	// Java String.hashCode("hello") = 99162322（原 utils_test.go TestHashHex 断言迁入）
	if got := Hash.JavaDefaultHash("hello"); got != 99162322 {
		t.Errorf("JavaDefaultHash(hello) = %d, want 99162322", got)
	}
	if got := Hash.JavaDefaultHash("中文"); got != 646394 {
		t.Errorf("JavaDefaultHash(中文) = %d, want 646394", got)
	}
	// MixHash：高 32 位为 JavaDefaultHash，低 32 位为 FnvHash（有符号承载）
	if got := Hash.MixHash("hello"); got != 425898930336059529 {
		t.Errorf("MixHash(hello) = %d, want 425898930336059529", got)
	}
	if got := Hash.MixHash(""); got != 1494218850 {
		t.Errorf("MixHash(\"\") = %d, want 1494218850（高 32 位为 0，低 32 位即 FnvHash 空串值）", got)
	}
	// 合成语义反证：高低 32 位分别等于两个分量
	mix := Hash.MixHash("中文")
	if int32(mix>>32) != Hash.JavaDefaultHash("中文") || int32(mix) != int32(Hash.FnvHash("中文")) {
		t.Errorf("MixHash 高低 32 位合成语义错误: %d", mix)
	}
}

/* ---------------- Murmur3 ---------------- */

// TestHashMurmur Murmur 家族：标准向量 + 空串 + 128 位双分量。
func TestHashMurmur(t *testing.T) {
	// MurmurHash3 x86_32 标准测试向量（原 utils_test.go TestHashHex 断言迁入）
	if got := Hash.Murmur32("hello"); got != 0x248bfa47 {
		t.Errorf("Murmur32(hello) = %#x, want 0x248bfa47", got)
	}
	if got := Hash.Murmur32(""); got != 0 {
		t.Errorf("Murmur32(\"\") = %#x, want 0", got)
	}
	if got := Hash.Murmur32("中文"); got != 0xdb267197 {
		t.Errorf("Murmur32(中文) = %#x, want 0xdb267197", got)
	}
	// hutool 版 64 位
	if got := Hash.Murmur64("hello"); got != 0xd35fde0222ad4f2f {
		t.Errorf("Murmur64(hello) = %#x, want 0xd35fde0222ad4f2f", got)
	}
	if got := Hash.Murmur64(""); got != 0 {
		t.Errorf("Murmur64(\"\") = %#x, want 0", got)
	}
	// 128 位：两个 64 位分量分别断言
	m128 := Hash.Murmur128("hello")
	if m128[0] != 0xcbd8a7b341bd9b02 || m128[1] != 0x5b1e906a48ae1d19 {
		t.Errorf("Murmur128(hello) = [%#x, %#x], want [0xcbd8a7b341bd9b02, 0x5b1e906a48ae1d19]",
			m128[0], m128[1])
	}
	if zero := Hash.Murmur128(""); zero[0] != 0 || zero[1] != 0 {
		t.Errorf("Murmur128(\"\") = [%#x, %#x], want [0, 0]", zero[0], zero[1])
	}
}
