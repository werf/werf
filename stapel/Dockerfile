ARG TARGETARCH=amd64

FROM --platform=linux/${TARGETARCH} alpine:3.19 AS alpine-tools

ARG TARGETARCH=amd64

RUN apk add --no-cache \
  bash \
  git \
  coreutils \
  findutils \
  gawk \
  sed \
  tar \
  attr \
  acl \
  sudo \
  shadow \
  gnupg \
  patchelf \
  file \
  unzip \
  util-linux \
  wget

RUN apk add --no-cache --virtual .build-deps \
  gcc \
  musl-dev \
  make \
  python3 \
  autoconf \
  automake \
  libtool

RUN cd /tmp && \
  wget https://github.com/RsyncProject/rsync/archive/v3.1.2.tar.gz && \
  tar xzf v3.1.2.tar.gz && \
  cd rsync-3.1.2 && \
  ./configure --prefix=/usr && \
  make && \
  make install && \
  cd / && \
  rm -rf /tmp/rsync-3.1.2 /tmp/v3.1.2.tar.gz && \
  apk del .build-deps

ENV TOOLS=/.werf/stapel

RUN mkdir -p \
  $TOOLS/embedded/bin \
  $TOOLS/embedded/lib \
  $TOOLS/embedded/libexec \
  $TOOLS/embedded/share \
  $TOOLS/lib \
  $TOOLS/bin \
  $TOOLS/sbin

# Copy embedded binaries that werf uses
RUN for b in bash git rsync tar install ls rm mkdir cut head stat sort chown true xargs find base64; do \
  cp "$(which $b)" $TOOLS/embedded/bin/$b; \
  done

# Copy git-core libexec (needed for git subcommands and credential helpers)
RUN cp -r /usr/libexec/git-core $TOOLS/embedded/libexec/ && \
  cp -r /usr/share/git-core $TOOLS/embedded/share/ 2>/dev/null || true

# Copy gpg and related binaries to bin/
RUN for b in gpg gpg-agent gpgconf gpgsm gpgv gpgtar; do \
  b_path="$(which $b 2>/dev/null)" && [ -n "$b_path" ] && cp "$b_path" $TOOLS/bin/$b || true; \
  done

# Copy musl linker to $TOOLS/lib — this is the ELF interpreter for all Alpine binaries
RUN case "$TARGETARCH" in \
  arm64) MUSL_LD=ld-musl-aarch64.so.1 ;; \
  amd64) MUSL_LD=ld-musl-x86_64.so.1 ;; \
  *) echo "Unsupported arch: $TARGETARCH" && exit 1 ;; \
  esac && \
  cp "/lib/$MUSL_LD" $TOOLS/lib/

# Collect shared libs for all binaries (direct deps)
RUN find $TOOLS/embedded/bin $TOOLS/embedded/libexec $TOOLS/bin $TOOLS/sbin -type f | while read f; do \
  ldd "$f" 2>/dev/null | awk '{print $3}' | grep -E '^/' | while read lib; do \
  [ -f "$lib" ] && cp -Lp "$lib" $TOOLS/embedded/lib/ 2>/dev/null || true; \
  done; \
  done

# Second pass: transitive deps of the libs we just copied
RUN find $TOOLS/embedded/lib -type f -name '*.so*' | while read f; do \
  ldd "$f" 2>/dev/null | awk '{print $3}' | grep -E '^/' | while read lib; do \
  [ -f "$lib" ] && cp -Lp "$lib" $TOOLS/embedded/lib/ 2>/dev/null || true; \
  done; \
  done

# Patch all ELF binaries: set interpreter to stapel-internal musl linker and rpath
RUN case "$TARGETARCH" in \
  arm64) MUSL_INTERP=$TOOLS/lib/ld-musl-aarch64.so.1 ;; \
  amd64) MUSL_INTERP=$TOOLS/lib/ld-musl-x86_64.so.1 ;; \
  esac && \
  find $TOOLS/embedded/bin $TOOLS/embedded/libexec $TOOLS/bin $TOOLS/sbin -type f | while read f; do \
  if file "$f" 2>/dev/null | grep -q 'ELF.*dynamically linked'; then \
  patchelf --set-interpreter "$MUSL_INTERP" \
  --set-rpath "$TOOLS/embedded/lib:$TOOLS/lib" \
  "$f" 2>/dev/null || true; \
  fi; \
  done

FROM scratch AS final
CMD ["no-such-command"]
ENV TOOLS=/.werf/stapel
ARG TARGETARCH=amd64
COPY --from=alpine-tools /.werf/stapel /.werf/stapel
