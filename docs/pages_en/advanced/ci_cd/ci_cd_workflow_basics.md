---
title: CI/CD workflow basics
permalink: advanced/ci_cd/ci_cd_workflow_basics.html
---

In this article, we will discuss what CI/CD means and overview CI/CD workflows that can be implemented with werf.

As you know, the primary purpose of using CI/CD for any project is to deliver application code changes to the end user as fast as possible. CI/CD consists of 2 parts: CI means continuous integration of application code changes into the main codebase, and CD means continuous delivery of these changes to the end user.

We will start with defining some basic terms like environment and workflow, and then move on to workflow building blocks and describe some ready-to-use workflows that can be constructed using provided workflow building blocks.

We also recommend you to read one the following guides about configuring the specific CI/CD system:
 - [GitLab CI/CD]({{ "advanced/ci_cd/gitlab_ci_cd.html" | true_relative_url }});
 - [GitHub Actions]({{ "advanced/ci_cd/github_actions.html" | true_relative_url }}).

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

The real CI/CD process implies a quick and effortless deployment of new code to the production-like environment (staging, testing, or the production itself). The common anti-pattern in the CI/CD practice is the deployment of many accumulated changes in the source code to the production-like or production environment all at once. In other words, do not build a waterfall-like model on top of the components specially designed for CI/CD. It could be due to the wrong project organization that does not allow developers to deploy code to production-like environments all by themselves, without input from the management (automatic testing is still in place, obviously). Or it could be due to the fact that the CI/CD system is based on hard-to-use tools and primitives. In this case, the system looks like a regular CI/CD process, but in fact, it is totally developer-unfriendly (for example, using git tags with a message for each release).

A good CI/CD system should be able to deliver changes to production within minutes; it must be able to roll back/test these changes on-the-fly while avoiding unnecessary barriers and pointless repetitive actions.

The synchronization of the state of the code in git and an application deployed to the production-like/production environments is also a sign of a well-built CI/CD system. For example, the state of the application deployed to the staging environment corresponds to the last change made to the staging branch (that is just one of the possible options for organizing the structure). If the code in the staging branch does not work, then the application deployed to the staging environment does not work either. In this case, you cannot use the app. At the same time, you cannot use the code from the staging branch either since the non-working code cannot serve as a basis for new changes. That is why we keep the code in the staging branch (the last commit to this branch) and the application in a synchronized state instead of rolling back the staging environment to the previous commit. Thus, in such circumstances, it makes perfect sense to either fix the problem quickly and roll out the correct code to the branch (with the following automatic deployment to the environment) or revert changes (also with the automatic deployment to the environment).

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
 - The [GitLab CI/CD]({{ "advanced/ci_cd/gitlab_ci_cd.html" | true_relative_url }}) guide.
 - The [GitHub Actions]({{ "advanced/ci_cd/github_actions.html" | true_relative_url }}) guide.

If you want to learn more about how to create a workflow, or no instructions exist for your particular CI/CD system, then you can read the following sections where the [workflow components](#workflow-components-for-various-environments) and [ready-made workflow configurations](#ready-made-workflow-configurations) are defined. After reading them, you will be able to choose a ready-made configuration (or create your own) and implement it for your CI/CD system using werf.

## Workflow components for various environments

Next, we will discuss various options for rolling out to production and other environments in conjunction with the git. Each section describes a building block that you can use to work with a specific environment. We will call such a block a "workflow block". You can build your own workflow using workflow blocks or make use of a ready-made configuration (see [Ready-made workflow configurations](#ready-made-workflow-configurations) below).

Review environments are created and deleted dynamically as per the developers' request. Thus, there are some specificities when deploying code to these environments. In the review-related sections, we will describe how to create as well as delete a review environment.

### Automatically deploy to production from master

A merge request or a commit to the master branch triggers the pipeline to deploy directly to production.

The state of the branch reflects the state of the environment at any given time. Therefore, this option complies with the true CI/CD approach.

Rollback options:
 - Recommended: rollback via reverting the commit in the master branch. In this case, the state of the branch is kept synchronized with the state of the environment, so this is the preferred option for preserving schema integrity.
 - By means of the CI/CD system; the user can [manually re-run the pipeline](#run-a-pipeline-manually) on an old commit (e.g., in GitLab CI/CD, the "rollback" button performs precisely these steps).

### Deploy to production from master at the click of a button

You can manually run a pipeline to deploy to production for a commit in the master branch only. The pipeline is triggered [manually](#run-a-pipeline-manually) using the CI/CD system's tools, such as a dedicated button or an API call.

Rollback options:
 - Recommended: By means of the CI/CD system; the user can [manually re-run the pipeline](#run-a-pipeline-manually) on an old commit (e.g., in GitLab CI/CD, the "rollback" button performs precisely these steps).
 - By reverting a commit in the master branch and then [manually](#run-a-pipeline-manually) run the pipeline using the CI/CD system: either by hitting a button in the CI/CD or by an API call. We do not recommend this option because the state of the master is not consistent with the state of the environment at all times (as opposed to the automatic deployment). Thus, it does not make sense to revert exclusively for the rollback.

### Automatically deploy to production using a tag

Creating a new tag automatically triggers the pipeline to deploy the code to the production environment on the basis of the commit associated with this tag.

Rollback options:
 - Recommended: by means of CI/CD system; [re-run the pipeline manually](#run-a-pipeline-manually) using the old tag.
 - By creating a new tag linked to the old commit; the pipeline will be run automatically for the new tag. It is not the best option since it leads to the generation of unnecessary tags.

### Deploy to production using a tag at the click of a button

The pipeline to deploy to the production environment can only be triggered using the existing tag in git. The pipeline is triggered [manually](#run-a-pipeline-manually) using the CI/CD system's tools, such as a dedicated button or an API call.

Rollback options:
- By means of the CI/CD system; [re-run the pipeline manually](#run-a-pipeline-manually) using the old tag.

### Automatically deploy to production from a branch

A merge request or a commit to the specific branch triggers the pipeline to deploy to production (this option is similar to the (master-automatically) (#master-automatically), but involves a separate branch).

The state of the specific branch is consistent with the state of the environment at any given time. Therefore, this option complies with the true CI/CD approach.

Rollback options:
- Recommended: rollback via reverting the commit in the branch. In this case, the state of the branch is kept synchronized with the state of the environment, so this is the preferred option for preserving schema integrity.
 - By means of the CI/CD system; the user can [manually run the pipeline](#run-a-pipeline-manually) on an old commit (e.g., in GitLab CI/CD, the "undo" button performs precisely these steps).
 - By reverting the commit in the master, then fast-forwarding merge to the specific branch.
 - By deleting the commit in the specific branch (delete the commit in git and then use the git push-force procedure).

### Deploy to production from a branch at the click of a button

Manually run a pipeline to deploy to production for a commit to the specific branch only. The pipeline is triggered [manually](#run-a-pipeline-manually) using the CI/CD system's tools, such as a dedicated button or an API call.

Rollback options:
 - Recommended: By means of the CI/CD system; the user can [manually re-run the pipeline](#run-a-pipeline-manually) on an old commit (e.g., in GitLab CI/CD, the "rollback" button performs precisely these steps).
 - By reverting the commit in the specific branch and then [manually](#run-a-pipeline-manually) run the pipeline using the means of the CI/CD system: either by hitting a button in the CI/CD or by an API call. We do not recommend this option because the state of the master is not consistent with the state of the environment at all times (as opposed to the automatic deployment). Thus, it does not make sense to revert exclusively for the rollback.

### Deploy to production-like using a pull request at the click of a button

The pipeline to deploy to production can be run on any commit to the pull request. The pipeline is triggered [manually](#run-a-pipeline-manually) using the CI/CD system's tools, such as a dedicated button or an API call.

Rollback options:
 - By means of the CI/CD system; the user can [manually](#run-a-pipeline-manually) run the pipeline on an old commit (e.g., in GitLab CI/CD, the "rollback" button performs precisely these steps).

### Automatically deploy to staging from master

A merge request or a commit to the master branch triggers the pipeline to deploy directly to the staging environment.

The state of the branch reflects the state of the environment at any given time. Therefore, this option complies with the true CI/CD approach.

Rollback options:
 - Recommended: rollback via reverting the commit in the master branch. In this case, the state of the branch is kept synchronized with the state of the environment, so this is the preferred option for preserving schema integrity.
 - By means of the CI/CD system; the user can [manually](#run-a-pipeline-manually) run the pipeline on an old commit (e.g., in GitLab CI/CD, the "rollback" button performs precisely these steps).

### Deploy to staging from master at the click of a button

The pipeline to deploy to staging can only be [run manually](#run-a-pipeline-manually) on a commit from the master branch. The pipeline is triggered manually using the CI/CD system's tools, such as a dedicated button or an API call.

Rollback options:
 - Recommended: By means of the CI/CD system; the user can [manually re-run the pipeline](#run-a-pipeline-manually) on an old commit (e.g., in GitLab CI/CD, the "rollback" button performs precisely these steps).
 - Revert the commit in the master branch and then [manually](#run-a-pipeline-manually) run the pipeline using the CI/CD system: either by hitting a button in the CI/CD or by an API call. We do not recommend this option because the state of the master does not consistent with the state of the environment at all times (as opposed to the automatic deployment from master to staging). Thus, it does not make sense to revert exclusively for the rollback.

### Automatically deploy to production-like from a branch

A merge request or a commit to the specific branch triggers the pipeline to deploy to the production-like environment (this option is similar to (master-automatically) (#master-automatically), but involves a separate branch). A separate branch is used for each specific production-like environment, such as staging or testing.

The state of the specific branch is consistent with the state of the environment at any given time. Therefore, this option complies with the true CI/CD approach.

Rollback options:
 - Recommended: rollback via reverting the commit in the branch. In this case, the state of the branch is kept synchronized with the state of the environment, so this is the preferred option for preserving schema integrity.
 - By means of the CI/CD system; the user can [manually](#run-a-pipeline-manually) run the pipeline on an old commit (e.g., in GitLab CI/CD, the "rollback" button performs precisely these steps).
 - Revert the commit in the master, then fast-forward merge to the specific branch.
 - Delete the commit in the specific branch (by deleting the commit in git and then using the git push-force procedure).

### Deploy to production-like from a branch at the click of a button

Manually run a pipeline to deploy to the production-like environment for a commit to the specific branch only. The pipeline is triggered [manually](#run-a-pipeline-manually) using the CI/CD system's tools, such as a dedicated button or an API call.

Rollback options:
 - Recommended: By means of the CI/CD system; the user can [manually re-run the pipeline](#run-a-pipeline-manually) on an old commit (e.g., in GitLab CI/CD, the "rollback" button performs precisely these steps).
 - Revert the commit in the specific branch and then [manually](#run-a-pipeline-manually) run the pipeline using the means of the CI/CD system: either by hitting a button in the CI/CD or by an API call. We do not recommend this option because the state of the branch is not consistent with the state of the environment at all times (as opposed to the automatic deployment to the production-like environment). Thus, it does not make sense to revert exclusively for the purposes of a rollback.

### Automatically deploy to review using a pull request

Further commits to the branch tied to the pull request automatically trigger roll-outs to the review environment.

Rollback options:
 - Recommended: rollback via reverting the commit in the branch. In this case, the state of the branch is kept synchronized with the state of the environment, so this is the preferred option for preserving schema integrity.
 - By means of the CI/CD system; the user can [manually re-run the pipeline](#run-a-pipeline-manually) on an old commit (e.g., in GitLab CI/CD, the "rollback" button performs precisely these steps).

Deleting the review environment:
 - By closing or merging a PR.
 - Automatically when the time-lo-live period since the last deployment expires (in other words, if there is no activity in the environment).

### Automatically deploy to review from a branch using a pattern

Creating a pull request for a branch that matches some  specific pattern automatically triggers a roll-out to a separate review environment. The name of this environment is associated with the name of the branch. Further commits to the branch associated with the pull request automatically trigger a roll-out to the review environment.

For example, if you have a `review_*` pattern, then the creation of a pull request for the branch called `review_myfeature1` would automatically lead to the creation of the corresponding review-environment.

Rollback options:
 - Recommended: rollback via reverting the commit in the branch. In this case, the state of the branch is kept synchronized with the state of the environment, so this is the preferred option for preserving schema integrity.
 - By means of the CI/CD system; the user can [manually](#run-a-pipeline-manually) run the pipeline on an old commit (e.g., in GitLab CI/CD, the "rollback" button performs precisely these steps).

Deleting the review environment:
 - By closing or merging a PR.
 - Automatically when the time-lo-live period since the last deployment expires (in other words, if there is no activity in the environment).

### Deploy to review using a pull request at the click of a button

The pipeline to deploy to the review environment can only be run manually on a commit from the branch tied to this environment. The name of this environment is associated with the name of the branch. The pipeline is triggered [manually](#run-a-pipeline-manually) using the CI/CD system's tools, such as a dedicated button, an API call, or by assigning a label.

Rollback options:
 - Recommended: By means of the CI/CD system; the user can [manually re-run the pipeline](#run-a-pipeline-manually) on an old commit (e.g., in GitLab CI/CD, the "rollback" button performs precisely these steps).
 - Revert the commit in the specific branch and then [manually](#run-a-pipeline-manually) run the pipeline using the means of the CI/CD system: either by hitting a button in the CI/CD or by an API call. We do not recommend this option because the state of the branch does not always correspond to the state of the environment (in contrast to "deploy automatically using the pull request" and "deploy automatically using the pull request and a pattern"), so it does not make sense to perform a revert solely for the rollback.

Deleting the review environment:
 - By closing or merging a PR.
 - Automatically when the time-lo-live period since the last deployment expires (in other words, if there is no activity in the environment).

### Automatically deploy to review using a pull request; manual triggering

The review environment is created for the pull request after the pipeline is manually run by the means of the CI/CD system. From this point on, any commit to the branch tied to the pull request leads to an automatic roll-out to the review environment. After the work is complete, you can deactivate the review environment manually using the CI/CD system.

The pipeline to deploy to the review environment can only be run manually on a commit from the branch tied to this environment. The name of this environment is associated with the name of the branch. The pipeline to activate the environment is triggered [manually](#run-a-pipeline-manually) using the CI/CD system's tools, such as a dedicated button, an API call, or by assigning a label. 

Rollback options:
 - Recommended: rollback via reverting the commit in the branch. In this case, the state of the branch is kept synchronized with the state of the environment, so this is the preferred option for preserving schema integrity.
 - By means of the CI/CD system; the user can [manually](#run-a-pipeline-manually) run the pipeline on an old commit (e.g., in GitLab CI/CD, the "rollback" button performs precisely these steps).

Deleting the review environment:
 - The pipeline to deactivate the environment is triggered [manually](#run-a-pipeline-manually) using the CI/CD system's tools, such as a dedicated button, an API call, or by assigning a label.
 - By closing or merging a PR.
 - Automatically when the time-lo-live period since the previous deployment expires (in other words, if there is no activity in the environment).

## Comparing building blocks for creating a workflow

### The level of git control

|   | **Rollback via git** | **Rollback manually** |
|:---|:---|:---|
| **Deploy via git** | [Automatically deploy to production from master](#automatically-deploy-to-production-from-master);<br>[Automatically deploy to production using a tag](#automatically-deploy-to-production-using-a-tag);<br>[Automatically deploy to production from a branch](#automatically-deploy-to-production-from-a-branch);<br>[Automatically deploy to staging from master](#automatically-deploy-to-staging-from-master);<br>[Automatically deploy to production-like from a branch](#automatically-deploy-to-production-like-from-a-branch);<br>[Automatically deploy to review using a pull request](#automatically-deploy-to-review-using-a-pull-request);<br>[Automatically deploy to review from a branch using a pattern](#automatically-deploy-to-review-from-a-branch-using-a-pattern);<br>[Automatically deploy to review using a pull request; manual triggering (semi-automatic)](#automatically-deploy-to-review-using-a-pull-request-manual-triggering); | [Automatically deploy to production from master](#automatically-deploy-to-production-from-master);<br>[Automatically deploy to production using a tag](#automatically-deploy-to-production-using-a-tag);<br>[Automatically deploy to production from a branch](#automatically-deploy-to-production-from-a-branch);<br>[Automatically deploy to staging from master](#automatically-deploy-to-staging-from-master);<br>[Automatically deploy to production-like from a branch](#automatically-deploy-to-production-like-from-a-branch);<br>[Automatically deploy to review using a pull request](#automatically-deploy-to-review-using-a-pull-request);<br>[Automatically deploy to review from a branch using a pattern](#automatically-deploy-to-review-from-a-branch-using-a-pattern);<br>[Automatically deploy to review using a pull request; manual triggering (semi-automatic)](#automatically-deploy-to-review-using-a-pull-request-manual-triggering); |
| **Deploy manually** | [Automatically deploy to review using a pull request; manual triggering (semi-automatic)](#automatically-deploy-to-review-using-a-pull-request-manual-triggering); | [Deploy to production from master at the click of a button](#deploy-to-production-from-master-at-the-click-of-a-button);<br>[Deploy to production using a tag at the click of a button](#deploy-to-production-using-a-tag-at-the-click-of-a-button);<br>[Deploy to production from a branch at the click of a button](#deploy-to-production-from-a-branch-at-the-click-of-a-button);<br>[Deploy to production-like using a pull request at the click of a button](#deploy-to-production-like-using-a-pull-request-at-the-click-of-a-button);<br>[Deploy to review using a pull request at the click of a button](#deploy-to-review-using-a-pull-request-at-the-click-of-a-button);<br>[Automatically deploy to review using a pull request; manual triggering (semi-automatic)](#automatically-deploy-to-review-using-a-pull-request-manual-triggering); |

### Does the building block meet the criteria for CI/CD?

|   | **True CI/CD** | **Recommended for werf** |
|:---|:---:|:---:|
| [**Automatically deploy to production from master**](#automatically-deploy-to-production-from-master) | yes | yes |
| [**Deploy to production from master at the click of a button**](#deploy-to-production-from-master-at-the-click-of-a-button) | no | no |
| [**Automatically deploy to production using a tag**](#automatically-deploy-to-production-using-a-tag) | no | yes |
| [**Deploy to production using a tag at the click of a button**](#deploy-to-production-using-a-tag-at-the-click-of-a-button) | no | yes |
| [**Automatically deploy to production from a branch**](#automatically-deploy-to-production-from-a-branch) | yes | no |
| [**Deploy to production from a branch at the click of a button**](#deploy-to-production-from-a-branch-at-the-click-of-a-button) | no | no |
| [**Deploy to production-like using a pull request at the click of a button**](#deploy-to-production-like-using-a-pull-request-at-the-click-of-a-button) | no | yes |
| [**Automatically deploy to staging from master**](#automatically-deploy-to-staging-from-master) | yes | yes |
| [**Deploy to staging from master at the click of a button**](#deploy-to-staging-from-master-at-the-click-of-a-button) | no | no |
| [**Automatically deploy to production-like from a branch**](#automatically-deploy-to-production-like-from-a-branch) | yes | no |
| [**Deploy to production-like from a branch at the click of a button**](#deploy-to-production-like-from-a-branch-at-the-click-of-a-button) | no | no |
| [**Automatically deploy to review using a pull request**](#automatically-deploy-to-review-using-a-pull-request) | yes | no |
| [**Automatically deploy to review from a branch using a pattern**](#automatically-deploy-to-review-from-a-branch-using-a-pattern) | yes | no |
| [**Deploy to review using a pull request at the click of a button**](#deploy-to-review-using-a-pull-request-at-the-click-of-a-button) | no | no |
| [**Automatically deploy to review using a pull request; manual triggering**](#automatically-deploy-to-review-using-a-pull-request-manual-triggering) | yes | yes |

## Ready-made workflow configurations

In this section, we offer the user the selection of ready-made workflow configurations tailored to various use cases. These configurations are made up of the workflow blocks listed above. In the documentation, these ready-made configurations can also be referred to as workflow strategies.

The exact configuration for each type of the CI/CD system you can find in the documentation for that system. For example, GitLab CI/CD: link, GitHub Actions: link.

### №1 Fast and Furious

We recommend this configuration as the most consistent with the principles of CI/CD of those that can be implemented using werf.

This configuration supports any number of production-like environments, such as testing, staging, development, qa, and so on.

| **Environment** | **The workflow block** |
| :--- | :--- |
| Production | [Automatically deploy to production from master](#automatically-deploy-to-production-from-master) + rollback changes via revert |
| Staging / Testing / Development / QA | [Deploy to production-like using a pull request at the click of a button](#deploy-to-production-like-using-a-pull-request-at-the-click-of-a-button) |
| Review | [Automatically deploy to review using a pull request; manual triggering](#automatically-deploy-to-review-using-a-pull-request-manual-triggering) |

### №2 Push the button

| **Environment** | **The workflow block** |
| :--- | :--- |
| Production | [Deploy to production from master at the click of a button](#deploy-to-production-from-master-at-the-click-of-a-button) |
| Staging | [Automatically deploy to staging from master](#automatically-deploy-to-staging-from-master) |
| Testing / Development / QA | [Automatically deploy to production-like from a branch](#automatically-deploy-to-production-like-from-a-branch) |
| Review | [Deploy to review using a pull request at the click of a button](#deploy-to-review-using-a-pull-request-at-the-click-of-a-button) |

### №3 Tag everything

Not all projects are ready to implement the CI/CD approach. These projects use the more classical approach, that implies making releases after the active development phase is complete. The transition to the CI/CD approach in such projects requires both developers and DevOps engineers to change personal habits and rethink the workflow. Therefore, for such projects, we also offer a classic configuration. We recommend to use it with werf if [fast & furious](#1-fast-and-furious) does not suit you.

| **Environment** | **The workflow block** |
| :--- | :--- |
| Production | [Automatically deploy to production using a tag](#automatically-deploy-to-production-using-a-tag) |
| Staging | [Automatically deploy to staging from master or Deploy to staging from master at the click of a button](#automatically-deploy-to-staging-from-master) |
| Review | [Automatically deploy to review using a pull request; manual triggering](#automatically-deploy-to-review-using-a-pull-request-manual-triggering) |

### №4 Branch, branch, branch

Directly manage the deployment process via git by using branches and git merge, rebase, push-force procedures. By creating branches with specific names, you can trigger an automatic deployment to review environments.

We recommend this configuration for those who would like to manage the CI/CD process via git only. Note that this approach is also consistent with the CI/CD principles (just like the fast & furious approach described above).

| **Environment** | **The workflow block** |
| :--- | :--- |
| Production | [Automatically deploy to production from master](#automatically-deploy-to-production-from-master) + rollback changes via revert |
| Staging | [Automatically deploy to staging from master](#automatically-deploy-to-staging-from-master) |
| Review | [Automatically deploy to review from a branch using a pattern](#automatically-deploy-to-review-from-a-branch-using-a-pattern) |
