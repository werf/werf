---
title: Релиз
permalink: advanced/helm/releases/release.html
---

В то время как чарт — набор конфигурационных файлов вашего приложения, релиз (**release**) — это объект времени выполнения, экземпляр вашего приложения, развернутого с помощью werf.

## Хранение релизов

Информация о каждой версии релиза хранится в самом кластере Kubernetes. werf поддерживает сохранение в произвольном namespace в объектах Secret или ConfigMap.

По умолчанию, werf хранит информацию о релизах в объектах Secret в целевом namespace, куда происходит деплой приложения. Это полностью совместимо с конфигурацией по умолчанию по хранению релизов в [Helm 3](https://helm.sh), что полностью совместимо с конфигурацией [Helm 2](https://helm.sh) по умолчанию. Место хранения информации о релизах может быть указано при деплое с помощью параметров werf: `--helm-release-storage-namespace=NS` и `--helm-release-storage-type=configmap|secret`.

Для получения информации обо всех созданных релизах можно использовать команду [werf helm list]({{ "reference/cli/werf_helm_list.html" | true_relative_url }}), а для просмотра истории конкретного релиза [werf helm history]({{ "reference/cli/werf_helm_history.html" | true_relative_url }}).

### Замечание о совместимости с Helm

werf полностью совместим с уже установленным Helm 2, т.к. хранение информации о релизах осуществляется одним и тем же образом, как и в Helm. Если вы используете в Helm специфичное место хранения информации о релизах, а не значение по умолчанию, то вам нужно указывать место хранения с помощью опций werf `--helm-release-storage-namespace` и `--helm-release-storage-type`.

Информация о релизах, созданных с помощью werf, может быть получена с помощью Helm, например, командами `helm list` и `helm get`. С помощью werf также можно обновлять релизы, развернутые ранее с помощью Helm.

Команда [`werf helm list -A`]({{ "reference/cli/werf_helm_list.html" | true_relative_url }}) выводит список релизов созданных werf или Helm 3. Релизы, созданные через werf могут свободно просматриваться через утилиту helm командами `helm list` или `helm get` и другими.

### Совместимость с Helm 2

Существующие релизы helm 2 (созданные например через werf v1.1) могут быть конвертированы в helm 3 либо автоматически во время работы команды [`werf converge`]({{ "/reference/cli/werf_converge.html" | true_relative_url }}), либо с помощью команды [`werf helm migrate2to3`]({{ "/reference/cli/werf_helm_migrate2to3.html" | true_relative_url }}).
