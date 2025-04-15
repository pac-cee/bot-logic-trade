FROM golang:1.20-alpine
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o matching_engine_redis ./matching_engine_redis.go
EXPOSE 8081
CMD ["./matching_engine_redis"]
