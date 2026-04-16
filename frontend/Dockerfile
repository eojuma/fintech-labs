# Step 1: Build the Go binary
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN go build -o main .

# Step 2: Final Runtime Image
FROM alpine:latest
WORKDIR /root/

# Copy the binary from the builder
COPY --from=builder /app/main .

# IMPORTANT: Copy the frontend files so the Go app can find them
COPY frontend/ ./frontend/

# Expose the port your Go app runs on (e.g., 8080)
EXPOSE 8080

CMD ["./main"]