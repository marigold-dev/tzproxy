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
	"github.com/spf13/viper"
	"github.com/ulule/limiter/v3"
)

type Config struct {
	ConfigFile               *ConfigFile
	Rate                     *limiter.Rate
	CacheDisabledRoutesRegex []*regexp.Regexp
	BlockRoutesRegex         []*regexp.Regexp
	CacheStorage             *freecache.Cache
	CacheTTL                 time.Duration
	RequestLoggerConfig      *middleware.RequestLoggerConfig
	ProxyConfig              *middleware.ProxyConfig
}

type ConfigFile struct {
	Host                string          `mapstructure:"host"`
	MetricsHost         string          `mapstructure:"metrics_host"`
	TezosHost           string          `mapstructure:"tezos_host"`
	RateLimitEnabled    bool            `mapstructure:"rate_limit_enabled"`
	RateLimitMinutes    float64         `mapstructure:"rate_limit_minutes"`
	RateLimitMax        int             `mapstructure:"rate_limit_max"`
	BlockAddressEnabled bool            `mapstructure:"block_address_enabled"`
	BlockRoutesEnabled  bool            `mapstructure:"block_routes_enabled"`
	CORSEnabled         bool            `mapstructure:"cors_enabled"`
	CacheEnabled        bool            `mapstructure:"cache_enabled"`
	CacheTTL            int             `mapstructure:"cache_ttl"`
	PprofEnabled        bool            `mapstructure:"pprof_enabled"`
	GzipEnabled         bool            `mapstructure:"gzip_enabled"`
	CacheDisabledRoutes []string        `mapstructure:"cache_disabled_routes"`
	CacheSizeMB         int             `mapstructure:"cache_size_mb"`
	BlockAddress        map[string]bool `mapstructure:"block_address"`
	BlockRoutes         []string        `mapstructure:"block_routes"`
	CGPercent           int             `mapstructure:"cg_percent"`
}

var defaultConfig = &ConfigFile{
	Host:               "0.0.0.0:8080",
	MetricsHost:        "0.0.0.0:9000",
	BlockRoutesEnabled: true,
	BlockRoutes: []string{
		"/injection/block", "/injection/protocol", "/network.*", "/workers.*",
		"/worker.*", "/stats.*", "/config", "/chains/main/blocks/.*/helpers/baking_rights",
		"/chains/main/blocks/.*/helpers/endorsing_rights",
		"/helpers/baking_rights", "/helpers/endorsing_rights",
		"/chains/main/blocks/.*/context/contracts(/?)$",
	},
	CacheDisabledRoutes: []string{
		"/monitor/.*",
	},
	CacheEnabled:        true,
	CacheSizeMB:         100,
	CacheTTL:            5,
	CGPercent:           20,
	GzipEnabled:         true,
	PprofEnabled:        false,
	RateLimitEnabled:    true,
	RateLimitMinutes:    1.0,
	RateLimitMax:        300,
	CORSEnabled:         true,
	BlockAddressEnabled: false,
	BlockAddress:        map[string]bool{},
}

func NewConfig() *Config {
	// Set default values for configuration
	viper.SetDefault("host", defaultConfig.Host)
	viper.SetDefault("metrics_host", defaultConfig.MetricsHost)
	viper.SetDefault("tezos_host", defaultConfig.TezosHost)
	viper.SetDefault("rate_limit_enabled", defaultConfig.RateLimitEnabled)
	viper.SetDefault("rate_limit_minutes", defaultConfig.RateLimitMinutes)
	viper.SetDefault("rate_limit_max", defaultConfig.RateLimitMax)
	viper.SetDefault("block_address_enabled", defaultConfig.BlockAddressEnabled)
	viper.SetDefault("block_routes_enabled", defaultConfig.BlockRoutesEnabled)
	viper.SetDefault("cors_enabled", defaultConfig.CORSEnabled)
	viper.SetDefault("cache_enabled", defaultConfig.CacheEnabled)
	viper.SetDefault("cache_ttl", defaultConfig.CacheTTL)
	viper.SetDefault("pprof_enabled", defaultConfig.PprofEnabled)
	viper.SetDefault("gzip_enabled", defaultConfig.GzipEnabled)
	viper.SetDefault("cache_disabled_routes", defaultConfig.CacheDisabledRoutes)
	viper.SetDefault("cache_size_mb", defaultConfig.CacheSizeMB)
	viper.SetDefault("block_address", defaultConfig.BlockAddress)
	viper.SetDefault("block_routes", defaultConfig.BlockRoutes)
	viper.SetDefault("cg_percent", defaultConfig.CGPercent)

	// Set the configuration file name and path
	viper.SetConfigName("tzproxy")
	viper.SetConfigType("yaml")
	// viper.AutomaticEnv()
	viper.AddConfigPath(".")

	// Read the configuration file
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// Unmarshal the configuration into the Config struct
	var configFile ConfigFile
	err = viper.Unmarshal(&configFile)
	if err != nil {
		log.Fatalf("Error unmarshaling config: %v", err)
	}

	url, err := url.Parse(configFile.TezosHost)
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
		ConfigFile: &configFile,
		Rate: &limiter.Rate{
			Period: time.Duration(GetEnvFloat("TZPROXY_RATE_LIMIT_MINUTES", 1.0)) * time.Minute,
			Limit:  int64(configFile.RateLimitMax),
		},
		CacheStorage: freecache.NewCache(1024 * 1024 * configFile.CacheSizeMB),
		CacheTTL:     time.Duration(configFile.CacheTTL) * (time.Second),
		ProxyConfig:  &proxyConfig,
	}

	for _, route := range config.ConfigFile.BlockRoutes {
		regex, err := regexp.Compile(route)
		if err != nil {
			panic(err)
		}
		config.BlockRoutesRegex = append(config.BlockRoutesRegex, regex)
	}

	for _, route := range config.ConfigFile.CacheDisabledRoutes {
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
