package main

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/marigold-dev/tzproxy/middlewares"
	utils "github.com/marigold-dev/tzproxy/utils"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() - 1)
	println(runtime.NumCPU() - 1)
	config := utils.NewConfig()
	store := memory.NewStore()

	e := echo.New()

	url, err := url.Parse(config.TezosHost)
	if err != nil {
		e.Logger.Fatal(err)
	}

	targets := []*middleware.ProxyTarget{{URL: url}}
	balancer := middleware.NewRoundRobinBalancer(targets)

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   65 * time.Second,
			KeepAlive: 65 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          300,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	proxyConfig := middleware.ProxyConfig{
		Skipper:    middleware.DefaultSkipper,
		ContextKey: "target",
		Balancer:   balancer,
		Transport:  transport,
	}

	// Middlewares
	e.Use(middleware.Recover())
	e.Use(middleware.RequestLoggerWithConfig(config.RequestLoggerConfig))
	e.Use(echoprometheus.NewMiddleware("proxy"))
	e.Use(middlewares.CORS(config))
	e.Use(middlewares.RateLimit(store, config))
	e.Use(middlewares.BlockRoutes(config))
	e.Use(middlewares.Cache(config))
	e.Use(middlewares.Gzip(config))
	e.Use(middleware.ProxyWithConfig(proxyConfig))

	// Start metrics server
	eMetrics := echo.New()
	eMetrics.GET("/metrics", echoprometheus.NewHandler())
	go func() {
		if err := eMetrics.Start(":9000"); err != nil && err != http.ErrServerClosed {
			eMetrics.Logger.Fatal(err)
		}
	}()

	// Start proxy
	go func() {
		if err := e.Start(config.Host); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("Shutting down the server")
			e.Logger.Fatal(err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
