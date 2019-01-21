---
title: Deploy into Kubernetes
sidebar: how_to
permalink: how_to/deploy_into_kubernetes.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

## Task overview

How to deploy application into Kubernetes using Werf.

Werf uses Helm with some additions to deploy applications into Kubernetes. In this article we will create a simple web application, build all needed images, write helm templates and run it on your kubernetes cluster.

## Requirements

 * Working Kubernetes cluster. It may be Minikube or regular Kubernetes installation. Read [the article about Minikube setup]({{ site.baseurl }}/reference/deploy/minikube.html) to set up local minikube instance with docker-registry.

 * Working docker-registry.

   * Accessible from a host machine to push images to the registry and read info about images being deployed.

   * Accessible from kubernetes nodes to pull images from the registry.

 * Installed `werf` on a host machine (<{{ site.baseurl }}/how_to/installation.html>).

 * Installed `kubectl` on a host machine configured to access your kubernetes cluster (<https://kubernetes.io/docs/tasks/tools/install-kubectl/>).

 * Installed `helm` on a host machine (<https://docs.helm.sh/using_helm/#installing-helm>).

 * Minimal knowledge of `helm` and its concepts required (such as chart, templates, release etc.). Helm documentation is available here: <https://docs.helm.sh/>.

**NOTICE** In the following steps we will use `:minikube` as `REPO` argument of werf commands. If you are using own kubernetes and docker-registry installation, specify your own `REPO` address instead of `:minikube`.

## The application

Our example application is a simple web application. To run this application we only need a web server.

Desireable kubernetes layout:

     .----------------------.
     | backend (Deployment) |
     '----------------------'
                |
                |
      .--------------------.
      | frontend (Ingress) |
      '--------------------'

Where `backend` is a web server, `frontend` is a proxy for our application and also it is an entry point server to access kubernetes cluster from outside.

## Application files

Werf expects that all files needed to build and deploy are residing in the same directory with application source files itself (if any).

So let's create empty application directory on host machine:

```shell
mkdir myapp
cd myapp
```

## Prepare an image

We need to prepare main application image with a web server. Create the following `werf.yaml` in the root of the application directory:

```yaml
project: myapp
---

dimg: ~
from: python:alpine
ansible:
  install:
  - file:
      path: /app
      state: directory
      mode: 0755
  - name: Prepare main page
    copy:
      content:
        <!DOCTYPE html>
        <html>
          <body>
            <h2>Congratulations!</h2>
            <img src="https://github.com/flant/werf/raw/master/logo.png" style="max-height:100%;" height="25">
          </body>
        </html>
      dest: /app/index.html
```

Our web application consists of a single static web page which created right in the config. This page is served by python HTTP server.

Build and push an image with the following command:

```shell
werf dimg bp :minikube
```

The image name consists of `REPO` and `TAG`. Werf will use `latest` tag for the image by default. We have specified `:minikube` as a `REPO` — this is a shortcut for `localhost:5000/myapp`. So for our example werf will push into the docker-registry image with the name `localhost:5000/myapp:latest`.

## Prepare deploy configuration

Werf uses helm under the hood *to apply* kubernetes configuration. *To describe* kubernetes configuration werf also use helm configuration files (templates, values) with some extensions, such as secret files and secret values, additional helm golang templates to generate image names and some more.

First, we need to create a helm chart for our application. To do that place a file `.helm/Chart.yaml` with content:

```yaml
name: myapp
version: 0.1.0
```

### Backend configuration

Place a file `.helm/templates/010-backend.yaml` with configuration of `backend` and then we will see what's going on in detail:

{% raw %}
```yaml
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: {{ .Chart.Name }}-backend
spec:
  replicas: 4
  template:
    metadata:
      labels:
        service: {{ .Chart.Name }}-backend
    spec:
      containers:
      - name: backend
        workingDir: /app
        command: [ "python3", "-m", "http.server", "8080" ]
{{ include "werf_container_image" . | indent 8 }}
        livenessProbe:
          httpGet:
            path: /
            port: 8080
            scheme: HTTP
        readinessProbe:
          httpGet:
            path: /
            port: 8080
            scheme: HTTP
        ports:
        - containerPort: 8080
          name: http
          protocol: TCP
        env:
{{ include "werf_container_env" . | indent 8 }}
---
apiVersion: v1
kind: Service
metadata:
  name: {{ .Chart.Name }}-backend
spec:
  clusterIP: None
  selector:
    service: {{ .Chart.Name }}-backend
  ports:
  - name: http
    port: 8080
    protocol: TCP
```
{% endraw %}

In this configuration Deployment with name `myapp-backend` (after template {% raw %}`{{ .Chart.Name }}-backend`{% endraw %} expansion) with several replicas specified.

Construction {% raw %}`{{ include "werf_container_image" . | indent 8 }}`{% endraw %} is an addition of werf to helm which:

* always generates right image name (`localhost:5000/myapp:latest` in our case), and

* may generate other related fields (such as imagePullPolicy) based on some external conditions.

Go template `werf_container_image` is the valid way to specify image **from config** in kubernetes resource configuration. There may be multiple images described in config [see the reference for details]({{ site.baseurl }}/reference/deploy/templates.html#%D1%88%D0%B0%D0%B1%D0%BB%D0%BE%D0%BD-werf_container_image).

Construction {% raw %}`{{ include "werf_container_env" . | indent 8 }}`{% endraw %} is another addition of werf to helm which *may* generate environment variables section for the kubernetes resource. It is needed for kubernetes to shut down and restart deployment pods only when docker image has been changed, [see the reference for details]({{ site.baseurl }}/reference/deploy/templates.html#%D1%88%D0%B0%D0%B1%D0%BB%D0%BE%D0%BD-werf_container_env).

Finally, in this configuration Service `myapp-backend` specified to access Pods of Deployment `myapp-backend`.

### Frontend configuration

To describe `frontend` configuration place file `.helm/templates/090-frontend.yaml` with the following content:

```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: myapp-frontend
  annotations:
    kubernetes.io/ingress.class: "nginx"
    nginx.ingress.kubernetes.io/ssl-redirect: "false"
spec:
  rules:
  - host: myapp.local
    http:
      paths:
      - path: /
        backend:
          serviceName: myapp-backend
          servicePort: 8080
```

This Ingress configuration set up nginx proxy server for host `myapp.local` to our backend web server `myapp-backend`.

## Run deploy

If you use minikube then enable ingress addon before running deploy:

```shell
minikube addons enable ingress
```

Run deploy with werf:

```shell
werf kube deploy :minikube --namespace mynamespace
```

With this command werf will create all kubernetes resources using helm and watch until `myapp-backend` Deployment is ready (when all replicas Pods are up and running). Helm release with name `myapp-mynamespace` will be created. This name consists of chart name `myapp` and namespace `mynamespace`.

## Check your application

Now it is time to know the IP address of your kubernetes cluster. If you use minikube get it with (in the most cases the IP address will be `192.168.99.100`):

```shell
minikube ip
```

Make sure that host name `myapp.local` is resolving to this IP address on your machine. For example append this record to the `/etc/hosts` file:

```shell
192.168.99.100 myapp.local
```

Then you can check application by url: `http://myapp.local`.

## Delete application from kubernetes

To completely remove deployed application run this dismiss werf command:

```shell
werf kube dismiss :minikube --namespace mynamespace --with-namespace
```

## See also

For all werf deploy features such as secrets [take a look at reference]({{ site.baseurl }}/reference/deploy/deploy_to_kubernetes.html).
