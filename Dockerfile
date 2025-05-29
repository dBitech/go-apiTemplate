# Build stage
FROM golang:1.24-alpine AS build

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy go.mod and go.sum files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build application
RUN make build

# Final stage
FROM alpine:3.21

# Set working directory
WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from build stage
COPY --from=build /app/bin/api-template /app/

# Copy config file
COPY config.yaml /app/

# Set environment variables
ENV APP_SERVER_HOST="0.0.0.0" \
    APP_LOGGING_LEVEL="info" \
    APP_LOGGING_FORMAT="json"

# Expose port
EXPOSE 8080 9090

# Run application
ENTRYPOINT ["/app/api-template"]
