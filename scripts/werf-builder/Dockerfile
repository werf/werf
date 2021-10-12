FROM golang:1.17-bullseye

RUN dpkg --add-architecture arm64 && \
    apt-get update && \
    apt-get install -y gcc-aarch64-linux-gnu libc6-dev-arm64-cross libbtrfs-dev parallel && \
    apt-get install -y libbtrfs-dev:arm64 && \
    rm -rf /var/lib/apt/lists/*

ADD cmd /.werf-deps/cmd
ADD pkg /.werf-deps/pkg
ADD go.mod /.werf-deps/go.mod
ADD go.sum /.werf-deps/go.sum
ADD scripts /.werf-deps/scripts

RUN cd /.werf-deps && \
    ./scripts/build_release_v3.sh base && \
    rm -rf /.werf-deps
