FROM ubuntu:20.04
ENV DEBIAN_FRONTEND="noninteractive"

RUN apt-get -y update && apt-get -y install fuse-overlayfs git uidmap libcap2-bin && \
    rm -rf /var/cache/apt/* /var/lib/apt/lists/* /var/log/*

# Fix messed up setuid/setgid capabilities.
RUN setcap cap_setuid+ep /usr/bin/newuidmap && \
    setcap cap_setgid+ep /usr/bin/newgidmap && \
    chmod u-s,g-s /usr/bin/newuidmap /usr/bin/newgidmap

RUN useradd -m build
USER build:build
RUN mkdir -p /home/build/.local/share/containers
VOLUME /home/build/.local/share/containers

WORKDIR /home/build
ENV WERF_BUILDAH_MODE=auto
