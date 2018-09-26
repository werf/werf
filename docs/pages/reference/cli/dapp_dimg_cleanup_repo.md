---
title: dapp dimg cleanup repo
sidebar: reference
permalink: reference/cli/dapp_dimg_cleanup_repo.html
---

Удалить теги приложений проекта, исходя из соответствующих схем тегирования.

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
* Лишние, в случае, если привышен лимит в 10 тегов на приложение, исходя из времени загрузки образа в registry (`git_tag`, `git_commit`).

```
dapp dimg cleanup repo [options] [DIMG ...] REPO
```

### `--with-stages`
Соответствует вызову команды `dapp dimg stages cleanup local` с опцией `--improper-repo-cache`.

### `--without-kube`
Отключает проверку использования образов в кластерах. См. подробнее [о работе очистки](cleanup_for_advanced_build.html#автоматическая-очистка-по-политикам).
