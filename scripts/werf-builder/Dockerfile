FROM golang:1.13-alpine

RUN apk add gcc bash libc-dev git

ADD cmd /werf/cmd
ADD pkg /werf/pkg
ADD go.mod /werf/go.mod
ADD go.sum /werf/go.sum
ADD scripts/lib /werf/scripts/lib

WORKDIR /werf

RUN bash -ec "echo '142.93.108.123 gonum.org www.gonum.org' >> /etc/hosts && source scripts/lib/release/global_data.sh && source scripts/lib/release/build.sh && go_mod_download"
