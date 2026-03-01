FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0

RUN apk add --no-cache make git && \
    make build

FROM alpine:3.21

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/bin/desolabs-data* /app/desolabs-data

WORKDIR /app
RUN mkdir -p /data

ENTRYPOINT ["./desolabs-data"]

EXPOSE 5000
