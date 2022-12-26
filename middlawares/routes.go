package middlewares

import (
	"net/http"

	"github.com/labstack/echo/v4"
	utils "github.com/marigold-dev/tzproxy/utils"
)

func BlockRoutes(config *utils.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if !config.BlockRoutesEnable {
				return next(c)
			}

			for _, regex := range config.BlockRoutesRegex {
				if regex.MatchString(c.Request().URL.Path) {
					return c.JSON(http.StatusForbidden, echo.Map{
						"success": false,
						"message": "This route is blocked, please contact the infra@marigold.dev",
					})
				}
			}

			return next(c)
		}
	}
}
