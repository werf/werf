---
title: Телеметрия
permalink: resources/telemetry.html
---

Чтобы работать над улучшением функций werf и направлять разработку в правильное русло, мы собираем анонимные данные об использовании. Эти данные никак не привязываются к пользователям и не содержат никакой персонализированной информации.

Они помогают понять, как используется werf, и бросить силы на улучшение нужных функций.

## Пример передаваемых данных и их расшифровка

Ниже приведены примеры передаваемых данных:

```json
{
  "ts": 1658231825280,
  "executionID": "2f75d020-684e-4224-9013-35e95e1b7721",
  "projectID": "b4c2d019070529344a6967d8c73d578101c6554fcdcae3d00bb93a9692523cb1",
  "command": "werf converge",
  "attributes": {
    "arch": "amd64",
    "os": "linux",
    "version": "dev",
    "ci": true,
    "ciName": "gitlab"
  },
  "eventType": "CommandStarted",
  "eventData": {
    "commandOptions": [
      {
        "name": "repo",
        "asCli": false,
        "asEnv": false,
        "count": 0
      }
    ]
  },
  "schemaVersion": 1
}
```

```json
{
  "ts": 1658231836102,
  "executionID": "2f75d020-684e-4224-9013-35e95e1b7721",
  "projectID": "b4c2d019070529344a6967d8c73d578101c6554fcdcae3d00bb93a9692523cb1",
  "command": "werf converge",
  "attributes": {
    "arch": "amd64",
    "os": "linux",
    "version": "dev",
    "ci": true,
    "ciName": "gitlab"
  },
  "eventType": "CommandExited",
  "eventData": {
    "exitCode": 0,
    "durationMs": 10827
  },
  "schemaVersion": 1
}
```

Здесь передаются следующие данные об использовании:

* `ts` — timestamp;
* `executionID` — UUID;
* `projectID` — контрольная сумма (SHA-256) от git origin URL проекта;
* `command` — название выполняемой команды werf;
* `attributes` — атрибуты окружения:
  * `os`;
  * `arch`;
  * `trdl group-channel`;
  * `werf version`.
  * `ci` — используется ли CI/CD система;
  * `ciName` — имя используемой CI/CD системы (gitlab, github-actions, и т.п.).
* `eventType` — тип события:
  * `CommandStarted`;
  * `CommandExited`;
* `eventData` — данные события, включают в себя:
  * exit code;
  * длительность работы команды;
  * имена используемых опций.

Вы можете сами убедиться в том, что мы собираем только обезличенную информацию, как указано в примере выше. Для этого см. исходные код werf, отвечающий за телеметрию: файлы [event.go](https://github.com/werf/werf/blob/main/pkg/telemetry/event.go) и [telemetrywerfio.go](https://github.com/werf/werf/blob/main/pkg/telemetry/telemetrywerfio.go) пакета [telemetry](https://github.com/werf/werf/tree/main/pkg/telemetry).

## Выгрузка отчета телеметрии

Мы доверяем нашим пользователям и хотим, чтобы они так же доверяли нам, поэтому сделали все максимально прозрачно. Используя переменную окружения `WERF_TELEMETRY_LOG_FILE`, можно задать путь к лог-файлу, куда будут складываться все данные, передаваемые телеметрией.

## Отключение телеметрии

Пользователь может отключить телеметрию с помощью переменной окружения `WERF_TELEMETRY`. Для отключения необходимо выставить значение в `0`:

```shell
export WERF_TELEMETRY=0
``` 
