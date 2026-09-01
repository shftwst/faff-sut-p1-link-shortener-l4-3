# Multi-stage build: compile a static Go binary, then run it in a minimal Alpine
# image as a dedicated non-root user.
FROM golang:1.22-alpine AS build
WORKDIR /src

# Cache module downloads independently of source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# CGO disabled -> a static binary that runs in a minimal image with no libc deps.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.20
# wget backs the container healthcheck; ca-certificates for TLS if ever needed.
RUN apk add --no-cache wget ca-certificates \
    && adduser -D -u 10001 appuser
COPY --from=build /out/api /usr/local/bin/api
# The (empty in E1) migrations directory the startup runner reads from /migrations.
COPY --from=build /src/migrations /migrations
USER appuser
WORKDIR /
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/api"]
