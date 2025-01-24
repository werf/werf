---
title: Authentication and authorization
permalink: usage/deploy/authentication.html
---

## Specify kubeconfig to access Kubernetes

By default we use the `~/.kube/config` file to get access to the Kubernetes cluster. You can specify a different kubeconfig file with the following options:

1. `--kube-config=<path>` or `$WERF_KUBE_CONFIG=<path>`: set the path to the kubeconfig file.
1. `--kube-config-base64=<base64>` or `$WERF_KUBE_CONFIG_BASE64=<base64>`: pass the kubeconfig file encoded in base64 via the command line or the environment variable.

## Override kubeconfig configuration

You can override the kubeconfig configuration with the following options:

1. `--kube-context=<context>` or `$WERF_KUBE_CONTEXT=<context>`: change the kubeconfig context.
1. `--kube-token=<token>` or `$WERF_KUBE_TOKEN=<token>`: set the Kubernetes bearer token.
1. `--kube-api-server=<url>` or `$WERF_KUBE_API_SERVER=<url>`: change the Kubernetes API Server URL.
1. `--kube-tls-server=<url>` or `$WERF_KUBE_TLS_SERVER=<url>`: change the server name used for Kubernetes API certificate validation.
1. `--kube-ca-path=<path>` or `$WERF_KUBE_CA_PATH=<path>`: change the path to the CA certificate file.
1. `--skip-tls-verify-kube=<bool>` or `$WERF_SKIP_TLS_VERIFY_KUBE=<bool>`: should we verify the Kubernetes API server certificate.

## Access Helm charts or werf bundles in a private repository

Use `werf cr login` to log in to the private OCI registry with Helm charts or werf bundles:

```shell
werf cr login -u myuser -p mypassword localhost:5000
```

Alternatively, for the private HTTP Helm chart repository use `werf helm registry login`:

```shell
werf helm registry login -u myuser -p mypassword localhost:5000
```

Now you can pull or push Helm charts or werf bundles from the repository, be it with `werf helm dependency build/update`, `werf helm pull`, `werf bundle publish/apply` or other commands.
