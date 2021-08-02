package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/log"
)

func main() {
	ctx := context.Background()
	errChan := make(chan error)

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	var svc Service
	svc = CryptoAESService{}

	svc = LoggingMiddleware(logger)(svc)

	cryptoAESEndpoint := MakeCryptoAESEndpoints(svc)
	healthEndpoint := MakeHealthCheckEndpoint(svc)

	endpts := CryptoEndpoints{
		CryptoAESEndpoint:   cryptoAESEndpoint,
		HealthCheckEndpoint: healthEndpoint,
	}

	r := MakeHttpHandler(ctx, endpts, logger)

	go func() {
		fmt.Println("Http Server start at port:9000")
		handle := r
		errChan <- http.ListenAndServe(":9000", handle)
	}()
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	fmt.Println(<-errChan)
}
