---
title: Comparison with other solutions
permalink: resources/comparison.html
---

## Giterminism vs GitOps

GitOps only defines the deployment of pre-built application artifacts, while giterminism can define (shape) the entire CI/CD process, including building, testing, distributing, and deploying.

GitOps requires a CD solution that continuously synchronizes the desired state with the actual one, while giterminism imposes no restrictions: the user can decide how to perform such synchronization themselves.

GitOps mandates separation of development and operation, while giterminism allows for them to be separated or integrated into a single process so that the DevOps methodology can be followed.

## werf vs Helm

Helm is used only for chart deployment and distribution, while werf can also be used for developing, building, testing, distributing images and bundles, and cleaning up the container registry. 

Helm is built into werf and is enhanced with additional features such as advanced tracking, customizing the deployment order not only for hooks but also for regular resources, and more.

## werf vs Argo CD

Argo CD is only used for deploying, while werf also supports developing, building, testing, distributing, and cleaning up the container registry. 

Deployment in werf follows a Helm push model, though integration with Argo CD to implement GitOps is also available. Read [this article](https://blog.werf.io/new-werf-mode-combining-werf-argo-cd-into-a-unified-ci-cd-process-d49bce7f3be1) to learn more about the integration and how & why Argo CD might complement werf.

## werf vs Skaffold/DevSpace

Skaffold and DevSpace are essentially wrappers for the popular builders and deployment tools with additional development-oriented functionality. 

werf, on the other hand, focuses on CI/CD and tighter integration of the unified way of building and deploying. Users are provided with the solutions to their application problems, not just tools (those are secondary in the werf's context).
