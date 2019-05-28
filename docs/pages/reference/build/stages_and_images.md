---
title: Stages and Images
sidebar: reference
permalink: reference/build/stages_and_images.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
---

We propose to divide the assembly proccess into steps, intermediate images (like layers in Docker), with clear functions and assignments.
In Werf, such step is called [stage](#stage) and result [image](#image) consists of a set of built stages.
All stages are kept in a [stages storage](#stages-storage) and defining build cache of application (not really cache but part of building context).

## Stages

Stages are steps in the assembly process, building blocks for constructing images.
A ***stage*** is built from a logically grouped set of config instructions, taking into account the assembly conditions and rules.
Each _stage_ relates to one Docker image.

The werf assembly process assumes a sequential build of stages using _stage conveyor_.  A _stage conveyor_ is a sequence with the predefined order and set of stages. Werf uses different _stage conveyor_ for assembling a particular type of build object.

<div class="tabs">
  <a href="javascript:void(0)" class="tabs__btn active" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'image')">Image</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'artifact')">Artifact</a>
</div>

<div id="image" class="tabs__content active">
<a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vRKB-_Re-ZhkUSB45jF9GcM-3gnE2snMjTOEIQZSyXUniNHKK-eCQl8jw3tHFF-a6JLAr2sV73lGAdw/pub?w=2000&amp;h=881" data-featherlight="image">
<img src="https://docs.google.com/drawings/d/e/2PACX-1vRKB-_Re-ZhkUSB45jF9GcM-3gnE2snMjTOEIQZSyXUniNHKK-eCQl8jw3tHFF-a6JLAr2sV73lGAdw/pub?w=821&amp;h=362" >
</a>
</div>

<div id="artifact" class="tabs__content">
<a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vRD-K_z7KEoliEVT4GpTekCkeaFMbSPWZpZkyTDms4XLeJAWEnnj4EeAxsdwnU3OtSW_vuKxDaaFLgD/pub?w=1800&amp;h=850" data-featherlight="image">
<img src="https://docs.google.com/drawings/d/e/2PACX-1vRD-K_z7KEoliEVT4GpTekCkeaFMbSPWZpZkyTDms4XLeJAWEnnj4EeAxsdwnU3OtSW_vuKxDaaFLgD/pub?w=640&amp;h=301">
</a>
</div>

**All works with _stages_ are done by werf, and you only need to write config correctly.**

For every _stage_ at each build, Werf calculates build stage identifier called _stage signature_.
Each _stage_ is assembled in an ***assembly container*** based on the previous _stage_, and saved in [stages storage](#stages-storage).
The _stage signature_ is used for [tagging](#stage-naming) _stage_ in _stages storage_.
Werf does not build stages that already exist in _stages storage_ (like caching in Docker, but more complex).

The ***stage signature*** is the checksum of [stage dependencies]({{ site.baseurl }}/reference/build/stages_and_images.html#stage-dependencies) and previous _stage signature_. In the absence of _stage dependencies_, the _stage_ is skipped.

It means that the _stage conveyor_, can be reduced to several _stages_ or even to single _from_ stage.

<a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vR6qxP5dbQNlHXik0jCvEcKZS2gKbdNmbFa8XIem8pixSHSGvmL1n7rpuuQv64YWl48wLXfpwbLQEG_/pub?w=572&amp;h=577" data-featherlight="image">
<img src="https://docs.google.com/drawings/d/e/2PACX-1vR6qxP5dbQNlHXik0jCvEcKZS2gKbdNmbFa8XIem8pixSHSGvmL1n7rpuuQv64YWl48wLXfpwbLQEG_/pub?w=286&amp;h=288">
</a>

## Stage dependencies

Most _stage dependencies_ are specified in werf.yaml, others relate to a runtime.
Changing these dependencies affects on a _signature_, stages reassembling.

The tables bellow represent image and [artifact]({{ site.baseurl }}/reference/build/artifact.html) _stages dependencies_ and contain relevant references for more information.

<div class="tabs">
  <a href="javascript:void(0)" id="image-dependencies" class="tabs__btn dependencies-btn">Image</a>
  <a href="javascript:void(0)" id="artifact-dependencies" class="tabs__btn dependencies-btn">Artifact</a>
</div>

<div id="dependencies">
{% for stage in site.data.stages.entries %}
<div class="stage {{stage.type}}">
  <div class="stage-body">
    <div class="stage-base">
      <p>{{ stage.name | escape }}</p>

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
  if ($("a[id=image-dependencies]").hasClass('active')) {
    $(".artifact").addClass('hidden')
    $(".image").removeClass('hidden')
  }
  else if ($("a[id=artifact-dependencies]").hasClass('active')) {
    $(".image").addClass('hidden')
    $(".artifact").removeClass('hidden')
  }
  else {
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

## Image

_Image_ is a **ready-to-use** Docker image, corresponding to a specific application state and [tagging strategy]({{ site.baseurl }}/reference/registry/image_naming.html#image-tag-parameters).

As it is written [above](#stages), _stages_ are steps in the assembly process, building blocks for constructing _images_. 
_Stages_ are not intended for direct use, unlike images. The main difference between images and stages is [cleaning policies]({{ site.baseurl }}/reference/registry/cleaning.html#cleanup-policies) due to stored meta-information.
The _stages storage_ cleanup is only based on the related images in _images repo_. 

Werf creates _images_ using _stages storage_.
Currently, _images_ can only be created in a [_publishing process_]({{ site.baseurl }}/reference/registry/publish.html) and be saved in [_images repo_]({{ site.baseurl }}/reference/registry/image_naming.html#images-repo).
