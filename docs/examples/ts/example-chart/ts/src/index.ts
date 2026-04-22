import { WerfRenderContext, RenderResult, render } from '@nelm/chart-ts-sdk';
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
