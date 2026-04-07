# Stage 1: Build the Go binary
FROM golang:alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the statically linked executable
RUN CGO_ENABLED=0 GOOS=linux go build -o scarrow-api ./cmd/api/main.go

# Stage 2: Create a lightweight runtime image
FROM alpine:latest

WORKDIR /app

# Install root CA certificates (needed for TLS/HTTPS calls) and tzdata for timezone support
RUN apk --no-cache add ca-certificates tzdata

# Copy the compiled binary from the builder stage
COPY --from=builder /app/scarrow-api .
# Copy .env file
COPY .env .

# Expose the API port
EXPOSE 38192

# Run the binary
CMD ["./scarrow-api"]