package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	cache "github.com/fraidev/go-echo-cache"
	"github.com/labstack/echo/v4"
	utils "github.com/marigold-dev/tzproxy/utils"
)

func Cache(config *utils.Config) echo.MiddlewareFunc {
	return cache.New(&cache.Config{
		TTL: config.CacheTTL,
		Cache: func(r *http.Request) bool {
			if !config.CacheEnabled || r.Method != http.MethodGet {
				return false
			}

			// Don't cache binary rpc calls
			acceptHeader := r.Header.Get("Accept")
			if mediaIsAccept(acceptHeader, "application/bson") {
				return false
			}
			if mediaIsAccept(acceptHeader, "application/octet-stream") {
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

func mediaIsAccept(acceptHeader, media string) bool {
	if strings.Contains(acceptHeader, media) {
		acceptValuesByQuallity := parseQValues(acceptHeader)
		mediaQValue, containsOctet := acceptValuesByQuallity[media]

		if containsOctet {
			allQValue, containsAll := acceptValuesByQuallity["*/*"]
			jsonQValue, containsJson := acceptValuesByQuallity["application/json"]
			if containsAll || containsJson {
				if mediaQValue > allQValue && mediaQValue > jsonQValue {
					return true
				}
			} else {
				return true
			}
		}
	}

	return false
}

func parseQValues(header string) map[string]float32 {
	qValues := make(map[string]float32)

	if header == "" {
		return qValues
	}

	for _, mediaRange := range strings.Split(header, ",") {
		parts := strings.Split(strings.TrimSpace(mediaRange), ";")
		mediaType := parts[0]
		qValue := float32(1)

		for _, param := range parts[1:] {
			if strings.HasPrefix(param, "q=") {
				q := strings.TrimPrefix(param, "q=")
				fmt.Sscanf(q, "%f", &qValue)
			}
		}

		qValues[mediaType] = qValue
	}

	return qValues
}
