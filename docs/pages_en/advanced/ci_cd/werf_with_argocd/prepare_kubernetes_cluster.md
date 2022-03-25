---
title: Prepare Kubernetes cluster
permalink: advanced/ci_cd/werf_with_argocd/prepare_kubernetes_cluster.html
---

## 1. Install ArgoCD with werf plugin

[Install ArgoCD](https://argo-cd.readthedocs.io/en/stable/getting_started/#1-install-argo-cd).

Enable werf sidecar plugin:

1. Edit `deploy/argocd-repo-server`:
    ```shell
    kubectl -n argocd edit deploy argocd-repo-server
    ```
2. Add sidecar container and apparmor annotation:
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

## 2. Install ArgoCD Image Updater

Install ArgoCD Image Updater with the ["continuous deployment of OCI Helm chart type application" patch](https://github.com/argoproj-labs/argocd-image-updater/pull/405):

```shell
kubectl apply -n argocd -f https://raw.githubusercontent.com/werf/3p-argocd-image-updater/master/manifests/install.yaml
```
