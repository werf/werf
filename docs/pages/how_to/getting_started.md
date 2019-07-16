---
title: Getting started
sidebar: how_to
permalink: how_to/getting_started.html
author: Artem Kladov <artem.kladov@flant.com>
---

## Task Overview

In this tutorial, we will build an image of a simple application — [Linux Tweet App](https://github.com/dockersamples/linux_tweet_app). It has a Dockerfile, and we will build the application image, then push it to the Docker registry with Werf.

## Requirements

* Minimal knowledge of [Docker](https://www.docker.com/) and [Dockerfile instructions](https://docs.docker.com/engine/reference/builder/).
* Installed Docker and [Multiwerf](https://github.com/flant/multiwerf) on the host system.

### Select Werf version

This command should be run prior running any Werf command in your shell session:

```shell
source <(multiwerf use 1.0 beta)
```

## Step 1: Add a config

To implement these steps and requirements with werf we will add a special file called `werf.yaml` to the application's source code.

1. Clone the [Linux Tweet App](https://github.com/dockersamples/linux_tweet_app) repository to get the source code:

    ```shell
    git clone https://github.com/dockersamples/linux_tweet_app.git
    cd linux_tweet_app
    ```

1.  In the project root directory create a `werf.yaml` with the following contents:

    ```yaml
    project: g-started
    configVersion: 1
    ---
    image: ~
    dockerfile: Dockerfile
    ```

## Step 2: Build and Run the Application

Let's build and run our first application.

1.  Build an image (execute in the project root directory):

    ```shell
    werf build --stages-storage :local
    ```

1.  Run a container from the image:

    ```shell
    werf --stages-storage :local run --docker-options="-d -p 80:80"
    ```

1.  Check that the application runs and responds by opening http://localhost:8000 in your browser or executing:

    ```shell
    curl localhost:80
    ```

## Step 3: Push image into docker registry

Werf can be used to push a built image into docker-registry.

1. Run local docker-registry:

    ```shell
    docker run -d -p 5000:5000 --restart=always --name registry registry:2
    ```

2. Publish image with werf using custom tagging strategy with docker tag `v0.1.0`:

    ```shell
    werf publish --stages-storage :local --images-repo localhost:5000/g-started --tag-custom v0.1.0
    ```

## What's next?

This example is quite simple and only shows how you can easily use Werf to build images in your existing project. Please see at the [more complex example]({{ site.base_url}}/how_to/building_application.html) of building application image by using Werf stage conveyor. If you want to know how easy it is to deploy an application to a Kubernetes cluster, check out [this]({{ site.base_url}}/how_to/deploy_into_kubernetes.html) example.
