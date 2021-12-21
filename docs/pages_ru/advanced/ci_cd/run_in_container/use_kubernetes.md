---
title: С помощью Kubernetes
permalink: advanced/ci_cd/run_in_container/use_kubernetes.html
---

> ПРИМЕЧАНИЕ: в настоящее время werf поддерживает сборку образов с _использованием Docker-сервера_ или _без его использования_ (в экспериментальном режиме). Эта страница содержит инструкции, которые подходят только для экспериментального режима _без Docker-сервера_. На данный момент для этого типа сборки доступен только сборщик образов на основе Dockerfile'ов. Сборщик Stapel будет доступен через некоторое время.

## 1. Подготовьте кластер Kubernetes

Выберите один из [доступных режимов работы]({{ "advanced/ci_cd/run_in_container/how_it_works.html#modes-of-operation" | true_relative_url }}) и перейдите к нему.

### Ядро Linux с поддержкой OverlayFS в режиме rootless

Дополнительные действия не требуются.

### Ядро Linux без поддержки OverlayFS в режиме rootless и использование привилегированного контейнера

Дополнительные действия не требуются.

### Ядро Linux без поддержки OverlayFS в режиме rootless и использование непривилегированного контейнера

Для подключения устройства `/dev/fuse` в контейнерах под управлением werf необходим плагин [Fuse device](https://github.com/kuberenetes-learning-group/fuse-device-plugin):

```
# werf-fuse-device-plugin-ds.yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: werf-fuse-device-plugin
spec:
  selector:
    matchLabels:
      name: werf-fuse-device-plugin
  template:
    metadata:
      labels:
        name: werf-fuse-device-plugin
    spec:
      hostNetwork: true
      containers:
      - image: soolaugust/fuse-device-plugin:v1.0
        name: fuse-device-plugin-ctr
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop: ["ALL"]
        volumeMounts:
          - name: device-plugin
            mountPath: /var/lib/kubelet/device-plugins
      volumes:
        - name: device-plugin
          hostPath:
            path: /var/lib/kubelet/device-plugins
```

Примените приведенный выше манифест плагина в пространстве имен `kube-system`:

```
kubectl -n kube-system apply -f werf-fuse-device-plugin-ds.yaml
```

## 2. Настройте доступ к реестру контейнеров

Подготовьте конфигурацию Docker в формате base64 для доступа к реестру.

```yaml
# registrysecret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: registrysecret
data:
  .dockerconfigjson: <base64 of ~/.docker/config.json>
type: kubernetes.io/dockerconfigjson
```

Создайте `registrysecret` в пространстве имен приложения:

```shell
kubectl -n quickstart-application apply -f registrysecret.yaml
```

## 3. Настройте service account для werf

Service account необходим werf для доступа к кластеру kubernetes при развертывании приложения.

```yaml
# werf-service-account.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: werf
  namespace: quickstart-application
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: werf
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
  - kind: ServiceAccount
    name: werf
    namespace: quickstart-application
```

Создайте описанный выше service account `werf` с ролью `cluster-admin` в пространстве имен приложения:

```shell
kubectl -n quickstart-application apply -f werf-service-account.yaml
```

## 4. Выполните развертывание приложения

Выберите один из [доступных режимов работы]({{ "advanced/ci_cd/run_in_container/how_it_works.html#режимы=работы" | true_relative_url }}) и перейдите к нему.

### Ядро Linux с поддержкой OverlayFS в режиме rootless

Команда werf converge будет выполняться в специальном Pod'е. Обратите внимание, что `CONTAINER_REGISTRY_REPO` следует заменить на реальный адрес репозитория реестра контейнеров, для которого на предыдущем шаге мы настраивали `registrysecret`.


```yaml
# werf-converge.yaml
apiVersion: v1
kind: Pod
metadata:
  name: werf-converge
  annotations:
    "container.apparmor.security.beta.kubernetes.io/werf-converge": "unconfined"
spec:
  serviceAccount: werf
  automountServiceAccountToken: true
  restartPolicy: Never
  containers:
  - name: werf-converge
    image: ghcr.io/werf/werf
    args:
      - "sh"
      - "-ec"
      - |
        git clone --depth 1 https://github.com/werf/quickstart-application.git $HOME/quickstart-application &&
        cd $HOME/quickstart-application &&
        werf converge --release quickstart-application --repo CONTAINER_REGISTRY_REPO
    env:
      - name: WERF_NAMESPACE
        valueFrom:
          fieldRef:
            fieldPath: metadata.namespace
    volumeMounts:
      - mountPath: /home/build/.docker/
        name: registrysecret
  volumes:
    - name: registrysecret
      secret:
        secretName: registrysecret
        items:
          - key: .dockerconfigjson
            path: config.json
```

Создайте pod и разверните приложение в соответствующем пространстве имен:

```shell
kubectl -n quickstart-application apply -f werf-converge.yaml
kubectl -n quickstart-application logs -f werf-converge
```

### Ядро Linux без поддержки OverlayFS в режиме rootless и использование привилегированного контейнера

Команда werf converge будет выполняться в специальном Pod'е. Обратите внимание, что `CONTAINER_REGISTRY_REPO` следует заменить на реальный адрес репозитория реестра контейнеров, для которого на предыдущем шаге мы настраивали `registrysecret`.

```yaml
# werf-converge.yaml
apiVersion: v1
kind: Pod
metadata:
  name: werf-converge
spec:
  serviceAccount: werf
  automountServiceAccountToken: true
  restartPolicy: Never
  containers:
  - name: werf-converge
    image: ghcr.io/werf/werf
    securityContext:
      privileged: true
    args:
      - "sh"
      - "-ec"
      - |
        git clone --depth 1 https://github.com/werf/quickstart-application.git $HOME/quickstart-application &&
        cd $HOME/quickstart-application &&
        werf converge --release quickstart-application --repo CONTAINER_REGISTRY_REPO
    env:
      - name: WERF_NAMESPACE
        valueFrom:
          fieldRef:
            fieldPath: metadata.namespace
    volumeMounts:
      - mountPath: /home/build/.docker/
        name: registrysecret
  volumes:
    - name: registrysecret
      secret:
        secretName: registrysecret
        items:
          - key: .dockerconfigjson
            path: config.json
```

Создайте pod и разверните приложение в соответствующем пространстве имен:

```shell
kubectl -n quickstart-application apply -f werf-converge.yaml
kubectl -n quickstart-application logs -f werf-converge
```

### Ядро Linux без поддержки OverlayFS в режиме rootless и использование непривилегированного контейнера

Команда werf converge будет выполняться в специальном Pod'е. Обратите внимание, что `CONTAINER_REGISTRY_REPO` следует заменить на реальный адрес репозитория реестра контейнеров, для которого на предыдущем шаге мы настраивали `registrysecret`.

```yaml
# werf-converge.yaml
apiVersion: v1
kind: Pod
metadata:
  name: werf-converge
  annotations:
    "container.apparmor.security.beta.kubernetes.io/werf-converge": "unconfined"
spec:
  serviceAccount: werf
  automountServiceAccountToken: true
  restartPolicy: Never
  containers:
  - name: werf-converge
    image: ghcr.io/werf/werf
    resources:
      limits:
        github.com/fuse: 1
    args:
      - "sh"
      - "-ec"
      - |
        git clone --depth 1 https://github.com/werf/quickstart-application.git $HOME/quickstart-application &&
        cd $HOME/quickstart-application &&
        werf converge --release quickstart-application --repo CONTAINER_REGISTRY_REPO
    env:
      - name: WERF_NAMESPACE
        valueFrom:
          fieldRef:
            fieldPath: metadata.namespace
    volumeMounts:
      - mountPath: /home/build/.docker/
        name: registrysecret
  volumes:
    - name: registrysecret
      secret:
        secretName: registrysecret
        items:
          - key: .dockerconfigjson
            path: config.json
```

Создайте pod и разверните приложение в целевом пространстве имен:

```shell
kubectl -n quickstart-application apply -f werf-converge.yaml
kubectl -n quickstart-application logs -f werf-converge
```

## Устранение проблем

Если у вас возникли какие-либо сложности, пожалуйста, обратитесь к разделу [Устранение проблем]({{ "advanced/ci_cd/run_in_container/how_it_works.html#troubleshooting" | true_relative_url }})
