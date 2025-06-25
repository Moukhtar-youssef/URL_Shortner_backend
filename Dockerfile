FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internl/ ./internl/
COPY pkg/ ./pkg/

RUN go build -o /bin/backend ./cmd/backend/main.go

FROM alpine:latest
WORKDIR /app

COPY --from=builder /bin/backend .

EXPOSE 8081
CMD ["./backend"]
