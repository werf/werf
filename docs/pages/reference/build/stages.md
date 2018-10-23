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

A _stage_ is a logically grouped set of dappfile instructions, as well as the conditions and rules by which these instructions are assembled. 

The dapp assembly process is a sequential build of _stages_. Dapp uses different _stage conveyor_ for assembling a particular type of build object. A _stage conveyor_ is a statically defined sequence of _stages_. The set of _stages_ and their order is predetermined. 

<div class="tab">
  <button class="tablinks active" onclick="openTab(event, 'dimg')">Dimg</button>
  <button class="tablinks" onclick="openTab(event, 'artifact')">Artifact</button>
</div>

<div id="dimg" class="tabcontent active">
<a href="https://docs.google.com/drawings/d/e/2PACX-1vRbqae63cNHREeseGvz2WDNExunn__HVzTSH9Umuvo8-WD0D9waBDdz_Z0GrRwuDIA5GSalmRgSyJI4/pub?w=2035&amp;h=859" data-featherlight="image">
<img src="https://docs.google.com/drawings/d/e/2PACX-1vRbqae63cNHREeseGvz2WDNExunn__HVzTSH9Umuvo8-WD0D9waBDdz_Z0GrRwuDIA5GSalmRgSyJI4/pub?w=1017&amp;h=429" >
</a>
</div>

<div id="artifact" class="tabcontent">
<a href="https://docs.google.com/drawings/d/e/2PACX-1vRPnqkxbv8wSziAE7QVhcP4rsb58AfIGOmOvVUbWKtZdvNhGItnL0RX8ZFZgCxxNZTtYdZ6YbVuItix/pub?w=1914&amp;h=721" data-featherlight="image">
<img src="https://docs.google.com/drawings/d/e/2PACX-1vRPnqkxbv8wSziAE7QVhcP4rsb58AfIGOmOvVUbWKtZdvNhGItnL0RX8ZFZgCxxNZTtYdZ6YbVuItix/pub?w=957&amp;h=360">
</a>
</div>

**All works with _stages_ are done by dapp, and you only need to write dappfile correctly.** 

Each _stage_ is assembled in an _assembly container_ based on an image of the previous _stage_. The result of the assembly _stage_ and _stage conveyor_ in general is the _stages cache_: each _stage_ relates to one docker image. 

Using a cache for re-assemblies is possible due to the build stage identifier called _signature_. The _signature_ is calculated for the _stages_ at each build. At the last step of the build when saving _stages cache_, the _signature_ is used for tagging (`dimgstage-<project name>:<signature>`). This logic allows to assembly only _stages_ whose the _stages cache_ does not exist in the docker. More information about _stages cache_ in a [separate article]({{ site.baseurl }}/reference/build/cache.html).

<div class="rsc" markdown="1">

<div class="rsc-description" markdown="1">
  
  The _stage signature_ is the checksum of _stage dependencies_ and previous _stage signature_. In the absence of _stage dependencies_, the _stage_ is skipped. 
  
  It means that the _stage conveyor_, e.g., dimg _stage conveyor_, can be reduced to several _stages_ or even to single _from stage_.
  
</div>

<div class="rsc-example">
<a href="https://docs.google.com/drawings/d/e/2PACX-1vSL81NRgq51uWSBUdSG4amon-e-loGKtLGJLWu35Anw-EyE9VVsBxJfP89TiUpWQRHrIXbTTijeedsF/pub?w=572&amp;h=577" data-featherlight="image">
<img src="https://docs.google.com/drawings/d/e/2PACX-1vSL81NRgq51uWSBUdSG4amon-e-loGKtLGJLWu35Anw-EyE9VVsBxJfP89TiUpWQRHrIXbTTijeedsF/pub?w=286&amp;h=288">
</a>
</div>

</div>

<div style="clear: both;"></div>
