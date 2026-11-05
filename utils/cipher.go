package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"encoding/binary"
	"fmt"
	"math/bits"
)

// ------------------ 内部辅助：PKCS7 填充与 IV 归一化 ------------------

// cipherPkcs7Pad 按 PKCS7 规则填充（hutool 的 PKCS5Padding 在 8/16 字节分组下与 PKCS7 等价）。
func cipherPkcs7Pad(data []byte, blockSize int) []byte {
	padLen := blockSize - len(data)%blockSize
	out := make([]byte, len(data)+padLen)
	copy(out, data)
	for i := len(data); i < len(out); i++ {
		out[i] = byte(padLen)
	}
	return out
}

// cipherPkcs7Unpad 去除 PKCS7 填充，密文长度非法或填充内容非法时返回错误。
func cipherPkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	n := len(data)
	if n == 0 || n%blockSize != 0 {
		return nil, fmt.Errorf("密文长度 %d 不是分组大小 %d 的整数倍", n, blockSize)
	}
	padLen := int(data[n-1])
	if padLen <= 0 || padLen > blockSize || padLen > n {
		return nil, fmt.Errorf("非法的 PKCS7 填充长度 %d", padLen)
	}
	for _, b := range data[n-padLen:] {
		if int(b) != padLen {
			return nil, fmt.Errorf("非法的 PKCS7 填充内容")
		}
	}
	return data[:n-padLen], nil
}

// cipherNormalizeIv 归一化 IV：nil/空按全 0 处理（与 hutool 传 null iv 的 JDK 默认行为对应），
// 长度不符时返回错误。
func cipherNormalizeIv(iv []byte, blockSize int) ([]byte, error) {
	if len(iv) == 0 {
		return make([]byte, blockSize), nil
	}
	if len(iv) != blockSize {
		return nil, fmt.Errorf("IV 长度须为 %d 字节，当前为 %d", blockSize, len(iv))
	}
	return iv, nil
}

// ------------------ AES ------------------

// AesEcbEncrypt AES/ECB/PKCS7 加密（对应 hutool AES 默认的 ECB/PKCS5Padding）。
// key 须为 16/24/32 字节；ECB 模式下相同明文块得到相同密文块，敏感数据建议改用 AesCbcEncrypt。
// 场景：与 hutool SecureUtil.aes(key).encrypt(...) 的旧 Java 服务互操作。
func (c *utilCrypto) AesEcbEncrypt(data, key []byte) ([]byte, error) {
	switch len(key) {
	case 16, 24, 32:
	default:
		return nil, fmt.Errorf("AES 密钥长度须为 16/24/32 字节，当前为 %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	padded := cipherPkcs7Pad(data, aes.BlockSize)
	out := make([]byte, len(padded))
	for off := 0; off < len(padded); off += aes.BlockSize {
		block.Encrypt(out[off:off+aes.BlockSize], padded[off:off+aes.BlockSize])
	}
	return out, nil
}

// AesEcbDecrypt AES/ECB/PKCS7 解密。密文长度非 16 的整数倍或填充非法时返回错误。
func (c *utilCrypto) AesEcbDecrypt(data, key []byte) ([]byte, error) {
	switch len(key) {
	case 16, 24, 32:
	default:
		return nil, fmt.Errorf("AES 密钥长度须为 16/24/32 字节，当前为 %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 || len(data)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("密文长度 %d 不是 %d 的整数倍", len(data), aes.BlockSize)
	}
	out := make([]byte, len(data))
	for off := 0; off < len(data); off += aes.BlockSize {
		block.Decrypt(out[off:off+aes.BlockSize], data[off:off+aes.BlockSize])
	}
	return cipherPkcs7Unpad(out, aes.BlockSize)
}

// AesCbcEncrypt AES/CBC/PKCS7 加密。key 须为 16/24/32 字节；iv 长度须为 16，
// 传 nil 时按全 0 IV 处理（对应 hutool 使用 null iv 时 JDK Cipher 的默认行为）。
// 场景：跨语言报文加密互通（CBC 为最常用的分组工作模式）。
func (c *utilCrypto) AesCbcEncrypt(data, key, iv []byte) ([]byte, error) {
	switch len(key) {
	case 16, 24, 32:
	default:
		return nil, fmt.Errorf("AES 密钥长度须为 16/24/32 字节，当前为 %d", len(key))
	}
	iv, err := cipherNormalizeIv(iv, aes.BlockSize)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	padded := cipherPkcs7Pad(data, aes.BlockSize)
	out := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(out, padded)
	return out, nil
}

// AesCbcDecrypt AES/CBC/PKCS7 解密。iv 规则同 AesCbcEncrypt；
// 密文长度非 16 的整数倍或填充非法时返回错误。
func (c *utilCrypto) AesCbcDecrypt(data, key, iv []byte) ([]byte, error) {
	switch len(key) {
	case 16, 24, 32:
	default:
		return nil, fmt.Errorf("AES 密钥长度须为 16/24/32 字节，当前为 %d", len(key))
	}
	iv, err := cipherNormalizeIv(iv, aes.BlockSize)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 || len(data)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("密文长度 %d 不是 %d 的整数倍", len(data), aes.BlockSize)
	}
	out := make([]byte, len(data))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(out, data)
	return cipherPkcs7Unpad(out, aes.BlockSize)
}

// ------------------ DES / 3DES ------------------

// DesEncrypt 单重 DES/ECB/PKCS7 加密（hutool DES 默认），key 须为 8 字节。
// 注意 DES 强度早已不足，仅用于兼容旧系统，新系统请使用 AES。
// 场景：对接银行/老设备的 8 字节密钥旧协议。
func (c *utilCrypto) DesEncrypt(data, key []byte) ([]byte, error) {
	if len(key) != 8 {
		return nil, fmt.Errorf("DES 密钥长度须为 8 字节，当前为 %d", len(key))
	}
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	padded := cipherPkcs7Pad(data, des.BlockSize)
	out := make([]byte, len(padded))
	for off := 0; off < len(padded); off += des.BlockSize {
		block.Encrypt(out[off:off+des.BlockSize], padded[off:off+des.BlockSize])
	}
	return out, nil
}

// DesDecrypt 单重 DES/ECB/PKCS7 解密。密文长度非 8 的整数倍或填充非法时返回错误。
func (c *utilCrypto) DesDecrypt(data, key []byte) ([]byte, error) {
	if len(key) != 8 {
		return nil, fmt.Errorf("DES 密钥长度须为 8 字节，当前为 %d", len(key))
	}
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 || len(data)%des.BlockSize != 0 {
		return nil, fmt.Errorf("密文长度 %d 不是 %d 的整数倍", len(data), des.BlockSize)
	}
	out := make([]byte, len(data))
	for off := 0; off < len(data); off += des.BlockSize {
		block.Decrypt(out[off:off+des.BlockSize], data[off:off+des.BlockSize])
	}
	return cipherPkcs7Unpad(out, des.BlockSize)
}

// Des3Encrypt 3DES/ECB/PKCS7 加密（hutool DESede 默认），key 须为 24 字节。
// 场景：金融行业旧协议（IC 卡、POS 报文）的兼容加解密。
func (c *utilCrypto) Des3Encrypt(data, key []byte) ([]byte, error) {
	if len(key) != 24 {
		return nil, fmt.Errorf("3DES 密钥长度须为 24 字节，当前为 %d", len(key))
	}
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}
	padded := cipherPkcs7Pad(data, des.BlockSize)
	out := make([]byte, len(padded))
	for off := 0; off < len(padded); off += des.BlockSize {
		block.Encrypt(out[off:off+des.BlockSize], padded[off:off+des.BlockSize])
	}
	return out, nil
}

// Des3Decrypt 3DES/ECB/PKCS7 解密。密文长度非 8 的整数倍或填充非法时返回错误。
func (c *utilCrypto) Des3Decrypt(data, key []byte) ([]byte, error) {
	if len(key) != 24 {
		return nil, fmt.Errorf("3DES 密钥长度须为 24 字节，当前为 %d", len(key))
	}
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 || len(data)%des.BlockSize != 0 {
		return nil, fmt.Errorf("密文长度 %d 不是 %d 的整数倍", len(data), des.BlockSize)
	}
	out := make([]byte, len(data))
	for off := 0; off < len(data); off += des.BlockSize {
		block.Decrypt(out[off:off+des.BlockSize], data[off:off+des.BlockSize])
	}
	return cipherPkcs7Unpad(out, des.BlockSize)
}

// ------------------ RC4（标准库已移除，手写实现） ------------------

// Rc4Crypt RC4 流密码加密/解密（加解密同为异或运算，本方法两者通用）。
// 标准库已移除 RC4，这里按算法手写：KSA 初始化 S 盒 + PRGA 逐字节异或，对照 hutool RC4.crypt。
// key 建议 1~256 字节；空 key 时原样返回 data（hutool 此时抛异常，这里放宽处理）。
// 注意：RC4 已不安全，仅用于兼容旧协议。
// 场景：兼容老系统/老设备的私有协议报文加解密。
func (c *utilCrypto) Rc4Crypt(data, key []byte) []byte {
	if len(key) == 0 || len(data) == 0 {
		return data
	}
	var s [256]byte
	for i := 0; i < 256; i++ {
		s[i] = byte(i)
	}
	j := 0
	for i := 0; i < 256; i++ {
		j = (j + int(s[i]) + int(key[i%len(key)])) % 256
		s[i], s[j] = s[j], s[i]
	}
	out := make([]byte, len(data))
	i, k := 0, 0
	for n := 0; n < len(data); n++ {
		i = (i + 1) % 256
		k = (k + int(s[i])) % 256
		s[i], s[k] = s[k], s[i]
		out[n] = data[n] ^ s[(int(s[i])+int(s[k]))%256]
	}
	return out
}

// ------------------ ChaCha20（RFC 8439，手写实现） ------------------

// cipherQuarterRound ChaCha20 的四分之轮运算（RFC 8439 §2.1.1）。
func cipherQuarterRound(a, b, c, d uint32) (uint32, uint32, uint32, uint32) {
	a += b
	d ^= a
	d = bits.RotateLeft32(d, 16)
	c += d
	b ^= c
	b = bits.RotateLeft32(b, 12)
	a += b
	d ^= a
	d = bits.RotateLeft32(d, 8)
	c += d
	b ^= c
	b = bits.RotateLeft32(b, 7)
	return a, b, c, d
}

// cipherChaCha20Block 生成 64 字节 ChaCha20 密钥流块
// （RFC 8439 §2.3：初始状态经 20 轮混淆后与原状态逐字相加，小端输出）。
func cipherChaCha20Block(key *[32]byte, counter uint32, nonce *[12]byte) [64]byte {
	var state [16]uint32
	state[0], state[1], state[2], state[3] = 0x61707865, 0x3320646e, 0x79622d32, 0x6b206574 // "expand 32-byte k"
	for i := 0; i < 8; i++ {
		state[4+i] = binary.LittleEndian.Uint32(key[i*4:])
	}
	state[12] = counter
	for i := 0; i < 3; i++ {
		state[13+i] = binary.LittleEndian.Uint32(nonce[i*4:])
	}
	working := state
	for i := 0; i < 10; i++ { // 10 个双轮，共 20 轮
		// 列轮
		working[0], working[4], working[8], working[12] = cipherQuarterRound(working[0], working[4], working[8], working[12])
		working[1], working[5], working[9], working[13] = cipherQuarterRound(working[1], working[5], working[9], working[13])
		working[2], working[6], working[10], working[14] = cipherQuarterRound(working[2], working[6], working[10], working[14])
		working[3], working[7], working[11], working[15] = cipherQuarterRound(working[3], working[7], working[11], working[15])
		// 对角轮
		working[0], working[5], working[10], working[15] = cipherQuarterRound(working[0], working[5], working[10], working[15])
		working[1], working[6], working[11], working[12] = cipherQuarterRound(working[1], working[6], working[11], working[12])
		working[2], working[7], working[8], working[13] = cipherQuarterRound(working[2], working[7], working[8], working[13])
		working[3], working[4], working[9], working[14] = cipherQuarterRound(working[3], working[4], working[9], working[14])
	}
	var out [64]byte
	for i := 0; i < 16; i++ {
		binary.LittleEndian.PutUint32(out[i*4:], working[i]+state[i])
	}
	return out
}

// ChaCha20Xor ChaCha20 流加密/解密（RFC 8439，手写实现：quarter round + 20 轮 + 密钥流异或，
// 不含 Poly1305；初始计数器取 1，与 RFC 8439 §2.4 加密测试向量一致，
// 实现已用 RFC 8439 §2.3.2 密钥流向量与 §2.4.2 加密向量自校验）。
// key 须为 32 字节，nonce 须为 12 字节，长度非法返回错误；加解密为同一异或运算，两者通用。
// 场景：高性能流加密的协议兼容实现（TLS 1.3 / WireGuard 同族算法）。
func (c *utilCrypto) ChaCha20Xor(key, nonce, data []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("ChaCha20 密钥长度须为 32 字节，当前为 %d", len(key))
	}
	if len(nonce) != 12 {
		return nil, fmt.Errorf("ChaCha20 nonce 长度须为 12 字节，当前为 %d", len(nonce))
	}
	var keyArr [32]byte
	copy(keyArr[:], key)
	var nonceArr [12]byte
	copy(nonceArr[:], nonce)
	out := make([]byte, len(data))
	var stream [64]byte
	for pos, counter := 0, uint32(1); pos < len(data); pos, counter = pos+64, counter+1 {
		stream = cipherChaCha20Block(&keyArr, counter, &nonceArr)
		for i := 0; i < 64 && pos+i < len(data); i++ {
			out[pos+i] = data[pos+i] ^ stream[i]
		}
	}
	return out, nil
}

// ------------------ XXTEA（忠实移植 hutool XXTEA.java） ------------------

// cipherXxteaDelta XXTEA 的黄金比例常量。
const cipherXxteaDelta uint32 = 0x9E3779B9

// cipherXxteaFixKey 按 hutool 规则把密钥截断/补零到 16 字节。
func cipherXxteaFixKey(key []byte) []byte {
	if len(key) == 16 {
		return key
	}
	fixed := make([]byte, 16)
	copy(fixed, key)
	return fixed
}

// cipherXxteaToIntArray 字节流转 uint32 小端数组；includeLength 为 true 时末尾附加原始字节长度
// （hutool toIntArray 同名逻辑，是密文头 4 字节长度字段的来源）。
func cipherXxteaToIntArray(data []byte, includeLength bool) []uint32 {
	n := len(data) >> 2
	if len(data)&3 != 0 {
		n++
	}
	var result []uint32
	if includeLength {
		result = make([]uint32, n+1)
		result[n] = uint32(len(data))
	} else {
		result = make([]uint32, n)
	}
	for i := 0; i < len(data); i++ {
		result[i>>2] |= uint32(data[i]) << uint((i&3)<<3)
	}
	return result
}

// cipherXxteaToByteArray uint32 数组转字节流；includeLength 为 true 时按末尾长度字段截取，
// 长度字段非法（密文损坏或密钥不匹配）时返回 ok=false。
func cipherXxteaToByteArray(data []uint32, includeLength bool) ([]byte, bool) {
	n := len(data) << 2
	if includeLength {
		m := int(data[len(data)-1])
		n -= 4
		if m < n-3 || m > n {
			return nil, false
		}
		n = m
	}
	result := make([]byte, n)
	for i := 0; i < n; i++ {
		result[i] = byte(data[i>>2] >> uint((i&3)<<3))
	}
	return result, true
}

// cipherXxteaMx XXTEA 非线性函数（对照 hutool mx；Go 与 Java 均按 32 位回绕运算，逻辑右移一致）。
func cipherXxteaMx(sum, y, z, p, e uint32, k []uint32) uint32 {
	return ((z>>5 ^ y<<2) + (y>>3 ^ z<<4)) ^ ((sum ^ y) + (k[(p&3)^e] ^ z))
}

// cipherXxteaEncryptArray XXTEA 加密核心，轮数 q = 6 + 52/(n+1)（hutool encrypt 同名逻辑）。
func cipherXxteaEncryptArray(v, k []uint32) []uint32 {
	n := len(v) - 1
	if n < 1 {
		return v
	}
	var p int
	var sum, e, y, z uint32
	z = v[n]
	for q := 6 + 52/(n+1); q > 0; q-- {
		sum += cipherXxteaDelta
		e = (sum >> 2) & 3
		for p = 0; p < n; p++ {
			y = v[p+1]
			z = v[p] + cipherXxteaMx(sum, y, z, uint32(p), e, k)
			v[p] = z
		}
		y = v[0]
		z = v[n] + cipherXxteaMx(sum, y, z, uint32(p), e, k)
		v[n] = z
	}
	return v
}

// cipherXxteaDecryptArray XXTEA 解密核心，sum 从 q*DELTA 回绕递减至 0（hutool decrypt 同名逻辑）。
func cipherXxteaDecryptArray(v, k []uint32) []uint32 {
	n := len(v) - 1
	if n < 1 {
		return v
	}
	q := 6 + 52/(n+1)
	sum := uint32(q) * cipherXxteaDelta
	var p int
	var e, y, z uint32
	y = v[0]
	for sum != 0 {
		e = (sum >> 2) & 3
		for p = n; p > 0; p-- {
			z = v[p-1]
			y = v[p] - cipherXxteaMx(sum, y, z, uint32(p), e, k)
			v[p] = y
		}
		z = v[n]
		y = v[0] - cipherXxteaMx(sum, y, z, uint32(p), e, k)
		v[0] = y
		sum -= cipherXxteaDelta
	}
	return v
}

// XxteaEncrypt XXTEA 加密（忠实移植 hutool XXTEA.java，保持密文格式兼容：
// 明文长度以 4 字节小端写入密文尾部字段，密文总长为 4 的整数倍；空数据原样返回）。
// key 按密文兼容规则截断/补零到 16 字节（hutool fixKey，建议直接传 16 字节）。
// 场景：嵌入式设备/配置混淆等轻量加密，与 hutool XXTEA 结果逐字节互通。
func (c *utilCrypto) XxteaEncrypt(data, key []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}
	if len(key) == 0 {
		return nil, fmt.Errorf("XXTEA 密钥不能为空")
	}
	v := cipherXxteaEncryptArray(
		cipherXxteaToIntArray(data, true),
		cipherXxteaToIntArray(cipherXxteaFixKey(key), false))
	out, _ := cipherXxteaToByteArray(v, false)
	return out, nil
}

// XxteaDecrypt XXTEA 解密。密文须为 4 的整数倍且尾部含合法明文长度字段，
// 否则（密文损坏或密钥不匹配）返回错误。
func (c *utilCrypto) XxteaDecrypt(data, key []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}
	if len(key) == 0 {
		return nil, fmt.Errorf("XXTEA 密钥不能为空")
	}
	if len(data)%4 != 0 {
		return nil, fmt.Errorf("非法 XXTEA 密文：长度 %d 不是 4 的整数倍", len(data))
	}
	v := cipherXxteaDecryptArray(
		cipherXxteaToIntArray(data, false),
		cipherXxteaToIntArray(cipherXxteaFixKey(key), false))
	out, ok := cipherXxteaToByteArray(v, true)
	if !ok {
		return nil, fmt.Errorf("非法 XXTEA 密文或密钥不匹配")
	}
	return out, nil
}

// ------------------ 维吉尼亚密码（对照 hutool Vigenere.java） ------------------

// VigenereEncrypt 维吉尼亚密码加密（对照 hutool Vigenere.java）。
// 注意：hutool 实际是对所有字符按 95 个可打印 ASCII 字符（0x20~0x7E）循环位移，
// 并非仅对字母位移，本实现与 hutool 逐字符一致以便互通；密钥为空时原样返回 data
// （hutool 此时抛出异常）。按 Unicode 码点处理，BMP 字符与 Java char 语义一致。
// 场景：教学演示、CTF、与旧系统的简易字符混淆互通。
func (c *utilCrypto) VigenereEncrypt(data, key string) string {
	if key == "" || data == "" {
		return data
	}
	runes := []rune(data)
	keyRunes := []rune(key)
	for i := range runes {
		runes[i] = rune((int(runes[i])+int(keyRunes[i%len(keyRunes)])-64)%95 + 32)
	}
	return string(runes)
}

// VigenereDecrypt 维吉尼亚密码解密，与 VigenereEncrypt 配对（对照 hutool Vigenere.decrypt）。
func (c *utilCrypto) VigenereDecrypt(data, key string) string {
	if key == "" || data == "" {
		return data
	}
	runes := []rune(data)
	keyRunes := []rune(key)
	for i := range runes {
		d := int(runes[i]) - int(keyRunes[i%len(keyRunes)])
		if d < 0 {
			d += 95
		}
		runes[i] = rune(d%95 + 32)
	}
	return string(runes)
}
