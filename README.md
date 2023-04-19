# TzProxy

TzProxy is a reverse proxy specificly for Tezos Nodes written entirely in Go.

## Features

- [x] Rate limit
- [x] Block IPs
- [x] Blocklist routes

## Variables

- `HOST` is the host of the proxy.
- `TEZOS_HOST` is the host of the tezos node.
- `RATE_LIMIT_ENABLED` is a flag to enable rate limiting.
- `RATE_LIMIT_MINUTES` is the minutes of the period of rate limiting. 
- `RATE_LIMIT_MAX` is the max of requests permitted in a period.
- `BLOCK_ADDRESSES_ENABLED` is a flag to block IP addresses.
- `BLOCK_ADDRESSES` is the IP Address that will be blocked on the proxy.
- `BLOCK_ROUTES_ENABLED` is a flag to block the Tezos node's routes. 
- `BLOCK_ROUTES` is the Tezos nodes routes that will be blocked on the proxy.

