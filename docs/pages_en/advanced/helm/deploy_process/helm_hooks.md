---
title: Helm hooks
permalink: advanced/helm/deploy_process/helm_hooks.html
---

The helm hook is an arbitrary Kubernetes resource marked with the `helm.sh/hook` annotation. For example:

```yaml
kind: Job
metadata:
  name: somejob
  annotations:
    "helm.sh/hook": pre-upgrade,pre-install
    "helm.sh/hook-weight": "1"
```

A lot of various helm hooks come into play during the deploy process. We have already discussed `pre|post-install|upgrade` hooks during the [deploy process steps]({{ "/advanced/helm/deploy_process/steps.html" | true_relative_url }}). These hooks are often used to perform tasks such as migrations (in the case of `pre-upgrade` hooks) or some post-deploy actions. The full list of available hooks can be found in the [helm docs](https://helm.sh/docs/topics/charts_hooks/).

Hooks are sorted in the ascending order specified by the `helm.sh/hook-weight` annotation (hooks with the same weight are sorted by the name). After that, hooks are created and executed sequentially. werf by default recreates the Kubernetes resource for each hook if that resource already exists in the cluster. Hooks are remained existing in the Kubernetes cluster after execution.

Created hooks resources will not be deleted after completion, unless there is [special annotation `"helm.sh/hook-delete-policy": hook-succeeded,hook-failed`](https://helm.sh/docs/topics/charts_hooks/).
