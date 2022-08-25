package util

import (
	"time"

	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

type Config struct {
	Host            string
	TezosHost       string
	RateLimitConfig *middleware.RateLimiterMemoryStoreConfig
}

func NewConfig() *Config {
	return &Config{
		Host:      GetEnv("HOST", "localhost:8080"),
		TezosHost: GetEnv("TEZOS_HOST", "http://127.0.0.1:8732"),
		RateLimitConfig: &middleware.RateLimiterMemoryStoreConfig{
			Rate:      rate.Limit(GetEnvFloat("RATE_LIMIT_REQUESTS_PER_SECOND", 5)),
			Burst:     GetEnvInt("RATE_LIMIT_BURST", 5),
			ExpiresIn: time.Duration(GetEnvInt("RATE_LIMIT_EXPIRES_IN", 180)) * time.Second,
		},
	}
}
