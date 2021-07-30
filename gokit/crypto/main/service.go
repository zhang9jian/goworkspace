package main

import "errors"

/*
 定义基础结构
*/
type Service interface {
	Padding(plainText []byte, blockSize int) []byte
	UnPadding(origData []byte) []byte
	Encrypt(origData, key []byte) ([]byte, error)
	Decrypt(encrypted, key []byte) ([]byte, error)
	HealthCheck() bool
}
