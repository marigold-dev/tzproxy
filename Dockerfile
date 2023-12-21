FROM golang:1.21 as builder
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . .
RUN go build -o /tzproxy

FROM debian:12.4-slim
COPY --from=builder /tzproxy ./
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
EXPOSE 8080
ENTRYPOINT ["/tzproxy"]
CMD [""]
