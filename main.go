package main

import (
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	utils "github.com/marigold-dev/tzproxy/utils"
)

func main() {
	config := utils.NewConfig()

	e := echo.New()
	store := middleware.NewRateLimiterMemoryStoreWithConfig(*config.RateLimitConfig)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RateLimiter(store))

	url, err := url.Parse(config.TezosHost)
	if err != nil {
		e.Logger.Fatal(err)
	}

	targets := []*middleware.ProxyTarget{{URL: url}}
	balance := middleware.NewRoundRobinBalancer(targets)
	e.Use(middleware.Proxy(balance))

	e.Logger.Fatal(e.Start(config.Host))
}
