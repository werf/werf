---
title: Releases
permalink: releases.html
sidebar: documentation
layout: default
---

<link rel="stylesheet" href="{{ site.baseurl }}/css/releases.css">

{% assign releases = site.data.releases.releases %}
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