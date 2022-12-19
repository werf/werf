---
title: Reducing image size and speeding up a build by mounts
permalink: usage/build_draft/stapel/mounts.html
author: Artem Kladov <artem.kladov@flant.com>, Alexey Igrychev <alexey.igrychev@flant.com>
directive_summary: mount
---

Quite often when you build an image, you have auxiliary files that should be excluded from the image:
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

`mount` config directive is used for defining volumes. Host and assembly container mount folders determine each volume (accordingly in `from`/`fromPath` and `to` directives).
When specifying the host mount point, you can choose an arbitrary file or folder, defined in `fromPath`, or one of the service folders, defined in `from`:
- `tmp_dir` is an individual temporary image directory, created new for each build;
- `build_dir` is a collectively shared directory, stored between builds (`~/.werf/shared_context/mounts/projects/<project name>/<mount id>/`).
Project images can use this common directory to share and store assembly data (e.g., cache).

> werf binds host mount folders for reading/writing on each stage build.
If you need to keep assembly data from these directories in an image, you should copy them to another directory during build

On `from` stage werf adds mount points definitions to stage image labels.
Then each stage uses these definitions for adding volumes to an assembly container.
The implementation allows inheriting mount points from [base image]({{ "usage/build_draft/stapel/base.html" | true_relative_url }}).

Also, on `from` stage werf cleans assembly container mount points in a [base image]({{ "usage/build_draft/stapel/base.html" | true_relative_url }}).
Therefore, these folders are empty in an image.

> By default, the use of the `fromPath` directive and `from: build_dir` are not allowed by giterminism (read more about it [here]({{ "/usage/project_configuration/giterminism.html#mount" | true_relative_url }}))
