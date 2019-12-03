---
title: Stages and Images
sidebar: documentation
permalink: documentation/reference/stages_and_images.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
---

We propose to divide the assembly proccess into steps, intermediate images (like layers in Docker), with clear functions and assignments.
In Werf, such step is called [stage](#stages) and result [image](#images) consists of a set of built stages.
All stages are kept in a [stages storage](#stages-storage) and defining build cache of application (not really cache but part of building context).

## Stages

Stages are steps in the assembly process, building blocks for constructing images.
A ***stage*** is built from a logically grouped set of config instructions, taking into account the assembly conditions and rules.
Each _stage_ relates to one Docker image.

The Werf assembly process assumes a sequential build of stages using _stage conveyor_.  A _stage conveyor_ is a sequence with the predefined order and set of stages. Werf uses different _stage conveyor_ for assembling a particular type of build object.

<div class="tabs">
  <a href="javascript:void(0)" class="tabs__btn active" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'dockerfile-image-tab')">Dockerfile Image</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'stapel-image-tab')">Stapel Image</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'stapel-artifact-tab')">Stapel Artifact</a>
</div>

<div id="dockerfile-image-tab" class="tabs__content active">
<a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vRrzxht-PmC-4NKq95DtLS9E7JrvtuHy0JpMKdylzlZtEZ5m7bJwEMJ6rXTLevFosWZXmi9t3rDVaPB/pub?w=2031&amp;h=144" data-featherlight="image">
<img src="https://docs.google.com/drawings/d/e/2PACX-1vRrzxht-PmC-4NKq95DtLS9E7JrvtuHy0JpMKdylzlZtEZ5m7bJwEMJ6rXTLevFosWZXmi9t3rDVaPB/pub?w=821&amp;h=59">
</a>
</div>

<div id="stapel-image-tab" class="tabs__content">
<a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vRKB-_Re-ZhkUSB45jF9GcM-3gnE2snMjTOEIQZSyXUniNHKK-eCQl8jw3tHFF-a6JLAr2sV73lGAdw/pub?w=2000&amp;h=881" data-featherlight="image">
<img src="https://docs.google.com/drawings/d/e/2PACX-1vRKB-_Re-ZhkUSB45jF9GcM-3gnE2snMjTOEIQZSyXUniNHKK-eCQl8jw3tHFF-a6JLAr2sV73lGAdw/pub?w=821&amp;h=362" >
</a>
</div>

<div id="stapel-artifact-tab" class="tabs__content">
<a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vRD-K_z7KEoliEVT4GpTekCkeaFMbSPWZpZkyTDms4XLeJAWEnnj4EeAxsdwnU3OtSW_vuKxDaaFLgD/pub?w=1800&amp;h=850" data-featherlight="image">
<img src="https://docs.google.com/drawings/d/e/2PACX-1vRD-K_z7KEoliEVT4GpTekCkeaFMbSPWZpZkyTDms4XLeJAWEnnj4EeAxsdwnU3OtSW_vuKxDaaFLgD/pub?w=640&amp;h=301">
</a>
</div>

**User only needs to write a config correсtly the rest of the work with stages are done by Werf.**

For every _stage_ at each build, Werf calculates build stage identifier called _stage signature_.
Each _stage_ is assembled in an ***assembly container*** based on the previous _stage_, and saved in [stages storage](#stages-storage).
The _stage signature_ is used for [tagging](#stage-naming) _stage_ in _stages storage_.
Werf does not build stages that already exist in _stages storage_ (like caching in Docker, but more complex).

The ***stage signature*** is the checksum of [stage dependencies]({{ site.baseurl }}/documentation/reference/stages_and_images.html#stage-dependencies) and previous _stage signature_. In the absence of _stage dependencies_, the _stage_ is skipped.

It means that the _stage conveyor_, can be reduced to several _stages_ or even to single _from_ stage.

<a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vR6qxP5dbQNlHXik0jCvEcKZS2gKbdNmbFa8XIem8pixSHSGvmL1n7rpuuQv64YWl48wLXfpwbLQEG_/pub?w=572&amp;h=577" data-featherlight="image">
<img src="https://docs.google.com/drawings/d/e/2PACX-1vR6qxP5dbQNlHXik0jCvEcKZS2gKbdNmbFa8XIem8pixSHSGvmL1n7rpuuQv64YWl48wLXfpwbLQEG_/pub?w=286&amp;h=288">
</a>

## Stage dependencies

_Stage dependency_ is some piece of data that affects stage _signature_. Stage dependency may be represented by:

 - some file from git repo with its content;
 - instructions to build stage specified in `werf.yaml`;
 - arbitrary string specified by user in `werf.yaml`; etc.

Most _stage dependencies_ are specified in `werf.yaml`, others relate to a runtime.

Tables below represent Dockerfile image, Stapel image and [Stapel artifact]({{ site.baseurl }}/documentation/configuration/stapel_artifact.html) _stages dependencies_.
Each row describes dependencies for certain stage.
Left column consists of short descriptions of dependencies, right includes related `werf.yaml` directives and contains relevant references for more information.

<div class="tabs">
  <a href="javascript:void(0)" id="image-from-dockerfile-dependencies" class="tabs__btn dependencies-btn">Dockerfile Image</a>
  <a href="javascript:void(0)" id="image-dependencies" class="tabs__btn dependencies-btn">Stapel Image</a>
  <a href="javascript:void(0)" id="artifact-dependencies" class="tabs__btn dependencies-btn">Stapel Artifact</a>
</div>

<div id="dependencies">
{% for stage in site.data.stages.entries %}
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

<link rel="stylesheet" href="{{ site.baseurl }}/css/stages.css">
<script src="{{ site.baseurl }}/js/jquery-3.1.0.min.js"></script>
<script>
function application() {
  if ($("a[id=image-from-dockerfile-dependencies]").hasClass('active')) {
    $(".image").addClass('hidden')
    $(".artifact").addClass('hidden')
    $(".image-from-dockerfile").removeClass('hidden')
  }
  else if ($("a[id=image-dependencies]").hasClass('active')) {
    $(".image-from-dockerfile").addClass('hidden')
    $(".artifact").addClass('hidden')
    $(".image").removeClass('hidden')
  }
  else if ($("a[id=artifact-dependencies]").hasClass('active')) {
    $(".image-from-dockerfile").addClass('hidden')
    $(".image").addClass('hidden')
    $(".artifact").removeClass('hidden')
  }
  else {
    $(".image-from-dockerfile").addClass('hidden')
    $(".image").addClass('hidden')
    $(".artifact").addClass('hidden')
  }
}

$('.tabs').on('click', '.dependencies-btn', function() {
  $(this).toggleClass('active').siblings().removeClass('active');
  application()
});

application()
$.noConflict();
</script>

## Stages storage

_Stages storage_ keeps project stages.
Stages can be stored in Docker Repo or locally, on a host machine.

Most commands use _stages_ and require specified _stages storage_, defined by `--stages-storage` option or `WERF_STAGES_STORAGE` environment variable.
At the moment, only local storage, `:local`, is supported.

### Stage naming

_Stages_ in _local stages storage_ are named by the following schema — `werf-stages-storage/PROJECT_NAME:STAGE_SIGNATURE`.

## Images

_Image_ is a **ready-to-use** Docker image, corresponding to a specific application state and [tagging strategy]({{ site.baseurl }}/documentation/reference/publish_process.html).

As it is written [above](#stages), _stages_ are steps in the assembly process, building blocks for constructing _images_.
_Stages_ are not intended for direct use, unlike images. The main difference between images and stages is [cleaning policies]({{ site.baseurl }}/documentation/reference/cleaning_process.html#cleanup-policies) due to stored meta-information.
The _stages storage_ cleanup is only based on the related images in _images repo_.

Werf creates _images_ using _stages storage_.
Currently, _images_ can only be created in a [_publishing process_]({{ site.baseurl }}/documentation/reference/publish_process.html) and be saved in _images repo_.

Images should be defined in the werf configuration file `werf.yaml`.

[See more info about build process]({{ site.baseurl }}/documentation/reference/build_process.html).
