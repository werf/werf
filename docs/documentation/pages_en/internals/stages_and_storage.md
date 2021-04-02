---
title: Stages and storage
permalink: internals/stages_and_storage.html
---

We divide the build process of the images described in [the werf configuration file]({{ "reference/werf_yaml.html" | true_relative_url }}) into stages, with [clear functions and purposes](#stage-dependencies). Each stage corresponds to an intermediate image, like layers in Docker. The final image corresponds to the last built stage for a particular git state and the werf configuration file.

Stages are steps in the build process. A ***stage*** is built from a logically grouped set of config instructions. It takes into account the assembly conditions and rules. Each _stage_ relates to a single Docker image and saved in the [storage](#storage).

You can consider the stages as a building cache of an application, however, that are not the cache but a part of a building context.

## Stage conveyor

A _stage conveyor_ is an ordered sequence of conditions and rules for carrying out stages. werf uses different _stage conveyors_ to assemble various types of images depending on their configuration.

<div class="tabs">
  <a href="javascript:void(0)" class="tabs__btn active" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'dockerfile-image-tab')">Dockerfile Image</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'stapel-image-tab')">Stapel Image</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'stapel-artifact-tab')">Stapel Artifact</a>
</div>

<div id="dockerfile-image-tab" class="tabs__content active">
<a class="google-drawings" href="{{ "images/reference/stages_and_images1.png" | true_relative_url }}" data-featherlight="image">
<img src="{{ "images/reference/stages_and_images1_preview.png" | true_relative_url }}">
</a>
</div>

<div id="stapel-image-tab" class="tabs__content">
<a class="google-drawings" href="{{ "images/reference/stages_and_images2.png" | true_relative_url }}" data-featherlight="image">
<img src="{{ "images/reference/stages_and_images2_preview.png" | true_relative_url }}">
</a>
</div>

<div id="stapel-artifact-tab" class="tabs__content">
<a class="google-drawings" href="{{ "images/reference/stages_and_images3.png" | true_relative_url }}" data-featherlight="image">
<img src="{{ "images/reference/stages_and_images3_preview.png" | true_relative_url }}">
</a>
</div>

**The user only needs to write a correct configuration: werf performs the rest of the work with stages**

For each _stage_ at every build, werf calculates the unique identifier of the stage called [stage digest](#stage-digest).

If a _stage_ has no [stage dependencies](#stage-dependencies), it is skipped, and, accordingly, the _stage conveyor_ is reduced by one stage. It means that the _stage conveyor_ can be reduced to several _stages_ or even to a single _from_ stage.

<a class="google-drawings" href="{{ "images/reference/stages_and_images4.png" | true_relative_url }}" data-featherlight="image">
<img src="{{ "images/reference/stages_and_images4_preview.png" | true_relative_url }}">
</a>

## Stage digest

The _stage digest_ is used for [tagging](#stage-naming) a _stage_ (digest is the part of image tag) in the _storage_.
werf does not build stages that already exist in the _storage_ (similar to caching in Docker yet more complex).

The ***stage digest*** is calculated as the checksum of:
 - checksum of [stage dependencies]({{ "internals/stages_and_storage.html#stage-dependencies" | true_relative_url }});
 - previous _stage digest_;
 - git commit-id related with the previous stage (if previous stage is git-related).

Digest identifier of the stage represents content of the stage and depends on git history which lead to this content.

## Stage dependencies

_Stage dependency_ is a piece of data that affects the stage _digest_. Stage dependency may be represented by:

 - some file from a git repo with its contents;
 - instructions to build stage defined in the `werf.yaml`;
 - the arbitrary string specified by the user in the `werf.yaml`;
 - and so on.

Most _stage dependencies_ are specified in the `werf.yaml`, others relate to a runtime.

The tables below illustrate dependencies of a Dockerfile image, a Stapel image, and a [Stapel artifact]({{ "advanced/building_images_with_stapel/artifacts.html" | true_relative_url }}) _stages dependencies_.
Each row describes dependencies for a certain stage.
The left column contains a short description of dependencies, the right column includes related `werf.yaml` directives and contains relevant references for more information.

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
        <li><a href="{{ reference.link | true_relative_url }}">{{ reference.name }}</a></li>
    {% endfor %}
    </ul>
</div>
{% endif %}

</div>

    </div>
</div>
{% endfor %}
</div>

<link rel="stylesheet" type="text/css" href="{{ assets["stages.css"].digest_path | true_relative_url }}" />

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

## Storage

_Storage_ contains the stages of the project. Stages can be stored in the container registry or locally on a host machine.

Most werf commands use _stages_. Such commands require specifying the location of the _storage_ using the `--repo` key or the `WERF_REPO` environment variable.

There are 2 types of storage:
 1. _Local storage_. Uses local docker server runtime to store stages as docker images. 
 2. _Remote storage_. Uses container registry to store images. Remote storage is selected by param `--repo=CONTAINER_REGISTRY_REPO`, for example `--repo=registry.mycompany.com/web/frontend/stages`. **NOTE** Each project should specify unique docker repo domain, that used only by this project.

Stages are [named differently](#stage-naming) depending on local or remote storage used.

When container registry is used as the storage for the project there is also a cache of local docker images on each host where werf is running. This cache is cleared by the werf itself or can be freely removed by other tools (such as `docker rmi`).

### Stage naming

Stages in the _local storage_ are named using the following schema: `PROJECT_NAME:STAGE_DIGEST-TIMESTAMP_MILLISEC`. For example:

```
myproject                   9f3a82975136d66d04ebcb9ce90b14428077099417b6c170e2ef2fef-1589786063772   274bd7e41dd9        16 seconds ago      65.4MB
myproject                   7a29ff1ba40e2f601d1f9ead88214d4429835c43a0efd440e052e068-1589786061907   e455d998a06e        18 seconds ago      65.4MB
myproject                   878f70c2034f41558e2e13f9d4e7d3c6127cdbee515812a44fef61b6-1589786056879   771f2c139561        23 seconds ago      65.4MB
myproject                   5e4cb0dcd255ac2963ec0905df3c8c8a9be64bbdfa57467aabeaeb91-1589786050923   699770c600e6        29 seconds ago      65.4MB
myproject                   14df0fe44a98f492b7b085055f6bc82ffc7a4fb55cd97d30331f0a93-1589786048987   54d5e60e052e        31 seconds ago      64.2MB
```

Stages in the _remote storage_ are named using the following schema: `CONTAINER_REGISTRY_REPO:STAGE_DIGEST-TIMESTAMP_MILLISEC`. For example:

```
localhost:5000/myproject-stages                 d4bf3e71015d1e757a8481536eeabda98f51f1891d68b539cc50753a-1589714365467   7c834f0ff026        20 hours ago        66.7MB
localhost:5000/myproject-stages                 e6073b8f03231e122fa3b7d3294ff69a5060c332c4395e7d0b3231e3-1589714362300   2fc39536332d        20 hours ago        66.7MB
localhost:5000/myproject-stages                 20dcf519ff499da126ada17dbc1e09f98dd1d9aecb85a7fd917ccc96-1589714359522   f9815cec0867        20 hours ago        65.4MB
localhost:5000/myproject-stages                 1dbdae9cc1c9d5d8d3721e32be5ed5542199def38ff6e28270581cdc-1589714352200   6a37070d1b46        20 hours ago        65.4MB
localhost:5000/myproject-stages                 f88cb5a1c353a8aed65d7ad797859b39d357b49a802a671d881bd3b6-1589714347985   5295f82d8796        20 hours ago        65.4MB
localhost:5000/myproject-stages                 796e905d0cc975e718b3f8b3ea0199ea4d52668ecc12c4dbf85a136d-1589714344546   a02ec3540da5        20 hours ago        64.2MB
```

_Digest_ identifier of the stage represents content of the stage and depends on git history which lead to this content.

- `PROJECT_NAME` — the project name.
- `CONTAINER_REGISTRY_REPO` — the specified container registry repository with the `--repo` option.
- `STAGE_DIGEST` — [the stage digest][#stage-digest].
- `TIMESTAMP_MILLISEC` — the timestamp that is generated during [stage saving procedure]({{ "internals/build_process.html#saving-stages-to-the-storage" | true_relative_url }}) after stage built. It is guaranteed that timestamp will be unique within specified storage.
