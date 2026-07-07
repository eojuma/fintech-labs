FROM golang:1.21-alpine AS builder

# Install gcc and sqlite for CGO
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Download dependencies first (cached layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build the binary from the new cmd/server location
RUN CGO_ENABLED=1 GOOS=linux go build -o african-vault ./cmd/server/main.go

# Final lightweight image
FROM alpine:latest

RUN apk add --no-cache sqlite ca-certificates

WORKDIR /app

# Copy the binary
COPY --from=builder /app/african-vault .

# Copy web assets (templates and static files)
COPY --from=builder /app/web ./web

# Create data directory for SQLite database
RUN mkdir -p /data && chmod 777 /data

ENV DATABASE_PATH=/data/transaction.db
ENV PORT=8080

EXPOSE 8080

CMD ["./african-vault"]