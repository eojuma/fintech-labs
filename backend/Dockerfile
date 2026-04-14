FROM golang:1.26-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o fintech-backend .

FROM alpine:latest

RUN apk add --no-cache sqlite ca-certificates

WORKDIR /app

COPY --from=builder /app/fintech-backend .

COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates

RUN mkdir -p /data && chmod 777 /data

ENV DB_PATH=/data/transaction.db
ENV PORT=8080

EXPOSE 8080

CMD ["./fintech-backend"]