package helm

import (
	"fmt"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/resource"
)

var metadataAccessor = meta.NewAccessor()

type HelmKubeClientExtender struct{}

func NewHelmKubeClientExtender() *HelmKubeClientExtender {
	return &HelmKubeClientExtender{}
}

func (extender *HelmKubeClientExtender) BeforeCreateResource(info *resource.Info) error {
	resourceName := info.ObjectName()

	annotations, err := metadataAccessor.Annotations(info.Object)
	if err != nil {
		return err
	}

	if value, hasKey := annotations[ReplicasOnCreationAnnoName]; hasKey {
		intValue, err := strconv.Atoi(value)
		if err != nil || intValue < 0 {
			return fmt.Errorf("%s annotation %s with invalid value %s: positive or zero integer expected", resourceName, ReplicasOnCreationAnnoName, value)
		}

		switch value := asVersioned(info).(type) {
		case *appsv1.Deployment:
			value.Spec.Replicas = new(int32)
			*value.Spec.Replicas = int32(intValue)
			info.Object = value
		case *appsv1beta1.Deployment:
			value.Spec.Replicas = new(int32)
			*value.Spec.Replicas = int32(intValue)
			info.Object = value
		case *appsv1beta2.Deployment:
			value.Spec.Replicas = new(int32)
			*value.Spec.Replicas = int32(intValue)
			info.Object = value
		case *extensions.Deployment:
			value.Spec.Replicas = new(int32)
			*value.Spec.Replicas = int32(intValue)
			info.Object = value
		case *appsv1.StatefulSet:
			value.Spec.Replicas = new(int32)
			*value.Spec.Replicas = int32(intValue)
			info.Object = value
		case *appsv1beta1.StatefulSet:
			value.Spec.Replicas = new(int32)
			*value.Spec.Replicas = int32(intValue)
			info.Object = value
		case *appsv1beta2.StatefulSet:
			value.Spec.Replicas = new(int32)
			*value.Spec.Replicas = int32(intValue)
			info.Object = value
		case *extensions.ReplicaSet:
			value.Spec.Replicas = new(int32)
			*value.Spec.Replicas = int32(intValue)
			info.Object = value
		case *appsv1beta2.ReplicaSet:
			value.Spec.Replicas = new(int32)
			*value.Spec.Replicas = int32(intValue)
			info.Object = value
		case *appsv1.ReplicaSet:
			value.Spec.Replicas = new(int32)
			*value.Spec.Replicas = int32(intValue)
			info.Object = value
		}
	}

	return nil
}

func (extender *HelmKubeClientExtender) BeforeUpdateResource(info *resource.Info) error {
	return nil
}

func (extender *HelmKubeClientExtender) BeforeDeleteResource(info *resource.Info) error {
	return nil
}
