# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app

# First copy only module files
COPY go.mod go.sum ./
RUN go mod download

# Then copy only the necessary source files
COPY cmd/ ./cmd/
COPY internl/ ./internl/
COPY pkg/ ./pkg/

# Build the application
RUN go build -o /bin/backend ./cmd/backend/main.go

# Final stage
FROM alpine:latest
WORKDIR /app

# Copy the binary
COPY --from=builder /bin/backend .


EXPOSE 8081
CMD ["./backend"]
