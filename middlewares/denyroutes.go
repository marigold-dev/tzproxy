package middlewares

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/marigold-dev/tzproxy/config"
)

func DenyRoutes(config *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if !config.ConfigFile.DenyRoutes.Enabled {
				return next(c)
			}

			r := c.Request()

			path := r.URL.Path

			regexRoutesByMethod, has := config.DenyRoutesRegex[r.Method]
			if !has {
				return next(c)
			}

			for _, regex := range regexRoutesByMethod {
				if regex.MatchString(path) {
					msg := fmt.Sprintf("You don't have access %s route", path)
					return c.JSON(http.StatusForbidden, echo.Map{
						"success": false,
						"message": msg,
					})
				}
			}

			return next(c)
		}
	}
}
