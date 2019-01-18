---
title: Stages
sidebar: reference
permalink: reference/build/stages.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
---

What usually needs for build application image?

* Choose a base image
* Add source code
* Install system dependencies and software
* Install application dependencies
* Configure system software
* Configure application

In what order do you need to perform these steps for the effective assembly (re-assembly) process?

We propose to divide the assembly into steps with clear functions and purposes. In dapp, such steps are called _stages_.

## What is a stage?

A ***stage*** is a logically grouped set of dappfile instructions, as well as the conditions and rules by which these instructions are assembled.

The dapp assembly process is a sequential build of _stages_. Dapp uses different _stage conveyor_ for assembling a particular type of build object. A ***stage conveyor*** is a statically defined sequence of _stages_. The set of _stages_ and their order is predetermined.

<div class="tab">
  <button class="tablinks active" onclick="openTab(event, 'dimg')">Dimg</button>
  <button class="tablinks" onclick="openTab(event, 'artifact')">Artifact</button>
</div>

<div id="dimg" class="tabcontent active">
<a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vRKB-_Re-ZhkUSB45jF9GcM-3gnE2snMjTOEIQZSyXUniNHKK-eCQl8jw3tHFF-a6JLAr2sV73lGAdw/pub?w=2000&amp;h=881" data-featherlight="image">
<img src="https://docs.google.com/drawings/d/e/2PACX-1vRKB-_Re-ZhkUSB45jF9GcM-3gnE2snMjTOEIQZSyXUniNHKK-eCQl8jw3tHFF-a6JLAr2sV73lGAdw/pub?w=821&amp;h=362" >
</a>
</div>

<div id="artifact" class="tabcontent">
<a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vRD-K_z7KEoliEVT4GpTekCkeaFMbSPWZpZkyTDms4XLeJAWEnnj4EeAxsdwnU3OtSW_vuKxDaaFLgD/pub?w=1800&amp;h=850" data-featherlight="image">
<img src="https://docs.google.com/drawings/d/e/2PACX-1vRD-K_z7KEoliEVT4GpTekCkeaFMbSPWZpZkyTDms4XLeJAWEnnj4EeAxsdwnU3OtSW_vuKxDaaFLgD/pub?w=640&amp;h=301">
</a>
</div>

<!-- 301 -->

**All works with _stages_ are done by dapp, and you only need to write dappfile correctly.**

Each _stage_ is assembled in an ***assembly container*** based on an image of the previous _stage_. The result of the assembly _stage_ and _stage conveyor_, in general, is the ***stages cache***: each _stage_ relates to one docker image.

Using a cache for re-assemblies is possible due to the build stage identifier called _signature_. The _signature_ is calculated for the _stages_ at each build. At the last step of the build when saving _stages cache_, the _signature_ is used for tagging (`dimgstage-<project name>:<signature>`). This logic allows to assembly only _stages_ whose the _stages cache_ does not exist in the docker. More information about _stages cache_ in a [separate article]({{ site.baseurl }}/reference/build/cache.html).

<div class="rsc" markdown="1">

<div class="rsc-description" markdown="1">

  The ***stage signature*** is the checksum of _stage dependencies_ and previous _stage signature_. In the absence of _stage dependencies_, the _stage_ is skipped.

  It means that the _stage conveyor_, e.g., dimg _stage conveyor_, can be reduced to several _stages_ or even to single _from stage_.

</div>

<div class="rsc-example">
<a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vR6qxP5dbQNlHXik0jCvEcKZS2gKbdNmbFa8XIem8pixSHSGvmL1n7rpuuQv64YWl48wLXfpwbLQEG_/pub?w=572&amp;h=577" data-featherlight="image">
<img src="https://docs.google.com/drawings/d/e/2PACX-1vR6qxP5dbQNlHXik0jCvEcKZS2gKbdNmbFa8XIem8pixSHSGvmL1n7rpuuQv64YWl48wLXfpwbLQEG_/pub?w=286&amp;h=288">
</a>
</div>

</div>

<div style="clear: both;"></div>

## Dapp build

Dapp build command launches assembly process for dimgs specified in the dappfile.

### Multiple builds on the same host

Multiple build commands can run at the same time on the same host. When building _stage_ dapp acquires a **lock** using _stage signature_ as ID so that only one build process is active for a stage with a particular signature at the same time.

When another build process is holding a lock for a stage, dapp waits until this process releases a lock. Then dapp proceeds to the next stage.

The reason is no need to build the same stage multiple times. Dapp build process can wait until another process finishes build and puts _stage_ into the _stages cache_.

### Syntax

```bash
dapp dimg build [options] [DIMG ...]
  --introspect-stage STAGE
  --introspect-before STAGE
  --introspect-artifact-before STAGE
  --introspect-artifact-stage STAGE
  --introspect-before-error
  --introspect-error
  --ssh-key SSH_KEY
  --name NAME
  --lock-timeout TIMEOUT
```

##### DIMG

The `DIMG` optional parameter — is a name of dimg from a dappfile. Specifying `DIMG` one or multiple times allows building only certain dimgs from dappfile. By default, dapp builds all dimgs.

##### \-\-introspect-before-error

Introspect dimg or artifact stage in the clean state, when no assembly instructions from the failed stage have been run yet.

##### \-\-introspect-error

Introspect dimg or artifact stage in the state, right after running assembly instruction, which failed.

##### \-\-ssh-key SSH_KEY

Make ssh-key available during assembly process for git operations and in assembly containers.

`SSH_KEY` is the path to the private ssh-key file.

The option can be specified multiple times to make multiple ssh-keys available in the build.

The use of this option disables system ssh-agent for the build. Only specified ssh-keys is available during assembly process if this option has been specified at least once.

##### \-\-name NAME

Use custom [dapp name](https://flant.github.io/dapp/reference/glossary.html#dapp-name). Changing default name causes full cache rebuild because dapp name affects stages cache images naming.

##### \-\-lock-timeout TIMEOUT

Specify build lock timeout for dapp to wait until another process builds some stage. 24 hours by default.

### Examples

#### Build all dimgs

Build all dimgs of dappfile:

```bash
dapp dimg build
```

#### Build specified dimgs

Given dappfile with dimgs `backend`, `frontend` and `api`, build only `backend` and `api` dimgs:

```bash
dapp dimg build backend api
```

#### Build with introspection

Run build and enable drop-in shell session in the failed assembly container in the case when an error occurred:

```bash
dapp dimg build --introspect-error
```
