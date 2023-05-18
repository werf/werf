package helm

import (
	"fmt"

	"helm.sh/helm/v3/pkg/werf/common"
	"helm.sh/helm/v3/pkg/werf/mutator"
	"helm.sh/helm/v3/pkg/werf/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ mutator.RuntimeResourceMutator = (*ExtraLabelsMutator)(nil)

func NewExtraLabelsMutator(extraLabels map[string]string) *ExtraLabelsMutator {
	return &ExtraLabelsMutator{
		extraLabels: extraLabels,
	}
}

type ExtraLabelsMutator struct {
	extraLabels map[string]string
}

func (m *ExtraLabelsMutator) Mutate(res resource.Resourcer, operationType common.ClientOperationType) (resource.Resourcer, error) {
	if !res.PartOfRelease() {
		return res, nil
	}

	switch operationType {
	case common.ClientOperationTypeCreate, common.ClientOperationTypeUpdate, common.ClientOperationTypeSmartApply:
	default:
		return res, nil
	}

	for k, v := range m.extraLabels {
		if err := unstructured.SetNestedField(res.Unstructured().UnstructuredContent(), v, "metadata", "labels", k); err != nil {
			return nil, fmt.Errorf("error adding extra labels: %w", err)
		}
	}

	return res, nil
}
