---
title: Команды очистки
sidebar: reference
permalink: dimg_cleanup.html
folder: command
---
### dapp dimg cleanup
Убраться в системе после некорректного завершения работы dapp, удалить нетегированные docker-образы и docker-контейнеры.

```
dapp dimg cleanup [options]
```

#### Примеры

##### Запустить
```bash
$ dapp dimg cleanup
```

##### Посмотреть, какие команды могут быть выполнены
```bash
$ dapp dimg cleanup --dry-run
backend
  docker rm -f dd4ec7v33
  docker rmi dimgstage-dapp-test-project:07758b3ec8aec701a01 dimgstage-dapp-test-project:ec701a0107758b3ec8a
```

### dapp dimg cleanup repo
Удалить теги [приложений](definitions.html#dimg) [проекта](definitions.html#проект), исходя из соответствующих схем тегирования.

<table class="tag-scheme">
  <tr>
    <td>Опция тегирования</td>
    <td>--tag<br />--tag-slug<br />--tag-plain</td>
    <td>--tag-branch</td>
    <td>--tag-commit</td>
    <td>--tag-build-id</td>
    <td>--tag-ci</td>
  </tr>
  <tr>
    <td>Схема тегирования</td>
    <td>custom</td>
    <td>git_branch</td>
    <td>git_commit</td>
    <td>ci</td>
    <td>git_tag или git_branch</td>
  </tr>
</table>

* Имена которых содержат неактуальные данные:
    * ветка или тег удалены из репозитория (`git_branch`, `git_tag`);
    * комит отсутствует в репозитории, был сделан rebase (`git_commit`).
* Загружены в registry более одного месяца назад (`git_tag`, `git_commit`);
* Лишние, в случае, если привышен лимит в 10 тегов на [приложение](definitions.html#dimg), исходя из времени загрузки образа в registry (`git_tag`, `git_commit`).

```
dapp dimg cleanup repo [options] [DIMG ...] REPO
```

#### `--with-stages`
Соответствует вызову команды `dapp dimg stages cleanup local` с опцией `--improper-repo-cache`.

#### `--without-kube`
Отключает проверку использования образов в кластерах. См. подробнее [о работе очистки](cleanup_for_advanced_build.html#автоматическая-очистка-по-политикам).

