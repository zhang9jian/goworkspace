package main

import (
	"context"
	"errors"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/metrics"
	"golang.org/x/time/rate"
)

var (
	ErrLimitExceed = errors.New("Rate Limit Exceed")
)

func NewTokenBucketLimitterWithBuildIn(bkt *rate.Limiter) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (reponse interface{}, err error) {
			if !bkt.Allow() {
				return nil, ErrLimitExceed
			}
			return next(ctx, request) //Endpoint是匿名函数，这里是调next方法
		}
	}
}

// metricMiddleware 定义监控中间件，嵌入Service
type metricMiddleware struct {
	Service
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
}

//返回方法签名，构建监控中间件
func Metrics(requestCount metrics.Counter, requestLatency metrics.Histogram) ServiceMiddleware {
	return func(next Service) Service {
		return metricMiddleware{
			next,
			requestCount,
			requestLatency,
		}
	}
}

func (mw metricMiddleware) Encrypt(origData, key, mode string) (string, error) {
	defer func(beign time.Time) {
		lvs := []string{"method", "Encrypt"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(beign).Seconds())
	}(time.Now())

	ret, err := mw.Service.Encrypt(origData, key, mode)
	return ret, err
}

func (mw metricMiddleware) Decrypt(encrypted, key, mode string) (string, error) {
	defer func(beign time.Time) {
		lvs := []string{"method", "Decrypt"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(beign).Seconds())
	}(time.Now())

	ret, err := mw.Service.Decrypt(encrypted, key, mode)
	return ret, err
}

func (mw metricMiddleware) HealthCheck() bool {
	defer func(begin time.Time) {
		lvs := []string{"method", "HealthCheck"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	result := mw.Service.HealthCheck()
	return result

}
