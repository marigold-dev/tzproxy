package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	echocache "github.com/fraidev/go-echo-cache"
	"github.com/labstack/echo/v4"
	"github.com/marigold-dev/tzproxy/config"
)

func Cache(config *config.Config) echo.MiddlewareFunc {
	return echocache.New(&echocache.Config{
		TTL: config.CacheTTL,
		Cache: func(r *http.Request) bool {
			if !config.ConfigFile.Cache.Enabled ||
				r.Method != http.MethodGet ||
				strings.Contains(r.URL.Path, "mempool") ||
				strings.Contains(r.URL.Path, "monitor") {
				return false
			}

			for _, regex := range config.CacheDisabledRoutesRegex {
				if regex.MatchString(r.URL.Path) {
					return false
				}
			}

			return true
		},
		GetKey: func(r *http.Request) []byte {
			base := r.Method + "|" + r.URL.Path + "|" + r.URL.Query().Encode()

			gzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")

			acceptHeader := r.Header.Get("Accept")
			if mediaIsUsed(acceptHeader, "application/bson") {
				base += "|bson"
			} else if mediaIsUsed(acceptHeader, "application/octet-stream") {
				base += "|octet"
			} else {
				base += "|json"
			}

			if gzip {
				base += "|gzip"
			}

			return []byte(base)
		},
	}, config.Store)
}

func mediaIsUsed(acceptHeader, media string) bool {
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
