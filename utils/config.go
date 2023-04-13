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
	RateEnable           bool
	BlockAddressEnable   bool
	BlockRoutesEnable    bool
	CacheEnable          bool
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

	blockAddress := GetEnvSet("BLOCK_ADDRESSES", map[string]bool{})
	blockRoutes := GetEnvSlice("BLOCK_ROUTES", []string{
		"/injection/block", "/injection/protocol", "/network.*", "/workers.*",
		"/worker.*", "/stats.*", "/config", "/chains/main/blocks/.*/helpers/baking_rights",
		"/chains/main/blocks/.*/helpers/endorsing_rights",
		"/helpers/baking_rights", "/helpers/endorsing_rights",
		"(.*?)context/contracts",
	})
	dontCacheRoutes := GetEnvSlice("CACHE_ROUTES", []string{
		"/chains/main/blocks/.*/context/contracts",
		"/monitor/.*",
	})

	configs := &Config{
		Host:       GetEnv("HOST", "0.0.0.0:8080"),
		TezosHost:  GetEnv("TEZOS_HOST", "http://127.0.0.1:8732"),
		RateEnable: GetEnvBool("RATE_LIMIT_ENABLE", true),
		Rate: &limiter.Rate{
			Period: time.Duration(GetEnvFloat("RATE_LIMIT_MINUTES", 1.0)) * time.Minute,
			Limit:  int64(GetEnvInt("RATE_LIMIT_MAX", 300)),
		},
		BlockAddress:       blockAddress,
		BlockAddressEnable: GetEnvBool("BLOCK_ADDRESSES_ENABLE", len(blockAddress) > 0),
		BlockRoutesEnable:  GetEnvBool("BLOCK_ROUTES_ENABLE", len(blockRoutes) > 0),
		BlockRoutes:        blockRoutes,
		CacheEnable:        GetEnvBool("CACHE_ENABLE", true),
		DontCacheRoutes:    dontCacheRoutes,
		Cache:              freecache.NewCache(1024 * 1024 * 10),
		CacheTTL:           time.Duration(GetEnvInt("CACHE_TTL", 5)) * (time.Second),
	}

	for _, route := range configs.BlockRoutes {
		regex, err := regexp.Compile(route)
		if err != nil {
			panic(err)
		}
		configs.BlockRoutesRegex = append(configs.BlockRoutesRegex, regex)
	}

	for _, route := range configs.DontCacheRoutes {
		regex, err := regexp.Compile(route)
		if err != nil {
			panic(err)
		}
		configs.DontCacheRoutesRegex = append(configs.DontCacheRoutesRegex, regex)
	}

	wr := diode.NewWriter(os.Stdout, 1000, 10*time.Millisecond, func(missed int) {
		log.Printf("Logger Dropped %d messages", missed)
	})
	configs.Logger = zerolog.New(wr)
	configs.RequestLoggerConfig = middleware.RequestLoggerConfig{
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
			configs.Logger.Info().
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

	return configs
}
