---
title: Installation
permalink: installation.html
layout: default
---

{%- asset releases.css %}

{%- assign releases = site.data.releases.releases %}

[Installation.]({{ site.base_url}}/installation_old.html)

<div class="page__container page_releases">

<div class="releases__block-title">
    Release channels <a href="/feed.xml" title="RSS" target="_blank" class="page__icon page__icon_rss page__icon_block-title page__icon_link"></a>
</div>

<!-- Releases description -->
<div class="releases__info">
    Each werf release progresses through all release channels, starting with Alpha → Beta → Early-Access → Stable → Rock-Solid. You can think of each release on a lower channel as a release-candidate for the higher one. Once a release is considered bug-free, it is promoted to the next channel.
</div>

{%- assign groups = site.data.releases_history.history | map: "group" | uniq | reverse %}
{%- assign channels_sorted = site.data.channels_info.channels | sort: "stability" %}
{%- assign channels_sorted_reverse = site.data.channels_info.channels | sort: "stability" | reverse  %}

<div class="releases__menu">
{%- for channel in channels_sorted_reverse %}
{%- assign channel_latest_versions = site.data.releases_history.latest | where: "name",  channel.name | first| map: "versions" | first | reverse | default: nil %}
    <div class="releases__menu-item">
        <div class="releases__menu-item-header">            
            <div class="releases__menu-item-title">
                {{ channel.title }}
            </div>
            <div class="releases__menu-item-versions">
            {%- for version in channel_latest_versions %}
            {%- if version != nil  %}
            {%- assign version_info = site.data.releases.releases | where: "tag_name", version | first %}
                <a href="{{ version_info.html_url }}" class="releases__btn">
                {{ version }}
                </a>
            {%- endif %}
            {%- endfor %}
            </div>
        </div>        
        <div class="releases__menu-item-description">
            {{ channel.description[page.lang] }}
        </div>
    </div>
{%- endfor %}
</div>

<div class="releases__block-title">
    Changelog history of releases within channels
</div>

<div class="releases">

<div class="releases__block-subtitle">
    Release:
</div>

<div class="tabs">
  {%- for group in groups %}
  <a href="javascript:void(0)" class="tabs__btn tabs__group__btn{% if group == groups[0] %} active{% endif %}" onclick="openTab(event, 'tabs__group__btn', 'tabs__group__content', 'group-{{group}}')">{{group}}</a>
  {%- endfor %}
</div>

{%- for group in groups %}
<div id="group-{{group}}" class="tabs__content tabs__group__content{% if group == groups[0] %} active{% endif %}">
    <div class="releases__block-subtitle">
        Channel:
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
      <div class="releases__info">
        <p>
          {{ channel.tooltip[page.lang] }}
          <a href="/feed-{{group}}-{{ channel.name }}.xml" title="RSS" target="_blank" class="page__icon page__icon_rss page__icon_text page__icon_link"></a>
        </p>
        <p class="releases__info-text">{{ channel.description[page.lang] }}</p>
      </div>

      {%- assign group_history = site.data.releases_history.history | reverse | where: "group", group %}
      {%- assign channel_history = group_history | where: "name", channel.name %}
  
      {%- if channel_history.size > 0 %}
        {%- for channel_action in channel_history %}
           {%- assign release = site.data.releases.releases | where: "tag_name", channel_action.version | first %}            
            <div class="releases__header">
                <a href="{{ release.html_url }}" class="releases__title">{{ release.tag_name }}</a>
                <div class="releases__date">{{ channel_action.ts | date: "%b %-d, %Y at %H:%M %z" }}</div>
            </div>
            <div class="releases__body">
                {{ release.body | markdownify }}
            </div>
        {%- endfor %}
      {%- else %}
        <div class="releases__info releases__info_notification">
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
      <div class="releases__info">
          <p>This is a list of all of the releases (Alpha, Beta, Early-Access, Stable and Rock-Solid) combined in chronological order.</p>
      </div>
      {%- assign group_history = site.data.releases_history.history | reverse | where: "group", group | map: "version" | reverse | uniq %}
      {%- for release_data in group_history %}
          {%- assign release = site.data.releases.releases | where: "tag_name", release_data | first %}
          <div class="releases__header">
              <div class="releases__date">{{ channel_action.ts | date: "%b %-d, %Y at %H:%M %z" }}</div>
              <a href="{{ release.html_url }}" class="releases__title">{{ release.tag_name }}</a>              
          </div>
          <div class="releases__body">
              {{ release.body | markdownify }}
          </div>
      {%- endfor %}
    </div>
    {%- endif %}
    {%- endcomment %}
</div>
{%- endfor %}

## Stability channels

All changes in werf go through all stability channels:

- `alpha` channel can bring new features but can be unstable;
- `beta` channel is for more broad testing of new features to catch regressions;
- `ea` channel is mostly safe and can be used in non-critical environments or for local development;
- `stable` channel is mostly safe and we encourage you to use this version everywhere.
  We **guarantee** that `ea` release should become `stable` not earlier than 1 week after internal tests;
- `rock-solid` channel is a generally available version and recommended for use in critical environments with tight SLAs.
  We **guarantee** that `stable` release should become a `rock-solid` release not earlier than after 2 weeks of extensive testing.

The relations between channels and werf releases are described in [multiwerf.json](https://github.com/werf/werf/blob/multiwerf/multiwerf.json). The usage of werf within the channel should be carried out with [multiwerf](https://github.com/werf/multiwerf). 

> When using release channels, you do not specify a version, because the version is managed automatically within the channel
  
Stability channels and frequent releases allow receiving continuous feedback on new changes, quickly rolling problem changes back, ensuring the high stability of the software, and preserving an acceptable development speed at the same time.

### Backward compatibility promise

> _Note:_ This promise was introduced with werf 1.0 and does not apply to previous versions.

werf follows a versioning strategy called [Semantic Versioning](https://semver.org). It means that major releases (1.0, 2.0) can break backward compatibility. In the case of werf, an update to the next major release _may_ require to do a full re-deploy of applications or to perform other non-scriptable actions.

Minor releases (1.1, 1.2, etc.) may introduce new global features, but have to do so without significant backward compatibility breaks with a major branch (1.x).
In the case of werf, this means that an update to the next minor release goes smoothly most of the time. However, it _may_ require running a provided upgrade script.

Patch releases (1.1.0, 1.1.1, 1.1.2) may introduce new features, but must do so without breaking backward compatibility within the minor branch (1.1.x).
In the case of werf, this means that an update to the next patch release should be smooth and can be done automatically.

- We do **not guarantee** backward compatibility between:
  - `alpha` releases;
  - `beta` releases;
  - `ea` releases.
- We **guarantee** backward compatibility between:
  - `stable` releases within the minor branch (1.1.x);
  - `rock-solid` releases within the minor branch (1.1.x).
