FROM golang:1.26-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o fintech-app .


FROM alpine:latest

RUN apk add --no-cache sqlite ca-certificates

WORKDIR /root/

COPY  --from=builder /app/fintech-app .

ENV DB_NAME="transaction.db" 

ENV PORT=8080

EXPOSE 8080

CMD ["./fintech-app"]