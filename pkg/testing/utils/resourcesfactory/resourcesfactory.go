package resourcesfactory

import (
	"encoding/json"

	"sigs.k8s.io/yaml"

	v1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	. "github.com/onsi/gomega"
)

func unmarshalObject(manifestYaml string, obj interface{}) {
	manifestJson, err := yaml.YAMLToJSON([]byte(manifestYaml))
	Expect(err).To(Succeed())
	Expect(json.Unmarshal(manifestJson, obj)).To(Succeed())
}

func NewNamespace(manifestYaml string) *corev1.Namespace {
	obj := &corev1.Namespace{}
	unmarshalObject(manifestYaml, &obj)
	return obj
}

func NewJob(manifestYaml string) *batchv1.Job {
	obj := &batchv1.Job{}
	unmarshalObject(manifestYaml, &obj)
	return obj
}

func NewDeployment(manifestYaml string) *v1.Deployment {
	obj := &v1.Deployment{}
	unmarshalObject(manifestYaml, &obj)
	return obj
}
