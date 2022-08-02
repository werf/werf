---
title: Synchronization in werf
permalink: advanced/synchronization.html
---

Synchronization is a group of service components of the werf to coordinate multiple werf processes when selecting and publishing stages into storage. There are 2 such synchronization components:

 1. _Stages storage cache_ is an internal werf cache, which significantly improves performance of the werf invocations when stages already exists in the storage. Stages storage cache contains the mapping of stages existing in storage by the digest (or in other words this cache contains precalculated result of stages selection by digest algorithm). This cache should be coherent with storage itself and werf will automatically reset this cache automatically when detects an inconsistency between storage cache and storage.
 2. _Lock manager_. Locks are needed to organize correct publishing of new stages into storage, publishing images into images-repo and for concurrent deploy processes that uses the same release name.

All commands that requires storage (`--repo`) param also use _synchronization service components_ address, which defined by the `--synchronization` option or `WERF_SYNCHRONIZATION=...` environment variable.

There are 3 types of synchronization components:
 1. Local. Selected by `--synchronization=:local` param.
   - Local _storage cache_ is stored in the `~/.werf/shared_context/storage/stages_storage_cache/1/PROJECT_NAME/DIGEST` files by default, each file contains a mapping of images existing in storage by some digest.
   - Local _lock manager_ uses OS file-locks in the `~/.werf/service/locks` as implementation of locks.
 2. Kubernetes. Selected by `--synchronization=kubernetes://NAMESPACE[:CONTEXT][@(base64:CONFIG_DATA)|CONFIG_PATH]` param.
  - Kubernetes _storage cache_ is stored in the specified `NAMESPACE` in ConfigMap named by project `cm/PROJECT_NAME`.
  - Kubernetes _lock manager_  uses ConfigMap named by project `cm/PROJECT_NAME` (the same as storage cache) to store distributed locks in the annotations. [Lockgate library](https://github.com/werf/lockgate) is used as implementation of distributed locks using kubernetes resource annotations.
 3. Http. Selected by `--synchronization=http[s]://DOMAIN` param.
  - There is a public instance of synchronization server available at domain `https://synchronization.werf.io`.
  - Custom http synchronization server can be run with `werf synchronization` command.

werf uses `--synchronization=:local` (local _storage cache_ and local _lock manager_) by default when _local storage_ is used.

werf uses `--synchronization=https://synchronization.werf.io` (http _storage cache_ and http _lock manager_) by default when container registry is used as _storage_.

User may force arbitrary non-default address of synchronization service components if needed using explicit `--synchronization=:local|(kubernetes://NAMESPACE[:CONTEXT][@(base64:CONFIG_DATA)|CONFIG_PATH])|(http[s]://DOMAIN)` param.

**NOTE:** Multiple werf processes working with the same project should use the same _storage_ and _synchronization_.
