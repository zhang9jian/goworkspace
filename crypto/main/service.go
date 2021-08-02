package main

/*
 基础接口，定义基本的业务功能
*/
type Service interface {
	Encrypt(origData, key, mode string) (string, error)
	Decrypt(encrypted, key, mode string) (string, error)
	HealthCheck() bool
}

//密钥
var secretKey string = "0123456789abcdef"

//中间件方法
type ServiceMiddleware func(Service) Service
