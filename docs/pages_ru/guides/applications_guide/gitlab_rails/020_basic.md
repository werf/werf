---
title: Базовые настройки
sidebar: applications_guide
guide_code: gitlab_rails
permalink: documentation/guides/applications_guide/gitlab_rails/020_basic.html
layout: guide
toc: false
---

В этой главе мы возьмём приложение, которое будет выводить сообщение "hello world" по http и опубликуем его в kubernetes с помощью Werf. Сперва мы разберёмся со сборкой и добьёмся того, чтобы образ оказался в Registry, затем — разберёмся с деплоем собранного приложения в Kubernetes, и, наконец, организуем CI/CD-процесс силами Gitlab CI. 

Наше приложение будет состоять из одного docker образа собранного с помощью werf. В этом образе будет работать один основной процесс который запустит веб сервер для ruby. Управлять маршрутизацией запросов к приложению будет Ingress в Kubernetes кластере. Мы реализуем два стенда: [production](https://ru.werf.io/documentation/reference/ci_cd_workflows_overview.html#production) и [staging](https://ru.werf.io/documentation/reference/ci_cd_workflows_overview.html#staging).

<div>
    <a href="030_dependencies.html" class="nav-btn">Далее: Подключение зависимостей</a>
</div>
