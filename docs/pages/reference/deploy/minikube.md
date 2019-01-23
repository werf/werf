---
title: Minikube
sidebar: reference
permalink: reference/deploy/minikube.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

Using werf to build and deploy on your localhost requires to setup the following components:

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

1. Instal [minikube](https://github.com/kubernetes/minikube).
2. Start minikube:

     ```
     minikube start
     ```

3. Enable minikube registry addon:

     ```
     minikube addons enable registry
     ```

4. Run custom `werf-registry` service with the binding to 5000 port:

     ```
     kubectl -n kube-system expose rc/registry --type=ClusterIP --port=5000 --target-port=5000 --name=werf-registry
     ```

5. Setup registry domain in minikube VM:

     ```
     export REGISTRY_IP=$(kubectl -n kube-system get svc/werf-registry -o=template={{.spec.clusterIP}})
     minikube ssh "echo '$REGISTRY_IP werf-registry.kube-system.svc.cluster.local' | sudo tee -a /etc/hosts"
     ```

6. Setup registry domain in host system:

     ```
     echo "127.0.0.1 werf-registry.kube-system.svc.cluster.local" | sudo tee -a /etc/hosts
     cat /etc/hosts
     ```

7. Run port forwarder on host system in a separate terminal:

     ```
     kubectl port-forward --namespace kube-system service/werf-registry 5000
     ```

8. Check connectivity on host system:

     ```
     curl -X GET werf-registry.kube-system.svc.cluster.local:5000/v2/_catalog
     ```
