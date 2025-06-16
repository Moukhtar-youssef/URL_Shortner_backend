# Build stage
FROM --platform=$BUILDPLATFORM golang:1.21-alpine as builder
WORKDIR /app

# Enable Go's build cache for faster rebuilds
ENV GOCACHE=/tmp/go-build
ENV GOPATH=/tmp/go

# Copy only what's needed for dependencies
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/tmp/go-build \
    go mod download

# Copy source files
COPY . .

# Build for target platform
ARG TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/tmp/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH \
    go build -o /bin/URL_shortner_backend ./cmd/backend

# Final stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /bin/URL_shortner_backend .
COPY --from=builder /app/internl/public ./internl/public
EXPOSE 8081
CMD ["./URL_shortner_backend"]
