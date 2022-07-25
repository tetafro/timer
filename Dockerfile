FROM golang:1.18-alpine3.16 AS build

WORKDIR /build

RUN apk add --no-cache git gcc musl-dev

COPY . .

RUN go build -o ./bin/timer .

FROM alpine:3.16

WORKDIR /app

COPY --from=build /build/bin/timer /app/

RUN apk add --no-cache ca-certificates && \
    addgroup -S -g 5000 timer && \
    adduser -S -u 5000 -G timer timer && \
    chown -R timer:timer .

USER timer

ENTRYPOINT ["/app/timer"]
