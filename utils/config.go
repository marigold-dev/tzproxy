package util

import (
	"log"
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
	Host                 string
	TezosHost            string
	Rate                 *limiter.Rate
	RateEnabled          bool
	BlockAddressEnabled  bool
	BlockRoutesEnabled   bool
	CORSEnabled          bool
	CacheEnabled         bool
	DontCacheRoutes      []string
	DontCacheRoutesRegex []*regexp.Regexp
	BlockAddress         map[string]bool
	BlockRoutes          []string
	BlockRoutesRegex     []*regexp.Regexp
	Logger               zerolog.Logger
	Cache                *freecache.Cache
	CacheTTL             time.Duration
	RequestLoggerConfig  middleware.RequestLoggerConfig
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
	dontCacheRoutes := GetEnvSlice("TZPROXY_CACHE_ROUTES", []string{
		"/monitor/.*",
	})

	config := &Config{
		Host:        GetEnv("TZPROXY_HOST", "0.0.0.0:8080"),
		TezosHost:   GetEnv("TZPROXY_TEZOS_HOST", "http://127.0.0.1:8732"),
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
		DontCacheRoutes:     dontCacheRoutes,
		Cache:               freecache.NewCache(1024 * 1024 * 10),
		CacheTTL:            time.Duration(GetEnvInt("TZPROXY_CACHE_TTL", 5)) * (time.Second),
	}

	for _, route := range config.BlockRoutes {
		regex, err := regexp.Compile(route)
		if err != nil {
			panic(err)
		}
		config.BlockRoutesRegex = append(config.BlockRoutesRegex, regex)
	}

	for _, route := range config.DontCacheRoutes {
		regex, err := regexp.Compile(route)
		if err != nil {
			panic(err)
		}
		config.DontCacheRoutesRegex = append(config.DontCacheRoutesRegex, regex)
	}

	wr := diode.NewWriter(os.Stdout, 1000, 10*time.Millisecond, func(missed int) {
		log.Printf("Logger Dropped %d messages", missed)
	})
	config.Logger = zerolog.New(wr)
	config.RequestLoggerConfig = middleware.RequestLoggerConfig{
		LogLatency:      true,
		LogProtocol:     true,
		LogRemoteIP:     true,
		LogMethod:       true,
		LogURI:          true,
		LogRoutePath:    true,
		LogUserAgent:    true,
		LogStatus:       true,
		LogError:        true,
		LogResponseSize: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			config.Logger.Info().
				Str("ip", v.RemoteIP).
				Str("protocol", v.Protocol).
				Int("status", v.Status).
				Str("method", v.Method).
				Str("uri", v.URI).
				Str("route", v.RoutePath).
				Err(v.Error).
				Str("elapsed", v.Latency.String()).
				Str("user_agent", v.UserAgent).
				Int64("response_size", v.ResponseSize).
				Msg("request")

			return nil
		},
	}

	return config
}
