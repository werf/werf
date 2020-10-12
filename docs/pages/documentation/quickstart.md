---
title: Quickstart
permalink: documentation/quickstart.html
sidebar: documentation
---

In this article we will show how to setup deployment of [example application](https://github.com/werf/quickstart-application) with werf.

[Install werf]({{ site.baseurl }}/installation.html) first.

## Setup local Kubernetes and Docker Registry

_NOTE: Skip this section if you already have running Kubernetes instance and Docker Registry, which are available from your host._

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

## Deploy example application

 1. Clone example application repo

    {% raw %}
    ```shell
    git clone https://github.com/werf/quickstart-application
    cd quickstart-application
    ```
    {% endraw %}

 2. Run converge using `localhost:5000/quickstart-application` docker repository to store images.

    {% raw %}
    ```shell
    werf converge --repo localhost:5000/quickstart-application
    ```
    {% endraw %}

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

Due to the [introduction]({{ site.baseurl }}/introduction.html) we should define the desired state in the Git.

 1. We have following dockerfiles in our repo:

    {% raw %}
    ```
    vote/Dockerfile
    result/Dockerfile
    worker/Dockerfile
    ```
    {% endraw %}

 2. `werf.yaml` which uses these dockerfiles:

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

Checkout [using werf with CI/CD systems article]({{ site.baseurl }}/documentation/using_with_ci_cd_systems.html) or go to the [guides](https://ru.werf.io/applications_guide_ru).
