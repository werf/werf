FROM ubuntu:22.04 AS builder

RUN apt-get update && \
    apt-get install -y curl

FROM scratch

COPY --from=builder /usr/bin/curl /usr/bin/curl
