---
title: Overview
without_auto_heading: true
permalink: index.html
description: Extensive and clear werf documentation
editme_button: false
---

<link rel="stylesheet" type="text/css" href="{{ assets["overview.css"].digest_path | true_relative_url }}" />
<link rel="stylesheet" type="text/css" href="/css/guides.css" />


<div class="overview">
    <div class="overview__title">Must-Read</div>
    <div class="overview__row">
        <div class="overview__step">
            <div class="overview__step-header">
                <div class="overview__step-num">1</div>
                <div class="overview__step-time">5 minutes</div>
            </div>
            <div class="overview__step-title">Start your journey with basics</div>
            <div class="overview__step-actions">
                <a class="overview__step-action" href="/introduction.html">Introduction</a>
            </div>
        </div>
        <div class="overview__step">
            <div class="overview__step-header">
                <div class="overview__step-num">2</div>
                <div class="overview__step-time">15 minutes</div>
            </div>
            <div class="overview__step-title">Install and give werf a try on an example project</div>
            <div class="overview__step-actions">
                <a class="overview__step-action" href="/installation.html">Installation</a>
                <a class="overview__step-action" href="{{ "quickstart.html" | true_relative_url }}">Quickstart</a>
            </div>
        </div>
    </div>
    <div class="overview__step">
        <div class="overview__step-header">
            <div class="overview__step-num">3</div>
            <div class="overview__step-time">15 minutes</div>
        </div>
        <div class="overview__step-title">Learn the essentials of using werf in any CI/CD system</div>
        <div class="overview__step-actions">
            <a class="overview__step-action" href="{{ "using_with_ci_cd_systems.html" | true_relative_url }}">Using werf with CI/CD systems</a>
        </div>
    </div>
    <div class="overview__step">
        <div class="overview__step-header">
            <div class="overview__step-num">4</div>
            <div class="overview__step-time">several hours</div>
        </div>
        <div class="overview__step-title">Find a guide suitable for your project</div>
        <div class="overview__step-actions">
            <a class="overview__step-action" href="/guides.html">Guides</a>
        </div>
        <div class="overview__step-info">
            Find a guide suitable for your project (filter by a programming language, framework, CI/CD system, etc.) and deploy your first real application into the Kubernetes cluster with werf.
        </div>
    </div>
    <!--#include virtual="/guides/includes/landing-tiles.html" -->
    <div class="overview__title">Reference</div>
    <div class="overview__step">
        <div class="overview__step-title">Use Reference for structured information about werf configuration and commands</div>
        <div class="overview__step-actions">
            <a class="overview__step-action" href="{{ "reference/werf_yaml.html" | true_relative_url }}">Reference</a>
        </div>
        <div class="overview__step-info">
<div markdown="1">
 - An application should be properly configured via the [`werf.yaml`]({{ "reference/werf_yaml.html" | true_relative_url }}) file to use werf.
 - werf also uses [annotations]({{ "reference/deploy_annotations.html" | true_relative_url }}) in resources definitions to configure the deploy behaviour.
 - The [command line interface]({{ "reference/cli/overview.html" | true_relative_url }}) article contains the full list of werf commands with a description.
</div>
        </div>
    </div>
    <div class="overview__title">The extra mile</div>
    <div class="overview__step">
        <div class="overview__step-title">Get the deep knowledge, which you will need eventually during werf usage</div>
        <div class="overview__step-actions">
            <a class="overview__step-action" href="{{ "advanced/giterminism.html" | true_relative_url }}">Advanced</a>
        </div>
        <div class="overview__step-info">
<div markdown="1">
 - [Giterminism]({{ "advanced/giterminism.html" | true_relative_url }}) describes how determinism is implemented with git, what limitations it imposes, and why.
 - [Helm]({{ "advanced/helm/overview.html" | true_relative_url }})** describes the deploy essentials: how to configure werf for deploying to Kubernetes, what helm chart and release is. Here you may find the basics of templating Kubernetes resources, algorithms for using built images defined in your `werf.yaml` file during the deploy process and working with secrets, plus other useful stuff. Read this chapter if you want to learn more about organizing the deploy process with werf.
 - [Cleanup]({{ "advanced/cleanup.html" | true_relative_url }}) explains werf cleanup concepts and main commands to perform cleaning tasks.
 - [CI/CD]({{ "advanced/ci_cd/ci_cd_workflow_basics.html" | true_relative_url }}) describes main aspects of organizing CI/CD workflows with werf. Here you will learn how to use werf with GitLab CI/CD, GitHub Actions, or any other CI/CD system.
 - [Building images with stapel]({{ "reference/werf_yaml.html#image-section" | true_relative_url }}) introduces werf's custom builder. It currently implements the distributed building algorithm to enable lightning-fast build pipelines with distributed caching and incremental rebuilds based on the Git history of your application.
 - [Development and debug]({{ "advanced/development_and_debug/stage_introspection.html" | true_relative_url }}) describes debugging build and deploy processes of your application when something goes wrong and provides instructions for setting up a local development environment.
 - [Supported container registries]({{ "advanced/supported_container_registries.html" | true_relative_url }}) contains information about the peculiarities of using various container registries.
</div>
        </div>
    </div>
    <div class="overview__step">
        <div class="overview__step-title">Dive into overview of werf's inner workings</div>
        <div class="overview__step-actions">
            <a class="overview__step-action" href="{{ "internals/build_process.html" | true_relative_url }}">Internals</a>
        </div>
        <div class="overview__step-info">
            <p>You do not have to read through this section to make full use of werf. However, those interested in werf's internal mechanics will find some valuable info here.</p>
<div markdown="1">
 - [Building images]({{ "internals/build_process.html" | true_relative_url }}) — what image builder and stages are, how stages storage works, what is the synchronization server, other info related to the building process.
 - [How does the CI/CD integration work?]({{ "internals/how_ci_cd_integration_works/general_overview.html" | true_relative_url }}).
 - [The slug algorithm for naming]({{ "internals/names_slug_algorithm.html" | true_relative_url }}) describes the algorithm that werf uses under-the-hood to automatically replace invalid characters in input names so that other systems (such as Kubernetes namespaces or release names) can consume them.
 - [Integration with SSH agent]({{ "internals/integration_with_ssh_agent.html" | true_relative_url }}) shows how to integrate ssh-keys with the building process in werf.
 - [Development]({{ "internals/development/stapel_image.html" | true_relative_url }}) — this developers zone contains service/maintenance manuals and other docs written by werf developers and for werf developers. All this information sheds light on how specific werf subsystems work, describes how to keep the subsystem current, how to write and build new code for the werf, etc.
</div>
        </div>
    </div>
</div>
