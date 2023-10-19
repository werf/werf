FROM fedora:38

RUN dnf -y install fuse-overlayfs git git-lfs gnupg nano jq bash make ca-certificates openssh-clients telnet iputils iproute dnsutils && \
    dnf clean all && rm -rf /var/cache /var/log/dnf* /var/log/yum.*

RUN curl -sSLO https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 && \
    mv yq_linux_amd64 /usr/local/bin/yq && \
    chmod +x /usr/local/bin/yq

# Fix messed up setuid/setgid capabilities.
RUN setcap cap_setuid+ep /usr/bin/newuidmap && \
    setcap cap_setgid+ep /usr/bin/newgidmap && \
    chmod u-s,g-s /usr/bin/newuidmap /usr/bin/newgidmap

RUN useradd build
USER build:build
RUN mkdir -p /home/build/.local/share/containers
VOLUME /home/build/.local/share/containers

# Fix fatal: detected dubious ownership in repository.
RUN git config --global --add safe.directory '*'

WORKDIR /home/build

ENV WERF_CONTAINERIZED=yes
ENV WERF_BUILDAH_MODE=auto
