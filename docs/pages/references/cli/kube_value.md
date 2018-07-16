---
title: dapp kube value
sidebar: reference
permalink: kube_value.html
folder: command
---
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
