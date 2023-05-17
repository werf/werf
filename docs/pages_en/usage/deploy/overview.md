---
title: Overview
permalink: usage/deploy/overview.html
---

When setting up application delivery in Kubernetes, you must decide which format to use for managing the deployment configuration (parameterization, dependency management, configuration for different environments, etc.), as well as how to apply this configuration (the deployment mechanism).

Helm is built into werf and is used for these tasks. Configuration development and maintenance is carried out using Helm charts. As for the deployment process, Helm in werf provides some advanced features:

- tracking the status of deployed resources (with the option to change the behaviour [for each resource]({{ "/reference/deploy_annotations.html" | true_relative_url }})):
  - smart waiting for resources to become ready;
  - instant termination of a failed deployment without the need to wait for a timeout;
  - deployment progress, logs, system events and application errors.
- arbitrary deployment order for any resources, not just hooks;
- waiting for resources that are not part of the release to be created and ready;
- integrating building and deployment, and much more.

werf strives to make working with Helm easier, more convenient and flexible without breaking backward compatibility with Helm charts, Helm templates, and Helm releases.

## Basic deployment example

You only need two files and the `werf converge` command (run it in the application's Git repository) to deploy a basic application:

```yaml
# .helm/templates/hello.yaml:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello
spec:
  selector:
    matchLabels:
      app: hello
  template:
    metadata:
      labels:
        app: hello
    spec:
      containers:
      - image: nginxdemos/hello:plain-text
```

```yaml
# werf.yaml:
configVersion: 1
project: hello
```

```shell
werf converge --repo registry.example.org/repo --env production
```

The above command will deploy the `hello` Deployment to the `hello-production` Namespace.

## Advanced deployment example

Here is a more complex deployment example involving image assembly and external Helm charts:

```yaml
# werf.yaml:
configVersion: 1
project: myapp
---
image: backend
dockerfile: Dockerfile
```

```dockerfile
# Dockerfile:
FROM node
WORKDIR /app
COPY . .
RUN npm ci
CMD ["node", "server.js"]
```

```yaml
# .helm/Chart.yaml:
dependencies:
- name: postgresql
  version: "~12.1.9"
  repository: https://charts.bitnami.com/bitnami
```

```yaml
# .helm/values.yaml:
backend:
  replicas: 1
```

{% raw %}

```yaml
# .helm/templates/backend.yaml:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
spec:
  replicas: {{ $.Values.backend.replicas }}
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
    spec:
      containers:
      - image: {{ $.Values.werf.image.backend }}
```

{% endraw %}

```shell
werf converge --repo registry.example.org/repo --env production
```

The above command will build the `backend` image and then deploy the `backend` Deployment and the resources of the `postgresql` chart to the `myapp-production` Namespace.
