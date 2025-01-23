# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git make
RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY . .
RUN go mod download
RUN make build

# Runtime stage
FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache postgresql-client ca-certificates

COPY --from=builder /app/bin/go-link-shortener .

# Ensure database initialization happens through GORM
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

# Read SERVER_PORT from .env file
COPY .env .env
RUN SERVER_PORT=$(grep SERVER_PORT .env | cut -d '=' -f2) && \
  echo "Exposing port $SERVER_PORT" && \
  echo "EXPOSE $SERVER_PORT" >> /app/Dockerfile.tmp

# Expose the port
EXPOSE $SERVER_PORT

ENTRYPOINT ["./entrypoint.sh"]