FROM registry.werf.io/base/alpine:3.21 AS artifact
RUN echo "123" > /artifact

FROM registry.werf.io/base/alpine:3.21
COPY --from=artifact /artifact /artifact
RUN date > /date
