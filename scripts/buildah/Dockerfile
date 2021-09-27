ARG version

FROM quay.io/buildah/stable:$version

# make /dev/fuse available for user build in Docker Desktop
RUN usermod -a -G root build

# extend subuid/subgid range
RUN sed -i -e 's/build:2000:50000/build:2000:100000/' /etc/subuid /etc/subgid
