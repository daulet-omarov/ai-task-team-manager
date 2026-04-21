FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main ./cmd/api/main.go

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache curl

COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations

EXPOSE 7777

CMD ["./main"]
