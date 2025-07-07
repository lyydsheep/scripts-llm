FROM golang:1.24-alpine AS builder

COPY . /src
WORKDIR /src

RUN GOPROXY=https://goproxy.cn  mkdir -p bin/ &&  go build -o ./bin/server

FROM debian:11-slim
USER root

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    netbase \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get autoremove -y && apt-get autoclean -y


COPY --from=builder /src/bin /app

WORKDIR /app
EXPOSE 31721

CMD ["./server"]
