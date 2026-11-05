package utils

import (
	crand "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	mrand "math/rand/v2"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// 雪花算法位分配常量（对照 hutool Snowflake 的同名常量，加 sf 前缀避免包内冲突）
const (
	// sfDefaultTwepoch 默认起始纪元：Thu, 04 Nov 2010 01:42:54 GMT（对照 hutool DEFAULT_TWEPOCH）
	sfDefaultTwepoch int64 = 1288834974657
	// sfDefaultTimeOffset 默认容忍的时钟回拨毫秒数：2 秒（对照 hutool DEFAULT_TIME_OFFSET）
	sfDefaultTimeOffset int64 = 2000

	sfWorkerIdBits       = 5                                                    // 机器 ID 位数
	sfDataCenterIdBits   = 5                                                    // 数据中心 ID 位数
	sfSequenceBits       = 12                                                   // 毫秒内序列号位数
	sfMaxWorkerId        = -1 ^ (-1 << sfWorkerIdBits)                          // 机器 ID 最大值 31
	sfMaxDataCenterId    = -1 ^ (-1 << sfDataCenterIdBits)                      // 数据中心 ID 最大值 31
	sfSequenceMask       = -1 ^ (-1 << sfSequenceBits)                          // 序列掩码，限制序列最大 4095
	sfWorkerIdShift      = sfSequenceBits                                       // 机器 ID 左移位数
	sfDataCenterIdShift  = sfSequenceBits + sfWorkerIdBits                      // 数据中心 ID 左移位数
	sfTimestampLeftShift = sfSequenceBits + sfWorkerIdBits + sfDataCenterIdBits // 时间戳左移位数
)

// Snowflake 雪花 ID 生成器（Twitter Snowflake 算法，结构对照 hutool Snowflake）。
//
// ID 结构（每部分用 - 分隔）：
//
//	0 - 0000000000 0000000000 0000000000 0000000000 0 - 00000 - 00000 - 000000000000
//	符号位 - 41 位毫秒时间戳 - 5 位数据中心 - 5 位机器 - 12 位序列
//
// 41 位时间戳可用 69 年，10 位节点标识最多支持 1024 个节点（32 数据中心 x 32 机器），
// 12 位序列支持单节点每毫秒 4096 个 ID，ID 总体按时间趋势递增。
// 实例内部用互斥锁保证并发安全；同一实例必须全进程单例使用（多实例会产生重复 ID），
// 快捷方式：utils.Id.SnowflakeNextId()（包级默认单例）或 utils.NewSnowflake(workerId, datacenterId)。
type Snowflake struct {
	mu            sync.Mutex // 保护下列可变状态，保证多协程下 ID 不重复
	twepoch       int64      // 起始纪元，ID 中记录的是相对该时间点的毫秒数
	workerId      int64      // 机器 ID（0~31）
	dataCenterId  int64      // 数据中心 ID（0~31）
	timeOffset    int64      // 容忍的时钟回拨毫秒数（默认 2000）
	sequence      int64      // 当前毫秒内的序列号（0~4095）
	lastTimestamp int64      // 上次生成 ID 的时间戳（毫秒），-1 表示尚未生成
}

// newSnowflake 创建雪花 ID 生成器（utils.NewSnowflake 的实现）。
// workerId/datacenterId 合法范围为 0~31：hutool 越界时通过 Assert 直接抛出 IllegalArgumentException；
// Go 版本选择收敛到合法范围（对 32 取模，负数归一化为非负，如 -1 -> 31、33 -> 1），
// 避免构造时 panic 中断业务，同时保证生成的 ID 结构仍然合法。
func newSnowflake(workerId, datacenterId int64) *Snowflake {
	return &Snowflake{
		twepoch:       sfDefaultTwepoch,
		workerId:      sfClampId(workerId),
		dataCenterId:  sfClampId(datacenterId),
		timeOffset:    sfDefaultTimeOffset,
		lastTimestamp: -1,
	}
}

// sfClampId 把节点 ID 收敛到 [0,31]：先对 32 取模，负数再加 32 归一化为非负
func sfClampId(v int64) int64 {
	v %= sfMaxWorkerId + 1
	if v < 0 {
		v += sfMaxWorkerId + 1
	}
	return v
}

// NextId 生成下一个全局唯一 ID（线程安全）。
//
// 生成规则：当前毫秒时间戳相对起始纪元的差值左移 22 位，或上数据中心 ID 左移 17 位、
// 机器 ID 左移 12 位、毫秒内序列号。同一毫秒内序列号自增，耗尽（超过 4095）时
// 自旋等待到下一毫秒再继续。同一实例内的 ID 严格按时间趋势递增。
//
// 时钟回拨处理（对照 hutool）：回拨幅度小于 timeOffset（默认 2 秒，多为 NTP 微调）时，
// hutool 沿用上一时间戳继续生成以容忍回拨；回拨过大时 hutool 抛出 IllegalStateException。
// Go 版本为不中断服务，回拨过大时同样沿用上一时间戳继续生成（不抛错），
// 差异说明：极端回拨期间生成的 ID 时间部分会略超前于真实时间，但唯一性不受影响；
// 需要严格感知时钟异常的场景请自行比对系统时间。
func (s *Snowflake) NextId() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	timestamp := s.genTime()
	if timestamp < s.lastTimestamp {
		// 时钟回拨：小于 timeOffset 容忍（hutool 同款），过大时 hutool 抛错、此处沿用上一时间戳继续
		timestamp = s.lastTimestamp
	}

	if timestamp == s.lastTimestamp {
		// 同一毫秒：序列号自增，超出 4095 则自旋到下一毫秒
		s.sequence = (s.sequence + 1) & sfSequenceMask
		if s.sequence == 0 {
			timestamp = s.tilNextMillis(s.lastTimestamp)
		}
	} else {
		// 新的毫秒：序列号归零重新计数（hutool 的随机起始序列默认 randomSequenceLimit=0 不启用）
		s.sequence = 0
	}
	s.lastTimestamp = timestamp

	return ((timestamp - s.twepoch) << sfTimestampLeftShift) |
		(s.dataCenterId << sfDataCenterIdShift) |
		(s.workerId << sfWorkerIdShift) |
		s.sequence
}

// NextIdStr 同 NextId，返回字符串形式。场景：ID 直接进 JSON、URL、日志等文本载体。
func (s *Snowflake) NextIdStr() string {
	return strconv.FormatInt(s.NextId(), 10)
}

// tilNextMillis 自旋等待直到系统时钟越过 lastTimestamp（到达下一毫秒）。
// 对照 hutool：hutool 在自旋中发现时间倒退会抛异常；本实现为容忍回拨继续自旋，
// 与 NextId 的时钟回拨策略保持一致。
func (s *Snowflake) tilNextMillis(lastTimestamp int64) int64 {
	timestamp := s.genTime()
	for timestamp <= lastTimestamp {
		timestamp = s.genTime()
	}
	return timestamp
}

// genTime 获取当前毫秒时间戳
func (s *Snowflake) genTime() int64 {
	return time.Now().UnixMilli()
}

// utilId ID 生成工具。
//
// 命名约定：SimpleUUID 为 32 位无连字符短 UUID；SnowflakeNextId 走包级默认单例；
// 需要显式分配节点号（多实例部署）时用 utils.NewSnowflake 自行创建实例。
// 本对象中 UUID/ObjectId/NanoId 使用 crypto/rand 密码学安全随机源。
type utilId struct{}

// SimpleUUID 生成 32 位无连字符的 UUID v4 小写十六进制串（对照 hutool simpleUUID）。
// 返回形如 "9a8b7c6d0e1f4a2b3c4d5e6f7a8b9c0d"，密码学安全随机源，含版本位与变体位。
// 场景：请求追踪 ID、分布式去重键；需要标准带连字符格式时按 8-4-4-4-12 自行插入。
func (i *utilId) SimpleUUID() string {
	var uuid [16]byte
	// Go 1.24 起 crypto/rand.Read 保证填满且不返回错误；保留检查以兼容旧版本语义
	if _, err := crand.Read(uuid[:]); err != nil {
		return ""
	}
	uuid[6] = (uuid[6] & 0x0F) | 0x40 // 版本号置为 4（随机 UUID）
	uuid[8] = (uuid[8] & 0x3F) | 0x80 // 变体位 置为 RFC 4122
	return hex.EncodeToString(uuid[:])
}

// ObjectId 相关的进程级状态：机器标识进程内固定，计数器进程内自增。
var (
	// objectIdMachine ObjectId 的 5 字节机器标识：进程启动时一次性随机生成、进程内固定。
	// 对照 hutool ObjectId.MACHINE（网络接口信息散列 + 进程 ID），Go 简化为随机值，
	// 作用等价：区分不同进程/机器生成的 ID。
	objectIdMachine = idRandomBytes(5)
	// objectIdInc ObjectId 的 3 字节自增计数器：初值随机，每生成一次加 1（模 2^24 回绕），
	// 保证同一进程同一秒内生成的 ObjectId 唯一（对照 hutool NEXT_INC 初值 randomInt）。
	objectIdInc = newObjectIdInc()
)

// newObjectIdInc 初始化随机起点的自增计数器
func newObjectIdInc() *atomic.Uint32 {
	b := idRandomBytes(3)
	c := new(atomic.Uint32)
	c.Store(uint32(b[0])<<16 | uint32(b[1])<<8 | uint32(b[2]))
	return c
}

// ObjectId 生成 MongoDB 风格的 ObjectId：24 位小写十六进制（12 字节）。
// 结构：4 字节秒级时间戳 + 5 字节机器标识（进程内固定随机值）+ 3 字节自增计数器，
// 可从 ID 反推生成时间，且同一进程内按时间趋势递增。
// 对照 hutool ObjectId.next()：hutool 机器段为"主机信息散列 + 进程 ID"，Go 用随机值等价替代。
// 场景：无协调服务的分布式主键、消息/工单等业务单号前缀。
func (i *utilId) ObjectId() string {
	var b [12]byte
	binary.BigEndian.PutUint32(b[0:4], uint32(time.Now().Unix())) // 4 字节秒级时间
	copy(b[4:9], objectIdMachine)                                 // 5 字节机器标识
	inc := (objectIdInc.Add(1) - 1) & 0xFFFFFF                    // 3 字节自增计数（回绕）
	b[9] = byte(inc >> 16)
	b[10] = byte(inc >> 8)
	b[11] = byte(inc)
	return hex.EncodeToString(b[:])
}

// idNanoAlphabet NanoID 字母表：URL 安全的 64 个字符（对照 hutool NanoId.DEFAULT_ALPHABET）
const idNanoAlphabet = "-0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// NanoId 生成长度为 size 的 NanoID（小型、URL 安全的唯一字符串 ID，对照 hutool nanoId）。
// 使用 crypto/rand 密码学安全随机源 + 拒绝采样，保证每个字符等概率分布；
// size<=0 时使用默认长度 21（hutool 同款默认值；hutool 对非法 size 抛异常，Go 便捷处理为默认）。
// 场景：短邀请码、分享标识、会暴露给前端的随机 ID（比 UUID 短且 URL 安全）。
func (i *utilId) NanoId(size int) string {
	if size <= 0 {
		size = 21
	}
	// 拒绝采样：字母表长度为 64，256 能被其整除，任意随机字节均落在合法范围内，
	// 实际不会触发丢弃，但保留通用拒绝采样写法以忠实对照 hutool/官方 NanoId 算法
	max := 256 - 256%len(idNanoAlphabet)
	out := make([]byte, 0, size)
	buf := make([]byte, size)
	for len(out) < size {
		if _, err := crand.Read(buf); err != nil {
			// 理论不可达（Go 1.24 起 crypto/rand.Read 保证不失败），兜底返回已生成部分
			return string(out)
		}
		for _, v := range buf {
			if int(v) >= max {
				continue // 拒绝采样：超出字母表覆盖范围的字节丢弃重取
			}
			out = append(out, idNanoAlphabet[int(v)%len(idNanoAlphabet)])
			if len(out) == size {
				break
			}
		}
	}
	return string(out)
}

// 包级默认雪花 ID 生成器：懒加载单例，全进程仅初始化一次（对照 hutool getSnowflake()）
var (
	idDefaultSnowflakeOnce sync.Once
	idDefaultSnowflake     *Snowflake
)

// SnowflakeNextId 用包级默认雪花 ID 生成器生成下一个 ID（int64，线程安全）。
// 默认单例懒加载且全进程唯一：workerId/datacenterId 在首次调用时以 0~31 的随机值确定，
// 之后整个进程复用同一实例（对照 hutool getSnowflake()：hutool 由 PID+MAC 推导节点号，
// Go 简化为随机初始化一次）。注意：随机节点号在多实例部署时有小概率撞号，
// 严格场景请用 utils.NewSnowflake 显式分配互不冲突的节点号。
func (i *utilId) SnowflakeNextId() int64 {
	idDefaultSnowflakeOnce.Do(func() {
		idDefaultSnowflake = newSnowflake(int64(idRandomIntn(32)), int64(idRandomIntn(32)))
	})
	return idDefaultSnowflake.NextId()
}

// SnowflakeNextIdStr 同 SnowflakeNextId，返回字符串形式。场景：ID 直接进 JSON、URL、日志。
func (i *utilId) SnowflakeNextIdStr() string {
	return strconv.FormatInt(i.SnowflakeNextId(), 10)
}

// idRandomBytes 读取 n 个密码学安全随机字节。
// Go 1.24 起 crypto/rand.Read 保证填满且不返回错误；错误分支仅为极端旧环境兜底
// （退化为自动播种的伪随机），保证仍能产出格式合法的 ID。
func idRandomBytes(n int) []byte {
	b := make([]byte, n)
	if _, err := crand.Read(b); err != nil {
		for idx := range b {
			b[idx] = byte(mrand.Uint64())
		}
	}
	return b
}

// idRandomIntn 返回 [0, n) 内的密码学安全随机整数，n<=0 时返回 0。
// 用拒绝采样消除模偏差，保证 [0, n) 内每个取值等概率。
func idRandomIntn(n int) int {
	if n <= 0 {
		return 0
	}
	max := 256 - 256%n
	for {
		b := idRandomBytes(1)
		if int(b[0]) < max {
			return int(b[0]) % n
		}
	}
}
