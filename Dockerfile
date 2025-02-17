FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./cmd/...

FROM alpine:3.21

WORKDIR /root/

COPY --from=builder /app/main .
CMD ["./main"]