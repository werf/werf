---
title: Quickstart
permalink: quickstart.html
description: Deploy your first application with werf
---

In this article we will show you how to set up the deployment of an [example application](https://github.com/werf/quickstart-application) (a cool voting app in our case) using werf. It is better to start with a [short introduction](/introduction.html) first if you haven't read it yet.

## Prepare your host

 1. Install [dependencies](/installation.html#install-dependencies) (Docker and Git).
 2. Install [multiwerf and werf](/installation.html#install-werf).

Make sure you have `werf` command available in your shell before proceeding to the next step:
   
```
werf version
```

## Prepare your Kubernetes and container registry

You should have access to the Kubernetes cluster and be able to push images to your container registry. The container registry should also be accessible from the Kubernetes cluster to pull images.

If your Kubernetes and container registry are running already:

 1. Perform the standard docker login procedure into your container registry from your host.
 2. Make sure your Kubernetes cluster is accessible from your host (if the `kubectl` tool is already set up and running, then `werf` will also work just fine).

<br>

Or use one of the following instructions to set up the local Kubernetes cluster and container registry in your OS:

<div class="details">
<a href="javascript:void(0)" class="details__summary">Windows</a>
<div class="details__content" markdown="1">
 1. Install [minikube](https://github.com/kubernetes/minikube#installation).
 2. Start minikube:

    {% raw %}
    ```shell
    minikube start --driver=docker
    ```
    {% endraw %}
    
    **IMPORTANT** If you have already started minikube make sure it uses `docker` driver, restart minikube otherwise (by using `minikube delete` and start command shown above).

 3. Enable the minikube registry addon:

    {% raw %}
    ```shell
    minikube addons enable registry
    ```
    {% endraw %}
    
    Output should contain a line like:
 
    {% raw %}
    ```
    ...
    * Registry addon on with docker uses 32769 please use that instead of default 5000
    ...
    ```
    {% endraw %}

    Remember the port `32769`.
    
 4. **In the separate terminal** run the following port forwarder replacing `32769` port with the port from the previous step:

    {% raw %}
    ```shell
    docker run -ti --rm --network=host alpine ash -c "apk add socat && socat TCP-LISTEN:5000,reuseaddr,fork TCP:host.docker.internal:32769"
    ```
    {% endraw %}

    **NOTE** This command should stay in foreground in this terminal, do not interrupt it.

 5. Run a service with a binding to 5000 port:

    {% raw %}
    ```shell
    kubectl -n kube-system expose rc/registry --type=ClusterIP --port=5000 --target-port=5000 --name=werf-registry --selector=actual-registry=true
    ```
    {% endraw %}

 6. **In the separate terminal** run the following port forwarder:

    {% raw %}
    ```shell
    kubectl port-forward --namespace kube-system service/werf-registry 5000
    ```
    {% endraw %}

    **NOTE** This command should stay in foreground in this terminal, do not interrupt it.
</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">MacOS</a>
<div class="details__content" markdown="1">
 1. Install [Docker Desktop](https://docs.docker.com/docker-for-mac/install/) or update to the latest version.
 2. Enable Kubernetes cluster in the Docker Desktop settings. If something goes wrong and Kubernetes is not starting — try to restart Docker Desktop and reset to the factory settings.
 3. Run container registry in the terminal:

    {% raw %}
    ```shell
    docker run -d -p 5000:5000 --restart=always --name registry registry:2
    ```
    {% endraw %}
</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">MacOS — alternative way</a>
<div class="details__content" markdown="1">
 1. Install [minikube](https://github.com/kubernetes/minikube#installation).
 2. Start minikube:

    {% raw %}
    ```shell
    minikube start --driver=docker
    ```
    {% endraw %}
    
    **IMPORTANT** If you have already started minikube make sure it uses `docker` driver, restart minikube otherwise (by using `minikube delete` and then start command shown above).

 3. Enable the minikube registry addon:

    {% raw %}
    ```shell
    minikube addons enable registry
    ```
    {% endraw %}
    
    Output should contain a line like:
 
    {% raw %}
    ```
    ...
    * Registry addon on with docker uses 32769 please use that instead of default 5000
    ...
    ```
    {% endraw %}

    Remember the port `32769`.
    
 4. **In the separate terminal** run the following port forwarder replacing `32769` port with the port from the step 3:

    {% raw %}
    ```shell
    docker run -ti --rm --network=host alpine ash -c "apk add socat && socat TCP-LISTEN:5000,reuseaddr,fork TCP:host.docker.internal:32769"
    ```
    {% endraw %}

    **NOTE** This command should stay in foreground in this terminal, do not interrupt it.

 5. **In the separate another terminal** run the following port forwarder replacing `32769` port with the port from the step 3:

    {% raw %}
    ```shell
    brew install socat
    socat TCP-LISTEN:5000,reuseaddr,fork TCP:localhost:32769
    ```
    {% endraw %}

    **NOTE** This command should stay in foreground in this terminal, do not interrupt it.
</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">Linux</a>
<div class="details__content" markdown="1">
 1. Install [minikube](https://github.com/kubernetes/minikube#installation).
 2. Start minikube:

    {% raw %}
    ```shell
    minikube start --driver=docker
    ```
    {% endraw %}

    **IMPORTANT** If you have already started minikube make sure it uses `docker` driver, restart minikube otherwise (by using `minikube delete` and then start command shown above).

 3. Enable the minikube registry addon:

    {% raw %}
    ```shell
    minikube addons enable registry
    ```
    {% endraw %}
 
 4. **In the separate another terminal** run the following port forwarder:

    {% raw %}
    ```shell
    sudo apt-get install -y socat
    socat -d -d TCP-LISTEN:5000,reuseaddr,fork TCP:$(minikube ip):5000
    ```
    {% endraw %}

    **NOTE** This command should stay in foreground in this terminal, do not interrupt it.
</div>
</div>

## Deploy an example application

 1. Clone the example application's repository:

    {% raw %}
    ```shell
    git clone https://github.com/werf/quickstart-application
    cd quickstart-application
    ```
    {% endraw %}

 2. Run the converge command using your container registry for storing images (`localhost:5000/quickstart-application` repository in the case of a local container registry).

    {% raw %}
    ```shell
    werf converge --repo localhost:5000/quickstart-application
    ```
    {% endraw %}
    
_NOTE: `werf` uses the same settings to connect to the Kubernetes cluster as the `kubectl` tool does: the `~/.kube/config` file and the `KUBECONFIG` environment variable. werf also supports `--kube-config` and `--kube-config-base64` parameters for specifying custom kubeconfig files._

## Check the result

When the converge command is successfully completed, it is safe to assume that our application is up and running.

Our application is a basic voting system. Let’s check it!

<div class="details">
<a href="javascript:void(0)" class="details__summary">Windows</a>
<div class="details__content" markdown="1">
 1. Go to the following URL to vote:

    {% raw %}
    ```
    minikube service --namespace quickstart-application --url vote
    ```
    {% endraw %}

 2. Go to the following URL to check the result of voting:

    {% raw %}
    ```
    minikube service --namespace quickstart-application --url result
    ```
    {% endraw %}
</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">MacOS</a>
<div class="details__content" markdown="1">
 1. Go to the following URL to vote: [http://localhost:31160](http://localhost:31160).
 2. Go to the following URL to see the result of voting: [http://localhost:31161](http://localhost:31161).
</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">MacOS — alternative way</a>
<div class="details__content" markdown="1">
 1. Go to the following URL to vote:

    {% raw %}
    ```
    minikube service --namespace quickstart-application --url vote
    ```
    {% endraw %}

 2. Go to the following URL to check the result of voting:

    {% raw %}
    ```
    minikube service --namespace quickstart-application --url result
    ```
    {% endraw %}
</div>
</div>

<div class="details">
<a href="javascript:void(0)" class="details__summary">Linux</a>
<div class="details__content" markdown="1">
 1. Go to the following URL to vote: 

    {% raw %}
    ```
    minikube service --namespace quickstart-application --url vote
    ```
    {% endraw %}

 2. Go to the following URL to check the result of voting:

    {% raw %}
    ```
    minikube service --namespace quickstart-application --url result
    ```
    {% endraw %}
</div>
</div>

## How it works

To deploy an application using werf, we should define the desired state in the Git (as set out in the [introduction](/introduction.html)).

 1. We have the following Dockerfiles in our repository:

    {% raw %}
    ```
    vote/Dockerfile
    result/Dockerfile
    worker/Dockerfile
    ```
    {% endraw %}

 2. The `werf.yaml` file references those Dockerfiles:

    {% raw %}
    ```
    configVersion: 1
    project: quickstart-application
    ---
    image: vote
    dockerfile: vote/Dockerfile
    context: vote
    ---
    image: result
    dockerfile: result/Dockerfile
    context: result
    ---
    image: worker
    dockerfile: worker/Dockerfile
    context: worker
    ```
    {% endraw %}

 3. Kubernetes templates for `vote`, `db`, `redis`, `result` and `worker` components of the application are described in the files of a `.helm/templates/` directory. Components interact with each other as shown in the diagram:

  ![Component interaction diagram]({{ "images/quickstart-architecture.svg" | true_relative_url }})

   - A front-end web app in Python or ASP.NET Core lets you vote for one of the two options;
   - A Redis or NATS queue collects new votes;
   - A .NET Core, Java or .NET Core 2.1 worker consumes votes and stores them in…
   - A Postgres or TiDB database backed by a Docker volume;
   - A Node.js or ASP.NET Core SignalR web-app shows the results of the voting in real-time.

## What's next?

Check the ["Using werf with CI/CD systems" article](using_with_ci_cd_systems.html) or refer to the [guides](/guides.html).
