---
title: Поддержка git (#TODO)
sidebar: doc_sidebar
permalink: git_for_build.html
folder: build
---

Поддерживается chef-сборщиком и ansible-сборщиком


Требуется собрать образ с зашитым исходным кодом приложения из git.

Первая проблема, с которой предстоит столкнуться: как добиться минимальной задержки между коммитом и получением готового образа для дальнейшего тестирования/деплоя. Большая часть коммитов в репозиторий приложения относится к обновлению кода самого самого приложения. В этом случае сборка нового образа должна представлять собой не более, чем применение патча к файлам предыдущего образа.

Вторая проблема: если после сделанных изменений в исходном коде приложения и сборки образа эти изменения были отменены (например, через git revert) в следующем коммите — должен сработать кэш сборщика образов. Т.е. собираемые сборщиком образа должны кэшироваться по содержимому изменений в файлах git репозитория, а не по факту наличия этих изменений.

Третья проблема возникает, когда подготовка образа включает в себя, например, установку внешних зависимостей gem'ов в образ на основе Gemfile и Gemfile.lock из git-репозитория. В таком случае необходимо пересобирать стадию, на которой происходит установка этих зависимостей, если поменялся Gemfile или Gemfile.lock.

Четвертая проблема: добавлять файлы из git репозитория в образ путем копирования всего дерева исходников необходимо выполнять редко, основным способом добавления изменения должно служить наложение патчей.

* Problem: i want to copy only particular files from my repo to image. Or i want to exclude some directories.
* your app is in git repository. .dockerignore
* directives: git
* Enhanced layers cache for faster builds
  * Problems:
    * cache depends only on Dockerfile content, not on my changes of source files.
    * cache is local.
    * Stages: what is a checksum? what is a sequence of stages/lifecycle of build process (the order of stage execution)? 
  * directives: git.add.stage_dependencies
