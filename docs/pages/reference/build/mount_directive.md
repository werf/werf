---
title: Reducing image size and speeding up a build by mounts
sidebar: reference
permalink: reference/build/mount_directive.html
author: Artem Kladov <artem.kladov@flant.com>, Alexey Igrychev <alexey.igrychev@flant.com>
---

Quite often when you build an image, you have auxiliary files that should be excluded from the image. E.g.:
- Most package managers create a system-wide cache of packages and other files.
  - [APT](https://wiki.debian.org/Apt) saves the package list in the `/var/lib/apt/lists/` directory.
  - APT also saves packages in the `/var/cache/apt/` directory when installs them.
  - [YUM](http://yum.baseurl.org/) can leave downloaded packages in the `/var/cache/yum/.../packages/` directory.
- Package managers for programming languages like ​npm (nodejs), glide (go), pip (python) store files in their cache folder.
- Building C, C++ and similar applications leaves behind itself the object files and other files of the used build systems.

Thus, those files:
- are useless in the image;
- might enormously increase image size;
- might be useful in the next image builds.

Reducing image size and speeding up a build can be performed by mounting external folders into assembly containers. Docker implements a mounting mechanism using [volumes](https://docs.docker.com/storage/volumes/).

Mount directive in a dappfile allows defining volumes. Host and assembly container mount folders determine each volume (accordingly in `from`/`fromPath` and `to` directives). When specifying the host mount point, you can choose an arbitrary folder or one of the service folders:
- `tmp_dir` is an individual temporary dimg directory that is created only for one build;
- `build_dir` is a directory that is saved between builds. All dimgs in dappfile can use the common directory to store and to share assembly data (e.g., cache). The folder `~/.dapp/builds/<project name>/` store directories of this type.

Dapp binds host mount folders for reading/writing on each stage build. If you need to keep assembly data from these directories in a image, you should copy them to another directory during build.

On `from` stage dapp adds mount points definitions in stage image labels and then each stage uses theirs to add volumes in an assembly container. This implementation allows inheriting mount points from [base image]({{ site.baseurl }}/reference/build/base_image.html). 

Also on `from` stage dapp cleans assembly container mount points in a [base image]({{ site.baseurl }}/reference/build/base_image.html). Therefore these folders are empty in a image. 

## Syntax
```yaml
mount:
- from: tmp_dir
  to: <absolute_path>
- from: build_dir
  to: <absolute_path>
- fromPath: <absolute_path>
  to: <absolute_path>
```
