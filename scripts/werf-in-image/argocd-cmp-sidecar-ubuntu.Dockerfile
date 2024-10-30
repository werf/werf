FROM ubuntu:20.04
ENV DEBIAN_FRONTEND="noninteractive"

RUN apt-get -y update && apt-get -y install curl fuse-overlayfs git uidmap libcap2-bin tzdata && \
    rm -rf /var/cache/apt/* /var/lib/apt/lists/* /var/log/*

RUN ARCH=`uname -m` && \
    if [[ $ARCH -eq "aarch64" ]]; then ARCH=arm64; fi && \
    curl -sL "https://github.com/google/go-containerregistry/releases/download/v0.20.2/go-containerregistry_Linux_$ARCH.tar.gz" > go-containerregistry.tar.gz && \
    tar -zxvf go-containerregistry.tar.gz -C /usr/local/bin/ crane && \
    rm -f go-containerregistry.tar.gz \

# Fix messed up setuid/setgid capabilities.
RUN setcap cap_setuid+ep /usr/bin/newuidmap && \
    setcap cap_setgid+ep /usr/bin/newgidmap && \
    chmod u-s,g-s /usr/bin/newuidmap /usr/bin/newgidmap

RUN useradd -m argocd -r && echo 'argocd:100000:65536' | tee /etc/subuid >/etc/subgid
USER argocd:argocd

COPY argocd-cmp-sidecar-plugin.yaml /home/argocd/cmp-server/config/plugin.yaml

RUN mkdir -p /home/argocd/.local/share/containers
VOLUME /home/argocd/.local/share/containers

WORKDIR /home/argocd

ENV WERF_CONTAINERIZED=yes
ENV WERF_BUILDAH_MODE=auto

ENTRYPOINT ["/var/run/argocd/argocd-cmp-server"]
