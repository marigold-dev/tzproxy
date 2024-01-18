package middlewares

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/marigold-dev/tzproxy/config"
)

func DenyIPs(config *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if !config.ConfigFile.DenyIPs.Enabled {
				return next(c)
			}

			value, has := config.DenyIPsTable[c.RealIP()]
			if value && has {
				return c.JSON(http.StatusForbidden, echo.Map{
					"success": false,
					"message": "Your IP is blocked",
				})
			}
			return next(c)
		}
	}
}
