# dapp [![Gem Version](https://badge.fury.io/rb/dapp.svg)](https://badge.fury.io/rb/dapp) [![Build Status](https://travis-ci.org/flant/dapp.svg)](https://travis-ci.org/flant/dapp) [![Code Climate](https://codeclimate.com/github/flant/dapp/badges/gpa.svg)](https://codeclimate.com/github/flant/dapp) [![Test Coverage](https://codeclimate.com/github/flant/dapp/badges/coverage.svg)](https://codeclimate.com/github/flant/dapp/coverage)

Утилита для реализации и сопровождения процессов CI/CD (Continuous Integration и Continuous Delivery). Она предназначена для использования DevOps-специалистами в качестве связующего звена между кодом приложений (поддерживается Git), описанной кодом инфраструктурой (Chef) и используемой PaaS (Kubernetes). При этом dapp спроектирована с мыслями о быстроте/эффективности работы, её цель — упростить DevOps-инженерам разработку кода для сборки и уменьшить время ожидания сборки по очередному коммиту.

На данный момент dapp поддерживает только сборку образов Docker-контейнеров*, делая это быстро и эффективно. При сборке предусмотрены:
* кэширование;
* поддержка Chef (образ контейнера может настраиваться по cookbook);
* создание множества образов по одному декларативному файлу;
* применение сторонних инструментов внутри образа с сохранением только нужных файлов в финальном образе;
* … и ряд других особенностей, призванных ускорить сборку и сделать образы минимальными по объёму.

Подробнее о возможностях dapp и их применении на практике читайте в следующих разделах документации.

\* В планах развития dapp — поддержка полного цикла CI/CD.

## Особенности

* Уменьшение среднего времени сборки.
* Использование общего кэша между сборками.
* Поддержка распределённой сборки при использовании общего registry.
* Уменьшение размера образа, за счёт вынесения исходных данных и инструментов сборки.
* Продвинутые инструменты отладки.

## Установка

### [Install rvm and ruby](https://rvm.io/rvm/install)
### [Install docker](https://docs.docker.com/engine/installation/)
### Install dapp
```bash
gem install dapp
```

## [Документация](http://flant.github.io/dapp/)

## [Лицензия](https://github.com/flant/dapp/blob/master/LICENSE.txt)
