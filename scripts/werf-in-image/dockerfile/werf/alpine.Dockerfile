FROM alpine:3.18
ARG TARGETARCH
ARG USERS="build"

RUN apk add --no-cache git git-lfs curl gnupg nano jq bash make ca-certificates openssh-client iproute2-ss busybox-extras tzdata

RUN curl -sSLO https://github.com/mikefarah/yq/releases/latest/download/yq_linux_${TARGETARCH} && \
    mv yq_linux_${TARGETARCH} /usr/local/bin/yq && \
    chmod +x /usr/local/bin/yq

RUN ARCH=`uname -m` && \
    case "$ARCH" in "aarch64") ARCH=arm64 ;; esac && \
    curl -sL "https://github.com/google/go-containerregistry/releases/download/v0.20.2/go-containerregistry_Linux_$ARCH.tar.gz" > go-containerregistry.tar.gz && \
    tar -zxvf go-containerregistry.tar.gz -C /usr/local/bin/ crane && \
    rm -f go-containerregistry.tar.gz

RUN set -eux; \
    for u in $USERS; do \
    adduser -D $u; \
    mkdir -p /home/$u/.werf; \
    chown -R $u:$u /home/$u; \
    done

USER build:build

# Fix fatal: detected dubious ownership in repository.
RUN git config --global --add safe.directory '*'

WORKDIR /home/build

ENV WERF_CONTAINERIZED=yes
ENV WERF_DISABLE_AUTO_HOST_CLEANUP=1
