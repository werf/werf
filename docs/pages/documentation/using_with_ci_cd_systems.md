---
title: Using with CI/CD systems
permalink: documentation/using_with_ci_cd_systems.html
sidebar: documentation
---

Werf can be embedded into any CI/CD system. Each CI/CD job should perform following steps:

 1. Prepare cloned git work tree of your project, checkout target git commit.

 2. Login into docker repo of your project if needed.

```
docker login -u USER -p PASSWORD REPO
```

 3. Call `werf converge --repo REPO --env ENV` command to deploy application. If you use different CI/CD environments then determine current environment value to pass to the werf, if not — then omit `--env` param.
