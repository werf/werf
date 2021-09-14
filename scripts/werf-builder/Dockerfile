FROM golang:1.17-alpine

RUN apk add gcc bash libc-dev git
RUN go install github.com/mitchellh/gox@8c3b2b9e647dc52457d6ee7b5adcf97e2bafe131

ADD cmd /.werf-deps/cmd
ADD pkg /.werf-deps/pkg
ADD go.mod /.werf-deps/go.mod
ADD go.sum /.werf-deps/go.sum
ADD scripts /.werf-deps/scripts

RUN bash -ec "cd /.werf-deps && ./scripts/build_release_v3.sh base && rm -rf /.werf-deps"
