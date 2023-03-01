package helpers

import (
	"context"

	"helm.sh/helm/v3/pkg/chartutil"

	"github.com/werf/werf/pkg/deploy/helm/chart_extender/helpers/secrets"
)

type ChartExtenderValuesMerger struct{}

func (m *ChartExtenderValuesMerger) MergeValues(ctx context.Context, inputVals, serviceVals map[string]interface{}, secretsRuntimeData *secrets.SecretsRuntimeData) (map[string]interface{}, error) {
	vals := make(map[string]interface{})

	DebugPrintValues(ctx, "service", serviceVals)
	chartutil.CoalesceTables(vals, serviceVals) // NOTE: service values will not be saved into the marshalled release

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
