package main

import (
	"context"
	"net"
	"net/http"
	"net/http/pprof"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/fraidev/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/marigold-dev/tzproxy/middlewares"
	utils "github.com/marigold-dev/tzproxy/utils"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func main() {

	config := utils.NewConfig()
	store := memory.NewStore()

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

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
	e.Use(middlewares.CORS(config))
	e.Use(middlewares.RateLimit(store, config))
	e.Use(middlewares.BlockRoutes(config))
	e.Use(middlewares.Cache(config))
	e.Use(middlewares.Gzip(config))
	e.Use(middleware.ProxyWithConfig(proxyConfig))

	// Start metrics server
	go func() {
		metrics := echo.New()

		pp := http.NewServeMux()
		pp.HandleFunc("/debug/pprof/", pprof.Index)
		pp.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		pp.HandleFunc("/debug/pprof/profile", pprof.Profile)
		pp.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		pp.HandleFunc("/debug/pprof/trace", pprof.Trace)
		metrics.GET("/debug/pprof/*", echo.WrapHandler(pp))

		metrics.GET("/metrics", echoprometheus.NewHandler())
		metrics.HideBanner = true
		metrics.HidePort = true
		if err := metrics.Start(":9000"); err != nil && err != http.ErrServerClosed {
			metrics.Logger.Fatal(err)
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
