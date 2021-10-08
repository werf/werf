FROM golang:1.17-bullseye

RUN dpkg --add-architecture arm64 && \
    apt-get update && \
    apt-get install -y gcc-aarch64-linux-gnu libc6-dev-arm64-cross \
      libgpgme-dev libgpgme-dev:arm64 libbtrfs-dev libbtrfs-dev:arm64 && \
    rm -rf /var/lib/apt/lists/*

ADD cmd pkg go.mod go.sum scripts /.werf-deps/

RUN cd /.werf-deps && \
    ./scripts/build_release_v3.sh base && \
    rm -rf /.werf-deps
