package utils

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
)

// ------------------ RSA ------------------

// GenerateRsaKeyPair 生成 RSA 密钥对：私钥为 PKCS#8 DER 的 Base64，公钥为
// X.509 SubjectPublicKeyInfo DER 的 Base64（与 hutool RSA 生成的 Base64 密钥格式互通）。
// keySize<=0 时按 1024 位（hutool DEFAULT_KEY_SIZE），生产环境建议 2048 位及以上。
// 场景：接口报文加解密、开放平台签名验签的密钥准备。
func (c *utilCrypto) GenerateRsaKeyPair(keySize int) (privateKeyB64, publicKeyB64 string, err error) {
	if keySize <= 0 {
		keySize = 1024
	}
	priv, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return "", "", err
	}
	privDer, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return "", "", err
	}
	pubDer, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return "", "", err
	}
	return base64.StdEncoding.EncodeToString(privDer), base64.StdEncoding.EncodeToString(pubDer), nil
}

// RsaEncrypt 使用 RSA 公钥加密（RSA/ECB/PKCS1Padding，即 rsa.EncryptPKCS1v15）。
// publicKeyB64 为 X.509 SubjectPublicKeyInfo DER 的 Base64（兼容 PKCS#1 公钥）；
// 明文长度须小于模长-11 字节（如 1024 位密钥最多 117 字节），长数据应改用对称+RSA 混合加密。
// 场景：解密侧为 Java（hutool RSA）时的兼容加密。
func (c *utilCrypto) RsaEncrypt(data []byte, publicKeyB64 string) ([]byte, error) {
	pub, err := rsaLoadPublicKey(publicKeyB64)
	if err != nil {
		return nil, err
	}
	return rsa.EncryptPKCS1v15(rand.Reader, pub, data)
}

// RsaDecrypt 使用 RSA 私钥解密（RSA/ECB/PKCS1Padding）。
// privateKeyB64 为 PKCS#8 DER 的 Base64（兼容 PKCS#1 私钥）。
// 场景：解密 Java 侧（hutool RSA）加密的数据。
func (c *utilCrypto) RsaDecrypt(data []byte, privateKeyB64 string) ([]byte, error) {
	priv, err := rsaLoadPrivateKey(privateKeyB64)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, priv, data)
}

// RsaSignSha1 用 SHA1withRSA（PKCS1v15）签名，返回小写 hex（对应 hutool Sign 的 signHex）。
// 场景：旧开放平台接口签名，新系统建议 RsaSignSha256。
func (c *utilCrypto) RsaSignSha1(data []byte, privateKeyB64 string) (string, error) {
	return rsaSignHash(data, privateKeyB64, crypto.SHA1, sha1.New)
}

// RsaSignSha256 用 SHA256withRSA（PKCS1v15）签名，返回小写 hex。
// 场景：开放平台请求签名、License 授权文件签名。
func (c *utilCrypto) RsaSignSha256(data []byte, privateKeyB64 string) (string, error) {
	return rsaSignHash(data, privateKeyB64, crypto.SHA256, sha256.New)
}

// RsaVerifySha1 校验 SHA1withRSA 签名，signHex 为 hex 签名串；签名不匹配返回 false，密钥/hex 非法返回 error。
func (c *utilCrypto) RsaVerifySha1(data []byte, signHex, publicKeyB64 string) (bool, error) {
	return rsaVerifyHash(data, signHex, publicKeyB64, crypto.SHA1, sha1.New)
}

// RsaVerifySha256 校验 SHA256withRSA 签名，signHex 为 hex 签名串；签名不匹配返回 false，密钥/hex 非法返回 error。
func (c *utilCrypto) RsaVerifySha256(data []byte, signHex, publicKeyB64 string) (bool, error) {
	return rsaVerifyHash(data, signHex, publicKeyB64, crypto.SHA256, sha256.New)
}

// rsaSignHash RSA PKCS1v15 签名的公共实现：按指定哈希摘要后签名，返回小写 hex。
func rsaSignHash(data []byte, privateKeyB64 string, hashAlg crypto.Hash, newHash func() hash.Hash) (string, error) {
	priv, err := rsaLoadPrivateKey(privateKeyB64)
	if err != nil {
		return "", err
	}
	h := newHash()
	h.Write(data)
	sig, err := rsa.SignPKCS1v15(nil, priv, hashAlg, h.Sum(nil))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sig), nil
}

// rsaVerifyHash RSA PKCS1v15 验签的公共实现：签名不匹配返回 false，密钥或 hex 非法返回 error。
func rsaVerifyHash(data []byte, signHex, publicKeyB64 string, hashAlg crypto.Hash, newHash func() hash.Hash) (bool, error) {
	pub, err := rsaLoadPublicKey(publicKeyB64)
	if err != nil {
		return false, err
	}
	sig, err := hex.DecodeString(signHex)
	if err != nil {
		return false, err
	}
	h := newHash()
	h.Write(data)
	if err := rsa.VerifyPKCS1v15(pub, hashAlg, h.Sum(nil), sig); err != nil {
		return false, nil
	}
	return true, nil
}

// ------------------ Ed25519 ------------------

// GenerateEd25519KeyPair 生成 Ed25519 密钥对。
// 格式说明（与 hutool 存原始字节的行为一致）：私钥为 ed25519.PrivateKey 原始 64 字节
// （32 字节种子 + 32 字节公钥）的 Base64；公钥为 32 字节原始公钥的 Base64（非 X.509 DER）。
// 场景：高性能签名（如区块链、软件签名），签名比 RSA 小且快。
func (c *utilCrypto) GenerateEd25519KeyPair() (privateKeyB64, publicKeyB64 string, err error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}
	return base64.StdEncoding.EncodeToString(priv), base64.StdEncoding.EncodeToString(pub), nil
}

// Ed25519Sign 使用 Ed25519 私钥签名，返回小写 hex。
// privateKeyB64 支持 64 字节（种子+公钥，GenerateEd25519KeyPair 的格式）或 32 字节纯种子。
func (c *utilCrypto) Ed25519Sign(data []byte, privateKeyB64 string) (string, error) {
	raw, err := rsaB64Decode(privateKeyB64)
	if err != nil {
		return "", err
	}
	var priv ed25519.PrivateKey
	switch len(raw) {
	case ed25519.PrivateKeySize: // 64 字节：种子+公钥
		priv = ed25519.PrivateKey(raw)
	case ed25519.SeedSize: // 32 字节：纯种子，兼容
		priv = ed25519.NewKeyFromSeed(raw)
	default:
		return "", fmt.Errorf("Ed25519 私钥长度须为 64 字节（种子+公钥）或 32 字节（种子），当前为 %d", len(raw))
	}
	return hex.EncodeToString(ed25519.Sign(priv, data)), nil
}

// Ed25519Verify 校验 Ed25519 签名，signHex 为小写 hex；publicKeyB64 为 32 字节原始公钥的 Base64。
func (c *utilCrypto) Ed25519Verify(data []byte, signHex, publicKeyB64 string) (bool, error) {
	pubRaw, err := rsaB64Decode(publicKeyB64)
	if err != nil {
		return false, err
	}
	if len(pubRaw) != ed25519.PublicKeySize {
		return false, fmt.Errorf("Ed25519 公钥长度须为 32 字节，当前为 %d", len(pubRaw))
	}
	sig, err := hex.DecodeString(signHex)
	if err != nil {
		return false, err
	}
	return ed25519.Verify(ed25519.PublicKey(pubRaw), data, sig), nil
}

// ------------------ ECDSA ------------------

// EcdsaSignSha256 用 SHA256withECDSA 签名（ecdsa.SignASN1，ASN.1/DER 编码，对应 hutool 默认），
// 返回小写 hex。privateKeyB64 为 PKCS#8 DER 的 Base64（与 RSA 私钥传递格式一致，建议 P-256 曲线）。
// 场景：JWT ES256、物联网设备身份认证。
func (c *utilCrypto) EcdsaSignSha256(data []byte, privateKeyB64 string) (string, error) {
	priv, err := rsaLoadEcdsaPrivateKey(privateKeyB64)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(data)
	sig, err := ecdsa.SignASN1(rand.Reader, priv, digest[:])
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sig), nil
}

// EcdsaVerifySha256 校验 SHA256withECDSA 签名，signHex 为小写 hex（ASN.1/DER 编码）；
// publicKeyB64 为 X.509 SubjectPublicKeyInfo DER 的 Base64。
func (c *utilCrypto) EcdsaVerifySha256(data []byte, signHex, publicKeyB64 string) (bool, error) {
	pub, err := rsaLoadEcdsaPublicKey(publicKeyB64)
	if err != nil {
		return false, err
	}
	sig, err := hex.DecodeString(signHex)
	if err != nil {
		return false, err
	}
	digest := sha256.Sum256(data)
	return ecdsa.VerifyASN1(pub, digest[:], sig), nil
}

// ------------------ 内部辅助：密钥解析（Base64 <-> DER） ------------------

// rsaB64Decode 标准 Base64 解码，失败时返回带上下文的错误。
func rsaB64Decode(s string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("Base64 解码失败: %w", err)
	}
	return data, nil
}

// rsaLoadPrivateKey 解析 Base64 私钥为 RSA 私钥：优先 PKCS#8（hutool 默认），兼容 PKCS#1。
func rsaLoadPrivateKey(privateKeyB64 string) (*rsa.PrivateKey, error) {
	der, err := rsaB64Decode(privateKeyB64)
	if err != nil {
		return nil, err
	}
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("私钥不是 RSA 私钥")
		}
		return rsaKey, nil
	}
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}
	return nil, fmt.Errorf("解析 RSA 私钥失败：须为 PKCS#8 或 PKCS#1 DER 的 Base64")
}

// rsaLoadPublicKey 解析 Base64 公钥为 RSA 公钥：优先 X.509 SubjectPublicKeyInfo（hutool 默认），兼容 PKCS#1。
func rsaLoadPublicKey(publicKeyB64 string) (*rsa.PublicKey, error) {
	der, err := rsaB64Decode(publicKeyB64)
	if err != nil {
		return nil, err
	}
	if key, err := x509.ParsePKIXPublicKey(der); err == nil {
		rsaKey, ok := key.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("公钥不是 RSA 公钥")
		}
		return rsaKey, nil
	}
	if key, err := x509.ParsePKCS1PublicKey(der); err == nil {
		return key, nil
	}
	return nil, fmt.Errorf("解析 RSA 公钥失败：须为 X.509 SubjectPublicKeyInfo 或 PKCS#1 DER 的 Base64")
}

// rsaLoadEcdsaPrivateKey 解析 Base64 私钥为 ECDSA 私钥（PKCS#8 DER 的 Base64）。
func rsaLoadEcdsaPrivateKey(privateKeyB64 string) (*ecdsa.PrivateKey, error) {
	der, err := rsaB64Decode(privateKeyB64)
	if err != nil {
		return nil, err
	}
	key, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return nil, fmt.Errorf("解析 ECDSA 私钥失败：须为 PKCS#8 DER 的 Base64: %w", err)
	}
	ecdsaKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("私钥不是 ECDSA 私钥")
	}
	return ecdsaKey, nil
}

// rsaLoadEcdsaPublicKey 解析 Base64 公钥为 ECDSA 公钥（X.509 SubjectPublicKeyInfo DER 的 Base64）。
func rsaLoadEcdsaPublicKey(publicKeyB64 string) (*ecdsa.PublicKey, error) {
	der, err := rsaB64Decode(publicKeyB64)
	if err != nil {
		return nil, err
	}
	key, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return nil, fmt.Errorf("解析 ECDSA 公钥失败：须为 X.509 SubjectPublicKeyInfo DER 的 Base64: %w", err)
	}
	ecdsaKey, ok := key.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("公钥不是 ECDSA 公钥")
	}
	return ecdsaKey, nil
}
