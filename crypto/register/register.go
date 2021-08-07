package main

import (
	"os"
	"strconv"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"github.com/pborman/uuid"
)

func Register(consulHost, consulPort, consulSvr, svcHost, svcPort string, logger log.Logger) (registar sd.Registrar) {

	// 创建Consul客户端连接
	var client consul.Client
	{
		consulCfg := api.DefaultConfig()
		consulCfg.Address = consulHost + ":" + consulPort
		consulClient, err := api.NewClient(consulCfg)
		if err != nil {
			logger.Log("create consul client error:", err)
			os.Exit(1)
		}

		client = consul.NewClient(consulClient)
	}

	// 设置Consul对服务健康检查的参数，consul 服务端会自己发送请求，来进行健康检查
	check := api.AgentServiceCheck{
		HTTP:     "http://" + svcHost + ":" + svcPort + "/health",
		Interval: "10s",
		Timeout:  "2s",
		Notes:    "Consul  health check service status.",
	}

	port, _ := strconv.Atoi(svcPort)

	//设置微服务向Consul的注册信息
	reg := api.AgentServiceRegistration{
		ID:      consulSvr + uuid.New(),
		Name:    consulSvr,
		Address: svcHost,
		Port:    port,
		Tags:    []string{consulSvr},
		Check:   &check,
	}

	// 执行注册
	registar = consul.NewRegistrar(client, &reg, logger)

	return
}
