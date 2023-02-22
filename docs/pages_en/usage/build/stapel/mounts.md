---
title: Using mounts to reduce image size and speed up a build
permalink: usage/build/stapel/mounts.html
author: Artem Kladov <artem.kladov@flant.com>, Alexey Igrychev <alexey.igrychev@flant.com>
directive_summary: mount
---

Quite often when you build an image, you have auxiliary files that should be excluded from the image:
- Most package managers create a system-wide cache of packages and other files.
  - [APT](https://wiki.debian.org/Apt) saves the package list in the `/var/lib/apt/lists/` directory.
  - APT also saves packages being installed in the `/var/cache/apt/` directory.
  - [YUM](http://yum.baseurl.org/) may leave downloaded packages in the `/var/cache/yum/.../packages/` directory.
- Package managers for programming languages like ​npm (nodejs), glide (go), pip (python) store files in their cache folder.
- Building C, C++ and similar applications leaves behind object and other files created by the build system.

These kinds of files:
- are not needed in the image;
- may significantly increase the image size;
- may be useful when rebuilding the image.

Mounting external folders into build containers can reduce image size and speed up the build. Docker implements the mount mechanism using [volumes](https://docs.docker.com/storage/volumes/).

The `mount` config directive is used for defining volumes. The host and build container mount folders define each volume (in the `from` / `fromPath` and `to` directives, respectively).
When specifying the host mount point, you can choose an arbitrary file or folder defined in `fromPath` or one of the service folders defined in `from`:
- `tmp_dir` is an individual temporary image directory; it is created anew for each build;
- `build_dir` is a collectively shared directory kept between builds (`~/.werf/shared_context/mounts/projects/<project name>/<mount id>/`).
Project images can use this shared directory to share and store build data (e.g., cache).

> werf mounts the service directories for reading/writing at each build; no contents of these directories will be left in the image. To keep assembly data from these directories in the image, copy them to another directory during the build.

At the `from` stage, werf adds mount point definitions to the stage image labels.
Each stage then uses these definitions to add volumes to the assembly container.
The implementation allows for inheriting mount points from the [base image]({{ "usage/build/stapel/base.html" | true_relative_url }}).

Also, during the `from` stage, werf cleans up assembly container mount points in the [base image]({{ "usage/build/stapel/base.html" | true_relative_url }}), so these directories in the image are empty.

> By default, the `fromPath` and `from: build_dir` directives are not allowed by giterminism (read more about it [here]({{ "/usage/project_configuration/giterminism.html#mount" | true_relative_url }})).