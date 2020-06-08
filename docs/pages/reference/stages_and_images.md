---
title: Stages and Images
sidebar: documentation
permalink: documentation/reference/stages_and_images.html
author: Alexey Igrychev, Timofey Kirillov <alexey.igrychev@flant.com,timofey.kirillov@flant.com>
---

We propose to divide the assembly process into steps. Every step corresponds to the intermediate image (like layers in Docker) with specific functions and assignments.
In werf, we call every such step a [stage](#stages). So the final [image](#images) consists of a set of built stages.
All stages are kept in a [stages storage](#stages-storage). You can view it as a building cache of an application, however, that isn't a cache but merely a part of a building context.

## Stages

Stages are steps in the assembly process. They act as building blocks for constructing images.
A ***stage*** is built from a logically grouped set of config instructions. It takes into account the assembly conditions and rules.
Each _stage_ relates to a single Docker image.

The werf assembly process involves a sequential build of stages using the _stage conveyor_.  A _stage conveyor_ is an ordered sequence of conditions and rules for carrying out stages. werf uses different _stage conveyors_ to assemble various types of images depending on their configuration.

<div class="tabs">
  <a href="javascript:void(0)" class="tabs__btn active" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'dockerfile-image-tab')">Dockerfile Image</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'stapel-image-tab')">Stapel Image</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'stapel-artifact-tab')">Stapel Artifact</a>
</div>

<div id="dockerfile-image-tab" class="tabs__content active">
<a class="google-drawings" href="../../images/reference/stages_and_images1.png" data-featherlight="image">
<img src="../../images/reference/stages_and_images1_preview.png">
</a>
</div>

<div id="stapel-image-tab" class="tabs__content">
<a class="google-drawings" href="../../images/reference/stages_and_images2.png" data-featherlight="image">
<img src="../../images/reference/stages_and_images2_preview.png" >
</a>
</div>

<div id="stapel-artifact-tab" class="tabs__content">
<a class="google-drawings" href="../../images/reference/stages_and_images3.png" data-featherlight="image">
<img src="../../images/reference/stages_and_images3_preview.png">
</a>
</div>

**The user only needs to write a correct configuration: werf performs the rest of the work with stages**

For each _stage_ at every build, werf calculates the unique identifier of the stage called _stage signature_.
Each _stage_ is assembled in the ***assembly container*** that is based on the previous _stage_ and saved in the [stages storage](#stages-storage).
The _stage signature_ is used for [tagging](#stage-naming) a _stage_ (signature is the part of image tag) in the _stages storage_.
werf does not build stages that already exist in the _stages storage_ (similar to caching in Docker yet more complex).

The ***stage signature*** is calculated as the checksum of:
 - checksum of [stage dependencies]({{ site.baseurl }}/documentation/reference/stages_and_images.html#stage-dependencies);
 - previous _stage signature_;
 - git commit-id related with the previous stage (if previous stage is git-related).

Signature identifier of the stage represents content of the stage and depends on git history which lead to this content. There may be multiple built images for a single signature. Stage for different git branches can have the same signature, but werf will prevent cache of different git branches from
being reused for totally different branches, [see stage selection algorithm]({{ site.baseurl }}/documentation/reference/stages_and_images.html#stage-selection).

It means that the _stage conveyor_ can be reduced to several _stages_ or even to a single _from_ stage.

<a class="google-drawings" href="../../images/reference/stages_and_images4.png" data-featherlight="image">
<img src="../../images/reference/stages_and_images4_preview.png">
</a>

## Stage dependencies

_Stage dependency_ is a piece of data that affects the stage _signature_. Stage dependency may be represented by:

 - some file from a git repo with its contents;
 - instructions to build stage defined in the `werf.yaml`;
 - the arbitrary string specified by the user in the `werf.yaml`;
 - and so on.

Most _stage dependencies_ are specified in the `werf.yaml`, others relate to a runtime.

The tables below illustrate dependencies of a Dockerfile image, a Stapel image, and a [Stapel artifact]({{ site.baseurl }}/documentation/configuration/stapel_artifact.html) _stages dependencies_.
Each row describes dependencies for a certain stage.
Left column contains a short description of dependencies, right column includes related `werf.yaml` directives and contains relevant references for more information.

<div class="tabs">
  <a href="javascript:void(0)" id="image-from-dockerfile-dependencies" class="tabs__btn dependencies-btn active">Dockerfile Image</a>
  <a href="javascript:void(0)" id="image-dependencies" class="tabs__btn dependencies-btn">Stapel Image</a>
  <a href="javascript:void(0)" id="artifact-dependencies" class="tabs__btn dependencies-btn">Stapel Artifact</a>
</div>

<div id="dependencies">
{% for stage in site.data.stages.en.entries %}
<div class="stage {{stage.type}}">
  <div class="stage-body">
    <div class="stage-base">
      <p>stage {{ stage.name | escape }}</p>

      {% if stage.dependencies %}
      <div class="dependencies">
        {% for dependency in stage.dependencies %}
        <div class="dependency">
          {{ dependency | escape }}
        </div>
        {% endfor %}
      </div>
      {% endif %}
    </div>

<div class="werf-config" markdown="1">

{% if stage.werf_config %}
```yaml
{{ stage.werf_config }}
```
{% endif %}

{% if stage.references %}
<div class="references">
    References:
    <ul>
    {% for reference in stage.references %}
        <li><a href="{{ reference.link }}">{{ reference.name }}</a></li>
    {% endfor %}
    </ul>
</div>
{% endif %}

</div>

    </div>
</div>
{% endfor %}
</div>

{% asset stages.css %}
<script src="https://cdnjs.cloudflare.com/ajax/libs/jquery/3.4.1/jquery.min.js"></script>
<script>
function application() {
  if ($("a[id=image-from-dockerfile-dependencies]").hasClass('active')) {
    $(".image").addClass('hidden');
    $(".artifact").addClass('hidden');
    $(".image-from-dockerfile").removeClass('hidden')
  }
  else if ($("a[id=image-dependencies]").hasClass('active')) {
    $(".image-from-dockerfile").addClass('hidden');
    $(".artifact").addClass('hidden');
    $(".image").removeClass('hidden')
  }
  else if ($("a[id=artifact-dependencies]").hasClass('active')) {
    $(".image-from-dockerfile").addClass('hidden');
    $(".image").addClass('hidden');
    $(".artifact").removeClass('hidden')
  }
  else {
    $(".image-from-dockerfile").addClass('hidden');
    $(".image").addClass('hidden');
    $(".artifact").addClass('hidden')
  }
}

$('.tabs').on('click', '.dependencies-btn', function() {
  $(this).toggleClass('active').siblings().removeClass('active');
  application()
});

application();
$.noConflict();
</script>

## Stages storage

_Stages storage_ contains the stages of the project. Stages can be stored in the Docker Repo or locally on a host machine.

Most commands use _stages_ and require the reference to a specific _stages storage_ defined by the `--stages-storage` option or `WERF_STAGES_STORAGE` environment variable.

There are 2 types of stages storage:
 1. _Local stages storage_. Uses local docker server runtime to store stages as docker-images. Local stages storage is selected by param `--stages-storage=:local`. This was the only supported choise for stages storage prior version v1.1.10.
 2. _Remote stages storage_. Uses docker registry to store images. Remote stages storage is selected by param `--stages-storage=DOCKER_REPO_DOMAIN`, for example `--stages-storage=registry.mycompany.com/web/frontend/stages`. **NOTE** Each project should specify unique docker repo domain, that used only by this project.

Stages will be [named differently](#stage-naming) depending on local or remote stages storage is being used.

When docker registry is used as the stages storage for the project there is also a cache of local docker images on each host where werf is running. This cache is cleared by the werf itself or can be freely removed by other tools (such as `docker rmi`).

It is recommended though to use docker registry as a stages storage, werf uses this mode with [CI/CD systems by default]({{ site.baseurl }}/documentation/reference/plugging_into_cicd/overview.html).

Host requirements to use remote stages storage:
 - Connection to docker registry.
 - Connection to the Kubernetes cluster (used to synchronize multiple build/publish/deploy processes running from different machines, see more info below).

Note that all werf commands that need an access to the stages should specify the same stages storage. So if it is a local stages storage, then all commands should run from the same host. It is irrelevant on which host werf command is running as long as the same remote stages storage used for the commands like: build, publish, cleanup, deploy, etc.

### Stage naming

Stages in the _local stages storage_ are named using the following schema: `werf-stages-storage/PROJECT_NAME:SIGNATURE-TIMESTAMP_MILLISEC`. For example:

```
werf-stages-storage/myproject                   9f3a82975136d66d04ebcb9ce90b14428077099417b6c170e2ef2fef-1589786063772   274bd7e41dd9        16 seconds ago      65.4MB
werf-stages-storage/myproject                   7a29ff1ba40e2f601d1f9ead88214d4429835c43a0efd440e052e068-1589786061907   e455d998a06e        18 seconds ago      65.4MB
werf-stages-storage/myproject                   878f70c2034f41558e2e13f9d4e7d3c6127cdbee515812a44fef61b6-1589786056879   771f2c139561        23 seconds ago      65.4MB
werf-stages-storage/myproject                   5e4cb0dcd255ac2963ec0905df3c8c8a9be64bbdfa57467aabeaeb91-1589786050923   699770c600e6        29 seconds ago      65.4MB
werf-stages-storage/myproject                   14df0fe44a98f492b7b085055f6bc82ffc7a4fb55cd97d30331f0a93-1589786048987   54d5e60e052e        31 seconds ago      64.2MB
```

Stages in the _remote stages storage_ are named using the following schema: `DOCKER_REPO_ADDRESS:SIGNATURE-TIMESTAMP_MILLISEC`. For example:

```
localhost:5000/myproject-stages                 d4bf3e71015d1e757a8481536eeabda98f51f1891d68b539cc50753a-1589714365467   7c834f0ff026        20 hours ago        66.7MB
localhost:5000/myproject-stages                 e6073b8f03231e122fa3b7d3294ff69a5060c332c4395e7d0b3231e3-1589714362300   2fc39536332d        20 hours ago        66.7MB
localhost:5000/myproject-stages                 20dcf519ff499da126ada17dbc1e09f98dd1d9aecb85a7fd917ccc96-1589714359522   f9815cec0867        20 hours ago        65.4MB
localhost:5000/myproject-stages                 1dbdae9cc1c9d5d8d3721e32be5ed5542199def38ff6e28270581cdc-1589714352200   6a37070d1b46        20 hours ago        65.4MB
localhost:5000/myproject-stages                 f88cb5a1c353a8aed65d7ad797859b39d357b49a802a671d881bd3b6-1589714347985   5295f82d8796        20 hours ago        65.4MB
localhost:5000/myproject-stages                 796e905d0cc975e718b3f8b3ea0199ea4d52668ecc12c4dbf85a136d-1589714344546   a02ec3540da5        20 hours ago        64.2MB
```

_Signature_ identifier of the stage represents content of the stage and depends on git history which lead to this content.

`TIMESTAMP_MILLISEC` is generated during [stage saving procedure](#stage-building-and-saving) after stage built. It is guaranteed that timestamp will be unique within specified stages storage.

### Stage selection

Werf stage selection algorithm is based on the git commits ancestry detection:

 1. Calculate a stage signature for some stage.
 2. There may be multiple stages in the stages storage by this signature — so select all suitable stages by the signature.
 3. If current stage is related to git (git-archive, user stage with git patch, git cache or git latest patch), then select only those stages which are related to the commit that is ancestor of current git commit.
 4. Select the _oldest_ by the `TIMESTAMP_MILLISEC` from the remaining stages.

There may be multiple built images for a single signature. Stage for different git branches can have the same signature, but werf will prevent cache of different git branches from being reused for different branch.

### Stage building and saving

If suitable stage has not been found by target signature during stage selection, werf starts building a new image for stage.

Note that multiple processes (on a single or multiple hosts) may start building the same stage at the same time. Werf uses optimistic locking when saving newly built image into the stages storage: when a new stage has been built werf locks stages storage and saves newly built stage image into storage stages cache only if there are no suitable already existing stages exists. Newly saved image will have a guaranteed unique identifier `TIMESTAMP_MILLISEC`. In the case when already existing stage has been found in the stages storage werf will discard newly built image and use already existing one as a cache.

In other words: the first process which finishes the build (the fastest one) will have a chance to save newly built stage into the stages storage. The slow build process will not block faster processes from saving build results and building next stages.

To select stages and save new ones into the stages storage werf uses [synchronization service components](#synchronization-locks-and-stages-storage-cache) to coordinate multiple werf processes and store stages cache needed for werf builder.

### Image stages signature

_Stages signature_ of the image is a signature which represents content of the image and depends on the history of git commits which lead to this content.

***Stages signature*** calculated similarly to the regular stage signature as the checksum of:
 - _stage signature_ of last non empty image stage;
 - git commit-id related with the last non empty image stage (if this last stage is git-related).

The ***stage signature*** is calculated as the checksum of:
 - checksum of [stage dependencies]({{ site.baseurl }}/documentation/reference/stages_and_images.html#stage-dependencies);
 - previous _stage signature_;
 - git commit-id related with the previous stage (if previous stage is git-related).

This signature used in [content based tagging]({{ site.baseurl }}/documentation/reference/publish_process.html#content-based-tagging) and used to import files from artifacts or images (stages signature of artifact or image will affect imports stage signature of the target image).

## Images

_Image_ is a **ready-to-use** Docker image corresponding to a specific application state and [tagging strategy]({{ site.baseurl }}/documentation/reference/publish_process.html).

As mentioned [above](#stages), _stages_ are steps in the assembly process. They act as building blocks for constructing _images_.
Unlike images, _stages_ are not intended for the direct use. The main difference between images and stages is in [cleaning policies]({{ site.baseurl }}/documentation/reference/cleaning_process.html#cleanup-policies) due to the stored meta-information.
The process of cleaning up the _stages storage_ is only based on the related images in the _images repo_.

werf creates _images_ using the _stages storage_.
Currently, _images_ can only be created during the [_publishing process_]({{ site.baseurl }}/documentation/reference/publish_process.html) and saved in the _images repo_.

Images should be defined in the werf configuration file `werf.yaml`.

To publish new images into the images repo werf uses [synchronization service components](#synchronization-locks-and-stages-storage-cache) to coordinate multiple werf processes. Only a single werf process can perform publishing of the same image at a time.

## Synchronization: locks and stages storage cache

Synchronization is a group of service components of the werf to coordinate multiple werf processes when selecting and saving stages into stages storage and publishing images into images repo. There are 2 such synchronization components:

 1. _Stages storage cache_ is an internal werf cache, which significantly improves performance of the werf invocations when stages already exists in the stages storage. Stages storage cache contains the mapping of stages existing in stages storage by the signature (or in other words this cache contains precalculated result of stages selection by signature algorithm). This cache should be coherent with stages storage itself and werf will automatically reset this cache automatically when detects an inconsistency between stages storage cache and stages storage.
 2. _Lock manager_. Locks are needed to organize correct publishing of new stages into stages-storage, publishing images into images-repo and for concurrent deploy processes that uses the same release name.

All commands that requires stages storage (`--stages-storage`) and images repo (`--images-repo`) params also use _synchronization service components_ address, which defined by the `--synchronization` option or `WERF_SYNCHRONIZATION=...` environment variable.

There are 2 types of sycnhronization components:
 1. Local. Selected by `--synchronization=:local` param.
   - Local _stages storage cache_ is stored in the `~/.werf/shared_context/storage/stages_storage_cache/1/PROJECT_NAME/SIGNATURE` files by default, each file contains a mapping of images existing in stages storage by some signature.
   - Local _lock manager_ uses OS file-locks in the `~/.werf/service/locks` as implementation of locks.
 2. Kubernetes. Selected by `--synchronization=kubernetes://NAMESPACE` param.
  - Kubernetes _stages storage cache_ is stored in the specified `NAMESPACE` in ConfigMap named by project `cm/PROJECT_NAME`.
  - Kubernetes _lock manager_  uses ConfigMap named by project `cm/PROJECT_NAME` (the same as stages storage cache) to store distributed locks in the annotations. [Lockgate library](https://github.com/werf/lockgate) is used as implementation of distributed locks using kubernetes resource annotations.

Werf uses `--synchronization=:local` (local _stages storage cache_ and local _lock manager_) by default when _local stages storage_ is used (`--stages-storage=:local`).

Werf uses `--synchronization=kubernetes://werf-synchronization` (kubernetes _stages storage cache_ and kubernetes _lock manager_) by default when docker-registry is used as _stages storage_. Stages storage cache and locks for each project is stored in the `cm/PROJECT_NAME` in the common namespace `werf-synchronization`.

User may force arbitrary non-default address of synchronization service components if needed using explicit `--synchronization=:local|kubernetes://NAMESPACE` param (arbitrary namespace may be specified, `werf-synchronization` is the default one).

**NOTE:** Multiple werf processes working with the same project should use the same _stages storage_ and _syncrhonization_.

## Working with stages

### Sync command

`werf stages sync --from=:local|REPO --to=:local|REPO`

 - Command will copy only difference of stages from one stages-storage to another.
 - Command will copy multiple stages in parallel.
 - Command run result is idempotent: sync can be called multiple times, interrupted, then called again — the result will be the same. Stages that are already synced will not be synced again on subsequent sync calls.
 - There are delete options: `--remove-source` and `--cleanup-local-cache`, which control whether werf will delete synced stages from source stages-storage and whether werf will cleanup localhost from temporary docker images created during sync process.
 - This command can be used to download project stages-storage to the localhost for development purpose as well as backup and migrating purposes.
 
### Switch-from-local command

`werf stages switch-from-local --to=REPO`

 - Command will automatically [sync](#sync-command) existing stages from :local stages storage to the specified REPO.
 - Command will block project from being used with `:local` stages-storage.
   - This means after werf stages switch-from-local is done, any werf command that specifies `:local` stages-storage for the project will fail preventing storing and using build results from different stages-storages.
   - Note that project is blocked after all existing stages has been synced.

See [switching to distributed mode article]({{ site.baseurl }}/documentation/guides/switch_to_distributed_mode.html) for guided steps.

## Further reading

Learn more about the [build process of stapel and Dockerfile builders]({{ site.baseurl }}/documentation/reference/build_process.html).
