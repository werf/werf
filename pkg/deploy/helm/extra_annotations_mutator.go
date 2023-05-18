package helm

import (
	"fmt"

	"helm.sh/helm/v3/pkg/werf/common"
	"helm.sh/helm/v3/pkg/werf/mutator"
	"helm.sh/helm/v3/pkg/werf/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

	for k, v := range m.extraAnnos {
		if err := unstructured.SetNestedField(res.Unstructured().UnstructuredContent(), v, "metadata", "annotations", k); err != nil {
			return nil, fmt.Errorf("error adding extra annotation: %w", err)
		}
	}

	return res, nil
}
