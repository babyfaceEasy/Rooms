# Stage 1: Build the Go binary.
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Download dependencies first to maximize Docker layer caching.
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build a statically linked binary.
COPY . .

ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build -ldflags="-w -s" -o bin/api ./cmd/api

# Stage 2: Minimal production image based on Google's Distroless.
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Bring in CA certificates and timezone data for outbound TLS connections.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=builder /app/bin/api .

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["./api"]
