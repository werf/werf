---
title: Shell сборщик
sidebar: not_used
permalink: shell_builder.html
---

Shell-инструкции аналогичны `RUN` в Dockerfile. Главное отличие состоит в том, что в dapp сборка происходит по стадиям (последовательность и их назначение описаны в главе [Стадии сборки](stages.html)).

## Пример dappfile.yaml для [Java приложения](https://habr.com/company/flant/blog/348436/)

```
artifact: appserver
from: maven:latest
git:
  - add: '/app'
    to: '/usr/src/atsea'
shell:
  install:
    - cd /usr/src/atsea
    - mvn -B -f pom.xml -s /usr/share/maven/ref/settings-docker.xml dependency:resolve
    - mvn -B -s /usr/share/maven/ref/settings-docker.xml package -DskipTests
---
artifact: storefront
from: node:latest
git:
  - add: /app/react-app
    to: /usr/src/atsea/app/react-app
shell:
  install:
    - cd /usr/src/atsea/app/react-app
    - npm install
    - npm run build
---
dimg: app
from: java:8-jdk-alpine
shell:
  beforeInstall:
    - mkdir /app
    - adduser -Dh /home/gordon gordon
import:
  - artifact: appserver
    add: '/usr/src/atsea/target/AtSea-0.0.1-SNAPSHOT.jar'
    to: '/app/AtSea-0.0.1-SNAPSHOT.jar'
    after: install
  - artifact: storefront
    add: /usr/src/atsea/app/react-app/build
    to: /static
    after: install
docker:
  ENTRYPOINT: ["java", "-jar", "/app/AtSea-0.0.1-SNAPSHOT.jar"]
  CMD: ["--spring.profiles.active=postgres"]
```

## Пример Dappfile для [Symfony приложения](https://habr.com/company/flant/blog/336212/)

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
