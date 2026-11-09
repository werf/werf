---
title: TypeScript templates
permalink: usage/deploy/typescript.html
---

> **Note**: TypeScript templates are an experimental feature. To enable it, set the environment variable `NELM_FEAT_TYPESCRIPT=true`.

## Overview

In addition to [Helm templates]({{ "/usage/deploy/templates.html" | true_relative_url }}), werf can generate Kubernetes manifests with TypeScript. Helm templates and TypeScript templates can coexist in the same chart — resulting manifests are merged into a single multi-doc YAML document.

TypeScript templates are rendered with [Node.js](https://nodejs.org/): deploying a chart that contains a `ts/` directory requires Node.js 26 or newer and a package manager (npm, pnpm or yarn) installed on the host.

### Why TypeScript

Helm's templating language works well for simple cases but becomes hard to maintain as chart complexity grows: primitive language with lots of gotchas, limited library, performance issues, debug difficulties, poor IDE/editor support and so on. TypeScript in werf solves these problems without complicating the deployment workflow.

### Features

- IDE support — full autocompletion, type checking, go-to-definition, and refactoring in any editor with TypeScript support (VS Code, JetBrains, Neovim, etc.).
- Standard syntax — proper functions, loops, and conditionals instead of awkward template engine constructs.
- Pure TypeScript — `ts` directory is a regular Node.js TypeScript project, and can be rendered without werf with `npm run dev`.
- Large ecosystem — TypeScript is one of the most popular languages with extensive documentation, community resources, and tooling.
- Almost any third-party TypeScript/JavaScript library can be used, for example [kubernetes-models](https://github.com/tommy351/kubernetes-models-ts), [cdk8s](https://cdk8s.io/) or any other library from the npm ecosystem.
- Testing — test your code using common TypeScript libraries and tooling.
- Isolated environments — npm modules are bundled into the chart by default, so no network calls will be done during the deployment, except to the Kubernetes itself.
- Security — code runs under the Node.js permission model with no access to the network, environment variables, or process execution. Filesystem access is limited to reading chart files.

## Quick start

Initialize TypeScript files in an existing chart:

```shell
werf chart ts init
```

It will bootstrap the `.helm/ts/` directory, which contains a TypeScript project skeleton and a few files with sample resources. Try modifying `ts/src/deployment.ts` — for example, change the number of replicas — then check the result:

```shell
werf render --dev
```

To deploy:

```shell
werf converge --dev
```

## Chart structure

{% tree_file_viewer 'examples/ts/example-chart' default_file='ts/src/index.ts' expanded=true %}

## Developing a chart with TypeScript templates

Install [Node.js](https://nodejs.org/en/download) 26 or newer and set up TypeScript support in your IDE/editor (VS Code, JetBrains, Neovim, etc.).

Initialize TypeScript files in the chart if not already initialized:

```shell
werf chart ts init
```

Open the `ts/` directory in your editor as a regular Node.js/TypeScript project. You can work with it the same way you would with any TypeScript codebase — run scripts, write tests, use a debugger.

The codebase can be organized as you wish. The only requirement is that `ts/src/index.ts` exists, and `render` function from `@nelm/chart-ts-sdk` **must** be called. Otherwise, no TypeScript rendering happens.

To debug templates rendering in an environment that is very close to how werf runs Node.js, you can use the `dev` script from `ts/package.json`:

```shell
cd .helm/ts
npm run dev
```
TypeScript engine will call `render` function from `ts/src/index.ts` with the example context from `ts/input.example.yaml`. The resulting YAML will be printed to the console below the `Rendered manifests:` message.

Install libraries using `npm install`, for example, try to install [kubernetes-models](https://github.com/tommy351/kubernetes-models-ts) — library for strict Kubernetes resource typing:

```shell
npm install kubernetes-models
```

The dependency is added to `package.json` automatically. Now you can import and use it:

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

To ensure that everything actually works when werf renders the chart, run:
```shell
werf lint --dev
```

```shell
werf render --dev
```

## How to deploy a chart with TypeScript templates

Simply run `werf converge`: TypeScript templates will be rendered with Node.js and deployed.

> **Note**: According to [giterminism policies]({{ "/usage/project_configuration/giterminism.html" | true_relative_url }}), all changed files must be committed.

## Deploying into isolated environments

An isolated environment requires only Node.js on the target machine and the chart itself to be self-contained:

1. Publish the chart:
   ```shell
   werf bundle publish --repo example.org/mycompany/myapp
   ```
   All npm modules will be minified and bundled inside, so that the chart can be installed even without Internet access.

2. On the target machine with an isolated environment (no network access), run:
   ```shell
   werf bundle apply --repo example.org/mycompany/myapp
   ```
   TypeScript templates will be rendered and deployed using pre-compiled files from the chart bundle.

To change such a chart and pack it again without network access, pass `--ts-no-rebuild` to `nelm chart pack`: it reuses the pre-compiled files instead of rebuilding them, see [Changing a published chart]({{ "/usage/distribute/bundles.html#changing-a-published-chart" | true_relative_url }}). Helm, Argo CD and Flux do not run TypeScript templates, so they do not deploy the resources these templates generate.

## SDK API overview

TypeScript engine uses the [@nelm/chart-ts-sdk](https://github.com/werf/nelm-chart-ts-sdk) package.

### "render" and "generate" functions

`index.ts` must call the `render()` function. The function `generate()`, which will actually generate the manifests, should be passed to the `render()` function as an argument, for example:

```typescript
// .helm/ts/src/index.ts:
await render(generate);
```

### "WerfRenderContext" object

The `generate` function receives the root context in the `$` variable of type `WerfRenderContext` — the same context as in Helm templates:

| Field | Type | Description |
|-------|------|-------------|
| `$.Values` | `WerfServiceValues` | Chart parameters + service values at `$.Values.global.werf` |
| `$.Release` | `Release` | Release information |
| `$.Chart` | `ChartMetadata` | Metadata from Chart.yaml |
| `$.Capabilities` | `Capabilities` | Cluster capabilities (API versions, Kubernetes version) |
| `$.Files` | `Record<string, Uint8Array>` | Raw chart files (except `templates/` and `ts/`) |

See the example context in `ts/input.example.yaml`. For details on parameters and how they are constructed, see [Parametrize templates]({{ "/usage/deploy/values.html" | true_relative_url }}).

### "RenderResult" object

The `generate` function returns `RenderResult` — an object with a `manifests` array. Each element is a plain JavaScript object representing a Kubernetes resource. Example output:

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
