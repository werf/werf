---
title: Adding docker instructions
sidebar: reference
permalink: reference/build/docker_directive.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
summary: |
  <a href="https://docs.google.com/drawings/d/e/2PACX-1vTZB0BLxL7mRUFxkrOMaj310CQgb5D5H_V0gXe7QYsTu3kKkdwchg--A1EoEP2CtKbO8pp2qARfeoOK/pub?w=2031&amp;h=144" data-featherlight="image">
    <img src="https://docs.google.com/drawings/d/e/2PACX-1vTZB0BLxL7mRUFxkrOMaj310CQgb5D5H_V0gXe7QYsTu3kKkdwchg--A1EoEP2CtKbO8pp2qARfeoOK/pub?w=1016&amp;h=72">
  </a>
  
  <div class="language-yaml highlighter-rouge"><pre class="highlight"><code><span class="s">docker</span><span class="pi">:</span>
    <span class="s">VOLUME</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;volume&gt;</span>
    <span class="s">EXPOSE</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;expose&gt;</span>
    <span class="s">ENV</span><span class="pi">:</span>
      <span class="s">&lt;env_name&gt;</span><span class="pi">:</span> <span class="s">&lt;env_value&gt;</span>
    <span class="s">LABEL</span><span class="pi">:</span>
      <span class="s">&lt;label_name&gt;</span><span class="pi">:</span> <span class="s">&lt;label_value&gt;</span>
    <span class="s">ENTRYPOINT</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;entrypoint&gt;</span>
    <span class="s">CMD</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;cmd&gt;</span>
    <span class="s">ONBUILD</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;onbuild&gt;</span>
    <span class="s">WORKDIR</span><span class="pi">:</span> <span class="s">&lt;workdir&gt;</span>
    <span class="s">USER</span><span class="pi">:</span> <span class="s">&lt;user&gt;</span></code></pre>
  </div>
---

Docker can build images by [Dockerfile](https://docs.docker.com/engine/reference/builder/) instructions. These instructions can be divided into two groups: build-time instructions and other instructions that effect on a running built image and a built image in general.  

Build-time instructions don't make sense in a dapp build process, therefore, dapp supports only following instructions:

* `USER` to set the user and the group to use when running the image (read more [here](https://docs.docker.com/engine/reference/builder/#user)).
* `WORKDIR` to set the working directory (read more [here](https://docs.docker.com/engine/reference/builder/#workdir)).
* `VOLUME` to add mount point (read more [here](https://docs.docker.com/engine/reference/builder/#volume)).
* `ENV` to set the environment variable (read more [here](https://docs.docker.com/engine/reference/builder/#env)).
* `LABEL` to add metadata to an image (read more [here](https://docs.docker.com/engine/reference/builder/#label)).
* `EXPOSE` to inform Docker that the container listens on the specified network ports at runtime (read more [here](https://docs.docker.com/engine/reference/builder/#expose))
* `ENTRYPOINT` to configure a container that will run as an executable (read more [here](https://docs.docker.com/engine/reference/builder/#entrypoint)).
* `CMD` to provide default arguments for the `ENTRYPOINT` to configure a container that will run as an executable (read more [here](https://docs.docker.com/engine/reference/builder/#cmd)).
* `ONBUILD` to add to the image a trigger instruction to be executed at a later time when the image is used as the base for another build (read more [here](https://docs.docker.com/engine/reference/builder/#onbuild)).

These instructions can be defined in the dappfile `docker` section.

Here are the example of using docker instructions:

```yaml
docker:
  WORKDIR: /app
  CMD: ['python', './index.py']
  EXPOSE: '5000'
  ENV:
    TERM: xterm
    LC_ALL: en_US.UTF-8
```

Defined docker instructions are applied on the last stage called `docker_instructions`. Thus instructions don't effect on other stages and a dapp build process in general, they simply will be applied on a result image. 

If need to use special environment variables in build-time of your application image, such as `TERM` environment, you should use a [base image]({{ site.baseurl }}/reference/build/base_image.html) with these variables.

> Tip: you can also implement exporting environment variables right in [_user stage_]({{ site.baseurl }}/reference/build/assembly_instructions.html#what-is-user-stages) instructions.
