---
title: Command Line Interface
sidebar: cli
permalink: cli/
---

```bash
Main Commands:
  build             Build stages
  publish           Build final images from stages and push into images repo
  build-and-publish Build stages and publish images
  run               Run container for specified project image
  deploy            Deploy application into Kubernetes
  dismiss           Delete application from Kubernetes
  cleanup           Safely cleanup unused project images and stages
  purge             Purge all project images from images repo and stages from 
                    stages storage

Toolbox:
  slugify           Print slugged string by specified format
  ci-env            Generate werf environment variables for specified CI system

Lowlevel Management Commands:
  stages            Work with stages, which are cache for images
  images            Work with images
  helm              Manage application deployment with helm
  host              Work with werf cache and data of all projects on the host 
                    machine

Other Commands:
  completion        Generate bash completion scripts
  version           Print version
```