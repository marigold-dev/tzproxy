package config

import (
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func initViper() *ConfigFile {

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
	viper.SetDefault("dev_mode", false)
	viper.SetDefault("host", defaultConfig.Host)
	viper.SetDefault("tezos_host", defaultConfig.TezosHost)
	viper.SetDefault("tezos_host_retry", defaultConfig.TezosHostRetry)
	viper.SetDefault("redis.host", defaultConfig.Redis.Host)
	viper.SetDefault("redis.enabled", defaultConfig.Redis.Enabled)
	viper.SetDefault("load_balancer.ttl", defaultConfig.LoadBalancer.TTL)
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
	viper.SetDefault("deny_routes.enabled", defaultConfig.DenyRoutes.Enabled)
	viper.SetDefault("deny_routes.values", defaultConfig.DenyRoutes.Values)
	viper.SetDefault("metrics.enabled", defaultConfig.Metrics.Enabled)
	viper.SetDefault("metrics.pprof", defaultConfig.Metrics.Pprof)
	viper.SetDefault("metrics.host", defaultConfig.Metrics.Host)
	viper.SetDefault("cors.enabled", defaultConfig.CORS.Enabled)
	viper.SetDefault("gzip.enabled", defaultConfig.GZIP.Enabled)
	viper.SetDefault("gc.optimize_memory_store", defaultConfig.GC.OptimizeMemoryStore)
	viper.SetDefault("gc.percent", defaultConfig.GC.Percent)

	// Unmarshal the configuration into the Config struct
	var configFile ConfigFile
	err := viper.Unmarshal(&configFile)
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("unable to decode configuration")
	}
	viper.SafeWriteConfig()
	viper.WatchConfig()

	return &configFile
}
