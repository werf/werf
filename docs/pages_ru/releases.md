---
title: Releases
permalink: releases.html
sidebar: documentation
layout: default
---

<link rel="stylesheet" href="{{ site.baseurl }}/css/releases.css">

{% assign releases = site.data.releases.releases %}
<!--
<div class="main-container page__container releases">
    {% for release in releases %}
        <div class="releases__title">
            <a href="{{ release.html_url }}">
                {{ release.name }}
            </a>
        </div>
        <div class="releases__body">
            {{ release.body | markdownify }}
        </div>
    {% endfor %}
</div>
--> 

<div class="page__container">

<strong>Каналы обновлений</strong><br /><br />

<!-- latest version per channel -->
{% assign channels_sorted = site.data.releases_info.channels | sort: "stability" %}
{% assign channels_sorted_reverse = site.data.releases_info.channels | sort: "stability" | reverse  %}

<div class="page__container releases__content">
{% for channel in channels_sorted_reverse %}
{% assign version = site.data.releases_info.versions | where: "name",  channel.name | map: "version" | first %}
    <div class="releases__menu-item">
       <div><b>{{ channel.title }}</b>
       </div>
       <small><span class="label label-primary ">{% if version != null %}{{ version }}{% else %}-{% endif %}</span></small>
       <div class="releases__menu-item">
          {{ channel.description }}
       </div>
       <div class="releases__menu-item">
         <a href="https://github.com/flant/werf/releases/tag/{{ version }}">View on GitHub</a>
       </div>
    </div>
  {% endfor %}
</div>
<br />
<!-- Releases description -->

<div class="page__container releases">
<p>| <em>Note:</em> Настоящее обещание относится к werf, начиная с версии 1.0, и не относится к предыдущим версиям или версиям dapp.</p>

<p>werf использует <a href="https://semver.org/lang/ru/">семантическое версионирование</a>. Это значит, что мажорные версии (1.0, 2.0) могут быть обратно не совместимыми между собой. В случае werf это означает, что обновление на следующую мажорную версию <em>может</em> потребовать полного передеплоя приложений, либо других ручных операций.</p>

<p>Минорные версии (1.1, 1.2, etc) могут добавлять новые “значительные” изменения, но без существенных проблем обратной совместимости в пределах мажорной версии. В случае werf это означает, что обновление на следующую минорную версию в большинстве случаев будет беспроблемным, но <em>может</em> потребоваться запуск предоставленных скриптов миграции.</p>

<p>Патч-версии (1.1.0, 1.1.1, 1.1.2) могут добавлять новые возможности, но без каких-либо проблем обратной совместимости в пределах минорной версии (1.1.x). В случае werf это означает, что обновление на следующий патч (следующую патч-версию) не должно вызывать проблем и требовать каких-либо ручных действий.</p>

<p>Патч-версии делятся на каналы обновлений. Канал обновлений — это префикс пререлизной части версии (1.1.0-alpha.2, 1.1.0-beta.3, 1.1.0-ea.1). Версия без пререлизной части считается версией стабильного канала обновлений.</p>
</div>

<hr />

<strong>Релизы</strong><br /><br />

<div class="tabs">
{% for channel in channels_sorted_reverse %}
<a href="javascript:void(0)" class="tabs__btn{% if channel == channels_sorted_reverse[0] %} active{% endif %}" onclick="openTab(event, 'tabs__btn', 'tabs__content', '{{channel.name}}')">Канал {{channel.title}} </a>
{% endfor %}
<a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'all')">Все каналы</a>
</div>

{% for channel in channels_sorted_reverse %}
<div id="{{ channel.name }}" class="tabs__content{% if channel == channels_sorted_reverse[0] %} active{% endif %}">
  <p>| {{ channel.tooltip }}</p>
  <p>{{ channel.description }}</p>

  {% assign _releases = site.data.releases.releases | where: "channel", channel.name %}
  {% if _releases.size > 0 %}
    {% for release in _releases %}
        <div class="releases__title">
            <a href="{{ release.html_url }}">
                {{ release.name }}
            </a>
        </div>
        <div class="releases__body">
            {{ release.body | markdownify }}
        </div>
    {% endfor %}
  {% else %}
      <p>На канале пока нет версий, но обязательно скоро появятся.</p>
  {% endif %}

</div>
{% endfor %}

{% assign _releases = site.data.releases.releases %}
<div id="all" class="tabs__content">
{% for release in _releases %}
  <p>Все релизы в хронологическом порядке</p>
        <div class="releases__title">
            <a href="{{ release.html_url }}">
                {{ release.name }}
            </a>
        </div>
        <div class="releases__body">
            {{ release.body | markdownify }}
        </div>
{% endfor %}
</div>


</div>