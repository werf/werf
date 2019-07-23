---
title: Releases
permalink: releases.html
sidebar: documentation
layout: default
---

<link rel="stylesheet" href="{{ site.baseurl }}/css/publications.css">

{% assign releases = site.data.releases.releases %}

{% for release in releases %}

<h1><a href="{{ release.html_url }}">{{ release.name }}</a></h1>
{{ release.body | markdownify }}
<hr>

{% endfor %}