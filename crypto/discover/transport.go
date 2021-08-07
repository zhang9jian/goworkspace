package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/zipkin"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	gozipkin "github.com/openzipkin/zipkin-go"
)

func MakeHttpHandler(ctx context.Context, endpoint endpoint.Endpoint, zipkinTracer *gozipkin.Tracer, logger log.Logger) http.Handler {

	r := mux.NewRouter()
	zipkinServer := zipkin.HTTPServerTrace(zipkinTracer, zipkin.Name("dis-http-handle"))

	options := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(kithttp.DefaultErrorEncoder),
		zipkinServer,
	}

	r.Methods("POST").Path("/Crypto").Handler(kithttp.NewServer(
		endpoint,
		decodeDiscoverRequest,
		encodeDiscoverResponse,
		options...,
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

func decodeDiscoverRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var request CryptoRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}
	err := CircurBreaker("AES", logger)
	fmt.Println("In decode")
	if err != nil {
		fmt.Println("decode error")
		return err, nil
	}
	fmt.Println("decode noerror")
	return request, nil
}

func encodeDiscoverResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
