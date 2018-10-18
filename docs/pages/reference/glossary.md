---
title: Glossary
sidebar: reference
permalink: reference/glossary.html
---

## Dappfile
— is a configuration to build docker images by dapp.

## Dapp name

_Dapp name_ is either:

* a last element of git repository path from `remote.origin.url` git config parameter;
* or directory name, where dappfile reside, in the case when no git repository used or `remote.origin.url` parameter is absent.

_Dapp name_ can be explicitly specified with `--name` basic option of the most dapp commands.

## Dimg
— is the named set of rules to build one docker image. 

— is result image.

## Artifact
— is special _dimg_ that is used by another _dimgs_ and _artifacts_ to isolate the build process and build tools resources (environments, software, data).

## Stage
— is a logically grouped set of dappfile instructions, as well as the conditions and rules by which these instructions are assembled.

The dapp assembly process is a sequential build of _stages_.

## Stage assembly container
— is container for assembling stage instructions based on previous stage image (or on _base image_ for _from stage_).

## Stage signature
— is build _stage_ identifier. The _stage signature_ determines whether to build the _stage_.

The _stage signature_ is the checksum of _stage dependencies_ and previous _stage signature_.

## Stage conveyor
— is a statically defined sequence of stages.

