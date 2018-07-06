---
title: Команды развёртывания приложений
sidebar: doc_sidebar
permalink: kube_commands.html
folder: command
---

### dapp kube deploy
Развернуть приложение, chart описанный в `.helm` согласно [документации helm](https://github.com/kubernetes/helm/blob/master/docs/index.md).

```
dapp kube deploy [options] REPO
```

#### --namespace NAMESPACE
Задаёт `namespace`, по умолчанию используется context namespace.

#### Опции тегирования `--tag TAG`, `--tag-branch`, `--tag-commit`, `--tag-build-id`, `--tag-ci`, `--tag-slug TAG`, `--tag-plain TAG`
Версия образа из указанного репозитория. Значение опций соответствует указываемым в [dapp dimg push](base_commands.html#dapp-dimg-push).

#### --tmp_dir_prefix PREFIX
Переопределяет префикс временной директории, временного chart-a, который удаляется после запуска.

#### --set STRING_ARRAY
Пробрасывает значения в `helm upgrade` (значения могут передаваться в одной опции, в формате `key1=val1,key2=val2`, или разбиваться на несколько).

#### --values FILE_PATH
Пробрасывает значения описанные в YAML файле в `helm upgrade` (опция может использоваться несколько раз).

#### --secret-values FILE_PATH
Декодирует значения описанные в YAML файле, используя [ключ шифрования](#dapp-kube-secret-key-generate), затем пробрасывает их в `helm upgrade` (опция может использоваться несколько раз).

### dapp kube dismiss
Удалить релиз.

```
dapp kube dismiss [options]
```

#### --namespace NAMESPACE
Задаёт `namespace`, по умолчанию используется context namespace.

#### --with_namespace
При удалении релиза также удаляет `namespace`.

### dapp kube lint
Вызывает [`helm lint`](https://docs.helm.sh/helm/#helm-lint) для helm чарта, - тестирует чарт на возможные ошибки.

```
dapp kube lint [options] [REPO]
```

#### --namespace NAMESPACE
Задаёт `namespace`, по умолчанию используется context namespace.

#### Опции тегирования `--tag TAG`, `--tag-branch`, `--tag-commit`, `--tag-build-id`, `--tag-ci`, `--tag-slug TAG`, `--tag-plain TAG`
Влияют на то, что будет в результате использования шаблона `dapp_container_image`. Значение опций соответствует указываемым в [dapp dimg push](base_commands.html#dapp-dimg-push).

#### --set STRING_ARRAY
Пробрасывает значения в `helm lint` (значения могут передаваться в одной опции, в формате `key1=val1,key2=val2`, или разбиваться на несколько).

#### --values FILE_PATH
Определяет YAML фаил, значения из которого пробрасываются в `helm lint` (опция может использоваться несколько раз).

#### --secret-values FILE_PATH
Декодирует значения описанные в YAML файле, используя [ключ шифрования](#dapp-kube-secret-key-generate), затем пробрасывает их в `helm lint` (опция может использоваться несколько раз).

### dapp kube render
Вызывает [`helm template`](https://docs.helm.sh/helm/#helm-template) для helm чарта, - выполняет локальный рендеринг чарта и выводит результат.

```
dapp kube render [options] [REPO]
```

#### --namespace NAMESPACE
Задаёт `namespace`, по умолчанию используется context namespace.

#### Опции тегирования `--tag TAG`, `--tag-branch`, `--tag-commit`, `--tag-build-id`, `--tag-ci`, `--tag-slug TAG`, `--tag-plain TAG`
Влияют на то, что будет в результате использования шаблона `dapp_container_image`. Значение опций соответствует указываемым в [dapp dimg push](base_commands.html#dapp-dimg-push).

#### --tmp-dir-prefix PREFIX
Переопределяет префикс временной директории для размещения временных папок используемых при рендеринге (по умолчанию - /tmp).

#### --set STRING_ARRAY
Пробрасывает значения в `helm render` (значения могут передаваться в одной опции, в формате `key1=val1,key2=val2`, или разбиваться на несколько).

#### --values FILE_PATH
Определяет YAML фаил, значения из которого пробрасываются в `helm render` (опция может использоваться несколько раз).

#### --secret-values FILE_PATH
Декодирует значения описанные в YAML файле, используя [ключ шифрования](#dapp-kube-secret-key-generate), затем пробрасывает их в `helm render` (опция может использоваться несколько раз).

#### -t, --templates GLOB_PATTERN
Задает шаблон имени файлов чарта, которые необходимо рендерить. В качестве параметра указывается шаблон:
* пути до chart `-t .helm/templates/*_custom`;
* пути до subchart `-t .helm/charts/custom_chart/templates/*_custom`;
* имени (`-t *_custom`).

Пути могут быть как относительными, так и абсолютными. Наличие `*` учитывает только один уровень иерархии и не выполняет рекурсивный обход. Ко всем шаблонам в конец добавляется `*`. Например, `dapp kube render -t '*/*-app'` выполнит рендеринг yaml файлов чарта - `./.helm/templates/10-app.yaml` и `./.helm/templates/50-app-api.yaml`.

### dapp kube secret key generate
Сгенерировать ключ шифрования.

```
dapp kube secret key generate
```

### dapp kube secret generate
Зашифровать данные ключом `DAPP_SECRET_KEY`.

```
dapp kube secret generate [FILE_PATH] [options]
```

#### --values
Зашифровать secret-values файл `FILE_PATH`.

#### -o OUTPUT_FILE_PATH
Перенаправляет зашифрованные данные в файл `OUTPUT_FILE_PATH`.

### dapp kube secret extract
Расшифровать данные ключом `DAPP_SECRET_KEY`.

```
dapp kube secret extract [FILE_PATH] [options]
```

#### --values
Расшифровать secret-values файл `FILE_PATH`.

#### -o OUTPUT_FILE_PATH
Перенаправляет расшифрованные данные в файл `OUTPUT_FILE_PATH`.

### dapp kube secret edit
Отредактировать секрет `FILE_PATH`.

```
dapp kube secret edit `FILE_PATH` [options]
```

#### --values
Отредактировать secret-values файл `FILE_PATH`.

### dapp kube secret regenerate
Перегенерировать секреты ключом `DAPP_SECRET_KEY`.

```
dapp kube secret regenerate [SECRET_VALUES_FILE_PATH ...] [options]
```

#### --old-secret-key KEY
Использовать ключ `KEY` для декодирования.

### dapp kube create chart
Создает шаблон чарта в папке `.helm`.

```
dapp kube chart create [options]
```

#### --force
Удалить существующий chart (по умолчанию происходит наложение, при этом существующие файлы не перетираются).

### dapp kube value get

```
dapp kube value get [options] VALUE_KEY [REPO]
```

Выводит значение переменных, устанавливаемых dapp. Список некоторых переменных:
* `global.namespace` - namespace (см. опцию `--namespace`)
* `global.dapp.name` - имя проекта dapp (см. опцию `--name`)
* `global.dapp.repo` - имя репозитория
* `global.dapp.docker_tag` - тэг образа
* `global.dapp.dimg.[<имя dimg>.]docker_image` - docker-образа включая адрес репозитория и тэг
* `global.dapp.dimg.[<имя dimg>.]docker_image_id` - хэш образа, например - `sha256:cce87e0fe251a295a9ae81c8343b48472a74879cd75a0fbbd035bb50f69a2b02`, либо `-` если docker-registry не доступен
* `global.dapp.ci.is_branch` - имеет значение `true`, в случае если происходит деплой из ветки (в случае использования dapp только для deploy, будет `false`)
* `global.dapp.ci.is_tag` - имеет значение `true`, в случае если происходит деплой из тэга
* `global.dapp.ci.tag` - имеет истинное значение только если `is_tag=true`, иначе `-`
* `global.dapp.ci.branch` - имеет истинное значение только если is_branch=true, иначе `-`
* `global.dapp.ci.ref` - тоже что tag или branch, - если `is_tag=false` и `is_branch=false`, то имеет значение `-`

Существует возможность получить `commit_id` для любого git-артефакта, который описан для dimg-образа. Чтобы такое значение появилось в helm-values, надо в dappfile при описании git-артефакта указать директиву `as <name>`, тогда в helm появится значение: `global.dapp.dimg.<имя dimg>.git.<имя, указанное в as>.commit_id`. Данный `commit_id` - это тот коммит, который соответствует состоянию git-артефакта в соответствующем dimg'е. Например:
```
dimg "rails" do
git "https://github.com/hello/world" do
  as "hello_world"
  ...
end
end
```
Получаем значение: `global.dapp.dimg.rails.git.hello_world.commit_id=abcd1010`.

Примеры использования `dapp kube value get`:

```
$ dapp kube value get . :minikube
---
global:
  namespace: default
  dapp:
    name: dapp-test-submodules
    repo: localhost:5000/dapp-test-submodules
    docker_tag: :latest
    ci:
      is_tag: false
      is_branch: true
      branch: master
      tag: \"-\"
      ref: master
    is_nameless_dimg: true
    dimg:
      docker_image: localhost:5000/dapp-test-submodules:latest
      docker_image_id: sha256:cce87e0fe251a295a9ae81c8343b48472a74879cd75a0fbbd035bb50f69a2b02
```

```
$ dapp kube value get global.dapp.ci :minikube
---
is_tag: false
is_branch: true
branch: master
tag: \"-\"
ref: master
```
