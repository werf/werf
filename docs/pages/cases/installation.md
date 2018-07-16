---
title: Установка
sidebar: how_to
permalink: installation.html
---

Для работы dapp требуется:

## Ruby

Версия >= 2.1.

[Установка ruby с помощью rvm](https://rvm.io/rvm/install)

## Docker

Версия >= 1.10.0.

[Установка docker](https://docs.docker.com/engine/installation/)

## Заголовочные файлы libssh2 для работы с git-репозиториями через ssh

### Ubuntu

```bash
apt-get install libssh2-1-dev
```

### Centos

```bash
yum install libssh2-devel
```

## Cmake для установки зависимого gem rugged

### Ubuntu

```bash
apt-get install cmake
```

### Centos

```bash
yum install cmake
```

## Gem dapp

```bash
gem install dapp
```
