---
title: Introduction
permalink: introduction.html
sidebar: documentation
---

{% asset introduction.css %}
{% asset introduction.js %}

## What is werf?

Werf is a CLI tool to implement full cycle of deployment of your application using Git as a single source of truth. Werf supports:
 - Building of docker images.
 - Deploying of the application into Kubernetes.
 - Making sure application is up and running well after deploy.
 - Rebuilding of docker images when application code changes.
 - Redeploying of the application into Kubernetes when necessary.
 - Cleaning up old unused images.

## How it works?

<div id="introduction-presentation" class="introduction-presentation">
    <div class="introduction-presentation__container">
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/dds-1.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                Werf configuration should be described in the application Git repository, right where application code resides.
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/dds-2.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    1. Put Dockerfiles into application repo
                </div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/dds-3.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    2. Create <code>werf.yaml</code> configuration file
                </div>
<div markdown="1">
Pay attention to an important param `project`, which is a _project name_ — werf will use it during _converge process_. This name cannot be changed easily later when your application is up and running without downtime and redeploying of the application.
</div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/dds-4.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    3. Describe helm chart templates to deploy your app
                </div>
<div markdown="1">
`werf_image` is a special template function to generate the full image name which has been built. This function has name parameter which corresponds to the image defined in the `werf.yaml` (`"frontend"` or `"backend"` in our case).
</div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/dds-5.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    4. Commit
                </div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    Converge your commit
                </div>
<div markdown="1">
Next step is to _converge_ our Git commit that contains defined state of the application.

Building of docker images (and rebuilding when something changes), deploying an application into the Kubernetes (and redeploying when necessary) and making sure application is up and running after deploying we will call a **converge process**.

Converge process is called for certain Git commit to perform synchronization of running application with the configuration defined in this commit. When converge process is done successfully it is safe to assume that your application is up and running and conforms to the state defined in the current Git commit.
</div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/c-1.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    1. Calculate configured images from current Git commit state.
                </div>
<div markdown="1">
During this step werf calculates needed images names. Images names may change or not change from one commit to another depending on the changes made into the Git repo. It is important to know, that each Git commit will have deterministic reproducible images names.
</div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/c-2.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    2. Read images existing in the Docker Registry.
                </div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/c-3.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    3. Calculate the difference between images needed for current Git commit state and images which are already exist in the Docker Registry.
                </div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/c-4.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    4. Build and publish only those images which do not exist in the Docker Registry (if any).
                </div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/c-5.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    5. Read needed configuration of Kubernetes resources from current Git commit state.
                </div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/c-6.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    6. Read configuration of existing Kubernetes resources from the cluster.
                </div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/c-7.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    7. Calculate the difference between needed configuration of Kubernetes resources defined for current Git commit state and configuration of existing Kubernetes resources of the cluster.
                </div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/c-8.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    8. Apply needed changes (if any) to the Kubernetes resources to conform with the state defined in the Git commit.
                </div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/c-9.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    9. Make sure all resources are ready to work and report any errors immediately (optionally rollback to the previous state if error occurred).
                </div>
            </div>
        </div>
    </div>
</div>

## What is converge?

Werf implements converge process by `werf converge` command. Converge command should be called either manually, by CI/CD system or operator when a state of the application has been altered in the Git. Run following command in the root of your project:

```
werf converge --docker-repo myregistry.domain.org/example-app [--kube-config ~/.kube/config]
```

Generally, converge command has one required argument: the address of the docker repo. Werf will use this docker repo to store built images and use them in Kubernetes (so, this docker repo should be accessible from within Kubernetes cluster).

Kube-config optional argument defines Kubernetes cluster to connect to.

There is of course such optional param as [env](TODO) to deploy application into different environments of the same kubernetes cluster. By default werf will use _the project name_ as a Kubernetes namespace.

_NOTE: Your application may not have custom docker images (and use only publicly available images for example), in such case it is not required to pass `--docker-repo` param — just omit it._

## What's next?

Deploy your first example application with [quickstart]({{ site.baseurl }}/documentation/quickstart.html) or checkout available [guides]({{ site.baseurl }}/documentation/guides.html) which cover configuration of wide variety of applications which use different programming languages and frameworks. It is recommended to find a suitable guide for your application and follow instructions.

### Converge your production

If you feel ready to dig into general overview of CI/CD workflows, which could be implemented with werf, then go [this article]({{ site.baseurl }}/documentation/advanced/ci_cd_workflows_overview.html).
