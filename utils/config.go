package util

import (
	"regexp"
	"time"

	"github.com/ulule/limiter/v3"
)

type Config struct {
	Host               string
	TezosHost          string
	Rate               *limiter.Rate
	RateEnable         bool
	BlockAddressEnable bool
	BlockRoutesEnable  bool
	BlockAddress       []string
	BlockRoutes        []string
	BlockRoutesRegex   []*regexp.Regexp
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

	return configs
}
