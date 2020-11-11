---
title: Quickstart
permalink: documentation/quickstart.html
description: Deploy your first application with werf
sidebar: documentation
---

In this article we will show you how to set up the deployment of an [example application](https://github.com/werf/quickstart-application) (a cool voting app in our case) using werf. It is better to start with a [short introduction]({{ "introduction.html" | relative_url }}) first if you haven't read it yet.

You have to [install werf]({{ "installation.html" | relative_url }}) to use this quick-start guide.

## Prepare your Kubernetes and Docker Registry

You should have access to the Kubernetes cluster and be able to push images to your Docker Registry. Docker Registry should also be accessible from the Kubernetes cluster to pull images.

If your Kubernetes and Docker Registry are running already:

 1. Perform the standard docker login procedure into your Docker Registry from your host.
 2. Make sure your Kubernetes cluster is accessible from your host (if the `kubectl` tool is already set up and running, then `werf` will also work just fine).

<br>

<div class="details">
<a href="javascript:void(0)" class="details__summary">Or follow these steps to set up the local Kubernetes cluster and Docker Registry.</a>
<div class="details__content" markdown="1">
 1. Install [minikube](https://github.com/kubernetes/minikube#installation).
 2. Start minikube:

    {% raw %}
    ```shell
    minikube start
    ```
    {% endraw %}

 3. Enable the minikube registry addon:

    {% raw %}
    ```shell
    minikube addons enable registry
    ```
    {% endraw %}

 4. Run a service with a binding to 5000 port:

    {% raw %}
    ```shell
    kubectl -n kube-system expose rc/registry --type=ClusterIP --port=5000 --target-port=5000 --name=werf-registry --selector='actual-registry=true'
    ```
    {% endraw %}

 5. Run a port forwarder on the host system in a separate terminal:

    {% raw %}
    ```shell
    kubectl port-forward --namespace kube-system service/werf-registry 5000
    ```
    {% endraw %}
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

 2. Run the converge command using your Docker Registry for storing images (`localhost:5000/quickstart-application` repository in the case of a local Docker Registry).

    {% raw %}
    ```shell
    werf converge --repo localhost:5000/quickstart-application
    ```
    {% endraw %}
    
_NOTE: `werf` uses the same settings to connect to the Kubernetes cluster as the `kubectl` tool does: the `~/.kube/config` file and the `KUBECONFIG` environment variable. werf also supports `--kube-config` and `--kube-config-base64` parameters for specifying custom kubeconfig files._

## Check the result

When the converge command is successfully completed, it is safe to assume that our application is up and running. Let’s check it!

Our application is a basic voting system. Go to the following URL to vote ([http://172.17.0.3:31000](http://172.17.0.3:31000) in our case):

{% raw %}
```
minikube service --namespace quickstart-application --url vote
```
{% endraw %}

Go to the following URL to check the result of voting ([http://172.17.0.3:31001](http://172.17.0.3:31001) in our case):

{% raw %}
```
minikube service --namespace quickstart-application --url result
```
{% endraw %}

## How it works

To deploy an application using werf, we should define the desired state in the Git (as set out in the [introduction]({{ "introduction.html" | relative_url }})).

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

  ![architecture](/images/quickstart-architecture.svg)

   - A front-end web app in Python or ASP.NET Core lets you vote for one of the two options;
   - A Redis or NATS queue collects new votes;
   - A .NET Core, Java or .NET Core 2.1 worker consumes votes and stores them in…
   - A Postgres or TiDB database backed by a Docker volume;
   - A Node.js or ASP.NET Core SignalR web-app shows the results of the voting in real-time.

## What's next?

Check the ["Using werf with CI/CD systems" article]({{ "documentation/using_with_ci_cd_systems.html" | relative_url }}) or refer to the [guides]({{ "documentation/guides.html" | relative_url }}).
