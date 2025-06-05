FROM alpine:3.18
ARG TARGETARCH
ARG USERS="build build1001"

RUN apk add --no-cache fuse-overlayfs git shadow-uidmap libcap git-lfs curl gnupg nano jq bash make ca-certificates openssh-client iproute2-ss busybox-extras tzdata

RUN curl -sSLO https://github.com/mikefarah/yq/releases/latest/download/yq_linux_${TARGETARCH} && \
    mv yq_linux_${TARGETARCH} /usr/local/bin/yq && \
    chmod +x /usr/local/bin/yq

RUN ARCH=`uname -m` && \
    case "$ARCH" in "aarch64") ARCH=arm64 ;; esac && \
    curl -sL "https://github.com/google/go-containerregistry/releases/download/v0.20.2/go-containerregistry_Linux_$ARCH.tar.gz" > go-containerregistry.tar.gz && \
    tar -zxvf go-containerregistry.tar.gz -C /usr/local/bin/ crane && \
    rm -f go-containerregistry.tar.gz

# Fix messed up setuid/setgid capabilities.
RUN setcap cap_setuid+ep /usr/bin/newuidmap && \
    setcap cap_setgid+ep /usr/bin/newgidmap && \
    chmod u-s,g-s /usr/bin/newuidmap /usr/bin/newgidmap

RUN set -eux; \
    OFFSET=100000; \
    for u in $USERS; do \
    adduser -D $u; \
    mkdir -p /home/$u/.local/share/containers /home/$u/.werf; \
    chown -R $u:$u /home/$u; \
    echo "$u:$OFFSET:65536" >> /etc/subuid; \
    echo "$u:$OFFSET:65536" >> /etc/subgid; \
    OFFSET=$((OFFSET + 65536)); \
    done

USER build1001:build1001
VOLUME /home/build1001/.local/share/containers

# Fix fatal: detected dubious ownership in repository.
RUN git config --global --add safe.directory '*'

USER build:build
VOLUME /home/build/.local/share/containers

# Fix fatal: detected dubious ownership in repository.
RUN git config --global --add safe.directory '*'

WORKDIR /home/build

ENV WERF_CONTAINERIZED=yes
ENV WERF_BUILDAH_MODE=auto
ENV WERF_DISABLE_AUTO_HOST_CLEANUP=1
