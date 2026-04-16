---
title: TypeScript templates
permalink: usage/deploy/typescript.html
---

> **Note**: TypeScript templates are an experimental feature. To enable it, set the environment variable `NELM_FEAT_TYPESCRIPT=true`.

## Overview

In addition to [Helm templates]({{ "/usage/deploy/templates.html" | true_relative_url }}), werf supports describing Kubernetes manifests in TypeScript. Helm templates and TypeScript templates can coexist in the same chart — the output of both engines is merged into a single YAML document.

TypeScript templates work out of the box: deploying a chart that contains `ts/` requires no additional tools or configuration — werf downloads the Deno runtime automatically and builds the bundle in memory.

### Why TypeScript

Helm's templating language works well for simple cases but becomes hard to maintain as chart complexity grows: deep nesting, lack of type checking, no IDE support, and no way to write tests. TypeScript solves these problems while keeping the same deployment workflow.

### Features

- Type safety and IDE autocompletion via `@nelm/chart-ts-sdk`.
- Standard syntax — regular functions, loops, and conditionals instead of template engine constructs.
- Third-party libraries (e.g. [kubernetes-models](https://github.com/tommy351/kubernetes-models-ts) for strict Kubernetes resource typing).
- Testing — ability to cover manifest generation logic with regular TypeScript tests.
- Security — code runs in an isolated sandbox via [Deno](https://deno.com/) runtime with no access to the network, environment variables, or process execution. Filesystem access is limited to data exchange files between werf and Deno.

## Quick start

Run the following command in your chart directory to initialize TypeScript templates files:

```shell
werf chart ts init
```

A `ts/` subdirectory with a ready-made example appears in the chart directory. Try modifying `ts/src/deployment.ts` — for example, change the number of replicas — then preview the result:

```shell
werf render --dev
```

Deploy:

```shell
werf converge --dev
```

## Chart structure

{% tree_file_viewer 'examples/ts/example-chart' default_file='ts/src/index.ts' expanded=true %}

## Development flow

Install [Deno](https://docs.deno.com/runtime/getting_started/installation/) and follow the [setup guide](https://docs.deno.com/runtime/getting_started/setup_your_environment/) for your IDE/editor (VS Code, JetBrains, Neovim, etc.).

Initialize TypeScript templates in the chart if not already initialized:

```shell
werf chart ts init
```

### Working with codebase

Open the `ts/` directory in your editor as a regular Deno/TypeScript project. You can work with it the same way you would with any TypeScript codebase — run scripts, write tests, use a debugger. Deno provides a rich set of tools for testing, linting, formatting, and more. See [Deno documentation](https://docs.deno.com/runtime/) for details.

The codebase can be organized as you wish. The only requirement is that `ts/src/index.ts` exists and calls `render` from `@nelm/chart-ts-sdk`.

To debug templates rendering, you can use `dev` task from `ts/deno.json`:

```shell
cd .helm/ts
deno task dev
```
This will run `render` function from `ts/src/index.ts` with the example context from `ts/input.example.yaml`. The resulting YAML will be printed to the console below the `Rendered manifests:` message.

### Third-party libraries

Install libraries using `deno add`, for example, try to install [kubernetes-models](https://github.com/tommy351/kubernetes-models-ts):

```shell
deno add npm:kubernetes-models
```

The dependency is added to `deno.json` automatically. Now you can import and use it:

```typescript
// .helm/ts/src/deployment.ts:
import { Deployment } from 'kubernetes-models/apps/v1';

export function newDeployment($: WerfRenderContext): object {
  return new Deployment({
    metadata: { name: 'myapp' },
    spec: {
      // other fields
    },
  }).toJSON();
}
```

To ensure that everything actually works with the werf deno runtime, run:
```shell
werf lint --dev
```

```shell
werf render --dev
```

## Default installation flow

Run `werf converge` — TypeScript is built into a bundle in memory and executed via Deno. The Deno binary is downloaded automatically on first use and cached locally. No extra steps needed.

> **Note**: According to [giterminism policies]({{ "/usage/project_configuration/giterminism.html" | true_relative_url }}), all changed files must be committed.

## Isolated environments installation/distribution flow

For the isolated environments, where Deno cannot be downloaded automatically:

1. Publish the chart:
   ```shell
   werf bundle publish --repo example.org/mycompany/myapp
   ```
   This includes the prebuilt TypeScript bundle in the package.

2. On the target machine with an isolated environment (no network access), download Deno manually and run:
   ```shell
   werf bundle apply --repo example.org/mycompany/myapp --deno-binary-path /usr/local/bin/deno
   ```
   Where `/usr/local/bin/deno` is the path to the local Deno binary. The TypeScript engine skips the build step and uses the prebuilt bundle for rendering.

## SDK API overview

TypeScript engine uses the [@nelm/chart-ts-sdk](https://github.com/werf/nelm-chart-ts-sdk) package.

### `render(generate)`

The entry point must call `render()`, passing a handler function that receives the render context and returns a `RenderResult`:

```typescript
await render(generate);
```

Here `generate` could be a regular function or an async function.

### `WerfRenderContext`

The `generate` function receives `$` of type `WerfRenderContext` — the same context as in Helm templates:

| Field | Type | Description |
|-------|------|-------------|
| `$.Values` | `WerfServiceValues` | Chart parameters + service values at `$.Values.global.werf` |
| `$.Release` | `Release` | Release information |
| `$.Chart` | `ChartMetadata` | Metadata from Chart.yaml |
| `$.Capabilities` | `Capabilities` | Cluster capabilities (API versions, Kubernetes version) |
| `$.Files` | `Record<string, Uint8Array>` | Raw chart files (except `templates/` and `ts/`) |

See the example context in `ts/input.example.yaml`. For details on parameters and how they are constructed, see [Parametrize templates]({{ "/usage/deploy/values.html" | true_relative_url }}).

### `RenderResult`

The generate function returns `RenderResult` — an object with a `manifests` array. Each element is a plain JavaScript object representing a Kubernetes resource. Example output:

```json
{
  "manifests": [
    {
      "apiVersion": "apps/v1",
      "kind": "Deployment",
      "metadata": { "name": "myapp" },
      "spec": { "..." }
    },
    {
      "apiVersion": "v1",
      "kind": "Service",
      "metadata": { "name": "myapp" },
      "spec": { "..." }
    }
  ]
}
```

Each object is serialized to YAML and included in the final rendered output.
