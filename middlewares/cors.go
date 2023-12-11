package middlewares

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/marigold-dev/tzproxy/utils"
)

func CORS(config *utils.Config) echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		Skipper: func(c echo.Context) bool {
			return !config.ConfigFile.CORS.Enabled
		},
		AllowHeaders: []string{"*"},
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"*"},
	})
}
