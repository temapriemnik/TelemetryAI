FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o log-collector .

FROM alpine:3.19

RUN apk add --no-cache docker-cli ca-certificates

COPY --from=builder /app/log-collector .
COPY config.json .

ENTRYPOINT ["./log-collector"]