package main

import "errors"

/*
AES加密CBC填充
*/
type AesCbcCrypto struct {
}

//实现接口方法
func (s AesCbcCrypto) Padding(plainText []byte, blockSize int) []byte {
	PKCS5Padding(plainText, blockSize)
}

func (s AesCbcCrypto) UnPadding(origData []byte) []byte {
	PKCS5UnPadding(origData)
}

func (s AesCbcCrypto) Encrypt(origData, key []byte) ([]byte, error) {
	AESEncryptCBC(origData, key)
}

func (s AesCbcCrypto) Decrypt(encrypted, key []byte) ([]byte, error) {
	AESDecryptCBC(encrypted, key)

}

func (s AesCbcCrypto) HealthCheck() bool {
	return true
}

//实现接口具体逻辑
func (s AesCbcCrypto) PKCS5Padding(plaintext []byte, blockSize int) []byte {
	padding := blockSize - len(plaintext)%blockSize
	paddingText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plaintext, paddingText...)
}

func (s AesCbcCrypto) PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unPadding := int(origData[length-1])
	return origData[:(length - unPadding)]
}

func (s AesCbcCrypto) AESEncryptCBC(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	//AES分组长度为128位，所以blockSize=16，单位字节
	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize]) //初始向量的长度必须等于块block的长度16字节
	encrypted := make([]byte, len(origData))
	blockMode.CryptBlocks(encrypted, origData)
	return encrypted, nil
}

func (s AesCbcCrypto) AESDecryptCBC(encrypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	//AES分组长度为128位，所以blockSize=16，单位字节
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize]) //初始向量的长度必须等于块block的长度16字节
	origData := make([]byte, len(encrypted))
	blockMode.CryptBlocks(origData, encrypted)
	origData = PKCS5UnPadding(origData)
	return origData, nil
}
