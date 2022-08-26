package util

import (
	"time"

	"github.com/ulule/limiter/v3"
)

type Config struct {
	Host            string
	TezosHost       string
	RateEnable      bool
	Rate            *limiter.Rate
	BlocklistEnable bool
	Blocklist       []string
	RateLimit       bool
}

func NewConfig() *Config {
	return &Config{
		Host:       GetEnv("HOST", "0.0.0.0:8080"),
		TezosHost:  GetEnv("TEZOS_HOST", "http://127.0.0.1:8732"),
		RateEnable: GetEnvBool("RATE_LIMIT_ENABLE", true),
		Rate: &limiter.Rate{
			Period: time.Duration(GetEnvFloat("RATE_LIMIT_MINUTES", 1.0)) * time.Minute,
			Limit:  int64(GetEnvInt("RATE_LIMIT_MAX", 300)),
		},
		BlocklistEnable: GetEnvBool("BLOCKLIST_ENABLE", true),
		Blocklist:       GetEnvSlice("BLOCKLIST", []string{}),
	}
}
