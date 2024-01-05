package middlewares

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/marigold-dev/tzproxy/config"
)

func Gzip(config *config.Config) echo.MiddlewareFunc {
	return middleware.GzipWithConfig(middleware.GzipConfig{
		Skipper: func(c echo.Context) bool {
			if config.ConfigFile.GZIP.Enabled && strings.Contains(c.Request().Header.Get("Accept-Encoding"), "gzip") {
				return false
			}
			return true
		},
	})
}
