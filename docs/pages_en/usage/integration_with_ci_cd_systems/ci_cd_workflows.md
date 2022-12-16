---
title: CI/CD workflows
permalink: usage/integration_with_ci_cd_systems/ci_cd_workflows.html
---

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
