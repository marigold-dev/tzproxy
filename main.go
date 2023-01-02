package main

import (
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	middlewares "github.com/marigold-dev/tzproxy/middlewares"
	utils "github.com/marigold-dev/tzproxy/utils"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func main() {
	config := utils.NewConfig()
	store := memory.NewStore()

	e := echo.New()

	e.Use(middleware.RequestLoggerWithConfig(config.RequestLoggerConfig))
	e.Use(middleware.Recover())
	e.Use(middlewares.BlockIP(config))
	e.Use(middlewares.RateLimit(store, config))
	e.Use(middlewares.BlockRoutes(config))

	url, err := url.Parse(config.TezosHost)
	if err != nil {
		e.Logger.Fatal(err)
	}

	targets := []*middleware.ProxyTarget{{URL: url}}
	balance := middleware.NewRoundRobinBalancer(targets)

	e.Use(middleware.Proxy(balance))
	e.Logger.Fatal(e.Start(config.Host))
}
