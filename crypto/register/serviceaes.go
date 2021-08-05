package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"strings"
)

/* 文件名：sericeaescbd.go
 * 功能：实现CBC填充模式的AES加解密方法
 * Encrypt：
 *    输入：string明文 ；输出： Base64编码密文
 * Decrypt：
 *    输入：Base64编码密文； 输出： string明文
 */
type CryptoAESService struct {
}

//实现接口具体逻辑
func (s CryptoAESService) PKCS5Padding(plaintext []byte, blockSize int) []byte {
	padding := blockSize - len(plaintext)%blockSize
	paddingText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plaintext, paddingText...)
}
func (s CryptoAESService) PKCS5UnPadding(origData []byte) []byte {
	defer func() {
		fmt.Println("error PKCS5UnPadding")
	}()
	length := len(origData)
	unPadding := int(origData[length-1])
	return origData[:(length - unPadding)]
}

func (s CryptoAESService) PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func (s CryptoAESService) PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func (s CryptoAESService) Encrypt(origData, key, mode string) (string, error) {
	var (
		encryptedByte []byte
		err           error
	)

	if strings.EqualFold(mode, "CBC") {
		encryptedByte, err = s.AESEncryptCBC([]byte(origData), []byte(key))

	} else if strings.EqualFold(mode, "ECB") {
		encryptedByte, err = s.AESEncryptECB([]byte(origData), []byte(key))

	}
	ret := base64.StdEncoding.EncodeToString(encryptedByte)
	return ret, err

}

func (s CryptoAESService) Decrypt(data, key, mode string) (string, error) {
	//Base64解码为字节
	encryptedByte, e := base64.StdEncoding.DecodeString(data)
	if e != nil {
		return "", e
	}
	//定义明文数据结构
	var (
		decryptByte []byte
		decryptData string
		err         error
	)
	//执行解密
	if strings.EqualFold(mode, "CBC") {
		decryptByte, err = s.AESDecryptCBC(encryptedByte, []byte(key))

	} else if strings.EqualFold(mode, "ECB") {
		decryptByte, err = s.AESDecryptECB(encryptedByte, []byte(key))
	}
	if err != nil {
		return "", err
	}
	//解密[]byte转为string
	decryptData = string(decryptByte)
	return decryptData, err

}

func (s CryptoAESService) AESEncryptCBC(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	//AES分组长度为128位，所以blockSize=16，单位字节
	blockSize := block.BlockSize()
	origData = s.PKCS7Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize]) //初始向量的长度必须等于块block的长度16字节
	encrypted := make([]byte, len(origData))
	blockMode.CryptBlocks(encrypted, origData)
	return encrypted, nil
}

func (s CryptoAESService) AESDecryptCBC(encrypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	//AES分组长度为128位，所以blockSize=16，单位字节
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize]) //初始向量的长度必须等于块block的长度16字节
	origData := make([]byte, len(encrypted))
	blockMode.CryptBlocks(origData, encrypted)
	origData = s.PKCS7UnPadding(origData)
	return origData, nil
}

func (s CryptoAESService) AESEncryptECB(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	data = s.PKCS7Padding(data, block.BlockSize())
	decrypted := make([]byte, len(data))
	size := block.BlockSize()

	for bs, be := 0, size; bs < len(data); bs, be = bs+size, be+size {
		block.Encrypt(decrypted[bs:be], data[bs:be])
	}

	return decrypted, nil
}

func (s CryptoAESService) AESDecryptECB(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	decrypted := make([]byte, len(data))
	size := block.BlockSize()

	for bs, be := 0, size; bs < len(data); bs, be = bs+size, be+size {
		block.Decrypt(decrypted[bs:be], data[bs:be])
	}

	return s.PKCS7UnPadding(decrypted), nil
}

func (s CryptoAESService) HealthCheck() bool {
	return true
}
