package middlewares

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/marigold-dev/tzproxy/utils"
)

func Denylist(config *utils.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if !config.ConfigFile.DenyList.Enabled {
				return next(c)
			}

			value, has := config.DenyListTable[c.RealIP()]
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
