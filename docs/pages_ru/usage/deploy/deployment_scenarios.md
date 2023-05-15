---
title: Сценарии развертывания
permalink: usage/deploy/deployment_scenarios.html
---

## Обычное развертывание

Обычно развертывание осуществляется командой `werf converge`, которая собирает образы и развертывает приложение, но требует запуска из Git-репозитория приложения. Пример:

```shell
werf converge --repo example.org/mycompany/myapp
```

Если требуется разделить шаги сборки и развертывания, то это можно сделать так:

```shell
werf build --repo example.org/mycompany/myapp
```

```shell
werf converge --require-built-images --repo example.org/mycompany/myapp
```

## Развертывание с использованием произвольных тегов образов

По умолчанию собранные образы получают тег на основе их содержимого, который становится доступен в Values для их дальнейшего использования в шаблонах при развертывании. Но если возникает необходимость тегировать образы иным тегом, то можно использовать параметр `--use-custom-tag`, например:

```shell
werf converge --use-custom-tag '%image%-v1.0.0' --repo example.org/mycompany/myapp
```

Результат: образы были собраны и опубликованы с тегами `<имя image>-v1.0.0`, после чего теги этих образов стали доступны в Values, на основе которых были сформированы и применены конечные манифесты Kubernetes.

В имени тега, указываемом в параметре `--use-custom-tag`, можно использовать шаблоны `%image%`, `%image_slug%` и `%image_safe_slug%` для подставления имени образа и `%image_content_based_tag%` для подставления оригинального тега на основе содержимого.

> Обратите внимание, что при указании произвольного тега публикуется также и образ с тегом на основе содержимого. В дальнейшем при вызове `werf cleanup` образ с тегом на основе содержимого и образы с произвольными тегами удаляются вместе.

Если требуется разделить шаги сборки и развертывания, то это можно сделать так:

```shell
werf build --add-custom-tag '%image%-v1.0.0' --repo example.org/mycompany/myapp
```

```shell
werf converge --require-built-images --use-custom-tag '%image%-v1.0.0' --repo example.org/mycompany/myapp
```

## Развертывание без доступа к Git-репозиторию приложения

Если нужно развернуть приложение без доступа к Git-репозиторию приложения, то необходимо выполнить три шага:

1. Сборка образов и их публикация в container registry.

2. Добавление переданных параметров и публикация основного чарта в OCI-репозиторий. Чарт содержит указатели на опубликованные в первом шаге образы.

3. Применение опубликованного бандла в кластер.

Первые два шага выполняются командой `werf bundle publish`, находясь в Git-репозитории приложения, например:

```shell
werf bundle publish --tag latest --repo example.org/mycompany/myapp
```

А третий шаг выполняется командой `werf bundle apply` уже без необходимости находиться в Git-репозитории приложения, например:

```shell
werf bundle apply --tag latest --release myapp --namespace myapp-production --repo example.org/mycompany/myapp
```

Конечный результат будет тот же самый, что и при использовании `werf converge`.

Если требуется разделить первый и второй шаг, то это можно сделать так:

```shell
werf build --repo example.org/mycompany/myapp
```

```
werf bundle publish --require-built-images --tag latest --repo example.org/mycompany/myapp
```

## Развертывание без доступа к Git-репозиторию и container registry приложения

Если нужно развернуть приложение без доступа к Git-репозиторию приложения и без доступа к container registry приложения, то необходимо выполнить пять шагов:

1. Сборка образов и их публикация в container registry приложения.

2. Добавление переданных параметров и публикация основного чарта в OCI-репозиторий. Чарт содержит указатели на опубликованные в первом шаге образы.

3. Экспорт бандла и связанных с ним образов в локальный архив.

4. Импорт заархивированного бандла и его образов в container registry, доступный из Kubernetes-кластера, используемого для развертывания.

5. Применение в кластер бандла, опубликованного в новом container registry.

Первые два шага выполняются командой `werf bundle publish`, находясь в Git-репозитории приложения, например:

```shell
werf bundle publish --tag latest --repo example.org/mycompany/myapp
```

Третий шаг выполняется командой `werf bundle copy` уже без необходимости находиться в Git-репозитории приложения, например:

```shell
werf bundle copy --from example.org/mycompany/myapp:latest --to archive:myapp-latest.tar.gz
```

Теперь полученный локальный архив `myapp-latest.tar.gz` переносится удобным способом туда, откуда имеется доступ в container registry, используемый для развертывания в Kubernetes-кластер, и снова выполняется команда `werf bundle copy`, например:

```shell
werf bundle copy --from archive:myapp-latest.tar.gz --to registry.internal/mycompany/myapp:latest
```

В результате чарт и связанные с ним образы опубликуются в новый container registry, к которому из Kubernetes-кластера уже есть доступ. Осталось только развернуть опубликованный бандл в кластер командой `werf bundle apply`, например:

```shell
werf bundle apply --tag latest --release myapp --namespace myapp-production --repo registry.internal/mycompany/myapp
```

На этом шаге уже не требуется доступ ни в Git-репозиторий приложения, ни в его первоначальный container registry. Конечный результат развертывания бандла будет тот же самый, что и при использовании `werf converge`.

Если требуется разделить первый и второй шаг, то это можно сделать так:

```shell
werf build --repo example.org/mycompany/myapp
```

```
werf bundle publish --require-built-images --tag latest --repo example.org/mycompany/myapp
```

## Развертывание сторонним инструментом

Если нужно выполнить применение конечных манифестов приложения не с werf, а с использованием другого инструмента (kubectl, Helm, ...), то необходимо выполнить три шага:

1. Сборка образов и их публикация в container registry.

2. Формирование конечных манифестов.

3. Развертывание получившихся манифестов в кластер, используя сторонний инструмент.

Первые два шага выполняются командой `werf render`, находясь в Git-репозитории приложения:

```shell
werf render --output manifests.yaml --repo example.org/mycompany/myapp
```

Теперь полученные манифесты можно передать в сторонний инструмент для дальнейшего развертывания, например:

```shell
kubectl apply -f manifests.yaml
```

> Обратите внимание, что некоторые специальные возможности werf вроде возможности изменения порядка развертывания ресурсов на основании их веса (аннотация `werf.io/weight`) скорее всего не будут поддерживаться при применении манифестов сторонним инструментом.

Если требуется разделить первый и второй шаг, то это можно сделать так:

```shell
werf build --repo example.org/mycompany/myapp
```

```
werf render --require-built-images --output manifests.yaml --repo example.org/mycompany/myapp
```

## Развертывание сторонним инструментом без доступа к Git-репозиторию приложения

Если нужно выполнить применение конечных манифестов приложения не с werf, а с использованием другого инструмента (kubectl, Helm, ...), при этом не имея доступа к Git-репозиторию приложения, то необходимо выполнить три шага:

1. Сборка образов и их публикация в container registry.

2. Добавление переданных параметров и публикация основного чарта в OCI-репозиторий. Чарт содержит указатели на опубликованные в первом шаге образы.

3. Формирование из бандла конечных манифестов.

4. Развертывание получившихся манифестов в кластер используя сторонний инструмент.

Первые два шага выполняются командой `werf bundle publish`, находясь в Git-репозитории приложения:

```shell
werf bundle publish --tag latest --repo example.org/mycompany/myapp
```

А третий шаг выполняется командой `werf bundle render` уже без необходимости находиться в Git-репозитории приложения, например:

```shell
werf bundle render --output manifests.yaml --tag latest --release myapp --namespace myapp-production --repo example.org/mycompany/myapp
```

Теперь полученные манифесты можно передать в сторонний инструмент для дальнейшего развертывания, например:

```shell
kubectl apply -f manifests.yaml
```

> Обратите внимание, что некоторые специальные возможности werf, вроде возможности изменения порядка развертывания ресурсов на основании их веса (аннотация `werf.io/weight`), скорее всего не будут поддерживаться при применении манифестов сторонним инструментом.

Если требуется разделить первый и второй шаг, то это можно сделать так:

```shell
werf build --repo example.org/mycompany/myapp
```

```
werf bundle publish --require-built-images --tag latest --repo example.org/mycompany/myapp
```

## Сохранение отчета о развертывании

Команды `werf converge` и `werf bundle apply` имеют параметр `--save-deploy-report`, который позволяет сохранить отчёт о последнем развертывании в файл. Отчёт содержит имя релиза, Namespace, статус развертывания и ряд других данных. Пример:

```shell
werf converge --save-deploy-report
```

Результат: после развертывания появится файл `.werf-deploy-report.json`, содержащий информацию о последнем релизе.

Путь к отчёту о развертывании можно изменить параметром `--deploy-report-path`.

## Удаление развернутого приложения

Удалить развернутое приложение можно командой `werf dismiss`, запущенной из Git-репозитория приложения, например:

```shell
werf dismiss --env staging
```

При отсутствии доступа к Git-репозиторию приложения можно явно указать имя релиза и Namespace:

```shell
werf dismiss --release myapp-staging --namespace myapp-staging
```

... или использовать отчёт о предыдущем развертывании, включаемый опцией `--save-deploy-report` у `werf converge` и `werf bundle apply`, который содержит имя релиза и Namespace:

```shell
werf converge --save-deploy-report
cp .werf-deploy-report.json /anywhere
cd /anywhere
werf dismiss --use-deploy-report
```

Путь к отчёту о развертывании можно изменить параметром `--deploy-report-path`.
