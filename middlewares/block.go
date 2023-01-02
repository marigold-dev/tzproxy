package middlewares

import (
	"net/http"

	"github.com/labstack/echo/v4"
	utils "github.com/marigold-dev/tzproxy/utils"
	"golang.org/x/exp/slices"
)

func BlockIP(config *utils.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if !config.BlockAddressEnable {
				return next(c)
			}
			if slices.Contains(config.BlockAddress, c.RealIP()) {
				return c.JSON(http.StatusForbidden, echo.Map{
					"success": false,
					"message": "Your IP is blocked",
				})
			}
			return next(c)
		}
	}
}
