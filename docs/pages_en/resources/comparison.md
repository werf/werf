---
title: Comparing to other tools
permalink: resources/comparison.html
---

## werf vs Helm

Helm is used only for chart deployment and distribution, while werf can also be used for developing, building, testing, distributing images and bundles, and cleaning up the container registry. Helm is built into werf and is enhanced with additional features such as advanced tracking, customizing the deployment order not only for hooks but also for regular resources, and more.

## werf vs ArgoCD

ArgoCD is only used for deploying (production-like environments), while werf also supports developing, building, testing, distributing, and cleaning up the container registry. Deployment in werf follows a Helm push model, though integration with ArgoCD to implement GitOps is also available.

## werf vs Scaffold/DevSpace

Skaffold and DevSpace are essentially wrappers for the popular builders and deployment tools with additional development-oriented functionality. werf, on the other hand, focuses on CI/CD and tighter integration of the unified way of building and deploying. Users are provided with the solutions to their application problems, not just tools (those are secondary in the werf's context).
