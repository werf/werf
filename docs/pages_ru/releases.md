---
title: Releases
permalink: releases.html
sidebar: documentation
layout: default
---

{%- asset releases.css %}

{%- assign releases = site.data.releases.releases %}

<div class="page__container page_releases">

<div class="releases__block-title">
    Каналы обновлений
</div>

<!-- Releases description -->
<div class="releases__info">
    Релизный процесс подразумевает последовательное прохождение версий по каналам в порядке повышения стабильности Alpha → Beta → Early-Access → Stable → Rock-Solid. Таким образом каждую версию на менее стабильном канале можно рассматривать как кандидата на переход в более стабильный канал.
</div>

{%- assign groups = site.data.releases_history.history | map: "group" | uniq | reverse %}
{%- assign channels_sorted = site.data.channels_info.channels | sort: "stability" %}
{%- assign channels_sorted_reverse = site.data.channels_info.channels | sort: "stability" | reverse  %}

<div class="releases__menu">
{%- for channel in channels_sorted_reverse %}
{%- assign channel_latest_versions = site.data.releases_history.latest | where: "name",  channel.name | first| map: "versions" | first| default: nil %}
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
    Релизы
</div>

<div class="releases__block-subtitle">
    Канал обновлений:
</div>

<div class="tabs">
  {%- for channel in channels_sorted_reverse %}
  <a href="javascript:void(0)" class="tabs__btn tabs__channel__btn{% if channel == channels_sorted_reverse[0] %} active{% endif %}" onclick="openTab(event, 'tabs__channel__btn', 'tabs__channel__content', 'id-{{channel.name}}')">{{channel.title}}</a>
  {%- endfor %}
  <a href="javascript:void(0)" class="tabs__btn tabs__channel__btn" onclick="openTab(event, 'tabs__channel__btn', 'tabs__channel__content', 'id-all-channels')">Все каналы</a>
</div>

{%- for channel in channels_sorted_reverse %}
<div id="id-{{channel.name}}" class="tabs__channel__content{% if channel == channels_sorted_reverse[0] %} active{% endif %}">
    <div class="releases__block-subtitle">
        Версия:
    </div>
    <div class="tabs">
    {%- assign not_activated = true %}
    {%- for group in groups %}
      {%- assign group_activity = site.data.releases_history.history | reverse | where: "group", group | where: "name", channel.name | size %}
      {%- if group_activity < 1 %}
        {% continue %} 
      {% endif %}
      <a href="javascript:void(0)" class="tabs__btn tabs__{{channel.name}}__btn{%- if group_activity > 0 and not_activated %} active{% endif %}" 
         onclick="openTab(event, 'tabs__{{channel.name}}__btn', 'tabs__{{channel.name}}__content', 'id-{{group}}-{{ channel.name }}')">{{group}}</a>
         {%- if group_activity > 0 and not_activated %}
         {%- assign not_activated = false %}
         {%- endif %}
    {%- endfor %}
    </div>

    <div class="releases__info">
        <p>{{ channel.tooltip[page.lang] }}</p>
        <p class="releases__info-text">{{ channel.description[page.lang] }}</p>
    </div>
    {%- assign not_activated = true %}
    {%- for group in groups %}
      {%- assign group_activity = site.data.releases_history.history | reverse | where: "group", group | where: "name", channel.name | size %}
      {%- if group_activity < 1 %}
        {% continue %} 
      {% endif %}
      <div id="id-{{group}}-{{ channel.name }}" class="tabs__content tabs__{{channel.name}}__content{%- if group_activity > 0 and not_activated %} active{% endif %}">
        
        {%- assign group_history = site.data.releases_history.history | reverse | where: "group", group %}
        {%- assign channel_history = group_history | where: "name", channel.name %}
        
        {%- if channel_history.size > 0 %}
            {%- for channel_action in channel_history %}
               {%- assign release = site.data.releases.releases | where: "tag_name", channel_action.version | first %}
                <div class="releases__title">
                    <a href="{{ release.html_url }}">
                        {{ release.tag_name }}
                    </a>
                </div>
                <div class="releases__body">
                    {{ release.body | markdownify }}
                </div>
            {%- endfor %}
        {%- else %}
            <div class="releases__info releases__info_notification">
            <p>На канале пока нет версий, но обязательно скоро появятся.</p>
        </div>
    {%- endif %}

      </div>
      {%- if group_activity > 0 and not_activated %}
      {%- assign not_activated = false %}
      {%- endif %}

    {%- endfor %}
</div>
{%- endfor %}

<div id="id-all-channels" class="tabs__content tabs__channel__content">
    <div class="releases__block-subtitle">
        Версия:
    </div>
    <div class="tabs">
    {%- for group in groups %}
    {%- assign group_activity = site.data.releases_history.history | reverse | where: "group", group | where: "name", channel.name | size %}
    <a href="javascript:void(0)" class="tabs__btn tabs__all-channel__btn{% if group == groups[0] %} active{% endif %}
             {%- if group_activity < 1 %} tabs__btn__empty{% endif %}" 
             onclick="openTab(event, 'tabs__all-channel__btn', 'tabs__all-channel__content', 'id-{{group}}-all-channel')">{{group}}</a>
    {%- endfor %}
    </div>

    <div class="releases__info">
            <p>Список всех версий на каналах в хронологическом порядке.</p>
    </div>

    {%- for group in groups %}
      <div id="id-{{group}}-all-channel" class="tabs__content tabs__all-channel__content{% if group == groups[0] %} active{% endif %}">

      {%- assign group_history = site.data.releases_history.history | reverse | where: "group", group %}

      {%- for release_data in group_history %}
              {%- assign release = site.data.releases.releases | where: "tag_name", release_data.version | first %}
              <div class="releases__title">
                  <a href="{{ release.html_url }}">
                      {{ release.tag_name }}
                  </a>
              </div>
              <div class="releases__body">
                  {{ release.body | markdownify }}
              </div>
      {%- endfor %}
      </div>
    {%- endfor %}
</div>
