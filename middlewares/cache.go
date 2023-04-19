package middlewares

import (
	"net/http"

	cache "github.com/fraidev/go-echo-cache"
	"github.com/labstack/echo/v4"
	utils "github.com/marigold-dev/tzproxy/utils"
)

func Cache(config *utils.Config) echo.MiddlewareFunc {
	return cache.New(&cache.Config{
		TTL: config.CacheTTL,
		Cache: func(r *http.Request) bool {
			if !config.CacheEnable || r.Method != http.MethodGet {
				return false
			}

			for _, regex := range config.DontCacheRoutesRegex {
				if regex.MatchString(r.URL.Path) {
					return false
				}
			}
			return true
		},
	}, config.Cache)
}
