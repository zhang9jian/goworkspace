package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd/consul"
	kitzipkin "github.com/go-kit/kit/tracing/zipkin"
	"github.com/hashicorp/consul/api"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
)

func main() {

	// 创建环境变量
	var (
		consulHost = flag.String("consul.host", "", "consul server ip address")
		consulPort = flag.String("consul.port", "", "consul server port")
		zipkinURL  = flag.String("zipkin.url", "", "Zipkin server url")
	)
	flag.Parse()

	//创建日志组件
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}
	var zipkinTracer *zipkin.Tracer
	{
		var (
			err           error
			hostPort      = "localhost:9090"
			serviceName   = "discover-service"
			useNoopTracer = (*zipkinURL == "")
			reporter      = zipkinhttp.NewReporter(*zipkinURL)
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
			logger.Log("tracer", "Zipkin", "type", "Native", "URL", *zipkinURL)
		}
	}

	//创建consul客户端对象
	var client consul.Client
	{
		consulConfig := api.DefaultConfig()

		consulConfig.Address = "http://" + *consulHost + ":" + *consulPort
		consulClient, err := api.NewClient(consulConfig)

		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
		client = consul.NewClient(consulClient)

	}

	ctx := context.Background()
	//创建Endpoint
	discoverEndpoint := MakeDiscoverEndpoint(ctx, client, logger)
	//增加链路追踪
	discoverEndpoint = kitzipkin.TraceEndpoint(zipkinTracer, "discover-endpoint")(discoverEndpoint)
	//创建传输层
	r := MakeHttpHandler(ctx, discoverEndpoint, zipkinTracer, logger)

	errc := make(chan error)

	//启用hystrix实时监控，监听端口为9010
	hystrixStreamHandler := hystrix.NewStreamHandler()
	hystrixStreamHandler.Start()
	go func() {
		errc <- http.ListenAndServe(net.JoinHostPort("", "9010"), hystrixStreamHandler)
	}()

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	//开始监听
	go func() {
		logger.Log("transport", "HTTP", "addr", "9001")
		errc <- http.ListenAndServe(":9001", r)
	}()

	// 开始运行，等待结束
	logger.Log("exit", <-errc)
}
