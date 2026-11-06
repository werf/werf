FROM registry.werf.io/base/alpine
# PROJECT_NAME is substituted by the spec: a project-unique layer keeps this build out of the
# cache shared with parallel specs, so the image the failing step started from is this project's own.
RUN echo PROJECT_NAME > /marker
RUN false
