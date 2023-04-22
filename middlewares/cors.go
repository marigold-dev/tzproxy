package middlewares

import (
	"github.com/labstack/echo/v4"
	utils "github.com/marigold-dev/tzproxy/utils"
)

func CORS(config *utils.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.CORSEnabled {
				resHeader := c.Response().Header()
				allowOriginHeader := resHeader.Get(echo.HeaderAccessControlAllowOrigin)
				if allowOriginHeader == "" {
					resHeader.Set(echo.HeaderAccessControlAllowOrigin, "*")
				}
			}
			return next(c)
		}
	}
}
