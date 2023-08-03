package helm

import (
	"helm.sh/helm/v3/pkg/werf/common"
	"helm.sh/helm/v3/pkg/werf/mutator"
	"helm.sh/helm/v3/pkg/werf/resource"

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

	annos := res.Unstructured().GetAnnotations()
	if annos == nil {
		annos = make(map[string]string)
	}
	annos["werf.io/version"] = werf.Version
	annos["project.werf.io/name"] = m.werfProject
	annos["project.werf.io/env"] = m.werfEnv
	res.Unstructured().SetAnnotations(annos)

	return res, nil
}
