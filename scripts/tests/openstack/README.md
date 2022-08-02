Скрипт создания kubernetes кластеров и docker registry для werf.io
=======

Данный скрипт создает:
- 6 кластеров Kubernetes (1.11 - 1.16) с одной нодой (мастером) и доступом с наружи по IP и по доменам `kubernetes-1-11.ci.werf.io`, ... .
- Одну машину для docker registry с доменом `registry.ci.werf.io` с Lets Encrypt сертификатом.

## Пример запуска скрипта

```shell
export SSH_PRIVATE_KEY_PATH=~/.ssh/id_rsa
./run
```

## Как удалить все

```shell
./run destroy
```
