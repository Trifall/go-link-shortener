# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install system dependencies and swag in a single layer
RUN apk add --no-cache git make && \
  go install github.com/swaggo/swag/cmd/swag@latest

# Copy dependency files first to leverage layer caching
COPY go.mod go.sum ./

RUN go mod download

# Copy remaining source files
COPY . .

# Build the application
RUN make build

# Runtime stage
FROM alpine:3.21.3

WORKDIR /app

# Install runtime dependencies in a single layer
RUN apk add --no-cache postgresql-client ca-certificates

COPY --from=builder /app/bin/go-link-shortener .
COPY entrypoint.sh ./

EXPOSE 8080

ENTRYPOINT ["./entrypoint.sh"]