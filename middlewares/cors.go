package middlewares

import (
	"github.com/labstack/echo/v4"
	utils "github.com/marigold-dev/tzproxy/utils"
)

func CORS(config *utils.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.CORSEnabled {
				allowOriginHeader := c.Request().Header.Get("Access-Control-Allow-Origin")
				if allowOriginHeader == "" {
					c.Request().Header.Set("Access-Control-Allow-Origin", "*")
				}
			}
			return next(c)
		}
	}
}
