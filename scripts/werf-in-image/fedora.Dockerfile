FROM registry.fedoraproject.org/fedora-minimal:36

RUN microdnf -y install fuse-overlayfs git && \
    microdnf clean all && rm -rf /var/cache /var/log/dnf* /var/log/yum.*

# Fix messed up setuid/setgid capabilities.
RUN setcap cap_setuid+ep /usr/bin/newuidmap && \
    setcap cap_setgid+ep /usr/bin/newgidmap && \
    chmod u-s,g-s /usr/bin/newuidmap /usr/bin/newgidmap

RUN useradd build
USER build:build
RUN mkdir -p /home/build/.local/share/containers
VOLUME /home/build/.local/share/containers

ENV WERF_BUILDAH_MODE=auto
