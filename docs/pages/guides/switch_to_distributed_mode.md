---
title: Switch to distributed mode
sidebar: documentation
permalink: guides/switch_to_distributed_mode.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

Let's say you have some project in Gitlab CI/CD, that uses werf and [`:local` stages storage]({{ site.baseurl }}/reference/stages_and_images.html#stages-storage). In this article we provide steps to switch such project to remote stages storage. Remote stages storage allows running werf from multiple hosts in a distributed manner.

Required werf version: >= v1.1.10.

Hosts requirements:
 - Connection to the docker registry which will be used as a stages storage.
 - Connection to the Kubernetes cluster.

**Important.** It is required to run switch-from-local command to prevent future stages storage integrity violation due to usage of local and remote stages storage at the same time. Please do not skip this step.

**Important.** There will be outage of CI/CD operation for git branches that not merged/rebased onto changes introduced by MR which changes to remote stages storage. This outage will be activated by switch-from-local command. After this command was called all active git branches that use CI/CD should start using remote stages storage (because local stages storage will refuse to work anymore).

## Steps to switch

 1. _Optional._ Copy already existing stages from gitlab runner to the registry. This step is 100% safe and will not break CI/CD operation for the project. This step is recommended because it will speed up further switch-from-local command call and it will prevent unnecessary rebuild of stages which already exist in the local stages storage when preparing MR in one of the further steps.
   - Go to the project directory on gitlab runner.
   - Get docker repo address of the project using `Gitlab CI/CD / Packages / Container Registry` interface.
   - Copy address of the project docker repo and add `/stages` suffix. For example: `registry.mycompany.com/web/catalog/myproject/stages`. **NOTE:** Werf ci-env command will automatically use `CI_REGISTRY_IMAGE/stages` as a stages storage, that's why it is required to use `/stages` suffix and not an arbitrary one.
   - Login into docker registry by running `docker login registry.mycompany.com/web/catalog/myproject/stages -uUSER -pTOKEN`. You can get a `TOKEN` for your user using Gitlab CI/CD interface.
   - Run stages sync command: `werf stages sync --from :local --to registry.mycompany.com/web/catalog/myproject/stages`.
 2. _Required._ Prepare MR which changes `.gitlab-ci.yml`: remove explicit `--stages-storage :local` param (or environment variable `WERF_STAGES_STORAGE`) in all werf commands. Wait until Gitlab CI/CD build stage for this MR is done.
 3. _Required._ Run switch command from local to registry: `werf stages switch-from-local --to registry.mycompany.com/web/catalog/myproject/stages`. **NOTE:** After this command has done werf will refuse to work with `:local` stages storage anymore, so all git branches that use `:local` stages storage should merge changes from step 2 to enable normal CI/CD operation if needed.
 4. Merge MR from step 2.
 5. Actualize all git branches with active development: merge/rebase onto new changes introduced in MR from step 4.

**NOTE:** If you need to switch the whole gitlab to remote stages storage, then it is possible to write a script, that makes a sync (from step 1) for all projects of gitlab, because it is long-running operation. While this script is working you can prepare MR for all projects and then perform switch-from-local and merge MR steps for each project sequentially.

## How to rollback to local

 1. If switch-from-local command ([step 3](#steps-to-switch)) was successful, then there are no more local stages, because switch command has removed local stages. To restore this cache run sync command: `werf stages sync --from registry.mycompany.com/web/catalog/myproject/stages --to :local`.
 2. Remove block-file which prevents usage of `:local` stages storage for your project: `rm ~/.werf/service/stages_switch_from_local_block/PROJECT_NAME` — this block-file was created in the [step 3](#steps-to-switch).
 3. Create MR to revert changes introduced in [step 2](#steps-to-switch).
 4. Stop all project jobs that run werf process and prevent running new ones.
 5. Merge new MR and bring these new changes into all active git branches that need CI/CD operation and already use remote stages storage.
