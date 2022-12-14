package util

import (
	"os"
	"regexp"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/ulule/limiter/v3"
)

type Config struct {
	Host                string
	TezosHost           string
	Rate                *limiter.Rate
	RateEnable          bool
	BlockAddressEnable  bool
	BlockRoutesEnable   bool
	BlockAddress        []string
	BlockRoutes         []string
	BlockRoutesRegex    []*regexp.Regexp
	Logger              zerolog.Logger
	RequestLoggerConfig middleware.RequestLoggerConfig
}

func NewConfig() *Config {
	configs := &Config{
		Host:       GetEnv("HOST", "0.0.0.0:8080"),
		TezosHost:  GetEnv("TEZOS_HOST", "http://127.0.0.1:8732"),
		RateEnable: GetEnvBool("RATE_LIMIT_ENABLE", true),
		Rate: &limiter.Rate{
			Period: time.Duration(GetEnvFloat("RATE_LIMIT_MINUTES", 1.0)) * time.Minute,
			Limit:  int64(GetEnvInt("RATE_LIMIT_MAX", 300)),
		},
		BlockAddressEnable: GetEnvBool("BLOCK_ADDRESSES_ENABLE", true),
		BlockAddress:       GetEnvSlice("BLOCK_ADDRESSES", []string{}),
		BlockRoutesEnable:  GetEnvBool("BLOCK_ROUTES_ENABLE", true),
		BlockRoutes: GetEnvSlice("BLOCK_ROUTES", []string{
			"/injection/block", "/injection/protocol", "/network.*", "/workers.*",
			"/worker.*", "/stats.*", "/config", "/chains/main/blocks/.*/helpers/baking_rights",
			"/chains/main/blocks/.*/helpers/endorsing_rights",
			"/helpers/baking_rights", "/helpers/endorsing_rights",
		}),
	}

	for _, route := range configs.BlockRoutes {
		regex, err := regexp.Compile(route)
		if err != nil {
			panic(err)
		}
		configs.BlockRoutesRegex = append(configs.BlockRoutesRegex, regex)
	}

	configs.Logger = zerolog.New(os.Stdout)
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
