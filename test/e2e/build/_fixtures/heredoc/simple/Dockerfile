FROM registry.werf.io/base/ubuntu:22.04 AS builder
ARG FOO=bar
ENV BAR=baz

RUN mkdir -p /etc/myapp && \
    cat > /etc/myapp/env <<EOT
FOO=${FOO}
BAR=${BAR}
EOT
