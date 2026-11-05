package utils

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"testing"
)

/* ---------------- RSA 密钥生成与加解密 ---------------- */

func TestRsaKeyPairAndRoundtrip(t *testing.T) {
	// 2048 位密钥对生成
	privB64, pubB64, err := Crypto.GenerateRsaKeyPair(2048)
	if err != nil {
		t.Fatalf("GenerateRsaKeyPair(2048) 失败: %v", err)
	}
	if privB64 == "" || pubB64 == "" {
		t.Fatal("生成的密钥对不应为空")
	}

	// 中文 + 英文长文加解密往返（长度不超过模长-11 字节）
	plain := []byte("go-utils RSA 往返测试：The quick brown fox jumps over the lazy dog. " +
		"敏感报文加密场景验证，包含中英混排与数字 0123456789。")
	enc, err := Crypto.RsaEncrypt(plain, pubB64)
	if err != nil {
		t.Fatalf("RsaEncrypt 失败: %v", err)
	}
	dec, err := Crypto.RsaDecrypt(enc, privB64)
	if err != nil || !bytes.Equal(dec, plain) {
		t.Errorf("RSA 往返失败: %q, %v", dec, err)
	}

	// keySize<=0 默认 1024 位，应可用不 panic
	priv1024, pub1024, err := Crypto.GenerateRsaKeyPair(0)
	if err != nil {
		t.Fatalf("GenerateRsaKeyPair(0) 失败: %v", err)
	}
	if priv1024 == "" || pub1024 == "" {
		t.Fatal("默认密钥对不应为空")
	}
	small := []byte("1024")
	enc2, err := Crypto.RsaEncrypt(small, pub1024)
	if err != nil {
		t.Fatalf("RsaEncrypt(默认密钥) 失败: %v", err)
	}
	dec2, err := Crypto.RsaDecrypt(enc2, priv1024)
	if err != nil || !bytes.Equal(dec2, small) {
		t.Errorf("默认 1024 位密钥往返失败: %q, %v", dec2, err)
	}
}

/* ---------------- RSA 加解密错误路径 ---------------- */

func TestRsaEncryptDecryptErrors(t *testing.T) {
	_, pubB64, err := Crypto.GenerateRsaKeyPair(2048)
	if err != nil {
		t.Fatalf("GenerateRsaKeyPair 失败: %v", err)
	}
	privX, _, err := Crypto.GenerateRsaKeyPair(2048) // 另一对密钥
	if err != nil {
		t.Fatalf("GenerateRsaKeyPair(第二对) 失败: %v", err)
	}

	enc, err := Crypto.RsaEncrypt([]byte("x"), pubB64)
	if err != nil {
		t.Fatalf("RsaEncrypt 失败: %v", err)
	}
	// 错误路径：用错误私钥解密
	if _, err := Crypto.RsaDecrypt(enc, privX); err == nil {
		t.Error("用不匹配的私钥解密应返回错误")
	}
	// 错误路径：非法 Base64 密钥
	if _, err := Crypto.RsaEncrypt([]byte("x"), "not-base64!!!"); err == nil {
		t.Error("RsaEncrypt 非法 Base64 公钥应返回错误")
	}
	if _, err := Crypto.RsaDecrypt(enc, "not-base64!!!"); err == nil {
		t.Error("RsaDecrypt 非法 Base64 私钥应返回错误")
	}
	// 错误路径：Base64 合法但内容不是密钥 DER
	garbage := base64.StdEncoding.EncodeToString([]byte("garbage key bytes"))
	if _, err := Crypto.RsaEncrypt([]byte("x"), garbage); err == nil {
		t.Error("RsaEncrypt 非密钥 DER 应返回错误")
	}
	if _, err := Crypto.RsaDecrypt(enc, garbage); err == nil {
		t.Error("RsaDecrypt 非密钥 DER 应返回错误")
	}
}

/* ---------------- RSA 签名与验签（SHA1 / SHA256） ---------------- */

func TestRsaSignVerifySha1Sha256(t *testing.T) {
	privB64, pubB64, err := Crypto.GenerateRsaKeyPair(2048)
	if err != nil {
		t.Fatalf("GenerateRsaKeyPair 失败: %v", err)
	}
	data := []byte("签名数据 sign me 2026")

	for _, tc := range []struct {
		name   string
		sign   func(data []byte, privateKeyB64 string) (string, error)
		verify func(data []byte, signHex, publicKeyB64 string) (bool, error)
	}{
		{"Sha1", Crypto.RsaSignSha1, Crypto.RsaVerifySha1},
		{"Sha256", Crypto.RsaSignSha256, Crypto.RsaVerifySha256},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// 签验通过
			sig, err := tc.sign(data, privB64)
			if err != nil {
				t.Fatalf("RsaSign%s 失败: %v", tc.name, err)
			}
			if sig == "" {
				t.Fatal("签名不应为空")
			}
			if ok, err := tc.verify(data, sig, pubB64); err != nil || !ok {
				t.Errorf("RsaVerify%s 应通过: %v, %v", tc.name, ok, err)
			}
			// PKCS1v15 签名确定性：同数据同密钥签名一致
			sig2, _ := tc.sign(data, privB64)
			if sig2 != sig {
				t.Errorf("PKCS1v15 签名应确定: %q vs %q", sig, sig2)
			}
			// 篡改数据不通过
			if ok, err := tc.verify([]byte("篡改后的数据"), sig, pubB64); err != nil || ok {
				t.Errorf("被篡改数据验签不应通过: %v, %v", ok, err)
			}
			// 签名内容错误 → false 而非 error
			if ok, err := tc.verify(data, "00", pubB64); err != nil || ok {
				t.Errorf("错误签名应返回 false 而非 error: %v, %v", ok, err)
			}
			// 错误路径：非法私钥/公钥/十六进制签名
			if _, err := tc.sign(data, "bad-key!!!"); err == nil {
				t.Errorf("RsaSign%s 非法 Base64 私钥应返回错误", tc.name)
			}
			if _, err := tc.verify(data, sig, "bad-pub!!!"); err == nil {
				t.Errorf("RsaVerify%s 非法 Base64 公钥应返回错误", tc.name)
			}
			if _, err := tc.verify(data, "zz-非hex", pubB64); err == nil {
				t.Errorf("RsaVerify%s 非法 hex 签名应返回错误", tc.name)
			}
		})
	}
}

/* ---------------- Ed25519 签名与验签 ---------------- */

func TestRsaEd25519(t *testing.T) {
	privB64, pubB64, err := Crypto.GenerateEd25519KeyPair()
	if err != nil {
		t.Fatalf("GenerateEd25519KeyPair 失败: %v", err)
	}
	if privB64 == "" || pubB64 == "" {
		t.Fatal("Ed25519 密钥对不应为空")
	}

	data := []byte("ed25519 签名测试数据")
	sig, err := Crypto.Ed25519Sign(data, privB64)
	if err != nil {
		t.Fatalf("Ed25519Sign 失败: %v", err)
	}
	// 验签通过
	if ok, err := Crypto.Ed25519Verify(data, sig, pubB64); err != nil || !ok {
		t.Errorf("Ed25519Verify 应通过: %v, %v", ok, err)
	}
	// 篡改数据不通过
	if ok, err := Crypto.Ed25519Verify([]byte("篡改"), sig, pubB64); err != nil || ok {
		t.Errorf("被篡改数据验签不应通过: %v, %v", ok, err)
	}
	// 错公钥不通过（第二对密钥的公钥）
	_, pubX, err := Crypto.GenerateEd25519KeyPair()
	if err != nil {
		t.Fatalf("GenerateEd25519KeyPair(第二对) 失败: %v", err)
	}
	if ok, err := Crypto.Ed25519Verify(data, sig, pubX); err != nil || ok {
		t.Errorf("错公钥验签不应通过: %v, %v", ok, err)
	}

	// 32 字节纯种子私钥兼容格式：签名应与 64 字节私钥一致
	raw, err := base64.StdEncoding.DecodeString(privB64)
	if err != nil || len(raw) != 64 {
		t.Fatalf("私钥原始长度应 64 字节: %d, %v", len(raw), err)
	}
	seedB64 := base64.StdEncoding.EncodeToString(raw[:32])
	sigSeed, err := Crypto.Ed25519Sign(data, seedB64)
	if err != nil {
		t.Fatalf("Ed25519Sign(32 字节种子) 失败: %v", err)
	}
	if sigSeed != sig {
		t.Error("同一种子的两种私钥格式签名应一致")
	}
	if ok, _ := Crypto.Ed25519Verify(data, sigSeed, pubB64); !ok {
		t.Error("种子私钥的签名应通过验签")
	}

	// 错误路径：私钥长度非法（既非 64 也非 32 字节）
	if _, err := Crypto.Ed25519Sign(data, base64.StdEncoding.EncodeToString([]byte("short-key"))); err == nil {
		t.Error("Ed25519Sign 私钥长度非法应返回错误")
	}
	// 错误路径：非法 Base64
	if _, err := Crypto.Ed25519Sign(data, "bad-b64!!!"); err == nil {
		t.Error("Ed25519Sign 非法 Base64 私钥应返回错误")
	}
	// 错误路径：公钥长度非法（非 32 字节）
	if _, err := Crypto.Ed25519Verify(data, sig, base64.StdEncoding.EncodeToString([]byte("tooshortpubkey"))); err == nil {
		t.Error("Ed25519Verify 公钥长度非法应返回错误")
	}
	// 错误路径：非法 Base64 公钥 / 非法 hex 签名
	if _, err := Crypto.Ed25519Verify(data, sig, "bad-b64!!!"); err == nil {
		t.Error("Ed25519Verify 非法 Base64 公钥应返回错误")
	}
	if _, err := Crypto.Ed25519Verify(data, "zz-非hex", pubB64); err == nil {
		t.Error("Ed25519Verify 非法 hex 签名应返回错误")
	}
}

/* ---------------- ECDSA（SHA256withECDSA） ---------------- */

func TestRsaEcdsa(t *testing.T) {
	// 测试内生成 P-256 密钥对（PKCS#8 / X.509 DER 的 Base64，与 RSA 传递格式一致）
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("生成 ECDSA 密钥失败: %v", err)
	}
	privDer, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("序列化私钥失败: %v", err)
	}
	pubDer, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatalf("序列化公钥失败: %v", err)
	}
	privB64 := base64.StdEncoding.EncodeToString(privDer)
	pubB64 := base64.StdEncoding.EncodeToString(pubDer)

	data := []byte("ecdsa 签名测试 ES256")
	sig, err := Crypto.EcdsaSignSha256(data, privB64)
	if err != nil {
		t.Fatalf("EcdsaSignSha256 失败: %v", err)
	}
	// 验签通过
	if ok, err := Crypto.EcdsaVerifySha256(data, sig, pubB64); err != nil || !ok {
		t.Errorf("EcdsaVerifySha256 应通过: %v, %v", ok, err)
	}
	// 篡改数据不通过
	if ok, err := Crypto.EcdsaVerifySha256([]byte("篡改"), sig, pubB64); err != nil || ok {
		t.Errorf("被篡改数据验签不应通过: %v, %v", ok, err)
	}
	// 错误路径：非法 Base64 私钥/公钥、非法 hex 签名
	if _, err := Crypto.EcdsaSignSha256(data, "bad-key!!!"); err == nil {
		t.Error("EcdsaSignSha256 非法 Base64 私钥应返回错误")
	}
	if _, err := Crypto.EcdsaVerifySha256(data, sig, "bad-pub!!!"); err == nil {
		t.Error("EcdsaVerifySha256 非法 Base64 公钥应返回错误")
	}
	if _, err := Crypto.EcdsaVerifySha256(data, "zz-非hex", pubB64); err == nil {
		t.Error("EcdsaVerifySha256 非法 hex 签名应返回错误")
	}
	// 错误路径：RSA 公钥不是 ECDSA 公钥（1024 位快速生成）
	_, rsaPub, err := Crypto.GenerateRsaKeyPair(1024)
	if err != nil {
		t.Fatalf("GenerateRsaKeyPair(1024) 失败: %v", err)
	}
	if _, err := Crypto.EcdsaVerifySha256(data, sig, rsaPub); err == nil {
		t.Error("用 RSA 公钥验 ECDSA 签名应返回错误")
	}
}
