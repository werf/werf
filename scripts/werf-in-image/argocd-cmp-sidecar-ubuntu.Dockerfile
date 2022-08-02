FROM ubuntu:20.04
ENV DEBIAN_FRONTEND="noninteractive"

RUN apt-get -y update && apt-get -y install fuse-overlayfs git uidmap libcap2-bin && \
    rm -rf /var/cache/apt/* /var/lib/apt/lists/* /var/log/*

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
