---
title: Introduction
permalink: introduction.html
sidebar: documentation
---

{% asset introduction.css %}
{% asset introduction.js %}

## What is werf?

werf is a CLI tool for implementing a full deployment cycle for your application using Git as a single source of truth. werf can:

 - Build docker images.
 - Deploy the application into the Kubernetes cluster.
 - Make sure the application is up and running well after the deploy is complete.
 - Re-build docker images when application code changes.
 - Re-deploy the application into the Kubernetes cluster when necessary.
 - Clean up irrelevant and unused images.

## How it works?

<div id="introduction-presentation" class="introduction-presentation">
    <div class="introduction-presentation__container">
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/dds-1.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                werf configuration should be stored in the application Git repository along with the application code.
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/dds-2.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    1. Put Dockerfiles into the application repository
                </div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/dds-3.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    2. Create the <code>werf.yaml</code> configuration file
                </div>
<div markdown="1">
Pay attention to the crucial `project` parameter that stores a _project name_ since werf will be using it  extensively during the _converge process_. Changing this name later, when your application is up and running, involves downtime, and requires re-deploying the application.
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
`werf_image` is a special template function for generating the full name of the image that was built. This function has a name parameter that corresponds to the image defined in the `werf.yaml` (`"frontend"` or `"backend"` in our case).
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
            <img src="{% asset introduction/c-1.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    5. Calculate configured images based on the current Git commit state.
                </div>
<div markdown="1">
At this step, werf calculates the target image names. Image names may change or stay the same after the new commit is made depending on the changes in the Git repository. Note that each Git commit is associated with deterministic and reproducible image names.
</div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/c-2.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    6. Read images that are already available in the Docker Registry.
                </div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/c-3.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    7. Calculate the difference between images required for the current Git commit state and those that are already available in the Docker Registry.
                </div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/c-4.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    8. Build and publish only those images that do not exist in the Docker Registry (if any).
                </div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/c-5.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    9. Read the target configuration of Kubernetes resources based on the current Git commit state.
                </div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/c-6.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    10. Read the configuration of existing Kubernetes resources from the cluster.
                </div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/c-7.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    11. Calculate the difference between the target configuration of Kubernetes resources defined for the current Git commit state and the configuration of existing Kubernetes resources in the cluster.
                </div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/c-8.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    12. Apply changes (if any) to Kubernetes resources so that they conform to the state defined in the Git commit.
                </div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/c-9.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    13. Make sure all resources are ready to operate and report any errors immediately (with an option to roll back to the previous state if an error occurred).
                </div>
            </div>
        </div>
    </div>
</div>

## What is converge?

**Converge** is the process of building docker images (and re-building them in response to changes), deploying an application into the Kubernetes cluster (and re-deploying it when necessary), and making sure that the application is up and running.

The `werf converge` command starts the converging process. The command can be invoked by a user, by a CI/CD system, or by an operator in response to changes in the state of an application in Git. The behavior of the `werf converge` command is fully deterministic and transparent from the standpoint of the Git repository. After the converge is complete, your application will be up and running in the state defined in the target Git commit. Typically, to roll back your application to the previous version, you only need to run the converge on the corresponding earlier commit (werf will use correct images for this commit).

Run the following command in the root of your project to converge:

```
werf converge --docker-repo myregistry.domain.org/example-app [--kube-config ~/.kube/config]
```

Generally, the converge command has only one mandatory argument: the address of the docker repository. werf will use this docker repository to store built images and use them in Kubernetes (thus, this docker repository must be accessible from within the Kubernetes cluster). Kube-config is another optional argument that defines the Kubernetes cluster to connect to. Also, there is an optional `--env` parameter (and the `WERF_ENV` environment variable) allowing you to deploy an application into various [environments]({{ site.baseurl }}/documentation/advanced/ci_cd/ci_cd_workflow_basics.html#environment).

_NOTE: If your application does not use custom docker images (e.g., it uses only publicly available ones), you do not have to pass `--docker-repo` parameter to the command and may just omit it._

## What's next?

Deploy your first demo application with this [quickstart]({{ site.baseurl }}/documentation/quickstart.html) or check the [guides]({{ site.baseurl }}/documentation/guides.html) – they cover the configuration of a wide variety of applications based on different programming languages and frameworks. We recommend finding a manual suitable for your application and follow instructions.

Refer to [this article]({{ site.baseurl }}/documentation/advanced/ci_cd/ci_cd_workflow_basics.html) if you feel like you are ready to dig deeper into the general overview of CI/CD workflows that can be implemented with werf.
