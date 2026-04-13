---
title: TypeScript manifests
permalink: usage/deploy/typescript.html
---

> Note that TypeScript rendering is an experimental feature. To enable it, set the environment variable `NELM_FEAT_TYPESCRIPT=true`.

## TypeScript rendering features

In addition to the standard way of describing Kubernetes manifests via [Helm templates]({{ "/usage/deploy/templates.html" | true_relative_url }}), werf supports describing manifests in TypeScript. Using TypeScript gives you:

- type safety — a typo in a parameter name is caught while writing code, not during deployment;
- IDE support — autocompletion, code navigation, and refactoring;
- standard debugging — error messages include stack traces and point to the exact location in code;
- standard syntax — regular functions, loops, and conditionals instead of template engine constructs.

TypeScript code receives the same context as Helm templates (`Values`, `Release`, `Chart`, etc.) and can coexist with them in the same chart.

## Basic example

Initialize a TypeScript structure in an existing chart:

```shell
werf chart ts init
```

This adds a `ts/` subdirectory to the chart directory with a ready-made example containing a `Deployment` and a `Service`.

Now render the chart:

```shell
werf render
```

TypeScript manifests are generated and merged with the output of Helm templates from `templates/`.

### Comparison with Helm templates

In Helm templates:

{% raw %}

```yaml
# .helm/templates/deployment.yaml:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ $.Release.Name }}-myapp
  labels:
    app.kubernetes.io/name: {{ $.Chart.Name }}
    app.kubernetes.io/instance: {{ $.Release.Name }}
spec:
  replicas: {{ $.Values.replicaCount | default 1 }}
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ $.Chart.Name }}
      app.kubernetes.io/instance: {{ $.Release.Name }}
  template:
    metadata:
      labels:
        app.kubernetes.io/name: {{ $.Chart.Name }}
        app.kubernetes.io/instance: {{ $.Release.Name }}
    spec:
      containers:
      - name: {{ $.Release.Name }}-myapp
        image: {{ $.Values.image.repository }}:{{ $.Values.image.tag | default "latest" }}
        ports:
        - name: http
          containerPort: {{ $.Values.service.port | default 80 }}
```

{% endraw %}

In TypeScript:

```typescript
// .helm/ts/src/deployment.ts:
import type { WerfRenderContext } from '@nelm/chart-ts-sdk';

export function newDeployment($: WerfRenderContext): object {
  const name = `${$.Release.Name}-myapp`;

  return {
    apiVersion: 'apps/v1',
    kind: 'Deployment',
    metadata: {
      name,
      labels: {
        'app.kubernetes.io/name': $.Chart.Name,
        'app.kubernetes.io/instance': $.Release.Name,
      },
    },
    spec: {
      replicas: $.Values.replicaCount ?? 1,
      selector: {
        matchLabels: {
          'app.kubernetes.io/name': $.Chart.Name,
          'app.kubernetes.io/instance': $.Release.Name,
        },
      },
      template: {
        metadata: {
          labels: {
            'app.kubernetes.io/name': $.Chart.Name,
            'app.kubernetes.io/instance': $.Release.Name,
          },
        },
        spec: {
          containers: [
            {
              name,
              image: ($.Values.image?.repository ?? 'nginx') + ':' + ($.Values.image?.tag ?? 'latest'),
              ports: [{ name: 'http', containerPort: $.Values.service?.port ?? 80 }],
            },
          ],
        },
      },
    },
  };
}
```

Both produce identical manifests.

## TypeScript chart structure

After running `werf chart ts init`, the chart gets a `ts/` directory:

```
.helm/
  templates/              # Helm templates (as usual)
    deployment.yaml
    _helpers.tpl
  ts/                     # TypeScript manifests
    src/
      index.ts            # Entry point
      helpers.ts          # Helper functions
      deployment.ts       # Deployment generator
      service.ts          # Service generator
    deno.json             # Deno configuration and dependencies
    tsconfig.json         # TypeScript configuration
    input.example.yaml    # Example rendering context for local development
  values.yaml
  Chart.yaml
```

### Entry point

werf looks for the entry point in this order: `ts/src/index.ts`, then `ts/src/index.js`. If neither file exists, TypeScript rendering is skipped for that chart.

The `index.ts` file calls `render` with a function that accepts the rendering context and returns an array of manifests:

```typescript
// .helm/ts/src/index.ts:
import { render, WerfRenderContext, RenderResult } from '@nelm/chart-ts-sdk';
import { newDeployment } from './deployment.ts';
import { newService } from './service.ts';

function generate($: WerfRenderContext): RenderResult {
  const manifests: object[] = [];

  manifests.push(newDeployment($));

  if ($.Values.service?.enabled !== false) {
    manifests.push(newService($));
  }

  return { manifests };
}

await render(generate);
```

The function passed to `render` receives the rendering context `$` of type `WerfRenderContext` and returns a `RenderResult` object with a `manifests` array of Kubernetes manifests as plain JavaScript objects. Each object is serialized to YAML.

## Rendering context

TypeScript code receives the same context as Helm templates. When initialized via `werf chart ts init`, the context is passed as a typed `WerfRenderContext` object from the [`@nelm/chart-ts-sdk`](https://github.com/werf/nelm-chart-ts-sdk) package. This package is installed automatically during chart initialization and contains the types and the `render` function for running rendering.

The `WerfRenderContext` type extends the base `RenderContext`, adding typed access to werf service parameters at `$.Values.global.werf`.

| Field | Type | Description | Helm template equivalent |
|-------|------|-------------|--------------------------|
| `$.Values` | `WerfServiceValues` | Chart parameters with typed werf service values | `$.Values` |
| `$.Release` | `Release` | Release information | `$.Release` |
| `$.Chart` | `ChartMetadata` | Metadata from Chart.yaml | `$.Chart` |
| `$.Capabilities` | `Capabilities` | Cluster capabilities (API versions, Kubernetes version) | `$.Capabilities` |
| `$.Files` | `Record<string, Uint8Array>` | Additional chart files (not from `templates/` or `ts/`) | `$.Files` |

### The `Values` field

The `$.Values` dictionary is built the same way as for Helm templates: from `values.yaml`, `secret-values.yaml`, `--set`, `--values`, and other [parameter sources]({{ "/usage/deploy/values.html" | true_relative_url }}). All parameterization mechanisms work identically.

When using `WerfRenderContext`, typed werf service parameters are available at `$.Values.global.werf`:

```typescript
$.Values.global.werf.name       // project name
$.Values.global.werf.version    // werf version
$.Values.global.werf.repo       // container registry address
$.Values.global.werf.commit     // commit information (hash, date)
$.Values.global.werf.images     // built images with tags and digests
```

**Release:**

```typescript
$.Release.Name        // release name
$.Release.Namespace   // release namespace
$.Release.Revision    // revision number
$.Release.IsInstall   // true if this is the first install
$.Release.IsUpgrade   // true if this is an upgrade
```

**Chart:**

```typescript
$.Chart.Name          // chart name
$.Chart.Version       // chart version
$.Chart.AppVersion    // application version
$.Chart.Description   // description
$.Chart.Keywords      // keywords
$.Chart.Home          // homepage
$.Chart.Sources       // source code links
```

**Capabilities:**

```typescript
$.Capabilities.KubeVersion.Version  // e.g. "v1.28.0"
$.Capabilities.KubeVersion.Major    // "1"
$.Capabilities.KubeVersion.Minor    // "28"
$.Capabilities.APIVersions          // list of available API versions
```

## Writing manifests

### Conditional resource creation

In Helm templates, conditional resource creation requires wrapping the entire file in an `if` block:

{% raw %}

```
# templates/service.yaml:
{{ if $.Values.service.enabled }}
apiVersion: v1
kind: Service
# ...
{{ end }}
```

{% endraw %}

In TypeScript, it's a regular `if` inside the `render` function:

```typescript
await render(($: WerfRenderContext): RenderResult => {
  const manifests: object[] = [];

  manifests.push(newDeployment($));

  if ($.Values.service?.enabled) {
    manifests.push(newService($));
  }

  return { manifests };
});
```

### Loops and data transformations

In Helm templates, iterating over data uses `range`:

{% raw %}

```yaml
# templates/configmaps.yaml:
{{ range $name, $data := $.Values.configmaps }}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ $name }}
data:
  {{ $data | toYaml | nindent 2 }}
{{ end }}
```

{% endraw %}

In TypeScript, you use standard language features:

```typescript
await render(($: WerfRenderContext): RenderResult => {
  const manifests: object[] = [];

  for (const [name, data] of Object.entries($.Values.configmaps ?? {})) {
    manifests.push({
      apiVersion: 'v1',
      kind: 'ConfigMap',
      metadata: { name },
      data,
    });
  }

  return { manifests };
});
```

### Code reuse

In Helm templates, reuse relies on named templates defined in `_*.tpl` files:

{% raw %}

```
# templates/_helpers.tpl:
{{ define "myapp.labels" }}
app.kubernetes.io/name: {{ $.Chart.Name }}
app.kubernetes.io/instance: {{ $.Release.Name }}
{{ end }}
```

{% endraw %}

In TypeScript, you use regular functions and modules:

```typescript
// ts/src/helpers.ts:
import type { WerfRenderContext } from '@nelm/chart-ts-sdk';

export function getLabels($: WerfRenderContext): Record<string, string> {
  return {
    'app.kubernetes.io/name': $.Chart.Name,
    'app.kubernetes.io/instance': $.Release.Name,
  };
}
```

```typescript
// ts/src/deployment.ts:
import { getLabels } from './helpers.ts';

export function newDeployment($: WerfRenderContext): object {
  return {
    // ...
    metadata: {
      labels: getLabels($),
    },
    // ...
  };
}
```

### Third-party libraries

TypeScript charts support third-party libraries. Install them using `deno install` from the chart's `ts/` directory. For example, to add [kubernetes-models](https://github.com/tommy351/kubernetes-models-ts) for strict Kubernetes resource typing:

```shell
cd .helm/ts
deno install npm:kubernetes-models
```

The dependency is added to the `imports` section of `deno.json` automatically.

```typescript
// .helm/ts/src/deployment.ts:
import { Deployment } from 'kubernetes-models/apps/v1';

export function newDeployment($: WerfRenderContext): object {
  return new Deployment({
    metadata: { name: 'myapp' },
    spec: {
      replicas: $.Values.replicaCount ?? 1,
      selector: { matchLabels: { app: 'myapp' } },
      template: {
        metadata: { labels: { app: 'myapp' } },
        spec: {
          containers: [{ name: 'myapp', image: 'nginx:latest' }],
        },
      },
    },
  }).toJSON();
}
```

With such libraries, the IDE suggests valid resource fields, their types, and which ones are required — manifest structure errors are caught before rendering.

## Using TypeScript alongside Helm templates

TypeScript manifests and Helm templates can coexist in the same chart. The output of both rendering approaches is merged before deployment.

This makes incremental migration straightforward:

1. Add a TypeScript structure to an existing chart: `werf chart ts init`.
2. Move one resource from `templates/` to `ts/src/`.
3. Verify that `werf render` produces the expected output.
4. Repeat for the remaining resources.

> Note that TypeScript manifests and Helm templates are independent resource sources. Named templates (`define`/`include`) from `templates/_*.tpl` are not available in TypeScript code, and vice versa.

### Subchart dependencies

Dependent charts can also contain TypeScript code. Rendering is recursive: for each dependent chart that has an entry point (`ts/src/index.ts` or `ts/src/index.js`), TypeScript rendering runs as well.

Rebuilding the bundle from source is only possible for local dependent charts. If a dependent chart was downloaded from a repository but already contains a prebuilt `ts/dist/bundle.js` (for example, one built during publishing via `werf bundle publish`), rendering uses that file.

The same rules that apply to Helm templates apply to dependent charts: `Values` are scoped to the dependent chart, while release information, cluster capabilities, and runtime data are inherited from the parent chart.

## Building and distribution

### Automatic build

When running `werf converge`, `werf render`, `werf lint`, and `werf plan`, TypeScript code is built automatically. There's no need to run a build command manually — the bundle is assembled in memory and passed to Deno for execution.

### Explicit build

To explicitly build TypeScript code into a JavaScript bundle, run:

```shell
werf chart ts build
```

The resulting bundle is written to `ts/dist/bundle.js`. This is useful for debugging or preparing a chart for manual publishing. Note that due to [giterminism]({{ "/usage/project_configuration/giterminism.html" | true_relative_url }}), the generated file must be committed to Git, or you must use the `--dev` flag.

By default, if `ts/dist/bundle.js` already exists, rendering uses it as-is. To force a rebuild from source, use the `--ignore-bundle-js` flag:

```shell
werf render --ignore-bundle-js
```

### Publishing and applying bundles

When publishing a bundle, the TypeScript bundle is built and included in the package automatically:

```shell
werf bundle publish --repo registry.example.org/myapp
```

When applying a bundle, TypeScript rendering also runs automatically:

```shell
werf bundle apply --repo registry.example.org/myapp --env production
```

Render manifests from a bundle without applying:

```shell
werf bundle render --repo registry.example.org/myapp --env production
```

Preview changes before applying a bundle:

```shell
werf bundle plan --repo registry.example.org/myapp --env production
```

## Activation and runtime environment

TypeScript rendering is experimental and disabled by default. To enable it, set the environment variable:

```shell
export NELM_FEAT_TYPESCRIPT=true
```

Without this variable:
- `werf chart ts init` and `werf chart ts build` will exit with an error;
- TypeScript rendering during `werf converge`, `werf render`, and other commands will be skipped silently.

### Deno

TypeScript code runs on the [Deno](https://deno.com/) runtime. The Deno binary is downloaded automatically on first use and cached locally.

If automatic download isn't available (for example, in an air-gapped environment), specify the path to a pre-installed binary:

```shell
werf render --deno-binary-path /usr/local/bin/deno
```

On the first run of `werf render`, `werf converge`, or another command that processes the chart, Deno creates a `deno.lock` file in the `ts/` directory. This file is also updated when dependencies are added or changed. After initializing a chart or modifying `deno.json`, it is recommended to run `deno install` from the `ts/` directory to install dependencies, update `deno.lock`, and ensure proper IDE support.

Due to [giterminism]({{ "/usage/project_configuration/giterminism.html" | true_relative_url }}), the `deno.lock` file must be committed to Git, otherwise subsequent werf runs will fail. Alternatively, you can use the `--dev` flag.

### Sandbox restrictions

TypeScript code runs in an isolated sandbox with strict restrictions:

- **no network access** — HTTP requests and sockets are not allowed;
- **no environment variable access** — `Deno.env` is unavailable;
- **no process execution** — `Deno.run` is unavailable;
- **restricted filesystem** — access is limited to the data exchange files between werf and Deno.

This ensures manifest rendering stays deterministic and secure.
