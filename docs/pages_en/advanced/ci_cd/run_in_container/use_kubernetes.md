---
title: Use Kubernetes
permalink: advanced/ci_cd/run_in_container/use_kubernetes.html
---

> NOTICE: werf currently supports building images _with the Docker server_ or _without the Docker server_ (in experimental mode). This page contains information applicable only to the experimental mode _without the Docker server_. For now, only the Dockerfile image builder is available for this mode. The Stapel image builder will be available soon.

## 1. Prepare the Kubernetes cluster

Select one of the [available operating modes]({{ "advanced/ci_cd/run_in_container/how_it_works.html#modes-of-operation" | true_relative_url }}) and navigate to it.

### Linux kernel with rootless OverlayFS

No additional actions are required.

### Linux kernel without rootless OverlayFS and privileged container

No additional actions are required.

### Linux kernel without rootless OverlayFS and non-privileged container

The [fuse device plugin](https://github.com/kuberenetes-learning-group/fuse-device-plugin) is required to enable the `/dev/fuse` device in werf containers:

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

Apply the above plugin manifest to the `kube-system` namespace:

```
kubectl -n kube-system apply -f werf-fuse-device-plugin-ds.yaml
```

## 2. Configure access to the container registry

Prepare the Docker configuration in base64 format to access the registry.

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

Create the `registrysecret` in the application namespace:

```shell
kubectl -n quickstart-application apply -f registrysecret.yaml
```

## 3. Configure service account for werf

A service account is needed for werf to access the Kubernetes cluster when deploying the application.

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

Create the `werf` service account described above with the `cluster-admin` role in the target application namespace:

```shell
kubectl -n quickstart-application apply -f werf-service-account.yaml
```

## 4. Deploy the application

Select one of the [available operating modes]({{ "advanced/ci_cd/run_in_container/how_it_works.html#modes-of-operation" | true_relative_url }}) and navigate to it.

### Linux kernel with rootless OverlayFS

The werf converge command will be executed in a special Pod.  Note that `CONTAINER_REGISTRY_REPO` must be replaced with the real address of the container registry repo for which we configured `registrysecret` at the previous step.

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

Create a Pod and deploy the application to the target namespace:

```shell
kubectl -n quickstart-application apply -f werf-converge.yaml
kubectl -n quickstart-application logs -f werf-converge
```

### Linux kernel without rootless OverlayFS and privileged container

The werf converge command will be executed in a special Pod.  Note that `CONTAINER_REGISTRY_REPO` must be replaced with the real address of the container registry repo for which we configured `registrysecret` at the previous step.

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

Create pod and perform application deployment in the target namespace:

```shell
kubectl -n quickstart-application apply -f werf-converge.yaml
kubectl -n quickstart-application logs -f werf-converge
```

### Linux kernel without rootless OverlayFS and non-privileged container

The werf converge command will be executed in a special Pod.  Note that `CONTAINER_REGISTRY_REPO` must be replaced with the real address of the container registry repo for which we configured `registrysecret` at the previous step.

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

Create pod and perform application deployment in the target namespace:

```shell
kubectl -n quickstart-application apply -f werf-converge.yaml
kubectl -n quickstart-application logs -f werf-converge
```

## Troubleshooting

In case of problems, refer to the [Troubleshooting section]({{ "advanced/ci_cd/run_in_container/how_it_works.html#troubleshooting" | true_relative_url }})
