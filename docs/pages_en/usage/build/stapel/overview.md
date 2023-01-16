---
title: Overview
permalink: usage/build/stapel/overview.html
---

werf has a built-in alternative syntax for describing assembly instructions called Stapel. Here are its distinctive features:

1. Easily support and parameterize complex configurations, reuse common snippets and generate configurations of the images of the same type using YAML format and templating.
2. Dedicated commands for integrating with git to enable incremental rebuilds based on the git repository history.
3. Inheriting images and importing files from images (similar to Dockerfile's multi-stage).
4. Run arbitrary build instructions, specify directory mount options, and use other advanced tools to build images.
5. More efficient caching mechanics for layers (a similar scheme is supported for Dockerfile layers when building with Buildah (currently pre-alpha)).

<!-- TODO(staged-dockerfile): удалить 5 пункт как неактуальный -->

To build images using the Stapel builder, you need to describe the build instructions in the `werf.yaml` configuration file. Stapel is supported for both the docker server builder backend (building via shell instructions or ansible) and for buildah (shell instructions only).

This section describes how to build images with the Stapel builder, its advanced features and how to use them.

<div class="details">
<a href="javascript:void(0)" class="details__summary">Как устроен конвеер стадий stapel</a>
<div class="details__content" markdown="1">

A _stage conveyor_ is an ordered sequence of conditions and rules for carrying out stages. werf uses different _stage conveyors_ to assemble various types of images depending on their configuration.

<div class="tabs">
  <a href="javascript:void(0)" class="tabs__btn active" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'stapel-image-tab')">Stapel Image</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'stapel-artifact-tab')">Stapel Artifact</a>
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

For each _stage_ at every build, werf calculates the unique identifier of the stage called stage digest.

If a _stage_ has no stage dependencies, it is skipped, and, accordingly, the _stage conveyor_ is reduced by one stage. It means that the _stage conveyor_ can be reduced to several _stages_ or even to a single _from_ stage.

<a class="google-drawings" href="{{ "images/reference/stages_and_images4.png" | true_relative_url }}" data-featherlight="image">
<img src="{{ "images/reference/stages_and_images4_preview.png" | true_relative_url }}">
</a>

_Stage dependency_ is a piece of data that affects the stage _digest_. Stage dependency may be represented by:

 - some file from a git repo with its contents;
 - instructions to build stage defined in the `werf.yaml`;
 - the arbitrary string specified by the user in the `werf.yaml`;
 - and so on.

Most _stage dependencies_ are specified in the `werf.yaml`, others relate to a runtime.

The tables below illustrate dependencies of a Dockerfile image, a Stapel image, and a [Stapel artifact]({{ "usage/build/stapel/imports.html" | true_relative_url }}) _stages dependencies_.
Each row describes dependencies for a certain stage.
The left column contains a short description of dependencies, the right column includes related `werf.yaml` directives and contains relevant references for more information.

<div class="tabs">
  <a href="javascript:void(0)" id="image-dependencies" class="tabs__btn dependencies-btn active">Stapel Image</a>
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
  if ($("a[id=image-dependencies]").hasClass('active')) {
    $(".artifact").addClass('hidden');
    $(".image").removeClass('hidden')
  }
  else if ($("a[id=artifact-dependencies]").hasClass('active')) {
    $(".image").addClass('hidden');
    $(".artifact").removeClass('hidden')
  }
  else {
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

</div>
</div>
