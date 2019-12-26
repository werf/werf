---
title: Getting started
sidebar: documentation
permalink: documentation/guides/getting_started.html
author: Artem Kladov <artem.kladov@flant.com>
---

## Task Overview

This tutorial demonstrates how you can easily start using werf with Dockerfile.
We will build the application image and publish it to the local Docker registry.
For example, we choose simple project — [Linux Tweet App](https://github.com/dockersamples/linux_tweet_app).

## Requirements

* Minimal knowledge of [Docker](https://www.docker.com/) and [Dockerfile instructions](https://docs.docker.com/engine/reference/builder/).
* Installed [werf dependencies]({{ site.baseurl }}/documentation/guides/installation.html#install-dependencies) on the host system.
* Installed [multiwerf](https://github.com/flant/multiwerf) on the host system.

### Select werf version

This command should be run prior running any werf command in your shell session:

```shell
. $(multiwerf use 1.0 stable --as-file)
```

## Step 1: Add a werf configuration

Add a special file called `werf.yaml` to the source code and define application image based on project [Dockerfile](https://github.com/dockersamples/linux_tweet_app/blob/master/Dockerfile).

1.  Clone the [Linux Tweet App](https://github.com/dockersamples/linux_tweet_app) repository to get the source code:

    ```shell
    git clone https://github.com/dockersamples/linux_tweet_app.git
    cd linux_tweet_app
    ```

1.  In the project root directory create a `werf.yaml` with the following content:

    ```yaml
    project: g-started
    configVersion: 1
    ---
    image: ~
    dockerfile: Dockerfile
    ```

## Step 2: Build and Check the Application

1.  Build an image (execute in the project root directory):

    ```shell
    werf build --stages-storage :local
    ```

1.  Run a container from the image:

    ```shell
    werf run --stages-storage :local --docker-options="-d -p 80:80"
    ```

1.  Check that the application runs and responds by opening `http://localhost:80` in your browser or executing:

    ```shell
    curl localhost:80
    ```

## Step 3: Publish built image to Docker registry

1.  Run local Docker registry:

    ```shell
    docker run -d -p 5000:5000 --restart=always --name registry registry:2
    ```

2.  Publish image using custom tagging strategy with docker tag `v0.1.0`:

    ```shell
    werf publish --stages-storage :local --images-repo localhost:5000/g-started --tag-custom v0.1.0
    ```

## What's next?

Firstly, you can plunge into the relevant documentation:
* [werf configuration file]({{ site.base_url}}/documentation/configuration/introduction.html).
* [Dockerfile Image: complete directive list]({{ site.base_url}}/documentation/configuration/dockerfile_image.html).
* [Build procedure]({{ site.base_url}}/documentation/reference/build_process.html).
* [Publish procedure]({{ site.base_url}}/documentation/reference/publish_process.html).

Or go further:
* [Deploy an application to a Kubernetes cluster]({{ site.base_url}}/documentation/guides/deploy_into_kubernetes.html).
* [Advanced build with Stapel image]({{ site.base_url}}/documentation/guides/advanced_build/first_application.html).
