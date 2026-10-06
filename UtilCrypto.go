package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
)

// 生成32位md5字串
type utilCrypto struct {
}

func (_self *utilCrypto) GetMd5(input string) string {
	hash := md5.New()
	hash.Write([]byte(input))
	return hex.EncodeToString(hash.Sum(nil))
}

func (_self *utilCrypto) GetSaltMD5(input, salt string) string {
	hash := md5.New()
	//salt = "salt123456" //盐值
	io.WriteString(hash, input+salt)
	result := fmt.Sprintf("%x", hash.Sum(nil))
	return result
}
