---
title: Overview
permalink: documentation/index.html
sidebar: documentation
---

Documentation of the werf project consists of several parts.

Start with **[introduction]({{ site.baseurl }}/introduction.html)** to dive into werf easily. It describes how to use your Dockerfiles  Learn how to integrate werf into your project smoothly and effortlessly.

**[Installation]({{ site.baseurl }}/installation.html)** describes werf depedencies and various installation methods, stability channels and backward compatibility promise.

**[Guides]()** contain plenty of guides to quickly setup deployment of your application into Kubernetes,
either for local development or production. This is recommended section to read after _introduction_.

To use werf, an application should be properly configured via `werf.yaml` file. The **[configuration article]({{ site.baseurl }}/documentation/reference/configuration/introduction.html)** includes:

1. Definition of the project meta information such as a project name (it will affect converge and dismiss commands).
2. Definition of images to be built.

**[Deploy section]()** describes deploy essentials: how to configure werf to deploy into Kubernetes, what is chart and release, templating of Kubernetes resources, how to use built images defined in your `werf.yaml` file during deploy process, working with secrets and other. Read this section if you want learn more about deploy process with werf.

**[CI/CD]()** describes main aspects of organizing CI/CD workflows with werf, using GitLab CI/CD or GitHub Actions or any other CI/CD system with werf.

**[Local development]()** describes how to use werf to ease local development of your applications, use the same configuration to deploy application either locally or into production.

Check out **[command line interface]()** article to see reference of the main werf CLI commands. Full CLI commands list also available in [internals section]().

Advanced section covers more complex topics which you will need eventually:
 - [Configuration]() contains information about templating of werf config files, deployment names generation (such as Kubernetes namespace or release name).
 - [Cleanup]() contains description of werf cleanup concepts and main commands to perform cleaning tasks.
 - [Supported registry implementation]() contains general info about supporting of implementations and info about authorization when using different implementations.
 - [Building images with stapel]() introduces werf custom builder which currently implements distributed building algorithm to enable really fast build pipelines with distributed caching and incremental rebuilds based on Git-history of your application.
 - [Development and debug]() describes how to debug build and deploy processes of your application when something goes wrong, how to setup local development environment.

Internals section contains general info about how werf works inside. It is not necessary to read this section to use werf fully. But if you interested how werf really works inside — you will find some useful info here.
 - [Building of images]() — what is images builder and stages, how stages storage works, what is syncrhonization server and other info related to building process.
 - [How CI/CD integration works]().
 - [Names slug algorithm]() contains description of this algorithm which werf uses underhood automatically to transform unacceptable characters in input names for other systems to consume (such as Kubernetes namespace, or release names).
 - [Integration with SSH agent]() — how to integrate ssh-keys with building process in werf.
 - [Development]() — this is developers zone, contains service and maintenance manuals and other docs written by developers for developers of the werf. This info helps to understand how a specific werf subsystem works, how to maintain subsystem in the actual state, how to write and build new code for the werf, etc.
 - [CLI reference]() contains full CLI commands list with descriptions.
