import type { RenderContext } from '@nelm/chart-ts-sdk';
import { getFullname, getLabels, getSelectorLabels } from './helpers.ts';

export function newDeployment($: RenderContext): object {
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
