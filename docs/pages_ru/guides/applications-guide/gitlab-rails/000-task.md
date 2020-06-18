---
title: Как использовать гайд
sidebar: applications-guide
permalink: documentation/guides/applications-guide/gitlab-rails/000-task.html
author: alexey.chazov <alexey.chazov@flant.com>
layout: guide
toc: false
author_team: "bravo"
author_name: "alexey.chazov"
ci: "gitlab"
language: "ruby"
framework: "rails"
is_compiled: 0
package_managers_possible:
 - bundler
package_managers_chosen: "bundler"
unit_tests_possible:
 - Rspec
unit_tests_chosen: "Rspec"
assets_generator_possible:
 - webpack
 - gulp
assets_generator_chosen: "webpack"
---

Этот гайд расскажет, как Ruby On Rails разработчику развернуть своё приложение в Kubernetes с помощью утилиты Werf.

Обязательны к прочтению главы "Подготовка к работе" и "Базовые настройки" — в них будут разобраны вопросы настройки окружения и основы работы с Werf, сборки и деплоя приложения в production. Однако, чтобы построить серьёзное приложение понадобится чуть больше навыков, раскрытых в других главах.

В рамках гайда мы будем рассматривать приложение с минимальным функционалом, которое постепенно будем дорабатывать. Исходный код всех вариантов приложения также прилагается.

<div>
    <a href="#" class="nav-btn">Далее: подготовка к работе</a>
</div>
