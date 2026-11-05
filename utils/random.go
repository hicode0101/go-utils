package utils

import (
	mrand "math/rand/v2"
	"strings"
)

// utilRandom 随机工具。
//
// 随机源强度说明：本对象方法统一使用 math/rand/v2 自动播种的伪随机源
// （对照 hutool RandomUtil 默认的 ThreadLocalRandom，非密码学安全，具备协程安全性，
// 适合模拟数据、抽样、负载均衡等常规业务场景）；
// 安全敏感场景（验证码、口令、NanoID）请使用 utils.Id 下的密码学安全实现。
//
// 泛型说明：Go 语言限制方法不能自带类型参数，RandomEle/RandomEles/RandomEleSet
// 三个泛型方法以包级泛型函数形式提供（用法 utils.RandomEle(slice)，
// 与包内 IndexOf/Shuffle 等集合泛型函数约定一致）。
type utilRandom struct{}

// randomIntn 统一的随机整数入口：返回 [0, n) 内的伪随机数（math/rand/v2，自动播种、并发安全）。
// n<=0 时返回 0，调用方无需重复判空。
func randomIntn(n int) int {
	if n <= 0 {
		return 0
	}
	return mrand.IntN(n)
}

// RandomBool 随机布尔值，true/false 各 50% 概率（对照 hutool randomBoolean）。
// 场景：灰度开关、模拟数据。
func (r *utilRandom) RandomBool() bool {
	return randomIntn(2) == 0
}

// RandomBytes 生成长度为 length 的随机字节数组（非密码学安全，见类型注释）。
// length<=0 返回空数组。场景：测试数据填充、模拟二进制报文。
func (r *utilRandom) RandomBytes(length int) []byte {
	if length <= 0 {
		return nil
	}
	b := make([]byte, length)
	for i := range b {
		b[i] = byte(randomIntn(256))
	}
	return b
}

// RandomInt 返回 [min, max) 区间内的随机整数（左闭右开）；min>=max 时返回 min。
// 场景：模拟数据取值范围，如 RandomInt(1, 7) 模拟骰子点数 1~6。
func (r *utilRandom) RandomInt(min, max int) int {
	if min >= max {
		return min
	}
	return min + randomIntn(max-min)
}

// RandomInt64 返回 [min, max) 区间内的随机 int64（左闭右开）；min>=max 时返回 min。
// 区间宽度可能超出 int 范围，故单独走 Int64N（随机源与 randomIntn 相同）。
func (r *utilRandom) RandomInt64(min, max int64) int64 {
	if min >= max {
		return min
	}
	return min + mrand.Int64N(max-min)
}

// RandomFloat 返回 [min, max) 区间内的随机浮点数；min>=max 时返回 min。
// 场景：模拟价格、概率采样。
func (r *utilRandom) RandomFloat(min, max float64) float64 {
	if min >= max {
		return min
	}
	return min + (max-min)*mrand.Float64()
}

// RandomEle 随机返回切片中的一个元素；空切片（或 nil）返回元素零值
// （hutool 对 null 列表返回 null，Go 无此语义，以零值代替）。
// 因 Go 方法不能自带类型参数，本方法以包级泛型函数形式提供（用法 utils.RandomEle(slice)）。
// 场景：抽奖、负载均衡随机路由。
func RandomEle[T any](slice []T) T {
	var zero T
	if len(slice) == 0 {
		return zero
	}
	return slice[randomIntn(len(slice))]
}

// RandomEles 随机返回切片中的 count 个元素：每次独立抽取、允许重复（对照 hutool randomEles）。
// count<=0 或源切片为空时返回空切片（hutool 对空列表会抛异常，Go 返回空切片更安全）。
// 因 Go 方法不能自带类型参数，本方法以包级泛型函数形式提供（用法 utils.RandomEles(slice, n)）。
// 场景：模拟数据批量生成、有放回抽样。
func RandomEles[T any](slice []T, count int) []T {
	if len(slice) == 0 || count <= 0 {
		return []T{}
	}
	result := make([]T, 0, count)
	for len(result) < count {
		result = append(result, slice[randomIntn(len(slice))])
	}
	return result
}

// RandomEleSet 不重复地随机返回切片中的 count 个元素（对照 hutool randomEleSet 的 Set 语义）。
// 源切片先去重；count 超过去重后的元素总数时返回全部元素（顺序随机；
// hutool 此场景抛 IllegalArgumentException，Go 收敛为取全部）；count<=0 返回空切片。
// 因 Go 方法不能自带类型参数，本方法以包级泛型函数形式提供（用法 utils.RandomEleSet(slice, n)）。
// 场景：无放回抽奖、随机抽查样本。
func RandomEleSet[T comparable](slice []T, count int) []T {
	// 先去重（对照 hutool CollUtil.distinct）
	seen := make(map[T]struct{}, len(slice))
	uniq := make([]T, 0, len(slice))
	for _, v := range slice {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			uniq = append(uniq, v)
		}
	}
	if count > len(uniq) {
		count = len(uniq)
	}
	result := make([]T, 0, count)
	if count <= 0 {
		return result
	}
	// 部分洗牌（Fisher-Yates 前 count 步）：每个不重复组合等概率被选出，且顺序随机
	for i := 0; i < count; i++ {
		j := i + randomIntn(len(uniq)-i)
		uniq[i], uniq[j] = uniq[j], uniq[i]
		result = append(result, uniq[i])
	}
	return result
}

// RandomString 生成长度为 length 的随机字符串（字母 + 数字 [a-zA-Z0-9]）。
// 委托 utils.String.RandomString（密码学安全随机源实现），避免同功能两套算法并存。
// 场景：验证码、临时文件名、请求 ID。
func (r *utilRandom) RandomString(length int) string {
	return String.RandomString(length)
}

// randomBaseNumber 纯数字字符集（对照 hutool BASE_NUMBER）
const randomBaseNumber = "0123456789"

// RandomNumbers 生成长度为 length 的纯数字随机串（对照 hutool randomNumbers）。
// length<=0 返回空串。场景：短信验证码、数字口令。
func (r *utilRandom) RandomNumbers(length int) string {
	if length <= 0 {
		return ""
	}
	b := make([]byte, length)
	for i := range b {
		b[i] = randomBaseNumber[randomIntn(len(randomBaseNumber))]
	}
	return string(b)
}

// randomBaseCharNumber 字母 + 数字字符集（对照 hutool BASE_CHAR_NUMBER：大写 + 小写 + 数字）
const randomBaseCharNumber = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// RandomStringWithoutStr 生成长度为 length 的随机字符串，并从字符集 [a-zA-Z0-9] 中
// 剔除 excludeChars 里出现的字符（区分大小写，对照 hutool randomStringWithoutStr）。
// 字符全部被剔除、或 length<=0 时返回空串。
// 场景：生成人眼易读的口令，剔除 oO0、lL1、q9Q、pP 等易混淆字符。
func (r *utilRandom) RandomStringWithoutStr(length int, excludeChars string) string {
	if length <= 0 {
		return ""
	}
	// 剔除 excludeChars 内的字符（对照 hutool StrUtil.removeAll）
	base := make([]rune, 0, len(randomBaseCharNumber))
	for _, c := range randomBaseCharNumber {
		if !strings.ContainsRune(excludeChars, c) {
			base = append(base, c)
		}
	}
	if len(base) == 0 {
		return ""
	}
	out := make([]rune, length)
	for i := range out {
		out[i] = base[randomIntn(len(base))]
	}
	return string(out)
}

// RandomChar 从给定字符集 base 中随机取一个字符（rune，支持中文等非 ASCII 字符集）；
// base 为空时返回 0（对应 hutool 的越界异常场景，Go 以空字符表示）。
// 场景：验证码单字符生成、随机符号。
func (r *utilRandom) RandomChar(base string) rune {
	runes := []rune(base)
	if len(runes) == 0 {
		return 0
	}
	return runes[randomIntn(len(runes))]
}

// RandomChinese 随机返回一个常用汉字（Unicode CJK 基本区 [0x4E00, 0x9FA5]；
// 对照 hutool randomChinese，hutool 区间为 [0x4E00, 0x9FFF)，此处按常用汉字收尾 0x9FA5 收敛）。
// 场景：模拟中文昵称、测试数据。
func (r *utilRandom) RandomChinese() string {
	return string(rune(0x4E00 + randomIntn(0x9FA5-0x4E00+1)))
}

// WeightRandom 加权随机选取器（对照 hutool WeightRandom）。
//
// 思路：把每个元素的权重累加成连续区间，随机数落在哪个元素的区间即命中该元素。
// 如 A、B、C、D 权重 1、2、3、4，区间分别为 [0,1)、[1,3)、[3,6)、[6,10)，
// 在 [0,10) 内随机取数，命中比例即为 1:2:3:4。场景：广告投放、按权重负载均衡。
// 实例构造后内部状态不可变，可安全并发使用。
type WeightRandom[T any] struct {
	weights []float64 // 与 items 一一对应的权重（应为非负数）
	items   []T       // 待选取的元素
	total   float64   // 权重总和
}

// newWeightRandom 创建加权随机选取器（utils.NewWeightRandom 的实现）。
// weights 与 items 一一对应；两者长度不一致时以较短者为准；
// 权重总和 <= 0（全部为 0 或负数）时退化为等概率随机选取。
func newWeightRandom[T any](weights []float64, items []T) *WeightRandom[T] {
	n := len(weights)
	if len(items) < n {
		n = len(items)
	}
	w := &WeightRandom[T]{
		weights: make([]float64, n),
		items:   make([]T, n),
	}
	// 拷贝一份，与调用方的切片解除引用，保证实例不可变
	copy(w.weights, weights)
	copy(w.items, items)
	for _, weight := range w.weights {
		w.total += weight
	}
	return w
}

// Next 按权重随机返回一个元素。无可用元素（空构造）时返回元素零值。
func (w *WeightRandom[T]) Next() T {
	var zero T
	n := len(w.items)
	if n == 0 {
		return zero
	}
	// 权重总和 <= 0 时退化为等概率（对照 hutool：权重 <=0 的元素不会进入权重表，
	// 全部无效时 next() 返回 null，Go 返回零值）
	if w.total <= 0 {
		return w.items[randomIntn(n)]
	}
	r := mrand.Float64() * w.total
	var acc float64
	for i, weight := range w.weights {
		acc += weight
		if r < acc {
			return w.items[i]
		}
	}
	// 浮点误差兜底：r 极端接近 total 时命中最后一个元素
	return w.items[n-1]
}
