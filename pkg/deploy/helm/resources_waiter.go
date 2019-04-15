package helm

import (
	"fmt"
	"io"
	"time"

	helmKube "github.com/flant/helm/pkg/kube"
	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/kubedog/pkg/tracker"
	"github.com/flant/kubedog/pkg/trackers/rollout"
	"github.com/flant/kubedog/pkg/trackers/rollout/multitrack"
	"github.com/flant/logboek"
	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	v1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes/scheme"
)

type ResourcesWaiter struct {
	Client *helmKube.Client
}

func (waiter *ResourcesWaiter) WaitForResources(timeout time.Duration, created helmKube.Result) error {
	specs := multitrack.MultitrackSpecs{}

	for _, v := range created {
		switch value := asVersioned(v).(type) {
		case *v1.Pod:
			specs.Pods = append(specs.Pods, multitrack.MultitrackSpec{
				ResourceName: value.Name,
				Namespace:    value.Namespace,
			})
		case *appsv1.Deployment:
			specs.Deployments = append(specs.Deployments, multitrack.MultitrackSpec{
				ResourceName: value.Name,
				Namespace:    value.Namespace,
			})
		case *appsv1beta1.Deployment:
			specs.Deployments = append(specs.Deployments, multitrack.MultitrackSpec{
				ResourceName: value.Name,
				Namespace:    value.Namespace,
			})
		case *appsv1beta2.Deployment:
			specs.Deployments = append(specs.Deployments, multitrack.MultitrackSpec{
				ResourceName: value.Name,
				Namespace:    value.Namespace,
			})
		case *extensions.Deployment:
			specs.Deployments = append(specs.Deployments, multitrack.MultitrackSpec{
				ResourceName: value.Name,
				Namespace:    value.Namespace,
			})
		case *extensions.DaemonSet:
			specs.DaemonSets = append(specs.DaemonSets, multitrack.MultitrackSpec{
				ResourceName: value.Name,
				Namespace:    value.Namespace,
			})
		case *appsv1.DaemonSet:
			specs.DaemonSets = append(specs.DaemonSets, multitrack.MultitrackSpec{
				ResourceName: value.Name,
				Namespace:    value.Namespace,
			})
		case *appsv1beta2.DaemonSet:
			specs.DaemonSets = append(specs.DaemonSets, multitrack.MultitrackSpec{
				ResourceName: value.Name,
				Namespace:    value.Namespace,
			})
		case *appsv1.StatefulSet:
			specs.StatefulSets = append(specs.StatefulSets, multitrack.MultitrackSpec{
				ResourceName: value.Name,
				Namespace:    value.Namespace,
			})
		case *appsv1beta1.StatefulSet:
			specs.StatefulSets = append(specs.StatefulSets, multitrack.MultitrackSpec{
				ResourceName: value.Name,
				Namespace:    value.Namespace,
			})
		case *appsv1beta2.StatefulSet:
			specs.StatefulSets = append(specs.StatefulSets, multitrack.MultitrackSpec{
				ResourceName: value.Name,
				Namespace:    value.Namespace,
			})
		case *v1.ReplicationController:
		case *extensions.ReplicaSet:
		case *appsv1beta2.ReplicaSet:
		case *appsv1.ReplicaSet:
		case *v1.PersistentVolumeClaim:
		case *v1.Service:
		}
	}

	return logboek.LogSecondaryProcess("Waiting for release resources to become ready", logboek.LogProcessOptions{}, func() error {
		return multitrack.Multitrack(kube.Kubernetes, specs, multitrack.MultitrackOptions{})
	})
}

func (waiter *ResourcesWaiter) WatchUntilReady(namespace string, reader io.Reader, timeout time.Duration) error {
	watchStartTime := time.Now()

	infos, err := waiter.Client.BuildUnstructured(namespace, reader)
	if err != nil {
		return err
	}

	for _, info := range infos {
		name := info.Name
		namespace := info.Namespace
		kind := info.Mapping.GroupVersionKind.Kind

		switch kind {
		case "Job":
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
