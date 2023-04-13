package middlewares

import (
	"net/http"

	cache "github.com/gitsight/go-echo-cache"
	"github.com/labstack/echo/v4"
	utils "github.com/marigold-dev/tzproxy/utils"
)

func Cache(config *utils.Config) echo.MiddlewareFunc {
	return cache.New(&cache.Config{
		TTL: config.CacheTTL,
		Cache: func(r *http.Request) bool {
			path := r.URL.Path
			for _, regex := range config.DontCacheRoutesRegex {
				if regex.MatchString(path) {
					return false
				}
			}
			return true
		},
	}, config.Cache)
}
