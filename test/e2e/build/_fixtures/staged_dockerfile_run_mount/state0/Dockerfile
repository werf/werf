FROM registry.werf.io/base/ubuntu:22.04 AS os
RUN mkdir /apk && echo pkg-v1 > /apk/pkg.txt

FROM registry.werf.io/base/ubuntu:22.04
RUN --mount=type=bind,from=os,source=/apk,target=/apk cat /apk/pkg.txt > /result && echo run-v1
