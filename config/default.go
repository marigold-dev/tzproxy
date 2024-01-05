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
