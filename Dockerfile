# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o timelimiter ./cmd/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install postgres client for healthcheck
RUN apk --no-cache add postgresql-client

# Copy binary from builder
COPY --from=builder /app/timelimiter .

EXPOSE 3030

CMD ["./timelimiter", "-port=3030", "-backends=http://backend1:80,http://backend2:80,http://backend3:80"]