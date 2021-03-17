---
title: Использование монтирования
sidebar: documentation
permalink: guides/advanced_build/mounts.html
author: Artem Kladov <artem.kladov@flant.com>
---

## Обзор задачи

В руководстве сначала рассматривается сборка тестового приложения на GO, а затем она оптимизируется для существенного сокращения размера конечного образа, используя монтирование.

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
```
project: gowebapp
configVersion: 1
---

{{ $_ := set . "GoDlPath" "https://dl.google.com/go/" }}
{{ $_ := set . "GoTarball" "go1.14.7.linux-amd64.tar.gz" }}
{{ $_ := set . "GoTarballChecksum" "sha256:4a7fa60f323ee1416a4b1425aefc37ea359e9d64df19c326a58953a97ad41ea5" }}
{{ $_ := set . "BaseImage" "ubuntu:18.04" }}

image: gowebapp
from: {{ .BaseImage }}
docker:
  WORKDIR: /app
ansible:
  beforeInstall:
  - name: Install essential utils
    apt:
      name: ['curl','git','tree']
      update_cache: yes
  - name: Download the Go tarball
    get_url:
      url: {{ .GoDlPath }}{{ .GoTarball }}
      dest: /usr/local/src/{{ .GoTarball }}
      checksum:  {{ .GoTarballChecksum }}
  - name: Extract the Go tarball if Go is not yet installed or not the desired version
    unarchive:
      src: /usr/local/src/{{ .GoTarball }}
      dest: /usr/local
      copy: no
  install:
  - name: Getting packages
    shell: |
{{ include "export go vars" . | indent 6 }}
      go get github.com/josephspurrier/gowebapp
  setup:
  - file:
      path: /app
      state: directory
  - name: Copying config
    shell: |
{{ include "export go vars" . | indent 6 }}
      cp -r $GOPATH/src/github.com/josephspurrier/gowebapp/config /app/config
      cp -r $GOPATH/src/github.com/josephspurrier/gowebapp/static /app/static
      cp -r $GOPATH/src/github.com/josephspurrier/gowebapp/template /app/template
      cp $GOPATH/bin/gowebapp /app/

{{- define "export go vars" -}}
export GOPATH=/go
export PATH=$GOPATH/bin:$PATH:/usr/local/go/bin
{{- end -}}
```
{% endraw %}

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
werf-stages-storage/gowebapp   84d7...44992265   07cdc430e1c8      10 minutes ago      708MB
```

Размер всех связанных образов можно увидеть в выводе команды `werf build --stages-storage :local`. При сборке werf выводит подробную информацию для каждой стадии в сборке. Среди этой информации имя, тег, размер и дельта, использованные инструкции и команды.

Обратите внимание, что размер образа приложения получился **более 700 Мегабайт**.

## Оптимизация сборки приложения

Часто после сборки в образе остается много ненужных файлов. В нашем примере это — APT-кэш, исходный код на GO, а также сам компилятор GO. Всё это можно удалить из конечного образа.

### Оптимизация APT-кэша

[APT](https://wiki.debian.org/Apt) хранит список пакетов в папке `/var/lib/apt/lists/`, а также во время установки сохраняет сами пакеты в папке `/var/cache/apt/`. Поэтому удобно хранить `/var/cache/apt/` вне образа и подключать его между сборками. Содержимое папки `/var/lib/apt/lists/` зависит от статуса установленных пакетов и будет не правильным использовать его между сборками, но хорошо бы хранить ее за пределами образа, чтобы уменьшить его размер.

Чтобы применить описанную выше оптимизацию, добавим следующие инструкции в список инструкций по сборке образа `gowebapp`:

```yaml
mount:
- from: tmp_dir
  to: /var/lib/apt/lists
- from: build_dir
  to: /var/cache/apt
```

Читайте больше об инструкциях монтирования [здесь]({{ site.baseurl }}/configuration/stapel_image/mount_directive.html).

В результате добавленных инструкций папка `/var/lib/apt/lists` будет наполняться во время сборки, но в конченом образе она будет пуста.

Содержимое папки `/var/cache/apt/` кэшируется в папке `~/.werf/shared_context/mounts/projects/gowebapp/var-cache-apt-28143ccf/`, но в конечном образе она пуста. Монтирование работает только во время сборки, поэтому, если изменяете инструкции сборки и пересобираете образы проекта, папка `/var/cache/apt/` уже будет содержать скачанные ранее пакеты.

В официальном образе Ubuntu есть специальные APT-хуки, которые удаляют APT-кэш после сборки образа. Чтобы отключить эти хуки, добавим следующие инструкции в стадию сборки ***beforeInstall***:

```yaml
- name: Disable docker hook for apt-cache deletion
  shell: |
    set -e
    sed -i -e "s/DPkg::Post-Invoke.*//" /etc/apt/apt.conf.d/docker-clean
    sed -i -e "s/APT::Update::Post-Invoke.*//" /etc/apt/apt.conf.d/docker-clean
```

### Оптимизация сборки

В приведенном выше примере сборки скачивается и распаковывается компилятор GO, но его исходные файлы не нужны в собранном образе. Поэтому используем монтирование папок `/usr/local/src` и `/usr/local/go`, чтобы хранить их содержимое за пределами образа.

При сборке приложения, на стадии ***setup*** используется папка `/go`, согласно переменной окружения `GOPATH`. Эта папка содержит все необходимые пакеты зависимостей и исходный код самого приложения. Результат сборки помещается в папку `/app`, а папка `/go` более не нужна для работы приложения, поэтому её можно вынести и использовать только при сборке.

Добавим следующие инструкции монтирования:
```yaml
- from: tmp_dir
  to: /go
- from: build_dir
  to: /usr/local/src
- from: build_dir
  to: /usr/local/go
```

### Итоговый файл werf.yaml

{% raw %}
```
project: gowebapp
configVersion: 1
---

{{ $_ := set . "GoDlPath" "https://dl.google.com/go/" }}
{{ $_ := set . "GoTarball" "go1.14.7.linux-amd64.tar.gz" }}
{{ $_ := set . "GoTarballChecksum" "sha256:4a7fa60f323ee1416a4b1425aefc37ea359e9d64df19c326a58953a97ad41ea5" }}
{{ $_ := set . "BaseImage" "ubuntu:18.04" }}

image: gowebapp
from: {{ .BaseImage }}
docker:
  WORKDIR: /app
mount:
- from: tmp_dir
  to: /var/lib/apt/lists
- from: build_dir
  to: /var/cache/apt
- from: tmp_dir
  to: /go
- from: build_dir
  to: /usr/local/src
- from: build_dir
  to: /usr/local/go
ansible:
  beforeInstall:
  - name: Disable docker hook for apt-cache deletion
    shell: |
      set -e
      sed -i -e "s/DPkg::Post-Invoke.*//" /etc/apt/apt.conf.d/docker-clean
      sed -i -e "s/APT::Update::Post-Invoke.*//" /etc/apt/apt.conf.d/docker-clean          
  - name: Install essential utils
    apt:
      name: ['curl','git','tree']
      update_cache: yes
  - name: Download the Go tarball
    get_url:
      url: {{ .GoDlPath }}{{ .GoTarball }}
      dest: /usr/local/src/{{ .GoTarball }}
      checksum:  {{ .GoTarballChecksum }}
  - name: Extract the Go tarball if Go is not yet installed or not the desired version
    unarchive:
      src: /usr/local/src/{{ .GoTarball }}
      dest: /usr/local
      copy: no
  install:
  - name: Getting packages
    shell: |
{{ include "export go vars" . | indent 6 }}
      go get github.com/josephspurrier/gowebapp
  setup:
  - file:
      path: /app
      state: directory
  - name: Copying config
    shell: |
{{ include "export go vars" . | indent 6 }}
      cp -r $GOPATH/src/github.com/josephspurrier/gowebapp/config /app/config
      cp -r $GOPATH/src/github.com/josephspurrier/gowebapp/static /app/static
      cp -r $GOPATH/src/github.com/josephspurrier/gowebapp/template /app/template
      cp $GOPATH/bin/gowebapp /app/

{{- define "export go vars" -}}
export GOPATH=/go
export PATH=$GOPATH/bin:$PATH:/usr/local/go/bin
{{- end -}}
```
{% endraw %}

Соберём приложение с измененным набором инструкций:
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
werf-stages-storage/gowebapp   84d7...44992265   07cdc430e1c8   10 minutes ago   181MB
```

### Анализ

Папки, смонтированные с помощью инструкций `from: build_dir` в `werf.yaml`, находятся по пути `~/.werf/shared_context/mounts/projects/gowebapp/`.
Для анализа содержимого смонтированных папок выполните следующую команду:

```shell
tree -L 3 ~/.werf/shared_context/mounts/projects/gowebapp
```

Пример вывода (некоторые строки пропущены для уменьшения размера вывода):
```shell
/home/user/.werf/shared_context/mounts/projects/gowebapp/
├── usr-local-go-a179aaae
│   ├── api
│   ├── lib
│   ├── pkg
...
│   └── src
├── usr-local-src-6ad4baf1
│   └── go1.14.7.linux-amd64.tar.gz
└── var-cache-apt-28143ccf
    └── archives
        ├── ca-certificates_20190110~18.04.1_all.deb
...
        └── xauth_1%3a1.0.10-1_amd64.deb
```

Как вы можете видеть, для каждой папки монтирования, определенной в `werf.yaml`, существует отдельная папка на узле сборки.

Проверьте размер папок, выполнив следующую команду:
```shell
sudo du -kh --max-depth=1 ~/.werf/shared_context/mounts/projects/gowebapp
```

Пример вывода:
```shell
19M     /home/user/.werf/shared_context/mounts/projects/gowebapp/var-cache-apt-28143ccf
119M    /home/user/.werf/shared_context/mounts/projects/gowebapp/usr-local-src-6ad4baf1
348M    /home/user/.werf/shared_context/mounts/projects/gowebapp/usr-local-go-a179aaae
485M    /home/user/.werf/shared_context/mounts/projects/gowebapp
```

`485M` — это общий размер файлов использованных при сборке, но исключенных из конечного образа. Несмотря на то, что эти файлы исключены из образа, они доступны при повторной сборке образа. Более того, эти файлы могут быть смонтированы при сборке других образов проекта. Например, если вы добавляете в проект еще один образ, основанный на Ubuntu, вы можете также смонтировать папку `/var/cache/apt`, используя инструкцию `from: build_dir` и при сборке будет использоваться хранилище с уже скачанными на предыдущих сборках пакетами.

Также, примерно `70Мб` дискового пространства занято файлами в папках, смонтированных с помощью инструкций `from: tmp_dir`. Эти файлы так же исключаются из конечного образа и удаляются с узла сборки после окончания процесса сборки.

Общая разница межу размером образа с использованием инструкций монтирования и без, составляет около `527Мб` (708Мб — 181Мб).

**Приведенный пример показывает, что с использованием инструкций монтирования в werf размер образа оказался меньше более чем на 70% по сравнению с первоначальным!**

## Что можно улучшить

* Использовать вместо базового образа Ubuntu образ меньшего размера, например, [alpine](https://hub.docker.com/_/alpine/) или [golang](https://hub.docker.com/_/golang/).
* Использовать [артефакты]({{ site.baseurl }}/configuration/stapel_artifact.html). В большинстве случаев может дать еще большую оптимизацию по размеру. Размер папки `/app` в образе примерно 15Мб (можете проверить, выполнив `werf run --stages-storage :local --docker-options="--rm" gowebapp -- du -kh --max-depth=0 /app`). Соответственно, можно выполнить сборку приложения, поместив результат в папку `/app` в артефакте, а затем импортировать в конечный образ приложения только содержимое папки `/app`.
