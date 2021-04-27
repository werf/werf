---
title: Installation
permalink: installation.html
layout: default
sidebar: none
description: How to install werf?
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
      <div class="installation-selector__title">Version</div>
      <div class="tabs">
        <a href="javascript:void(0)" class="tabs__btn"
          data-install-tab-group="version" data-install-tab="1.2">1.2</a>
        <a href="javascript:void(0)" class="tabs__btn"
          data-install-tab-group="version" data-install-tab="1.1">1.1</a>
      </div>
    </div><!-- /selector -->
    <div class="installation-selector">
      <div class="installation-selector__title">Stability channel</div>
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
      <div class="installation-selector__title">OS</div>
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
      <div class="installation-selector__title">Installation method</div>
      <div class="tabs">
        <a href="javascript:void(0)" class="tabs__btn"
          data-install-tab-group="method" data-install-tab="multiwerf">using multiwerf (recommended)</a>
        <a href="javascript:void(0)" class="tabs__btn"
          data-install-tab-group="method" data-install-tab="binary">downloading binary package</a>
        <a href="javascript:void(0)" class="tabs__btn"
          data-install-tab-group="method" data-install-tab="source">compiling from source</a>
      </div>
    </div><!-- /selector -->
  </div><!-- /selector-row -->

  <div class="installation-instruction">
      <h1 class="installation-instruction__title">
        Install <span data-install-info="channel"></span> werf <span data-install-info="version"></span><br>
        for <span data-install-info="os"></span> by <span data-install-info="method"></span>
      </h1>
      <div class="docs">
<div class="details">
<h2 id="install-dependencies"><a href="javascript:void(0)" class="details__summary">Install dependencies</a></h2>
<div class="details__content" markdown="1">
{% include en/installation/multiwerf_dependencies.md %}
</div>
</div>
<h2 id="install-werf">Install werf</h2>
<div class="installation-instruction__tab-content" data-install-content-group="method" data-install-content="multiwerf">
<h3 id="install-multiwerf">Installing multiwerf</h3>
<div class="installation-instruction__tab-content" data-install-content-group="os" data-install-content="linux">
  {% for version in page.versions %}
    <div class="installation-instruction__tab-content" data-install-content-group="version" data-install-content="{{ version }}">
      {% for channel in page.channels %}
        <div class="installation-instruction__tab-content" data-install-content-group="channel" data-install-content="{{ channel }}">
<div markdown="1">{% include en/installation/multiwerf_unix.md version=version channel=channel %}</div>
<div class="details">
<h2><a href="javascript:void(0)" class="details__summary">How to use in the CI/CD system?</a></h2>
<div class="details__content" markdown="1">
{% include en/installation/multiwerf_unix/how_to_use_in_the_ci_cd.md version=version channel=channel %}
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
<div markdown="1">{% include en/installation/multiwerf_unix.md version=version channel=channel %}</div>
<div class="details">
<h2><a href="javascript:void(0)" class="details__summary">How to use in the CI/CD system?</a></h2>
<div class="details__content" markdown="1">
{% include en/installation/multiwerf_unix/how_to_use_in_the_ci_cd.md version=version channel=channel %}
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
<div markdown="1">{% include en/installation/multiwerf_windows.md version=version channel=channel %}</div>
        </div>
      {% endfor %}
    </div>
  {% endfor %}
</div><!-- /os -->

<div class="details">
<h2><a href="javascript:void(0)" class="details__summary">How it works?</a></h2>
<div class="details__content" markdown="1">
{% include en/installation/how_it_works.md %}
</div>
</div>
      </div><!-- /method -->
      <div class="installation-instruction__tab-content" data-install-content-group="method" data-install-content="binary">
<div markdown="1">
The latest release can be found [at this page](https://github.com/werf/werf/releases/)
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
    <h2 class="installation-channels__title" id="all-changes-in-werf-go-through-all-stability-channels">
      All changes in werf<br>
      go through all stability channels
    </h2>
    <ul class="installation-channels__channels">
      <li class="installation-channels__channel">
        <div class="installation-channels__channel-title">
          Alpha
        </div>
        <div class="installation-channels__channel-description">
          can bring new features<br>
          but can be unstable
        </div>
      </li>
      <li class="installation-channels__channel installation-channels__channel_beta">
        <div class="installation-channels__channel-title">
          Beta
        </div>
        <div class="installation-channels__channel-description">
          for more broad testing<br>
          of new features to catch<br>
          regressions
        </div>
      </li>
      <li class="installation-channels__channel installation-channels__channel_ea">
        <div class="installation-channels__channel-title">
          Early-Access
        </div>
        <div class="installation-channels__channel-description">
          is mostly safe and can be used<br>
          in non-critical environments<br>
          or for local development
        </div>
      </li>
      <li class="installation-channels__channel installation-channels__channel_stable">
        <div class="installation-channels__channel-title">
          Stable
        </div>
        <div class="installation-channels__channel-description">
          is mostly safe and we<br>
          encourage you to use<br>
          this version everywhere
        </div>
      </li>
      <li class="installation-channels__channel installation-channels__channel_rocksolid">
        <div class="installation-channels__channel-title">
          Rock-Solid
        </div>
        <div class="installation-channels__channel-description">
          the most stable channel<br>
          and recommended for usage<br>
          in critical environments with tight SLA
        </div>
      </li>
    </ul>
    <div class="installation-channels__info">
      <div class="installation-channels__info-versions">
        <p>When using release channels, you do not specify a&nbsp;version, because the version is&nbsp;managed automatically within the&nbsp;channel Stability channels and&nbsp;frequent releases allow receiving continuous feedback on&nbsp;new changes, quickly rolling problem changes back, ensuring the&nbsp;high stability of&nbsp;the&nbsp;software, and&nbsp;preserving an&nbsp;acceptable development speed at&nbsp;the&nbsp;same time.</p>
        <p>The relations between channels&nbsp;and werf releases are&nbsp;described in&nbsp;<a href="https://github.com/werf/werf/blob/multiwerf/multiwerf.json">multiwerf.json</a>.</p>
      </div>
      <div class="installation-channels__info-guarantees">
        <div class="installation-channels__info-guarantee">
          <strong>We guarantee</strong> that <i>Early-Access</i> release should become <i>Stable</i> not earlier than 1 week after internal tests.
        </div>
        <div class="installation-channels__info-guarantee">
          <strong>We guarantee</strong> that <i>Stable</i> release should become <i>Rock-Solid</i> release not earlier than after 2 weeks of&nbsp;extensive testing.
        </div>
      </div>
    </div>
  </div>
  <div class="installation-compatibility">
    <h2 class="installation-compatibility__title" id="backward-compatibility-promise">Backward compatibility promise</h2>
<div markdown="1" class="docs">
{% include en/installation/backward-compatibility.md %}
</div>
  </div>
  <div class="installation-releases">
  <div class="installation-releases__block-title">
      Changelog history of releases within channels
      <a href="/feed.xml" title="RSS" target="_blank" class="page__icon page__icon_rss page__icon_block-title page__icon_link"></a>
  </div>
  <div class="installation-releases__block-subtitle">
      Release
  </div>

  <div class="tabs">
    {%- for group in groups %}
    <a href="javascript:void(0)" class="tabs__btn tabs__group__btn{% if group == groups[0] %} active{% endif %}" onclick="openTab(event, 'tabs__group__btn', 'tabs__group__content', 'group-{{group}}')">{{group}}</a>
    {%- endfor %}
  </div>

  {%- for group in groups %}
  <div id="group-{{group}}" class="tabs__content tabs__group__content{% if group == groups[0] %} active{% endif %}">
      <div class="installation-releases__block-subtitle">
          Channel
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
          <a href="javascript:void(0)" class="tabs__btn tabs__{{group}}__channel__btn" onclick="openTab(event, 'tabs__{{group}}__channel__btn', 'tabs__{{group}}__channel__content', 'id-{{group}}-all')">All channels</a>
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
              <p>There are no versions on the channel yet, but they will appear soon.</p>
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
            <p>This is a list of all of the releases (Alpha, Beta, Early-Access, Stable and Rock-Solid) combined in chronological order.</p>
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
