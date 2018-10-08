---
title: Image naming
sidebar: reference
permalink: reference/registry/image_naming.html
---

TODO

## Docker image naming

For all commands related to docker registry dapp use single parameter named `REPO`, which is a [docker repository](https://docs.docker.com/glossary/?term=repository). However `REPO` is a _base part_ of docker repository for current dapp project, which means:

* If dapp project contains several dimgs, then dapp will add dimg name to construct final docker repository name `REPO/DIMG_NAME` for each dimg.
* If dapp project contains single dimg, then `REPO` is a final single docker repository for this dimg.
