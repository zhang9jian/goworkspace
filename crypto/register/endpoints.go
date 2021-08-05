package main

import (
	"context"
	"errors"
	"strings"

	"github.com/go-kit/kit/endpoint"
)

/*
 * 定义端点服务以及请求响应结构
 */
type CryptoEndpoints struct {
	CryptoAESEndpoint   endpoint.Endpoint
	HealthCheckEndpoint endpoint.Endpoint
}

//define type
var (
	errInvalidRequestType = errors.New("Error Crypto Type or Mode")
)

//定义加解密请求结构
type CryptoRequest struct {
	CryptoType string `json:"type"` //加解密类型:AES
	TypeofOp   string `json:"oper"` //操作类型:0:加密 1:解密
	ModeOfWork string `json:"mode"` //工作模式: CBC ECB
	Data       string `json:"data"` //待处理数据
}

//定义加解密响应结构
type CryptoResponse struct {
	Result string `json:"result"`
	Errmsg error  `json:"errmsg"`
}

//定义AES(CBC)端点服务endpoint.Endpoint
func MakeCryptoAESEndpoints(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(CryptoRequest)
		var (
			cryptoType       string
			typeofOp         string
			modeOfWork, data string
			res              string
			calError         error
		)
		cryptoType = req.CryptoType
		typeofOp = req.TypeofOp
		modeOfWork = req.ModeOfWork
		data = req.Data

		if !strings.EqualFold(cryptoType, "AES") {
			return nil, errInvalidRequestType
		}

		if typeofOp == "0" { //加密
			res, calError = svc.Encrypt(data, secretKey, modeOfWork)
		} else if typeofOp == "1" { //解密
			res, calError = svc.Decrypt(data, secretKey, modeOfWork)
		} else {
			return nil, errInvalidRequestType
		}
		return CryptoResponse{Result: res, Errmsg: calError}, nil
	}
}

// HealthRequest 健康检查请求结构
type HealthRequest struct{}

// HealthResponse 健康检查响应结构
type HealthResponse struct {
	Status bool `json:"status"`
}

// MakeHealthCheckEndpoint 创建健康检查Endpoint
func MakeHealthCheckEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		status := svc.HealthCheck()
		return HealthResponse{status}, nil
	}
}
