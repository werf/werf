import { type RenderResult, render } from '@nelm/chart-ts-sdk';
import type { ChartContext } from './context.ts';
import { newDeployment } from './deployment.ts';
import { newService } from './service.ts';

function generate($: ChartContext): RenderResult {
  const manifests: object[] = [];

  manifests.push(newDeployment($));

  if ($.Values.service?.enabled !== false) {
    manifests.push(newService($));
  }

  return { manifests };
}

await render(generate);
