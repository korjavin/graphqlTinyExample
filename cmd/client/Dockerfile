FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/client ./cmd/client

# Create a minimal image
FROM alpine:latest

# Add necessary tools and SSL certificates
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/client /app/client

# Set environment variables (will be overridden by Docker Compose)
ENV SERVER_URL=http://server:8080/graphql

# Set the entry point to the client with default parameters
ENTRYPOINT ["/app/client"]

# Default command is to show all sellers
CMD ["-query", "sellers"]