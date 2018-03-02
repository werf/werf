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
RUN bundle exec omnibus build -o append_timestamp:false dappdeps-gitartifact

RUN mkdir /tmp/result && \
dpkg -x /omnibus/pkg/dappdeps-gitartifact_0.2.1-1_amd64.deb /tmp/result

# Import tools into dappdeps/gitartifact scratch

FROM scratch
CMD ["no-such-command"]
COPY --from=0 /tmp/result/.dapp/deps/gitartifact/0.2.1 /.dapp/deps/gitartifact/0.2.1
