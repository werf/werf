---
permalink: documentation/index.html
sidebar: documentation
---

Start with **[introduction]({{ site.baseurl }}/introduction.html)**.
 > Dive into werf basics.
 > (5 min)

Continue with **[installation]({{ site.baseurl }}/installation.html)** and **[quickstart]({{ site.baseurl }}/introduction.html)**.
 > Install and try werf with example project.
 > (15 min)

**[Guides](https://ru.werf.io/applications_guide_ru)** section contains plenty of guides to setup deployment of your application.
 > Find a guide suitable for your project (by programming language, by framework, by CI/CD system, etc.) and deploy your first real application into Kubernetes with werf.
 > (several hours)

**[Using werf with CI/CD systems]()**
 > Learn important things to use werf with any CI/CD system.
 > (15 min)

**[Reference]()** contains structured information about werf configuration and commands.
 - To use werf, an application should be properly configured via [`werf.yaml`]() file.
 - Werf also uses [annotations]({{ site.baseurl }}/documentation/reference/configuration/deploy_annotations.html) in resources definitions to configure deploy behaviour.
 - [Command line interface]() article contains reference of werf commands.

<!-- **[Local development]()** describes how to use werf to ease local development of your applications, use the same configuration to deploy application either locally or into production. -->

**[Advanced]()** section covers more complex topics which you will need eventually.
 - [Configuration]() contains information about templating of werf config files, deployment names generation (such as Kubernetes namespace or release name).
 - [Helm]()** describes deploy essentials: how to configure werf to deploy into Kubernetes, what is helm chart and release, templating of Kubernetes resources, how to use built images defined in your `werf.yaml` file during deploy process, working with secrets and other. Read this section if you want learn more about deploy process with werf.
 - [Cleanup]() contains description of werf cleanup concepts and main commands to perform cleaning tasks.
 - [CI/CD]() describes main aspects of organizing CI/CD workflows with werf, using GitLab CI/CD or GitHub Actions or any other CI/CD system with werf.
 - [Building images with stapel]() introduces werf custom builder which currently implements distributed building algorithm to enable really fast build pipelines with distributed caching and incremental rebuilds based on Git-history of your application.
 - [Development and debug]() describes how to debug build and deploy processes of your application when something goes wrong, how to setup local development environment.
 - [Supported registry implementation]() contains general info about supporting of implementations and info about authorization when using different implementations.

**[Internals]()** section contains general info about how werf works inside. It is not necessary to read through this section to use werf fully. But if you interested how werf really works inside — you will find some useful info here.
 - [Building of images]() — what is images builder and stages, how stages storage works, what is syncrhonization server and other info related to building process.
 - [How CI/CD integration works]().
 - [Names slug algorithm]() contains description of this algorithm which werf uses underhood automatically to transform unacceptable characters in input names for other systems to consume (such as Kubernetes namespace, or release names).
 - [Integration with SSH agent]() — how to integrate ssh-keys with building process in werf.
 - [Development]() — this is developers zone, contains service and maintenance manuals and other docs written by developers for developers of the werf. This info helps to understand how a specific werf subsystem works, how to maintain subsystem in the actual state, how to write and build new code for the werf, etc.
 - [CLI reference]() contains full CLI commands list with descriptions.
