package middlewares

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	utils "github.com/marigold-dev/tzproxy/utils"
)

func Gzip(config *utils.Config) echo.MiddlewareFunc {
	return middleware.GzipWithConfig(middleware.GzipConfig{
		Skipper: func(c echo.Context) bool {
			if config.GzipEnabled && strings.Contains(c.Request().Header.Get("Accept-Encoding"), "gzip") {
				return false
			}
			return true
		},
	})
}
