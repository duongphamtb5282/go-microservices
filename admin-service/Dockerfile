# Build stage
FROM golang:1.19-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o admin-service cmd/server/main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/admin-service .
COPY --from=builder /app/config ./config
COPY --from=builder /app/migrations ./migrations

# Expose ports
EXPOSE 8086 50051

# Run the application
CMD ["./admin-service"]

