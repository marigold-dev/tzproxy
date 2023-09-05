# TzProxy

TzProxy is a reverse proxy specifically designed for Tezos nodes, offering a range of features that enhance node performance, security, and manageability.

## Features

- [x] Rate limit
- [x] Block IPs
- [x] Blocklist routes
- [x] Cache
- [x] CORS
- [x] GZIP
- [x] Metrics

## Congiguration

### Yaml File
Here a default `tzproxy.yaml` file:

```yaml
cache:
    disabled_routes:
        - /monitor/.*
    enabled: true
    size_mb: 100
    ttl: 5
cors:
    enabled: true
deny_list:
    enabled: false
    values: []
deny_routes:
    enabled: true
    values:
        - /injection/block
        - /injection/protocol
        - /network.*
        - /workers.*
        - /worker.*
        - /stats.*
        - /config
        - /chains/main/blocks/.*/helpers/baking_rights
        - /chains/main/blocks/.*/helpers/endorsing_rights
        - /helpers/baking_rights
        - /helpers/endorsing_rights
        - /chains/main/blocks/.*/context/contracts(/?)$
gc:
    percent: 20
gzip:
    enabled: true
host: 0.0.0.0:8080
logger:
    bunch_size: 1000
    pool_interval_seconds: 10
metrics:
    enabled: true
    host: 0.0.0.0:9000
    pprof: false
rate_limit:
    enabled: false
    max: 300
    minutes: 1
tezos_host: 127.0.0.1:8732
```

### Environment Variables

You can also configure or overwrite TzProxy with environment variables, using the same structure:


- `TZPROXY_HOST` is the host of the proxy.
- `TZPROXY_TEZOS_HOST` is the host of the tezos node.
- `TZPROXY_LOGGER_BUNCH_SIZE` is the bunch size of the logger.
- `TZPROXY_LOGGER_POOL_INTERVAL_SECONDS` is the pool interval of the logger.
- `TZPROXY_RATE_LIMIT_ENABLED` is a flag to enable rate limiting.
- `TZPROXY_RATE_LIMIT_MINUTES` is the minutes of the period of rate limiting. 
- `TZPROXY_RATE_LIMIT_MAX` is the max of requests permitted in a period.
- `TZPROXY_DENY_LIST_ENABLED` is a flag to block IP addresses.
- `TZPROXY_DENY_LIST_VALUES` is the IP Address that will be blocked on the proxy.
- `TZPROXY_DENY_ROUTES_ENABLED` is a flag to block the Tezos node's routes. 
- `TZPROXY_DENY_ROUTES_VALUES` is the Tezos nodes routes that will be blocked on the proxy.conf.
- `TZPROXY_CACHE_ENABLED` is the flag to cache enable cache.
- `TZPROXY_CACHE_DISABLED_ROUTES` is the routes to cache.
- `TZPROXY_CACHE_SIZE_MB` is the size of the cache in megabytes.
- `TZPROXY_CACHE_TTL` is the time to live in seconds for cache.
- `TZPROXY_METRICS_ENABLED` is the flag to enable metrics.
- `TZPROXY_METRICS_PPROF` is the flag to enable pprof.
- `TZPROXY_METRICS_HOST` is the host of the prometheus metrics and pprof (if enabled).
- `TZPROXY_GZIP_ENABLED` is the flag to enable gzip.
- `TZPROXY_CORS_ENABLED` is the flag to enable cors.
- `TZPROXY_GC_PERCENT` is the percent of the garbage collector.
