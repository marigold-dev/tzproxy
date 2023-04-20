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

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	middlewares "github.com/marigold-dev/tzproxy/middlewares"
	utils "github.com/marigold-dev/tzproxy/utils"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
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
		MaxIdleConns:          100,
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

	e.Use(middleware.Recover())
	e.Use(middleware.RequestLoggerWithConfig(config.RequestLoggerConfig))
	e.Use(middlewares.CORS(config))
	e.Use(middlewares.RateLimit(store, config))
	e.Use(middlewares.BlockRoutes(config))
	e.Use(middlewares.Cache(config))
	e.Use(middleware.Gzip())
	e.Use(middleware.ProxyWithConfig(proxyConfig))

	// Start server
	go func() {
		if err := e.Start(config.Host); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("Shutting down the server")
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
