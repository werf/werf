FROM golang:1.18.2-bullseye

RUN apt-get update && \
    apt-get install -y gcc-aarch64-linux-gnu libbtrfs-dev parallel && \
    rm -rf /var/lib/apt/lists/*

ADD cmd /.werf-deps/cmd
ADD pkg /.werf-deps/pkg
ADD go.mod /.werf-deps/go.mod
ADD go.sum /.werf-deps/go.sum
ADD scripts /.werf-deps/scripts

RUN cd /.werf-deps && \
    ./scripts/build_release_v3.sh base && \
    rm -rf /.werf-deps
