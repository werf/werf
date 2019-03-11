# Build

FROM ubuntu:18.04

RUN rm /bin/sh && ln -s /bin/bash /bin/sh
RUN >/etc/profile && >/root/.profile
SHELL ["/bin/bash", "-lc"]

RUN apt update && apt install -y curl git fakeroot gettext gpg ruby ruby-bundler ruby-dev make file m4 xz-utils

ADD ./werf-stapel-toolchain.tar /werf-stapel-toolchain
RUN tar xf /werf-stapel-toolchain/**/layer.tar -C /

RUN git config --global user.name flant && git config --global user.name 256@flant.com

ADD ./omnibus /omnibus
WORKDIR /omnibus
ENV BUNDLE_GEMFILE=/omnibus/Gemfile
RUN bundle install --without development

ENV PATH=/.werf/stapel/toolchain/0.2.0/x86_64-lfs-linux-gnu/bin:/.werf/stapel/toolchain/0.2.0/bin:$PATH
RUN bundle exec omnibus build -o append_timestamp:false werf-stapel-base

RUN mkdir /tmp/result && \
dpkg -x /omnibus/pkg/werf-stapel-base_*.deb /tmp/result

# Import

FROM scratch
CMD ["no-such-command"]
COPY --from=0 /tmp/result/.werf/stapel/base /.werf/stapel/base
