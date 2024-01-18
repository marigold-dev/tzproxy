package middlewares

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/marigold-dev/tzproxy/config"
)

func AllowRoutes(config *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if !config.ConfigFile.AllowRoutes.Enabled {
				return next(c)
			}

			r := c.Request()

			path := r.URL.Path

			if r.Method == http.MethodOptions {
				return next(c)
			}

			regexRoutesByMethod, has := config.AllowRoutesRegex[r.Method]
			if !has {
				msg := fmt.Sprintf("You don't have access %s route", path)
				return c.JSON(http.StatusForbidden, echo.Map{
					"success": false,
					"message": msg,
				})
			}

			for _, regex := range regexRoutesByMethod {
				if regex.MatchString(path) {
					return next(c)
				}
			}

			msg := fmt.Sprintf("You don't have access %s route", path)
			return c.JSON(http.StatusForbidden, echo.Map{
				"success": false,
				"message": msg,
			})

		}
	}
}
