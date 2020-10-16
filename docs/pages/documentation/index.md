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

**[Using werf with CI/CD systems]({{ site.baseurl }}/documentation/using_with_ci_cd_systems.html)**
 > Learn important things to use werf with any CI/CD system.
 > (15 min)

**[Guides]({{ site.baseurl }}/documentation/guides.html)** section contains plenty of guides to setup deployment of your application.
 > Find a guide suitable for your project (by programming language, by framework, by CI/CD system, etc.) and deploy your first real application into Kubernetes with werf.
 > (several hours)

**[Reference]({{ site.baseurl }}/documentation/reference/werf_yaml.html)** contains structured information about werf configuration and commands.
 - To use werf, an application should be properly configured via [`werf.yaml`]({{ site.baseurl }}/documentation/reference/werf_yaml.html) file.
 - Werf also uses [annotations]({{ site.baseurl }}/documentation/reference/deploy_annotations.html) in resources definitions to configure deploy behaviour.
 - [Command line interface]({{ site.baseurl }}/documentation/reference/cli/index.html) article contains reference of werf commands.

<!-- **[Local development]()** describes how to use werf to ease local development of your applications, use the same configuration to deploy application either locally or into production. -->

**[Advanced]({{ site.baseurl }}/documentation/advanced/configuration/supported_go_templates.html)** section covers more complex topics which you will need eventually.
 - [Configuration]({{ site.baseurl }}/documentation/advanced/configuration/supported_go_templates.html) contains information about templating of werf config files, deployment names generation (such as Kubernetes namespace or release name).
 - [Helm]({{ site.baseurl }}/documentation/advanced/helm/basics.html)** describes deploy essentials: how to configure werf to deploy into Kubernetes, what is helm chart and release, templating of Kubernetes resources, how to use built images defined in your `werf.yaml` file during deploy process, working with secrets and other. Read this section if you want learn more about deploy process with werf.
 - [Cleanup]({{ site.baseurl }}/documentation/advanced/cleanup.html) contains description of werf cleanup concepts and main commands to perform cleaning tasks.
 - [CI/CD]({{ site.baseurl }}/documentation/advanced/ci_cd/ci_cd_workflow_basics.html) describes main aspects of organizing CI/CD workflows with werf, using GitLab CI/CD or GitHub Actions or any other CI/CD system with werf.
 - [Building images with stapel]({{ site.baseurl }}/documentation/advanced/building_images_with_stapel/naming.html) introduces werf custom builder which currently implements distributed building algorithm to enable really fast build pipelines with distributed caching and incremental rebuilds based on Git-history of your application.
 - [Development and debug]({{ site.baseurl }}/documentation/advanced/development_and_debug/stage_introspection.html) describes how to debug build and deploy processes of your application when something goes wrong, how to setup local development environment.
 - [Supported registry implementations]({{ site.baseurl }}/documentation/advanced/supported_registry_implementations.html) contains general info about supporting of implementations and info about authorization when using different implementations.

**[Internals]({{ site.baseurl }}/documentation/internals/building_of_images/build_process.html)** section contains general info about how werf works inside. It is not necessary to read through this section to use werf fully. But if you interested how werf really works inside — you will find some useful info here.
 - [Building of images]({{ site.baseurl }}/documentation/internals/building_of_images/build_process.html) — what is images builder and stages, how stages storage works, what is syncrhonization server and other info related to building process.
 - [How CI/CD integration works]({{ site.baseurl }}/documentation/internals/how_ci_cd_integration_works/general_overview.html).
 - [Names slug algorithm]({{ site.baseurl }}/documentation/internals/names_slug_algorithm.html) contains description of this algorithm which werf uses underhood automatically to transform unacceptable characters in input names for other systems to consume (such as Kubernetes namespace, or release names).
 - [Integration with SSH agent]({{ site.baseurl }}/documentation/internals/integration_with_ssh_agent.html) — how to integrate ssh-keys with building process in werf.
 - [Development]({{ site.baseurl }}/documentation/internals/development/stapel_image.html) — this is developers zone, contains service and maintenance manuals and other docs written by developers for developers of the werf. This info helps to understand how a specific werf subsystem works, how to maintain subsystem in the actual state, how to write and build new code for the werf, etc.
