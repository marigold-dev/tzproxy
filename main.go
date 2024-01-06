package main

import (
	"context"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"runtime/debug"
	"time"

	"github.com/fraidev/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/marigold-dev/tzproxy/config"
	"github.com/marigold-dev/tzproxy/middlewares"
	"github.com/ziflex/lecho/v3"
)

func main() {
	config := config.NewConfig()

	debug.SetGCPercent(config.ConfigFile.GC.Percent)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Middlewares
	e.Logger = lecho.From(config.Logger)
	e.Use(middleware.Recover())
	e.Use(middleware.RequestLoggerWithConfig(*config.RequestLoggerConfig))
	e.Use(middlewares.CORS(config))
	e.Use(middlewares.RateLimit(config))
	e.Use(middlewares.DenyRoutes(config))
	e.Use(middlewares.Cache(config))
	e.Use(middlewares.Gzip(config))
	e.Use(middlewares.Retry(config))
	e.Use(middleware.ProxyWithConfig(*config.ProxyConfig))

	// Start metrics server
	if config.ConfigFile.Metrics.Enabled {
		go func() {
			metrics := echo.New()

			if config.ConfigFile.Metrics.Pprof {
				pp := http.NewServeMux()
				pp.HandleFunc("/debug/pprof/", pprof.Index)
				pp.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
				pp.HandleFunc("/debug/pprof/profile", pprof.Profile)
				pp.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
				pp.HandleFunc("/debug/pprof/trace", pprof.Trace)
				metrics.GET("/debug/pprof/*", echo.WrapHandler(pp))
			}
			metrics.GET("/metrics", echoprometheus.NewHandler())
			metrics.HideBanner = true
			metrics.HidePort = true
			if err := metrics.Start(config.ConfigFile.Metrics.Host); err != nil && err != http.ErrServerClosed {
				metrics.Logger.Fatal(err)
			}
		}()
	}

	// Start proxy
	go func() {
		if err := e.Start(config.ConfigFile.Host); err != nil && err != http.ErrServerClosed {
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
