---
title: Чарты
permalink: usage/distribute/charts.html
---

## Публикация чартов

Рекомендуемый способ публикации чарта — публикация бандла (который по существу и является чартом) в OCI-репозиторий:

1. Разместите чарт в `.helm`;

2. Если ещё нет `werf.yaml`, то создайте его:
   
   ```yaml
   # werf.yaml:
   project: mychart
   configVersion: 1
   ```

3. Опубликуйте содержимое `.helm` как чарт `example.org/charts/mychart:v1.0.0` в виде OCI-образа:
   
   ```shell
   werf bundle publish --repo example.org/charts --tag v1.0.0
   ```

### Публикация нескольких чартов из одного Git-репозитория

Разместите `.helm` с содержимым чарта и соответствующий ему `werf.yaml` в отдельную директорию для каждого чарта:

```
chart1/
  .helm/
  werf.yaml
chart2/
  .helm/
  werf.yaml
```

Теперь опубликуйте каждый чарт по отдельности:

```shell
cd chart1
werf bundle publish --repo example.org/charts --tag v1.0.0

cd ../chart2
werf bundle publish --repo example.org/charts --tag v1.0.0
```

### .helmignore

Файл `.helmignore`, находящийся в корне чарта, может содержать фильтры по именам файлов, при соответствии которым файлы *не будут добавляться* в чарт при публикации. Формат правил такой же, как и в [.gitignore](https://git-scm.com/docs/gitignore), за исключением:

- `**` не поддерживается;

- `!` в начале строки не поддерживается;

- `.helmignore` не исключает сам себя по умолчанию.
