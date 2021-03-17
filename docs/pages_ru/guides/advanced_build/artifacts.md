---
title: Использование артефактов
sidebar: documentation
permalink: guides/advanced_build/artifacts.html
author: Artem Kladov <artem.kladov@flant.com>
---

## Обзор задачи

Часто в процессе сборки образа приложения происходит скачивание временных файлов — пакетов установки, архивов программ и т.п. В результате в образе могут оставаться файлы, которые необходимы в процессе сборки, но уже не требуется конечному пользователю собранного образа для запуска приложения.

werf может [импортировать]({{ site.baseurl }}/configuration/stapel_image/import_directive.html) ресурсы из других образов и образов [артефактов]({{ site.baseurl }}/configuration/stapel_artifact.html). Это позволяет вынести часть процесса сборки (сборку вспомогательных инструментов) в отдельный образ, копируя в конечный образ приложения только необходимые файлы. Этот функционал werf похож на [соответствующий в Docker](https://docs.docker.com/develop/develop-images/multistage-build/) (поддерживаемый, начиная с Docker версии 17.05), но в werf имеется больше возможностей, в частности, по импорту файлов.

В руководстве сначала рассматривается сборка тестового приложения на GO, а затем оно оптимизируется для существенного уменьшения размера конечного образа с использованием артефактов.

## Требования

* Установленные [зависимости werf]({{ site.baseurl }}/guides/installation.html#установка-зависимостей).
* Установленный [multiwerf](https://github.com/werf/multiwerf).

### Выбор версии werf

Перед началом работы необходимо выбрать версию werf. Для выбора актуальной версии werf в канале stable, релиза 1.1, выполним следующую команду:

```shell
. $(multiwerf use 1.1 stable --as-file)
```

## Тестовое приложение

Возьмем в качестве примера приложение [Go Web App](https://github.com/josephspurrier/gowebapp), написанное на [Go](https://golang.org/).

### Сборка

Создадим папку `gowebapp` и файл `werf.yaml` со следующим содержимым:
{% raw %}
```yaml
project: gowebapp
configVersion: 1
---

image: gowebapp
from: golang:1.14
docker:
  WORKDIR: /app
ansible:
  install:
  - name: Getting packages
    shell: go get github.com/josephspurrier/gowebapp
  setup:
  - file:
      path: /app
      state: directory
  - name: Copying config
    shell: |
      cp -r $GOPATH/src/github.com/josephspurrier/gowebapp/config /app/config
      cp -r $GOPATH/src/github.com/josephspurrier/gowebapp/static /app/static
      cp -r $GOPATH/src/github.com/josephspurrier/gowebapp/template /app/template
      cp $GOPATH/bin/gowebapp /app/
```
{% endraw %}

Приведенные инструкции описывают сборку одного образа — `gowebapp`.

Соберём образ приложения, выполнив следующую команду в папке `gowebapp`:

```shell
werf build --stages-storage :local
```

### Запуск

Запустим приложение, выполнив следующую команду в папке `gowebapp`:
```shell
werf run --stages-storage :local --docker-options="-d -p 9000:80 --name gowebapp"  gowebapp -- /app/gowebapp
```

Убедитесь, что контейнер запустился, выполнив следующую команду:
```shell
docker ps -f "name=gowebapp"
```

Вы должны увидеть запущенный контейнер `gowebapp`, например, вывод может быть подобен следующему:
```shell
CONTAINER ID  IMAGE                                          COMMAND           CREATED        STATUS        PORTS                  NAMES
41d6f49798a8  werf-stages-storage/gowebapp:84d7...44992265   "/app/gowebapp"   2 minutes ago  Up 2 minutes  0.0.0.0:9000->80/tcp   gowebapp
```

Откройте в браузере адрес [http://localhost:9000](http://localhost:9000) — вы должны увидеть страницу `Go Web App`. Перейдите по ссылке "Click here to login", где вы сможете зарегистрироваться и авторизоваться в приложении.

### Размер собранного образа

Получим размер собранного образа, выполнив:

{% raw %}
```shell
docker images `docker ps -f "name=gowebapp" --format='{{.Image}}'`
```
{% endraw %}

Пример вывода:
```shell
REPOSITORY                 TAG                   IMAGE ID          CREATED             SIZE
werf-stages-storage/gowebapp   84d7...44992265   07cdc430e1c8      10 minutes ago      857MB
```

Обратите внимание, что размер образа приложения получился **более 800 Мегабайт**.

## Оптимизация сборки приложения с использованием артефактов

Непосредственно для запуска приложения необходимы только файлы в папке `/app`, поэтому из образа можно удалить скачанные пакеты и сам компилятор Go. Использование функционала [артефактов в werf]({{ site.baseurl }}/configuration/stapel_artifact.html) позволяет импортировать в образ только конкретные файлы.

### Сборка

Заменим имеющийся файл `werf.yaml` следующим содержимым:
{% raw %}
```yaml
project: gowebapp
configVersion: 1
---

artifact: gowebapp-build
from: golang:1.14
ansible:
  install:
  - name: Getting packages
    shell: go get github.com/josephspurrier/gowebapp
  setup:
  - file:
      path: /app
      state: directory
  - name: Copying config
    shell: |
      cp -r $GOPATH/src/github.com/josephspurrier/gowebapp/config /app/config
      cp -r $GOPATH/src/github.com/josephspurrier/gowebapp/static /app/static
      cp -r $GOPATH/src/github.com/josephspurrier/gowebapp/template /app/template
      cp $GOPATH/bin/gowebapp /app/
---
image: gowebapp
docker:
  WORKDIR: /app
from: ubuntu:18.04
import:
- artifact: gowebapp-build
  add: /app
  to: /app
  after: install
```
{% endraw %}

В оптимизированных инструкциях сборки само приложение собирается в артефакте `gowebapp-build`, после чего получившиеся файлы импортируются в образ `gowebapp`.

Обратите внимание, что при сборке образа `gowebapp` используется образ `ubuntu`, а не `golang`.

Соберём приложение с измененным файлом инструкций:
```shell
werf build --stages-storage :local
```

### Запуск

Перед запуском измененного приложения нужно остановить и удалить запущенный контейнер `gowebapp`, собранный и запущенный ранее. В противном случае новый контейнер не сможет запуститься из-за того, что контейнер с таким именем уже существует, порт 9000 занят. Например, выполните следующие команды для остановки и удаления контейнера `gowebapp`:

```shell
docker rm -f gowebapp
```

Запустим измененное приложение, выполнив следующую команду:
```shell
werf run --stages-storage :local --docker-options="-d -p 9000:80 --name gowebapp" gowebapp -- /app/gowebapp
```

Убедитесь, что контейнер запустился, выполнив следующую команду:
```shell
docker ps -f "name=gowebapp"
```

Вы должны увидеть запущенный контейнер `gowebapp`, например, вывод может быть следующим:
```shell
CONTAINER ID  IMAGE                                          COMMAND          CREATED        STATUS        PORTS                   NAMES
41d6f49798a8  werf-stages-storage/gowebapp:84d7...44992265   "/app/gowebapp"  2 minutes ago  Up 2 minutes  0.0.0.0:9000->80/tcp   gowebapp
```

Откройте в браузере адрес [http://localhost:9000](http://localhost:9000) — вы должны увидеть страницу `Go Web App`. Перейдите по ссылке "Click here to login", где вы сможете зарегистрироваться и авторизоваться в приложении.

### Размер собранного образа

Получим размер образа, выполнив:
{% raw %}
```shell
docker images `docker ps -f "name=gowebapp" --format='{{.Image}}'`
```
{% endraw %}

Пример вывода:
```shell
REPOSITORY                     TAG               IMAGE ID       CREATED          SIZE
werf-stages-storage/gowebapp   84d7...44992265   07cdc430e1c8   10 minutes ago   79MB
```

Сравнивая размеры можно увидеть, что **образ, собранный с использованием артефактов, меньше на 90%**, чем первоначальный.

## Вывод

Приведенный пример показывает, что использование артефактов — отличный способ выбросить ненужное из конечного образа. Более того, вы можете использовать один и тот же артефакт (или артефакты) в нескольких образах, определенных в одном `werf.yaml`. Этот прием позволяет увеличить скорость сборки и сократить размер конечного образа.
