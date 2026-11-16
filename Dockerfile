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

FROM gcr.io/distroless/static-debian12:nonroot AS runtime

WORKDIR /app
COPY --from=builder /out/telemetry-server /app/telemetry-server

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/telemetry-server"]
