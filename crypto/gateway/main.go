package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/go-kit/kit/log"
	"github.com/hashicorp/consul/api"
	"github.com/openzipkin/zipkin-go"
	zipkinhttpsvr "github.com/openzipkin/zipkin-go/middleware/http"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
)

/*
  ./gateway
*/
func main() {
	var (
		consulHost  = flag.String("consul.host", "127.0.0.1", "consul ip address")
		consulPort  = flag.String("consul.port", "8500", "consul port")
		zipkinHost  = flag.String("zipkin.host", "127.0.0.1", "Zipkin server url")
		zipkinPort  = flag.String("zipkin.url", "9411", "Zipkin server port")
		gatewayPort = flag.String("gateway.port", "9091", "gateway port")
		hystrixPort = flag.String("hystrix.port", "9010", "hystrix port")
	)
	flag.Parse()
	zipkinURL := fmt.Sprintf("http://%s:%s/api/v2/spans", *zipkinHost, *zipkinPort)

	//创建日志组件
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	logger.Log("consulPort", consulHost, "consulPort", consulPort)
	logger.Log("zipkinHost", zipkinHost, "zipkinPort", zipkinPort)
	logger.Log("zipkinURL", zipkinURL)
	logger.Log("gatewayPort", gatewayPort, "hystrixPort", hystrixPort)

	var zipkinTracer *zipkin.Tracer
	{
		var (
			err           error
			hostPort      = "localhost:" + *gatewayPort
			serviceName   = "gateway-service"
			useNoopTracer = (zipkinURL == "")
			reporter      = zipkinhttp.NewReporter(zipkinURL)
		)
		defer reporter.Close()
		zEP, _ := zipkin.NewEndpoint(serviceName, hostPort)
		zipkinTracer, err = zipkin.NewTracer(
			reporter, zipkin.WithLocalEndpoint(zEP), zipkin.WithNoopTracer(useNoopTracer),
		)
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
		if !useNoopTracer {
			logger.Log("tracer", "Zipkin", "type", "Native", "URL", zipkinURL)
		}
	}
	// 创建consul api客户端
	consulConfig := api.DefaultConfig()
	consulConfig.Address = "http://" + *consulHost + ":" + *consulPort
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		logger.Log("err", err)
		os.Exit(1)
	}
	//创建反向代理
	//proxy := NewReverseProxy(consulClient, logger)
	tags := map[string]string{
		"component": "gateway_server",
	}

	hystrixRouter := Routes(consulClient, zipkinTracer, "Circuit Breaker:Service unavailable", logger)

	handler := zipkinhttpsvr.NewServerMiddleware(
		zipkinTracer,
		zipkinhttpsvr.SpanName("gateway"),
		zipkinhttpsvr.TagResponseSize(true),
		zipkinhttpsvr.ServerTags(tags),
	)(hystrixRouter)

	errc := make(chan error)
	//启用hystrix实时监控，监听端口为9010
	hystrixStreamHandler := hystrix.NewStreamHandler()
	hystrixStreamHandler.Start()
	go func() {
		errc <- http.ListenAndServe(net.JoinHostPort("", *hystrixPort), hystrixStreamHandler)
	}()
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	//开始监听
	go func() {
		logger.Log("transport", "HTTP", "addr", *gatewayPort)
		errc <- http.ListenAndServe(":"+*gatewayPort, handler)
	}()

	// 开始运行，等待结束
	logger.Log("exit", <-errc)
}

/*
// NewReverseProxy 创建反向代理处理方法
func NewReverseProxy(client *api.Client, logger log.Logger) *httputil.ReverseProxy {

	//创建Director
	director := func(req *http.Request) {

		//查询原始请求路径，如：/Crypto
		reqPath := req.URL.Path
		if reqPath == "" {
			return
		}
		//按照分隔符'/'对路径进行分解，获取服务名称serviceName
		pathArray := strings.Split(reqPath, "/")
		serviceName := pathArray[1]

		//调用consul api查询serviceName的服务实例列表
		result, _, err := client.Catalog().Service(serviceName, "", nil)
		if err != nil {
			logger.Log("ReverseProxy failed", "query service instace error", err.Error())
			return
		}

		if len(result) == 0 {
			logger.Log("ReverseProxy failed", "no such service instance", serviceName)
			return
		}

		//重新组织请求路径，去掉服务名称部分
		//destPath := strings.Join(pathArray[2:], "/")

		//随机选择一个服务实例
		tgt := result[rand.Int()%len(result)]
		logger.Log("service id", tgt.ServiceID)

		//设置代理服务地址信息
		req.URL.Scheme = "http"
		req.URL.Host = fmt.Sprintf("%s:%d", tgt.ServiceAddress, tgt.ServicePort)
		s, _ := ioutil.ReadAll(req.Body)
		//需要重新设置body
		req.Body = ioutil.NopCloser(bytes.NewBuffer(s))
		destPath := retriveServiceURI(string(s))
		req.URL.Path += "/" + destPath

			fmt.Println("req.RequestURI is :" + req.RequestURI + " reqPath :" + reqPath)
			fmt.Println("req.Host is :" + req.Host + " req.Method :" + req.Method)
			fmt.Println("req.URL.Host is :" + req.URL.Host + " req.URL.Path :" + req.URL.Path)
			fmt.Println("req.URL.Hostname() is :" + req.URL.Hostname() + " req.URL.Port() :" + req.URL.Port())
			fmt.Println("req.URL.RequestURI() is :" + req.URL.RequestURI())
			fmt.Println("string(s) is :" + string(s))

	}
	return &httputil.ReverseProxy{Director: director}

}
*/
