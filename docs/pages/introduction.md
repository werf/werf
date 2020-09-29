---
title: Introduction
permalink: introduction.html
sidebar: documentation
---

Werf is a CLI tool to implement full cycle of deployment of your application using Git as a single source of truth:
 - Building of docker images.
 - Deploying application into Kubernetes.
 - Making sure that application is up and running well after deploy.
 - Rebuilding docker images when application code changes.
 - Redeploying application into Kubernetes when necessary.
 - Cleaning up old unused images.

TODO: visualize the process above

# 1. Define desired state in the Git

Werf configuration should be described in the application Git repository right where application code resides.

 1. Put Dockerfiles into `.dockerfiles` directory in the root of application repo. For example we have two Dockerfiles to build 2 images of application: backend and frontend:

```
$ ls .dockerfiles
~/myapp$ ls .dockerfiles/
Dockerfile-backend  Dockerfile-frontend
```

 2. Create `werf.yaml` main werf configuration file with the following content:

```
project: myapp
configVersion: 1
---
image: frontend
dockerfile: .dockerfiles/Dockerfile-frontend
---
image: backend
dockerfile: .dockerfiles/Dockerfile-backend
```

    Pay attention to an important param `project`, which is a _project name_ — werf will use it during converge process. This name cannot be changed easily later when your application is up and running without downtime and redeploying of the application.

 3. 

# 2. Converge application

A process of docker images building (and rebuilding when something changes), deploying an application into the Kubernetes (and redeploying when necessary) and making sure that application is up and running after deploying we will call a _converge process_.

Converge process is called for certain Git commit to perform synchronization of running application with the configuration defined in this commit. When converge process is done it is safe to assume that your application is up and running and conforms to the state defined in the current Git commit.

Werf implements converge process by `werf converge` command. Converge command should be called either manually, by CI/CD system or operator when a state of the application has been altered in the Git.

```
werf converge --repo myregistry.domain.org/example-app
```

Generally, converge command has one required argument: the address of the docker repo. Werf will use this docker repo to store built images and use them in Kubernetes (so, this docker repo should be accessible from within Kubernetes cluster). There is of course such optional param as [env](TODO) to deploy application into different environments of the same kubernetes cluster. By default werf will use _the project name_ as a Kubernetes namespace.

_NOTE: Your application may not have custom docker images (and use only publicly available images for example), in such case it is not required to pass `--repo` param — just omit it._

Werf will take Kubernetes cluster connection settings from environment (`~/.kube/config` or `KUBECONFIG`) the same way as `kubectl` tool does.

Converge will sync images in the docker repo to the defined state in the Git and sync Kubernetes to the defined state in the current Git commit.

# When to converge?

Converge process is applyable to the any Git commit. 

There are plenty of CI/CD workflows variations to implement with werf. Have a look at the article .
