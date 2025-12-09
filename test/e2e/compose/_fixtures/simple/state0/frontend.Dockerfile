FROM registry.werf.io/base/ubuntu:22.04
WORKDIR /www

RUN touch /created-by-run
