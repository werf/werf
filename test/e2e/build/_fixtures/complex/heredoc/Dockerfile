FROM registry.werf.io/base/ubuntu:22.04 AS builder
RUN <<EOF
set -e
mkdir -p /app
echo 'hello-from-heredoc' > /app/heredoc.txt
EOF
