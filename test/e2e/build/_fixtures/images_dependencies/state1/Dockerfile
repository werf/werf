FROM registry.werf.io/base/alpine:latest AS build-env-stage-1
RUN echo "hello-werf" > /stage-1-file

FROM registry.werf.io/base/alpine:latest AS build-env-stage-2
COPY --from=registry.werf.io/base/alpine:latest /etc/os-release /stage-2-file

FROM registry.werf.io/base/alpine:latest AS build-env-stage-3
COPY --from=build-env-stage-1 /stage-1-file /

FROM registry.werf.io/base/alpine:latest AS runtime-env
COPY --from=build-env-stage-3 /stage-1-file/ /

FROM runtime-env AS app
RUN echo "hello-werf"