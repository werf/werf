---
title: Shell сборщик
sidebar: reference
permalink: build_shell.html
folder: build
---

**Shell сборщик устарел, переходите на Ansible-сборщик**

Shell сборщик это аналог RUN команд из Dockerfile. Главное отличие состоит в том, что в dapp сборка с помощью shell-команд происходит по стадиям (последовательность и их назначение описаны в главе [Стадии сборки](stages.html)).

## Структура файлов

- `/Dappfile` - Инструкции для сборки

## Синтаксист dappfile

Синтаксис dappfile очень похож на [вариант chef-синтаксиса](build_chef.html). Смотрите на пример:

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