FROM ubuntu:18.04 AS base

ARG TARGETARCH=amd64
ARG LFS_TGT=x86_64-lfs-linux-gnu

RUN rm /bin/sh && ln -s /bin/bash /bin/sh
RUN >/etc/profile && >/root/.profile
RUN echo "set +h" >> /root/.profile && \
  echo "umask 022" >> /root/.profile

SHELL ["/bin/bash", "-lc"]

ENV DEBIAN_FRONTEND=noninteractive

RUN apt update && apt install -y \
  build-essential wget curl gawk flex bison bzip2 liblzma5 texinfo file \
  gettext python python3 curl git fakeroot gettext gpg ruby ruby-bundler \
  ruby-dev make file m4 xz-utils texlive vim rsync

RUN git config --global user.name flant && git config --global user.email 256@flant.com

ENV LFS=/mnt/lfs
ENV TOOLS=/.werf/stapel
ENV LFS_TGT=$LFS_TGT

RUN mkdir -pv $LFS$TOOLS && mkdir -pv $LFS/sources && chmod -v a+wt $LFS/sources
ADD stapel/wget-list-before-omnibus $LFS/sources/wget-list-before-omnibus
RUN wget --input-file=$LFS/sources/wget-list-before-omnibus --continue --directory-prefix=$LFS/sources || true
ADD stapel/wget-list-before-omnibus.md5sums $LFS/sources/wget-list-before-omnibus.md5sums
RUN bash -c "pushd $LFS/sources && md5sum -c $LFS/sources/wget-list-before-omnibus.md5sums && popd"
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
  --enable-initfini-array \
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
  --enable-kernel=3.2 \
  --with-headers=$TOOLS/include \
  libc_cv_forced_unwind=yes \
  libc_cv_c_cleanup=yes \
  libc_cv_slibdir=$TOOLS/lib
WORKDIR $LFS/sources/glibc/build
RUN make
RUN make install
RUN mkdir /.werf/stapel/lib/locale
RUN /.werf/stapel/bin/localedef -i POSIX -f UTF-8 C.UTF-8 || true

RUN echo "Libstdc++ pass 1" && cd $LFS/sources/ && \
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
  --disable-libstdcxx-pch \
  --disable-libstdcxx-threads \
  --with-gxx-include-dir=$TOOLS/$LFS_TGT/include/c++/11.2.0
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
  --host=$LFS_TGT \
  --disable-nls \
  --enable-shared \
  --disable-werror \
  --enable-64-bit-bfd \
  --with-lib-path=$TOOLS/lib

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
  --build=$(../config.guess) \
  --host=$LFS_TGT \
  --target=$LFS_TGT \
  --prefix=$TOOLS \
  CC_FOR_TARGET=$LFS_TGT-gcc \
  --with-build-sysroot=$LFS \
  --enable-initfini-array \
  --disable-nls \
  --disable-multilib \
  --disable-decimal-float \
  --disable-libatomic \
  --disable-libgomp \
  --disable-libquadmath \
  --disable-libssp \
  --disable-libvtv \
  --disable-libstdcxx \
  --disable-lto \
  --with-local-prefix=$TOOLS \
  --with-sysroot=$LFS \
  --with-native-system-header-dir=$TOOLS/include \
  --oldincludedir=$TOOLS/include \
  --enable-languages=c,c++
WORKDIR $LFS/sources/gcc/build
RUN make
RUN make DESTDIR=$LFS install
RUN ln -sv gcc $TOOLS/bin/cc
RUN for tool in $(ls /.werf/stapel/bin/ | grep $LFS_TGT) ; do ln -fs /.werf/stapel/bin/$tool /.werf/stapel/bin/$(echo $tool | sed -e "s|$LFS_TGT-||") ; done

# libffi
RUN echo "libffi" && cd $LFS/sources && \
  mkdir libffi && \
  tar xf libffi-*.tar.* -C libffi --strip-components 1
WORKDIR $LFS/sources/libffi
RUN PATH=/bin:/usr/bin:/usr/local/bin ./configure --prefix=$TOOLS --disable-static
RUN PATH=/bin:/usr/bin:/usr/local/bin make
RUN PATH=/bin:/usr/bin:/usr/local/bin make install

ENV PATH=/sbin:/usr/sbin:/usr/local/sbin:/bin:/usr/bin:/usr/local/bin
RUN apt install -y libssl-dev autoconf automake libffi-dev libgdbm-dev libncurses5-dev libsqlite3-dev libtool libyaml-dev pkg-config sqlite3 zlib1g-dev libreadline-dev libssl-dev
RUN curl -sSL https://rvm.io/mpapis.asc | gpg --import -
RUN curl -sSL https://rvm.io/pkuczynski.asc | gpg --import -
RUN curl -sSL https://get.rvm.io -o /tmp/rvm.sh && cat /tmp/rvm.sh | bash -s stable
RUN source /etc/profile.d/rvm.sh && rvm install 2.7

ADD stapel/omnibus /omnibus
WORKDIR /omnibus
ENV BUNDLE_GEMFILE=/omnibus/Gemfile
RUN source /etc/profile.d/rvm.sh && bundle install --without development
RUN source /etc/profile.d/rvm.sh && \
  omnibus_software_path="$(bundle show omnibus-software)" && \
  find "$omnibus_software_path/config/software" -name '*.rb' -type f -print0 | xargs -0 sed -i \
  -e 's|https://ftp.gnu.org/pub/gnu/|https://ftpmirror.gnu.org/gnu/|g' \
  -e 's|https://ftp.gnu.org/gnu/|https://ftpmirror.gnu.org/gnu/|g' \
  -e 's|http://ftp.gnu.org/pub/gnu/|https://ftpmirror.gnu.org/gnu/|g' \
  -e 's|http://ftp.gnu.org/gnu/|https://ftpmirror.gnu.org/gnu/|g' \
  -e 's|ftp://ftp.gnu.org/pub/gnu/|https://ftpmirror.gnu.org/gnu/|g' \
  -e 's|ftp://ftp.gnu.org/gnu/|https://ftpmirror.gnu.org/gnu/|g' \
  -e 's|https://ftpmirror.gnu.org/pub/gnu/|https://ftpmirror.gnu.org/gnu/|g' \
  -e 's|http://ftpmirror.gnu.org/gnu/|https://ftpmirror.gnu.org/gnu/|g'

ENV PATH=$TOOLS/$LFS_TGT/bin:$TOOLS/bin:$PATH

# Dpkg-architecture binary will make python-omnibus-package fail to build,
# because of python setup.py, which hardcodes /usr/include/... into preceeding include paths,
# in the case when dpkg-architecture is available in system: https://github.com/python/cpython/blob/master/setup.py#L485
# It is needed to remove that binary before omnibus-building.
RUN mv $(which dpkg-architecture) /tmp/dpkg-architecture

#ENV CC=$LFS_TGT-gcc
#ENV CXX=$LFS_TGT-g++
#ENV AR=$LFS_TGT-ar
#ENV RANLIB=$LFS_TGT-ranlib
ENV CC_FOR_TARGET=$LFS_TGT-gcc
ENV PKG_CONFIG_PATH="/.werf/stapel/lib/pkgconfig:/.werf/stapel/embedded/lib/pkgconfig"

RUN source /etc/profile.d/rvm.sh && \
  if [ "$TARGETARCH" = "arm64" ]; then \
  export ac_cv_build=aarch64-unknown-linux-gnu ac_cv_host=aarch64-unknown-linux-gnu ac_cv_target=aarch64-unknown-linux-gnu; \
  fi && \
  env CC=/usr/bin/gcc CXX=/usr/bin/g++ bundle exec omnibus build -o append_timestamp:false werf-stapel

# Fix ELF interpreter paths for arm64: omnibus binaries are linked against the system Ubuntu glibc
# (interpreter /lib/ld-linux-aarch64.so.1), which is not present in build containers using the
# stapel volume. Patch every arm64 ELF binary to use /.werf/stapel/lib/ld-linux-aarch64.so.1,
# and copy the system linker there so it is accessible via the volume mount.
RUN if [ "$TARGETARCH" = "arm64" ]; then \
  apt-get install -y patchelf && \
  cp /lib/aarch64-linux-gnu/ld-linux-aarch64.so.1 $TOOLS/lib/ld-linux-aarch64.so.1 && \
  find $TOOLS/embedded/bin $TOOLS/bin $TOOLS/embedded/sbin $TOOLS/sbin $TOOLS/embedded/libexec -type f 2>/dev/null | while read f; do \
  if file "$f" 2>/dev/null | grep -q 'ELF.*ARM aarch64.*dynamically linked'; then \
  patchelf --set-interpreter $TOOLS/lib/ld-linux-aarch64.so.1 "$f" 2>/dev/null || true; \
  fi; \
  done; \
  fi

ENV PATH=/sbin:/usr/sbin:/usr/local/sbin:/bin:/usr/bin:/usr/local/bin
ENV CC=/usr/bin/gcc
ENV CXX=/usr/bin/g++
ENV AR=/usr/bin/ar
ENV RANLIB=/usr/bin/ranlib
ENV PKG_CONFIG_PATH=$TOOLS/lib/pkgconfig

RUN if [ "$TARGETARCH" = "arm64" ]; then \
  echo "export LIBC_SEARCH_PATHS='-L/lib/aarch64-linux-gnu -L/usr/lib/aarch64-linux-gnu'" > /etc/profile.d/libc_paths.sh; \
  else \
  echo "export LIBC_SEARCH_PATHS='-L/lib/x86_64-linux-gnu -L/usr/lib/x86_64-linux-gnu'" > /etc/profile.d/libc_paths.sh; \
  fi && \
  chmod +x /etc/profile.d/libc_paths.sh

ADD stapel/wget-list-after-omnibus $LFS/sources/wget-list-after-omnibus
ADD stapel/wget-list-after-omnibus.md5sums $LFS/sources/wget-list-after-omnibus.md5sums
RUN wget --input-file=$LFS/sources/wget-list-after-omnibus --continue --directory-prefix=$LFS/sources || true
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
RUN ./configure --prefix=$TOOLS --with-gpg-error-prefix=$TOOLS
RUN . /etc/profile.d/libc_paths.sh && make LDFLAGS="$LIBC_SEARCH_PATHS -Wl,-rpath,$TOOLS/lib"
RUN make install

# libassuan
RUN echo "libassuan" && cd $LFS/sources && \
  mkdir libassuan && \
  tar xf libassuan*.tar.* -C libassuan --strip-components 1
WORKDIR $LFS/sources/libassuan
RUN ./configure --prefix=$TOOLS --with-libgpg-error-prefix=$TOOLS
RUN . /etc/profile.d/libc_paths.sh && make LDFLAGS="$LIBC_SEARCH_PATHS -Wl,-rpath,$TOOLS/lib"
RUN make install

# libksba
RUN echo "libksba" && cd $LFS/sources && \
  mkdir libksba && \
  tar xf libksba*.tar.* -C libksba --strip-components 1
WORKDIR $LFS/sources/libksba
RUN ./configure --prefix=$TOOLS --with-libgpg-error-prefix=$TOOLS
RUN . /etc/profile.d/libc_paths.sh && make LDFLAGS="$LIBC_SEARCH_PATHS -Wl,-rpath,$TOOLS/lib"
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
RUN ./configure --prefix=$TOOLS --disable-static --disable-documentation
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
RUN . /etc/profile.d/libc_paths.sh && \
  CFLAGS="-I$TOOLS/include" \
  LDFLAGS="$LIBC_SEARCH_PATHS -L$TOOLS/lib -Wl,-rpath,$TOOLS/lib" \
  PKG_CONFIG_PATH="$TOOLS/lib/pkgconfig" \
  ./configure --prefix=$TOOLS --disable-guile --disable-cxx --without-p11-kit --with-included-libtasn1 --with-included-unistring
RUN . /etc/profile.d/libc_paths.sh && make LDFLAGS="$LIBC_SEARCH_PATHS -L$TOOLS/lib -Wl,-rpath,$TOOLS/lib"
RUN make install

# gnupg
RUN echo "gnupg" && cd $LFS/sources && \
  mkdir gnupg && \
  tar xf gnupg*.tar.* -C gnupg --strip-components 1
WORKDIR $LFS/sources/gnupg
RUN perl -0pi -e 's@static const char \*\nget_signal_name\( int signum \)\n\{\n.*?\n\}\n#endif /\*!HAVE_DOSISH_SYSTEM\*/@static const char *\nget_signal_name( int signum )\n{\n  (void)signum;\n  return NULL;\n}\n#endif /*!HAVE_DOSISH_SYSTEM*/@s' common/signal.c
RUN PKG_CONFIG_PATH="$TOOLS/lib/pkgconfig" \
  ./configure --prefix=$TOOLS --enable-symcryptrun --localstatedir=/.werf/stapel/var --sysconfdir=/etc --disable-sqlite --disable-dirmngr \
  --with-libgpg-error-prefix=$TOOLS --with-libgcrypt-prefix=$TOOLS --with-libassuan-prefix=$TOOLS \
  --with-ksba-prefix=$TOOLS --with-npth-prefix=$TOOLS
RUN . /etc/profile.d/libc_paths.sh && make LDFLAGS="$LIBC_SEARCH_PATHS -L$TOOLS/lib -Wl,-rpath,$TOOLS/lib"
RUN make install

# util-linux
RUN echo "util-linux" && cd $LFS/sources && \
  mkdir util-linux && \
  tar xf util-linux-*.tar.* -C util-linux --strip-components 1
WORKDIR $LFS/sources/util-linux
RUN ./configure --prefix=$TOOLS --without-python --disable-all-programs --disable-makeinstall-chown --without-systemdsystemunitdir PKG_CONFIG=""
RUN . /etc/profile.d/libc_paths.sh && make LDFLAGS="$LIBC_SEARCH_PATHS -Wl,-rpath,$TOOLS/lib"
RUN make install

# glib
RUN PATH=/bin:/usr/bin:/usr/local/bin:/sbin:/usr/sbin:/usr/local/sbin apt update
RUN PATH=/bin:/usr/bin:/usr/local/bin:/sbin:/usr/sbin:/usr/local/sbin apt install -y python3-pip xsltproc ninja-build libpcre3-dev
RUN PATH=/bin:/usr/bin:/usr/local/bin:/sbin:/usr/sbin:/usr/local/sbin pip3 install meson
RUN echo "glib" && cd $LFS/sources && \
  mkdir glib && \
  tar xf glib-*.tar* -C glib --strip-components 1
WORKDIR $LFS/sources/glib
RUN mkdir -p /tmp/glib-linker-stash && \
  shopt -s nullglob && \
  for p in $TOOLS/lib/libresolv.so* $TOOLS/lib/libdl.so*; do mv "$p" /tmp/glib-linker-stash/; done && \
  rm -rf _build && \
  PKG_CONFIG_PATH="$TOOLS/lib/pkgconfig" \
  CC="/usr/bin/gcc" \
  CXX="/usr/bin/g++" \
  AR="/usr/bin/ar" \
  RANLIB="/usr/bin/ranlib" \
  CFLAGS="-pthread" \
  LDFLAGS="-pthread" \
  meson _build -Diconv=auto -Dman=false -Dtests=false -Dinstalled_tests=false -Dforce_posix_threads=true --prefix=$TOOLS -Dselinux=disabled
RUN ninja -C _build -v
RUN ninja -C _build install
RUN shopt -s nullglob && \
  for p in /tmp/glib-linker-stash/*; do mv "$p" $TOOLS/lib/; done && \
  rmdir /tmp/glib-linker-stash

# libxml2
RUN echo "libxml2" && cd $LFS/sources && \
  mkdir libxml2 && \
  tar xf libxml2-*.tar.* -C libxml2 --strip-components 1
WORKDIR $LFS/sources/libxml2
RUN ./configure --prefix=$TOOLS --disable-static --with-history --without-python --with-iconv=no
RUN . /etc/profile.d/libc_paths.sh && make LDFLAGS="$LIBC_SEARCH_PATHS -Wl,-rpath,$TOOLS/lib"
RUN make install

# Restore PATH *not to* use stapel build tools anymore
ENV PATH=/sbin:/bin:/usr/sbin:/usr/bin:/usr/local/sbin:/usr/local/bin

# Install golang
RUN wget https://dl.google.com/go/go1.17.linux-${TARGETARCH}.tar.gz -O /tmp/go.tar.gz
RUN tar -C /usr/local -xzf /tmp/go.tar.gz
ENV PATH=$PATH:/usr/local/go/bin
ENV GO111MODULE=on

ENV CC=
ENV CXX=
ENV AR=
ENV RANLIB=
ENV CC_FOR_TARGET=
ENV PKG_CONFIG_PATH=



# TODO: FIXME: https://github.com/werf/werf/issues/1798
# TODO: FIXME: try set nsswitch.conf file: http://www.linuxfromscratch.org/lfs/view/9.0-systemd-rc1/chapter06/glibc.html

ENV CC=$LFS_TGT-gcc
ENV CXX=$LFS_TGT-g++
ENV AR=$LFS_TGT-ar
ENV RANLIB=$LFS_TGT-ranlib
ENV CC_FOR_TARGET=$LFS_TGT-gcc
ENV PKG_CONFIG_PATH="/.werf/stapel/lib/pkgconfig:/.werf/stapel/embedded/lib/pkgconfig"

ENV PATH=$TOOLS/$LFS_TGT/bin:$TOOLS/bin:$PATH



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
  $TOOLS/$LFS_TGT/include \
  $TOOLS/$LFS_TGT/lib \
  $TOOLS/lib/*.a \
  $TOOLS/lib/*.o \
  $TOOLS/lib64/*.a \
  $TOOLS/lib/gcc \
  $TOOLS/libexec/gcc \
  $TOOLS/var \
  $TOOLS/embedded/lib/*.a \
  $TOOLS/embedded/include \
  $TOOLS/embedded/lib/python2.7/test \
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

RUN cp $TOOLS/$LFS_TGT/bin/ld /tmp && \
  rm -rf $TOOLS/$LFS_TGT/bin && \
  mkdir -p $TOOLS/$LFS_TGT/bin && \
  mv /tmp/ld $TOOLS/$LFS_TGT/bin

RUN cp $TOOLS/lib/gconv/UNICODE.so /tmp && \
  rm -rf $TOOLS/lib/gconv && \
  mkdir -p $TOOLS/lib/gconv && \
  mv /tmp/UNICODE.so $TOOLS/lib/gconv

RUN if [ -d "$TOOLS/embedded/lib/python2.7" ]; then find "$TOOLS/embedded/lib/python2.7" -name '*.py[oc]' -delete; fi


RUN mkdir /tmp/bin && \
  cp $TOOLS/bin/gpg* /tmp/bin && \
  cp $TOOLS/bin/gnutls* /tmp/bin && \
  cp $TOOLS/bin/nettle* /tmp/bin && \
  cp $TOOLS/bin/libassuan* /tmp/bin && \
  cp $TOOLS/bin/libgcrypt* /tmp/bin && \
  cp $TOOLS/bin/npth* /tmp/bin && \
  cp $TOOLS/bin/dirmngr* /tmp/bin 2>/dev/null || true && \
  rm -rf $TOOLS/bin && \
  mv /tmp/bin $TOOLS/bin

# Import tools

FROM scratch AS final
CMD ["no-such-command"]
ENV TOOLS=/.werf/stapel
ARG LFS_TGT=x86_64-lfs-linux-gnu
COPY --from=0 $TOOLS/libexec $TOOLS/libexec
COPY --from=0 $TOOLS/$LFS_TGT $TOOLS/$LFS_TGT
COPY --from=0 $TOOLS/sbin $TOOLS/sbin
COPY --from=0 $TOOLS/bin $TOOLS/bin
COPY --from=0 $TOOLS/lib64 $TOOLS/lib64
COPY --from=0 $TOOLS/lib $TOOLS/lib
COPY --from=0 $TOOLS/etc $TOOLS/etc
COPY --from=0 $TOOLS/share $TOOLS/share
COPY --from=0 $TOOLS/embedded/bin $TOOLS/embedded/bin
COPY --from=0 $TOOLS/embedded/etc $TOOLS/embedded/etc
COPY --from=0 $TOOLS/embedded/lib $TOOLS/embedded/lib
COPY --from=0 $TOOLS/embedded/libexec $TOOLS/embedded/libexec
COPY --from=0 $TOOLS/embedded/sbin $TOOLS/embedded/sbin
COPY --from=0 $TOOLS/embedded/share $TOOLS/embedded/share
COPY --from=0 $TOOLS/embedded/ssl $TOOLS/embedded/ssl
