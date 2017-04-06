---
title: Shell сборщик
sidebar: doc_sidebar
permalink: stages_for_build.html
folder: build
---

Shell сборщик это аналог RUN команд из Dockerfile. Главное отличие состоит в том, что в dapp сборка с помощью shell-команд происходит по стадиям (последовательность и их назначение описаны в главе [Стадии сборки](stages_for_build.html)).

## Шаги сборки для приложения

Для начала разберём, что нужно выполнить для сборки образа php приложения, например demo приложения symfony, по списку шагов из предыдущей главы.

- установить системное ПО и системные зависимости

Нужно установить php, например, 7-ой версии. Понадобятся расширения php7.0-sqlite3 (для приложения) php7.0-xml php7.0-zip (для composer).

- настроить системное ПО

Для работы веб-сервера нужен пользователь. Это будет пользователь phpapp.

- установить прикладные зависимости

Для установки зависимостей проекта нужен composer. Его можно установить скачиванием phar файла, поэтому в системное ПО добавится curl.

- добавить код

Код будет располагаться в финальном образе в директории /demo. Всем файлам проекта нужно установить владельцем пользователя phpapp.

- настроить приложение

Никаких особых настроек производить не нужно. Единственной настройкой будет ip адрес, на котором  слушает веб-сервер, но эта настройка будет в скрипте /opt/start.sh, который будет запускаться при старте контейнера.

В качестве иллюстрации для стадии setup добавится создание файла version.txt с текущей датой.

## Dappfile

После изложенного должен быть понятен вот такой Dappfile:

```
dimg 'symfony-demo-app' do
  docker.from 'ubuntu:16.04'

  git do
    add '/' do
      to '/demo'
    end
  end

  shell do
    before_install do
      run 'apt-get update',
          'apt-get install -y curl php7.0',
          # пользователь phpapp
          'groupadd -g 242 phpapp',
          'useradd -m  -d /home/phpapp -g 242 -u 242 phpapp',
          # скрипт для простого запуска приложения
          "echo '#!/bin/bash' > /opt/start.sh",
          "echo 'cd /demo' >> /opt/start.sh",
          "echo su -c \"'php bin/console server:run 0.0.0.0:8000'\" phpapp >> /opt/start.sh",
          'chmod +x /opt/start.sh'
    end
    install do
      run 'apt-get install -y php7.0-sqlite3 php7.0-xml php7.0-zip',
          # установка composer
          'curl -LsS https://getcomposer.org/download/1.4.1/composer.phar -o /usr/local/bin/composer',
          'chmod a+x /usr/local/bin/composer'
    end
    before_setup do
      # исходным текстам нужно сменить права и запустить composer install
      run 'chown phpapp:phpapp -R /demo && cd /demo',
          "su -c 'composer install' phpapp"
    end
    setup do
      # используем текущую дату как версию приложения
      run 'echo `date` > /demo/version.txt',
          'chown phpapp:phpapp /demo/version.txt'
    end
  end

  # Порт совпадает с портом, указанным в /opt/start.sh
  docker.expose 8000
end
```


## Сборка и запуск

Для запуска этого Dappfile нужно склонировать репозиторий с приложением и создать в корне репозитория Dаppfile.

```
git clone https://github.com/symfony/symfony-demo.git
cd symfony-demo
vi Dappfile
```

Далее нужно собрать образ приложения можно командой

```
dapp dimg build
```

А запустить командой

```
dapp dimg run -d -p 8000:8000 -- /opt/start.sh
```

После чего проверить браузером или в консоли

```
curl host_ip:8000
```


## Что не так?

* Набор команд echo для создания файла start.sh вполне заменим на ещё одну директиву git и хранение файла в репозитории.
* Если директивой git можно копировать файлы, то почему бы в этой директиве не указать права на эти файлы?
* composer install требуется не каждый раз, а только при изменении файла package.json, поэтому было бы отлично, если эта команда запускалась только при изменении этого файла.

Эти проблемы будут разобраны в следующей главе [Поддержка git](git_for_build.html)
