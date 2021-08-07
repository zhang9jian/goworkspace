package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	kithttp "github.com/go-kit/kit/transport/http"
)

func cryptoFactory(_ context.Context, method, path string) sd.Factory {
	return func(instance string) (endpoint endpoint.Endpoint, closer io.Closer, err error) {
		if !strings.HasPrefix(instance, "http") {
			instance = "http://" + instance
		}

		tgt, err := url.Parse(instance)
		if err != nil {
			return nil, nil, err
		}
		tgt.Path = path

		var (
			enc kithttp.EncodeRequestFunc
			dec kithttp.DecodeResponseFunc
		)
		enc, dec = encodeCryptoAESRequest, decodeCryptoAESReponse

		return kithttp.NewClient(method, tgt, enc, dec).Endpoint(), nil, nil

	}
}

func encodeCryptoAESRequest(ctx context.Context, req *http.Request, request interface{}) error {
	reqStruct := request.(CryptoRequest)
	data, _ := json.MarshalIndent(reqStruct, " ", "")
	req.Body = ioutil.NopCloser(bytes.NewBuffer(data))
	return nil
}

func decodeCryptoAESReponse(ctx context.Context, resp *http.Response) (interface{}, error) {
	var response CryptoResponse
	var s map[string]interface{}

	if respCode := resp.StatusCode; respCode >= 400 {
		if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
			return nil, err
		}
		return nil, errors.New(s["error"].(string) + "\n")
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response, nil

}
