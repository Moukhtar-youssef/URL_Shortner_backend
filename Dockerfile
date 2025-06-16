
FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod go.sum .
RUN go mod download 

COPY . .

RUN GOOS=linux go build -o /bin/backend ./cmd/backend/main.go

EXPOSE 8081

FROM alpine:latest

WORKDIR /app

COPY --from=builder /bin/backend .

COPY --from=builder /app/internl/public ./internl/public

EXPOSE 8081

CMD ["./backend"]
