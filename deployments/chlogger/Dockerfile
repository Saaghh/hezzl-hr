FROM golang:1.22 as builder


COPY ../.. /src
WORKDIR /src

RUN CGO_ENABLED=0 GOOS=linux go build -o bin/chlogger cmd/chlogger/main.go

FROM debian:stable-slim

COPY --from=builder /src/bin/chlogger /app/bin/chlogger

WORKDIR /app

ENV CH_BINDADDR="hezzl_clickhouse:9000"
ENV CH_BATCH_SIZE="1000"
ENV NATS_BINDADDR="hezzl_nats"

ENTRYPOINT ["./bin/chlogger"]

