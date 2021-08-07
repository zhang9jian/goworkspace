package main

import (
	"errors"
	"os"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/go-kit/kit/log"
	"github.com/hashicorp/consul/api"
)

func CircurBreaker(serviceName string, logger log.Logger) error {
	hystrix.ConfigureCommand(serviceName, hystrix.CommandConfig{
		Timeout:                1000,  // 单次请求 超时时间
		MaxConcurrentRequests:  1,     // 最大并发量
		SleepWindow:            10000, // 熔断后多久去尝试服务是否可用
		RequestVolumeThreshold: 1,     // 验证熔断的 请求数量, 10秒内采样
		ErrorPercentThreshold:  1,     // 验证熔断的 错误百分比
	})
	consulConfig := api.DefaultConfig()

	consulConfig.Address = "http://" + "127.0.0.1" + ":8500"
	consulClient, err := api.NewClient(consulConfig)

	if err != nil {
		logger.Log("err", err)
		os.Exit(1)
	}

	return CircurBreakerStatus(consulClient, serviceName, logger)
}

func CircurBreakerStatus(consulAPI *api.Client, serviceName string, logger log.Logger) error {
	err := hystrix.Do(serviceName, func() (err error) {
		result, _, err := consulAPI.Catalog().Service(serviceName, "", nil)
		if err != nil {
			logger.Log("router failed", "query service  error", err.Error())
			return errors.New("query service instace error")
		}

		if len(result) == 0 {
			logger.Log("ReverseProxy failed", "no such service instance", serviceName)
			return errors.New("no such service instance")
		}

		return nil
	}, func(err error) error {
		//run执行失败，返回fallback信息
		logger.Log("fallback error description", err.Error())

		return errors.New("fallback error description")
	})

	// Do方法执行失败，响应错误信息
	if err != nil {
		return errors.New("do error")
	}

	return err
}
