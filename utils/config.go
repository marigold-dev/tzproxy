package util

import (
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/coocood/freecache"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
	"github.com/ulule/limiter/v3"
)

type Config struct {
	Host                     string
	MetricsHost              string
	TezosHost                string
	Rate                     *limiter.Rate
	RateEnabled              bool
	BlockAddressEnabled      bool
	BlockRoutesEnabled       bool
	CORSEnabled              bool
	CacheEnabled             bool
	PprofEnabled             bool
	GzipEnabled              bool
	CacheDisabledRoutes      []string
	CacheDisabledRoutesRegex []*regexp.Regexp
	BlockAddress             map[string]bool
	BlockRoutes              []string
	BlockRoutesRegex         []*regexp.Regexp
	Cache                    *freecache.Cache
	CacheTTL                 time.Duration
	CGPercent                int
	RequestLoggerConfig      *middleware.RequestLoggerConfig
	ProxyConfig              *middleware.ProxyConfig
}

func NewConfig() *Config {
	blockAddress := GetEnvSet("TZPROXY_BLOCK_ADDRESSES", map[string]bool{})
	blockRoutes := GetEnvSlice("TZPROXY_BLOCK_ROUTES", []string{
		"/injection/block", "/injection/protocol", "/network.*", "/workers.*",
		"/worker.*", "/stats.*", "/config", "/chains/main/blocks/.*/helpers/baking_rights",
		"/chains/main/blocks/.*/helpers/endorsing_rights",
		"/helpers/baking_rights", "/helpers/endorsing_rights",
		"/chains/main/blocks/.*/context/contracts(/?)$",
	})
	cacheDisabledRoutes := GetEnvSlice("TZPROXY_CACHE_DISABLED_ROUTES", []string{
		"/monitor/.*",
	})
	cacheSizeMB := GetEnvInt("TZPROXY_CACHE_SIZE_MB", 100)
	pprofEnabled := GetEnvBool("TZPROXY_ENABLE_PPROF", false)
	gzipEnabled := GetEnvBool("TZPROXY_ENABLE_GZIP", true)

	tezosHost := GetEnv("TZPROXY_TEZOS_HOST", "http://127.0.0.1:8732")
	url, err := url.Parse(tezosHost)
	if err != nil {
		log.Fatal(err)
	}
	balancer := middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{{URL: url}})

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   65 * time.Second,
			KeepAlive: 65 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          300,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	proxyConfig := middleware.ProxyConfig{
		Skipper:    middleware.DefaultSkipper,
		ContextKey: "target",
		Balancer:   balancer,
		Transport:  transport,
	}

	config := &Config{
		Host:        GetEnv("TZPROXY_HOST", "0.0.0.0:8080"),
		MetricsHost: GetEnv("TZPROXY_METRICS_HOST", "0.0.0.0:9000"),
		TezosHost:   tezosHost,
		RateEnabled: GetEnvBool("TZPROXY_RATE_LIMIT_ENABLED", true),
		Rate: &limiter.Rate{
			Period: time.Duration(GetEnvFloat("TZPROXY_RATE_LIMIT_MINUTES", 1.0)) * time.Minute,
			Limit:  int64(GetEnvInt("TZPROXY_RATE_LIMIT_MAX", 300)),
		},
		CORSEnabled:         GetEnvBool("TZPROXY_CORS_ENABLED", true),
		BlockAddress:        blockAddress,
		BlockAddressEnabled: GetEnvBool("TZPROXY_BLOCK_ADDRESSES_ENABLED", len(blockAddress) > 0),
		BlockRoutesEnabled:  GetEnvBool("TZPROXY_BLOCK_ROUTES_ENABLED", len(blockRoutes) > 0),
		BlockRoutes:         blockRoutes,
		CacheEnabled:        GetEnvBool("TZPROXY_CACHE_ENABLED", true),
		CacheDisabledRoutes: cacheDisabledRoutes,
		Cache:               freecache.NewCache(1024 * 1024 * cacheSizeMB),
		CacheTTL:            time.Duration(GetEnvInt("TZPROXY_CACHE_TTL", 5)) * (time.Second),
		CGPercent:           GetEnvInt("GO_GC", 20),
		ProxyConfig:         &proxyConfig,
		PprofEnabled:        pprofEnabled,
		GzipEnabled:         gzipEnabled,
	}

	for _, route := range config.BlockRoutes {
		regex, err := regexp.Compile(route)
		if err != nil {
			panic(err)
		}
		config.BlockRoutesRegex = append(config.BlockRoutesRegex, regex)
	}

	for _, route := range config.CacheDisabledRoutes {
		regex, err := regexp.Compile(route)
		if err != nil {
			panic(err)
		}
		config.CacheDisabledRoutesRegex = append(config.CacheDisabledRoutesRegex, regex)
	}

	wr := diode.NewWriter(os.Stdout, 1000, 1*time.Second, func(missed int) {
		log.Printf("Logger Dropped %d messages", missed)
	})
	zl := zerolog.New(wr)
	config.RequestLoggerConfig = &middleware.RequestLoggerConfig{
		LogLatency:      true,
		LogProtocol:     true,
		LogRemoteIP:     true,
		LogMethod:       true,
		LogURI:          true,
		LogUserAgent:    true,
		LogStatus:       true,
		LogError:        true,
		HandleError:     true,
		LogResponseSize: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error != nil {
				zl.Error().
					Err(v.Error).
					Str("ip", v.RemoteIP).
					Int("status", v.Status).
					Str("method", v.Method).
					Str("uri", v.URI).
					Str("user_agent", v.UserAgent).
					Msg("request error")
			} else {
				zl.Info().
					Str("ip", v.RemoteIP).
					Str("protocol", v.Protocol).
					Int("status", v.Status).
					Str("method", v.Method).
					Str("uri", v.URI).
					Int64("elapsed", int64(v.Latency)).
					Str("user_agent", v.UserAgent).
					Int64("response_size", v.ResponseSize).
					Msg("request")
			}

			return nil
		},
	}

	return config
}
