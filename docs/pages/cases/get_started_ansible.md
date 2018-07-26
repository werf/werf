---
title: Первое приложение на dapp (Ansible)
sidebar: how_to
permalink: get_started_ansible.html
---

В этой главе описана сборка простого php-приложения [Symfony Demo APP](https://github.com/symfony/demo) с помощью dapp и [Ansible сборщика](build_yaml.html). Перед изучением dapp желательно представлять, что такое Dockerfile и его основные [директивы](https://docs.docker.io/).

## Определение шагов сборки приложения

Для для сборки образа приложения [Symfony Demo APP](https://github.com/symfony/demo), сформируем следующие требования к этапам сборки:
- необходимо установить системное ПО и системные зависимости;
    * установить `php`;
    * установить расширения `php7.0-sqlite3` (для приложения) и `php7.0-xml`, `php7.0-zip` (для composer);
- необходимо настроить системное ПО;
    * выделить для работы веб-сервера  отдельного пользователя - `phpapp`;
- необходимо установить прикладные зависимости;
    * для установки зависимостей проекта нужен composer, который можно установить скачиванием phar файла, - поэтому в системное ПО добавится `curl`;
- необходимо добавить код приложения;
    * код приложения будет располагаться в финальном образе в директории `/demo`;
    * всем файлам в папке `/demo` нужно будет установить владельца - пользователя `phpapp`;
- необходимо настроить приложение;
    * никаких особых настроек производить не нужно и единственной настройкой будет указание ip адреса, на котором слушает веб-сервер, - эта настройка будет в скрипте `/opt/start.sh`, который будет запускаться при старте контейнера;
    * в качестве иллюстрации для стадии setup добавится создание файла version.txt с текущей датой.

## Подготовка dappfile

Согласно требований к этапам сборки которые были определены ранее, подготовим dappfile с инструкциями для сборки приложения. При сборке будем использовать shell сборщик и dappfile с YAML синтаксисом.

Склонируйте репозиторий приложения [Symfony Demo APP](https://github.com/symfony/demo)

```
git clone https://github.com/symfony/symfony-demo.git
cd symfony-demo
```

Создайте dappfile.yaml следующего содержания:

{% raw %}
```
dimg: symfony-demo-app
from: ubuntu:16.04
docker:
  # Workdir
  WORKDIR: /app
  # Non-root main application user
  USER: app
  EXPOSE: '80'
  ENV:
    LC_ALL: en_US.UTF-8
ansible:
  beforeInstall:
    - name: "Generate en_US.UTF-8 default locale"
      locale_gen:
        name: en_US.UTF-8
        state: present
    #  установка вспомогательных пакетов, добавление репозитория
    - name: "Install additional packages"
      apt:
        name: "{{`{{ item }}`}}"
        state: present
        update_cache: yes
      with_items:
        - software-properties-common
        - locales
        - curl
    # Install PHP
    - name: "Add PHP apt repository"
      apt_repository:
        repo: 'ppa:ondrej/php'
        codename: 'xenial'
        update_cache: yes
    - name: "Install PHP"
      apt:
        name: "php7.2"
        state: present
        update_cache: yes
    # создание скрипта запуска /opt/start.sh (просто для демонстрации)
    - name: "Create start script"
      copy:
        content: |
          #!/bin/bash
          cd /app
          su -c "php bin/console server:run 0.0.0.0:8000" app
        dest: /opt/start.sh
    - file:
        path: /opt/start.sh
        owner: app
        group: app
        mode: 0755
  install:
      # установка необходимых для приложения модулей php
    - name: "Install php moduiles"
      apt:
        name: "{{`{{ item }}`}}"
        state: present
        update_cache: yes
      with_items:
        - php-sqlite3
        - php-xml
        - php-zip
        - php-mbstring
      # установка composer
    - raw: curl -LsS https://getcomposer.org/download/1.6.5/composer.phar -o /usr/local/bin/composer
    - file:
        path: /usr/local/bin/composer
        mode: "a+x"
  beforeSetup:
      # смена прав файлам исходных текстов и запуск composer install
    - file:
        path: /app
        state: directory
        owner: app
        group: app
        recurse: yes
    - raw: cd /app && su -c 'composer install' app
  setup:
      # используем текущую дату как версию приложения
    - raw: echo `date` > /app/version.txt
    - raw: chown app:app /app/version.txt
git:
  - add: '/'
    to: '/app'

```
{% endraw %}


## Сборка и запуск

Для сборки приложения выполните в корне проекта команду:

```
dapp dimg build
```

Запустите контейнер командой

```
dapp dimg run -d -p 8000:8000 -- /opt/start.sh
```

После чего проверить браузером или в консоли

```
curl host_ip:8000
```

## Что не так?

* Набор команд для создания файла start.sh вполне заменим на ещё одну директиву git и хранение файла в репозитории.
* Если директивой git можно копировать файлы, то почему бы в этой директиве не указать права на эти файлы?
* composer install требуется не каждый раз, а только при изменении файла package.json, поэтому было бы отлично, если эта команда запускалась только при изменении этого файла.

Эти проблемы будут более подробно рассмотрены в главе [Поддержка git](git.html)
