ARG BASE_REGISTRY=registry.ci.werf.io/base
FROM ${BASE_REGISTRY}/ubuntu:22.04
WORKDIR /www

RUN touch /created-by-run
