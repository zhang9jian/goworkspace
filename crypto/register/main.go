package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	kitzipkin "github.com/go-kit/kit/tracing/zipkin"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"
)

/*
 ./register -consul.service Crypto
*/
func main() {
	var (
		consulHost   = flag.String("consul.host", "127.0.0.1", "consul ip address")
		consulPort   = flag.String("consul.port", "8500", "consul port")
		consulSevice = flag.String("consul.service", "", "consul register server")
		serviceHost  = flag.String("service.host", "127.0.0.1", "service ip address")
		servicePort  = flag.String("service.port", "9000", "service port")
		zipkinHost   = flag.String("zipkin.host", "127.0.0.1", "Zipkin server url")
		zipkinPort   = flag.String("zipkin.url", "9411", "Zipkin server port")
	)
	//解析命令行
	flag.Parse()
	zipkinURL := fmt.Sprintf("http://%s:%s/api/v2/spans", *zipkinHost, *zipkinPort)

	ctx := context.Background()
	errChan := make(chan error)
	//日志结构
	var logger log.Logger
	{

		logger = log.NewLogfmtLogger(os.Stdout) //Stderr
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}
	logger.Log("consulPort", consulHost, "consulPort", consulPort, "consulSevice", consulSevice)
	logger.Log("serviceHost", serviceHost, "servicePort", servicePort)
	logger.Log("zipkinHost", zipkinHost, "zipkinPort", zipkinPort)
	logger.Log("zipkinURL", zipkinURL)
	//链路跟踪
	var zipkinTracer *zipkin.Tracer
	{
		var (
			err           error
			hostPort      = *serviceHost + ":" + *servicePort
			serviceName   = "AES-Service"
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
	//监控指标
	fieldKeys := []string{"method"}
	requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "GO",
		Subsystem: "AES",
		Name:      "request_count",
		Help:      "Number of requests received for Crypto.",
	}, fieldKeys)
	requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "GO",
		Subsystem: "AES",
		Name:      "request_latency",
		Help:      "Total duration of requests in microseconds for Crypto.",
	}, fieldKeys)

	var svc Service
	svc = CryptoAESService{}
	svc = LoggingMiddleware(logger)(svc)
	svc = Metrics(requestCount, requestLatency)(svc)
	//AES服务端点
	cryptoAESEndpoint := MakeCryptoAESEndpoints(svc)
	//增加限流令牌桶每秒100请求。
	ratebucket := rate.NewLimiter(rate.Every(time.Second*1), 100)
	cryptoAESEndpoint = NewTokenBucketLimitterWithBuildIn(ratebucket)(cryptoAESEndpoint)
	//增加链路追踪
	cryptoAESEndpoint = kitzipkin.TraceEndpoint(zipkinTracer, "aes-endpoint")(cryptoAESEndpoint)

	//健康检查端点
	healthEndpoint := MakeHealthCheckEndpoint(svc)

	//把算术运算Endpoint和健康检查Endpoint封装至ArithmeticEndpoints
	endpts := CryptoEndpoints{
		CryptoAESEndpoint:   cryptoAESEndpoint,
		HealthCheckEndpoint: healthEndpoint,
	}
	//在transport上增加链路追踪
	r := MakeHttpHandler(ctx, endpts, zipkinTracer, logger)

	cryptoRegistar := Register(*consulHost, *consulPort, *consulSevice, *serviceHost, *servicePort, logger)

	go func() {
		logger.Log("Http Server start at port", *servicePort)
		//启动前执行注册
		cryptoRegistar.Register()
		handler := r
		errChan <- http.ListenAndServe(":"+*servicePort, handler)
	}()
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	error := <-errChan
	//服务退出取消注册
	cryptoRegistar.Deregister()
	fmt.Println(error)
}
