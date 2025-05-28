FROM alpine:3.18
ARG TARGETARCH

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

RUN addgroup -g 1001 github-runner && \
    adduser -D build && \
    adduser -u 1001 -G github-runner -D -h /home/github-runner github-runner && \
    echo 'build:100000:65536' >> /etc/subuid && \
    echo 'build:100000:65536' >> /etc/subgid && \
    echo 'github-runner:165536:65536' >> /etc/subuid && \
    echo 'github-runner:165536:65536' >> /etc/subgid && \
    mkdir -p /home/build/.local/share/containers /home/build/.werf && \
    mkdir -p /home/github-runner/.local/share/containers /home/github-runner/.werf && \
    chown -R build:build /home/build && \
    chown -R github-runner:github-runner /home/github-runner

VOLUME ["/home/github-runner/.local/share/containers", "/home/build/.local/share/containers"]

USER build:build

# Fix fatal: detected dubious ownership in repository.
RUN git config --global --add safe.directory '*'

WORKDIR /home/build

ENV WERF_CONTAINERIZED=yes
ENV WERF_BUILDAH_MODE=auto
ENV WERF_DISABLE_AUTO_HOST_CLEANUP=1
