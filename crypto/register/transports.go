package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/zipkin"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	gozipkin "github.com/openzipkin/zipkin-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ErrorBadRequest = errors.New("invalid request parameter")
)

func MakeHttpHandler(ctx context.Context, endpoints CryptoEndpoints, zipkinTracer *gozipkin.Tracer, logger log.Logger) http.Handler {
	r := mux.NewRouter()

	zipkinServer := zipkin.HTTPServerTrace(zipkinTracer, zipkin.Name("aes-http-handle"))

	options := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(kithttp.DefaultErrorEncoder),
		zipkinServer,
	}

	optionsNoTracer := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(kithttp.DefaultErrorEncoder),
	}

	r.Methods("POST").Path("/Crypto/AES").Handler(kithttp.NewServer(
		endpoints.CryptoAESEndpoint,
		decodeCryptoAESCBCRequest,
		encodeCryptoAESCBCResponse,
		options...,
	))

	// create health check handler
	r.Methods("GET").Path("/health").Handler(kithttp.NewServer(
		endpoints.HealthCheckEndpoint,
		decodeHealthCheckRequest,
		encodeHealthCheckResponse,
		optionsNoTracer..., //options...,
	))
	r.Path("/metrics").Handler(promhttp.Handler())

	return r
}

// decodeCryptoAESCBCRequest decode request params to struct
func decodeCryptoAESCBCRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req CryptoRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}

	return req, nil
}

// encodeCryptoAESCBCResponse encode response to return
func encodeCryptoAESCBCResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

// decodeHealthCheckRequest decode request
func decodeHealthCheckRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return HealthRequest{}, nil
}

func encodeHealthCheckResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
