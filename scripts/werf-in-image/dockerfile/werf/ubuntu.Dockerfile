FROM ubuntu:22.04
ENV DEBIAN_FRONTEND="noninteractive"
ARG TARGETARCH
ARG USERS="build build1001"

RUN apt-get -y update && \
    apt-get -y install fuse-overlayfs git uidmap libcap2-bin git-lfs curl gnupg nano jq bash make ca-certificates openssh-client iproute2 telnet iputils-ping dnsutils tzdata && \
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

RUN for u in $USERS; do \
    useradd -m $u && \
    mkdir -p /home/$u/.local/share/containers /home/$u/.werf && \
    chown -R $u:$u /home/$u && \
    runuser -u $u -- git config --global --add safe.directory '*' ; \
    done

USER build:build

WORKDIR /home/build

ENV WERF_CONTAINERIZED=yes
ENV WERF_BUILDAH_MODE=auto
ENV WERF_DISABLE_AUTO_HOST_CLEANUP=1