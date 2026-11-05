package main

import (
	"fmt"
	"time"

	"github.com/hicode0101/go-utils/utils"
)

// cryptoDemo 加解密与散列：Crypto（摘要/编码/HMAC/OTP/PBKDF2/对称加密/RSA/签名）/ Hex / Hash。
func cryptoDemo() {
	banner("Crypto 摘要与编码")
	fmt.Println(`MD5("abc") =`, utils.Crypto.MD5("abc"))
	fmt.Println(`SHA256("abc") =`, utils.Crypto.SHA256("abc"))
	fmt.Println(`MD5WithSalt("password", "salt") =`, utils.Crypto.MD5WithSalt("password", "salt"))
	if dec, err := utils.Crypto.Base64Decode(utils.Crypto.Base64Encode("hello")); err == nil {
		fmt.Println(`Base64Encode("hello") =`, utils.Crypto.Base64Encode("hello"), "| Base64Decode =", dec)
	}
	fmt.Println(`Base64UrlEncode("a+b/c=") =`, utils.Crypto.Base64UrlEncode("a+b/c="))
	fmt.Println(`HmacSha256("key", "data") =`, utils.Crypto.HmacSha256("key", "data"))
	fmt.Println("Uuid() =", utils.Crypto.Uuid())

	// 待签名参数：按 key 字典序排序拼接后再做摘要（对接支付/开放平台签名）
	params := map[string]string{"appId": "demo", "timestamp": "1700000000", "nonce": "abc123"}
	fmt.Println("SignParamsMd5(params) =", utils.Crypto.SignParamsMd5(params))

	banner("Crypto 对称加解密")
	key := []byte("0123456789abcdef")
	iv := []byte("fedcba9876543210")
	enc, _ := utils.Crypto.AesCbcEncrypt([]byte("hello go-utils"), key, iv)
	dec, _ := utils.Crypto.AesCbcDecrypt(enc, key, iv)
	fmt.Printf("AesCbcEncrypt/AesCbcDecrypt 往返: %q -> %x... -> %q\n", "hello go-utils", enc[:4], string(dec))
	ecbEnc, _ := utils.Crypto.AesEcbEncrypt([]byte("hello"), key)
	ecbDec, _ := utils.Crypto.AesEcbDecrypt(ecbEnc, key)
	fmt.Printf("AesEcbEncrypt/AesEcbDecrypt 往返: %q -> %q\n", "hello", string(ecbDec))
	desKey := []byte("12345678")
	desEnc, _ := utils.Crypto.DesEncrypt([]byte("des test"), desKey)
	desDec, _ := utils.Crypto.DesDecrypt(desEnc, desKey)
	fmt.Printf("DesEncrypt/DesDecrypt 往返: %q -> %q\n", "des test", string(desDec))
	rc4Enc := utils.Crypto.Rc4Crypt([]byte("rc4 test"), []byte("secret"))
	rc4Dec := utils.Crypto.Rc4Crypt(rc4Enc, []byte("secret")) // 异或对称：再调一次即解密
	fmt.Printf("Rc4Crypt 往返: %q -> %q\n", "rc4 test", string(rc4Dec))
	nonce := []byte("0123456789ab")                     // nonce 须 12 字节
	ccKey := []byte("0123456789abcdef0123456789abcdef") // ChaCha20 密钥须 32 字节
	ccEnc, _ := utils.Crypto.ChaCha20Xor(ccKey, nonce, []byte("chacha20"))
	ccDec, _ := utils.Crypto.ChaCha20Xor(ccKey, nonce, ccEnc)
	fmt.Printf("ChaCha20Xor 往返: %q -> %q\n", "chacha20", string(ccDec))
	xxEnc, _ := utils.Crypto.XxteaEncrypt([]byte("xxtea test"), key)
	xxDec, _ := utils.Crypto.XxteaDecrypt(xxEnc, key)
	fmt.Printf("XxteaEncrypt/XxteaDecrypt 往返: %q -> %q\n", "xxtea test", string(xxDec))
	fmt.Println(`VigenereEncrypt("HELLO", "KEY") =`, utils.Crypto.VigenereEncrypt("HELLO", "KEY"),
		"| VigenereDecrypt =", utils.Crypto.VigenereDecrypt(utils.Crypto.VigenereEncrypt("HELLO", "KEY"), "KEY"))

	banner("Crypto 动态口令（OTP）与 PBKDF2")
	secret := "JBSWY3DPEHPK3PXP" // Base32 密钥
	code, _ := utils.Crypto.Hotp(secret, 1, 6)
	fmt.Println("Hotp(secret, 1, 6) =", code)
	totp, _ := utils.Crypto.Totp(secret, time.Now(), 6)
	ok, _ := utils.Crypto.TotpValidate(secret, totp, 1)
	fmt.Println("Totp(secret, now, 6) =", totp, "| TotpValidate 验证自身 =", ok)
	gSec, gUri, _ := utils.Crypto.GoogleSecretKey("user@example.com", 16)
	fmt.Println("GoogleSecretKey secret =", gSec)
	fmt.Println("GoogleSecretKey uri =", gUri)
	fmt.Println("Pbkdf2Hex(password, salt) =", utils.Crypto.Pbkdf2Hex([]byte("password"), []byte("salt"))[:32]+"...")

	banner("Crypto RSA 与数字签名")
	privB64, pubB64, _ := utils.Crypto.GenerateRsaKeyPair(1024)
	msg := []byte("rsa 加密测试")
	rEnc, _ := utils.Crypto.RsaEncrypt(msg, pubB64)
	rDec, _ := utils.Crypto.RsaDecrypt(rEnc, privB64)
	fmt.Printf("RsaEncrypt/RsaDecrypt 往返: %q -> %q\n", string(msg), string(rDec))
	sig, _ := utils.Crypto.RsaSignSha256(msg, privB64)
	verified, _ := utils.Crypto.RsaVerifySha256(msg, sig, pubB64)
	fmt.Println("RsaSignSha256/RsaVerifySha256 验签通过 =", verified)
	edPriv, edPub, _ := utils.Crypto.GenerateEd25519KeyPair()
	edSig, _ := utils.Crypto.Ed25519Sign(msg, edPriv)
	edOk, _ := utils.Crypto.Ed25519Verify(msg, edSig, edPub)
	fmt.Println("Ed25519Sign/Ed25519Verify 验签通过 =", edOk)

	banner("Hex 十六进制")
	fmt.Println(`Encode([]byte("go")) =`, utils.Hex.Encode([]byte("go")),
		`| EncodeStr("go") =`, utils.Hex.EncodeStr("go"))
	if s, err := utils.Hex.DecodeStr("676f"); err == nil {
		fmt.Println(`DecodeStr("676f") =`, s)
	}
	hexNum, _ := utils.Hex.HexToInt("ff")
	fmt.Println("ToHex(255) =", utils.Hex.ToHex(255), `| HexToInt("ff") =`, hexNum)
	fmt.Println("EncodeColor(255, 136, 0) =", utils.Hex.EncodeColor(255, 136, 0))
	fmt.Println(`Format("676f8f") =`, utils.Hex.Format("676f8f"))

	banner("Hash 字符串散列")
	fmt.Println(`FnvHash("hello") =`, utils.Hash.FnvHash("hello"),
		"| ElfHash =", utils.Hash.ElfHash("hello"),
		"| BkdrHash =", utils.Hash.BkdrHash("hello"))
	fmt.Println(`JavaDefaultHash("hello") =`, utils.Hash.JavaDefaultHash("hello"), "（与 Java str.hashCode() 一致）")
	fmt.Println(`Murmur32("hello") =`, utils.Hash.Murmur32("hello"), "| Murmur128 =", utils.Hash.Murmur128("hello"))
}
