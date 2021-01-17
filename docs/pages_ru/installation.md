---
title: Installation
permalink: installation.html
layout: default
description: Как установить werf?
versions:
  - 1.2
  - 1.1
  - 1.0
channels:
  - alpha
  - beta
  - ea
  - stable
  - rock-solid
---
{%- asset installation.css %}

<div class="page__container page_installation">

  <div class="installation-selector-row">
    <div class="installation-selector">
      <div class="installation-selector__title">Версия</div>
      <div class="tabs">
        <a href="javascript:void(0)" class="tabs__btn"
          data-install-tab-group="version" data-install-tab="1.2">1.2</a>
        <a href="javascript:void(0)" class="tabs__btn"
          data-install-tab-group="version" data-install-tab="1.1">1.1</a>
        <a href="javascript:void(0)" class="tabs__btn"
          data-install-tab-group="version" data-install-tab="1.0">1.0</a>
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
Последний релиз может быть найден [на данной странице](https://bintray.com/flant/werf/werf/_latestVersion)
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
          the most stable channel<br>
          and recommended for usage<br>
          in critical environments with tight SLA
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
</div>
