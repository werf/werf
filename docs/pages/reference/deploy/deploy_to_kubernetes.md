---
title: Deploy to kubernetes
sidebar: reference
permalink: reference/deploy/deploy_to_kubernetes.html
author: Timofey Kirillov <timofey.kirillov@flant.com>
---

Dapp uses [helm](https://helm.sh/) kubernetes package manager to deploy applications into kubernetes.

## Installing helm

Before using dapp for deploy, you should install [helm](https://docs.helm.sh/using_helm/#installing-helm) and its back-end part — [tiller](https://docs.helm.sh/using_helm/#installing-tiller).

## Helm chart

The `.helm` directory in the project root describes a helm chart (starting now referred to as `chart`) which in turn provides a description of the application configuration and application component used for further release to the kubernetes cluster using the helm toolset. The chart for dapp, i.e. `.helm` folder, has the following structure:

```
.helm/
  templates/
    <name>.yaml
    <name>.tpl
  charts/
  secret/
  Chart.yaml
  values.yaml
  secret-values.yaml
```

The `Chart.yaml` file describes the application chart, and you must specify at least the name and version of the application. Example of a `Chart.yaml` file:


```
apiVersion: v1
description: Test RabbitMQ chart for Kubernetes
name: rabbit
version: 0.1.0
```

In the `templates` directory, YAML file (chart element) templates are stored, which describe the resources for their publishing in the cluster. Detailed information about template creation is available in a separate [section]({{ site.baseurl }}/reference/deploy/templates.html). The `charts` directory is used when you need to work with external charts.

Chart structure includes additional elements that are not contained within the structure of a standard helm chart – these are the `secret-values.yaml` file and the `secret` directory that are described in detail in the [working with secrets section]({{ site.baseurl }}/reference/deploy/secrets.html).

## Connection settings

Connection to kubernetes is setup using the same configuration file as for kubectl: `~/.kube/config`.

* `current-context` context is used if installed, or any other random context from the `contexts` list.
* The same default kubernetes `namespace` is used as for kubectl: from the namespace field of the current context.
  * If namespace is not set by default in `~/.kube/config`, `namespace=default` is used.

If you are using a self-signed SSL certificate, you can specify it using the `SSL_CERT_FILE` variable.
* Copy the certificate file in any folder and set the necessary permissions. Example:
```bash
cp  CERTFILE.crt ~gitlab-runner/.
chown -R gitlab-runner:gitlab-runner ~gitlab-runner/CERTFILE.crt
```

* Set the environment variable `SSL_CERT_FILE=/home/gitlab-runner/<domain>.crt]` (export it, or set it just before calling `dapp kube deploy`). Example:
```bash
SSL_CERT_FILE=/home/gitlab-runner/CERTFILE.crt dapp kube deploy
      --namespace ${CI_ENVIRONMENT_SLUG}
      --tag-ci
      ${CI_REGISTRY_IMAGE}
```

## Additions to helm

Dapp has several additions to the helm.

### Chart generation

During dapp deploy a temporary helm chart is created.

This chart contains:

* Additional generated go-templates: `dapp_container_image`, `dapp_container_env` and other. These templates are described in [the templates article]({{ site.baseurl }}/reference/deploy/templates.html).
* Decoded secret values yaml file. The secrets are described in [the secrets article]({{ site.baseurl }}/reference/deploy/secrets.html).

Temporary chart then passed to the helm. Dapp deletes this chart on the dapp deploy command termination.

### Watch resources

Dapp watches resources statuses and logs during deploy process. More info is avaiable in the [watch resources article]({{ site.baseurl }}/reference/deploy/watch_resources.html).

## Dapp deploy command

Dapp deploy command starts the helm-chart release process in kubernetes. A release named `<dapp name>-<NAMESPACE>` will be installed or updated by helm.

`DAPP_HELM_RELEASE_NAME` environment variable could be used to specify custom helm release name.

#### Syntax

```
dapp kube deploy REPO [--tag=TAG --tag-branch --tag-commit --tag-build-id --tag-ci] [--namespace=NAMESPACE] [--set=<value>] [--values=<values-path>] [--secret-values=<secret-values-path>]
```

##### `REPO`

Address of the repository, from which images will be retrieved. This parameter must coincide with the parameter specified in [`dapp dimg push`]({{ site.baseurl }}/reference/cli/dimg_push.html).

If a special value `:minikube` is set, the local proxy will be used for docker-registry from minikube, see [using minikube section]({{ site.baseurl }}/reference/deploy/minikube.html).

##### `--tag=TAG --tag-branch --tag-commit --tag-build-id --tag-ci`

Image version from the specified repository. Options are compliant with those specified in the [`dapp dimg push`]({{ site.baseurl }}/reference/cli/dimg_push.html).

##### `--namespace=NAMESPACE`

Use the specified kubernetes namespace. If not specified, the default namespace according to `~/.kube/config`, or, if not specified, `namespace=default` will be used.

##### `--set=<value>`

Passed unchanged to the `helm --set` parameter.

##### `--values=<values-path>`

Passed unchanged to the [`helm --values`](https://github.com/kubernetes/helm/blob/master/docs/chart_template_guide/values_files.md#values-files) parameter.

Enables specifying an additional values yaml file together with the standard `.helm/values.yaml`.

##### `--secret-values=<secret-values-path>`

Enables specifying an additional secret-values yaml file together with the standard `.helm/secret-values.yaml`. For detailed information about secrets, see the [working with secrets section]({{ site.baseurl }}/reference/deploy/secrets.html).

### dapp kube dismiss

Starts the process to remove release `<dapp name>-<NAMESPACE>` from helm.

```
dapp kube dismiss [--namespace=NAMESPACE] [--with-namespace]
```

##### `--namespace=NAMESPACE`

Use the specified kubernetes namespace. If not specified, the default namespace according to `~/.kube/config`, or, if not specified, `namespace=default` will be used.

##### `--with-namespace`

Remove the currently used kubernetes namespace after deleting the release from helm. The namespace is not deleted by default.
