---
title: Setup Minikube
sidebar: documentation
permalink: documentation/reference/development_and_debug/setup_minikube.html
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

1. Install [minikube](https://github.com/kubernetes/minikube).
2. Start minikube:

   {% raw %}
   ```shell
   minikube start
   ```
   {% endraw %}

3. Enable minikube registry addon:

   {% raw %}
   ```shell
   minikube addons enable registry
   ```
   {% endraw %}

4. Run custom `werf-registry` service with the binding to 5000 port:

   {% raw %}
   ```shell
   kubectl -n kube-system expose rc/registry --type=ClusterIP --port=5000 --target-port=5000 --name=werf-registry --selector='actual-registry=true'
   ```
   {% endraw %}

5. Setup registry domain in minikube VM:

   {% raw %}
   ```shell
   export REGISTRY_IP=$(kubectl -n kube-system get svc/werf-registry -o=template={{.spec.clusterIP}})
   minikube ssh "echo '$REGISTRY_IP werf-registry.kube-system.svc.cluster.local' | sudo tee -a /etc/hosts"
   ```
   {% endraw %}

6. Setup registry domain in host system:

   {% raw %}
   ```shell
   echo "127.0.0.1 werf-registry.kube-system.svc.cluster.local" | sudo tee -a /etc/hosts
   ```
   {% endraw %}

7. Run port forwarder on host system in a separate terminal:

   {% raw %}
   ```shell
   kubectl port-forward --namespace kube-system service/werf-registry 5000
   ```
   {% endraw %}

8. Check connectivity on host system:

   {% raw %}
   ```shell
   curl -X GET werf-registry.kube-system.svc.cluster.local:5000/v2/_catalog
   ```
   {% endraw %}
