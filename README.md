# TzProxy

TzProxy is a reverse proxy specificly for Tezos Nodes written entirely in Go.

## Features

- [x] Rate limit
- [x] Block IPs
- [x] Blocklist routes
- [x] Cache
- [x] CORS
- [x] Gzip
- [x] Metrics

## Variables

- `TZPROXY_HOST` is the host of the proxy.
- `TZPROXY_TEZOS_HOST` is the host of the tezos node.
- `TZPROXY_RATE_LIMIT_ENABLED` is a flag to enable rate limiting.
- `TZPROXY_RATE_LIMIT_MINUTES` is the minutes of the period of rate limiting. 
- `TZPROXY_RATE_LIMIT_MAX` is the max of requests permitted in a period.
- `TZPROXY_BLOCK_ADDRESSES_ENABLED` is a flag to block IP addresses.
- `TZPROXY_BLOCK_ADDRESSES` is the IP Address that will be blocked on the proxy.
- `TZPROXY_BLOCK_ROUTES_ENABLED` is a flag to block the Tezos node's routes. 
- `TZPROXY_BLOCK_ROUTES` is the Tezos nodes routes that will be blocked on the proxy.conf.
- `TZPROXY_CACHE_ENABLED` is the flag to cache enable cache.
- `TZPROXY_CACHE_ROUTES` is the routes to cache.
- `TZPROXY_CACHE_SIZE_MB` is the size of the cache in megabytes.
- `TZPROXY_CACHE_TTL` is the time to live in seconds for cache.
- `TZPROXY_ENABLE_PPROF` is the flag to enable pprof.
- `TZPROXY_ENABLE_GZIP` is the flag to enable gzip.
- `TZPROXY_CORS_ENABLED` is the flag to enable cors.
