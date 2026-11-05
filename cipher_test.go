package utils

import (
	"bytes"
	"encoding/hex"
	"testing"
)

/* ---------------- AES/ECB ---------------- */

func TestCipherAesEcb(t *testing.T) {
	plain := []byte("hello 加密世界, aes ecb test!")
	// 16/24/32 字节密钥各做一次往返
	for _, keyLen := range []int{16, 24, 32} {
		key := make([]byte, keyLen)
		for i := range key {
			key[i] = byte(i)
		}
		enc, err := Crypto.AesEcbEncrypt(plain, key)
		if err != nil {
			t.Fatalf("AesEcbEncrypt(key=%d 字节) 失败: %v", keyLen, err)
		}
		if len(enc) == 0 || len(enc)%16 != 0 {
			t.Errorf("AES-%d 密文长度非法: %d", keyLen*8, len(enc))
		}
		dec, err := Crypto.AesEcbDecrypt(enc, key)
		if err != nil || !bytes.Equal(dec, plain) {
			t.Errorf("AES-%d ECB 往返失败: %q, %v", keyLen*8, dec, err)
		}
	}

	// ECB 性质：相同明文块 → 相同密文块
	key := []byte("0123456789abcdef")
	rep, err := Crypto.AesEcbEncrypt(bytes.Repeat([]byte("A"), 32), key)
	if err != nil {
		t.Fatalf("AesEcbEncrypt 失败: %v", err)
	}
	if !bytes.Equal(rep[:16], rep[16:32]) {
		t.Error("ECB 模式相同明文块应得到相同密文块")
	}

	// 空明文：加密为一个填充块，解密还原为空
	encEmpty, err := Crypto.AesEcbEncrypt(nil, key)
	if err != nil {
		t.Fatalf("AesEcbEncrypt(空明文) 失败: %v", err)
	}
	if len(encEmpty) != 16 {
		t.Errorf("空明文应加密为 1 个填充块, got %d 字节", len(encEmpty))
	}
	decEmpty, err := Crypto.AesEcbDecrypt(encEmpty, key)
	if err != nil || len(decEmpty) != 0 {
		t.Errorf("空明文往返失败: %q, %v", decEmpty, err)
	}

	// 错误路径：密文长度不是块大小的整数倍
	if _, err := Crypto.AesEcbDecrypt([]byte{1, 2, 3}, key); err == nil {
		t.Error("密文长度非块大小整数倍应返回错误")
	}
	// 错误路径：密文长度合法但 PKCS7 填充内容非法（16 字节全零密文）
	if _, err := Crypto.AesEcbDecrypt(make([]byte, 16), key); err == nil {
		t.Error("填充非法的密文应返回错误")
	}
	// 错误路径：key 长度非法（加解密两侧）
	if _, err := Crypto.AesEcbEncrypt(plain, []byte("short")); err == nil {
		t.Error("AesEcbEncrypt 非法 key 长度应返回错误")
	}
	if _, err := Crypto.AesEcbDecrypt(make([]byte, 16), []byte("short")); err == nil {
		t.Error("AesEcbDecrypt 非法 key 长度应返回错误")
	}
}

/* ---------------- AES/CBC ---------------- */

func TestCipherAesCbc(t *testing.T) {
	plain := []byte("hello 加密世界, aes cbc test!")
	key := []byte("0123456789abcdef")
	iv := []byte("fedcba9876543210")

	// 往返
	enc, err := Crypto.AesCbcEncrypt(plain, key, iv)
	if err != nil {
		t.Fatalf("AesCbcEncrypt 失败: %v", err)
	}
	dec, err := Crypto.AesCbcDecrypt(enc, key, iv)
	if err != nil || !bytes.Equal(dec, plain) {
		t.Errorf("AES CBC 往返失败: %q, %v", dec, err)
	}
	// 不同 iv 密文应不同
	enc2, _ := Crypto.AesCbcEncrypt(plain, key, []byte("1111111111111111"))
	if bytes.Equal(enc, enc2) {
		t.Error("不同 IV 的 CBC 密文不应相同")
	}

	// iv=nil 等价全零 IV
	encNil, err := Crypto.AesCbcEncrypt(plain, key, nil)
	if err != nil {
		t.Fatalf("AesCbcEncrypt(nil iv) 失败: %v", err)
	}
	encZero, _ := Crypto.AesCbcEncrypt(plain, key, make([]byte, 16))
	if !bytes.Equal(encNil, encZero) {
		t.Error("nil IV 应等价于全零 IV")
	}
	decNil, err := Crypto.AesCbcDecrypt(encNil, key, nil)
	if err != nil || !bytes.Equal(decNil, plain) {
		t.Errorf("nil IV 往返失败: %q, %v", decNil, err)
	}

	// 空明文往返（密文为一个填充块）
	encE, err := Crypto.AesCbcEncrypt(nil, key, iv)
	if err != nil {
		t.Fatalf("AesCbcEncrypt(空明文) 失败: %v", err)
	}
	if len(encE) != 16 {
		t.Errorf("空明文应加密为 1 个填充块, got %d 字节", len(encE))
	}
	decE, err := Crypto.AesCbcDecrypt(encE, key, iv)
	if err != nil || len(decE) != 0 {
		t.Errorf("空明文往返失败: %q, %v", decE, err)
	}

	// 错误路径：iv 长度错
	if _, err := Crypto.AesCbcEncrypt(plain, key, []byte("short")); err == nil {
		t.Error("AesCbcEncrypt IV 长度错应返回错误")
	}
	if _, err := Crypto.AesCbcDecrypt(enc, key, make([]byte, 8)); err == nil {
		t.Error("AesCbcDecrypt IV 长度错应返回错误")
	}
	// 错误路径：密文长度不是块大小整数倍
	if _, err := Crypto.AesCbcDecrypt([]byte{1, 2, 3}, key, iv); err == nil {
		t.Error("AesCbcDecrypt 密文长度非法应返回错误")
	}
	// 错误路径：key 长度非法
	if _, err := Crypto.AesCbcEncrypt(plain, []byte("badkey"), iv); err == nil {
		t.Error("AesCbcEncrypt 非法 key 长度应返回错误")
	}
	if _, err := Crypto.AesCbcDecrypt(enc, []byte("badkey"), iv); err == nil {
		t.Error("AesCbcDecrypt 非法 key 长度应返回错误")
	}
}

/* ---------------- DES ---------------- */

func TestCipherDes(t *testing.T) {
	plain := []byte("des 往返测试 data 123")
	key := []byte("8bytekey")
	enc, err := Crypto.DesEncrypt(plain, key)
	if err != nil {
		t.Fatalf("DesEncrypt 失败: %v", err)
	}
	if len(enc) == 0 || len(enc)%8 != 0 {
		t.Errorf("DES 密文长度非法: %d", len(enc))
	}
	dec, err := Crypto.DesDecrypt(enc, key)
	if err != nil || !bytes.Equal(dec, plain) {
		t.Errorf("DES 往返失败: %q, %v", dec, err)
	}
	// 错误路径：key 非 8 字节（加解密两侧）
	if _, err := Crypto.DesEncrypt(plain, []byte("7bytes!")); err == nil {
		t.Error("DesEncrypt key 非 8 字节应返回错误")
	}
	if _, err := Crypto.DesDecrypt(enc, []byte("9bytes!!")); err == nil {
		t.Error("DesDecrypt key 非 8 字节应返回错误")
	}
	// 错误路径：密文长度非 8 的整数倍
	if _, err := Crypto.DesDecrypt([]byte{1, 2, 3, 4, 5, 6, 7}, key); err == nil {
		t.Error("DesDecrypt 密文长度非法应返回错误")
	}
}

/* ---------------- 3DES ---------------- */

func TestCipherDes3(t *testing.T) {
	plain := []byte("3des 往返测试 data 456")
	key := []byte("123456789012345678901234")
	enc, err := Crypto.Des3Encrypt(plain, key)
	if err != nil {
		t.Fatalf("Des3Encrypt 失败: %v", err)
	}
	if len(enc) == 0 || len(enc)%8 != 0 {
		t.Errorf("3DES 密文长度非法: %d", len(enc))
	}
	dec, err := Crypto.Des3Decrypt(enc, key)
	if err != nil || !bytes.Equal(dec, plain) {
		t.Errorf("3DES 往返失败: %q, %v", dec, err)
	}
	// 错误路径：key 非 24 字节（加解密两侧）
	if _, err := Crypto.Des3Encrypt(plain, []byte("16bytes_key_xxxx")); err == nil {
		t.Error("Des3Encrypt key 非 24 字节应返回错误")
	}
	if _, err := Crypto.Des3Decrypt(enc, []byte("short")); err == nil {
		t.Error("Des3Decrypt key 非 24 字节应返回错误")
	}
	// 错误路径：密文长度非 8 的整数倍
	if _, err := Crypto.Des3Decrypt([]byte{1, 2, 3}, key); err == nil {
		t.Error("Des3Decrypt 密文长度非法应返回错误")
	}
}

/* ---------------- RC4 ---------------- */

func TestCipherRc4(t *testing.T) {
	// 经典测试向量（RFC 6229 前身、广为引用的测试组）
	if got := hex.EncodeToString(Crypto.Rc4Crypt([]byte("Plaintext"), []byte("Key"))); got != "bbf316e8d940af0ad3" {
		t.Errorf("RC4 向量不符: %s, want bbf316e8d940af0ad3", got)
	}
	// 对称性：加密两次还原原文
	data := []byte("rc4 对称流密码 test 123")
	once := Crypto.Rc4Crypt(data, []byte("secret"))
	if bytes.Equal(once, data) {
		t.Error("RC4 加密后不应等于原文")
	}
	if twice := Crypto.Rc4Crypt(once, []byte("secret")); !bytes.Equal(twice, data) {
		t.Errorf("RC4 加密两次应还原原文: %q", twice)
	}
	// 空 key 原样返回（源码放宽语义，对应 hutool 抛异常的场景）
	if noKey := Crypto.Rc4Crypt(data, nil); !bytes.Equal(noKey, data) {
		t.Errorf("空 key 应原样返回: %q", noKey)
	}
	// 空数据原样返回
	if got := Crypto.Rc4Crypt(nil, []byte("k")); len(got) != 0 {
		t.Errorf("空数据应原样返回: %v", got)
	}
}

/* ---------------- ChaCha20（RFC 8439） ---------------- */

func TestCipherChaCha20(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	nonce := []byte{0, 0, 0, 0, 0, 0, 0, 0x4a, 0, 0, 0, 0}
	plain := []byte("Ladies and Gentlemen of the class of '99: If I could offer you only one tip for the future, sunscreen would be it.")

	enc, err := Crypto.ChaCha20Xor(key, nonce, plain)
	if err != nil {
		t.Fatalf("ChaCha20Xor 失败: %v", err)
	}
	// RFC 8439 §2.4.2 完整密文向量
	const wantFull = "6e2e359a2568f98041ba0728dd0d6981" +
		"e97e7aec1d4360c20a27afccfd9fae0b" +
		"f91b65c5524733ab8f593dabcd62b357" +
		"1639d624e65152ab8f530c359f0861d8" +
		"07ca0dbf500d6a6156a38e088a22b65e" +
		"52bc514d16ccf806818ce91ab7793736" +
		"5af90bbf74a35be6b40b8eedf2785e42" +
		"874d"
	if got := hex.EncodeToString(enc[:16]); got != "6e2e359a2568f98041ba0728dd0d6981" {
		t.Errorf("ChaCha20 前 16 字节向量不符: %s", got)
	}
	if got := hex.EncodeToString(enc); got != wantFull {
		t.Errorf("ChaCha20 全量向量不符: %s", got)
	}
	// 对称性：再异或一次还原
	dec, err := Crypto.ChaCha20Xor(key, nonce, enc)
	if err != nil || !bytes.Equal(dec, plain) {
		t.Errorf("ChaCha20 对称性失败: %q, %v", dec, err)
	}
	// 错误路径：key 非 32 字节
	if _, err := Crypto.ChaCha20Xor(make([]byte, 31), nonce, plain); err == nil {
		t.Error("ChaCha20 key 非 32 字节应返回错误")
	}
	// 错误路径：nonce 非 12 字节
	if _, err := Crypto.ChaCha20Xor(key, make([]byte, 11), plain); err == nil {
		t.Error("ChaCha20 nonce 非 12 字节应返回错误")
	}
}

/* ---------------- XXTEA ---------------- */

func TestCipherXxtea(t *testing.T) {
	key := []byte("1234567890123456")
	// 往返：空串、1 字节、16 字节、超 16 字节多块
	cases := [][]byte{
		nil,
		[]byte("a"),
		[]byte("0123456789abcdef"),
		bytes.Repeat([]byte("x"), 100),
	}
	for i, plain := range cases {
		enc, err := Crypto.XxteaEncrypt(plain, key)
		if err != nil {
			t.Fatalf("XxteaEncrypt(用例 %d) 失败: %v", i, err)
		}
		if len(enc)%4 != 0 {
			t.Errorf("XxteaEncrypt(用例 %d) 密文长度 %d 应为 4 的整数倍", i, len(enc))
		}
		dec, err := Crypto.XxteaDecrypt(enc, key)
		if err != nil || !bytes.Equal(dec, plain) {
			t.Errorf("Xxtea 往返(用例 %d) 失败: %q, %v", i, dec, err)
		}
	}

	// fixKey 语义：短于 16 字节的 key 补零后与手工补齐的 16 字节 key 密文一致
	short := []byte("1234567890123") // 13 字节
	padded := append(append([]byte{}, short...), 0, 0, 0)
	data := []byte("fixkey 语义验证 data")
	eShort, err := Crypto.XxteaEncrypt(data, short)
	if err != nil {
		t.Fatalf("XxteaEncrypt(短 key) 失败: %v", err)
	}
	ePadded, err := Crypto.XxteaEncrypt(data, padded)
	if err != nil {
		t.Fatalf("XxteaEncrypt(补齐 key) 失败: %v", err)
	}
	if !bytes.Equal(eShort, ePadded) {
		t.Error("短 key 补零后应与 16 字节 key 密文一致（fixKey 语义）")
	}
	if dec, err := Crypto.XxteaDecrypt(eShort, short); err != nil || !bytes.Equal(dec, data) {
		t.Errorf("短 key 往返失败: %q, %v", dec, err)
	}

	// 错误路径：空 key（非空数据）
	if _, err := Crypto.XxteaEncrypt([]byte("data"), nil); err == nil {
		t.Error("XxteaEncrypt 空 key 应返回错误")
	}
	if _, err := Crypto.XxteaDecrypt([]byte("data"), nil); err == nil {
		t.Error("XxteaDecrypt 空 key 应返回错误")
	}
	// 错误路径：密文长度不是 4 的整数倍
	if _, err := Crypto.XxteaDecrypt([]byte("abc"), key); err == nil {
		t.Error("XxteaDecrypt 密文长度非 4 的整数倍应返回错误")
	}
	// 错误路径：密钥不匹配（密文尾部长度字段非法）
	ct, _ := Crypto.XxteaEncrypt([]byte("hello world"), key)
	if _, err := Crypto.XxteaDecrypt(ct, []byte("wrong-wrong-wrong")); err == nil {
		t.Error("XxteaDecrypt 密钥不匹配应返回错误")
	}
}

/* ---------------- 维吉尼亚密码 ---------------- */

func TestCipherVigenere(t *testing.T) {
	// hutool SymmetricTest#vigenereTest 黄金值（key=CompleteVictory）
	const vigPlain = "Wherethereisawillthereisaway"
	const vigKey = "CompleteVictory"
	const vigCipher = "zXScRZ]KIOMhQjc0\\bYRXZOJK[Vi" // 含字面反斜杠字符
	if got := Crypto.VigenereEncrypt(vigPlain, vigKey); got != vigCipher {
		t.Errorf("VigenereEncrypt = %q, want %q", got, vigCipher)
	}
	if got := Crypto.VigenereDecrypt(vigCipher, vigKey); got != vigPlain {
		t.Errorf("VigenereDecrypt = %q, want %q", got, vigPlain)
	}

	// 可打印 ASCII 往返
	plain := "Hello, World! 0123 ~!@#$%^&*()_+-={}[]|\\:;<>?,./"
	key := "secret"
	enc := Crypto.VigenereEncrypt(plain, key)
	if enc == plain {
		t.Error("加密后不应等于原文")
	}
	if back := Crypto.VigenereDecrypt(enc, key); back != plain {
		t.Errorf("Vigenere 往返失败: %q", back)
	}

	// 空 key / 空数据按源码语义原样返回
	if got := Crypto.VigenereEncrypt("abc", ""); got != "abc" {
		t.Errorf("空 key 加密应原样返回: %q", got)
	}
	if got := Crypto.VigenereDecrypt("abc", ""); got != "abc" {
		t.Errorf("空 key 解密应原样返回: %q", got)
	}
	if got := Crypto.VigenereEncrypt("", "key"); got != "" {
		t.Errorf("空数据加密应原样返回: %q", got)
	}
	if got := Crypto.VigenereDecrypt("", "key"); got != "" {
		t.Errorf("空数据解密应原样返回: %q", got)
	}
}
