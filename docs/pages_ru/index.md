---
title: GitOps CLI-утилита
permalink: /
layout: default
---

<div class="welcome">
    <div class="page__container">
        <div class="welcome__content">
            <h1 class="welcome__title">
                GitOps-утилита
            </h1>
            <div class="welcome__subtitle">
                 Выкатывайте приложения быстро и просто.<br/>Open Source. Написана на Go.
            </div>
            <!--
            <form action="https://www.google.com/search" class="welcome__search" method="get" name="searchform" target="_blank">
                <input name="sitesearch" type="hidden" value="ru.werf.io">
                <input autocomplete="on" class="page__input welcome__search-input" name="q" placeholder="Поиск по документации" required="required"  type="text">
                <button type="submit" class="page__icon page__icon_search welcome__search-btn"></button>
            </form>
            -->
            <div class="welcome__extra-content">
                <div class="welcome__extra-content-title">
                    CLI-утилита для использования в <span>пайплайнах CI/CD</span>
                </div>
                <div class="welcome__extra-content-text">
                    <ul class="intro__list">
                        <li>
                            werf интегрирует <code>git</code>, <code>Helm</code> и <code>Docker</code>.
                        </li>
                        <li>
                            Может быть встроена в любую CI/CD-систему (например, GitLab CI) <br/>для построения пайплайнов, используя предложенный набор команд:
                            <ul>
                                <li><code>werf build-and-publish</code>;</li>
                                <li><code>werf deploy</code>;</li>
                                <li><code>werf dismiss</code>;</li>
                                <li><code>werf cleanup</code>.</li>
                            </ul>
                        </li>
                        <li>
                            Open Source, написана на Go.
                        </li>
                        <li>
                            werf — это не SAAS, а представитель высокоуровневых <br/>CI/CD-инструментов нового поколения.
                        </li>
                    </ul>
                </div>
            </div>
        </div>
    </div>
</div>

<div class="page__container">
    <div class="intro">
        <div class="intro__image"></div>        
    </div>
</div>

<div class="page__container">
    <ul class="intro-extra">
        <li class="intro-extra__item">
            <div class="intro-extra__item-title">
                Удобный деплой
            </div>
            <div class="intro-extra__item-text">
                <ul class="intro__list">
                    <li>Полная совместимость с Helm.</li>
                    <li>Простое использование RBAC.</li>
                    <li>Выкаченное приложение в Kubernetes == готовое к использованию.</li>
                    <li>Обнаружение проблем и быстрое завершение проблемного выката.</li>
                    <li>Получение в режиме реального времени информации о процессе деплоя.</li>
                    <li>Настраиваемый детектор ошибок и готовности ресурсов Kubernetes с использованием их аннотаций.</li>
                </ul>
            </div>
        </li>
        <li class="intro-extra__item">
            <div class="intro-extra__item-title">
                Управление всем жизненным циклом образа
            </div>
            <div class="intro-extra__item-text">
                <ul class="intro__list">
                    <li>Сборка образов с Dockerfile либо с нашим синтаксисом, учитывая особенности инкрементальной сборкой (основанной на истории Git), используя Ansible и многие другие особенности сборщика werf.</li>
                    <li>Публикация образов в Docker registry, используя множество различных схем тегирования.</li>
                    <li>Выкат приложения в Kubernetes.</li>
                    <li>Очистка Docker registry, основанная на встроенных политиках и используемых в Kubernetes-кластерах образах приложения.</li>
                </ul>
            </div>
        </li>
    </ul>
</div>

<div class="stats">
    <div class="page__container">
        <div class="stats__content">
            <div class="stats__title">Активная разработка</div>
            <ul class="stats__list">
                <li class="stats__list-item">
                    <div class="stats__list-item-num">4</div>
                    <div class="stats__list-item-title">релиза в неделю</div>
                    <div class="stats__list-item-subtitle">в среднем за прошлый год</div>
                </li>
                <li class="stats__list-item">
                    <div class="stats__list-item-num">1400</div>
                    <div class="stats__list-item-title">инсталляций</div>
                    <div class="stats__list-item-subtitle">в больших и маленьких проектах</div>
                </li>
                <li class="stats__list-item">
                    <div class="stats__list-item-num gh_counter">1470</div>
                    <div class="stats__list-item-title">звезд на GitHub</div>
                    <div class="stats__list-item-subtitle">поддержите проект ;)</div>
                </li>
            </ul>
        </div>
    </div>
</div>

<div class="features">
    <div class="page__container">
        <ul class="features__list">
            <li class="features__list-item">
                <div class="features__list-item-icon features__list-item-icon_lifecycle"></div>
                <div class="features__list-item-title">Управление полным жизненным циклом приложения</div>
                <div class="features__list-item-text">Управляйте процессом сборки, выкатом приложения в Kubernetes и очисткой неиспользуемых образов.</div>
            </li>
            <li class="features__list-item">
                <div class="features__list-item-icon features__list-item-icon_kubernetes"></div>
                <div class="features__list-item-title">Удобный деплой в <span>Kubernetes</span></div>
                <div class="features__list-item-text">Выкатывайте приложение в Kubernetes, используя стандартный менеджер пакетов с интерактивным отслеживанием процесса и получением событий и логов в режиме реального времени.</div>
            </li>
            <li class="features__list-item">
                <div class="features__list-item-icon features__list-item-icon_easy"></div>
                <div class="features__list-item-title">Легко начать</div>
                <div class="features__list-item-text">Начните использовать werf с существующим Dockerfile.</div>
            </li>
            <li class="features__list-item">
                <div class="features__list-item-icon features__list-item-icon_size"></div>
                <div class="features__list-item-title">Уменьшение размера образа</div>
                <div class="features__list-item-text">Сократите размер, исключив исходный код, инструменты сборки и кэши с помощью артефактов и монтирования.</div>
            </li>
            <li class="features__list-item">
                <div class="features__list-item-icon features__list-item-icon_ansible"></div>
                <div class="features__list-item-title">Сборка образов с <span>Ansible</span></div>
                <div class="features__list-item-text">Используйте популярный и мощный IaaS-инструмент.</div>
            </li>
            <li class="features__list-item">
                <div class="features__list-item-icon features__list-item-icon_debug"></div>
                <div class="features__list-item-title">Инструменты отладки сборочного процесса</div>
                <div class="features__list-item-text">Получайте доступ к любой стадии во время сборки с помощью опций интроспекции.</div>
            </li>
            <li class="features__list-item"></li>
            <li class="features__list-item">
                <div class="features__list-item-icon features__list-item-icon_config"></div>
                <div class="features__list-item-title">Компактный файл конфигурации</div>
                <div class="features__list-item-text">Собирайте несколько образов, используя один файл конфигурации, повторно используйте общие части с помощью Go-шаблонов.</div>
            </li>
            <li class="features__list-item"></li>
        </ul>        
    </div>
</div>

<div class="community">
    <div class="page__container">
        <div class="community__content">
            <div class="community__title">Растущее дружелюбное сообщество</div>
            <div class="community__subtitle">Мы всегда на связи с сообществом<br/> в Telegram, Twitter и Slack.</div>
            <div class="community__btns">
                <a href="{{ site.social_links[page.lang].telegram }}" target="_blank" class="page__btn page__btn_w community__btn">
                    <span class="page__icon page__icon_telegram"></span>
                    Мы в Telegram
                </a>
                <a href="{{ site.social_links[page.lang].twitter }}" target="_blank" class="page__btn page__btn_w community__btn">
                    <span class="page__icon page__icon_twitter"></span>
                    Мы в Twitter
                </a>
                <a href="#" data-open-popup="slack" target="_blank" class="page__btn page__btn_w community__btn">
                    <span class="page__icon page__icon_slack"></span>
                    Мы в Slack
                </a>
            </div>
        </div>
    </div>
</div>

<div class="roadmap">
    <div class="page__container">
        <div class="roadmap__title">
            Дорожная карта
        </div>
        <div class="roadmap__content">
            <div class="roadmap__goals">
                <div class="roadmap__goals-content">
                    <div class="roadmap__goals-title">Цели</div>
                    <ul class="roadmap__goals-list">
                        <li class="roadmap__goals-list-item">
                            Полнофункциональная версия werf, хорошо работающая на единственном хосте при выполнении всех операций werf (сборка, деплой и очистка).
                        </li>
                        <li class="roadmap__goals-list-item">
                            Проверенные подходы и готовые решения<br/>
                            для работы с популярными CI-системами.
                        </li>
                        <li class="roadmap__goals-list-item">
                            Развитие сборщика. Сборка образов без привязки к локальному Docker-демону и сборка в кластере Kubernetes.
                        </li>
                    </ul>
                </div>
            </div>
            <div class="roadmap__steps">
                <div class="roadmap__steps-content">
                    <div class="roadmap__steps-title">Этапы</div>
                    <ul class="roadmap__steps-list">
                        <li class="roadmap__steps-list-item" data-roadmap-step="1616">
                            <a href="https://github.com/flant/werf/issues/1616" class="roadmap__steps-list-item-issue" target="_blank">#1616</a>
                            <span class="roadmap__steps-list-item-text">
                                Использование <a href="https://kubernetes.io/docs/tasks/manage-kubernetes-objects/declarative-config/#merge-patch-calculation" target="_blank">3-х стороннего слияния</a><br> при обновлении Helm-релизов.
                            </span>
                        </li>
                        <li class="roadmap__steps-list-item" data-roadmap-step="1940">
                            <a href="https://github.com/flant/werf/issues/1940" class="roadmap__steps-list-item-issue" target="_blank">#1940</a>
                            <span class="roadmap__steps-list-item-text">
                                Локальная разработка приложений с werf.
                            </span>
                        </li>
                        <li class="roadmap__steps-list-item" data-roadmap-step="1184">
                            <a href="https://github.com/flant/werf/issues/1184" class="roadmap__steps-list-item-issue" target="_blank">#1184</a>
                            <span class="roadmap__steps-list-item-text">
                                Тегирование, основанное на контенте.
                            </span>
                        </li>
                        <li class="roadmap__steps-list-item" data-roadmap-step="1617">
                            <a href="https://github.com/flant/werf/issues/1617" class="roadmap__steps-list-item-issue" target="_blank">#1617</a>
                            <span class="roadmap__steps-list-item-text">
                            Готовые рецепты для интеграции<br/>
                            с наиболее популярными CI-системами.
                            </span>
                        </li>
                        <li class="roadmap__steps-list-item" data-roadmap-step="1614">
                            <a href="https://github.com/flant/werf/issues/1614" class="roadmap__steps-list-item-issue" target="_blank">#1614</a>
                            <span class="roadmap__steps-list-item-text">
                                Распределенная сборка с общим Docker registry.
                            </span>
                        </li>
                        <li class="roadmap__steps-list-item" data-roadmap-step="1606">
                            <a href="https://github.com/flant/werf/issues/1606" class="roadmap__steps-list-item-issue" target="_blank">#1606</a>
                            <span class="roadmap__steps-list-item-text">
                                Поддержка Helm 3.
                            </span>
                        </li>
                        <li class="roadmap__steps-list-item" data-roadmap-step="1618">
                            <a href="https://github.com/flant/werf/issues/1618" class="roadmap__steps-list-item-issue" target="_blank">#1618</a>
                            <span class="roadmap__steps-list-item-text">
                                <a href="https://github.com/GoogleContainerTools/kaniko" target="_blank">Kaniko</a>-подобная сборка без привязки<br>к локальному Docker-демону.
                            </span>
                        </li>
                    </ul>
                </div>
            </div>
        </div>
    </div>
</div>

<div class="page__container">
    <div class="documentation">
        <div class="documentation__image">
        </div>
        <div class="documentation__info">
            <div class="documentation__info-title">
                Исчерпывающая документация
            </div>
            <div class="documentation__info-text">
                Документация содержит более 100 статей, включающих описание частых случаев (первые шаги, деплой в Kubernetes, интеграция с CI/CD-системами и другое), полное описание функций, архитектуры и CLI-команд.
            </div>
        </div>
        <div class="documentation__btns">
            <a href="https://github.com/flant/werf" target="_blank" class="page__btn page__btn_b documentation__btn">
                Начать использовать
            </a>
            <a href="{{ site.baseurl }}/documentation/guides/getting_started.html" class="page__btn page__btn_o documentation__btn">
                Руководства для старта
            </a>
            <a href="{{ site.baseurl }}/documentation/cli/main/build.html" class="page__btn page__btn_o documentation__btn">
                CLI-команды
            </a>
        </div>
    </div>
</div>
