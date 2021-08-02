package main

import (
	"context"
	"errors"
	"github.com/go-kit/kit/endpoint"
	"strings"
)

type CryptoEndpoints struct {
	AesCrypotEndpoint   endpoint.Endpoint
	HealthCheckEndpoint endpoint.Endpoint
}

//define type
var (
	ErrInvalidRequestType = errors.New("Error")
)

type CryptoAESRequest struct {
	cryptoType string `json:"cryptoType"`
	modeOp     string `json:"modeOp"`
	originData string `json:"originData"`
}

type CryptoAESResponse struct {
	result string `json:"result"`
	errmsg error  `json:"errmsg"`
}

func MakeCryptoEndpoints(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{})  (response interface{}, err error) {
		req := request.(CryptoAESRequest)
		var(
			crytoType,modeOp,originData string
			err error
		)
		crytoType := req.cryptoType 
		if strings.EqualFold(,"")

	}
}
