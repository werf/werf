FROM ubuntu:24.04@sha256:fe81560990fae86dc4743202b2cb4d3df98d0bcf949173cecef15ce1127ef852
ENV DEBIAN_FRONTEND="noninteractive"
ARG TARGETARCH
ARG USERS="build build1001"

RUN apt-get -y update && \
    apt-get -y install fuse-overlayfs netavark passt aardvark-dns iptables git uidmap libcap2-bin git-lfs curl gnupg nano jq bash make ca-certificates openssh-client iproute2 telnet iputils-ping dnsutils tzdata && \
    rm -rf /var/cache/apt/* /var/lib/apt/lists/* /var/log/*

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

COPY --from=node:26.13.0-trixie-slim@sha256:5adad64f9a1fca2d51972caa65983634c8a774eb300d72b20f7f2a0f60368072 /usr/local/bin/node /usr/local/bin/
COPY --from=node:26.13.0-trixie-slim@sha256:5adad64f9a1fca2d51972caa65983634c8a774eb300d72b20f7f2a0f60368072 /usr/local/lib/node_modules/npm /usr/local/lib/node_modules/npm
RUN ln -s ../lib/node_modules/npm/bin/npm-cli.js /usr/local/bin/npm && \
    ln -s ../lib/node_modules/npm/bin/npx-cli.js /usr/local/bin/npx && \
    apt-get -y update && apt-get -y install libatomic1 ca-certificates git && \
    rm -rf /var/cache/apt/* /var/lib/apt/lists/* /var/log/* && \
    npm i -g pnpm@12.18.0 yarn@1.22.22 && npm cache clean --force
ENV NODE_USE_SYSTEM_CA=1

# Remove the stock ubuntu user (UID 1000) so build/build1001 keep UIDs 1000/1001
RUN userdel -r ubuntu && \
    for u in $USERS; do \
    useradd -m $u && \
    mkdir -p /home/$u/.local/share/containers /home/$u/.werf && \
    chown -R $u:$u /home/$u && \
    runuser -u $u -- git config --global --add safe.directory '*' ; \
    done

USER build1001:build1001
VOLUME /home/build1001/.local/share/containers

USER build:build
VOLUME /home/build/.local/share/containers

WORKDIR /home/build

ENV WERF_CONTAINERIZED=yes
ENV WERF_BUILDAH_MODE=auto
ENV WERF_DISABLE_AUTO_HOST_CLEANUP=1