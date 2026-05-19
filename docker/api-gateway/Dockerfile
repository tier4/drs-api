# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /build

# Copy go mod files
COPY services/api-gateway/go.mod services/api-gateway/go.sum ./services/api-gateway/
COPY tools/drs-cli/go.mod tools/drs-cli/go.sum ./tools/drs-cli/

# Download dependencies
WORKDIR /build/services/api-gateway
RUN go mod download

# Copy proto files and scripts first
WORKDIR /build
COPY proto ./proto
COPY scripts ./scripts

# Install protoc and Go protoc plugins
RUN apk add --no-cache bash protobuf protobuf-dev && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.6 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

# Generate proto files
RUN chmod +x scripts/generate-proto.sh && \
    ./scripts/generate-proto.sh --go-only

# Copy api-gateway source code
COPY services/api-gateway ./services/api-gateway

# Build the binary
WORKDIR /build/services/api-gateway
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -installsuffix cgo -ldflags="-w -s" -o api-gateway cmd/server/main.go

# Runtime stage
FROM gcr.io/distroless/static:nonroot

# Copy the binary
COPY --from=builder /build/services/api-gateway/api-gateway /api-gateway

# Copy default config
COPY --from=builder /build/services/api-gateway/config.yaml /etc/api-gateway/config.yaml

# Expose port (default 8080 from config.yaml)
EXPOSE 8080

# Set user
USER nonroot:nonroot

# Run the binary
ENTRYPOINT ["/api-gateway"]
CMD ["-config", "/etc/api-gateway/config.yaml"]
