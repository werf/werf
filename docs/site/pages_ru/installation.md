---
title: Installation
permalink: installation.html
layout: default
sidebar: none
description: Как установить werf?
versions:
  - 1.2
  - 1.1
channels:
  - alpha
  - beta
  - ea
  - stable
  - rock-solid
---
{%- asset installation.css %}
{%- asset installation.js %}
{%- asset releases.css %}

{%- assign releases = site.data.releases.releases %}
{%- assign groups = site.data.releases_history.history | map: "group" | uniq | reverse %}
{%- assign channels_sorted = site.data.channels_info.channels | sort: "stability" %}
{%- assign channels_sorted_reverse = site.data.channels_info.channels | sort: "stability" | reverse  %}

<div class="page__container page_installation">

  <div class="installation-selector-row">
    <div class="installation-selector">
      <div class="installation-selector__title">Версия</div>
      <div class="tabs">
        <a href="javascript:void(0)" class="tabs__btn"
          data-install-tab-group="version" data-install-tab="1.2">1.2</a>
        <a href="javascript:void(0)" class="tabs__btn"
          data-install-tab-group="version" data-install-tab="1.1">1.1</a>
      </div>
    </div><!-- /selector -->
    <div class="installation-selector">
      <div class="installation-selector__title">Уровень стабильности</div>
      <div class="tabs">
        <a href="javascript:void(0)" class="tabs__btn"
          data-install-tab-group="channel" data-install-tab="rock-solid">Rock-Solid</a>
        <a href="javascript:void(0)" class="tabs__btn"
          data-install-tab-group="channel" data-install-tab="stable">Stable</a>
        <a href="javascript:void(0)" class="tabs__btn"
          data-install-tab-group="channel" data-install-tab="ea">Early-Access</a>
        <a href="javascript:void(0)" class="tabs__btn"
          data-install-tab-group="channel" data-install-tab="beta">Beta</a>
        <a href="javascript:void(0)" class="tabs__btn"
          data-install-tab-group="channel" data-install-tab="alpha">Alpha</a>
      </div>
    </div><!-- /selector -->
    <div class="installation-selector">
      <div class="installation-selector__title">Операционная система</div>
      <div class="tabs">
        <a href="javascript:void(0)" class="tabs__btn"
          data-install-tab-group="os" data-install-tab="linux">Linux</a>
        <a href="javascript:void(0)" class="tabs__btn"
          data-install-tab-group="os" data-install-tab="macos">Mac OS</a>
        <a href="javascript:void(0)" class="tabs__btn"
          data-install-tab-group="os" data-install-tab="windows">Windows</a>
      </div>
    </div><!-- /selector -->
  </div><!-- /selector-row -->
  <div class="installation-selector-row">
    <div class="installation-selector">
      <div class="installation-selector__title">Метод установки</div>
      <div class="tabs">
        <a href="javascript:void(0)" class="tabs__btn"
          data-install-tab-group="method" data-install-tab="multiwerf">через multiwerf (рекомендовано)</a>
        <a href="javascript:void(0)" class="tabs__btn"
          data-install-tab-group="method" data-install-tab="binary">бинарным файлом</a>
        <a href="javascript:void(0)" class="tabs__btn"
          data-install-tab-group="method" data-install-tab="source">из исходников</a>
      </div>
    </div><!-- /selector -->
  </div><!-- /selector-row -->

  <div class="installation-instruction">
      <h1 class="installation-instruction__title">
        Установить <span data-install-info="channel"></span> werf <span data-install-info="version"></span><br>
        для <span data-install-info="os"></span> <span data-install-info="method"></span>
      </h1>
      <div class="docs">
<div class="details">
<h2 id="установка-зависимостей"><a href="javascript:void(0)" class="details__summary">Установка зависимостей</a></h2>
<div class="details__content" markdown="1">
{% include ru/installation/multiwerf_dependencies.md %}
</div>
</div>
<h2 id="установка-werf">Установка werf</h2>
<div class="installation-instruction__tab-content" data-install-content-group="method" data-install-content="multiwerf">
<h3>Установка multiwerf</h3>
<div class="installation-instruction__tab-content" data-install-content-group="os" data-install-content="linux">
  {% for version in page.versions %}
    <div class="installation-instruction__tab-content" data-install-content-group="version" data-install-content="{{ version }}">
      {% for channel in page.channels %}
        <div class="installation-instruction__tab-content" data-install-content-group="channel" data-install-content="{{ channel }}">
<div markdown="1">{% include ru/installation/multiwerf_unix.md version=version channel=channel %}</div>
<div class="details">
<h2 id="как-использовать-в-cicd-системе"><a href="javascript:void(0)" class="details__summary">Как использовать в CI/CD системе?</a></h2>
<div class="details__content" markdown="1">
{% include ru/installation/multiwerf_unix/how_to_use_in_the_ci_cd.md version=version channel=channel %}
</div>
</div>
        </div>
      {% endfor %}
    </div>
  {% endfor %}
</div><!-- /os -->
<div class="installation-instruction__tab-content" data-install-content-group="os" data-install-content="macos">
  {% for version in page.versions %}
    <div class="installation-instruction__tab-content" data-install-content-group="version" data-install-content="{{ version }}">
      {% for channel in page.channels %}
        <div class="installation-instruction__tab-content" data-install-content-group="channel" data-install-content="{{ channel }}">
<div markdown="1">{% include ru/installation/multiwerf_unix.md version=version channel=channel %}</div>
<div class="details">
<h2 id="как-использовать-в-cicd-системе"><a href="javascript:void(0)" class="details__summary">Как использовать в CI/CD системе?</a></h2>
<div class="details__content" markdown="1">
{% include ru/installation/multiwerf_unix/how_to_use_in_the_ci_cd.md version=version channel=channel %}
</div>
</div>
        </div>
      {% endfor %}
    </div>
  {% endfor %}
</div><!-- /os -->
<div class="installation-instruction__tab-content" data-install-content-group="os" data-install-content="windows">
  {% for version in page.versions %}
    <div class="installation-instruction__tab-content" data-install-content-group="version" data-install-content="{{ version }}">
      {% for channel in page.channels %}
        <div class="installation-instruction__tab-content" data-install-content-group="channel" data-install-content="{{ channel }}">
<div markdown="1">{% include ru/installation/multiwerf_windows.md version=version channel=channel %}</div>
        </div>
      {% endfor %}
    </div>
  {% endfor %}
</div><!-- /os -->

<div class="details">
<h2><a href="javascript:void(0)" class="details__summary">Как это работает?</a></h2>
<div class="details__content" markdown="1">
{% include ru/installation/how_it_works.md %}
</div>
</div>
      </div><!-- /method -->
      <div class="installation-instruction__tab-content" data-install-content-group="method" data-install-content="binary">
<div markdown="1">
Последний релиз может быть найден [на данной странице](https://github.com/werf/werf/releases/)
</div>
        <div class="installation-instruction__tab-content" data-install-content-group="os" data-install-content="linux">
  {% for version in page.versions %}
    <div class="installation-instruction__tab-content" data-install-content-group="version" data-install-content="{{ version }}">
      {% for channel in page.channels %}
        <div class="installation-instruction__tab-content" data-install-content-group="channel" data-install-content="{{ channel }}">
  {% capture version_key %}{{ channel }}-{{ version }}{% endcapture %}
  {% assign download_version = site.data.channels_versions.versions[version_key] %}
<div markdown="1">
{% include installation/binary_linux.md version=download_version %}
</div>
        </div>
      {% endfor %}
    </div>
  {% endfor %}

        </div>
        <div class="installation-instruction__tab-content" data-install-content-group="os" data-install-content="macos">
  {% for version in page.versions %}
    <div class="installation-instruction__tab-content" data-install-content-group="version" data-install-content="{{ version }}">
      {% for channel in page.channels %}
        <div class="installation-instruction__tab-content" data-install-content-group="channel" data-install-content="{{ channel }}">
  {% capture version_key %}{{ channel }}-{{ version }}{% endcapture %}
  {% assign download_version = site.data.channels_versions.versions[version_key] %}
<div markdown="1">
{% include installation/binary_macos.md version=download_version %}
</div>
        </div>
      {% endfor %}
    </div>
  {% endfor %}

        </div>
        <div class="installation-instruction__tab-content" data-install-content-group="os" data-install-content="windows">
  {% for version in page.versions %}
    <div class="installation-instruction__tab-content" data-install-content-group="version" data-install-content="{{ version }}">
      {% for channel in page.channels %}
        <div class="installation-instruction__tab-content" data-install-content-group="channel" data-install-content="{{ channel }}">
  {% capture version_key %}{{ channel }}-{{ version }}{% endcapture %}
  {% assign download_version = site.data.channels_versions.versions[version_key] %}
<div markdown="1">
{% include installation/binary_windows.md version=download_version %}
</div>
        </div>
      {% endfor %}
    </div>
  {% endfor %}

        </div>
      </div><!-- /method -->
      <div class="installation-instruction__tab-content" data-install-content-group="method" data-install-content="source">
<div markdown="1">
{% include installation/source.md %}
</div>
      </div><!-- /method -->
    </div>
  </div>

  <div class="installation-channels">
    <h2 class="installation-channels__title" id="все-изменения-в-werf-проходят-через-цепочку-каналов-стабильности">
      Все изменения в werf<br>
      проходят через цепочку каналов стабильности
    </h2>
    <ul class="installation-channels__channels">
      <li class="installation-channels__channel">
        <div class="installation-channels__channel-title">
          Alpha
        </div>
        <div class="installation-channels__channel-description">
          быстро доставляет новые возможности<br>
          однако может быть нестабильным
        </div>
      </li>
      <li class="installation-channels__channel installation-channels__channel_beta">
        <div class="installation-channels__channel-title">
          Beta
        </div>
        <div class="installation-channels__channel-description">
          для более широкого тестирования<br>
          новых возможностей с целью<br>
          обнаружить проблемы
        </div>
      </li>
      <li class="installation-channels__channel installation-channels__channel_ea">
        <div class="installation-channels__channel-title">
          Early-Access
        </div>
        <div class="installation-channels__channel-description">
          достаточно безопасен для использования<br>
          в некритичных окружениях и для<br>
          локальной разработки,<br>
          позволяет раньше получать новые возможности
        </div>
      </li>
      <li class="installation-channels__channel installation-channels__channel_stable">
        <div class="installation-channels__channel-title">
          Stable
        </div>
        <div class="installation-channels__channel-description">
          безопасен и рекомендуется для широкого<br>
          использования в любых окружениях,<br>
          как вариант по умолчанию.
        </div>
      </li>
      <li class="installation-channels__channel installation-channels__channel_rocksolid">
        <div class="installation-channels__channel-title">
          Rock-Solid
        </div>
        <div class="installation-channels__channel-description">
          наиболее стабильный канал,<br>
          рекомендован для критичных окружений<br>
          со строгими требованиями SLA.
        </div>
      </li>
    </ul>
    <div class="installation-channels__info">
      <div class="installation-channels__info-versions">
        <p>При использовании каналов стабильности не требуется указывать конкретную версию, т.к. конкретную версию активирует multiwerf, выступая в роли менеджера версий. Это позволяет автоматически и непрерывно получать как исправления проблем, так и новые возможности, оперативно откатывать проблемные изменения. В целом такая схема даёт баланс между достаточно высоким уровнем стабильности софта и быстрой разработкой новых возможностей.</p>
        <p>Связи между каналом стабильности и конкретной версией werf описываются в специально файле <a href="https://github.com/werf/werf/blob/multiwerf/multiwerf.json">multiwerf.json</a>.</p>
      </div>
      <div class="installation-channels__info-guarantees">
        <div class="installation-channels__info-guarantee">
          <strong>Мы гарантируем</strong>, что релиз из канал <i>Early-Access</i> попадёт в канал <i>Stable</i> не раньше, чем через 1 неделю после внутреннего тестирования.
        </div>
        <div class="installation-channels__info-guarantee">
          <strong>Мы гарантируем</strong>, что релиз из канала <i>Stable</i> должен попасть в канал <i>Rock-Solid</i> не раньше, чем через 2 недели активного тестирования.
        </div>
      </div>
    </div>
  </div>
  <div class="installation-compatibility">
    <h2 class="installation-compatibility__title" id="гарантии-обратной-совместимости">Гарантии обратной совместимости</h2>
<div markdown="1" class="docs">
{% include ru/installation/backward-compatibility.md %}
</div>
  </div>
  <div class="installation-releases">
  <div class="installation-releases__block-title">
      История изменений версий в каналах обновлений
      <a href="/feed.xml" title="RSS" target="_blank" class="page__icon page__icon_rss page__icon_block-title page__icon_link"></a>
  </div>
  <div class="installation-releases__block-subtitle">
      Релиз
  </div>

  <div class="tabs">
    {%- for group in groups %}
    <a href="javascript:void(0)" class="tabs__btn tabs__group__btn{% if group == groups[0] %} active{% endif %}" onclick="openTab(event, 'tabs__group__btn', 'tabs__group__content', 'group-{{group}}')">{{group}}</a>
    {%- endfor %}
  </div>

  {%- for group in groups %}
  <div id="group-{{group}}" class="tabs__content tabs__group__content{% if group == groups[0] %} active{% endif %}">
      <div class="installation-releases__block-subtitle">
          Уровень стабильности
      </div>
      <div class="tabs">
        {%- assign not_activated = true %}
        {%- assign active_channels = 0 %}
        {%- for channel in channels_sorted_reverse %}
          {%- assign channel_activity = site.data.releases_history.history | reverse | where: "group", group | where: "name", channel.name | size %}
          {%- if channel_activity < 1 %}
            {%- continue %} 
          {%- endif %}
          <a href="javascript:void(0)" class="tabs__btn tabs__{{group}}__channel__btn{% if channel_activity > 0 and not_activated and channel != channels_sorted_reverse[0] %} active{% endif %}" onclick="openTab(event, 'tabs__{{group}}__channel__btn', 'tabs__{{group}}__channel__content', 'id-{{group}}-{{channel.name}}')">{{channel.title}}</a>
          {%- if channel_activity > 0 and not_activated and channel != channels_sorted_reverse[0] %}
          {%- assign not_activated = false %}
          {% endif %}
          {%- assign active_channels = active_channels | plus: 1 %}
        {%- endfor %}
        {%- if active_channels > 10 %}
          <a href="javascript:void(0)" class="tabs__btn tabs__{{group}}__channel__btn" onclick="openTab(event, 'tabs__{{group}}__channel__btn', 'tabs__{{group}}__channel__content', 'id-{{group}}-all')">Все уровни</a>
        {%- endif %}
      </div>

      {%- assign not_activated = true %}
      {%- assign active_channels = 0 %}
      {%- for channel in channels_sorted_reverse %}
      {%- assign channel_activity = site.data.releases_history.history | reverse | where: "group", group | where: "name", channel.name | size %}
      {%- if channel_activity < 1 %}
        {% continue %} 
      {% endif %}
      <div id="id-{{group}}-{{ channel.name }}" class="tabs__content tabs__{{group}}__channel__content{% if channel_activity > 0 and not_activated and channel != channels_sorted_reverse[0]  %} active{% endif %}">
        <div class="installation-releases__info">
          <p>
            {{ channel.tooltip[page.lang] }}
            <a href="/feed-{{group}}-{{ channel.name }}.xml" title="RSS" target="_blank" class="page__icon page__icon_rss page__icon_text page__icon_link"></a>
          </p>
          <p class="installation-releases__info-text">{{ channel.description[page.lang] }}</p>
        </div>

        {%- assign group_history = site.data.releases_history.history | reverse | where: "group", group %}
        {%- assign channel_history = group_history | where: "name", channel.name %}
    
        {%- if channel_history.size > 0 %}
          {%- for channel_action in channel_history %}
            {%- assign release = site.data.releases.releases | where: "tag_name", channel_action.version | first %}            
              <div class="installation-releases__header">
                  <a href="{{ release.html_url }}" class="installation-releases__title">{{ release.tag_name }}</a>
                  <div class="installation-releases__date">{{ channel_action.ts | date: "%b %-d, %Y at %H:%M %z" }}</div>
              </div>
              <div class="installation-releases__body">
                  {{ release.body | markdownify }}
              </div>
          {%- endfor %}
        {%- else %}
          <div class="installation-releases__info releases__info_notification">
              <p>На канале пока нет версий, но обязательно скоро появятся.</p>
          </div>
        {%- endif %}

      </div>
      {%- if channel_activity > 0 and not_activated and channel != channels_sorted_reverse[0] %}
        {%- assign not_activated = false %}
      {%- endif %}
      {%- assign active_channels = active_channels | plus: 1 %}

      {%- endfor %}

      {%- comment %}
      {%- if active_channels > 10 %}
      <div id="id-{{group}}-all" class="tabs__content tabs__{{group}}__channel__content">
        <div class="installation-releases__info">
            <p>Список всех релизов (Alpha, Beta, Early-Access, Stable и Rock-Solid) в хронологическом порядке.</p>
        </div>
        {%- assign group_history = site.data.releases_history.history | reverse | where: "group", group | map: "version" | reverse | uniq %}
        {%- for release_data in group_history %}
            {%- assign release = site.data.releases.releases | where: "tag_name", release_data | first %}
            <div class="installation-releases__header">
                <div class="installation-releases__date">{{ channel_action.ts | date: "%b %-d, %Y at %H:%M %z" }}</div>
                <a href="{{ release.html_url }}" class="installation-releases__title">{{ release.tag_name }}</a>              
            </div>
            <div class="installation-releases__body">
                {{ release.body | markdownify }}
            </div>
        {%- endfor %}
      </div>
      {%- endif %}
      {%- endcomment %}
  </div>
  {%- endfor %}
</div>
