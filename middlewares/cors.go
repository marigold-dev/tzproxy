package middlewares

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	utils "github.com/marigold-dev/tzproxy/utils"
)

func CORS(config *utils.Config) echo.MiddlewareFunc {
	if config.CORSEnable {
		return middleware.CORS()
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	}
}
