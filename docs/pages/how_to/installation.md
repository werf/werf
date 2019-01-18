---
title: Installation
sidebar: how_to
permalink: how_to/installation.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

Werf requires a Linux operating system.
Support for macOS is coming soon (see issue [#661](https://github.com/flant/werf/issues/661)).

## Install Dependencies

1.  Ruby version 2.1 or later:
    [Ruby installation](https://www.ruby-lang.org/en/documentation/installation/).

1.  Docker version 1.10.0 or later:
    [Docker installation](https://docs.docker.com/engine/installation/).

1.  cmake and pkg-config (required to install `rugged` gem):

    on Ubuntu or Debian:

    ```bash
    apt-get install cmake pkg-config
    ```

    on CentOS or RHEL:

    ```bash
    yum install cmake pkgconfig
    ```

1.  libssh2 header files to work with git via SSH.

    on Ubuntu or Debian:

    ```bash
    apt-get install libssh2-1-dev
    ```

    on CentOS or RHEL:

    ```bash
    yum install libssh2-devel
    ```

1.  libssl header files to work with git via HTTPS.

    on Ubuntu or Debian:

    ```bash
    apt-get install libssl-dev
    ```

    on CentOS or RHEL:

    ```bash
    yum install openssl-devel
    ```

## Install werf

```bash
gem install werf
```

Now you have werf installed. Check it with `werf --version`.

Time to [make your first werf application]({{ site.baseurl }}/how_to/getting_started.html)!
