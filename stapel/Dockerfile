FROM ubuntu:18.04

RUN rm /bin/sh && ln -s /bin/bash /bin/sh
RUN >/etc/profile && >/root/.profile
RUN echo "set +h" >> /root/.profile && \
echo "umask 022" >> /root/.profile

SHELL ["/bin/bash", "-lc"]

RUN apt update && apt install -y \
build-essential wget curl gawk flex bison bzip2 liblzma5 texinfo file \
gettext python python3 curl git fakeroot gettext gpg ruby ruby-bundler \
ruby-dev make file m4 xz-utils texlive

RUN git config --global user.name flant && git config --global user.email 256@flant.com

ENV LFS=/mnt/lfs
ENV TOOLS=/.werf/stapel
ENV LFS_TGT=x86_64-lfs-linux-gnu

RUN mkdir -pv $LFS$TOOLS && mkdir -pv $LFS/sources && chmod -v a+wt $LFS/sources
ADD ./wget-list $LFS/sources/wget-list
ADD ./md5sums $LFS/sources/md5sums
RUN wget --input-file=$LFS/sources/wget-list --continue --directory-prefix=$LFS/sources
RUN bash -c "pushd $LFS/sources && md5sum -c $LFS/sources/md5sums && popd"
ADD version-check.sh $LFS/sources/version-check.sh
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

ADD ./gcc-before-configure.sh $LFS/sources/gcc-before-configure.sh
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

ADD ./omnibus /omnibus
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

# python-apt package will install all libs in docker.from image.
# python-apt will be installed automatically by ansible apt module on first run.
# flant/werf-stapel offers support for ansible-apt-module only for ubuntu:14.04 ubuntu:16.04 ubuntu:18.04
ENV PATH=/sbin:/bin:/usr/sbin:/usr/bin:/usr/local/sbin:/usr/local/bin
RUN apt install -y python-apt libzstd1
RUN \
cp /usr/lib/x86_64-linux-gnu/libapt-inst.so.2.0 $TOOLS/embedded/lib/ && \
cp /usr/lib/x86_64-linux-gnu/libapt-pkg.so.5.0 $TOOLS/embedded/lib/ && \
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
cp -r /usr/lib/python2.7/dist-packages/aptsources $TOOLS/embedded/lib/python2.7/

# Cleanup stapel: only runtime libs and binaries are needed
RUN rm -rf \
$TOOLS/LICENSE* \
$TOOLS/version-manifest.* \
$TOOLS/sbin \
$TOOLS/bin \
$TOOLS/include \
$TOOLS/x86_64-lfs-linux-gnu/include \
$TOOLS/x86_64-lfs-linux-gnu/lib \
$TOOLS/lib/*.a \
$TOOLS/lib/*.o \
$TOOLS/lib64/*.a \
$TOOLS/lib/gcc \
$TOOLS/libexec/gcc \
$TOOLS/share \
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
$TOOLS/embedded/share/doc \
$TOOLS/embedded/share/info \
$TOOLS/embedded/share/man \
$TOOLS/embedded/share/local \
$TOOLS/embedded/var \
$TOOLS/embedded/info \
$TOOLS/embedded/man

RUN cp $TOOLS/x86_64-lfs-linux-gnu/bin/ld /tmp && \
rm -rf $TOOLS/x86_64-lfs-linux-gnu/bin && \
mkdir -p $TOOLS/x86_64-lfs-linux-gnu/bin && \
mv /tmp/ld $TOOLS/x86_64-lfs-linux-gnu/bin

RUN cp $TOOLS/lib/gconv/UNICODE.so /tmp && \
rm -rf $TOOLS/lib/gconv && \
mkdir -p $TOOLS/lib/gconv && \
mv /tmp/UNICODE.so $TOOLS/lib/gconv

RUN find $TOOLS/embedded/lib/python2.7 -name *.py[oc] | xargs rm

# Import tools

FROM scratch
CMD ["no-such-command"]
COPY --from=0 /.werf/stapel /.werf/stapel
