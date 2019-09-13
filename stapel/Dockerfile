FROM ubuntu:18.04 as base

RUN rm /bin/sh && ln -s /bin/bash /bin/sh
RUN >/etc/profile && >/root/.profile
RUN echo "set +h" >> /root/.profile && \
echo "umask 022" >> /root/.profile

SHELL ["/bin/bash", "-lc"]

RUN apt update && apt install -y \
build-essential wget curl gawk flex bison bzip2 liblzma5 texinfo file \
gettext python python3 curl git fakeroot gettext gpg ruby ruby-bundler \
ruby-dev make file m4 xz-utils texlive vim

RUN git config --global user.name flant && git config --global user.email 256@flant.com

ENV LFS=/mnt/lfs
ENV TOOLS=/.werf/stapel
ENV LFS_TGT=x86_64-lfs-linux-gnu

RUN mkdir -pv $LFS$TOOLS && mkdir -pv $LFS/sources && chmod -v a+wt $LFS/sources
ADD stapel/wget-list $LFS/sources/wget-list
ADD stapel/md5sums $LFS/sources/md5sums
# TODO: rename md5sums & wget-list files
RUN wget --input-file=$LFS/sources/wget-list --continue --directory-prefix=$LFS/sources
RUN bash -c "pushd $LFS/sources && md5sum -c $LFS/sources/md5sums && popd"
ADD stapel/version-check.sh $LFS/sources/version-check.sh
RUN $LFS/sources/version-check.sh

RUN ln -sv $LFS/.werf /

ENV LC_ALL=POSIX
ENV PATH=$TOOLS/bin:/bin:/usr/bin
ENV MAKEFLAGS='-j 5'

RUN echo "Binutils pass 1" && cd $LFS/sources/ && \
mkdir binutils && \
tar xf binutils-*.tar.* -C binutils --strip-components 1 && \
cd binutils && \
mkdir -v build && \
cd build && \
../configure --prefix=$TOOLS \
--with-sysroot=$LFS \
--with-lib-path=$TOOLS/lib \
--target=$LFS_TGT \
--disable-nls \
--disable-werror
WORKDIR $LFS/sources/binutils/build
RUN make
RUN mkdir -pv $TOOLS/lib && ln -sv lib $TOOLS/lib64 && make install

ADD stapel/gcc-before-configure.sh $LFS/sources/gcc-before-configure.sh
RUN echo "GCC pass 1" && cd $LFS/sources/ && \
mkdir gcc && \
tar xf gcc-*.tar.* -C gcc --strip-components 1 && \
mkdir gcc/mpfr && \
tar xf mpfr*.tar.* -C gcc/mpfr --strip-components 1 && \
mkdir gcc/gmp && \
tar xf gmp*.tar.* -C gcc/gmp --strip-components 1 && \
mkdir gcc/mpc && \
tar xf mpc*.tar.* -C gcc/mpc --strip-components 1 && \
cd gcc && \
$LFS/sources/gcc-before-configure.sh && \
mkdir -v build && \
cd build && \
../configure \
--target=$LFS_TGT \
--prefix=$TOOLS \
--with-glibc-version=2.11 \
--with-sysroot=$LFS \
--with-newlib \
--without-headers \
--with-local-prefix=$TOOLS \
--with-native-system-header-dir=$TOOLS/include \
--disable-nls \
--disable-shared \
--disable-multilib \
--disable-decimal-float \
--disable-threads \
--disable-libatomic \
--disable-libgomp \
--disable-libmpx \
--disable-libquadmath \
--disable-libssp \
--disable-libvtv \
--disable-libstdcxx \
--enable-languages=c,c++
WORKDIR $LFS/sources/gcc/build
RUN make
RUN make install

RUN cd $LFS/sources/ && \
mkdir linux && \
tar xf linux*.tar.* -C linux --strip-components 1
WORKDIR $LFS/sources/linux
RUN echo "Linux API Headers" && make mrproper && \
make INSTALL_HDR_PATH=dest headers_install && \
cp -rv dest/include/* $TOOLS/include

RUN echo "Glibc" && cd $LFS/sources/ && \
mkdir glibc && \
tar xf glibc*.tar.* -C glibc --strip-components 1 && \
cd glibc && \
mkdir -v build && \
cd build && \
../configure \
--prefix=$TOOLS \
--host=$LFS_TGT \
--build=$(../scripts/config.guess) \
--enable-kernel=3.2 \
--with-headers=$TOOLS/include \
libc_cv_forced_unwind=yes \
libc_cv_c_cleanup=yes
WORKDIR $LFS/sources/glibc/build
RUN make
RUN make install

RUN echo "Libstdc++" && cd $LFS/sources/ && \
rm -rf gcc && \
mkdir gcc && \
tar xf gcc-*.tar.* -C gcc --strip-components 1 && \
mkdir gcc/mpfr && \
tar xf mpfr*.tar.* -C gcc/mpfr --strip-components 1 && \
mkdir gcc/gmp && \
tar xf gmp*.tar.* -C gcc/gmp --strip-components 1 && \
mkdir gcc/mpc && \
tar xf mpc*.tar.* -C gcc/mpc --strip-components 1 && \
cd gcc && \
$LFS/sources/gcc-before-configure.sh && \
mkdir -v build && \
cd build && \
../libstdc++-v3/configure \
--host=$LFS_TGT \
--prefix=$TOOLS \
--disable-multilib \
--disable-nls \
--disable-libstdcxx-threads \
--disable-libstdcxx-pch \
--with-gxx-include-dir=$TOOLS/$LFS_TGT/include/c++/8.2.0
WORKDIR $LFS/sources/gcc/build
RUN make
RUN make install

RUN echo "Binutils pass 2" && cd $LFS/sources/ && \
rm -rf binutils && \
mkdir binutils && \
tar xf binutils-*.tar.* -C binutils --strip-components 1 && \
cd binutils && \
mkdir -v build && \
cd build && \
CC=$LFS_TGT-gcc \
AR=$LFS_TGT-ar \
RANLIB=$LFS_TGT-ranlib \
../configure \
--prefix=$TOOLS \
--disable-nls \
--disable-werror \
--with-lib-path=$TOOLS/lib \
--with-sysroot

RUN echo "GCC pass 2" && cd $LFS/sources/ && \
rm -rf gcc && \
mkdir gcc && \
tar xf gcc-*.tar.* -C gcc --strip-components 1 && \
mkdir gcc/mpfr && \
tar xf mpfr*.tar.* -C gcc/mpfr --strip-components 1 && \
mkdir gcc/gmp && \
tar xf gmp*.tar.* -C gcc/gmp --strip-components 1 && \
mkdir gcc/mpc && \
tar xf mpc*.tar.* -C gcc/mpc --strip-components 1 && \
cd gcc && \
cat gcc/limitx.h gcc/glimits.h gcc/limity.h > \
  `dirname $($LFS_TGT-gcc -print-libgcc-file-name)`/include-fixed/limits.h && \
$LFS/sources/gcc-before-configure.sh && \
mkdir -v build && \
cd build && \
CC=$LFS_TGT-gcc \
CXX=$LFS_TGT-g++ \
AR=$LFS_TGT-ar \
RANLIB=$LFS_TGT-ranlib \
../configure \
--prefix=$TOOLS \
--with-local-prefix=$TOOLS \
--with-native-system-header-dir=$TOOLS/include \
--enable-languages=c,c++ \
--disable-libstdcxx-pch \
--disable-multilib \
--disable-bootstrap \
--disable-libgomp
WORKDIR $LFS/sources/gcc/build
RUN make
RUN make install
RUN ln -sv gcc $TOOLS/bin/cc

ADD stapel/omnibus /omnibus
WORKDIR /omnibus
ENV BUNDLE_GEMFILE=/omnibus/Gemfile

ENV PATH=/bin:/usr/bin:/usr/local/bin
RUN bundle install --without development

ENV PATH=$TOOLS/x86_64-lfs-linux-gnu/bin:$TOOLS/bin:$PATH

# Dpkg-architecture binary will make python-omnibus-package fail to build,
# because of python setup.py, which hardcodes /usr/include/... into preceeding include paths,
# in the case when dpkg-architecture is available in system: https://github.com/python/cpython/blob/master/setup.py#L485
# It is needed to remove that binary before omnibus-building.
RUN mv $(which dpkg-architecture) /tmp/dpkg-architecture

RUN bundle exec omnibus build -o append_timestamp:false werf-stapel
ENV PKG_CONFIG=/.werf/stapel/embedded/bin/pkg-config
ENV PKG_CONFIG_PATH="/.werf/stapel/lib/pkgconfig:/.werf/stapel/embedded/lib/pkgconfig"

ADD stapel/wget-list-after-omnibus $LFS/sources/wget-list-after-omnibus
ADD stapel/wget-list-after-omnibus.md5sums $LFS/sources/wget-list-after-omnibus.md5sums
RUN wget --input-file=$LFS/sources/wget-list-after-omnibus --continue --directory-prefix=$LFS/sources
RUN bash -c "pushd $LFS/sources && md5sum -c $LFS/sources/wget-list-after-omnibus.md5sums && popd"

# libgpg-error
RUN echo "libgpg-error" && cd $LFS/sources && \
mkdir libgpg-error && \
tar xf libgpg-error*.tar.* -C libgpg-error --strip-components 1
WORKDIR $LFS/sources/libgpg-error
RUN sed -i 's/namespace/pkg_&/' src/Makefile.{am,in} src/mkstrtable.awk
RUN ./configure --prefix=$TOOLS
RUN make
RUN make install

# libgcrypt
RUN echo "libgcrypt" && cd $LFS/sources && \
mkdir libgcrypt && \
tar xf libgcrypt*.tar.* -C libgcrypt --strip-components 1
WORKDIR $LFS/sources/libgcrypt
RUN ./configure --prefix=$TOOLS
RUN make
RUN make install

# libassuan
RUN echo "libassuan" && cd $LFS/sources && \
mkdir libassuan && \
tar xf libassuan*.tar.* -C libassuan --strip-components 1
WORKDIR $LFS/sources/libassuan
RUN ./configure --prefix=$TOOLS
RUN make
RUN make install

# libksba
RUN echo "libksba" && cd $LFS/sources && \
mkdir libksba && \
tar xf libksba*.tar.* -C libksba --strip-components 1
WORKDIR $LFS/sources/libksba
RUN ./configure --prefix=$TOOLS
RUN make
RUN make install

# npth
RUN echo "npth" && cd $LFS/sources && \
mkdir npth && \
tar xf npth*.tar.* -C npth --strip-components 1
WORKDIR $LFS/sources/npth
RUN ./configure --prefix=$TOOLS
RUN make
RUN make install

# gmp
RUN echo "gmp" && cd $LFS/sources && \
mkdir gmp && \
tar xf gmp*.tar.* -C gmp --strip-components 1
WORKDIR $LFS/sources/gmp
RUN ./configure --prefix=$TOOLS --disable-static
RUN make
RUN make install

# nettle
RUN echo "nettle" && cd $LFS/sources && \
mkdir nettle && \
tar xf nettle*.tar.* -C nettle --strip-components 1
WORKDIR $LFS/sources/nettle
RUN ./configure --prefix=$TOOLS --disable-static
RUN make
RUN make install && chmod -v 755 $TOOLS/lib/libnettle.so

## p11-kit
#RUN echo "p11-kit" && cd $LFS/sources && \
#mkdir p11-kit && \
#tar xf p11-kit*.tar.* -C p11-kit --strip-components 1
#WORKDIR $LFS/sources/p11-kit
#RUN ./configure --prefix=$TOOLS --sysconfigdir=$TOOLS/etc

# gnutls
RUN echo "gnutls" && cd $LFS/sources && \
mkdir gnutls && \
tar xf gnutls*.tar.* -C gnutls --strip-components 1
WORKDIR $LFS/sources/gnutls
RUN ./configure --prefix=$TOOLS --disable-guile --without-p11-kit --with-included-libtasn1 --with-included-unistring
RUN make
RUN make install

# gnupg
RUN echo "gnupg" && cd $LFS/sources && \
mkdir gnupg && \
tar xf gnupg*.tar.* -C gnupg --strip-components 1
WORKDIR $LFS/sources/gnupg
RUN ./configure --prefix=$TOOLS --enable-symcryptrun --localstatedir=/.werf/stapel/var
RUN make
RUN make install

# Restore PATH *not to* use stapel build tools anymore
ENV PATH=/sbin:/bin:/usr/sbin:/usr/bin:/usr/local/sbin:/usr/local/bin

# Install golang
RUN wget https://dl.google.com/go/go1.12.linux-amd64.tar.gz -O /tmp/go.tar.gz
RUN tar -C /usr/local -xzf /tmp/go.tar.gz
ENV PATH=$PATH:/usr/local/go/bin
ENV GO111MODULE=on

WORKDIR /werf
ADD go.mod /werf/go.mod
ADD go.sum /werf/go.sum
RUN bash -ec "echo '142.93.108.123 gonum.org www.gonum.org' >> /etc/hosts && go mod download -json"

ADD cmd /werf/cmd
ADD pkg /werf/pkg
RUN go build -o /.werf/stapel/embedded/lib/python2.7/_logboek.so -buildmode=c-shared github.com/flant/logboek/c_lib
RUN cp /root/go/pkg/mod/github.com/flant/logboek@*/logboek.py /.werf/stapel/embedded/lib/python2.7
RUN cp /werf/pkg/build/builder/ansible/crypt.py /.werf/stapel/embedded/lib/python2.7

# Create ansible tools overlay path
RUN mkdir -p $TOOLS/ansible_tools_overlay/bin $TOOLS/ansible_tools_overlay/lib

# python-apt package will install all libs in docker.from image.
# python-apt will be installed automatically by ansible apt module on first run.
# flant/werf-stapel offers support for ansible-apt-module only for ubuntu:14.04 ubuntu:16.04 ubuntu:18.04
RUN apt install -y python-apt libzstd1 apt-utils
RUN \
cp /usr/lib/x86_64-linux-gnu/libapt-private.so.0.0 $TOOLS/embedded/lib/ && \
cp /usr/lib/x86_64-linux-gnu/libapt-private.so.0.0.0 $TOOLS/embedded/lib/ && \
cp /usr/lib/x86_64-linux-gnu/libapt-pkg.so.5.0 $TOOLS/embedded/lib/ && \
cp /usr/lib/x86_64-linux-gnu/libapt-pkg.so.5.0.2 $TOOLS/embedded/lib/ && \
cp /usr/lib/x86_64-linux-gnu/libapt-inst.so.2.0 $TOOLS/embedded/lib/ && \
cp /usr/lib/x86_64-linux-gnu/libapt-inst.so.2.0.0 $TOOLS/embedded/lib/ && \
cp /usr/lib/x86_64-linux-gnu/liblz4.so.1 $TOOLS/embedded/lib/ && \
cp /lib/x86_64-linux-gnu/liblzma.so.5 $TOOLS/embedded/lib/ && \
cp /lib/x86_64-linux-gnu/libbz2.so.1.0 $TOOLS/embedded/lib/ && \
cp /lib/x86_64-linux-gnu/libudev.so.1 $TOOLS/embedded/lib/ && \
cp /usr/lib/x86_64-linux-gnu/libzstd.so* $TOOLS/embedded/lib/ && \
cp /usr/lib/python2.7/dist-packages/apt_inst.x86_64-linux-gnu.so /tmp/apt_inst.so && \
cp /tmp/apt_inst.so $TOOLS/embedded/lib/python2.7/ && \
cp /usr/lib/python2.7/dist-packages/apt_pkg.x86_64-linux-gnu.so /tmp/apt_pkg.so && \
cp /tmp/apt_pkg.so $TOOLS/embedded/lib/python2.7/ && \
cp -r /usr/lib/python2.7/dist-packages/apt $TOOLS/embedded/lib/python2.7/ && \
cp -r /usr/lib/python2.7/dist-packages/aptsources $TOOLS/embedded/lib/python2.7/ && \
cp /lib/x86_64-linux-gnu/libsystemd.so.0.21.0 $TOOLS/lib && \
cp /lib/x86_64-linux-gnu/libsystemd.so.0 $TOOLS/lib && \
cp /usr/bin/apt-extracttemplates $TOOLS/embedded/bin

RUN \
ln -fs $TOOLS/embedded/bin/tar $TOOLS/ansible_tools_overlay/bin/tar && \
ln -fs $TOOLS/embedded/lib/libapt-private.so.0.0 $TOOLS/ansible_tools_overlay/lib/libapt-private.so.0.0 && \
ln -fs $TOOLS/embedded/lib/libapt-private.so.0.0 $TOOLS/ansible_tools_overlay/lib/libapt-private.so.0.0.0 && \
ln -fs $TOOLS/embedded/lib/libapt-pkg.so.5.0 $TOOLS/ansible_tools_overlay/lib/libapt-private.so.5.0 && \
ln -fs $TOOLS/embedded/lib/libapt-pkg.so.5.0.2 $TOOLS/ansible_tools_overlay/lib/libapt-private.so.5.0.2 && \
ln -fs $TOOLS/embedded/lib/libapt-inst.so.2.0 $TOOLS/ansible_tools_overlay/lib/libapt-inst.so.2.0 && \
ln -fs $TOOLS/embedded/lib/libapt-inst.so.2.0.0 $TOOLS/ansible_tools_overlay/lib/libapt-inst.so.2.0.0

# TODO: FIXME: https://github.com/flant/werf/issues/1798
# TODO: FIXME: try set nsswitch.conf file: http://www.linuxfromscratch.org/lfs/view/9.0-systemd-rc1/chapter06/glibc.html
# yum-utils package needed for ansible yum module to work
#RUN apt install -y yum-utils
#RUN \
#cp -r /usr/lib/python2.7/dist-packages/yumutils $TOOLS/embedded/lib/python2.7/ && \
#cp -r /usr/lib/python2.7/dist-packages/yum $TOOLS/embedded/lib/python2.7/ && \
#cp -r /usr/lib/python2.7/dist-packages/rpm $TOOLS/embedded/lib/python2.7/ && \
#ln -fs $TOOLS/embedded/lib/python2.7/rpm/_rpm.x86_64-linux-gnu.so $TOOLS/embedded/lib/python2.7/rpm/_rpm.so && \
#ln -fs $TOOLS/embedded/lib/python2.7/rpm/_rpmb.x86_64-linux-gnu.so $TOOLS/embedded/lib/python2.7/rpm/_rpmb.so && \
#ln -fs $TOOLS/embedded/lib/python2.7/rpm/_rpms.x86_64-linux-gnu.so $TOOLS/embedded/lib/python2.7/rpm/_rpms.so && \
#cp /usr/lib/x86_64-linux-gnu/librpmbuild.so.8.0.1 $TOOLS/embedded/lib/ && \
#cp /usr/lib/x86_64-linux-gnu/librpmsign.so.8.0.1 $TOOLS/embedded/lib/ && \
#cp /usr/lib/x86_64-linux-gnu/librpm.so.8.0.1 $TOOLS/embedded/lib/ && \
#cp /usr/lib/x86_64-linux-gnu/librpm.so.8 $TOOLS/embedded/lib/ && \
#cp /usr/lib/x86_64-linux-gnu/librpmio.so.8 $TOOLS/embedded/lib/ && \
#cp /usr/lib/x86_64-linux-gnu/librpmsign.so.8 $TOOLS/embedded/lib/ && \
#cp /usr/lib/x86_64-linux-gnu/librpmio.so.8.0.1 $TOOLS/embedded/lib/ && \
#cp /usr/lib/x86_64-linux-gnu/librpmbuild.so.8 $TOOLS/embedded/lib/ && \
#cp /lib/x86_64-linux-gnu/libcap.so.2.25 $TOOLS/lib && \
#cp /lib/x86_64-linux-gnu/libcap.so.2 $TOOLS/lib && \
#cp /usr/lib/x86_64-linux-gnu/liblua5.2.so.0 $TOOLS/lib && \
#cp /usr/lib/x86_64-linux-gnu/libdb-5.3.so $TOOLS/lib && \
#cp /usr/lib/x86_64-linux-gnu/libelf.so.1 $TOOLS/lib && \
#cp /usr/lib/x86_64-linux-gnu/libplc4.so $TOOLS/lib && \
#cp /usr/lib/x86_64-linux-gnu/libplds4.so $TOOLS/lib && \
#cp /usr/lib/x86_64-linux-gnu/libnspr4.so $TOOLS/lib && \
#cp /lib/x86_64-linux-gnu/libnss_nis.so.2 $TOOLS/lib && \
#cp /lib/x86_64-linux-gnu/libnss_compat.so.2 $TOOLS/lib && \
#cp /lib/x86_64-linux-gnu/libnss_nis-2.27.so $TOOLS/lib && \
#cp /lib/x86_64-linux-gnu/libnss_files-2.27.so $TOOLS/lib && \
#cp /lib/x86_64-linux-gnu/libnss_nisplus-2.27.so $TOOLS/lib && \
#cp /lib/x86_64-linux-gnu/libnss_nisplus.so.2 $TOOLS/lib && \
#cp /lib/x86_64-linux-gnu/libnss_hesiod-2.27.so $TOOLS/lib && \
#cp /lib/x86_64-linux-gnu/libnss_dns.so.2 $TOOLS/lib && \
#cp /lib/x86_64-linux-gnu/libnss_hesiod.so.2 $TOOLS/lib && \
#cp /lib/x86_64-linux-gnu/libnss_dns-2.27.so $TOOLS/lib && \
#cp /lib/x86_64-linux-gnu/libnss_files.so.2 $TOOLS/lib && \
#cp /lib/x86_64-linux-gnu/libnss_compat-2.27.so $TOOLS/lib && \
#cp /usr/lib/x86_64-linux-gnu/libnss_files.so $TOOLS/lib && \
#cp /usr/lib/x86_64-linux-gnu/libnss_dns.so $TOOLS/lib && \
#cp /usr/lib/x86_64-linux-gnu/libnss_nis.so $TOOLS/lib && \
#cp /usr/lib/x86_64-linux-gnu/libnss_nisplus.so $TOOLS/lib && \
#cp /usr/lib/x86_64-linux-gnu/libnss3.so $TOOLS/lib && \
#cp /usr/lib/x86_64-linux-gnu/libnss_hesiod.so $TOOLS/lib && \
#cp /usr/lib/x86_64-linux-gnu/libnss_compat.so $TOOLS/lib && \
#cp /usr/lib/x86_64-linux-gnu/libnssutil3.so $TOOLS/lib && \
#cp -r /usr/lib/x86_64-linux-gnu/nss $TOOLS/lib
#
#RUN \
#ln -fs $TOOLS/embedded/lib/librpmbuild.so.8.0.1 $TOOLS/ansible_tools_overlay/lib/librpmbuild.so.8.0.1 && \
#ln -fs $TOOLS/embedded/lib/librpmsign.so.8.0.1 $TOOLS/ansible_tools_overlay/lib/librpmsign.so.8.0.1 && \
#ln -fs $TOOLS/embedded/lib/librpm.so.8.0.1 $TOOLS/ansible_tools_overlay/lib/librpm.so.8.0.1 && \
#ln -fs $TOOLS/embedded/lib/librpm.so.8 $TOOLS/ansible_tools_overlay/lib/librpm.so.8 && \
#ln -fs $TOOLS/embedded/lib/librpmio.so.8 $TOOLS/ansible_tools_overlay/lib/librpmio.so.8 && \
#ln -fs $TOOLS/embedded/lib/librpmsign.so.8 $TOOLS/ansible_tools_overlay/lib/librpmsign.so.8 && \
#ln -fs $TOOLS/embedded/lib/librpmio.so.8.0.1 $TOOLS/ansible_tools_overlay/lib/librpmio.so.8.0.1 && \
#ln -fs $TOOLS/embedded/lib/librpmbuild.so.8 $TOOLS/ansible_tools_overlay/lib/librpmbuild.so.8

# TODO: FIXME move to glibc
WORKDIR $LFS/sources/glibc/build
RUN mkdir /.werf/stapel/lib/locale
RUN /.werf/stapel/bin/localedef -i POSIX -f UTF-8 C.UTF-8 || true

# Cleanup stapel: only runtime libs and binaries are needed
RUN rm -rf \
$TOOLS/LICENSE* \
$TOOLS/version-manifest.* \
$TOOLS/include \
$TOOLS/share/aclocal \
$TOOLS/share/common-lisp \
$TOOLS/share/doc \
$TOOLS/share/gcc-8.2.0 \
$TOOLS/share/gnupg \
$TOOLS/share/info \
$TOOLS/share/libgpg-error \
$TOOLS/share/man \
$TOOLS/share/locale \
$TOOLS/x86_64-lfs-linux-gnu/include \
$TOOLS/x86_64-lfs-linux-gnu/lib \
$TOOLS/lib/*.a \
$TOOLS/lib/*.o \
$TOOLS/lib64/*.a \
$TOOLS/lib/gcc \
$TOOLS/libexec/gcc \
$TOOLS/var \
$TOOLS/embedded/lib/*.a \
$TOOLS/embedded/include \
$TOOLS/embedded/lib/python2.7/test \
$TOOLS/embedded/lib/python2.7/site-packages/ansible/galaxy \
$TOOLS/embedded/lib/python2.7/site-packages/ansible/modules/clustering \
$TOOLS/embedded/lib/python2.7/site-packages/ansible/modules/source_control \
$TOOLS/embedded/lib/python2.7/site-packages/ansible/modules/notification \
$TOOLS/embedded/lib/python2.7/site-packages/ansible/modules/web_infrastructure \
$TOOLS/embedded/lib/python2.7/site-packages/ansible/modules/remote_management \
$TOOLS/embedded/lib/python2.7/site-packages/ansible/modules/monitoring \
$TOOLS/embedded/lib/python2.7/site-packages/ansible/modules/windows \
$TOOLS/embedded/lib/python2.7/site-packages/ansible/modules/storage \
$TOOLS/embedded/lib/python2.7/site-packages/ansible/modules/cloud \
$TOOLS/embedded/lib/python2.7/site-packages/ansible/modules/network \
$TOOLS/embedded/lib/python2.7/site-packages/pip \
$TOOLS/embedded/lib/python2.7/site-packages/setuptools \
$TOOLS/embedded/lib/python2.7/site-packages/*.dist-info \
$TOOLS/embedded/share/aclocal \
$TOOLS/embedded/share/doc \
$TOOLS/embedded/share/info \
$TOOLS/embedded/share/man \
$TOOLS/embedded/share/local \
$TOOLS/embedded/share/locale \
$TOOLS/embedded/var \
$TOOLS/embedded/info \
$TOOLS/embedded/man

RUN cp $TOOLS/share/i18n/charmaps/UTF-8.gz /tmp && \
rm -rf $TOOLS/share/i18n/charmaps/ && \
mkdir -p $TOOLS/share/i18n/charmaps/ && \
mv /tmp/UTF-8.gz $TOOLS/share/i18n/charmaps/

RUN cp $TOOLS/share/i18n/locales/POSIX /tmp && \
rm -rf $TOOLS/share/i18n/locales/ && \
mkdir -p $TOOLS/share/i18n/locales/ && \
mv /tmp/POSIX $TOOLS/share/i18n/locales/

RUN cp $TOOLS/x86_64-lfs-linux-gnu/bin/ld /tmp && \
rm -rf $TOOLS/x86_64-lfs-linux-gnu/bin && \
mkdir -p $TOOLS/x86_64-lfs-linux-gnu/bin && \
mv /tmp/ld $TOOLS/x86_64-lfs-linux-gnu/bin

RUN cp $TOOLS/lib/gconv/UNICODE.so /tmp && \
rm -rf $TOOLS/lib/gconv && \
mkdir -p $TOOLS/lib/gconv && \
mv /tmp/UNICODE.so $TOOLS/lib/gconv

RUN find $TOOLS/embedded/lib/python2.7 -name *.py[oc] | xargs rm

RUN mkdir /tmp/bin && \
cp $TOOLS/bin/gpg* /tmp/bin && \
cp $TOOLS/bin/gnutls* /tmp/bin && \
cp $TOOLS/bin/nettle* /tmp/bin && \
cp $TOOLS/bin/libassuan* /tmp/bin && \
cp $TOOLS/bin/libgcrypt* /tmp/bin && \
cp $TOOLS/bin/npth* /tmp/bin && \
rm -rf $TOOLS/bin && \
mv /tmp/bin $TOOLS/bin

# Import tools

FROM scratch as final
CMD ["no-such-command"]
ENV TOOLS=/.werf/stapel
COPY --from=0 $TOOLS/ansible_tools_overlay $TOOLS/ansible_tools_overlay
COPY --from=0 $TOOLS/libexec $TOOLS/libexec
COPY --from=0 $TOOLS/x86_64-lfs-linux-gnu $TOOLS/x86_64-lfs-linux-gnu
COPY --from=0 $TOOLS/sbin $TOOLS/sbin
COPY --from=0 $TOOLS/bin $TOOLS/bin
COPY --from=0 $TOOLS/lib64 $TOOLS/lib64
COPY --from=0 $TOOLS/lib $TOOLS/lib
COPY --from=0 $TOOLS/share $TOOLS/share
COPY --from=0 $TOOLS/embedded/bin $TOOLS/embedded/bin
COPY --from=0 $TOOLS/embedded/etc $TOOLS/embedded/etc
COPY --from=0 $TOOLS/embedded/lib $TOOLS/embedded/lib
COPY --from=0 $TOOLS/embedded/libexec $TOOLS/embedded/libexec
COPY --from=0 $TOOLS/embedded/sbin $TOOLS/embedded/sbin
COPY --from=0 $TOOLS/embedded/share $TOOLS/embedded/share
COPY --from=0 $TOOLS/embedded/ssl $TOOLS/embedded/ssl
