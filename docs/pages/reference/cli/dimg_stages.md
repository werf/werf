---
title: dapp dimg stages
sidebar: reference
permalink: reference/cli/dimg_stages.html
---

### dapp dimg stages push
Выкатить кэш собранных приложений проекта в репозиторий.

```
dapp dimg stages push [options] [DIMG ...] REPO
```

#### Примеры

##### Выкатить кэш собранных dimg-ей проекта в репозиторий localhost:5000/test
```bash
$ dapp dimg stages push localhost:5000/test
```

##### Посмотреть, какие образы могут быть добавлены в репозиторий
```bash
$ dapp dimg stages push localhost:5000/test --dry-run
backend
  localhost:5000/test:dimgstage-be032ed31bd96506d0ed550fa914017452b553c7f1ecbb136216b2dd2d3d1623
  localhost:5000/test:dimgstage-2183f7db73687e727d9841594e30d8cb200312290a0a967ef214fe3771224ee2
  localhost:5000/test:dimgstage-f7d4c5c420f29b7b419f070ca45f899d2c65227bde1944f7d340d6e37087c68d
  localhost:5000/test:dimgstage-256f03ccf980b805d0e5944ab83a28c2374fbb69ef62b8d2a52f32adef91692f
  localhost:5000/test:dimgstage-31ed444b92690451e7fa889a137ffc4c3d4a128cb5a7e9578414abf107af13ee
  localhost:5000/test:dimgstage-02b636d9316012880e40da44ee5da3f1067cedd66caa3bf89572716cd1f894da
frontend
  localhost:5000/test:dimgstage-192c0e9d588a51747ce757e61be13acb4802dc2cdefbeec53ca254d014560d46
  localhost:5000/test:dimgstage-427b999000024f9268a46b889d66dae999efbfe04037fb6fc0b1cd7ebb4600b0
  localhost:5000/test:dimgstage-07fe13aec1e9ce0fe2d2890af4e4f81aaa984c89a2b91fbd0e164468a1394d46
  localhost:5000/test:dimgstage-ba95ec8a00638ddac413a13e303715dd2c93b80295c832af440c04a46f3e8555
  localhost:5000/test:dimgstage-02b636d9316012880e40da44ee5da3f1067cedd66caa3bf89572716cd1f894da
```

### dapp dimg stages pull
Импортировать необходимый кэш приложений проекта, если он присутствует в репозитории **REPO**.

Если не указана опция **`--all`**, импорт будет выполнен до первого найденного кэша стейджа для каждого dimg-a.

```
dapp dimg stages pull [options] [DIMG ...] REPO
```

#### `--all`
Попробовать импортировать весь необходимый кэш.

#### Примеры

##### Импортировать кэш dimg-ей проекта из репозитория localhost:5000/test
```bash
$ dapp dimg stages pull localhost:5000/test
```

##### Посмотреть, поиск каких образов в репозитории localhost:5000/test может быть выполнен
```bash
$ dapp dimg stages pull localhost:5000/test --all --dry-run
backend
  localhost:5000/test:dimgstage-be032ed31bd96506d0ed550fa914017452b553c7f1ecbb136216b2dd2d3d1623
  localhost:5000/test:dimgstage-2183f7db73687e727d9841594e30d8cb200312290a0a967ef214fe3771224ee2
  localhost:5000/test:dimgstage-f7d4c5c420f29b7b419f070ca45f899d2c65227bde1944f7d340d6e37087c68d
  localhost:5000/test:dimgstage-256f03ccf980b805d0e5944ab83a28c2374fbb69ef62b8d2a52f32adef91692f
  localhost:5000/test:dimgstage-31ed444b92690451e7fa889a137ffc4c3d4a128cb5a7e9578414abf107af13ee
  localhost:5000/test:dimgstage-02b636d9316012880e40da44ee5da3f1067cedd66caa3bf89572716cd1f894da
frontend
  localhost:5000/test:dimgstage-192c0e9d588a51747ce757e61be13acb4802dc2cdefbeec53ca254d014560d46
  localhost:5000/test:dimgstage-427b999000024f9268a46b889d66dae999efbfe04037fb6fc0b1cd7ebb4600b0
  localhost:5000/test:dimgstage-07fe13aec1e9ce0fe2d2890af4e4f81aaa984c89a2b91fbd0e164468a1394d46
  localhost:5000/test:dimgstage-ba95ec8a00638ddac413a13e303715dd2c93b80295c832af440c04a46f3e8555
  localhost:5000/test:dimgstage-02b636d9316012880e40da44ee5da3f1067cedd66caa3bf89572716cd1f894da
```

### dapp dimg stages cleanup local
Удалить неактуальный локальный кэш приложений проекта.

```
dapp dimg stages cleanup local [options] [REPO]
```

#### `--improper-repo-cache`
Удалить кэш, который не использовался при сборке приложений в репозитории **REPO**.

#### `--improper-git-commit`
Удалить кэш, связанный с отсутствующими в git-репозиториях коммитами.

#### `--improper-cache-version`
Удалить устаревший кэш приложений проекта.

#### Примеры

##### Оставить только актуальный кэш, исходя из приложений в localhost:5000/test
```bash
$ dapp dimg stages cleanup local --improper-repo-cache localhost:5000/test
```

##### Удалить кэш, версия которого не совпадает с текущей
```bash
$ dapp dimg stages cleanup local --improper-cache-version localhost:5000/test
```

##### Почистить кэш после rebase в одном из связанных git-репозиториев
```bash
$ dapp dimg stages cleanup local --improper-git-commit localhost:5000/test
```

### dapp dimg stages cleanup repo
Удалить неиспользуемый кэш приложений в репозитории **REPO**.

```
dapp dimg stages cleanup repo [options] REPO
```

#### `--improper-repo-cache`
Удалить кэш, который не использовался при сборке приложений в репозитории **REPO**.

#### `--improper-git-commit`
Удалить кэш, связанный с отсутствующими в git-репозиториях коммитами.

#### `--improper-cache-version`
Удалить устаревший кэш приложений проекта.

#### Примеры

##### Удалить неактуальный кэш в репозитории localhost:5000/test, где:

* Версия кэша не соответствует текущей.
* Не связан с приложениями в репозитории.
* Собран из коммитов, которые в данный момент отсутствуют в git-репозиториях проекта.

```bash
$ dapp dimg stages cleanup repo localhost:5000/test --improper-cache-version --improper-repo-cache --improper-git-commit
```

### dapp dimg stages flush local
Удалить кэш приложений проекта.

```
dapp dimg stages flush local [options]
```

#### Примеры

##### Удалить кэш приложений
```bash
$ dapp dimg stages flush local
```

### dapp dimg stages flush repo
Удалить кэш приложений проекта в репозитории **REPO**.

```
dapp dimg stages flush repo [options] REPO
```

#### Примеры

##### Удалить весь кэш приложений в репозитории localhost:5000/test
```bash
$ dapp dimg stages flush repo localhost:5000/test
```
