# Stage 1: Build Go application
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o backend ./cmd/backend

# Stage 2: Runtime
FROM alpine:latest
RUN apk --no-cache add ca-certificates wget
WORKDIR /root/
COPY --from=builder /app/backend .
COPY --from=builder /app/database/migrations ./database/migrations
EXPOSE 8080
CMD ["./backend", "serve"]
