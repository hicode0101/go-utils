package utils

import (
	"encoding/binary"
	"math/bits"
	"unicode/utf16"
)

// utilHash 哈希算法工具。
//
// 注意：本文件全部为非加密散列，仅用于哈希表散列、布隆过滤器、分片路由、简单指纹等场景，
// 不可用于密码存储与安全签名（安全摘要见 Crypto）。
// 各算法逐行对照 hutool HashUtil / MurmurHash 移植，输入统一按 UTF-8 字节处理
// （对 ASCII 串的输出与 Java 版逐位一致）；Java 的 int 为有符号 32 位数且其 >> 为符号扩展移位，
// 这里用 uint32 承载中间值并配合 hashSar32 还原该语义。
type utilHash struct{}

/* ---------------- 经典字符串哈希 ---------------- */

// AdditiveHash 加法哈希：串长与各字节累加后对质数 prime 取模。
// 场景：最简单的字符串散列，对散列质量要求不高时的快速分桶。
func (h *utilHash) AdditiveHash(key string, prime uint32) uint32 {
	hash := uint32(len(key))
	for i := 0; i < len(key); i++ {
		hash += uint32(key[i])
	}
	return hash % prime
}

// RotatingHash 旋转哈希：每步 (hash<<4) ^ (hash>>28) ^ 字节 循环混合。
// 场景：快速字符串散列；hutool 原版另带 prime 参数做取模收拢，
// 这里按约定返回旋转后的原始值，需要时可由调用方再 % prime 或 & mask。
func (h *utilHash) RotatingHash(key string) uint32 {
	hash := uint32(len(key))
	for i := 0; i < len(key); i++ {
		hash = (hash << 4) ^ hashSar32(hash, 28) ^ uint32(key[i])
	}
	return hash
}

// OneByOneHash 一次一个哈希：逐字节累加并做多轮移位混合，最后再整体雪崩一轮。
// 场景：通用字符串散列，混淆性优于加法哈希。
func (h *utilHash) OneByOneHash(key string) uint32 {
	var hash uint32
	for i := 0; i < len(key); i++ {
		hash += uint32(key[i])
		hash += hash << 10
		hash ^= hashSar32(hash, 6)
	}
	hash += hash << 3
	hash ^= hashSar32(hash, 11)
	hash += hash << 15
	return hash
}

// Bernstein DJB2 哈希：hash = hash*33 + 字节，初值 5381（hutool 原版初值为 0，此处按 DJB2 惯例实现）。
// 场景：符号表散列、简单缓存键。
func (h *utilHash) Bernstein(key string) uint32 {
	hash := uint32(5381)
	for i := 0; i < len(key); i++ {
		hash = hash*33 + uint32(key[i])
	}
	return hash
}

// Universal 通用哈希（Universal Hashing）：按每个字节的二进制位从随机表 tab 选取条目异或，最后 & mask 收拢。
// tab 长度需 >= len(key)*8（对应 hutool 的 int[] tab，逐位取用），不足时 panic（同 Java 数组越界）。
// 场景：需要对抗构造性冲突的哈希。
func (h *utilHash) Universal(key string, mask uint32, tab []uint32) uint32 {
	hash := uint32(len(key))
	for i := 0; i < len(key)<<3; i += 8 {
		k := key[i>>3]
		if k&0x01 == 0 {
			hash ^= tab[i]
		}
		if k&0x02 == 0 {
			hash ^= tab[i+1]
		}
		if k&0x04 == 0 {
			hash ^= tab[i+2]
		}
		if k&0x08 == 0 {
			hash ^= tab[i+3]
		}
		if k&0x10 == 0 {
			hash ^= tab[i+4]
		}
		if k&0x20 == 0 {
			hash ^= tab[i+5]
		}
		if k&0x40 == 0 {
			hash ^= tab[i+6]
		}
		if k&0x80 == 0 {
			hash ^= tab[i+7]
		}
	}
	return hash & mask
}

// Zobrist Zobrist 哈希：按"字节位置 + 字节值"查随机表异或，最后 & mask 收拢。
// tab 为扁平化的二维随机表，按 tab[i*256+b] 索引（i 为字节位置、b 为字节值，
// 对应 hutool 的 int[][] tab），长度需 >= len(key)*256，不足时 panic。
// 场景：棋局状态散列等增量更新式哈希。
func (h *utilHash) Zobrist(key string, mask uint32, tab []uint32) uint32 {
	hash := uint32(len(key))
	for i := 0; i < len(key); i++ {
		hash ^= tab[i*256+int(key[i])]
	}
	return hash & mask
}

// FnvHash 改进版 32 位 FNV-1a：先异或后乘质数 16777619，再做多轮移位雪崩，结果按 32 位有符号取绝对值。
// 场景：hutool 推荐的通用散列，缓存键、分片路由等。
func (h *utilHash) FnvHash(key string) uint32 {
	const p = 16777619
	hash := uint32(2166136261)
	for i := 0; i < len(key); i++ {
		hash = (hash ^ uint32(key[i])) * p
	}
	hash += hash << 13
	hash ^= hashSar32(hash, 7)
	hash += hash << 3
	hash ^= hashSar32(hash, 17)
	hash += hash << 5
	return hashAbs32(hash)
}

// RsHash RS 哈希：乘数 a 从 63689 起每步自乘 378551 变化，混合字节值后取正。
// 场景：字符串散列，分布较均匀，适合哈希表。
func (h *utilHash) RsHash(key string) uint32 {
	a := uint32(63689)
	var hash uint32
	for i := 0; i < len(key); i++ {
		hash = hash*a + uint32(key[i])
		a *= 378551
	}
	return hash & 0x7FFFFFFF
}

// JsHash JS 哈希：初值 1315423911，每步 (hash<<5) + 字节 + (hash>>2) 异或混合，结果按有符号取绝对值。
// 场景：通用字符串散列。
func (h *utilHash) JsHash(key string) uint32 {
	hash := uint32(1315423911)
	for i := 0; i < len(key); i++ {
		hash ^= (hash << 5) + uint32(key[i]) + hashSar32(hash, 2)
	}
	return hashAbs32(hash) & 0x7FFFFFFF
}

// PjwHash PJW 哈希：每步左移 4 位累加，高 4 位溢出时折返异或（Peter J. Weinberger 的经典算法）。
// 场景：符号表散列，ELF 系统的前身算法。
func (h *utilHash) PjwHash(key string) uint32 {
	// bitsInUnsignedInt=32，threeQuarters=24，oneEighth=4；
	// highBits 即 hutool 的 0xFFFFFFFF << 28（Java int 32 位环绕后为 0xF0000000）
	const (
		oneEighth     = 4
		threeQuarters = 24
		highBits      = uint32(0xF0000000)
	)
	var hash uint32
	for i := 0; i < len(key); i++ {
		hash = (hash << oneEighth) + uint32(key[i])
		if test := hash & highBits; test != 0 {
			hash = (hash ^ hashSar32(test, threeQuarters)) & ^highBits
		}
	}
	return hash & 0x7FFFFFFF
}

// ElfHash ELF 哈希：PjwHash 的变体，每步左移 4 位累加，高 4 位非 0 时异或折返并清除。
// 场景：Unix ELF 文件符号表散列，通用字符串散列。
func (h *utilHash) ElfHash(key string) uint32 {
	var hash uint32
	for i := 0; i < len(key); i++ {
		hash = (hash << 4) + uint32(key[i])
		if x := hash & 0xF0000000; x != 0 {
			hash ^= hashSar32(x, 24)
			hash &= ^x
		}
	}
	return hash & 0x7FFFFFFF
}

// BkdrHash BKDR 哈希：hash = hash*131 + 字节（种子可取 31/131/1313/13131 等）。
// 场景：字符串散列，实现简单且分布良好。
func (h *utilHash) BkdrHash(key string) uint32 {
	const seed = 131
	var hash uint32
	for i := 0; i < len(key); i++ {
		hash = hash*seed + uint32(key[i])
	}
	return hash & 0x7FFFFFFF
}

// SdbmHash SDBM 哈希：hash = 字节 + (hash<<6) + (hash<<16) - hash。
// 场景：SDBM 数据库开源项目采用的散列，散列效果优于 BKDR。
func (h *utilHash) SdbmHash(key string) uint32 {
	var hash uint32
	for i := 0; i < len(key); i++ {
		hash = uint32(key[i]) + (hash << 6) + (hash << 16) - hash
	}
	return hash & 0x7FFFFFFF
}

// DjbHash DJB 哈希：hash = ((hash<<5) + hash) + 字节，初值 5381，结果取正。
// 场景：通用字符串散列（与 Bernstein 同族，此处按 hutool 原版实现）。
func (h *utilHash) DjbHash(key string) uint32 {
	hash := uint32(5381)
	for i := 0; i < len(key); i++ {
		hash = ((hash << 5) + hash) + uint32(key[i])
	}
	return hash & 0x7FFFFFFF
}

// DekHash DEK 哈希：初值为串长，每步 ((hash<<5) ^ (hash>>27)) ^ 字节。
// 场景：Knuth 提出的字符串散列。
func (h *utilHash) DekHash(key string) uint32 {
	hash := uint32(len(key))
	for i := 0; i < len(key); i++ {
		hash = ((hash << 5) ^ hashSar32(hash, 27)) ^ uint32(key[i])
	}
	return hash & 0x7FFFFFFF
}

// ApHash Arash Partow 哈希：按位置奇偶交替使用不同的移位组合异或混合。
// 场景：通用字符串散列。
func (h *utilHash) ApHash(key string) uint32 {
	var hash uint32
	for i := 0; i < len(key); i++ {
		c := uint32(key[i])
		if i%2 == 0 {
			hash ^= (hash << 7) ^ c ^ hashSar32(hash, 3)
		} else {
			hash ^= ^((hash << 11) ^ c ^ hashSar32(hash, 5))
		}
	}
	return hash
}

// TianlHash TianL 哈希：基于"位置 × 字符平方"多项式取模累加，输出 64 位；
// 串长超过 96 时仅取末尾 96 个字节参与计算（对应 hutool tianlHash 的 long 返回值）。
// 场景：hutool 兼容场景的字符串散列。
func (h *utilHash) TianlHash(key string) int64 {
	iLength := len(key)
	if iLength == 0 {
		return 0
	}
	var hash int64
	if iLength <= 256 {
		hash = 16777216 * int64(iLength-1)
	} else {
		hash = 4278190080
	}
	if iLength <= 96 {
		for i := 1; i <= iLength; i++ {
			c := key[i-1]
			if c >= 'A' && c <= 'Z' {
				c += 'a' - 'A'
			}
			u := int64(c)
			hash += (3*int64(i)*u*u + 5*int64(i)*u + 7*int64(i) + 11*u) % 16777216
		}
	} else {
		for i := 1; i <= 96; i++ {
			c := key[i+iLength-96-1]
			if c >= 'A' && c <= 'Z' {
				c += 'a' - 'A'
			}
			u := int64(c)
			hash += (3*int64(i)*u*u + 5*int64(i)*u + 7*int64(i) + 11*u) % 16777216
		}
	}
	if hash < 0 {
		hash *= -1
	}
	return hash
}

// JavaDefaultHash 与 Java String.hashCode() 完全一致的有符号 32 位哈希：
// 按 h = 31*h + UTF-16 code unit 计算，BMP 外字符按 Java 规则拆成代理对（两个 code unit）参与运算，
// 保证与 JVM 端对同一字符串的计算结果一致。
// 场景：与 Java 服务的哈希值互认，如一致性问题路由、缓存键对齐。
func (h *utilHash) JavaDefaultHash(key string) int32 {
	var hash int32
	for _, unit := range utf16.Encode([]rune(key)) {
		hash = 31*hash + int32(unit)
	}
	return hash
}

// MixHash 混合哈希（64 位）：高 32 位为 JavaDefaultHash（String.hashCode），低 32 位为 FnvHash。
// 场景：需要 64 位散列值又不引入复杂算法的场合，如分表路由取模。
func (h *utilHash) MixHash(key string) int64 {
	hash := int64(h.JavaDefaultHash(key))
	hash <<= 32
	hash |= int64(int32(h.FnvHash(key)))
	return hash
}

/* ---------------- Murmur3 ---------------- */

// Murmur32 MurmurHash3 x86 32 位哈希（seed=0，输入为字符串 UTF-8 字节）。
// 场景：布隆过滤器、一致性哈希、分布式分片，与主流 MurmurHash3 实现互通。
func (h *utilHash) Murmur32(key string) uint32 {
	return hashMurmur32([]byte(key), 0)
}

// Murmur64 hutool 版 64 位 Murmur 哈希（seed=0）：以 128 位算法的常数逐 8 字节块混合，
// 对应 hutool MurmurHash.hash64（其文档标注为 128 位结果的主 64 位思想，非标准 x64 64 位截断）。
// 场景：需要 64 位指纹的场合，如 SimHash、去重键。
func (h *utilHash) Murmur64(key string) uint64 {
	return hashMurmur64([]byte(key), 0)
}

// Murmur128 MurmurHash3 x64 128 位哈希（seed=0），返回 [低位, 高位] 两个 uint64（对应 hutool long[0]/long[1]）。
// 场景：高精度指纹、HyperLogLog 等需要 128 位散列的场景。
func (h *utilHash) Murmur128(key string) [2]uint64 {
	return hashMurmur128([]byte(key), 0)
}

/* ---------------- 内部实现 ---------------- */

// Murmur3 x86 32 位变体常量（C1/C2 混合乘数，R1/R2 移位数，M/N 主体混合参数）
const (
	hashMurmur32C1 = 0xcc9e2d51
	hashMurmur32C2 = 0x1b873593
	hashMurmur32R1 = 15
	hashMurmur32R2 = 13
	hashMurmur32M  = 5
	hashMurmur32N  = 0xe6546b64
)

// Murmur3 x64 128 位变体常量（hutool hash64 / hash128 共用）
const (
	hashMurmurC1 = 0x87c37b91114253d5
	hashMurmurC2 = 0x4cf5ad432745937f
	hashMurmurR1 = 31
	hashMurmurR2 = 27
	hashMurmurR3 = 33
	hashMurmurM  = 5
	hashMurmurN1 = 0x52dce729
	hashMurmurN2 = 0x38495ab5
)

// hashSar32 算术右移：对应 Java 有符号 int 的 >>（符号位扩展）
func hashSar32(v uint32, n uint) uint32 {
	return uint32(int32(v) >> n)
}

// hashAbs32 按 32 位有符号语义取绝对值：对应 Java Math.abs(int)（Integer.MIN_VALUE 取反仍为自身）
func hashAbs32(v uint32) uint32 {
	if v>>31 == 1 {
		return -v
	}
	return v
}

// hashFmix32 Murmur3 32 位终态混合
func hashFmix32(hash uint32) uint32 {
	hash ^= hash >> 16
	hash *= 0x85ebca6b
	hash ^= hash >> 13
	hash *= 0xc2b2ae35
	hash ^= hash >> 16
	return hash
}

// hashFmix64 Murmur3 64 位终态混合
func hashFmix64(h uint64) uint64 {
	h ^= h >> 33
	h *= 0xff51afd7ed558ccd
	h ^= h >> 33
	h *= 0xc4ceb9fe1a85ec53
	h ^= h >> 33
	return h
}

// hashMurmur32 MurmurHash3 x86 32 位实现（对应 hutool MurmurHash.hash32）
func hashMurmur32(data []byte, seed uint32) uint32 {
	hash := seed
	nblocks := len(data) >> 2
	// 主体：按 4 字节小端块混合
	for i := 0; i < nblocks; i++ {
		k := binary.LittleEndian.Uint32(data[i<<2:])
		k *= hashMurmur32C1
		k = bits.RotateLeft32(k, hashMurmur32R1)
		k *= hashMurmur32C2
		hash ^= k
		hash = bits.RotateLeft32(hash, hashMurmur32R2)*hashMurmur32M + hashMurmur32N
	}
	// 尾部：处理不足 4 字节的剩余部分（switch 逐级 fallthrough，同 Java 原实现）
	var k1 uint32
	tail := data[nblocks<<2:]
	switch len(tail) {
	case 3:
		k1 ^= uint32(tail[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint32(tail[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint32(tail[0])
		k1 *= hashMurmur32C1
		k1 = bits.RotateLeft32(k1, hashMurmur32R1)
		k1 *= hashMurmur32C2
		hash ^= k1
	}
	// 终态
	hash ^= uint32(len(data))
	return hashFmix32(hash)
}

// hashMurmur64 hutool 版 64 位实现（对应 MurmurHash.hash64，128 位常数处理 8 字节块）
func hashMurmur64(data []byte, seed uint64) uint64 {
	hash := seed
	nblocks := len(data) >> 3
	// 主体：按 8 字节小端块混合
	for i := 0; i < nblocks; i++ {
		k := binary.LittleEndian.Uint64(data[i<<3:])
		k *= hashMurmurC1
		k = bits.RotateLeft64(k, hashMurmurR1)
		k *= hashMurmurC2
		hash ^= k
		hash = bits.RotateLeft64(hash, hashMurmurR2)*hashMurmurM + hashMurmurN1
	}
	// 尾部：处理不足 8 字节的剩余部分
	var k1 uint64
	tail := data[nblocks<<3:]
	switch len(tail) {
	case 7:
		k1 ^= uint64(tail[6]) << 48
		fallthrough
	case 6:
		k1 ^= uint64(tail[5]) << 40
		fallthrough
	case 5:
		k1 ^= uint64(tail[4]) << 32
		fallthrough
	case 4:
		k1 ^= uint64(tail[3]) << 24
		fallthrough
	case 3:
		k1 ^= uint64(tail[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint64(tail[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint64(tail[0])
		k1 *= hashMurmurC1
		k1 = bits.RotateLeft64(k1, hashMurmurR1)
		k1 *= hashMurmurC2
		hash ^= k1
	}
	// 终态
	hash ^= uint64(len(data))
	return hashFmix64(hash)
}

// hashMurmur128 MurmurHash3 x64 128 位实现（对应 hutool MurmurHash.hash128）
func hashMurmur128(data []byte, seed uint64) [2]uint64 {
	seed &= 0xffffffff // 避免负数种子（同 hutool）
	h1, h2 := seed, seed
	nblocks := len(data) >> 4
	// 主体：按 16 字节小端双块混合
	for i := 0; i < nblocks; i++ {
		i16 := i << 4
		k1 := binary.LittleEndian.Uint64(data[i16:])
		k2 := binary.LittleEndian.Uint64(data[i16+8:])

		// k1 混合
		k1 *= hashMurmurC1
		k1 = bits.RotateLeft64(k1, hashMurmurR1)
		k1 *= hashMurmurC2
		h1 ^= k1
		h1 = bits.RotateLeft64(h1, hashMurmurR2)
		h1 += h2
		h1 = h1*hashMurmurM + hashMurmurN1

		// k2 混合
		k2 *= hashMurmurC2
		k2 = bits.RotateLeft64(k2, hashMurmurR3)
		k2 *= hashMurmurC1
		h2 ^= k2
		h2 = bits.RotateLeft64(h2, hashMurmurR1)
		h2 += h1
		h2 = h2*hashMurmurM + hashMurmurN2
	}
	// 尾部：先混 k2（第 9 字节起），再混 k1（前 8 字节），逐级 fallthrough 同 Java 原实现
	var k1, k2 uint64
	tail := data[nblocks<<4:]
	switch len(tail) {
	case 15:
		k2 ^= uint64(tail[14]) << 48
		fallthrough
	case 14:
		k2 ^= uint64(tail[13]) << 40
		fallthrough
	case 13:
		k2 ^= uint64(tail[12]) << 32
		fallthrough
	case 12:
		k2 ^= uint64(tail[11]) << 24
		fallthrough
	case 11:
		k2 ^= uint64(tail[10]) << 16
		fallthrough
	case 10:
		k2 ^= uint64(tail[9]) << 8
		fallthrough
	case 9:
		k2 ^= uint64(tail[8])
		k2 *= hashMurmurC2
		k2 = bits.RotateLeft64(k2, hashMurmurR3)
		k2 *= hashMurmurC1
		h2 ^= k2
		fallthrough
	case 8:
		k1 ^= uint64(tail[7]) << 56
		fallthrough
	case 7:
		k1 ^= uint64(tail[6]) << 48
		fallthrough
	case 6:
		k1 ^= uint64(tail[5]) << 40
		fallthrough
	case 5:
		k1 ^= uint64(tail[4]) << 32
		fallthrough
	case 4:
		k1 ^= uint64(tail[3]) << 24
		fallthrough
	case 3:
		k1 ^= uint64(tail[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint64(tail[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint64(tail[0])
		k1 *= hashMurmurC1
		k1 = bits.RotateLeft64(k1, hashMurmurR1)
		k1 *= hashMurmurC2
		h1 ^= k1
	}
	// 终态
	h1 ^= uint64(len(data))
	h2 ^= uint64(len(data))

	h1 += h2
	h2 += h1

	h1 = hashFmix64(h1)
	h2 = hashFmix64(h2)

	h1 += h2
	h2 += h1

	return [2]uint64{h1, h2}
}
