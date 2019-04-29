package helm

import (
	"fmt"
	"io"
	"time"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/kubedog/pkg/tracker"
	"github.com/flant/kubedog/pkg/trackers/rollout"
	"github.com/flant/kubedog/pkg/trackers/rollout/multitrack"
	"github.com/flant/logboek"
	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes/scheme"
	helmKube "k8s.io/helm/pkg/kube"
)

type ResourcesWaiter struct {
	Client       *helmKube.Client
	LogsFromTime time.Time
}

func (waiter *ResourcesWaiter) WaitForResources(timeout time.Duration, created helmKube.Result) error {
	specs := multitrack.MultitrackSpecs{}

	for _, v := range created {
		switch value := asVersioned(v).(type) {
		case *v1.Pod:
			spec, err := makeMultitrackSpec(&value.ObjectMeta)
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.Pods = append(specs.Pods, *spec)
			}
		case *appsv1.Deployment:
			spec, err := makeMultitrackSpec(&value.ObjectMeta)
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.Deployments = append(specs.Deployments, *spec)
			}
		case *appsv1beta1.Deployment:
			spec, err := makeMultitrackSpec(&value.ObjectMeta)
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.Deployments = append(specs.Deployments, *spec)
			}
		case *appsv1beta2.Deployment:
			spec, err := makeMultitrackSpec(&value.ObjectMeta)
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.Deployments = append(specs.Deployments, *spec)
			}
		case *extensions.Deployment:
			spec, err := makeMultitrackSpec(&value.ObjectMeta)
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.Deployments = append(specs.Deployments, *spec)
			}
		case *extensions.DaemonSet:
			spec, err := makeMultitrackSpec(&value.ObjectMeta)
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.DaemonSets = append(specs.DaemonSets, *spec)
			}
		case *appsv1.DaemonSet:
			spec, err := makeMultitrackSpec(&value.ObjectMeta)
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.DaemonSets = append(specs.DaemonSets, *spec)
			}
		case *appsv1beta2.DaemonSet:
			spec, err := makeMultitrackSpec(&value.ObjectMeta)
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.DaemonSets = append(specs.DaemonSets, *spec)
			}

		case *appsv1.StatefulSet:
			spec, err := makeMultitrackSpec(&value.ObjectMeta)
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.StatefulSets = append(specs.StatefulSets, *spec)
			}
		case *appsv1beta1.StatefulSet:
			spec, err := makeMultitrackSpec(&value.ObjectMeta)
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.StatefulSets = append(specs.StatefulSets, *spec)
			}
		case *appsv1beta2.StatefulSet:
			spec, err := makeMultitrackSpec(&value.ObjectMeta)
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.StatefulSets = append(specs.StatefulSets, *spec)
			}
		case *v1.ReplicationController:
		case *extensions.ReplicaSet:
		case *appsv1beta2.ReplicaSet:
		case *appsv1.ReplicaSet:
		case *v1.PersistentVolumeClaim:
		case *v1.Service:
		}
	}

	return logboek.LogSecondaryProcess("Waiting for release resources to become ready", logboek.LogProcessOptions{}, func() error {
		return multitrack.Multitrack(kube.Kubernetes, specs, multitrack.MultitrackOptions{
			Options: tracker.Options{
				Timeout:      timeout,
				LogsFromTime: waiter.LogsFromTime,
			},
		})
	})
}

func makeMultitrackSpec(objMeta *metav1.ObjectMeta) (*multitrack.MultitrackSpec, error) {
	if objMeta.Annotations[TrackAnnoName] == string(TrackDisabled) {
		return nil, nil
	}

	return &multitrack.MultitrackSpec{
		ResourceName: objMeta.Name,
		Namespace:    objMeta.Namespace,
	}, nil
}

func (waiter *ResourcesWaiter) WatchUntilReady(namespace string, reader io.Reader, timeout time.Duration) error {
	watchStartTime := time.Now()

	infos, err := waiter.Client.BuildUnstructured(namespace, reader)
	if err != nil {
		return err
	}

TrackHooks:
	for _, info := range infos {
		name := info.Name
		namespace := info.Namespace
		kind := info.Mapping.GroupVersionKind.Kind

		switch value := asVersioned(info).(type) {
		case *batchv1.Job:
			if value.ObjectMeta.Annotations[TrackAnnoName] == string(TrackDisabled) {
				continue TrackHooks
			}

			loggerProcessMsg := fmt.Sprintf("Waiting for helm hook job/%s termination", name)
			if err := logboek.LogSecondaryProcess(loggerProcessMsg, logboek.LogProcessOptions{}, func() error {
				return logboek.WithFittedStreamsOutputOn(func() error {
					return rollout.TrackJobTillDone(name, namespace, kube.Kubernetes, tracker.Options{Timeout: timeout, LogsFromTime: watchStartTime})
				})
			}); err != nil {
				return err
			}

		default:
			return fmt.Errorf("helm hook kind '%s' not supported yet, only Job hooks can be used for now", kind)
		}
	}

	return nil
}

func asVersioned(info *resource.Info) runtime.Object {
	converter := runtime.ObjectConvertor(scheme.Scheme)
	groupVersioner := runtime.GroupVersioner(schema.GroupVersions(scheme.Scheme.PrioritizedVersionsAllGroups()))
	if info.Mapping != nil {
		groupVersioner = info.Mapping.GroupVersionKind.GroupVersion()
	}

	if obj, err := converter.ConvertToVersion(info.Object, groupVersioner); err == nil {
		return obj
	}
	return info.Object
}
