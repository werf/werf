FROM ubuntu:20.04
ENV DEBIAN_FRONTEND="noninteractive"

RUN apt-get -y update && apt-get -y install curl git tzdata && \
    rm -rf /var/cache/apt/* /var/lib/apt/lists/* /var/log/*

RUN ARCH=`uname -m` && \
    case "$ARCH" in "aarch64") ARCH=arm64 ;; esac && \
    curl -sL "https://github.com/google/go-containerregistry/releases/download/v0.20.2/go-containerregistry_Linux_$ARCH.tar.gz" > go-containerregistry.tar.gz && \
    tar -zxvf go-containerregistry.tar.gz -C /usr/local/bin/ crane && \
    rm -f go-containerregistry.tar.gz

RUN useradd -m argocd -r
USER argocd:argocd

COPY ../../argocd-cmp-sidecar-plugin.yaml /home/argocd/cmp-server/config/plugin.yaml

WORKDIR /home/argocd

ENV WERF_CONTAINERIZED=yes
ENV WERF_DISABLE_AUTO_HOST_CLEANUP=1

ENTRYPOINT ["/var/run/argocd/argocd-cmp-server"]
