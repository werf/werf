---
title: Debug werf.yaml
permalink: /usage/project_configuration/debug_werf_yaml.html
---

## Обзор

Файл `werf.yaml` может быть параметризован с помощью Go templates, что делает его мощным, но сложным в отладке. Werf предоставляет несколько команд для просмотра конечной конфигурации проекта и построения зависимостей: `render`, `list`, `graph`.

Пример, который будет использоваться

{% include pages/ru/werf_yaml_template_example.md.liquid %}

## `werf config render`

Показывает итоговую конфигурацию `werf.yaml` после рендеринга шаблонов.

### Пример команды

```bash
werf config render
```
### Пример вывода:

```yaml
project: my-project
configVersion: 1
---

image: rails
from: alpine
ansible:
  beforeInstall:
  - name: "Install mysql client"
    apt:
      name: "{{ item }}"
      update_cache: yes
    with_items:
    - libmysqlclient-dev
    - mysql-client
    - g++
  - command: gpg --keyserver hkp://keys.gnupg.net --recv-keys 409B6B1796C275462A1703113804BB82D39DC0E3
  - get_url:
      url: https://raw.githubusercontent.com/rvm/rvm/master/binscripts/rvm-installer
      dest: /tmp/rvm-installer
  - name: "Install rvm"
    command: bash -e /tmp/rvm-installer
  - name: "Install ruby 2.3.4"
    raw: bash -lec {{ item | quote }}
    with_items:
    - rvm install 2.3.4
    - rvm use --default 2.3.4
    - gem install bundler --no-ri --no-rdoc
    - rvm cleanup all
  install:
  - file:
      path: /root/.ssh
      state: directory
      owner: root
      group: root
      recurse: yes
  - name: "Setup ssh known_hosts used in Gemfile"
    shell: |
      set -e
      ssh-keyscan github.com >> /root/.ssh/known_hosts
      ssh-keyscan mygitlab.myorg.com >> /root/.ssh/known_hosts
    args:
      executable: /bin/bash
  - name: "Install Gemfile dependencies with bundler"
    shell: |
      set -e
      source /etc/profile.d/rvm.sh
      cd /app
      bundle install --without development test --path vendor/bundle
    args:
      executable: /bin/bas
```

### Когда полезно

* Проверка, что шаблоны Go корректно отрабатывают;
* Удостовериться, что значения переменных заданы как ожидалось;
* Отладка `dependencies` и других вложенных структур.

## `werf config list`

Выводит список всех образов, определённых в итоговом `werf.yaml`.

### Пример команды

```bash
werf config list
```

### Пример вывода:

```
rails
```

### Когда полезно

* Быстрая проверка всех образов, которые будут собираться;
* Диагностика генерации имён при использовании шаблонов.

## `werf config graph`

Строит граф зависимостей между образами.

### Пример команды

```bash
werf config graph
```

### Пример вывода:

```yaml
- image: rails
```

### Когда полезно

* Проверка, что зависимости подключены корректно;
* Отладка сложных схем сборки с `dependencies`.

## debug-templates flag

Включает режим отладки шаблонов, который позволяет просматривать отрендеренные шаблоны в файле werf.yaml (по умолчанию — значение переменной окружения $WERF_DEBUG_TEMPLATES или false).

### Пример команды

```bash
werf config render
```