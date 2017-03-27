---
title: Поддержка git (#TODO)
sidebar: doc_sidebar
permalink: git_for_build.html
folder: build
---
Требуется собрать образ с зашитым исходным кодом приложения из git.

Первая проблема, с которой предстоит столкнуться: как добиться минимальной задержки между коммитом и получением готового образа для дальнейшего тестирования/деплоя. Большая часть коммитов в репозиторий приложения относится к обновлению кода самого самого приложения. В этом случае сборка нового образа должна представлять собой не более, чем применение патча к файлам предыдущего образа.

Вторая проблема возникает, когда подготовка образа включает в себя, например, установку внешних зависимостей gem'ов в образ на основе Gemfile и Gemfile.lock из git-репозитория. В таком случае необходимо пересобирать стадию, на которой происходит установка этих зависимостей, в случае изменения Gemfile или Gemfile.lock.

* Problem: i want to copy only particular files from my repo to image. Or i want to exclude some directories.
* your app is in git repository. .dockerignore
* directives: git
* Enhanced layers cache for faster builds
  * Problems:
    * cache depends only on Dockerfile content, not on my changes of source files.
    * cache is local.
    * Stages: what is a checksum? what is a sequence of stages/lifecycle of build process (the order of stage execution)? 
  * directives: git.add.stage_dependencies
