FROM ubuntu:18.04

RUN apt update && apt install -y wget git curl build-essential gettext

RUN wget -O /tmp/go.tar.gz https://dl.google.com/go/go1.11.5.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf /tmp/go.tar.gz && rm /tmp/go.tar.gz

ENV GOROOT=/usr/local/go
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:/usr/local/go/bin:$PATH

ADD cmd /go/src/github.com/flant/werf/cmd
ADD pkg /go/src/github.com/flant/werf/pkg

ADD go-get.sh /go/src/github.com/flant/werf/go-get.sh
ADD scripts/lib /go/src/github.com/flant/werf/scripts/lib

WORKDIR /go/src/github.com/flant/werf

RUN bash -ec "source scripts/lib/release/global_data.sh && source /go/src/github.com/flant/werf/scripts/lib/release/build.sh && go_get"
