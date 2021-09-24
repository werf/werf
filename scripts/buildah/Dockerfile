ARG version

FROM quay.io/buildah/stable:$version

# make /dev/fuse available for user build in Docker Desktop
RUN usermod -a -G root build
