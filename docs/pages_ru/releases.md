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
    История изменений версий в каналах обновлений
</div>

<div class="releases">

<div class="releases__block-subtitle">
    Релиз:
</div>

<div class="tabs">
    {%- for group in groups %}
    <a href="javascript:void(0)" class="tabs__btn tabs__group__btn{% if group == groups[0] %} active{% endif %}" onclick="openTab(event, 'tabs__group__btn', 'tabs__group__content', 'group-{{group}}')">{{group}}</a>
    {%- endfor %}
</div>

{%- for group in groups %}
<div id="group-{{group}}" class="tabs__content tabs__group__content{% if group == groups[0] %} active{% endif %}">
    <div class="releases__block-subtitle">
        Канал:
    </div>
    <div class="tabs">
      {%- assign not_activated = true %}
      {%- assign active_channels = 0 %}
      {%- for channel in channels_sorted_reverse %}
        {%- assign channel_activity = site.data.releases_history.history | reverse | where: "group", group | where: "name", channel.name | size %}
        {%- if channel_activity < 1 %}
          {% continue %} 
        {%- endif %}
        <a href="javascript:void(0)" class="tabs__btn tabs__{{group}}__channel__btn{% if channel_activity > 0 and not_activated and channel != channels_sorted_reverse[0] %} active{% endif %}" onclick="openTab(event, 'tabs__{{group}}__channel__btn', 'tabs__{{group}}__channel__content', 'id-{{group}}-{{channel.name}}')">{{channel.title}}</a>
        {%- if channel_activity > 0 and not_activated and channel != channels_sorted_reverse[0] %}
        {%- assign not_activated = false %}
        {%- endif %}
        {%- assign active_channels = active_channels | plus: 1 %}
      {%- endfor %}
      {%- if active_channels > 10 %}
         <a href="javascript:void(0)" class="tabs__btn tabs__{{group}}__channel__btn" onclick="openTab(event, 'tabs__{{group}}__channel__btn', 'tabs__{{group}}__channel__content', 'id-{{group}}-all')">Все каналы</a>
      {%- endif %}
    </div>

    {%- assign not_activated = true %}
    {%- assign active_channels = 0 %}
    {%- for channel in channels_sorted_reverse %}
    {%- assign channel_activity = site.data.releases_history.history | reverse | where: "group", group | where: "name", channel.name | size %}
    {%- if channel_activity < 1 %}
      {% continue %} 
    {% endif %}
    <div id="id-{{group}}-{{ channel.name }}" class="tabs__content tabs__{{group}}__channel__content{% if channel_activity > 0 and not_activated and channel != channels_sorted_reverse[0] %} active{% endif %}">
      <div class="releases__info">
        <p>{{ channel.tooltip[page.lang] }}</p>
        <p class="releases__info-text">{{ channel.description[page.lang] }}</p>
      </div>
  
      {%- assign group_history = site.data.releases_history.history | reverse | where: "group", group %}
      {%- assign channel_history = group_history | where: "name", channel.name %}
  
      {%- if channel_history.size > 0 %}
        {%- for channel_action in channel_history %}
           {%- assign release = site.data.releases.releases | where: "tag_name", channel_action.version | first %}
            <div class="releases__header">
                <a href="{{ release.html_url }}" class="releases__title">{{ release.tag_name }}</a>
                <div class="releases__date">{{ channel_action.ts | date: "%e.%m.%Y, %H:%M %z" }}</div>
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
    {%- if channel_activity > 0 and not_activated and channel != channels_sorted_reverse[0] %}
      {%- assign not_activated = false %}
    {%- endif %}
    {%- assign active_channels = active_channels | plus: 1 %}

    {%- endfor %}

    {%- comment %}
    {%- if active_channels > 10 %}
    <div id="id-{{group}}-all" class="tabs__content tabs__{{group}}__channel__content">
      <div class="releases__info">
          <p>Список всех релизов (Alpha, Beta, Early-Access, Stable и Rock-Solid) в хронологическом порядке.</p>
      </div>
      {%- assign group_history = site.data.releases_history.history | reverse | where: "group", group | map: "version" | reverse | uniq %}
      {%- for release_data in group_history %}
          {%- assign release = site.data.releases.releases | where: "tag_name", release_data | first %}
          <div class="releases__header">
              <a href="{{ release.html_url }}" class="releases__title">{{ release.tag_name }}</a>
              <div class="releases__date">{{ channel_action.ts | date: "%e.%m.%Y, %H:%M %z" }}</div>
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
