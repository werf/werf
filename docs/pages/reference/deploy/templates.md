---
title: Release configuration
sidebar: reference
permalink: reference/deploy/templates.html
author: Artem Kladov <artem.kladov@flant.com>
---

In the course of the release, werf calls helm that uses the chart in the `.helm` folder in the project root for release configuration. Helm searches for YAML descriptions for kubernetes objects in the chart's `templates` folder and processes each file using the GO template rendering engine. Werf does not directly process charts and templates. It uses helm for this.

For more information:
* [GO templates description language](https://godoc.org/text/template)
* [Sprig reference](https://godoc.org/github.com/Masterminds/sprig) – a library that helm uses for GO template rendering
* [Advanced functions](https://docs.helm.sh/developing_charts/#chart-development-tips-and-tricks) added to helm for templates like include and required

You can create a chart structure using the `werf kube chart create` command, applying it from the project root directory. As a result, the `.helm` folder is created with the chart containing a sample description of kubernetes objects.

## Passing parameters

In the course of an application release, there is many ways of passing parameters:
* using the `values.yaml` or `secret-values.yaml` files. Both methods are identical for access. The only difference is in the storage method; values of variables are encrypted in the `secret-values.yaml` file. Further, in the text, this difference will be ignored.
* using the --set parameter in `werf kube deploy` commands (see [section](deploy_to_kubernetes.html))
* using environment variables.

The `values.yaml` and `secret-values.yaml` files contain a description of the variables that are available in the templates. For instance, we have a `values.yaml` file:

```yaml
replicas:
  production: 3
  staging: 2

db:
  host:
    production: prod-db.mycompany.int
    stage: stage-db
    _default: 192.168.6.3
  username:
    production: user-production
    stage: user-stage
  database:
    production: productiondb
    stage: stagedb
```

Then you can address the appropriate variables in the template using a construction like – {% raw %}`{{ .Values.db.username.production }}`{% endraw %}.

Werf sets and uses many variables that are also available in the templates. You can retrieve their values using the `werf kube value get VALUE_KEY` command.

For instance, you can perform the following actions to retrieve values for all variables:
```bash
werf kube value get .
```

## Features of chart template creation

Chart templates are published to the `templates` directory as YAML files.

Werf provides the following additional templates to be used:
* `werf_container_image`
* `werf_container_env`

### The `werf_container_image` template

This replaced the werf `dimg` template that was previously used in outdated versions. The template generates `image` and `imagePullPolicy` keys for the pod container.

A specific feature of the template is that `imagePullPolicy` is generated based on the `.Values.global.werf.is_branch` value, if tags are used, `imagePullPolicy: Always` is not set.

The template may return multiple strings, which is why it must be used together with the `indent` construction.

The logic of generating the `imagePullPolicy` key:
* The `.Values.global.werf.is_branch=true` value means that an image is being deployed based on the `latest` logic for a branch.
  * In this case, the image for an appropriate docker tag must be updated through docker pull, even if it already exists, to get the current `latest` version of the respective tag.
  * In this case – `imagePullPolicy=Always`.
* The `.Values.global.werf.is_branch=false` value means that a tag or a specific image commit is being deployed.
  * In this case, the image for an appropriate docker tag doesn't need to be updated through docker pull if it already exists.
  * In this case, `imagePullPolicy` is not specified, which is consistent with the default value currently adopted in kubernetes: `imagePullPolicy=IfNotPresent`.

An example of using a template in case multiple dimgs exist in the config:
* `tuple <dimg-name> . | include "werf_container_image" | indent <N-spaces>`

An example of using a template in case a single unnamed dimg exists in config:
* `tuple . | include "werf_container_image" | indent <N-spaces>`
* `include "werf_container_image" . | indent <N-spaces>` (additional simplified entry format)

### The `werf_container_env` template

Enables streamlining the release process if the image remains unchanged. Generates a block with the `DOCKER_IMAGE_ID` environment variable for the pod container, but only if `.Values.global.werf.is_branch=true`, because in this case the image for an appropriate docker tag might have been updated through its name remained unchanged. The `DOCKER_IMAGE_ID` variable contains a new id docker for an image, which forces kubernetes to update an asset. The template may return multiple strings, which is why it must be used together with `indent`.

An example of using a template in case multiple dimgs exist in the config:
* `tuple <dimg-name> . | include "werf_container_env" | indent <N-spaces>`

An example of using a template in case a single unnamed dimg exists in config:
* `tuple . | include "werf_container_env" | indent <N-spaces>`
* `include "werf_container_env" . | indent <N-spaces>` (additional simplified entry format)

## Example of configuration

A sample description of an application configuration that comprises frontend, backend, and db containers representing werf template use.

> This example covers only the crucial aspect of the configuration description. Based on your cluster configuration, to launch an application you might additionally be required to create such resources as an Ingress, a Secret (if you're working with a private repository), or a Service, or ensuring traffic routing, and so on.

Chart.yaml
```yaml
name: example-werf-deploy
version: 0.1.0
```

values.yaml
```yaml
replicas:
  production: 3
  staging: 1
```

werf.yaml
```yaml
dimg: "frontend"
from: "nginx"
---
dimg: "backend"
from: "alpine"
---
dimg: "db"
from: "mysql"
```

app.yaml
{% raw %}
```yaml
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: {{ .Chart.Name }}-frontend
spec:
  replicas: {{ .Values.replicas.production }}
  template:
    spec:
      containers:
        - name: frontend
{{ tuple "frontend" . | include "werf_container_image" | indent 10 }}
          env:
            - name: VAR1
              value: value
{{ tuple "frontend" . | include "werf_container_env" | indent 12 }}
            - name: VAR2
              value: value
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: {{ .Chart.Name }}-backend
spec:
  template:
    spec:
      containers:
        - name: backend
{{ tuple "backend" . | include "werf_container_image" | indent 10 }}
          env:
{{ tuple "backend" . | include "werf_container_env" | indent 12 }}
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: {{ .Chart.Name }}-db
spec:
  template:
    spec:
      containers:
        - name: db
{{ tuple "db" . | include "werf_container_image" | indent 10 }}
          env:
{{ tuple "db" . | include "werf_container_env" | indent 12 }}
```
{% endraw %}

Result:
```yaml
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: example-werf-deploy-frontend
spec:
  replicas: 3
  template:
    spec:
      containers:
        - name: frontend
          image: localhost:5000/example-werf-deploy/frontend:latest
          imagePullPolicy: Always
          env:
            - name: VAR1
              value: value
            - name: DOCKER_IMAGE_ID
              value: sha256:7a126ea38f24d3ca98207d28414e4f6ae5ae30458539828a125d029dea8a93cb
            - name: VAR2
              value: value
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: example-werf-deploy-backend
spec:
  template:
    spec:
      containers:
        - name: backend
          image: localhost:5000/example-werf-deploy/backend:latest
          imagePullPolicy: Always
          env:
            - name: DOCKER_IMAGE_ID
              value: sha256:b325c0788c80efb7a08e6eb7c04e11e412a035c9d39a1430260d776421ea1a4a
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: example-werf-deploy-db
spec:
  template:
    spec:
      containers:
        - name: db
          image: localhost:5000/example-werf-deploy/db:latest
          imagePullPolicy: Always
          env:
            - name: DOCKER_IMAGE_ID
              value: sha256:da27cdffc4fcafaa4f6ced8b3bc1409191b9876b9b75c0e91ddffaceba5b497c

```

The `.Chart.Name` mentioned in the configuration is a `name` key value from `Chart.yaml`, and the `.Values.replicas.production` value is retrieved from `values.yaml`.
