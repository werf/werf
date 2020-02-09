---
title: Matrix tests
sidebar: documentation
permalink: documentation/reference/deploy_process/matrix_tests.html
author: Maksim Nabokikh <maksim.nabokikh@flant.com>
---

Matrix Tests
============
Matrix tests are used to validate all variations of helm values.

## Problem
Sometimes user set helm values in CI deploy job with `werf deploy --set` command.

Gitlab example:
```yaml
Deploy Production:
  stage: tests
  script:
    - type multiwerf && source <(multiwerf use 1.0 beta)
    - type werf && source <(werf ci-env gitlab --tagging-strategy tag-or-branch --verbose)
    - werf deploy --stages-storage :local --namespace ${CI_ENVIRONMENT_SLUG}
  variables:
    WERF_ENV: prod
```
Helm template `global.env` value helps us to determine is additional port should be exposed for my-app Service.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-app
  labels:
    run: my-app
spec:
  ports:
  - port: 80
    protocol: TCP
    name: http
{{- eq if .Values.global.env "prod" }}
    - port: 9090
      protocol: TCP
      name: pprof
{{- end }}
  selector:
    run: my-app
``` 
Oh, we made a mistake that only refers to production environment. So, we have a great chance to catch it only on production deploy process.

Matrix tests helps us to create the matrix that contains all dynamic values variations to cover render errors for any combination of values. 

## Usage

Create the file named matrix_test.yaml in your chart directory (.helm).

Example for 1x1 matrix (matrix_test.yaml):
```yaml
replicas: 1
```

To increase your matrix dimensional you should use special pointers:

**NOTE:** Matrix tests generate values.yaml files, which will be combined with values.yaml in your chart directory.
Only dynamic values should be described in matrix_tests.yaml file.

#### Variations

Variation allows splitting your values.yaml.

* `__ConstantVariation__` - array of items you split values.yaml with.

Example values_matrix_test.yaml:
```yaml
replicas: 
  __ConstantVariation__: [1, 2]
```

It creates two different values.yaml files for tests:
```yaml
replicas: 1
```
```yaml
replicas: 2
```

You can add additional variation to create more variants:

Example matrix_test.yaml:
```yaml
replicas: 
  __ConstantVariation__: [1, 2]
environments:
  __ConstantVariation__:
  - ["test", "prod"]
  - ["dev", "stage"]
```

It creates four different values.yaml files for tests:
```yaml
replicas: 1
environments:
- test
- prod
```
```yaml
replicas: 2
environments:
- test
- prod
```
```yaml
replicas: 1
environments:
- dev
- stage
```
```yaml
replicas: 2
environments:
- dev
- stage
```

#### Items

Items are special pointers that can be used in matrix_test.yaml to override objects.

* `__EmptyItem__` - completely delete an object from map

Example matrix_test.yaml:
```yaml
image: 
  __ConstantVariation__: [registry.example.com, __EmptyItem__]
```
Generated values.yaml:
```yaml
image: registry.example.com
```
```yaml
```
