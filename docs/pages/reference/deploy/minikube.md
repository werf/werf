---
title: Using Minikube
sidebar: reference
permalink: reference/deploy/minikube.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

To use dapp for deployment of images in minikube:
* Collect required images from the host machine.
* Deploy minikube with docker-registry and proxy on the host machine, see [`dapp kube minikube setup`](#dapp-kube-minikube-setup).
* Upload collected images to docker-registry and specify `:minikube` as the `REPO` parameter through [`dapp dimg push :minikube`]({{ site.baseurl }}/reference/cli/dimg_push.html).
* Apply the kubernetes configuration and specify `:minikube` as the `REPO` parameter through [`dapp kube deploy :minikube`]({{ site.baseurl }}/reference/cli/kube_deploy.html).

### dapp kube minikube setup

```
dapp kube minikube setup
```

* Launches minikube, forces restart if it's already launched.
* Awaits kubernetes cluster readiness in minikube.
* Launches docker registry in minikube.
* Launches proxy for docker-registry at the system address: `localhost:5000`.
  * Proxy forwards directly to pod docker-registry inside minikube.
    * As a result, if pod fails, the setup command needs to be restarted.
