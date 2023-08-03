package helm

import (
	"helm.sh/helm/v3/pkg/werf/common"
	"helm.sh/helm/v3/pkg/werf/mutator"
	"helm.sh/helm/v3/pkg/werf/resource"
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

	labels := res.Unstructured().GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	for k, v := range m.extraLabels {
		labels[k] = v
	}
	res.Unstructured().SetLabels(labels)

	return res, nil
}
