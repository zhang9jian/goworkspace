package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

func MakeHttpHandler(endpoint endpoint.Endpoint) http.Handler {
	r := mux.NewRouter()

	r.Methods("POST").Path("/Crypto").Handler(kithttp.NewServer(
		endpoint,
		decodeDiscoverRequest,
		encodeDiscoverResponse,
	))

	return r
}

// CryptoRequest define request struct
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

func decodeDiscoverRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request CryptoRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	fmt.Println("request.CryptoType:" + request.CryptoType)
	data, _ := json.MarshalIndent(request, " ", "")
	fmt.Println("request is " + string(data))
	return request, nil
}

func encodeDiscoverResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
