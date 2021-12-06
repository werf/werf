---
title: Use kubernetes
permalink: advanced/ci_cd/run_in_container/use_kubernetes.html
---

> NOTICE: werf currently supports building of images _with docker server_ or _without docker server_ (in experimental mode). This page contains instructions, which are only applicable for experimental mode _without docker server_. Only dockerfile-images builder is available for this mode for now. Stapel-images builder will be available soon.

## 1. Prepare kubernetes cluster

Select and proceed with one of the [available modes of operation]({{ "advanced/ci_cd/run_in_container/how_it_works.html#modes-of-operation" | true_relative_url }}).

### Linux kernel with rootless overlayfs

No actions needed.

### Linux kernel without rootless overlayfs and privileged container

No actions needed.

### Linux kernel without rootless overlayfs and non-privileged container

[Fuse device plugin](https://github.com/kuberenetes-learning-group/fuse-device-plugin) needed to enable `/dev/fuse` device in containers running werf:

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

Apply provided device plugin in the `kube-system` namespace:

```
kubectl -n kube-system apply -f werf-fuse-device-plugin-ds.yaml
```

## 2. Configure access to container registry

Prepare base64 docker config to access your registry.

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

Create `registrysecret` in the target application namespace:

```shell
kubectl -n quickstart-application apply -f registrysecret.yaml
```

## 3. Configure service account for werf

werf need a service account to access kubernetes cluster when deploying application.

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

Create described special service account `werf` with `cluster-admin` role in the target application namespace:

```shell
kubectl -n quickstart-application apply -f werf-service-account.yaml
```

## 4. Perform application deployment

Select and proceed with one of the [available modes of operation]({{ "advanced/ci_cd/run_in_container/how_it_works.html#modes-of-operation" | true_relative_url }}).

### Linux kernel with rootless overlayfs

werf converge command will run in the one-shot Pod. Note that `CONTAINER_REGISTRY_REPO` should be replaced with real address of container registry repo, for which there is `registrysecret` configured in the previous step.

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
  securityContext:
    fsGroup: 1000
  restartPolicy: Never
  containers:
  - name: werf-converge
    image: ghcr.io/werf/werf
    securityContext:
      runAsUser: 1000
      runAsGroup: 1000
    args:
      - "bash"
      - "-lec"
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

### Linux kernel without rootless overlayfs and privileged container

werf converge command will run in the one-shot Pod. Note that `CONTAINER_REGISTRY_REPO` should be replaced with real address of container registry repo, for which there is `registrysecret` configured in the previous step.

```yaml
# werf-converge.yaml
apiVersion: v1
kind: Pod
metadata:
  name: werf-converge
spec:
  serviceAccount: werf
  automountServiceAccountToken: true
  securityContext:
    fsGroup: 1000
  restartPolicy: Never
  containers:
  - name: werf-converge
    image: ghcr.io/werf/werf
    securityContext:
      runAsUser: 1000
      runAsGroup: 1000
      privileged: true
    args:
      - "bash"
      - "-lec"
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

### Linux kernel without rootless overlayfs and non-privileged container

werf converge command will run in the one-shot Pod. Note that `CONTAINER_REGISTRY_REPO` should be replaced with real address of container registry repo, for which there is `registrysecret` configured in the previous step.

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
  securityContext:
    fsGroup: 1000
  restartPolicy: Never
  containers:
  - name: werf-converge
    image: ghcr.io/werf/werf
    securityContext:
      runAsUser: 1000
      runAsGroup: 1000
    resources:
      limits:
        github.com/fuse: 1
    args:
      - "bash"
      - "-lec"
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

If you have any problems please refer to the [troubleshooting section]({{ "advanced/ci_cd/run_in_container/how_it_works.html#troubleshooting" | true_relative_url }})
