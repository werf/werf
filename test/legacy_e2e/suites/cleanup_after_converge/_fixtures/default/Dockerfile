FROM registry.werf.io/base/alpine:3.21 AS artifact
ARG ARTIFACT_CACHE_VERSION
ARG ARTIFACT_DATA
RUN echo "cache=${ARTIFACT_CACHE_VERSION}" > /dev/null && echo "${ARTIFACT_DATA}" > /artifact

FROM registry.werf.io/base/python:alpine
COPY --from=artifact /artifact /artifact
RUN mkdir -p /app && chmod 0755 /app && \
    printf '%s\n' \
      '<!DOCTYPE html>' \
      '<html>' \
      '  <body>' \
      '    <h2>Congratulations!</h2>' \
      '    <img src="https://flant.com/images/logo_en.png" style="max-height:100%%;" height="76">' \
      '  </body>' \
      '</html>' \
      > /app/index.html
