# dapp [![Gem Version](https://badge.fury.io/rb/dapp.svg)](https://badge.fury.io/rb/dapp) [![Build Status](https://travis-ci.org/flant/dapp.svg)](https://travis-ci.org/flant/dapp) [![Code Climate](https://codeclimate.com/github/flant/dapp/badges/gpa.svg)](https://codeclimate.com/github/flant/dapp) [![Test Coverage](https://codeclimate.com/github/flant/dapp/badges/coverage.svg)](https://codeclimate.com/github/flant/dapp/coverage)

Утилита для реализации и сопровождения процессов CI/CD (Continuous Integration и Continuous Delivery). Предназначена для использования DevOps-специалистами в качестве связующего звена между кодом приложений (поддерживается Git), инфраструктурой, описанной кодом (Chef) и используемой PaaS (Kubernetes). При этом dapp спроектирована с мыслями о быстроте/эффективности работы, её цель — упростить DevOps-инженерам разработку кода для сборки и уменьшить время ожидания сборки по очередному коммиту.

_На данный момент dapp поддерживает только сборку образов Docker-контейнеров*, делая это быстро и эффективно._

_В планах поддержка полного цикла CI/CD. Вы можете помочь в обсуждениях [задач по kubernetes](https://github.com/flant/dapp/issues?q=is%3Aissue+is%3Aopen+label%3Akubernetes)._


## Особенности

* Уменьшение среднего времени сборки
* Использование общего кэша между сборками
* Поддержка распределённой сборки при использовании общего registry
* Уменьшение размера образа, за счёт вынесения исходных данных и инструментов сборки
* Поддержка chef
* Создание множества образов по одному файлу-описанию
* Продвинутые инструменты отладки собираемого образа


## Установка

Для работы dapp требуется ruby и docker.

* [Установка ruby с помощью rvm](https://rvm.io/rvm/install)
* [Установка docker](https://docs.docker.com/engine/installation/)

dapp распространяется в виде gem-а. Для установки достаточно набрать:

```bash
gem install dapp
```


## Документация

* [Быстрый старт](http://flant.github.io/dapp/get_started.html)

* Вся документация [http://flant.github.io/dapp/](http://flant.github.io/dapp/)


## Лицензия

dapp распространяется на условиях лицензии MIT.

Подробности в файле [LICENSE](https://github.com/flant/dapp/blob/master/LICENSE.txt)
