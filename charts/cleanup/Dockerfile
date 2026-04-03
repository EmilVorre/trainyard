# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Only copy what the cleanup binary needs
COPY charts/cleanup/cmd/cleanup/ ./charts/cleanup/cmd/cleanup/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" \
    -o /cleanup \
    ./charts/cleanup/cmd/cleanup/

# Runtime stage — distroless for minimal attack surface
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /cleanup /cleanup

ENTRYPOINT ["/cleanup"]