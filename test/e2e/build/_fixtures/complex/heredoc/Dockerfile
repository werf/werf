FROM registry.werf.io/base/ubuntu:22.04 AS builder
RUN apt-get update && apt-get install -y bash
RUN <<EOF
set -e
mkdir -p /app
echo 'hello-from-heredoc' > /app/heredoc.txt
EOF
