package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	middlewares "github.com/marigold-dev/tzproxy/middlawares"
	utils "github.com/marigold-dev/tzproxy/utils"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"net/url"
)

func main() {
	config := utils.NewConfig()
	store := memory.NewStore()

	e := echo.New()

	e.Use(middleware.Logger())
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
