---
title: Overview
permalink: usage/distribute/overview.html
---

The typical werf-based application delivery cycle looks like this: building images, publishing them, and deploying charts. A single run of the `werf converge` command is often enough to do all these things. But sometimes you need to separate the *distribution* of artifacts (images, charts) and their *deployment*, or even to deploy the artifacts without werf at all (using some third-party software).

This section discusses how to distribute images and bundles (charts with their related images) for subsequent deployment with or without werf. Refer to the Deployment section for instructions on deploying published artifacts.

## Example of image distribution

You only need two files and one `werf export` command running in the application's Git repository to distribute a single image built using the Dockerfile, which will be deployed by third-party software:

```yaml
# werf.yaml:
project: myproject
configVersion: 1
---
image: myapp
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

```shell
werf export myapp --repo example.org/myproject --tag other.example.org/myproject/myapp:latest
```

As a result of running the above command, the `other.example.org/myproject/myapp:latest` application image will be published and ready to be deployed by third-party software.

## Example of bundle distribution

In the most basic case, three files and one `werf bundle publish` command running in the application's Git repository are enough to distribute the bundle to be deployed with werf, hooking it up as a dependent bundle, or deploying it as a chart by third-party software:

```yaml
# werf.yaml:
project: mybundle
configVersion: 1
---
image: myapp
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

{% raw %}

```
# .helm/templates/myapp.yaml:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      - image: {{ $.Values.werf.image.myapp }}
```

{% endraw %}

```shell
werf bundle publish --repo example.org/bundles/mybundle
```

The above command will result in the `example.org/bundles/mybundle:latest` chart and its related built image being published.
