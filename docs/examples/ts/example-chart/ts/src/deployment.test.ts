import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import type { ChartContext } from './context.ts';
import { newDeployment } from './deployment.ts';

function newContext(values: Record<string, unknown>): ChartContext {
  return {
    Capabilities: {
      APIVersions: ['v1', 'apps/v1'],
      HelmVersion: {
        Version: 'v3.20.0',
        GitCommit: '',
        GitTreeState: 'clean',
        GoVersion: 'go1.26.0',
      },
      KubeVersion: { Major: '1', Minor: '35', Version: 'v1.35.0' },
    },
    Chart: {
      APIVersion: 'v2',
      Annotations: {},
      AppVersion: '1.0.0',
      Condition: '',
      Description: '',
      Home: '',
      Icon: '',
      Keywords: [],
      Maintainers: [],
      Name: 'example-chart',
      Sources: [],
      Tags: [],
      Type: 'application',
      Version: '0.1.0',
    },
    Files: {},
    Release: {
      IsInstall: true,
      IsUpgrade: false,
      Name: 'example-release',
      Namespace: 'default',
      Revision: 1,
      Service: 'Helm',
    },
    Runtime: {},
    Values: values as ChartContext['Values'],
  };
}

describe('newDeployment', () => {
  it('uses defaults with empty values', () => {
    assert.partialDeepStrictEqual(newDeployment(newContext({})), {
      apiVersion: 'apps/v1',
      kind: 'Deployment',
      metadata: { name: 'example-release-example-chart' },
      spec: {
        replicas: 1,
        template: {
          spec: {
            containers: [{ image: 'nginx:latest', ports: [{ containerPort: 80 }] }],
          },
        },
      },
    });
  });

  it('uses custom values', () => {
    const context = newContext({
      fullnameOverride: 'custom-app',
      replicaCount: 3,
      image: { repository: 'example.org/app', tag: '1.2.3' },
      service: { port: 8080 },
    });

    assert.partialDeepStrictEqual(newDeployment(context), {
      metadata: { name: 'custom-app' },
      spec: {
        replicas: 3,
        template: {
          spec: {
            containers: [{ image: 'example.org/app:1.2.3', ports: [{ containerPort: 8080 }] }],
          },
        },
      },
    });
  });

  it('uses defaults for a partially specified image', () => {
    for (const [image, expected] of [
      [{ repository: 'example.org/app' }, 'example.org/app:latest'],
      [{ tag: 'stable' }, 'nginx:stable'],
    ] as const) {
      assert.partialDeepStrictEqual(newDeployment(newContext({ image })), {
        spec: { template: { spec: { containers: [{ image: expected }] } } },
      });
    }
  });

  it('preserves zero replicas', () => {
    assert.partialDeepStrictEqual(newDeployment(newContext({ replicaCount: 0 })), {
      spec: { replicas: 0 },
    });
  });
});
