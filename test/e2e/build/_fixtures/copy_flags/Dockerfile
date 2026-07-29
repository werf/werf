FROM registry.werf.io/base/ubuntu:22.04

COPY --parents ./aaa/ ./bbb/ /target
COPY --parents ./aaa/./x /pivot
COPY --parents ./aaa/x /single
COPY --exclude=*.md ./docs /docs
COPY --exclude=**/*.md ./docs /docs-deep
