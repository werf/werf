package helm

import (
	"fmt"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/phases/stages"
	"helm.sh/helm/v3/pkg/phases/stages/externaldeps"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/resource"
)

func NewStagesExternalDepsGenerator(restClient *action.RESTClientGetter, defaultNamespace *string) *StagesExternalDepsGenerator {
	return &StagesExternalDepsGenerator{
		defaultNamespace: defaultNamespace,
		restClient:       restClient,
		metaAccessor:     metadataAccessor,
	}
}

type StagesExternalDepsGenerator struct {
	defaultNamespace *string
	restClient       *action.RESTClientGetter
	mapper           meta.RESTMapper
	gvkBuilder       externaldeps.GVKBuilder
	metaAccessor     meta.MetadataAccessor
	initialized      bool
}

func (s *StagesExternalDepsGenerator) init() error {
	if s.initialized {
		return nil
	}

	mapper, err := (*s.restClient).ToRESTMapper()
	if err != nil {
		return fmt.Errorf("error getting REST mapper: %w", err)
	}
	s.mapper = mapper

	s.gvkBuilder = NewGVKBuilder(mapper)

	s.initialized = true

	return nil
}

func (s *StagesExternalDepsGenerator) Generate(stages stages.SortedStageList) error {
	if err := s.init(); err != nil {
		return fmt.Errorf("error initializing external dependencies generator: %w", err)
	}

	for _, stage := range stages {
		if err := stage.DesiredResources.Visit(func(resInfo *resource.Info, err error) error {
			if err != nil {
				return err
			}

			annotations, err := s.metaAccessor.Annotations(resInfo.Object)
			if err != nil {
				return fmt.Errorf("error getting annotations for object: %w", err)
			}

			resExtDeps, err := s.resourceExternalDepsFromAnnotations(annotations)
			if err != nil {
				return fmt.Errorf("error generating external dependencies from resource annotations: %w", err)
			}

			stage.ExternalDependencies = append(stage.ExternalDependencies, resExtDeps...)

			return nil
		}); err != nil {
			return fmt.Errorf("error visiting resources list: %w", err)
		}
	}

	return nil
}

func (s *StagesExternalDepsGenerator) resourceExternalDepsFromAnnotations(annotations map[string]string) (externaldeps.ExternalDependencyList, error) {
	extDepsList, err := NewExternalDepsAnnotationsParser(*s.defaultNamespace).Parse(annotations)
	if err != nil {
		return nil, fmt.Errorf("error parsing external dependencies annotations: %w", err)
	}

	if len(extDepsList) == 0 {
		return nil, nil
	}

	for _, extDep := range extDepsList {
		if err := extDep.GenerateInfo(s.gvkBuilder, s.metaAccessor, s.mapper); err != nil {
			return nil, fmt.Errorf("error generating Info for external dependency: %w", err)
		}
	}

	return extDepsList, nil
}
