# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY backend/go.mod backend/go.sum ./

# Allow Go to download newer toolchain if needed
ENV GOTOOLCHAIN=auto

RUN go mod download

# Copy backend source code
COPY backend/ .

# Build the API binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/worker ./cmd/worker

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/api .
COPY --from=builder /app/worker .

# Copy static files from project root
COPY static ./static

# Expose port
EXPOSE 8080

# Run the API
CMD ["./api"]
