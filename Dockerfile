# --- Build Stage ---
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api

# --- Final Stage ---
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates and timezone data
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/main .
# Copy migrations
COPY --from=builder /app/migrations ./migrations

# Create uploads directory
RUN mkdir uploads

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./main"]
