---
title: Resources adoption
permalink: advanced/helm/deploy_process/resources_adoption.html
---

By default Helm and werf can only manage resources that was created by the Helm or werf itself. If you try to deploy chart with the new resource in the `.helm/templates` that already exists in the cluster and was created outside werf then the following error will occur:

```
Error: helm upgrade have failed: UPGRADE FAILED: rendered manifests contain a resource that already exists. Unable to continue with update: KIND NAME in namespace NAMESPACE exists and cannot be imported into the current release: invalid ownership metadata; label validation error: missing key "app.kubernetes.io/managed-by": must be set to "Helm"; annotation validation error: missing key "meta.helm.sh/release-name": must be set to RELEASE_NAME; annotation validation error: missing key "meta.helm.sh/release-namespace": must be set to RELEASE_NAMESPACE
```

This error prevents destructive behaviour when some existing resource would occasionally become part of helm release.

However, if it is the desired outcome, then edit this resource in the cluster and set the following annotations and labels:

```yaml
metadata:
  annotations:
    meta.helm.sh/release-name: "RELEASE_NAME"
    meta.helm.sh/release-namespace: "NAMESPACE"
  labels:
    app.kubernetes.io/managed-by: Helm
```

After the next deploy this resource will be adopted into the new release revision. Resource will be updated to the state defined in the chart manifest.
