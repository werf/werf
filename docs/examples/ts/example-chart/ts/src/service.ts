import type { ChartContext } from './context.ts';
import { getFullname, getLabels, getSelectorLabels } from './helpers.ts';

export function newService($: ChartContext): object {
  return {
    apiVersion: 'v1',
    kind: 'Service',
    metadata: {
      name: getFullname($),
      labels: getLabels($),
    },
    spec: {
      type: $.Values.service?.type ?? 'ClusterIP',
      ports: [
        {
          port: $.Values.service?.port ?? 80,
          targetPort: 'http',
        },
      ],
      selector: getSelectorLabels($),
    },
  };
}
