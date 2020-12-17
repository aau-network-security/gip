FROM golang:1.15-alpine AS builder
WORKDIR /ip
RUN apk add gcc g++ --no-cache
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o app -a -ldflags '-w -extldflags "-static"'  /ip/grpc/server/main.go

FROM alpine:latest
WORKDIR /app
RUN apk update
RUN apk add sudo --no-cache
RUN apk add iproute2 --no-cache
COPY --from=builder /ip/app /app/app
COPY --from=builder /ip/config/config.yml /app/config.yml
ENTRYPOINT ["/app/app"]