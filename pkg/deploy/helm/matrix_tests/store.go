package matrix_tests

import (
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ResourceIndex struct {
	Kind      string
	Name      string
	Namespace string
}

func (g *ResourceIndex) AsString() string {
	if g.Namespace == "" {
		return g.Kind + "/" + g.Name
	}
	return g.Namespace + "/" + g.Kind + "/" + g.Name
}

func GetResourceIndex(object unstructured.Unstructured) ResourceIndex {
	return ResourceIndex{
		Kind:      object.GetKind(),
		Name:      object.GetName(),
		Namespace: object.GetNamespace(),
	}
}

type UnstructuredObjectStore struct {
	Storage map[ResourceIndex]unstructured.Unstructured
}

func NewUnstructuredObjectStore() UnstructuredObjectStore {
	return UnstructuredObjectStore{Storage: make(map[ResourceIndex]unstructured.Unstructured)}
}

func (s UnstructuredObjectStore) Put(object map[string]interface{}) error {
	var u unstructured.Unstructured
	u.SetUnstructuredContent(object)

	index := GetResourceIndex(u)
	if _, ok := s.Storage[index]; ok {
		return fmt.Errorf("object %q already exists in the object store", index.AsString())
	}

	s.Storage[index] = u
	return nil
}

func (s UnstructuredObjectStore) Get(key ResourceIndex) unstructured.Unstructured {
	return s.Storage[key]
}
