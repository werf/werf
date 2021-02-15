---
title: werf-giterminism.yaml
permalink: documentation/reference/werf_giterminism_yaml.html
sidebar: documentation
description: Werf giterminism config file example
toc: false
---

The `werf-giterminism.yaml` allows the use of certain uncommitted configuration files and enables features that potentially depend on external factors. To take effect, this file should be committed to the project git repository.

> We recommend minimizing the use of `werf-giterminism.yaml` to create a configuration that is robust and easily reproducible

All the paths and globs in the config must be defined relative to the project directory.

{% include documentation/reference/werf_giterminism_yaml/table.html %}
