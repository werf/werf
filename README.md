<p align="center">
  <img src="https://github.com/flant/dapp/raw/master/logo.png" style="max-height:100%;" height="100">
</p>
<p align="center">
  <a href="https://badge.fury.io/rb/dapp"><img alt="Gem Version" src="https://badge.fury.io/rb/dapp.svg" style="max-width:100%;"></a>
  <a href="https://travis-ci.org/flant/dapp"><img alt="Build Status" src="https://travis-ci.org/flant/dapp.svg" style="max-width:100%;"></a>
  <a href="https://codeclimate.com/github/flant/dapp"><img alt="Code Climate" src="https://codeclimate.com/github/flant/dapp/badges/gpa.svg" style="max-width:100%;"></a>
  <a href="https://codeclimate.com/github/flant/dapp/coverage"><img alt="Test Coverage" src="https://codeclimate.com/github/flant/dapp/badges/coverage.svg" style="max-width:100%;"></a>
</p>

___

Dapp — это утилита для реализации и сопровождения процессов CI/CD (Continuous Integration и Continuous Delivery). Предназначена для использования DevOps-специалистами в качестве связующего звена между кодом приложений (поддерживается Git), инфраструктурой, описанной кодом (Chef) и используемой PaaS (Kubernetes).  При этом dapp спроектирована с мыслями о быстроте/эффективности работы, её предназначение — упростить DevOps-инженерам разработку кода для сборки, уменьшить время ожидания сборки по очередному коммиту, автоматизировать деплой.


## Что дает использование dapp?

* Уменьшение среднего времени сборки.
* Использование общего кэша между сборками.
* Возможность распределённой сборки при использовании общего registry.
* Уменьшение размера образа, за счёт вынесения исходных данных и инструментов сборки.
* Возможность сборки образа с помощью chef.
* Создание множества образов по одному файлу-описанию.
* Продвинутые инструменты отладки собираемого образа.
* Деплой в kubernetes через [helm](https://helm.sh/) (kubernetes package manager).


## Установка

Для работы dapp требуется:

#### Ruby

Версия >= 2.1.

[Установка ruby с помощью rvm](https://rvm.io/rvm/install)

#### Docker

Версия > 1.10.

[Установка docker](https://docs.docker.com/engine/installation/)

#### Заголовочные файлы libssh2 для работы с git-репозиториями через ssh

###### Ubuntu

```bash
apt-get install libssh2-1-dev
```

###### Centos

```bash
yum install libssh2-devel
```

#### Заголовочные файлы libssl для работы с git-репозиториями через ssl

###### Ubuntu

```bash
apt-get install libssl-dev
```

###### Centos

```bash
yum install openssl-devel
```

#### Cmake для установки зависимого gem rugged

###### Ubuntu

```bash
apt-get install cmake
```

###### Centos

```bash
yum install cmake
```

#### Gem dapp

```bash
gem install dapp
```

### Установка с автообновлением patch-релизов (экспериментально)

***Предупреждение:*** Будет установлен gem dapp с системными зависимостями и cron-job для автообновления до последнего patch-релиза x.x.**x** в рамках устанавливаемой версии.

```bash
curl https://raw.githubusercontent.com/flant/dapp/master/install.sh | bash
```

Установить dapp версии 0.13.5 и cron-job автообновления до последнего патч релиза в рамках версии 0.13.**x**:

```bash
curl https://raw.githubusercontent.com/flant/dapp/master/install.sh | bash -s 0.13.5
```

## Документация

* [Быстрый старт](http://flant.github.io/dapp/get_started.html)

* Вся документация [http://flant.github.io/dapp/](http://flant.github.io/dapp/)

* [Официальная русскоязычная группа поддержки в Telegram](https://t.me/dapp_ru)


## Лицензия

Dapp распространяется на условиях лицензии Apache v2.

Подробности в файле [LICENSE](https://github.com/flant/dapp/blob/master/LICENSE)
