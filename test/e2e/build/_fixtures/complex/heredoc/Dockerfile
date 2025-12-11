FROM alpine AS builder
RUN apk add --no-cache bash
RUN <<ADDUSER
addgroup -S app
adduser -S -h /app -s /bin/sh -D -G app app
ADDUSER
