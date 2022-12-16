---
title: CI/CD workflow basics
permalink: usage/integration_with_ci_cd_systems/ci_cd_workflow_basics.html
---

In this article, we will discuss what CI/CD means and overview CI/CD workflows that can be implemented with werf.

As you know, the primary purpose of using CI/CD for any project is to deliver application code changes to the end user as fast as possible. CI/CD consists of 2 parts: CI means continuous integration of application code changes into the main codebase, and CD means continuous delivery of these changes to the end user.

We will start with defining some basic terms like environment and workflow, and then move on to workflow building blocks and describe some ready-to-use workflows that can be constructed using provided workflow building blocks.

We also recommend you to read one the following guides about configuring the specific CI/CD system:
 - [GitLab CI/CD]({{ "how_to/integration_with_ci_cd_systems/gitlab_ci_cd/workflows.html" | true_relative_url }});
 - [GitHub Actions]({{ "how_to/integration_with_ci_cd_systems/github_actions/workflows.html" | true_relative_url }}).

## Basics

> NOTICE: In the text below, the term "git" can be interpreted as a common name for any version control system. However, werf supports only the Git version control system. We suppose our readers are familiar with the following git-related terms: branch, tag, master, commit, merge, rebase, fast-forward merge, and the git push-force procedure. To fully appreciate this article, readers are advised to get a good understanding of these terms.

### Environment

#### Production

Generally, end users interact with an application in the production environment. It is also the target environment for all application code changes made by application developers.

Also, there are so-called _production-like_ environments. Production-like environments are separate from production but have the same hardware and software configuration, network infrastructure, operating system, software versions, etc. In the real world, there may be some differences, though. Still, each project defines its acceptable risks related to the difference between production-like and production environments. Examples of production-like environments – [staging](#staging) and [testing](#testing) – are provided below.

#### Staging

Staging is a [production-like environment](#production) serving as a playground to examine your application before deploying it to production. Also, staging is the place to test new business functionality by managers, testers, and customers. End users do not interact with the staging environment in any way.

If your application uses some external services, then staging is the only environment typically (besides the production itself) that uses the same set of external services APIs and versions.

#### Testing

Testing is a [production-like environment](#production) with the following goal: to reveal problems with the new version of an application that may arise in the production environment. Some long-running application tests may use this environment to perform full-fledged application testing in a production-like environment. The higher the difference between testing and production environments, the higher the risk of deploying a buggy application to the production environment. That's why we recommend making a production-like environment as close as possible to a production one (same software versions, libraries, operating system, IP-addresses and ports, hardware, etc.).

#### Review

Review is a dynamic (temporary) environment used by application developers to test newly written code and conduct experiments not allowed in [production-like environments](#production).

It is possible to dynamically create an arbitrary number of review environments (as resources permit). Application developer usually creates and terminates such an environment using a CI/CD system. Also, a review environment could be automatically removed after a period of inactivity.

### Releases and CI/CD

The **release** — is a fully-functioning version of an application that contains some changes made since the previous application release.

In the waterfall model, releases are somewhat rare (monthly, half-yearly, yearly, and so on), and each release usually contains a big blob of changes. Deploying such a release is also a rare and unique event. Developer teams work hard to make the new version work properly while fixing all the bugs discovered. Usually, that is a separate and essential stage of the development process. And while this stage is in force, no new features are being implemented (some developers might work on an original, cool idea instead of debugging the latest release, but it will be available after the next release anyway). The latest version of an application contains many changes and has not been tested in real-life yet. So, there is a low probability that it will be running smoothly (even if an application uses automatic testing). Application users are also suffering during this stage.

On the other hand, in the CI/CD practice, releases contain small portions of changes and are frequently released. Developers are encouraged to implement new features in a way that makes it possible to instantly check and use the new version in production-like or production environments. Moreover, developers inspired by this approach usually want to push and test their newly written code to the production-like and production environments as fast as possible. The reason is simple: it is easier to make new changes on top of the existing ones that have proven to work well. Note that the code can only be considered working if it is running in production successfully.

In the sections below, we will define possible workflow configurations and take a look at to what extent each one of them is consistent with the CI/CD approach.

### Pipelines and workflows

As you know, the user makes changes to the project's code base using git. Every git commit triggers the so-called **pipeline** on the side of the CI/CD system.  The pipeline is split into stages that can be run sequentially or in parallel.

The main purpose of the pipeline is to deliver changes committed to git to some environment. If some stage of the pipeline fails, then the pipeline interrupts, and the user gets some feedback. A commit that successfully passes through all the stages ends up in a certain environment.

There are the following main types of stages:
 - Build-and-Publish — build app images and publish them to the image store.
 - Deploy — deploy an application to the environment.
 - Cleanup — clean up unused images and the associated cache from the repository.

#### Pull request

The user makes changes to the code base by creating commits in git and pull requests in the CI/CD system. Normally, a pull request is a thing that binds together a commit in git and a pipeline (also, it allows you to do a review, leave comments, and so on). Pull requests may have different names in various CI/CD systems (e.g., Merge Request in GitLab CI/CD) or may not exist at all (Jenkins).

#### Workflow

Pipelines and stages within pipelines can be triggered either automatically or manually. The so-called workflow defines pipelines activation, their structure, the way they integrate with git, as well as the required actions on the part of the user.
You can create multiple workflows to achieve the same goal. Next, we will focus on alternatives that can be implemented using werf.

### Run a pipeline manually

You can run a pipeline manually by one of the following methods:
 - By clicking a button in a CI/CD system (for example, in GitLab CI/CD).
 - By assigning a label in a CI/CD system (e.g., GitLab CI/CD or GitHub Actions). A label is usually assigned to/withdrawn from the pull request by the user or automatically by the CI/CD system while the pipeline is running.
   - For example, the user assigns the "run tests" label to his pull request, and the CI/CD system automatically triggers the corresponding pipeline. While this pipeline is running, the "run tests" label is revoked from the pull request. Then the user can assign this label again to restart the process,  and so on.
   - Label serves as an alternative to buttons or is used in cases when the CI/CD system does not support buttons (e.g., GitHub Actions).
   For review environments, assigning a label by the user is a signal to activate the review environment, and removing it (manually or automatically) is a signal to deactivate it. This option will be discussed in more detail (later) (#automatically-deploy-to-review-using-a-pull-request-manual-triggering).
 - By sending a request to the API of the CI/CD system (for example, via HTTP at a specific url in GitHub Actions).

### The testing stage

To implement a proper CI/CD, it is critical to automatically get instant feedback when making changes to the project codebase. The testing stage can be divided into two nominal stages: during the primary stage, tests run fast and cover most of the functions; during the secondary stage, tests can run for a prolonged time and verify more aspects of the application. Generally, primary tests are run automatically, and passing them is a prerequisite for changes to be released.

## Further reading

We recommend you to read the guide for your specific CI/CD system:
 - The [GitLab CI/CD]({{ "how_to/integration_with_ci_cd_systems/gitlab_ci_cd/workflows.html" | true_relative_url }}) guide.
 - The [GitHub Actions]({{ "how_to/integration_with_ci_cd_systems/github_actions/workflows.html" | true_relative_url }}) guide.

If you want to learn more about how to create a workflow, or no instructions exist for your particular CI/CD system, then you can read the following sections where the [workflow components](#workflow-components-for-various-environments) and [ready-made workflow configurations](#ready-made-workflow-configurations) are defined. After reading them, you will be able to choose a ready-made configuration (or create your own) and implement it for your CI/CD system using werf.
