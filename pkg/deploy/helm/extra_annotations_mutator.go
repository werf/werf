package helm

import (
	"helm.sh/helm/v3/pkg/werf/common"
	"helm.sh/helm/v3/pkg/werf/mutator"
	"helm.sh/helm/v3/pkg/werf/resource"
)

var _ mutator.RuntimeResourceMutator = (*ExtraAnnotationsMutator)(nil)

func NewExtraAnnotationsMutator(extraAnnos map[string]string) *ExtraAnnotationsMutator {
	return &ExtraAnnotationsMutator{
		extraAnnos: extraAnnos,
	}
}

type ExtraAnnotationsMutator struct {
	extraAnnos map[string]string
}

func (m *ExtraAnnotationsMutator) Mutate(res resource.Resourcer, operationType common.ClientOperationType) (resource.Resourcer, error) {
	if !res.PartOfRelease() {
		return res, nil
	}

	switch operationType {
	case common.ClientOperationTypeCreate, common.ClientOperationTypeUpdate, common.ClientOperationTypeSmartApply:
	default:
		return res, nil
	}

	annos := res.Unstructured().GetAnnotations()
	if annos == nil {
		annos = make(map[string]string)
	}
	for k, v := range m.extraAnnos {
		annos[k] = v
	}
	res.Unstructured().SetAnnotations(annos)

	return res, nil
}
