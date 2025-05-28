FROM fedora:38
ARG TARGETARCH

RUN dnf -y install fuse-overlayfs git git-lfs gnupg nano jq bash make ca-certificates openssh-clients telnet iputils iproute dnsutils tzdata && \
    dnf clean all && rm -rf /var/cache /var/log/dnf* /var/log/yum.*

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

RUN useradd -m build && \
    groupadd -g 1001 github-runner && \
    useradd -u 1001 -g 1001 -m github-runner && \
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
