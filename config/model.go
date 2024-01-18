package config

import (
	"regexp"
	"time"

	echocache "github.com/fraidev/go-echo-cache"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/ulule/limiter/v3"
)

type Config struct {
	Level                    uint
	HashBlock                string
	ConfigFile               *ConfigFile
	Rate                     *limiter.Rate
	DenyIPsTable             map[string]bool
	CacheDisabledRoutesRegex map[string][]*regexp.Regexp
	DenyRoutesRegex          map[string][]*regexp.Regexp
	AllowRoutesRegex         map[string][]*regexp.Regexp
	Store                    echocache.Cache
	CacheTTL                 time.Duration
	RequestLoggerConfig      *middleware.RequestLoggerConfig
	ProxyConfig              *middleware.ProxyConfig
	Redis                    *redis.Client
	Logger                   zerolog.Logger
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

type DenyIPs struct {
	Enabled bool     `mapstructure:"enabled"`
	Values  []string `mapstructure:"values"`
}

type AllowRoutes struct {
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
	OptimizeMemoryStore bool `mapstructure:"optimize_memory_store"`
	Percent             int  `mapstructure:"percent"`
}

type CORS struct {
	Enabled bool `mapstructure:"enabled"`
}

type GZIP struct {
	Enabled bool `mapstructure:"enabled"`
}

type Redis struct {
	Host    string `mapstructure:"host"`
	Enabled bool   `mapstructure:"enabled"`
}

type LoadBalancer struct {
	TTL int `mapstructure:"ttl"`
}

type ConfigFile struct {
	DevMode        bool         `mapstructure:"dev_mode"`
	LoadBalancer   LoadBalancer `mapstructure:"load_balancer"`
	Redis          Redis        `mapstructure:"redis"`
	Logger         Logger       `mapstructure:"logger"`
	RateLimit      RateLimit    `mapstructure:"rate_limit"`
	Cache          Cache        `mapstructure:"cache"`
	DenyIPs        DenyIPs      `mapstructure:"deny_ips"`
	DenyRoutes     DenyRoutes   `mapstructure:"deny_routes"`
	AllowRoutes    AllowRoutes  `mapstructure:"allow_routes"`
	Metrics        Metrics      `mapstructure:"metrics"`
	GC             GC           `mapstructure:"gc"`
	CORS           CORS         `mapstructure:"cors"`
	GZIP           GZIP         `mapstructure:"gzip"`
	Host           string       `mapstructure:"host"`
	TezosHost      []string     `mapstructure:"tezos_host"`
	TezosHostRetry string       `mapstructure:"tezos_host_retry"`
}
