FROM golang:1.18.10-alpine3.17 as builder

WORKDIR /build

RUN apk add libpcap-dev gcc musl-dev

COPY . .

RUN GOOS=linux go build -ldflags '-s -w' -o bin/address-book

FROM alpine:3.18.5

WORKDIR /app

RUN apk add --no-cache libpcap-dev

COPY --from=builder /build/bin/address-book /app/address-book

EXPOSE 9988

CMD ["/app/address-book"]


