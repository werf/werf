package helm

import (
	"fmt"
	"strings"

	"helm.sh/helm/v3/pkg/phases/stages/externaldeps"

	"github.com/werf/werf/pkg/slug"
)

func NewExternalDepsAnnotationsParser(defaultNamespace string) *ExternalDepsAnnotationsParser {
	return &ExternalDepsAnnotationsParser{
		defaultNamespace: defaultNamespace,
	}
}

type ExternalDepsAnnotationsParser struct {
	defaultNamespace string
}

func (s *ExternalDepsAnnotationsParser) Parse(annotations map[string]string) (externaldeps.ExternalDependencyList, error) {
	extDeps, err := s.parseResourceAnnotations(annotations)
	if err != nil {
		return nil, fmt.Errorf("error parsing ext deps resource annotations: %w", err)
	}

	extDeps, err = s.parseNamespaceAnnotations(extDeps, annotations)
	if err != nil {
		return nil, fmt.Errorf("error parsing ext deps namespace annotations: %w", err)
	}

	return extDeps, nil
}

func (s *ExternalDepsAnnotationsParser) parseResourceAnnotations(annotations map[string]string) (externaldeps.ExternalDependencyList, error) {
	externalDependencyList := externaldeps.ExternalDependencyList{}
	for annoKey, annoVal := range annotations {
		annoKey, annoVal = s.normalizeAnnotation(annoKey, annoVal)

		if !s.matchResourceAnnotation(annoKey) {
			continue
		}

		if err := s.validateResourceAnnotation(annoKey, annoVal); err != nil {
			return nil, fmt.Errorf("error validating external dependency resource annotation: %w", err)
		}

		name := s.parseResourceAnnotationKey(annoKey)
		resourceType, resourceName := s.parseResourceAnnotationValue(annoVal)

		externalDependencyList = append(externalDependencyList, externaldeps.NewExternalDependency(name, resourceType, resourceName))
	}

	return externalDependencyList, nil
}

func (s *ExternalDepsAnnotationsParser) parseNamespaceAnnotations(extDeps externaldeps.ExternalDependencyList, annotations map[string]string) (externaldeps.ExternalDependencyList, error) {
	for annoKey, annoVal := range annotations {
		annoKey, annoVal = s.normalizeAnnotation(annoKey, annoVal)

		if !s.matchNamespaceAnnotation(annoKey) {
			continue
		}

		if err := s.validateNamespaceAnnotation(annoKey, annoVal); err != nil {
			return nil, fmt.Errorf("error validating external dependency namespace annotation: %w", err)
		}

		name := s.parseNamespaceAnnotationKey(annoKey)

		for _, extDep := range extDeps {
			if extDep.Name == name {
				extDep.Namespace = annoVal
				break
			}
		}
	}

	for _, extDep := range extDeps {
		if extDep.Namespace == "" {
			extDep.Namespace = s.defaultNamespace
		}
	}

	return extDeps, nil
}

func (s *ExternalDepsAnnotationsParser) normalizeAnnotation(key, value string) (string, string) {
	key = strings.TrimSpace(key)
	key = strings.Trim(key, "/.")
	key = strings.TrimSpace(key)

	value = strings.TrimSpace(value)

	return key, value
}

func (s *ExternalDepsAnnotationsParser) matchResourceAnnotation(key string) bool {
	return strings.HasSuffix(key, ExternalDependencyResourceAnnoName)
}

func (s *ExternalDepsAnnotationsParser) matchNamespaceAnnotation(key string) bool {
	return strings.HasSuffix(key, ExternalDependencyNamespaceAnnoName)
}

func (s *ExternalDepsAnnotationsParser) validateResourceAnnotation(key, value string) error {
	if key == ExternalDependencyResourceAnnoName {
		return fmt.Errorf("annotation %q should have prefix specified, e.g. \"backend.%s\"", key, ExternalDependencyResourceAnnoName)
	}

	if value == "" {
		return fmt.Errorf("annotation %q value should be specified", key)
	}

	valueElems := strings.Split(value, "/")

	if len(valueElems) != 2 {
		return fmt.Errorf("wrong annotation %q value format, should be: type/name", key)
	}

	switch valueElems[0] {
	case "":
		return fmt.Errorf("in annotation %q resource type can't be empty", key)
	case "all":
		return fmt.Errorf("\"all\" resource type is not allowed in annotation %q", key)
	}

	resourceTypeParts := strings.Split(valueElems[0], ".")
	for _, part := range resourceTypeParts {
		if part == "" {
			return fmt.Errorf("resource type in annotation %q should have dots (.) delimiting only non-empty resource.version.group: %s", ExternalDependencyResourceAnnoName, key)
		}
	}

	if valueElems[1] == "" {
		return fmt.Errorf("in annotation %q resource name can't be empty", key)
	}

	return nil
}

func (s *ExternalDepsAnnotationsParser) validateNamespaceAnnotation(key, value string) error {
	if key == ExternalDependencyNamespaceAnnoName {
		return fmt.Errorf("annotation %q should have prefix specified, e.g. \"backend.%s\"", key, ExternalDependencyNamespaceAnnoName)
	}

	if value == "" {
		return fmt.Errorf("annotation %q value should be specified", key)
	}

	if err := slug.ValidateKubernetesNamespace(value); err != nil {
		return fmt.Errorf("error validating annotation \"%s=%s\" namespace name: %w", key, value, err)
	}

	return nil
}

func (s *ExternalDepsAnnotationsParser) parseResourceAnnotationKey(key string) (name string) {
	return strings.TrimSuffix(key, fmt.Sprint(".", ExternalDependencyResourceAnnoName))
}

func (s *ExternalDepsAnnotationsParser) parseResourceAnnotationValue(value string) (resourceType, resourceName string) {
	elems := strings.Split(value, "/")
	return elems[0], elems[1]
}

func (s *ExternalDepsAnnotationsParser) parseNamespaceAnnotationKey(key string) (name string) {
	return strings.TrimSuffix(key, fmt.Sprint(".", ExternalDependencyNamespaceAnnoName))
}
