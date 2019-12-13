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

<div class="page__container page_releases">

<div class="releases__block-title">
    Каналы обновлений
</div>

<!-- latest version per channel -->
{% assign channels_sorted = site.data.releases_info.channels | sort: "stability" %}
{% assign channels_sorted_reverse = site.data.releases_info.channels | sort: "stability" | reverse  %}

<div class="releases__menu">
{% for channel in channels_sorted_reverse %}
{% assign version = site.data.releases_info.versions | where: "name",  channel.name | map: "version" | first %}
    <div class="releases__menu-item">
        <div class="releases__menu-item-header">            
            <div class="releases__menu-item-title">
                {{ channel.title }}
            </div>
            {% if version != null %}
                <div class="releases__menu-item-version">
                    {{ version }}
                </div>
            {% endif %}
            <a href="https://github.com/flant/werf/releases/tag/{{ version }}" class="releases__btn">View on GitHub</a>
        </div>        
        <div class="releases__menu-item-description">
            {{ channel.description }}
        </div>
    </div>
  {% endfor %}
</div>

<!-- Releases description -->
<div class="releases__info">
    <div class="releases__info-note">
        <em>Note:</em> Настоящее обещание относится к werf, начиная с версии 1.0, и не относится к предыдущим версиям или версиям dapp.
    </div>

    <p>werf использует <a href="https://semver.org/lang/ru/">семантическое версионирование</a>. Это значит, что мажорные версии (1.0, 2.0) могут быть обратно не совместимыми между собой. В случае werf это означает, что обновление на следующую мажорную версию <em>может</em> потребовать полного передеплоя приложений, либо других ручных операций.</p>

    <p>Минорные версии (1.1, 1.2, etc) могут добавлять новые “значительные” изменения, но без существенных проблем обратной совместимости в пределах мажорной версии. В случае werf это означает, что обновление на следующую минорную версию в большинстве случаев будет беспроблемным, но <em>может</em> потребоваться запуск предоставленных скриптов миграции.</p>

    <p>Патч-версии (1.1.0, 1.1.1, 1.1.2) могут добавлять новые возможности, но без каких-либо проблем обратной совместимости в пределах минорной версии (1.1.x). В случае werf это означает, что обновление на следующий патч (следующую патч-версию) не должно вызывать проблем и требовать каких-либо ручных действий.</p>

    <p>Патч-версии делятся на каналы обновлений. Канал обновлений — это префикс пререлизной части версии (1.1.0-alpha.2, 1.1.0-beta.3, 1.1.0-ea.1). Версия без пререлизной части считается версией стабильного канала обновлений.</p>
</div>

<div class="releases__block-title">
    Релизы
</div>

<div class="releases">
    <div class="tabs">
    {% for channel in channels_sorted_reverse %}
    <a href="javascript:void(0)" class="tabs__btn{% if channel == channels_sorted_reverse[0] %} active{% endif %}" onclick="openTab(event, 'tabs__btn', 'tabs__content', '{{channel.name}}')">Канал {{channel.title}} </a>
    {% endfor %}
    <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'all')">Все каналы</a>
    </div>

    {% for channel in channels_sorted_reverse %}
    <div id="{{ channel.name }}" class="tabs__content{% if channel == channels_sorted_reverse[0] %} active{% endif %}">
    <div class="releases__info">
        <p>{{ channel.tooltip }}</p>
        <p>{{ channel.description }}</p>
    </div>

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
        <div class="releases__info">
            <p>На канале пока нет версий, но обязательно скоро появятся.</p>
        </div>
    {% endif %}

    </div>
    {% endfor %}

    {% assign _releases = site.data.releases.releases %}
    <div id="all" class="tabs__content">
    {% for release in _releases %}
        <div class="releases__info">
            <p>Все релизы в хронологическом порядке</p>
        </div>
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