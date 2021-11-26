FROM ghcr.io/werf/builder:afd00fa1348645658b718df6b2b7447c6cead90b@sha256:829a6fd3d5850fc24366b5a3fb10402fe2d5615dcfd045b7225819c7a8ec5011
ARG build_version
ADD cmd /werf/cmd
ADD pkg /werf/pkg
ADD go.mod /werf/go.mod
ADD go.sum /werf/go.sum
ADD scripts/build_release_v3.sh /werf/scripts/build_release_v3.sh
RUN cd /werf && go mod download
RUN cd /werf && ./scripts/build_release_v3.sh $build_version

FROM ghcr.io/werf/buildah:v1.22.3-1
ARG build_version
RUN yum install -y git curl gpg
COPY --from=0 /werf/release-build/$build_version/linux-amd64/bin/werf /usr/local/bin/werf
ENV WERF_CONTAINER_RUNTIME_BUILDAH=native-rootless
