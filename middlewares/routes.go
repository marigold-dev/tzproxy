package middlewares

import (
	"fmt"
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
					msg := fmt.Sprintf("You don't have accest %s route", c.Request().URL.Path)
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
