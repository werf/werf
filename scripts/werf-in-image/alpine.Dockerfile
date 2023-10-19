FROM alpine:3.18

RUN apk add --no-cache fuse-overlayfs git shadow-uidmap libcap git-lfs curl gnupg nano jq bash make ca-certificates openssh-client iproute2-ss busybox-extras

RUN curl -sSLO https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 && \
    mv yq_linux_amd64 /usr/local/bin/yq && \
    chmod +x /usr/local/bin/yq

# Fix messed up setuid/setgid capabilities.
RUN setcap cap_setuid+ep /usr/bin/newuidmap && \
    setcap cap_setgid+ep /usr/bin/newgidmap && \
    chmod u-s,g-s /usr/bin/newuidmap /usr/bin/newgidmap

RUN adduser -D build && echo 'build:100000:65536' | tee /etc/subuid >/etc/subgid
USER build:build
RUN mkdir -p /home/build/.local/share/containers
VOLUME /home/build/.local/share/containers

# Fix fatal: detected dubious ownership in repository.
RUN git config --global --add safe.directory '*'

WORKDIR /home/build

ENV WERF_CONTAINERIZED=yes
ENV WERF_BUILDAH_MODE=auto
