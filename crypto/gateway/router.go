package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/go-kit/kit/log"
	"github.com/hashicorp/consul/api"
	"github.com/openzipkin/zipkin-go"
	zipkinhttpsvr "github.com/openzipkin/zipkin-go/middleware/http"
)

// HystrixRouter hystrix路由
type HystrixRouter struct {
	svcMap       *sync.Map      //服务实例，存储已经通过hystrix监控服务列表
	logger       log.Logger     //日志工具
	fallbackMsg  string         //回调消息
	consulClient *api.Client    //consul客户端对象
	tracer       *zipkin.Tracer //服务追踪对象
}

func Routes(client *api.Client, zikkinTracer *zipkin.Tracer, fbMsg string, logger log.Logger) http.Handler {
	return HystrixRouter{
		svcMap:       &sync.Map{},
		logger:       logger,
		fallbackMsg:  fbMsg,
		consulClient: client,
		tracer:       zikkinTracer,
	}
}

func (router HystrixRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reqPath := r.URL.Path
	if reqPath == "" {
		return
	}
	//按照分隔符'/'对路径进行分解，获取服务名称serviceName
	pathArray := strings.Split(reqPath, "/")
	serviceName := pathArray[1]
	//检查是否已经加入监控
	if _, ok := router.svcMap.Load(serviceName); !ok {
		//把serviceName作为命令对象，设置参数
		hystrix.ConfigureCommand(serviceName, hystrix.CommandConfig{Timeout: 1000})
		router.svcMap.Store(serviceName, serviceName)
	}
	//执行命令
	err := hystrix.Do(serviceName, func() (err error) {
		//调用consul api查询serviceNam
		result, _, err := router.consulClient.Catalog().Service(serviceName, "", nil)
		if err != nil {
			router.logger.Log("ReverseProxy failed", "query service instace error", err.Error())
			return
		}
		if len(result) == 0 {
			router.logger.Log("ReverseProxy failed", "no such service instance", serviceName)
			return errors.New("no such service instance")
		}

		director := func(req *http.Request) {
			//随机选择一个服务实例
			tgt := result[rand.Int()%len(result)]
			router.logger.Log("service id", tgt.ServiceID)
			fmt.Println("service id" + tgt.ServiceID)
			//设置代理服务地址信息
			req.URL.Scheme = "http"
			req.URL.Host = fmt.Sprintf("%s:%d", tgt.ServiceAddress, tgt.ServicePort)
			s, _ := ioutil.ReadAll(req.Body)
			//需要重新设置body
			req.Body = ioutil.NopCloser(bytes.NewBuffer(s))
			destPath := retriveServiceURI(string(s))
			req.URL.Path += "/" + destPath
			/*fmt.Println("req.RequestURI is :" + req.RequestURI + " reqPath :" + reqPath)
			fmt.Println("req.Host is :" + req.Host + " req.Method :" + req.Method)
			fmt.Println("req.URL.Host is :" + req.URL.Host + " req.URL.Path :" + req.URL.Path)
			fmt.Println("req.URL.Hostname() is :" + req.URL.Hostname() + " req.URL.Port() :" + req.URL.Port())
			fmt.Println("req.URL.RequestURI() is :" + req.URL.RequestURI())
			fmt.Println("string(s) is :" + string(s))*/
		}

		var proxyError error = nil
		// 为反向代理增加追踪逻辑，使用如下RoundTrip代替默认Transport
		roundTrip, _ := zipkinhttpsvr.NewTransport(router.tracer, zipkinhttpsvr.TransportTrace(true))

		//反向代理失败时错误处理
		errorHandler := func(ew http.ResponseWriter, er *http.Request, err error) {
			proxyError = err
		}

		proxy := &httputil.ReverseProxy{
			Director:     director,
			Transport:    roundTrip,
			ErrorHandler: errorHandler,
		}
		proxy.ServeHTTP(w, r)

		return proxyError

	}, func(err error) error {
		//run执行失败，返回fallback信息
		router.logger.Log("fallback error description", err.Error())
		return errors.New(router.fallbackMsg)
	})

	// Do方法执行失败，响应错误信息
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	}

}

//定义加解密请求结构
type CryptoRequest struct {
	CryptoType string `json:"type"` //加解密类型:AES
	TypeofOp   string `json:"oper"` //操作类型:0:加密 1:解密
	ModeOfWork string `json:"mode"` //工作模式: CBC ECB
	Data       string `json:"data"` //待处理数据
}

func retriveServiceURI(reqBody string) string {
	var data CryptoRequest
	if err := json.Unmarshal([]byte(reqBody), &data); err == nil {
		return data.CryptoType
	} else {
		return ""

	}

}
