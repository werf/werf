---
title: Deploy into Kubernetes
sidebar: documentation
permalink: documentation/guides/deploy_into_kubernetes.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

## Task overview

How to deploy application into Kubernetes using werf.

werf uses Helm with some additions to deploy applications into Kubernetes. In this article we will create a simple web application, build all needed images, write helm templates and run it on your Kubernetes cluster.

## Requirements

 * Working Kubernetes cluster. It may be Minikube or regular Kubernetes installation. Read [the article about Minikube setup]({{ site.baseurl }}/documentation/reference/development_and_debug/setup_minikube.html) to set up local minikube instance with Docker registry.

 * Working Docker registry.

   * Accessible from a host machine to push images to the Docker registry and read info about images being deployed.

   * Accessible from Kubernetes nodes to pull images from the Docker registry.

 * Installed [werf dependencies]({{ site.baseurl }}/documentation/guides/installation.html#install-dependencies) on the host system.

 * Installed [multiwerf](https://github.com/flant/multiwerf) on the host system.

 * Installed `kubectl` on a host machine configured to access your Kubernetes cluster (<https://kubernetes.io/docs/tasks/tools/install-kubectl/>).

**NOTICE** In the following steps we will use `:minikube` as `REPO` argument of werf commands. If you are using own Kubernetes and Docker registry installation, specify your own `REPO` address instead of `:minikube`.

### Select werf version

This command should be run prior running any werf command in your shell session:

```shell
. $(multiwerf use 1.0 stable --as-file)
```

## The application

Our example application is a simple web application. To run this application we only need a web server.

Desirable Kubernetes layout:

     .----------------------.
     | backend (Deployment) |
     '----------------------'
                |
                |
      .--------------------.
      | frontend (Ingress) |
      '--------------------'

Where `backend` is a web server, `frontend` is a proxy for our application and also it is an entry point server to access Kubernetes cluster from outside.

## Application files

werf expects that all files needed to build and deploy are residing in the same directory with application source files itself (if any).

So let's create empty application directory on host machine:

```shell
mkdir myapp
cd myapp
```

## Prepare an image

We need to prepare main application image with a web server. Create the following `werf.yaml` in the root of the application directory:

```yaml
project: myapp
configVersion: 1
---

image: ~
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
            <img src="https://flant.com/images/logo_en.png" style="max-height:100%;" height="76">
          </body>
        </html>
      dest: /app/index.html
```

Our web application consists of a single static web page which created right in the config. This page is served by python HTTP server.

Build and push an image with the following command:

```shell
werf build-and-publish --stages-storage :local --tag-custom myapp --images-repo :minikube
```

The image name consists of `REPO` and `TAG`. We have specified `:minikube` as a `REPO` — this is a shortcut for `werf-registry.kube-system.svc.cluster.local:5000/myapp`. As we have specified `myapp` as a tag, for our example werf will push into the Docker registry image with the name `werf-registry.kube-system.svc.cluster.local:5000/myapp:myapp`.

## Prepare deploy configuration

werf uses helm under the hood *to apply* Kubernetes configuration. *To describe* kubernetes configuration werf also use helm configuration files (templates, values) with some extensions, such as secret files and secret values, additional Helm Go templates to generate image names and some more.

### Backend configuration

Place a file `.helm/templates/010-backend.yaml` with configuration of `backend` and then we will see what's going on in detail:

{% raw %}
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Chart.Name }}-backend
  labels:
    service: {{ .Chart.Name }}-backend
spec:
  replicas: 4
  selector:
    matchLabels:
      service: {{ .Chart.Name }}-backend
  template:
    metadata:
      labels:
        service: {{ .Chart.Name }}-backend
    spec:
      containers:
      - name: backend
        workingDir: /app
        command: [ "python3", "-m", "http.server", "8080" ]
{{ werf_container_image . | indent 8 }}
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
{{ werf_container_env . | indent 8 }}
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

Construction {% raw %}`{{ werf_container_image . | indent 8 }}`{% endraw %} is an addition of werf to helm which:

* always generates right image name (`werf-registry.kube-system.svc.cluster.local:5000/myapp:latest` in our case), and

* may generate other related fields (such as imagePullPolicy) based on some external conditions.

Go template function `werf_container_image` is the valid way to specify image **from config** in Kubernetes resource configuration. 
There may be multiple images described in config [see the reference for details]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#werf_container_image).

Construction {% raw %}`{{ werf_container_env . | indent 8 }}`{% endraw %} is another addition of werf to helm which *may* generate environment variables section for the Kubernetes resource. 
It is needed for Kubernetes to shut down and restart deployment pods only when docker image has been changed, [see the reference for details]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#werf_container_env).

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
werf deploy --stages-storage :local --images-repo :minikube --tag-custom myapp --env dev
```

With this command werf will create all Kubernetes resources using helm and watch until `myapp-backend` Deployment is ready (when all replicas Pods are up and running).

[Environment]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#environment) `--env` is a required param needed to generate helm release name and kubernetes namespace.

Helm release with name `myapp-dev` will be created. This name consists of [project name]({{ site.baseurl }}/documentation/configuration/introduction.html#meta-configuration-doc) `myapp` (which you've placed in the `werf.yaml`) and specified environment `dev`. Check docs for details about [helm release name generation]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#release-name).

Kubernetes namespace `myapp-dev` will also be used. This name also consists of [project name]({{ site.baseurl }}/documentation/configuration/introduction.html#meta-configuration-doc) `myapp` and specified environment `dev`. Check docs for details about [Kubernetes namespace generation]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html#kubernetes-namespace).

## Check your application

Now it is time to know the IP address of your Kubernetes cluster. If you use minikube get it with (in the most cases the IP address will be `192.168.99.100`):

```shell
minikube ip
```

Make sure that host name `myapp.local` is resolving to this IP address on your machine. For example append this record to the `/etc/hosts` file:

```shell
192.168.99.100 myapp.local
```

Then you can check application by url: `http://myapp.local`.

## Delete application from Kubernetes

To completely remove deployed application run this dismiss werf command:

```shell
werf dismiss --env dev --with-namespace
```

## See also

For all werf deploy features such as secrets [take a look at reference]({{ site.baseurl }}/documentation/reference/deploy_process/deploy_into_kubernetes.html).
