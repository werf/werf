---
title: Использование монтирования
sidebar: documentation
permalink: documentation/guides/advanced_build/mounts.html
author: Artem Kladov <artem.kladov@flant.com>
---

## Обзор задачи

В руководстве сначала рассматривается сборка тестового приложения на GO, а затем она оптимизируется для существенного сокращения размера конечного образа, используя монтирование.

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

Создадим папку `booking` и файл `werf.yaml` со следующим содержанием:
{% raw %}
```yaml
project: hotel-booking
configVersion: 1
---

{{ $_ := set . "GoDlPath" "https://dl.google.com/go/" }}
{{ $_ := set . "GoTarball" "go1.11.1.linux-amd64.tar.gz" }}
{{ $_ := set . "GoTarballChecksum" "sha256:2871270d8ff0c8c69f161aaae42f9f28739855ff5c5204752a8d92a1c9f63993" }}
{{ $_ := set . "BaseImage" "ubuntu:18.04" }}

image: go-booking
from: {{ .BaseImage }}
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
  - name: Install additional packages
    apt:
      name: ['gcc','sqlite3','libsqlite3-dev']
      update_cache: yes
  install:
  - name: Getting packages
    shell: |
{{ include "export golang vars" . | indent 6 }}
      go get -v github.com/revel/revel
      go get -v github.com/revel/cmd/revel
      (go get -v github.com/revel/examples/booking/... ; true )
  setup:
  - name: Preparing config and building application
    shell: |
{{ include "export golang vars" . | indent 6 }}
      sed -i 's/^http.addr=$/http.addr=0.0.0.0/' $GOPATH/src/github.com/revel/examples/booking/conf/app.conf
      revel build --run-mode dev github.com/revel/examples/booking /app

# Go template for exporting environment variables
{{- define "export golang vars" -}}
export GOPATH=/go
export PATH=$GOPATH/bin:$PATH:/usr/local/go/bin
{{- end -}}
```
{% endraw %}

Соберём образ приложения, выполнив следующую команду в папке `booking`:
```bash
werf build --stages-storage :local
```

### Запуск

Запустим приложение, выполнив следующую команду в папке `booking`:
```bash
werf run --stages-storage :local --docker-options="-d -p 9000:9000 --name go-booking" go-booking -- /app/run.sh
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

Откройте в браузере адрес [http://localhost:9000](http://localhost:9000) — вы должны увидеть страницу `revel framework booking demo`. Выполните авторизацию введя `demo` в качестве логина и пароля.

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

Размер всех связанных образов можно увидеть в выводе команды `werf build --stages-storage :local`. При сборке werf выводит подробную информацию для каждой стадии в сборке. Среди этой информации имя, тег, размер и дельта, использованные инструкции и команды.

Обратите внимание, что размер образа приложения получился **более 1 гигабайта**.

## Оптимизация сборки приложения

Часто после сборки в образе остается много ненужных файлов. В нашем примере это — APT-кэш, исходный код на GO, а также сам компилятор GO. Всё это можно удалить из конечного образа.

### Оптимизация APT-кэша

[APT](https://wiki.debian.org/Apt) хранит список пакетов в папке `/var/lib/apt/lists/`, а также во время установки сохраняет сами пакеты в папке `/var/cache/apt/`. Поэтому удобно хранить `/var/cache/apt/` вне образа и подключать его между сборками. Содержимое папки `/var/lib/apt/lists/` зависит от статуса установленных пакетов и будет не правильным использовать его между сборками, но хорошо бы хранить ее за пределами образа, чтобы уменьшить его размер.

Чтобы применить описанную выше оптимизацию, добавим следующие инструкции в список инструкций по сборке образа `go-booking`:

```yaml
mount:
- from: tmp_dir
  to: /var/lib/apt/lists
- from: build_dir
  to: /var/cache/apt
```

Читайте больше об инструкциях монтирования [здесь]({{ site.baseurl }}/documentation/configuration/stapel_image/mount_directive.html).

В результате добавленных инструкций папка `/var/lib/apt/lists` будет наполняться во время сборки, но в конченом образе она будет пуста.

Содержимое папки `/var/cache/apt/` кешируется в папке `~/.werf/shared_context/mounts/projects/hotel-booking/var-cache-apt-cf3c1428/`, но в конечном образе она пуста. Монтирование работает только во время сборки, поэтому, если изменяете инструкции сборки и пересобираете образы проекта, папка `/var/cache/apt/` уже будет содержать скачанные ранее пакеты.

В официальном образе Ubuntu есть специальные APT-хуки, которые удаляют APT-кэш после сборки образа. Чтобы отключить эти хуки добавим следующие инструкции в стадию сборки ***beforeInstall***:

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
```yaml
project: hotel-booking
configVersion: 1
---

{{ $_ := set . "GoDlPath" "https://dl.google.com/go/" }}
{{ $_ := set . "GoTarball" "go1.11.1.linux-amd64.tar.gz" }}
{{ $_ := set . "GoTarballChecksum" "sha256:2871270d8ff0c8c69f161aaae42f9f28739855ff5c5204752a8d92a1c9f63993" }}
{{ $_ := set . "BaseImage" "ubuntu:18.04" }}

image: go-booking
from: {{ .BaseImage }}
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
  - name: Install additional packages
    apt:
      name: ['gcc','sqlite3','libsqlite3-dev']
      update_cache: yes
  install:
  - name: Getting packages
    shell: |
{{ include "export golang vars" . | indent 6 }}
      go get -v github.com/revel/revel
      go get -v github.com/revel/cmd/revel
      (go get -v github.com/revel/examples/booking/... ; true )
  setup:
  - name: Preparing config and building application
    shell: |
{{ include "export golang vars" . | indent 6 }}
      sed -i 's/^http.addr=$/http.addr=0.0.0.0/' $GOPATH/src/github.com/revel/examples/booking/conf/app.conf
      revel build --run-mode dev github.com/revel/examples/booking /app

# Go template for exporting environment variables
{{- define "export golang vars" -}}
export GOPATH=/go
export PATH=$GOPATH/bin:$PATH:/usr/local/go/bin
{{- end -}}
```
{% endraw %}

Соберём приложение с измененным набором инструкций:
```bash
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

Откройте в браузере адрес [http://localhost:9000](http://localhost:9000) — вы должны увидеть страницу `revel framework booking demo`. Выполните авторизацию введя `demo` в качестве логина и пароля.

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
image-stage-hotel-booking    306aa6e8...f71dbe53      0a9943b0da6a     3 minutes ago      335 MB
```

### Анализ

Папки, смонтированные с помощью инструкций `from: build_dir` в `werf.yaml`, находятся по пути `~/.werf/shared_context/mounts/projects/hotel-booking/`.
Для анализа содержимого смонтированных папок выполните следующую команду:

```bash
tree -L 3 ~/.werf/shared_context/mounts/projects/hotel-booking
```

Пример вывода (некоторые строки пропущены для уменьшения размера вывода):
```bash
/home/user/.werf/shared_context/mounts/projects/hotel-booking
├── usr-local-go-a179aaae
│   ├── api
│   ├── lib
│   ├── pkg
...
│   └── src
├── usr-local-src-f1bad46a
│   └── go1.11.1.linux-amd64.tar.gz
└── var-cache-apt-28143ccf
    └── archives
        ├── binutils_2.30-21ubuntu1~18.04_amd64.deb
...
        └── xauth_1%3a1.0.10-1_amd64.deb
```

Как вы можете видеть, для каждой папки монтирования, определенной в `werf.yaml`, существует отдельная папка на узле сборки.

Проверьте размер папок, выполнив следующую команду:
```bash
sudo du -kh --max-depth=1 ~/.werf/shared_context/mounts/projects/hotel-booking
```

Пример вывода:
```bash
49M     /home/user/.werf/shared_context/mounts/projects/hotel-booking/var-cache-apt-28143ccf
122M    /home/user/.werf/shared_context/mounts/projects/hotel-booking/usr-local-src-f1bad46a
423M    /home/user/.werf/shared_context/mounts/projects/hotel-booking/usr-local-go-a179aaae
592M    /home/user/.werf/shared_context/mounts/projects/hotel-booking
```

`592MB` — это общий размер файлов использованных при сборке, но исключенных из конечного образа. Несмотря на то, что эти файлы исключены из образа, они доступны при повторной сборке образа. Более того, эти файлы могут быть смонтированы при сборке других образов проекта. Например, если вы добавляете в проект еще один образ, основанный на Ubuntu, вы можете также смонтировать папку `/var/cache/apt`, используя инструкцию `from: build_dir` и при сборке будет использоваться хранилище с уже скачанными на предыдущих сборках пакетами.

Также, примерно `77MB` дискового пространства занято файлами в папках, смонтированных с помощью инструкций `from: tmp_dir`. Эти файлы так же исключаются из конечного образа и удаляются с узла сборки после окончания процесса сборки.

Общая разница межу размером образа с использованием инструкций монтирования и без составляет около `730MB` (1.04GB и 335MB).

**Приведенный пример показывает, что с использованием инструкций монтирования в werf размер образа оказался меньше более чем на 68% по сравнению с первоначальным!**

## Что можно улучшить

* Использовать вместо базового образа Ubuntu образ меньшего размера, например, [alpine](https://hub.docker.com/_/alpine/) или [golang](https://hub.docker.com/_/golang/).
* Использовать [артефакты]({{ site.baseurl }}/documentation/configuration/stapel_artifact.html). В большинстве случаев может дать еще большую оптимизацию по размеру. Размер папки `/app` в образе примерно 17MB (можете проверить, выполнив `werf run --stages-storage :local --docker-options="--rm" go-booking -- du -kh --max-depth=0 /app`). Соответственно, можно выполнить сборку приложения, поместив результат в папку `/app` в артефакте, а затем импортировать в конечный образ приложения только содержимое папки `/app`.
