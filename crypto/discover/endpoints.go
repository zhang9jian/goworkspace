package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/afex/hystrix-go/hystrix"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
)

// MakeDiscoverEndpoint 使用consul.Client创建服务发现Endpoint
// 为了方便这里默认了一些参数
func MakeDiscoverEndpoint(ctx context.Context, client consul.Client, logger log.Logger) endpoint.Endpoint {
	serviceName := "AES"
	tags := []string{"AES"}
	passingOnly := true
	duration := 500 * time.Millisecond

	//基于consul客户端、服务名称、服务标签等信息，
	// 创建consul的连接实例，
	// 可实时查询服务实例的状态信息
	instancer := consul.NewInstancer(client, logger, serviceName, tags, passingOnly)

	//针对AES接口创建sd.Factory
	factory := cryptoFactory(ctx, "POST", "Crypto/AES")

	//使用consul连接实例（发现服务系统）、factory创建sd.Factory
	endpointer := sd.NewEndpointer(instancer, factory, logger)

	err := hystrix.Do(serviceName, func() (err error) {
		if endpointer == nil {
			fmt.Println("endpointer is nil, serviceName" + serviceName)
			logger.Log("ReverseProxy failed", "query service instace error", err.Error())
			return nil
		}
		fmt.Println("endpointer is not nil, serviceName" + serviceName)
		return nil
	}, func(err error) error {
		fmt.Println("in error 1111" + serviceName)
		//run执行失败，返回fallback信息
		logger.Log("fallback error description", err.Error())
		return errors.New("circur break")
	})
	if err != nil {
		fmt.Println("in error 2222" + err.Error())
		return nil
	}

	//创建RoundRibbon负载均衡器
	fmt.Println("balancer created ")
	balancer := lb.NewRoundRobin(endpointer)

	//为负载均衡器增加重试功能，同时该对象为endpoint.Endpoint
	retry := lb.Retry(1, duration, balancer)

	return retry
}
