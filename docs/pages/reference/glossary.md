---
title: Glossary
sidebar: reference
permalink: reference/glossary.html
---

## Werf config
— is a configuration to build docker images by Werf described in `werf.yaml` file and supplement `.werf/**/*.tmpl` files.

## Werf name

FIXME: project name and config

_Werf name_ is either:

* a last element of git repository path from `remote.origin.url` git config parameter;
* or directory name, where config reside, in the case when no git repository used or `remote.origin.url` parameter is absent.

_Werf name_ can be explicitly specified with `--name` basic option of the most werf commands.

## Image
— is the named set of rules to build one docker image.

— is result image.

## Artifact
— is special _image_ that is used by another _images_ and _artifacts_ to isolate the build process and build tools resources (environments, software, data).

## Stage
— is a logically grouped set of config instructions, as well as the conditions and rules by which these instructions are assembled.

The werf assembly process is a sequential build of _stages_.

## User stage
— is a _stage_ with assembly instructions from config.

## Stage assembly container
— is container for assembling stage instructions based on previous stage image (or on _base image_ for _from_ stage).

## Stage signature
— is build _stage_ identifier. The _stage signature_ determines whether to build the _stage_.

The _stage signature_ is the checksum of _stage dependencies_ and previous _stage signature_.

## Stage conveyor
— is a statically defined sequence of stages.

