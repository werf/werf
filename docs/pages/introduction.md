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

# Define desired state in the Git

Werf configuration should be described in the application Git repository right where application code resides.

1. Put Dockerfiles into application repo

   ![introduction-1]({% asset introduction/1.png @path %})

2. Create `werf.yaml` main werf configuration file

    ![introduction-2]({% asset introduction/2.png @path %})

    Pay attention to an important param `project`, which is a _project name_ — werf will use it during converge process. This name cannot be changed easily later when your application is up and running without downtime and redeploying of the application.

3. Describe helm chart templates to deploy your app

    ![introduction-3]({% asset introduction/3.png @path %})

    `werf_image` is a special template function to generate the full image name which has been built (like `myregistry.domain.org/example-app:f6e683effe19ca5f81b3c6df0ce398dd2ba50f5a991cec47b63c31b8-1601390730191`). This function has name parameter which corresponds to the image defined in the `werf.yaml` (`"frontend"` or `"backend"` in our case).

# Converge application

A process of docker images building (and rebuilding when something changes), deploying an application into the Kubernetes (and redeploying when necessary) and making sure that application is up and running after deploying we will call a _converge process_.

Converge process is called for certain Git commit to perform synchronization of running application with the configuration defined in this commit. When converge process is done it is safe to assume that your application is up and running and conforms to the state defined in the current Git commit.

Werf implements converge process by `werf converge` command. Converge command should be called either manually, by CI/CD system or operator when a state of the application has been altered in the Git. Run following command in the root of your project 

```
werf converge --repo myregistry.domain.org/example-app
```

Generally, converge command has one required argument: the address of the docker repo. Werf will use this docker repo to store built images and use them in Kubernetes (so, this docker repo should be accessible from within Kubernetes cluster). There is of course such optional param as [env](TODO) to deploy application into different environments of the same kubernetes cluster. By default werf will use _the project name_ as a Kubernetes namespace.

_NOTE: Your application may not have custom docker images (and use only publicly available images for example), in such case it is not required to pass `--repo` param — just omit it._

Werf will take Kubernetes cluster connection settings from environment (`~/.kube/config` or `KUBECONFIG`) the same way as `kubectl` tool does.

Converge will sync images in the docker repo to the defined state in the Git and sync Kubernetes to the defined state in the current Git commit.

# What's next?

## Converge your production

There are plenty of CI/CD workflows variations to implement with werf, have a look [at the article]({{ site.baseurl }}/documentation/advanced/ci_cd_workflows_overview.html).
