---
title: Installation
sidebar: how_to
permalink: how_to/installation.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

Dapp requires a Linux operating system.
Support for macOS is coming soon (see issue [#661](https://github.com/flant/dapp/issues/661)).

## Install Dependencies

1.  Ruby version 2.1 or later: 
    [Ruby installation](https://www.ruby-lang.org/en/documentation/installation/).

1.  Docker version 1.10.0 or later:
    [Docker installation](https://docs.docker.com/engine/installation/).    

1.  сmake (required to install `rugged` gem):

    on Ubuntu or Debian:

    ```bash
    apt-get install cmake
    ```

    on CentOS or RHEL:
    
    ```bash
    yum install cmake
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

## Install dapp

```bash
gem install dapp
```

Now you have dapp installed. Check it with `dapp --version`.

Time to [make your first dapp application]({{ site.baseurl }}/how_to/build_run_and_push.html)!
