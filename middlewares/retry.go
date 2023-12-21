package middlewares

import (
	"net/http"

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

			status := c.Response().Status
			if status == http.StatusNotFound || status == http.StatusForbidden {
				c.Set("retry", status)
				return next(c)
			}

			c.Set("retry", nil)

			return err
		}
	}
}
