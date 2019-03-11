# Build tools

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
# Dpkg-architecture binary will make python-omnibus-package fail to build,
# because of python setup.py, which hardcodes /usr/include/... into preceeding include paths,
# in the case dpkg-architecture is available in system: https://github.com/python/cpython/blob/master/setup.py#L485
# It is needed to remove that binary before omnibus-building.
RUN bundle exec omnibus build -o append_timestamp:false werf-stapel-ansible

RUN mkdir /tmp/result && \
dpkg -x /omnibus/pkg/werf-stapel-ansible_*.deb /tmp/result

# `python-apt` package will install all libs in docker.from image.
# `python-apt` will be installed automatically by ansible apt module on first run.
# flant/werf-stapel-ansible offers support for ansible-apt-module only for ubuntu:14.04 ubuntu:16.04 ubuntu:18.04
RUN apt install -y python-apt
RUN \
cp /usr/lib/x86_64-linux-gnu/libapt-inst.so.2.0 /tmp/result/.werf/stapel/ansible/*/embedded/lib/ && \
cp /usr/lib/x86_64-linux-gnu/libapt-pkg.so.5.0 /tmp/result/.werf/stapel/ansible/*/embedded/lib/ && \
cp /usr/lib/x86_64-linux-gnu/liblz4.so.1 /tmp/result/.werf/stapel/ansible/*/embedded/lib/ && \
cp /lib/x86_64-linux-gnu/liblzma.so.5 /tmp/result/.werf/stapel/ansible/*/embedded/lib/ && \
cp /lib/x86_64-linux-gnu/libbz2.so.1.0 /tmp/result/.werf/stapel/ansible/*/embedded/lib/ && \
cp /usr/lib/python2.7/dist-packages/apt_inst.x86_64-linux-gnu.so /tmp/apt_inst.so && \
cp /tmp/apt_inst.so /tmp/result/.werf/stapel/ansible/*/embedded/lib/python2.7/ && \
cp /usr/lib/python2.7/dist-packages/apt_pkg.x86_64-linux-gnu.so /tmp/apt_pkg.so && \
cp /tmp/apt_pkg.so /tmp/result/.werf/stapel/ansible/*/embedded/lib/python2.7/ && \
cp -r /usr/lib/python2.7/dist-packages/apt /tmp/result/.werf/stapel/ansible/*/embedded/lib/python2.7/ && \
cp -r /usr/lib/python2.7/dist-packages/aptsources /tmp/result/.werf/stapel/ansible/*/embedded/lib/python2.7/

# Import

FROM scratch
CMD ["no-such-command"]
COPY --from=0 /tmp/result/.werf/stapel/ansible /.werf/stapel/ansible
