---
permalink: documentation/index.html
sidebar: documentation
---

Start with the **[introduction]({{ site.baseurl }}/introduction.html)**.
 > Dive into werf basics.
 > (5 min)

Continue with the **[installation]({{ site.baseurl }}/installation.html)** and the **[quickstart]({{ site.baseurl }}/introduction.html)**.
 > Install and give werf a try on an example project.
 > (15 min)

**[Using werf with CI/CD systems]({{ site.baseurl }}/documentation/using_with_ci_cd_systems.html)**
 > Learn the essentials of using werf in any CI/CD system.
 > (15 min)

The **[Guides]({{ site.baseurl }}/documentation/guides.html)** section contains plenty of information to set up the deployment for your application.
 > Find a guide suitable for your project (filter by a programming language, framework, CI/CD system, etc.) and deploy your first real application into the Kubernetes cluster with werf.
 > (several hours)

The **[Reference]({{ site.baseurl }}/documentation/reference/werf_yaml.html)** list contains structured information about the werf configuration and commands.
 - An application should be properly configured via the [`werf.yaml`]({{ site.baseurl }}/documentation/reference/werf_yaml.html) file to use werf.
 - werf also uses [annotations]({{ site.baseurl }}/documentation/reference/deploy_annotations.html) in resources definitions to configure the deploy behaviour.
 - The [command line interface]({{ site.baseurl }}/documentation/reference/cli/overview.html) article contains the full list of werf commands with a description.

<!-- The **[Local development]()** section describes how werf simplifies and facilitates the local development of your applications, allowing you to use the same configuration to deploy an application either locally or remotely (into production). -->

The **[Advanced]({{ site.baseurl }}/documentation/advanced/configuration/supported_go_templates.html)** section covers more complex tasks that you will face eventually.
 - [Configuration]({{ site.baseurl }}/documentation/advanced/configuration/supported_go_templates.html) informs about templating principles of werf configuration files as well as generating deployment-related names (such as a Kubernetes namespace or a release name).
 - [Helm]({{ site.baseurl }}/documentation/advanced/helm/basics.html)** describes the deploy essentials: how to configure werf for deploying to Kubernetes, what helm chart and release is. Here you may find the basics of templating Kubernetes resources, algorithms for using built images defined in your `werf.yaml` file during the deploy process and working with secrets, plus other useful stuff. Read this section if you want to learn more about organizing the deploy process with werf.
 - [Cleanup]({{ site.baseurl }}/documentation/advanced/cleanup.html) explains werf cleanup concepts and main commands to perform cleaning tasks.
 - [CI/CD]({{ site.baseurl }}/documentation/advanced/ci_cd/ci_cd_workflow_basics.html) describes main aspects of organizing CI/CD workflows with werf. Here you will learn how to use werf with GitLab CI/CD, GitHub Actions, or any other CI/CD system.
 - [Building images with stapel]({{ site.baseurl }}/documentation/advanced/building_images_with_stapel/naming.html) introduces werf's custom builder. It currently implements the distributed building algorithm to enable lightning-fast build pipelines with distributed caching and incremental rebuilds based on the Git history of your application.
 - [Development and debug]({{ site.baseurl }}/documentation/advanced/development_and_debug/stage_introspection.html) describes debugging build and deploy processes of your application when something goes wrong and prvodes instructions for setting up a local development environment.
 - [Supported registry implementations]({{ site.baseurl }}/documentation/advanced/supported_registry_implementations.html) contains general info about supported implementations and authorization when using different implementations.

The **[Internals]({{ site.baseurl }}/documentation/internals/building_of_images/build_process.html)** section provides an overview of werf's inner workings. You do not have to read through this section to make full use of werf. However, those interested in werf's internal mechanics will find some valuable info here.
 - [Building images]({{ site.baseurl }}/documentation/internals/building_of_images/build_process.html) — what image builder and stages are, how stages storage works, what is the syncrhonization server, other info related to the building process.
 - [How does the CI/CD integration work?]({{ site.baseurl }}/documentation/internals/how_ci_cd_integration_works/general_overview.html).
 - [The slug algorithm for naming]({{ site.baseurl }}/documentation/internals/names_slug_algorithm.html) describes the algorithm that werf uses under-the-hood to automatically replace invalid characters in input names so that other systems (such as Kubernetes namespaces or release names) can consume them.
 - [Integration with SSH agent]({{ site.baseurl }}/documentation/internals/integration_with_ssh_agent.html) shows how to integrate ssh-keys with the building process in werf.
 - [Development]({{ site.baseurl }}/documentation/internals/development/stapel_image.html) — this developers zone contains service/maintenance manuals and other docs written by werf developers and for werf developers. All this information sheds light on how specific werf subsystems work, describes how to keep the subsystem current, how to write and build new code for the werf, etc.
