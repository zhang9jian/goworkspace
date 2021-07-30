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
	cryptoType string `json:"cryType"`
	mode       string `json:"padding"`
	data       string `json:"data"`
}

type CryptoAESResponse struct {
	result string `json:"result"`
	errmsg error  `json:"errmsg"`
}

func MakeCryptoEndpoints(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) {
		req := request.(CryptoAESRequest)

	}
}
