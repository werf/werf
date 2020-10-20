---
title: Настройка Minikube
sidebar: documentation
permalink: documentation/reference/development_and_debug/setup_minikube.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

Использование werf для сборки и деплоя локально требует настройки следующих компонент:

     .-----------------------------------------.
     |          Kubernetes (minikube)          |
     |                                         |
     |  .-------------.                        |
     |  | rc/registry |<-----( kubelet pull )  |
     |  '-------------'                        |
     |         `---------.                     |
     |                   |                     |
     |         .-----------------------.       |
     |         | service/werf-registry |       |
     |         '-----------------------'       |
     |                   |                     |
     '-------------------|---------------------'
                         |    .-----------------------------.
                         '----|   proxy localhost:5000 to   |
                              |    service/werf-registry    |<-----( werf push --repo :minikube )
                              |  with kubectl port-forward  |
                              '-----------------------------'

1. Установите [minikube](https://github.com/kubernetes/minikube).
2. Запустите minikube:

   {% raw %}
   ```shell
   minikube start
   ```
   {% endraw %}

3. Включите модуль registry в minikube, реализующий функционал Docker registry:

   {% raw %}
   ```shell
   minikube addons enable registry
   ```
   {% endraw %}

4. Запустите сервис `werf-registry` на порту 5000:

   {% raw %}
   ```shell
   kubectl -n kube-system expose rc/registry --type=ClusterIP --port=5000 --target-port=5000 --name=werf-registry
   ```
   {% endraw %}

5. Добавьте DNS-имя запущенного сервиса Docker registry в виртуальную машину:

   {% raw %}
   ```shell
   export REGISTRY_IP=$(kubectl -n kube-system get svc/werf-registry -o=template={{.spec.clusterIP}})
   minikube ssh "echo '$REGISTRY_IP werf-registry.kube-system.svc.cluster.local' | sudo tee -a /etc/hosts"
   ```
   {% endraw %}

6. Добавьте DNS-имя запущенного сервиса Docker registry на вашем хосте:

   {% raw %}
   ```shell
   echo "127.0.0.1 werf-registry.kube-system.svc.cluster.local" | sudo tee -a /etc/hosts
   ```
   {% endraw %}

7. В отдельном терминале на вашем хосте запустите проброс портов, для доступа к сервису Docker registry:

   {% raw %}
   ```shell
   kubectl port-forward --namespace kube-system service/werf-registry 5000
   ```
   {% endraw %}

8. Проверьте подключение на хосте:

   {% raw %}
   ```shell
   curl -X GET werf-registry.kube-system.svc.cluster.local:5000/v2/_catalog
   ```
   {% endraw %}
