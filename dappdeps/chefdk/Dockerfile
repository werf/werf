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

ADD https://github.com/chef/chef-dk/archive/v2.3.17.tar.gz /chefdk.tar.gz
RUN mkdir /omnibus && \
tar xf /chefdk.tar.gz -C /omnibus --strip-components 1
WORKDIR /omnibus
RUN gem install bundler && \
bundle install
RUN cd /omnibus/omnibus && \
bundle install

RUN echo 'install_dir "/.dapp/deps/chefdk/2.3.17-2"' >> omnibus_overrides.rb && \
sed -i -e 's@install_dir: /opt/chefdk@install_dir: /.dapp/deps/chefdk/2.3.17-2@g' omnibus/.kitchen.yml && \
sed -i -e 's@INSTALLER_DIR=/opt/chefdk@INSTALLER_DIR=/.dapp/deps/chefdk/2.3.17-2@g' omnibus/package-scripts/chefdk/postinst

ENV PATH=/.dapp/deps/toolchain/0.1.1/bin:$PATH
WORKDIR /omnibus/omnibus
ENV BUNDLE_GEMFILE=/omnibus/omnibus/Gemfile
RUN bundle exec omnibus build -o install_dir:/.dapp/deps/chefdk/2.3.17-2 -o append_timestamp:false chefdk

RUN mkdir /tmp/result && \
dpkg -x /omnibus/omnibus/pkg/chefdk_2.3.17-1_amd64.deb /tmp/result

# Import tools into dappdeps/chefdk scratch

FROM scratch
CMD ["no-such-command"]
COPY --from=0 /tmp/result/.dapp/deps/chefdk/2.3.17-2 /.dapp/deps/chefdk/2.3.17-2
