FROM ubuntu:20.04 AS builder
ARG INVALIDATE_CACHES=overridethis
ENV DEBIAN_FRONTEND=noninteractive

RUN apt update && apt upgrade -y && \
    apt install -y ninja-build gettext libtool libtool-bin autoconf automake cmake g++ pkg-config unzip git binutils

ADD neovim-v0.1.0 /neovim
ADD neovim-v0.5.0 /neovim
COPY neovim-v0.1.0 /neovim
COPY neovim-v0.5.0 /neovim
RUN rm -rf /neovim && \
    git clone --branch v0.1.0 https://github.com/neovim/neovim /neovim && \
    rm -rf /neovim/.git
RUN rm -rf /neovim && \
    git clone --branch v0.5.0 https://github.com/neovim/neovim /neovim && \
    rm -rf /neovim/.git

ADD https://github.com/neovim/neovim/archive/refs/tags/v0.1.0.zip /neovim.zip
ADD https://github.com/neovim/neovim/archive/refs/tags/v0.5.0.zip /neovim.zip

WORKDIR /neovim
RUN make CMAKE_BUILD_TYPE=Release && \
    make CMAKE_INSTALL_PREFIX=/usr/local install


FROM ubuntu:20.04
ARG INVALIDATE_CACHES=overridethis
COPY --from=builder /neovim /neovim
ADD neovim-v0.1.0 /neovim
