---
title: Использование артефактов
sidebar: documentation
permalink: documentation/guides/advanced_build/artifacts.html
author: Artem Kladov <artem.kladov@flant.com>
---

## Обзор задачи

Часто в процессе сборки образа приложения происходит скачивание временных файлов — пакетов установки, архивов программ и т.п. В результате в образе могут оставаться файлы, которые необходимы в процессе сборки, но уже не требуется конечному пользователю собранного образа для запуска приложения.

werf может [импортировать]({{ site.baseurl }}/documentation/configuration/stapel_image/import_directive.html) ресурсы из других образов и образов [артефактов]({{ site.baseurl }}/documentation/configuration/stapel_artifact.html). Это позволяет вынести часть процесса сборки, сборку вспомогательных инструментов в отдельный образ, копируя в конечный образ приложения только необходимый файлы. Этот функционал werf похож на [соответствующий в Docker](https://docs.docker.com/develop/develop-images/multistage-build/) (поддерживаемый, начиная с Docker версии 17.05), но в werf имеется больше возможностей, в частности, по импорту файлов.

В руководстве сначала рассматривается сборка тестового приложения на GO, а затем она оптимизируются для существенного уменьшения размера конечного образа с использованием артефактов.

## Требования

* Установленные [зависимости werf]({{ site.baseurl }}/documentation/guides/installation.html#install-dependencies).
* Установленный [multiwerf](https://github.com/flant/multiwerf).

### Выбор версии werf

Перед началом работы необходимо выбрать версию werf. Для выбора актуальной версии werf в канале stable, релиза 1.0, выполним следующую команду:

```shell
. $(multiwerf use 1.0 stable --as-file)
```

## Тестовое приложение

Возьмем в качестве примера приложение [Hotel Booking](https://github.com/revel/examples/tree/master/booking), написанное на [Go](https://golang.org/) и с фреймворком [Revel Framework](https://github.com/revel).

### Сборка

Создадим папку `booking` и файл `werf.yaml` со следующим содержимым:
{% raw %}
```yaml
project: hotel-booking
configVersion: 1
---

image: go-booking
from: golang:1.10
ansible:
  beforeInstall:
  - name: Install additional packages
    apt:
      update_cache: yes
      pkg:
      - gcc
      - sqlite3
      - libsqlite3-dev
  install:
  - name: Getting packages
    shell: |
      go get -v github.com/revel/revel
      go get -v github.com/revel/cmd/revel
      (go get -v github.com/revel/examples/booking/... ; true )
  setup:
  - name: Preparing config and building application
    shell: |
      sed -i 's/^http.addr=$/http.addr=0.0.0.0/' $GOPATH/src/github.com/revel/examples/booking/conf/app.conf
      revel build --run-mode dev github.com/revel/examples/booking /app
```
{% endraw %}

Приведенные инструкции описывают сборку одного образа — `go-booking`.

Соберём образ приложения, выполнив следующую команду в папке `booking`:

```bash
werf build --stages-storage :local
```

### Запуск

Запустим приложение, выполнив следующую команду в папке `booking`:
```bash
werf run --stages-storage :local --docker-options="-d -p 9000:9000 --name go-booking"  go-booking -- /app/run.sh
```

Убедитесь, что контейнер запустился, выполнив следующую команду:
```bash
docker ps -f "name=go-booking"
```

Вы должны увидеть запущенный контейнер `go-booking`, например, вывод может быть подобен следующему:
```bash
CONTAINER ID  IMAGE                                          COMMAND        CREATED        STATUS        PORTS                   NAMES
41d6f49798a8  image-stage-hotel-booking:f27efaf9...1456b0b4  "/app/run.sh"  3 minutes ago  Up 3 minutes  0.0.0.0:9000->9000/tcp  go-booking
```

Откройте в браузере адрес [http://localhost:9000](http://localhost:9000) — вы должны увидеть страницу `revel framework booking demo`. Выполните авторизацию, введя `demo` в качестве логина и пароля.

### Размер собранного образа

Получим размер собранного образа, выполнив:

{% raw %}
```bash
docker images `docker ps -f "name=go-booking" --format='{{.Image}}'`
```
{% endraw %}

Пример вывода:
```bash
REPOSITORY                 TAG                   IMAGE ID          CREATED             SIZE
image-stage-hotel-booking  f27efaf9...1456b0b4   0bf71cb34076      10 minutes ago      1.04 GB
```

Обратите внимание, что размер образа приложения получился **более 1 гигабайта**.

## Оптимизация сборки приложения с использованием артефактов

Непосредственно для запуска приложения необходимы только файлы в папке `/app`, поэтому из образа можно удалить скачанные пакеты и сам компилятор Go. Использование функционала [артефактов в werf]({{ site.baseurl }}/documentation/configuration/stapel_artifact.html) позволяет импортировать в образ только конкретные файлы.

### Сборка

Заменим имеющийся файл `werf.yaml` следующим содержимым:
{% raw %}
```yaml
project: hotel-booking
configVersion: 1
---

artifact: booking-app
from: golang:1.10
ansible:
  beforeInstall:
  - name: Install additional packages
    apt:
      update_cache: yes
      pkg:
      - gcc
      - sqlite3
      - libsqlite3-dev
  install:
  - name: Getting packages
    shell: |
      go get -v github.com/revel/revel
      go get -v github.com/revel/cmd/revel
      (go get -v github.com/revel/examples/booking/... ; true )
  setup:
  - name: Preparing config and building application
    shell: |
      sed -i 's/^http.addr=$/http.addr=0.0.0.0/' $GOPATH/src/github.com/revel/examples/booking/conf/app.conf
      revel build --run-mode dev github.com/revel/examples/booking /app
---
image: go-booking
from: ubuntu:18.04
import:
- artifact: booking-app
  add: /app
  to: /app
  after: install
```
{% endraw %}

В оптимизированных инструкциях сборки само приложение собирается в артефакте `booking-app`, после чего получившиеся файлы импортируются в образ `go-booking`.

Обратите внимание, что при сборке образа `go-booking` используется образ `ubuntu`, а не `golang`.

Соберём приложение с измененным файлом инструкций:
```yaml
werf build --stages-storage :local
```

### Запуск

Перед запуском измененного приложения нужно остановить и удалить запущенный контейнер `go-booking`, собранный и запущенный ранее. В противном случае новый контейнер не сможет запуститься из-за того, что контейнер с таким именем уже существует, порт 9000 занят. Например, выполните следующие команды для остановки и удаления контейнера `go-booking`:

```bash
docker stop go-booking && docker rm go-booking
```

Запустим измененное приложение, выполнив следующую команду:
```bash
werf run --stages-storage :local --docker-options="-d -p 9000:9000 --name go-booking" go-booking -- /app/run.sh
```

Убедитесь, что контейнер запустился, выполнив следующую команду:
```bash
docker ps -f "name=go-booking"
```

Вы должны увидеть запущенный контейнер `go-booking`, например, вывод может быть следующим:
```bash
CONTAINER ID  IMAGE                                          COMMAND        CREATED        STATUS        PORTS                   NAMES
41d6f49798a8  image-stage-hotel-booking:306aa6e8...f71dbe53  "/app/run.sh"  3 minutes ago  Up 3 minutes  0.0.0.0:9000->9000/tcp  go-booking
```

Откройте в браузере адрес [http://localhost:9000](http://localhost:9000) — вы должны увидеть страницу `revel framework booking demo`. Выполните авторизацию, введя `demo` в качестве логина и пароля.

### Размер собранного образа

Получим размер образа, выполнив:
{% raw %}
```bash
docker images `docker ps -f "name=go-booking" --format='{{.Image}}'`
```
{% endraw %}

Пример вывода:
```bash
REPOSITORY                   TAG                      IMAGE ID         CREATED            SIZE
image-stage-hotel-booking    306aa6e8...f71dbe53      0a9943b0da6a     3 minutes ago      103 MB
```

Сравнивая размеры можно увидеть, что **образ, собранный с использованием артефактов, меньше на 90%**, чем первоначальный.

## Вывод

Приведенный пример показывает, что использование артефактов — отличный способ выбросить ненужное из конечного образа. Более того, вы можете использовать один и тот же артефакт (или артефакты) в нескольких образах, определенных в одном `werf.yaml`. Этот прием позволяет увеличить скорость сборки и сократить размер конечного образа.
