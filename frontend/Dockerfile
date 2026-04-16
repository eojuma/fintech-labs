FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN go build -o main .

FROM alpine:latest
WORKDIR /root/

COPY --from=builder /app/main .

COPY frontend/ ./frontend/

EXPOSE 8080

CMD ["./main"]