---
title: Quickstart
permalink: documentation/quickstart.html
sidebar: documentation
---

In this article we will show how to setup deployment of [example application](https://github.com/werf/quickstart-application) — which is a cool voting application — with werf. It is better to start with [short introduction]({{ site.baseurl }}/introduction.html) first if you haven't read it yet.

It is required to [install werf]({{ site.baseurl }}/installation.html) for this quickstart guide.

## Prepare your Kubernetes and Docker Registry

You should have access to the Kubernetes cluster and be able to push images to your Docker Registry. Docker Registry should also be accessible from the Kubernetes cluster to pull images.

If you have already running Kubernetes and Docker Registry:

 1. Perform standard docker login procedure into your Docker Registry from your host.
 2. Make sure your Kubernetes cluster is accessible from your host (if `kubectl` tool is already set up and working then `werf` will also work).

<br>

<div class="details">
<div id="details_link">
<a href="javascript:void(0)" class="details__summary">Or follow these steps to setup local Kubernetes and Docker Registry.</a>
</div>
<div class="details__content" markdown="1">
 1. Install [minikube](https://github.com/kubernetes/minikube#installation).
 2. Start minikube:

    {% raw %}
    ```shell
    minikube start
    ```
    {% endraw %}

 3. Enable minikube registry addon:

    {% raw %}
    ```shell
    minikube addons enable registry
    ```
    {% endraw %}

 4. Run a service with the binding to 5000 port:

    {% raw %}
    ```shell
    kubectl -n kube-system expose rc/registry --type=ClusterIP --port=5000 --target-port=5000 --name=werf-registry --selector='actual-registry=true'
    ```
    {% endraw %}

 5. Run port forwarder on host system in a separate terminal:

    {% raw %}
    ```shell
    kubectl port-forward --namespace kube-system service/werf-registry 5000
    ```
    {% endraw %}
</div>
</div>

## Deploy example application

 1. Clone example application repo

    {% raw %}
    ```shell
    git clone https://github.com/werf/quickstart-application
    cd quickstart-application
    ```
    {% endraw %}

 2. Run converge using your Docker Registry to store images (`localhost:5000/quickstart-application` repository in the case of using local Docker Registry).

    {% raw %}
    ```shell
    werf converge --repo localhost:5000/quickstart-application
    ```
    {% endraw %}

_NOTE: `werf` tool respects the same settings to connect to the Kubernetes cluster as `kubectl` tool uses: `~/.kube/config`, `KUBECONFIG` environment variable. Werf also supports `--kube-config` and `--kube-config-base64` params to setup custom kube config._

## Check result

When converge command is done succesfully it is safe to assume that our application is up and running and it is time to check it.

Our application is a simple vote system. Go to the following url to vote ([http://172.17.0.3:31000](http://172.17.0.3:31000) in our case):

{% raw %}
```
minikube service --namespace quickstart-application --url vote
```
{% endraw %}

Go to the following url to check the result of voting ([http://172.17.0.3:31001](http://172.17.0.3:31001) in our case):

{% raw %}
```
minikube service --namespace quickstart-application --url result
```
{% endraw %}

## How it works

To deploy application using werf we should define the desired state in the Git (as described in the [introduction]({{ site.baseurl }}/introduction.html)).

 1. We have following dockerfiles in our repo:

    {% raw %}
    ```
    vote/Dockerfile
    result/Dockerfile
    worker/Dockerfile
    ```
    {% endraw %}

 2. We have `werf.yaml` which references these dockerfiles:

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

 3. We have kubernetes templates for `vote`, `db`, `redis`, `result` and `worker` components of the application described in the files of `.helm/templates/` directory. Components interact with each other as shown in the diagram:

  ![architecture](https://raw.githubusercontent.com/werf/quickstart-application/master/architecture.png)

   - A front-end web app in Python or ASP.NET Core which lets you vote between two options
   - A Redis or NATS queue which collects new votes
   - A .NET Core, Java or .NET Core 2.1 worker which consumes votes and stores them in…
   - A Postgres or TiDB database backed by a Docker volume
   - A Node.js or ASP.NET Core SignalR webapp which shows the results of the voting in real time

## What's next?

Checkout [using werf with CI/CD systems article]({{ site.baseurl }}/documentation/using_with_ci_cd_systems.html) or go to the [guides]({{ site.baseurl }}/documentation/guides.html).
