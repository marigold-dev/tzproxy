FROM golang:1.20.5 as builder
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . .
# To use the libc functions for net and os/user, and still get a static binary (for containers)
# https://github.com/remotemobprogramming/mob/issues/393
RUN go build -ldflags "-linkmode 'external' -extldflags '-static'" -o /tzproxy

FROM debian:12.0-slim
COPY --from=builder /tzproxy ./
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
EXPOSE 8080
ENTRYPOINT ["/tzproxy"]
CMD [""]
