---
title: TypeScript templates
permalink: usage/deploy/typescript.html
---

> Note that TypeScript templates are an experimental feature. To enable them, set the environment variable `NELM_FEAT_TYPESCRIPT=true`.

## Features

In addition to the standard way of describing Kubernetes manifests via [Helm templates]({{ "/usage/deploy/templates.html" | true_relative_url }}), werf supports describing manifests in TypeScript:

- type safety and IDE autocompletion;
- third-party libraries (e.g. [kubernetes-models](https://github.com/tommy351/kubernetes-models-ts) for strict Kubernetes resource typing);
- standard syntax — regular functions, loops, and conditionals instead of template engine constructs;
- testing — ability to cover manifest generation logic with regular tests;
- security — code runs in an isolated sandbox with no access to the network, environment variables, or filesystem.

TypeScript code receives the same [root context]({{ "/usage/deploy/values.html" | true_relative_url }}) as Helm templates (`Values`, `Release`, `Chart`, etc.) and can coexist with Helm templates in the same chart.

## Quick start

Initialize TypeScript files in an existing chart:

```shell
werf chart ts init
```

A `ts/` subdirectory with a ready-made example appears in the chart directory. Edit the generated files in `ts/src/` and run:

```shell
werf converge
```

Manifests are generated from the `ts/` directory by the TypeScript engine and from the `templates/` directory by the Helm templating engine, then merged together as one YAML multidoc document.

### Structure

```
.helm/
  templates/              # Helm templates (as usual)
  ts/                     # TypeScript templates
    src/
      index.ts            # Entry point
      helpers.ts          # Helper functions
      deployment.ts       # Deployment generator
      service.ts          # Service generator
    deno.json             # Deno configuration and dependencies
    tsconfig.json         # TypeScript configuration
  values.yaml
  Chart.yaml
```

werf looks for the entry point in this order: `ts/src/index.ts`, then `ts/src/index.js`. If neither file exists, TypeScript rendering is skipped for that chart.

### Entry point

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

The `generate` function receives the root context `$` of type `WerfRenderContext` and returns a `RenderResult` with a manifests array — plain JavaScript objects that are serialized to YAML.

## Root context

TypeScript code receives the same context as Helm templates via the variable `$` of the `WerfRenderContext` type:

| Field | Type | Description |
|-------|------|-------------|
| `$.Values` | `WerfServiceValues` | Chart parameters with typed werf service values |
| `$.Release` | `Release` | Release information |
| `$.Chart` | `ChartMetadata` | Metadata from Chart.yaml |
| `$.Capabilities` | `Capabilities` | Cluster capabilities (API versions, Kubernetes version) |
| `$.Files` | `Record<string, Uint8Array>` | Raw chart files (except `templates/` and `ts/`) |

For details on parameters and how they are constructed, see [Parametrize templates]({{ "/usage/deploy/values.html" | true_relative_url }}).

### werf service values

When using `WerfRenderContext`, typed werf service parameters are available at `$.Values.global.werf`:

```typescript
$.Values.global.werf.name       // project name
$.Values.global.werf.version    // werf version
$.Values.global.werf.repo       // container registry address
$.Values.global.werf.commit     // commit information (hash, date)
$.Values.global.werf.images     // built images with tags and digests
```

## Third-party libraries

Install libraries using `deno install` from the chart's `ts/` directory:

```shell
cd .helm/ts
deno install npm:kubernetes-models
```

The dependency is added to `deno.json` automatically. Usage example:

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

## Building and distribution

When running `werf converge`, `werf render`, `werf lint`, and `werf plan`, TypeScript code is built automatically — the bundle is assembled in memory and passed to Deno for execution.

To explicitly build a JavaScript bundle:

```shell
werf chart ts build
```

The resulting bundle is written to `ts/dist/bundle.js`. Note that due to [giterminism]({{ "/usage/project_configuration/giterminism.html" | true_relative_url }}), this file must be committed to Git, or you must use the `--dev` flag.

If `ts/dist/bundle.js` already exists, rendering uses it as-is. To force a rebuild from source:

```shell
werf render --ignore-bundle-js
```

### Bundles

When publishing a bundle, the TypeScript bundle is built and included in the package automatically:

```shell
werf bundle publish --repo registry.example.org/myapp
```

When applying a bundle, TypeScript rendering runs automatically:

```shell
werf bundle apply --repo registry.example.org/myapp --env production
```

Render manifests from a bundle without applying:

```shell
werf bundle render --repo registry.example.org/myapp --env production
```

Preview changes before applying:

```shell
werf bundle plan --repo registry.example.org/myapp --env production
```

### Dependent charts

Dependent charts can also contain TypeScript code — rendering is recursive. Rebuilding the bundle from source is only possible for local dependent charts. If a dependent chart was downloaded from a repository but contains a prebuilt `ts/dist/bundle.js` (e.g. built during publishing via `werf bundle publish`), rendering uses that file.

## Activation and runtime environment

The feature is experimental and disabled by default:

```shell
export NELM_FEAT_TYPESCRIPT=true
```

Without this variable, `werf chart ts init` and `werf chart ts build` will exit with an error, while TypeScript rendering during `werf converge`, `werf render`, and other commands will be skipped silently.

### Deno

TypeScript code runs on the [Deno](https://deno.com/) runtime. The Deno binary is downloaded automatically on first use and cached locally. If automatic download isn't available (e.g. in an air-gapped environment), specify the path to a pre-installed binary:

```shell
werf render --deno-binary-path /usr/local/bin/deno
```

On the first run of `werf render`, `werf converge`, or another command that processes the chart, Deno creates a `deno.lock` file in the `ts/` directory. This file is also updated when dependencies are added or changed. After initializing a chart or modifying `deno.json`, it is recommended to run `deno install` from the `ts/` directory to install dependencies, update `deno.lock`, and ensure proper IDE support.

Due to [giterminism]({{ "/usage/project_configuration/giterminism.html" | true_relative_url }}), the `deno.lock` file must be committed to Git, otherwise subsequent werf runs will fail. Alternatively, you can use the `--dev` flag.

### Sandbox

TypeScript code runs in an isolated sandbox: no access to the network, environment variables, or process execution; filesystem access is limited to data exchange files between werf and Deno. This ensures deterministic and secure rendering.
