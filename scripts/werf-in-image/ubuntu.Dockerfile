FROM ubuntu:22.04
ENV DEBIAN_FRONTEND="noninteractive"

RUN apt-get -y update && \
    apt-get -y install fuse-overlayfs git uidmap libcap2-bin git-lfs curl gnupg nano jq bash make ca-certificates openssh-client iproute2 telnet iputils-ping dnsutils && \
    rm -rf /var/cache/apt/* /var/lib/apt/lists/* /var/log/*

RUN curl -sSLO https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 && \
    mv yq_linux_amd64 /usr/local/bin/yq && \
    chmod +x /usr/local/bin/yq

# Fix messed up setuid/setgid capabilities.
RUN setcap cap_setuid+ep /usr/bin/newuidmap && \
    setcap cap_setgid+ep /usr/bin/newgidmap && \
    chmod u-s,g-s /usr/bin/newuidmap /usr/bin/newgidmap

RUN useradd -m build
USER build:build
RUN mkdir -p /home/build/.local/share/containers
VOLUME /home/build/.local/share/containers

# Fix fatal: detected dubious ownership in repository.
RUN git config --global --add safe.directory '*'

WORKDIR /home/build

ENV WERF_CONTAINERIZED=yes
ENV WERF_BUILDAH_MODE=auto
