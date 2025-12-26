FROM registry.werf.io/base/ubuntu:22.04

RUN apt update && apt install -yq openssh-client
RUN --mount=type=ssh ssh-add -L && echo "SSH agent is available"

COPY file /file

RUN touch /created-by-run
