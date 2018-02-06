# Build tools

FROM ubuntu:16.04

RUN rm /bin/sh && ln -s /bin/bash /bin/sh
RUN >/etc/profile && >/root/.profile
SHELL ["/bin/bash", "-lc"]

RUN apt update && apt install -y curl git fakeroot gettext

ADD ./dappdeps-toolchain.tar /dappdeps-toolchain
RUN tar xf /dappdeps-toolchain/**/layer.tar -C /

RUN git config --global user.name flant && git config --global user.name 256@flant.com

RUN gpg --keyserver hkp://keys.gnupg.net --recv-keys 409B6B1796C275462A1703113804BB82D39DC0E3 && \
curl -sSL https://get.rvm.io | bash && \
. /etc/profile.d/rvm.sh && \
rvm install 2.3
RUN echo "source /etc/profile.d/rvm.sh" >> /etc/profile

ADD ./omnibus /omnibus
WORKDIR /omnibus
ENV BUNDLE_GEMFILE=/omnibus/Gemfile
RUN gem install bundler && \
bundle install --without development

ENV PATH=/.dapp/deps/toolchain/0.1.1/bin:$PATH
# Dpkg-architecture binary will make python-omnibus-package fail to build,
# because of python setup.py, which hardcodes /usr/include/... into preceeding include paths,
# in the case dpkg-architecture is available in system: https://github.com/python/cpython/blob/master/setup.py#L485
# It is needed to remove that binary before omnibus-building.
RUN \
export ORIG_DPKG_ARCHITECTURE_PATH=$(which dpkg-architecture) && \
mv $ORIG_DPKG_ARCHITECTURE_PATH /tmp/dpkg-architecture && \
bundle exec omnibus build -o append_timestamp:false dappdeps-ansible && \
mv /tmp/dpkg-architecture $ORIG_DPKG_ARCHITECTURE_PATH

RUN mkdir /tmp/result && \
dpkg -x /omnibus/pkg/dappdeps-ansible_2.4.1.0~1-1_amd64.deb /tmp/result

# Import tools into dappdeps/ansible scratch

FROM scratch
CMD ["no-such-command"]
COPY --from=0 /tmp/result/.dapp/deps/ansible/2.4.1.0-1 /.dapp/deps/ansible/2.4.1.0-1
