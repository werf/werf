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
  "userID": "",
  "projectID": "b4c2d019070529344a6967d8c73d578101c6554fcdcae3d00bb93a9692523cb1",
  "command": "werf converge",
  "attributes": {
    "arch": "amd64",
    "os": "linux",
    "version": "dev",
    "ci": true,
    "ciName": "gitlab",
    "extra": {}
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
  "schemaVersion": 2
}
```

```json
{
  "ts": 1658231836102,
  "executionID": "2f75d020-684e-4224-9013-35e95e1b7721",
  "userID": "",
  "projectID": "b4c2d019070529344a6967d8c73d578101c6554fcdcae3d00bb93a9692523cb1",
  "command": "werf converge",
  "attributes": {
    "arch": "amd64",
    "os": "linux",
    "version": "dev",
    "ci": true,
    "ciName": "gitlab",
    "extra": {}
  },
  "eventType": "CommandExited",
  "eventData": {
    "exitCode": 0,
    "durationMs": 10827
  },
  "schemaVersion": 2
}
```

Пример начала сборки:

```json
{
  "ts": 1709251100000,
  "executionID": "3a8b9c01-123e-4567-8901-23e4567890ab",
  "userID": "",
  "projectID": "c4d3e019070529344a6967d8c73d578101c6554fcdcae3d00bb93a9692523cb1",
  "command": "werf build",
  "attributes": {
    "arch": "arm64",
    "os": "darwin",
    "version": "2.60.2",
    "ci": false,
    "extra": {}
  },
  "eventType": "BuildStarted",
  "eventData": {
    "imagesCount": 3,
    "containerBackend": "docker",
    "inContainer": false
  },
  "schemaVersion": 2
}
```

Пример завершения сборки:

```json
{
  "ts": 1709251150000,
  "executionID": "3a8b9c01-123e-4567-8901-23e4567890ab",
  "userID": "",
  "projectID": "c4d3e019070529344a6967d8c73d578101c6554fcdcae3d00bb93a9692523cb1",
  "command": "werf build",
  "attributes": {
    "arch": "arm64",
    "os": "darwin",
    "version": "2.60.2",
    "ci": false,
    "extra": {}
  },
  "eventType": "BuildFinished",
  "eventData": {
    "durationMs": 50000,
    "success": true,
    "imagesCount": 3
  },
  "schemaVersion": 2
}
```

Пример завершения сборки образа:

```json
{
  "ts": 1709251140000,
  "executionID": "3a8b9c01-123e-4567-8901-23e4567890ab",
  "userID": "",
  "projectID": "c4d3e019070529344a6967d8c73d578101c6554fcdcae3d00bb93a9692523cb1",
  "command": "werf build",
  "attributes": {
    "arch": "arm64",
    "os": "darwin",
    "version": "2.60.2",
    "ci": false,
    "extra": {}
  },
  "eventType": "ImageBuildFinished",
  "eventData": {
    "image": "backend",
    "durationMs": 15500,
    "rebuilt": true,
    "configType": "stapel"
  },
  "schemaVersion": 2
}
```

Пример завершения сборки стадии:

```json
{
  "ts": 1709251135000,
  "executionID": "3a8b9c01-123e-4567-8901-23e4567890ab",
  "userID": "",
  "projectID": "c4d3e019070529344a6967d8c73d578101c6554fcdcae3d00bb93a9692523cb1",
  "command": "werf build",
  "attributes": {
    "arch": "arm64",
    "os": "darwin",
    "version": "2.60.2",
    "ci": false,
    "extra": {}
  },
  "eventType": "StageBuildFinished",
  "eventData": {
    "image": "backend",
    "stage": "install",
    "durationMs": 8500,
    "fromCache": false,
    "baseImageSource": "repo",
    "baseImagePulled": true
  },
  "schemaVersion": 2
}
```

Здесь передаются следующие данные об использовании:

* `ts` — timestamp;
* `executionID` — UUID запуска;
* `userID` — анонимный идентификатор пользователя; может быть пустым, если не задан;
* `projectID` — контрольная сумма (SHA-256) от git origin URL проекта;
* `command` — название выполняемой команды werf;
* `attributes` — атрибуты окружения:
  * `os` — операционная система;
  * `arch` — архитектура;
  * `groupChannel` — используемый `trdl` group-channel, если задан через `TRDL_USE_WERF_GROUP_CHANNEL`;
  * `version` — версия werf;
  * `ci` — используется ли CI/CD система;
  * `ciName` — имя используемой CI/CD системы (`gitlab`, `github-actions` и т. п.);
  * `extra` — набор дополнительных анонимных атрибутов из переменных окружения `WERF_TELEMETRY_EXTRA_ATTRIBUTE_*`.
* `eventType` — тип события:
  * `CommandStarted` — запуск команды;
  * `CommandExited` — завершение команды;
  * `UnshallowFailed` — ошибка при `unshallow` в GitLab CI;
  * `BuildStarted` — начало сборки образов;
  * `BuildFinished` — завершение сборки образов;
  * `ImageBuildFinished` — завершение сборки отдельного образа;
  * `StageBuildFinished` — завершение сборки отдельной стадии.
* `eventData` — данные события, специфичные для каждого типа:
  * для `CommandStarted`: список используемых опций команды с флагами (`asCli`, `asEnv`, `count`);
  * для `CommandExited`: код выхода (`exitCode`), длительность выполнения команды в миллисекундах (`durationMs`);
  * для `UnshallowFailed`: сообщение об ошибке (`errorMessage`), версии GitLab Runner (`gitlabRunnerVersion`) и Server (`gitlabServerVersion`);
  * для `BuildStarted`: количество образов для сборки (`imagesCount`), используемый контейнерный бэкенд (`containerBackend`: `docker`/`buildkit`), а также флаг запуска werf внутри контейнера (`inContainer`);
  * для `BuildFinished`: длительность сборки (`durationMs`), успешность (`success`), количество образов (`imagesCount`);
  * для `ImageBuildFinished`: имя образа (`image`), время сборки (`durationMs`), флаг пересборки (`rebuilt`), тип конфигурации образа (`configType`: `stapel`/`dockerfile`/`staged`/`unknown`);
  * для `StageBuildFinished`: имя образа (`image`), имя стадии (`stage`), время сборки (`durationMs`), флаг использования кэша (`fromCache`), источник базового образа (`baseImageSource`: `repo`/`secondary`/пустое значение), флаг загрузки базового образа (`baseImagePulled`).

Вы можете сами убедиться в том, что мы собираем только обезличенную информацию, как указано в примере выше. Для этого см. исходный код werf, отвечающий за телеметрию: файлы [event.go](https://github.com/werf/werf/blob/main/pkg/telemetry/event.go) и [telemetrywerfio.go](https://github.com/werf/werf/blob/main/pkg/telemetry/telemetrywerfio.go) пакета [telemetry](https://github.com/werf/werf/tree/main/pkg/telemetry).

## Выгрузка отчета телеметрии

Мы доверяем нашим пользователям и хотим, чтобы они так же доверяли нам, поэтому сделали все максимально прозрачно. Используя переменную окружения `WERF_TELEMETRY_LOG_FILE`, можно задать путь к лог-файлу, куда будут складываться все данные, передаваемые телеметрией.

## Отключение телеметрии

Пользователь может отключить телеметрию с помощью переменной окружения `WERF_TELEMETRY`. Для отключения необходимо выставить значение в `0`:

```shell
export WERF_TELEMETRY=0
```
