FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o server ./cmd/server

FROM alpine:3.19

WORKDIR /app

ENV GIN_MODE=release

COPY --from=builder /app/server /app/server

EXPOSE 8001

CMD ["/app/server"]

