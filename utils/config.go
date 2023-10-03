package util

import (
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
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
	DenyListTable            map[string]bool
	Rate                     *limiter.Rate
	CacheDisabledRoutesRegex []*regexp.Regexp
	BlockRoutesRegex         []*regexp.Regexp
	CacheStorage             *freecache.Cache
	CacheTTL                 time.Duration
	RequestLoggerConfig      *middleware.RequestLoggerConfig
	ProxyConfig              *middleware.ProxyConfig
}

type Logger struct {
	BunchSize           int `mapstructure:"bunch_size"`
	PoolIntervalSeconds int `mapstructure:"pool_interval_seconds"`
}

type RateLimit struct {
	Enabled bool    `mapstructure:"enabled"`
	Minutes float64 `mapstructure:"minutes"`
	Max     int     `mapstructure:"max"`
}

type Cache struct {
	Enabled        bool     `mapstructure:"enabled"`
	TTL            int      `mapstructure:"ttl"`
	DisabledRoutes []string `mapstructure:"disabled_routes"`
	SizeMB         int      `mapstructure:"size_mb"`
}

type DenyList struct {
	Enabled bool     `mapstructure:"enabled"`
	Values  []string `mapstructure:"values"`
}

type DenyRoutes struct {
	Enabled bool     `mapstructure:"enabled"`
	Values  []string `mapstructure:"values"`
}

type Metrics struct {
	Host    string `mapstructure:"host"`
	Enabled bool   `mapstructure:"enabled"`
	Pprof   bool   `mapstructure:"pprof"`
}

type GC struct {
	Percent int `mapstructure:"percent"`
}

type CORS struct {
	Enabled bool `mapstructure:"enabled"`
}

type GZIP struct {
	Enabled bool `mapstructure:"enabled"`
}

type ConfigFile struct {
	Logger          Logger     `mapstructure:"logger"`
	RateLimit       RateLimit  `mapstructure:"rate_limit"`
	Cache           Cache      `mapstructure:"cache"`
	DenyList        DenyList   `mapstructure:"deny_list"`
	DenyRoutes      DenyRoutes `mapstructure:"deny_routes"`
	Metrics         Metrics    `mapstructure:"metrics"`
	GC              GC         `mapstructure:"gc"`
	CORS            CORS       `mapstructure:"cors"`
	GZIP            GZIP       `mapstructure:"gzip"`
	Host            string     `mapstructure:"host"`
	TezosHost       []string   `mapstructure:"tezos_host"`
	LoadBalancerTTL int        `mapstructure:"tezos_host_load_balancer_ttl"`
}

var defaultConfig = &ConfigFile{
	Host:            "0.0.0.0:8080",
	TezosHost:       []string{"127.0.0.1:8732"},
	LoadBalancerTTL: 600,
	Logger: Logger{
		BunchSize:           1000,
		PoolIntervalSeconds: 10,
	},
	RateLimit: RateLimit{
		Enabled: false,
		Minutes: 1,
		Max:     300,
	},
	Cache: Cache{
		Enabled: true,
		TTL:     5,
		DisabledRoutes: []string{
			"/monitor/.*",
			"/chains/.*/mempool",
			"/chains/.*/blocks.*head",
		},
		SizeMB: 100,
	},
	DenyList: DenyList{
		Enabled: false,
		Values:  []string{},
	},
	DenyRoutes: DenyRoutes{
		Enabled: true,
		Values: []string{
			"/injection/block", "/injection/protocol", "/network.*", "/workers.*",
			"/worker.*", "/stats.*", "/config", "/chains/.*/blocks/.*/helpers/baking_rights",
			"/chains/.*/blocks/.*/helpers/endorsing_rights",
			"/helpers/baking_rights", "/helpers/endorsing_rights",
			"/chains/.*/blocks/.*/context/contracts(/?)$",
			"/chains/.*/blocks/.*/context/raw/bytes",
		},
	},
	Metrics: Metrics{
		Host:    "0.0.0.0:9000",
		Enabled: true,
		Pprof:   false,
	},
	GC: GC{
		Percent: 20,
	},
	GZIP: GZIP{
		Enabled: true,
	},
	CORS: CORS{
		Enabled: true,
	},
}

func NewConfig() *Config {
	// Set the configuration file name and path
	viper.SetConfigName("tzproxy")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// Read the configuration file
	viper.ReadInConfig()

	// Set the environment variables prefix
	viper.SetEnvPrefix("TZPROXY")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()
	viper.SetTypeByDefaultValue(true)

	// Set default values for configuration
	viper.SetDefault("host", defaultConfig.Host)
	viper.SetDefault("tezos_host", defaultConfig.TezosHost)
	viper.SetDefault("load_balancer_ttl", defaultConfig.LoadBalancerTTL)
	viper.SetDefault("logger.bunch_size", defaultConfig.Logger.BunchSize)
	viper.SetDefault("logger.pool_interval_seconds", defaultConfig.Logger.PoolIntervalSeconds)
	viper.SetDefault("cache.enabled", defaultConfig.Cache.Enabled)
	viper.SetDefault("cache.ttl", defaultConfig.Cache.TTL)
	viper.SetDefault("cache.disabled_routes", defaultConfig.Cache.DisabledRoutes)
	viper.SetDefault("cache.size_mb", defaultConfig.Cache.SizeMB)
	viper.SetDefault("rate_limit.enabled", defaultConfig.RateLimit.Enabled)
	viper.SetDefault("rate_limit.minutes", defaultConfig.RateLimit.Minutes)
	viper.SetDefault("rate_limit.max", defaultConfig.RateLimit.Max)
	viper.SetDefault("deny_list.enabled", defaultConfig.DenyList.Enabled)
	viper.SetDefault("deny_list.values", defaultConfig.DenyList.Values)
	viper.SetDefault("metrics.enabled", defaultConfig.Metrics.Enabled)
	viper.SetDefault("metrics.pprof", defaultConfig.Metrics.Pprof)
	viper.SetDefault("metrics.host", defaultConfig.Metrics.Host)
	viper.SetDefault("deny_routes.enabled", defaultConfig.DenyRoutes.Enabled)
	viper.SetDefault("deny_routes.values", defaultConfig.DenyRoutes.Values)
	viper.SetDefault("cors.enabled", defaultConfig.CORS.Enabled)
	viper.SetDefault("gzip.enabled", defaultConfig.GZIP.Enabled)
	viper.SetDefault("gc.percent", defaultConfig.GC.Percent)

	// Unmarshal the configuration into the Config struct
	var configFile ConfigFile
	err := viper.Unmarshal(&configFile)
	if err != nil {
		log.Fatalf("Error unmarshaling config: %v", err)
	}

	var targets = []*middleware.ProxyTarget{}

	for _, host := range configFile.TezosHost {
		tezosHost := host

		if !strings.Contains(host, "http") {
			tezosHost = "http://" + host
		}

		url, err := url.ParseRequestURI(tezosHost)
		if err != nil {
			log.Fatal(err)
		}
		targets = append(targets, &middleware.ProxyTarget{URL: url})
	}
	balancer := NewSameNodeBalancer(targets, configFile.LoadBalancerTTL)

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
		CacheStorage: freecache.NewCache(1024 * 1024 * configFile.Cache.SizeMB),
		CacheTTL:     time.Duration(configFile.Cache.TTL) * (time.Second),
		ProxyConfig:  &proxyConfig,
	}

	for _, route := range config.ConfigFile.DenyRoutes.Values {
		regex, err := regexp.Compile(route)
		if err != nil {
			panic(err)
		}
		config.BlockRoutesRegex = append(config.BlockRoutesRegex, regex)
	}

	for _, route := range config.ConfigFile.Cache.DisabledRoutes {
		regex, err := regexp.Compile(route)
		if err != nil {
			panic(err)
		}
		config.CacheDisabledRoutesRegex = append(config.CacheDisabledRoutesRegex, regex)
	}

	bunchWriter := diode.NewWriter(
		os.Stdout,
		config.ConfigFile.Logger.BunchSize,
		time.Duration(configFile.Logger.PoolIntervalSeconds)*time.Second, func(missed int) {
			log.Printf("Logger Dropped %d messages", missed)
		})
	zl := zerolog.New(bunchWriter)
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
					Str("referer", v.Referer).
					Int64("response_size", v.ResponseSize).
					Msg("request")
			}

			return nil
		},
	}

	viper.SafeWriteConfig()
	viper.WatchConfig()

	return config
}
