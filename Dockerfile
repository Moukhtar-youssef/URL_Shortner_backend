# Build stage (named explicitly)
FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum .
RUN go mod download 

COPY . .

RUN go build -o /bin/backend ./cmd/backend/main.go

# Final stage
FROM alpine:latest
WORKDIR /app

# Copy from the named builder stage
COPY --from=builder /bin/backend .
COPY --from=builder /app/internl/public ./internl/public

EXPOSE 8081
CMD ["./backend"]
