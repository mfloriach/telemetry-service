# syntax=docker/dockerfile:1

FROM golang:1.27-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/telemetry-server ./cmd/telemetry-server
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/telemetry-relay ./cmd/relay

# Device simulator used by docker compose. Runs `make device` against the relay.
FROM alpine:3.21 AS device

RUN apk add --no-cache bash coreutils make netcat-openbsd websocat

WORKDIR /app
COPY Makefile .

ENV WS_URL=ws://telemetry-relay:9000/telemetry
CMD ["make", "device"]

FROM gcr.io/distroless/static-debian12:nonroot AS telemetry-runtime

WORKDIR /app
COPY --from=builder /out/telemetry-server /app/telemetry-server

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/telemetry-server"]

# Final/default stage: the WebSocket telemetry relay (cmd/relay).
FROM gcr.io/distroless/static-debian12:nonroot AS relay-runtime

WORKDIR /app
COPY --from=builder /out/telemetry-relay /app/telemetry-relay

EXPOSE 9000
USER nonroot:nonroot
ENTRYPOINT ["/app/telemetry-relay"]
