package main

import (
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	middlewares "github.com/marigold-dev/tzproxy/middlawares"
	utils "github.com/marigold-dev/tzproxy/utils"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func main() {
	config := utils.NewConfig()
	e := echo.New()

	store := memory.NewStore()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	if config.BlocklistEnable {
		e.Use(middlewares.BlockIP(config))
	}
	if config.RateEnable {
		e.Use(middlewares.RateLimit(store, *config.Rate))
	}

	url, err := url.Parse(config.TezosHost)
	if err != nil {
		e.Logger.Fatal(err)
	}

	targets := []*middleware.ProxyTarget{{URL: url}}
	balance := middleware.NewRoundRobinBalancer(targets)

	e.Use(middleware.Proxy(balance))
	e.Logger.Fatal(e.Start(config.Host))
}
