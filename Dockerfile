FROM golang:1.21-alpine3.19 AS build

WORKDIR /build

RUN apk add --no-cache git gcc musl-dev

COPY . .

RUN go build -o ./bin/timer .

FROM alpine:3.19

ENV DATA_FILE=/app/data.db
ENV TEMPLATES_DIR=/app/templates
ENV STATIC_DIR=/app/static

WORKDIR /app

COPY --from=build /build/bin/timer /app/
COPY templates /app/templates
COPY static /app/static

RUN apk add --no-cache ca-certificates && \
    addgroup -S -g 5000 timer && \
    adduser -S -u 5000 -G timer timer && \
    chown -R timer:timer .

USER timer

EXPOSE 8080

ENTRYPOINT ["/app/timer"]
