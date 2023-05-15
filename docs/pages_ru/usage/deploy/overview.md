---
title: Обзор
permalink: usage/deploy/overview.html
---

При организации доставки приложения в Kubernetes необходимо определиться с тем, какой формат выбрать для управления конфигурацией развёртывания (параметризации, управления зависимостями, конфигурации под различные окружения и т.д.), а также способом применения этой конфигурации – непосредственно механизмом развёртывания.

В werf встроен Helm, и именно он используется для решения перечисленных задач. Разработка и сопровождение конфигурации реализуется с помощью Helm-чарта, а для процесса развёртывания предлагается Helm c дополнительными возможностями:

- отслеживание состояния выкатываемых ресурсов (с возможностью изменения поведения [для каждого ресурса]({{ "/reference/deploy_annotations.html" | true_relative_url }})):
  - умное ожидание готовности ресурсов;
  - мгновенное завершение проблемного развертывания без необходимости ожидания таймаута;
  - прогресс развёртывания, логи, системные события и ошибки приложения.
- использование произвольного порядка развертывания для любых ресурсов, а не только для хуков;
- ожидание создания и готовности ресурсов, не принадлежащих релизу;
- интеграция сборки и развертывания и многое другое.

werf стремится сделать работу с Helm более простой, удобной и гибкой, при этом не ломая обратную совместимость с Helm-чартами, Helm-шаблонами и Helm-релизами.

## Простой пример развертывания

Для развертывания простого приложения достаточно двух файлов и команды `werf converge`, запущенной в Git-репозитории приложения:

```yaml
# .helm/templates/hello.yaml:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello
spec:
  selector:
    matchLabels:
      app: hello
  template:
    metadata:
      labels:
        app: hello
    spec:
      containers:
      - image: nginxdemos/hello:plain-text
```

```yaml
# werf.yaml:
configVersion: 1
project: hello
```

```shell
werf converge --repo registry.example.org/repo --env production
```

Результат: Deployment `hello` развёрнут в Namespace'е `hello-production`.

## Расширенный пример развертывания

Более сложный пример развертывания со сборкой образов и внешними Helm-чартами:

```yaml
# werf.yaml:
configVersion: 1
project: myapp
---
image: backend
dockerfile: Dockerfile
```

```dockerfile
# Dockerfile:
FROM node
WORKDIR /app
COPY . .
RUN npm ci
CMD ["node", "server.js"]
```

```yaml
# .helm/Chart.yaml:
dependencies:
- name: postgresql
  version: "~12.1.9"
  repository: https://charts.bitnami.com/bitnami
```

```yaml
# .helm/values.yaml:
backend:
  replicas: 1
```

{% raw %}

```yaml
# .helm/templates/backend.yaml:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
spec:
  replicas: {{ $.Values.backend.replicas }}
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
    spec:
      containers:
      - image: {{ $.Values.werf.image.backend }}
```

{% endraw %}

```shell
werf converge --repo registry.example.org/repo --env production
```

Результат: собран образ `backend`, а затем Deployment `backend` и ресурсы чарта `postgresql` развёрнуты в Namespace'е `myapp-production`.
