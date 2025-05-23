package helpers_for_werf_helm

import (
	"context"

	chartutil "github.com/werf/3p-helm-for-werf-helm/pkg/chartutil"
	secrets "github.com/werf/werf/v2/pkg/deploy/helm_for_werf_helm/chart_extender_for_werf_helm/helpers_for_werf_helm/secrets_for_werf_helm"
)

type ChartExtenderValuesMerger struct{}

func (m *ChartExtenderValuesMerger) MergeValues(ctx context.Context, inputVals, serviceVals map[string]interface{}, secretsRuntimeData *secrets.SecretsRuntimeData) (map[string]interface{}, error) {
	vals := make(map[string]interface{})

	DebugPrintValues(ctx, "service", serviceVals)
	chartutil.CoalesceTables(vals, serviceVals) // NOTE: service values will not be saved into the marshaled release

	if secretsRuntimeData != nil {
		if DebugSecretValues() {
			DebugPrintValues(ctx, "secret", secretsRuntimeData.DecryptedSecretValues)
		}
		chartutil.CoalesceTables(vals, secretsRuntimeData.DecryptedSecretValues)
	}

	DebugPrintValues(ctx, "input", inputVals)
	chartutil.CoalesceTables(vals, inputVals)

	if DebugSecretValues() {
		// Only print all values with secrets when secret values debug enabled
		DebugPrintValues(ctx, "all", vals)
	}

	return vals, nil
}
