package middlewares

import (
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/marigold-dev/tzproxy/utils"
)

func Retry(config *utils.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if config.ConfigFile.TezosHostRetry == "" {
				return next(c)
			}
			err = next(c)

			status := strconv.Itoa(c.Response().Status)
			if !strings.HasPrefix(status, "2") {
				c.Set("retry", true)
				return next(c)
			}

			c.Set("retry", false)
			return err
		}
	}
}
