package config

import (
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/coocood/freecache"
	echocache "github.com/fraidev/go-echo-cache"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/marigold-dev/tzproxy/balancers"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
	"github.com/rs/zerolog/log"
	"github.com/ulule/limiter/v3"
)

func NewConfig() *Config {
	configFile := initViper()

	if configFile.GC.OptimizeMemoryStore {
		if !configFile.Redis.Enabled {
			configFile.GC.Percent = 20
		}
	}

	var targets = []*middleware.ProxyTarget{}
	var retryTarget *middleware.ProxyTarget = nil
	if configFile.TezosHostRetry != "" {
		retryTarget = hostToTarget(configFile.TezosHostRetry)
	}
	for _, host := range configFile.TezosHost {
		targets = append(targets, hostToTarget(host))
	}

	var redisClient *redis.Client
	if configFile.Redis.Enabled {
		redisClient = redis.NewClient(&redis.Options{
			Addr: configFile.Redis.Host,
		})
	}
	store := buildStore(configFile, redisClient)
	balancer := balancers.NewIPHashBalancer(targets, retryTarget, configFile.LoadBalancer.TTL, store)
	logger := buildLogger(configFile.DevMode)
	proxyConfig := middleware.ProxyConfig{
		Skipper:    middleware.DefaultSkipper,
		ContextKey: "target",
		RetryCount: 0,
		Balancer:   balancer,
		RetryFilter: func(c echo.Context, err error) bool {
			if httpErr, ok := err.(*echo.HTTPError); ok {
				if httpErr.Code == http.StatusBadGateway || httpErr.Code == http.StatusNotFound || httpErr.Code == http.StatusGone {
					return true
				}
			}

			return false
		},
		ErrorHandler: func(c echo.Context, err error) error {
			logger.Error().Err(err).Msg("proxy error")
			c.Logger().Error(err)
			return err
		},
	}

	config := &Config{
		ConfigFile: configFile,
		DenyListTable: func() map[string]bool {
			table := make(map[string]bool)
			for _, ip := range configFile.DenyList.Values {
				table[ip] = true
			}
			return table
		}(),
		Rate: &limiter.Rate{
			Period: time.Duration(configFile.RateLimit.Minutes) * time.Minute,
			Limit:  int64(configFile.RateLimit.Max),
		},
		Store:       store,
		CacheTTL:    time.Duration(configFile.Cache.TTL) * (time.Second),
		ProxyConfig: &proxyConfig,
		Redis:       redisClient,
	}

	for _, route := range config.ConfigFile.DenyRoutes.Values {
		regex, err := regexp.Compile(route)
		if err != nil {
			log.Fatal().Err(err).Msg("unable to compile regex")
		}
		config.BlockRoutesRegex = append(config.BlockRoutesRegex, regex)
	}

	for _, route := range config.ConfigFile.Cache.DisabledRoutes {
		regex, err := regexp.Compile(route)
		if err != nil {
			log.Fatal().Err(err).Msg("unable to compile regex")
		}
		config.CacheDisabledRoutesRegex = append(config.CacheDisabledRoutesRegex, regex)
	}

	config.Logger = logger

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
		LogReferer:      true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error != nil {
				config.Logger.Error().
					Err(v.Error).
					Str("ip", v.RemoteIP).
					Int("status", v.Status).
					Str("method", v.Method).
					Str("uri", v.URI).
					Str("user_agent", v.UserAgent).
					Msg("request error")
				return v.Error
			}

			config.Logger.Info().
				Str("ip", v.RemoteIP).
				Str("protocol", v.Protocol).
				Int("status", v.Status).
				Str("method", v.Method).
				Str("uri", v.URI).
				Int64("elapsed", int64(v.Latency)).
				Str("user_agent", v.UserAgent).
				Str("referer", v.Referer).
				Int64("response_size", v.ResponseSize).
				Msg("request")

			return nil
		},
	}

	return config
}

func buildStore(cf *ConfigFile, redis *redis.Client) echocache.Cache {
	if cf.Redis.Enabled {
		redisStore := echocache.NewRedisCache(redis)
		return &redisStore
	}

	freeCache := freecache.NewCache(cf.Cache.SizeMB * 1024 * 1024)
	memoryStore := echocache.NewMemoryCache(freeCache)
	return &memoryStore
}

func buildLogger(devMode bool) zerolog.Logger {
	if !devMode {
		bunchWriter := diode.NewWriter(
			os.Stdout,
			1000,
			time.Second, func(missed int) {
				log.Printf("Logger Dropped %d messages", missed)
			})

		return zerolog.New(bunchWriter)
	}

	return log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func hostToTarget(host string) *middleware.ProxyTarget {
	hostWithScheme := host
	if !strings.Contains(host, "http") {
		hostWithScheme = "http://" + host
	}
	targetURL, err := url.ParseRequestURI(hostWithScheme)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to parse host")
	}

	return &middleware.ProxyTarget{URL: targetURL}
}
