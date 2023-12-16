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
- [x] Redis


## How to run

Make sure that your Tezos Node is running and set its host address in the `tezos_host` configuration.

If you want to test only TzProxy without a real Tezos Node, you can simulate a Tezos Node with our `flextesa.sh` script. Make sure that you have a docker.

```bash
./flextesa.sh
```

If you want custom configurations, create a file named as `tzproxy.yaml` in the same directory of the binary. This file will be created if you don't create it.

Then, just [download the binary](https://github.com/marigold-dev/tzproxy/releases) and run it:
```bash
./tzproxy
```

Finally, test it with:
```bash
curl http://localhost:8080/chains/main/blocks/head/header
```

## Configuration

### Yaml File
Here a default `tzproxy.yaml` file:

```yaml
cache:
    disabled_routes:
        - /monitor/.*
        - /chains/.*/mempool
        - /chains/.*/blocks.*head
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
        - /chains/.*/blocks/.*/helpers/baking_rights
        - /chains/.*/blocks/.*/helpers/endorsing_rights
        - /helpers/baking_rights
        - /helpers/endorsing_rights
        - /chains/.*/blocks/.*/context/contracts(/?)$
        - /chains/.*/blocks/.*/context/raw/bytes
gc:
    optimize_memory_store: true
    percent: 100
gzip:
    enabled: true
host: 0.0.0.0:8080
load_balancer:
    ttl: 600
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
redis:
    enabled: false
    host: ""
tezos_host:
    - 127.0.0.1:8732
```

### Environment Variables

You can also configure or overwrite TzProxy with environment variables, using the same structure:


- `TZPROXY_HOST` is the host of the proxy.
- `TZPROXY_TEZOS_HOST` are the hosts of the tezos nodes.
- `TZPROXY_TEZOS_HOST_RETRY` is the host used when finding a 404 or 410. It's recommended use full or archive nodes.
- `TZPROXY_REDIS_HOST` is the host of the redis.
- `TZPROXY_REDIS_ENABLE` is a flag to enable redis.
- `TZPROXY_LOAD_BALANCER_TTL` is the time to live to keep using the same node by user IP.
- `TZPROXY_LOGGER_BUNCH_SIZE` is the bunch size of the logger.
- `TZPROXY_LOGGER_POOL_INTERVAL_SECONDS` is the pool interval of the logger.
- `TZPROXY_CACHE_ENABLED` is the flag to cache enable cache.
- `TZPROXY_CACHE_DISABLED_ROUTES` is the variable with the routes to cache.
- `TZPROXY_CACHE_SIZE_MB` is the size of the cache in megabytes.
- `TZPROXY_CACHE_TTL` is the time to live in seconds for cache.
- `TZPROXY_RATE_LIMIT_ENABLED` is a flag to enable rate limiting.
- `TZPROXY_RATE_LIMIT_MINUTES` is the minutes of the period of rate limiting. 
- `TZPROXY_RATE_LIMIT_MAX` is the max of requests permitted in a period.
- `TZPROXY_DENY_LIST_ENABLED` is a flag to block IP addresses.
- `TZPROXY_DENY_LIST_VALUES` is the IP Address that will be blocked on the proxy.
- `TZPROXY_DENY_ROUTES_ENABLED` is a flag to block the Tezos node's routes. 
- `TZPROXY_DENY_ROUTES_VALUES` is the Tezos nodes routes that will be blocked on the proxy.conf.
- `TZPROXY_METRICS_ENABLED` is the flag to enable metrics.
- `TZPROXY_METRICS_PPROF` is the flag to enable pprof.
- `TZPROXY_METRICS_HOST` is the host of the prometheus metrics and pprof (if enabled).
- `TZPROXY_CORS_ENABLED` is the flag to enable cors.
- `TZPROXY_GZIP_ENABLED` is the flag to enable gzip.
- `TZPROXY_GC_OPTIMIZE_MEMORY_STORE` is a flag to optimize GC when it's using storage as memory allocations instead of redis.
- `TZPROXY_GC_PERCENT` is the percent of the garbage collector.

