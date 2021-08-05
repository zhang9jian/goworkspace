package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
		fmt.Println("instance is " + instance)
		tgt, err := url.Parse(instance)
		if err != nil {
			return nil, nil, err
		}
		tgt.Path = path
		fmt.Println("tgt.Host is " + tgt.Host)
		fmt.Println("tgt.Port is " + tgt.Port())
		fmt.Println("tgt.Path is " + tgt.Path)
		fmt.Println("tgt  is " + tgt.String())
		var (
			enc kithttp.EncodeRequestFunc
			dec kithttp.DecodeResponseFunc
		)
		enc, dec = encodeCryptoAESRequest, decodeCryptoAESReponse

		return kithttp.NewClient(method, tgt, enc, dec).Endpoint(), nil, nil

	}
}

func encodeCryptoAESRequest(ctx context.Context, req *http.Request, request interface{}) error {
	/*aesreq := request.(CryptoRequest)
	p := "/" + aesreq.CryptoType
	req.URL.Path += p
	fmt.Println("req.URL.Path:" + req.URL.Path)*/

	str := request.(CryptoRequest)
	fmt.Println("encodeCryptoAESRequest str :" + str.CryptoType)
	data, _ := json.MarshalIndent(str, " ", "")
	//	req.RequestURI = req.URL.Path
	fmt.Println("encodeCryptoAESRequest json :" + string(data))
	fmt.Println("req.Host:" + req.Host)
	fmt.Println("req.Method:" + req.Method)
	fmt.Println("req.RequestURI:" + req.RequestURI)
	fmt.Println("req.URL.Host:" + req.URL.Host)
	fmt.Println("req.URL.Path:" + req.URL.Path)

	req.Body = ioutil.NopCloser(bytes.NewBuffer(data))

	return nil
}

func decodeCryptoAESReponse(ctx context.Context, resp *http.Response) (interface{}, error) {
	var response CryptoResponse
	var s map[string]interface{}

	if respCode := resp.StatusCode; respCode >= 400 {
		if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
			fmt.Println(respCode)
			fmt.Println(s)
			return nil, err
		}
		fmt.Println("in error  respCode >= 400")
		return nil, errors.New(s["error"].(string) + "\n")
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Println("in error .Decode(&response)")
		return nil, err
	}
	return response, nil

}
