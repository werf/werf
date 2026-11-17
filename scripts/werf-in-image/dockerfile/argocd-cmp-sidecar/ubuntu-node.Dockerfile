FROM ubuntu:20.04@sha256:8feb4d8ca5354def3d8fce243717141ce31e2c428701f6682bd2fafe15388214
ENV DEBIAN_FRONTEND="noninteractive"

RUN apt-get -y update && apt-get -y install curl fuse-overlayfs git uidmap libcap2-bin tzdata && \
    rm -rf /var/cache/apt/* /var/lib/apt/lists/* /var/log/*

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

RUN useradd -m argocd -r && echo 'argocd:100000:65536' | tee /etc/subuid >/etc/subgid
USER argocd:argocd

COPY ../../argocd-cmp-sidecar-plugin.yaml /home/argocd/cmp-server/config/plugin.yaml

RUN mkdir -p /home/argocd/.local/share/containers
VOLUME /home/argocd/.local/share/containers

WORKDIR /home/argocd

ENV WERF_CONTAINERIZED=yes
ENV WERF_BUILDAH_MODE=auto
ENV WERF_DISABLE_AUTO_HOST_CLEANUP=1

ENTRYPOINT ["/var/run/argocd/argocd-cmp-server"]
