package config

var defaultConfig = &ConfigFile{
	DevMode:        false,
	Host:           "0.0.0.0:8080",
	TezosHost:      []string{"127.0.0.1:8732"},
	TezosHostRetry: "",
	Redis: Redis{
		Host:    "",
		Enabled: false,
	},
	LoadBalancer: LoadBalancer{
		TTL: 600,
	},
	Logger: Logger{
		BunchSize:           1000,
		PoolIntervalSeconds: 1,
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
			"GET/monitor/.*",
			"GET/chains/.*/mempool",
			"GET/chains/.*/blocks.*head",
		},
		SizeMB: 100,
	},
	DenyIPs: DenyIPs{
		Enabled: false,
		Values:  []string{},
	},
	AllowRoutes: AllowRoutes{
		Enabled: true,
		Values: []string{
			"GET/chains/.*/blocks",
			"GET/chains/.*/chain_id", "GET/chains.*/checkpoint",
			"GET/chains/.*/invalid_blocks", "GET/chains.*/invalid_blocks.*",
			"GET/chains/.*/is_bootstrapped", "GET/chains.*/mempool/filter",
			"GET/chains/.*/mempool/monitor_operations",
			"GET/chains/.*/mempool/pending_operations",
			"GET/config/network/user_activated_protocol_overrides",
			"GET/config/network/user_activated_upgrades",
			"GET/config/network/dal", "GET/describe.*", "GET/errors",
			"GET/monitor.*", "GET/network/greylist/ips",
			"GET/network/greylist/peers", "GET/network/self",
			"GET/network/stat", "GET/network/version", "GET/network/versions",
			"GET/protocols", "GET/protocols.*", "GET/protocols.*/environment",
			"GET/version",
			"POST/chains/.*/blocks/.*/helpers",
			"POST/chains/.*/blocks/.*/script",
			"POST/chains/.*/blocks/.*/context/contracts.*/big_map_get",
			"POST/injection/operation",
		},
	},
	DenyRoutes: DenyRoutes{
		Enabled: true,
		Values: []string{
			"GET/workers.*",
			"GET/worker.*",
			"GET/stats.*",
			"GET/chains/.*/blocks/.*/helpers/baking_rights",
			"GET/chains/.*/blocks/.*/helpers/endorsing_rights",
			"GET/helpers/baking_rights",
			"GET/helpers/endorsing_rights",
			"GET/chains/.*/blocks/.*/context/contracts(/?)$",
			"GET/chains/.*/blocks/.*/context/raw/bytes",
			"POST/injection/block",
			"POST/injection/protocol",
		},
	},
	Metrics: Metrics{
		Host:    "0.0.0.0:9000",
		Enabled: true,
		Pprof:   false,
	},
	GC: GC{
		OptimizeMemoryStore: true,
		Percent:             100,
	},
	GZIP: GZIP{
		Enabled: true,
	},
	CORS: CORS{
		Enabled: true,
	},
}
