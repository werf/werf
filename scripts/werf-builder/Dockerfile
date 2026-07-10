FROM golang:1.25.12-bookworm@sha256:544ae22bc85dae968aff6777784957963228cf4834adc6bf674ff72cf9e83335
ENV DEBIAN_FRONTEND=noninteractive
ENV TASK_X_REMOTE_TASKFILES=1

RUN apt-get -y update && \
    apt-get -y install apt-utils gcc-aarch64-linux-gnu libbtrfs-dev file xz-utils &&  \
    curl -sSLO https://github.com/go-task/task/releases/download/v3.33.1/task_linux_amd64.deb && \
    apt-get -y install ./task_linux_amd64.deb && \
    rm -rf ./task_linux_amd64.deb /var/cache/apt/* /var/lib/apt/lists/* /var/log/*

RUN wget https://github.com/upx/upx/releases/download/v5.2.0/upx-5.2.0-amd64_linux.tar.xz && \
    tar xf upx-5.2.0-amd64_linux.tar.xz &&  mv upx-5.2.0-amd64_linux/upx /usr/local/bin/ && \
    rm -rf upx-5.2.0-amd64_linux.tar.xz upx-5.2.0-amd64_linux

ADD cmd /.werf-deps/cmd
ADD pkg /.werf-deps/pkg
ADD go.mod go.sum Taskfile.dist.yaml /.werf-deps/
ADD scripts /.werf-deps/scripts

RUN cd /.werf-deps && \
    task --yes build:dist:all version=base && \
    task --yes verify:binaries:dist:all version=base && \
    rm -rf /.werf-deps

RUN git config --global --add safe.directory /git
