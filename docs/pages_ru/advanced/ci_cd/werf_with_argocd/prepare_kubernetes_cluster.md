---
title: Подготовка кластера Kubernetes
permalink: advanced/ci_cd/werf_with_argocd/prepare_kubernetes_cluster.html
---

## 1. Установите ArgoCD с плагином werf

[Установите ArgoCD](https://argo-cd.readthedocs.io/en/stable/getting_started/#1-install-argo-cd).

Включите sidecar-плагин werf:

1. Отредактируйте `deploy/argocd-repo-server`:
    ```shell
    kubectl -n argocd edit deploy argocd-repo-server
    ```
2. Добавьте sidecar-контейнер и аннотацию apparmor:
    ```yaml
    ...
    metadata:
      annotations:
        "container.apparmor.security.beta.kubernetes.io/werf-argocd-cmp-sidecar": "unconfined"
      ...
    spec:
      ...
      template:
        ...
        spec:
          containers:
          - image: registry.werf.io/werf/werf-argocd-cmp-sidecar:1.2-alpha
            imagePullPolicy: Always
            name: werf-argocd-cmp-sidecar
            volumeMounts:
            - mountPath: /var/run/argocd
              name: var-files
            - mountPath: /home/argocd/cmp-server/plugins
              name: plugins
            - mountPath: /tmp
              name: tmp
    ```

## 2. Установите ArgoCD Image Updater

Установите ArgoCD Image Updater с патчем [Непрерывное развертывание приложений на основе OCI-совместимых Helm-чартов](https://github.com/argoproj-labs/argocd-image-updater/pull/405):

```shell
kubectl apply -n argocd -f https://raw.githubusercontent.com/werf/3p-argocd-image-updater/master/manifests/install.yaml
```
