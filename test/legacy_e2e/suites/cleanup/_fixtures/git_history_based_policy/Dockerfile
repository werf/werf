FROM registry.werf.io/base/alpine:3.21 AS artifact
ARG FROM_CACHE_VERSION
RUN echo "cache=${FROM_CACHE_VERSION}" > /dev/null && echo "123" > /artifact

FROM registry.werf.io/base/alpine:3.21
ARG FROM_CACHE_VERSION
COPY --from=artifact /artifact /artifact
RUN echo "cache=${FROM_CACHE_VERSION}" > /dev/null && date > /date
