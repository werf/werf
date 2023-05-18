package helm

import (
	"fmt"

	"helm.sh/helm/v3/pkg/werf/common"
	"helm.sh/helm/v3/pkg/werf/mutator"
	"helm.sh/helm/v3/pkg/werf/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/werf/werf/pkg/werf"
)

var _ mutator.RuntimeResourceMutator = (*ServiceAnnotationsMutator)(nil)

func NewServiceAnnotationsMutator(werfEnv, werfProject string) *ServiceAnnotationsMutator {
	return &ServiceAnnotationsMutator{
		werfEnv:     werfEnv,
		werfProject: werfProject,
	}
}

type ServiceAnnotationsMutator struct {
	werfEnv     string
	werfProject string
}

func (m *ServiceAnnotationsMutator) Mutate(res resource.Resourcer, operationType common.ClientOperationType) (resource.Resourcer, error) {
	if !res.PartOfRelease() {
		return res, nil
	}

	switch operationType {
	case common.ClientOperationTypeCreate, common.ClientOperationTypeUpdate, common.ClientOperationTypeSmartApply:
	default:
		return res, nil
	}

	if err := unstructured.SetNestedField(res.Unstructured().UnstructuredContent(), werf.Version, "metadata", "annotations", "werf.io/version"); err != nil {
		return nil, fmt.Errorf("error adding werf version annotation: %w", err)
	}

	if err := unstructured.SetNestedField(res.Unstructured().UnstructuredContent(), m.werfProject, "metadata", "annotations", "project.werf.io/name"); err != nil {
		return nil, fmt.Errorf("error adding werf project name annotation: %w", err)
	}

	if err := unstructured.SetNestedField(res.Unstructured().UnstructuredContent(), m.werfEnv, "metadata", "annotations", "project.werf.io/env"); err != nil {
		return nil, fmt.Errorf("error adding werf project env annotation: %w", err)
	}

	return res, nil
}
