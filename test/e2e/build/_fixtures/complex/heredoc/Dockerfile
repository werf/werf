FROM registry.werf.io/base/ubuntu:22.04 AS builder
RUN <<ADDUSER
addgroup -S app
adduser -S -h /app -s /bin/sh -D -G app app
ADDUSER
