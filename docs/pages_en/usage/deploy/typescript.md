---
title: TypeScript templates
permalink: usage/deploy/typescript.html
---

> Note: TypeScript templates are an experimental feature. To enable it, set the environment variable `NELM_FEAT_TYPESCRIPT=true`.

## Features

In addition to the standard way of describing Kubernetes manifests via [Helm templates]({{ "/usage/deploy/templates.html" | true_relative_url }}), werf supports describing manifests in TypeScript:

- type safety and IDE autocompletion;
- third-party libraries (e.g. [kubernetes-models](https://github.com/tommy351/kubernetes-models-ts) for strict Kubernetes resource typing);
- standard syntax — regular functions, loops, and conditionals instead of template engine constructs;
- testing — ability to cover manifest generation logic with regular tests;
- security — code runs in an isolated sandbox via [Deno](https://deno.com/) runtime with no access to the network, environment variables, or process execution. Filesystem access is limited to data exchange files between werf and Deno.

## How it works

Manifests are generated from the `ts/` directory by the TypeScript engine and from the `templates/` directory by the Helm templating engine, then merged together as one YAML multidoc document.

## Quick start

Initialize TypeScript files in an existing chart:

```shell
werf chart ts init
```

A `ts/` subdirectory with a ready-made example appears in the chart directory. Edit the generated files in `ts/src/` and run:

```shell
werf converge --dev
```

### Chart structure

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
  values.yaml
  Chart.yaml
```

Example of TypeScript templates:

```typescript
// .helm/ts/src/deployment.ts:
import type { WerfRenderContext } from '@nelm/chart-ts-sdk';
import { getFullname, getLabels, getSelectorLabels } from './helpers.ts';

export function newDeployment($: WerfRenderContext): object {
  const name = getFullname($);

  return {
    apiVersion: 'apps/v1',
    kind: 'Deployment',
    metadata: {
      name,
      labels: getLabels($),
    },
    spec: {
      replicas: $.Values.replicaCount ?? 1,
      selector: {
        matchLabels: getSelectorLabels($),
      },
      template: {
        metadata: {
          labels: getSelectorLabels($),
        },
        spec: {
          containers: [
            {
              name: name,
              image: ($.Values.image?.repository ?? 'nginx') + ':' + ($.Values.image?.tag ?? 'latest'),
              ports: [
                {
                  name: 'http',
                  containerPort: $.Values.service?.port ?? 80,
                },
              ],
            },
          ],
        },
      },
    },
  };
}
```

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

The entry point (`ts/src/index.ts` or `index.js`) must call `render` from [@nelm/chart-ts-sdk](https://github.com/werf/nelm-chart-ts-sdk). `generate` takes `$` (`WerfRenderContext`) and returns `RenderResult` — a list of JavaScript objects serialized to YAML. If no entry point is found, TypeScript rendering is skipped.

## Root context

TypeScript code receives the same [root context]({{ "/usage/deploy/values.html" | true_relative_url }}) as Helm templates via the variable `$` of the `WerfRenderContext` type:

| Field | Type | Description                                                           |
|-------|------|-----------------------------------------------------------------------|
| `$.Values` | `WerfServiceValues` | Chart parameters + service values at `$.Values.global.werf` |
| `$.Release` | `Release` | Release information                                                   |
| `$.Chart` | `ChartMetadata` | Metadata from Chart.yaml                                              |
| `$.Capabilities` | `Capabilities` | Cluster capabilities (API versions, Kubernetes version)               |
| `$.Files` | `Record<string, Uint8Array>` | Raw chart files (except `templates/` and `ts/`)                       |

For details on parameters and how they are constructed, see [Parametrize templates]({{ "/usage/deploy/values.html" | true_relative_url }}).

## Development flow

### Environment Setup

Install [Deno](https://docs.deno.com/runtime/getting_started/installation/), then follow the [IDE setup guide](https://docs.deno.com/runtime/getting_started/setup_your_environment/) for your editor (VS Code, JetBrains, Neovim, etc.).

### General flow

1. Install third-party libraries if needed.
2. Edit TypeScript templates in `ts/src/`.
3. Commit changes.
4. Run werf commands (`werf converge`, `werf render`, etc.).

According to [giterminism policies]({{ "/usage/project_configuration/giterminism.html" | true_relative_url }}), any change must be committed to Git. Alternatively, you can use the `--dev` flag and skip the commit step.

### Third-party libraries

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

> Note: changed `deno.json` and `deno.lock` files must be committed to Git.

## Building and distribution

When running `werf converge`, `werf render`, `werf lint`, and `werf plan`, TypeScript code is built automatically — the bundle is assembled in memory and passed to Deno for execution.

To explicitly build a JavaScript bundle:

```shell
werf chart ts build
```

The resulting bundle is written to `ts/dist/bundle.js`. This file must be committed to Git.

If `ts/dist/bundle.js` already exists, rendering uses it as-is. To force a rebuild from source:

```shell
werf render --ignore-bundle-js
```

Bundle-based deployment is also supported — `werf bundle publish` automatically includes the TypeScript bundle in the package, and `werf bundle apply` runs TypeScript rendering.

### Dependent charts

Dependent charts can also contain TypeScript code — rendering is recursive. For local charts in `charts` directory, TypeScript is rebuilt from source as usual. 

For charts downloaded from a repository, only the prebuilt `ts/dist/bundle.js` is available (e.g. built during `werf bundle publish`), so rebuilding from source is not possible.

## Runtime environment

The Deno binary is downloaded automatically on first use and cached locally. If automatic download isn't available (e.g. in an air-gapped environment), specify the path to a pre-installed binary:

```shell
werf render --deno-binary-path /usr/local/bin/deno
```
