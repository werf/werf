---
title: How to migrate from v1.1 to v1.2
permalink: documentation/whats_new_in_v1_2/howto_from_v1_1_to_v1_2.html
description: How to migrate your application from v1.1 to v1.2
sidebar: documentation
---

**This article is under construction.**

 - Use `.Values.werf.image.IMAGE_NAME` instead of `werf_container_image`.
 - Completely remove `werf_container_env`.
 - Env-vars in the werf.yaml should be defined in the `werf-giterminism.yaml`.
 - V1.2 `werf converge` command will build and publish into container registry all needed images, try to render and validate project helm templates, and if successful — migrate existing helm 2 release to helm 3 release automatically, then perform usual helm upgrade procedure.
