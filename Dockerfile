# Build tools

FROM ubuntu:16.04

RUN rm /bin/sh && ln -s /bin/bash /bin/sh
RUN >/etc/profile && >/root/.profile
SHELL ["/bin/bash", "-lc"]

RUN apt update && apt install -y build-essential wget curl gawk flex bison bzip2 liblzma5 texinfo file

ENV LFS=/mnt/lfs
ENV TOOLS=/.dapp/deps/toolchain/0.1.0
ENV LFS_TGT=x86_64-lfs-linux-gnu

RUN mkdir -pv $LFS$TOOLS && mkdir -pv $LFS/sources && chmod -v a+wt $LFS/sources
ADD ./wget-list $LFS/sources/wget-list
ADD ./md5sums $LFS/sources/md5sums
RUN wget --input-file=$LFS/sources/wget-list --continue --directory-prefix=$LFS/sources
RUN bash -c "pushd $LFS/sources && md5sum -c $LFS/sources/md5sums && popd"
ADD version-check.sh $LFS/sources/version-check.sh
RUN $LFS/sources/version-check.sh

RUN ln -sv $LFS/.dapp /
RUN groupadd lfs && useradd -s /bin/bash -g lfs -m -k /dev/null lfs
RUN chown -R lfs:lfs $LFS
RUN >/home/lfs/.profile && \
echo "set +h" >> /home/lfs/.profile && \
echo "umask 022" >> /home/lfs/.profile

USER lfs
ENV LC_ALL=POSIX
ENV PATH=$TOOLS/bin:/bin:/usr/bin
ENV MAKEFLAGS='-j 5'

RUN cd $LFS/sources/ && \
mkdir binutils && \
tar xf binutils-*.tar.bz2 -C binutils --strip-components 1 && \
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

# gcc pass 1

ADD ./gcc-before-configure.sh $LFS/sources/gcc-before-configure.sh
RUN cd $LFS/sources/ && \
mkdir gcc && \
tar xf gcc-*.tar.xz -C gcc --strip-components 1 && \
mkdir gcc/mpfr && \
tar xf mpfr*.tar.xz -C gcc/mpfr --strip-components 1 && \
mkdir gcc/gmp && \
tar xf gmp*.tar.xz -C gcc/gmp --strip-components 1 && \
mkdir gcc/mpc && \
tar xf mpc*.tar.gz -C gcc/mpc --strip-components 1 && \
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
tar xf linux*.tar.xz -C linux --strip-components 1
WORKDIR $LFS/sources/linux
RUN make mrproper && \
make INSTALL_HDR_PATH=dest headers_install && \
cp -rv dest/include/* $TOOLS/include

# glibc pass 1

RUN cd $LFS/sources/ && \
mkdir glibc && \
tar xf glibc*.tar.xz -C glibc --strip-components 1 && \
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

RUN cd $LFS/sources/ && \
rm -rf gcc && \
mkdir gcc && \
tar xf gcc-*.tar.xz -C gcc --strip-components 1 && \
mkdir gcc/mpfr && \
tar xf mpfr*.tar.xz -C gcc/mpfr --strip-components 1 && \
mkdir gcc/gmp && \
tar xf gmp*.tar.xz -C gcc/gmp --strip-components 1 && \
mkdir gcc/mpc && \
tar xf mpc*.tar.gz -C gcc/mpc --strip-components 1 && \
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
--with-gxx-include-dir=$TOOLS/$LFS_TGT/include/c++/7.2.0
WORKDIR $LFS/sources/gcc/build
RUN make
RUN make install

# binutils pass 2

RUN cd $LFS/sources/ && \
rm -rf binutils && \
mkdir binutils && \
tar xf binutils-*.tar.bz2 -C binutils --strip-components 1 && \
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

# gcc pass 2

RUN cd $LFS/sources/ && \
rm -rf gcc && \
mkdir gcc && \
tar xf gcc-*.tar.xz -C gcc --strip-components 1 && \
mkdir gcc/mpfr && \
tar xf mpfr*.tar.xz -C gcc/mpfr --strip-components 1 && \
mkdir gcc/gmp && \
tar xf gmp*.tar.xz -C gcc/gmp --strip-components 1 && \
mkdir gcc/mpc && \
tar xf mpc*.tar.gz -C gcc/mpc --strip-components 1 && \
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

# Import tools into dappdeps/toolchain scratch

FROM scratch
COPY --from=0 /.dapp/deps/toolchain/0.1.0 /.dapp/deps/toolchain/0.1.0
