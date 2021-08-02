package main

import (
	"time"

	"github.com/go-kit/kit/log"
)

/*
 * 定义并实现中间件结构，中间件构造方法
 */
type loggingMiddleware struct {
	Service
	logger log.Logger
}

func LoggingMiddleware(logger log.Logger) ServiceMiddleware {
	return func(next Service) Service {
		return loggingMiddleware{next, logger}
	}
}

func (mw loggingMiddleware) Encrypt(data, key, mode string) (ret string, errmsg error) {
	defer func(beign time.Time) {
		mw.logger.Log(
			"function", "Encrypt",
			"Data", data,
			"key", key,
			"return code", errmsg,
			"return msg", ret,
			"took", time.Since(beign),
		)
	}(time.Now())

	ret, errmsg = mw.Service.Encrypt(data, key, mode)
	return ret, errmsg
}

func (mw loggingMiddleware) Decrypt(data, key, mode string) (ret string, errmsg error) {
	defer func(beign time.Time) {
		mw.logger.Log(
			"function", "Decrypt",
			"Data", data,
			"key", key,
			"return code", errmsg,
			"return msg", ret,
			"took", time.Since(beign),
		)
	}(time.Now())
	ret, errmsg = mw.Service.Decrypt(data, key, mode)
	return ret, errmsg
}
func (mw loggingMiddleware) HealthCheck() (result bool) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"function", "HealthChcek",
			"result", result,
			"took", time.Since(begin),
		)
	}(time.Now())
	result = mw.Service.HealthCheck()
	return
}
