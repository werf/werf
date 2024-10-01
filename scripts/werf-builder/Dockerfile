FROM --platform=linux/amd64 golang:1.22-bookworm@sha256:00c0724d8fe7bb3c56fe06113aec3781576d14cd836c2701cef9013c192fd43c
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get -y update && \
    apt-get -y install apt-utils gcc-aarch64-linux-gnu libbtrfs-dev file &&  \
    curl -sSLO https://github.com/go-task/task/releases/download/v3.33.1/task_linux_amd64.deb && \
    apt-get -y install ./task_linux_amd64.deb && \
    rm -rf ./task_linux_amd64.deb /var/cache/apt/* /var/lib/apt/lists/* /var/log/*

ADD cmd /.werf-deps/cmd
ADD pkg /.werf-deps/pkg
ADD go.mod go.sum Taskfile.dist.yaml /.werf-deps/
ADD scripts /.werf-deps/scripts

RUN cd /.werf-deps && \
    task build:dist:all version=base && \
    task verify:binaries:dist:all version=base && \
    rm -rf /.werf-deps

RUN git config --global --add safe.directory /git
