FROM golang:1.18-bullseye

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get -y update && \
    apt-get -y install gcc-aarch64-linux-gnu libbtrfs-dev &&  \
    curl -sSLO https://github.com/go-task/task/releases/download/v3.14.1/task_linux_amd64.deb && \
    apt-get -y install ./task_linux_amd64.deb && \
    rm -rf ./task_linux_amd64.deb /var/cache/apt/* /var/lib/apt/lists/* /var/log/*

ADD cmd /.werf-deps/cmd
ADD pkg /.werf-deps/pkg
ADD go.mod go.sum Taskfile.dist.yaml /.werf-deps/

RUN cd /.werf-deps && \
    task -p build:dist:all version=base && \
    rm -rf /.werf-deps
