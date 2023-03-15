FROM --platform=linux/amd64 golang:1.20-bullseye@sha256:51ff22f03320894402290ba7dfd83ee05b61e58b5381d76b40f2e3a370d81da3

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get -y update && \
    apt-get -y install gcc-aarch64-linux-gnu libbtrfs-dev file &&  \
    curl -sSLO https://github.com/go-task/task/releases/download/v3.14.1/task_linux_amd64.deb && \
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
