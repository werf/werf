package helm

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
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
			spec, err := makeMultitrackSpec(&value.ObjectMeta, "po")
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.Pods = append(specs.Pods, *spec)
			}
		case *appsv1.Deployment:
			spec, err := makeMultitrackSpec(&value.ObjectMeta, "deploy")
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.Deployments = append(specs.Deployments, *spec)
			}
		case *appsv1beta1.Deployment:
			spec, err := makeMultitrackSpec(&value.ObjectMeta, "deploy")
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.Deployments = append(specs.Deployments, *spec)
			}
		case *appsv1beta2.Deployment:
			spec, err := makeMultitrackSpec(&value.ObjectMeta, "deploy")
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.Deployments = append(specs.Deployments, *spec)
			}
		case *extensions.Deployment:
			spec, err := makeMultitrackSpec(&value.ObjectMeta, "deploy")
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.Deployments = append(specs.Deployments, *spec)
			}
		case *extensions.DaemonSet:
			spec, err := makeMultitrackSpec(&value.ObjectMeta, "ds")
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.DaemonSets = append(specs.DaemonSets, *spec)
			}
		case *appsv1.DaemonSet:
			spec, err := makeMultitrackSpec(&value.ObjectMeta, "ds")
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.DaemonSets = append(specs.DaemonSets, *spec)
			}
		case *appsv1beta2.DaemonSet:
			spec, err := makeMultitrackSpec(&value.ObjectMeta, "ds")
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.DaemonSets = append(specs.DaemonSets, *spec)
			}

		case *appsv1.StatefulSet:
			spec, err := makeMultitrackSpec(&value.ObjectMeta, "sts")
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.StatefulSets = append(specs.StatefulSets, *spec)
			}
		case *appsv1beta1.StatefulSet:
			spec, err := makeMultitrackSpec(&value.ObjectMeta, "sts")
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.StatefulSets = append(specs.StatefulSets, *spec)
			}
		case *appsv1beta2.StatefulSet:
			spec, err := makeMultitrackSpec(&value.ObjectMeta, "sts")
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

func makeMultitrackSpec(objMeta *metav1.ObjectMeta, kind string) (*multitrack.MultitrackSpec, error) {
	multitrackSpec, err := prepareMultitrackSpec(objMeta.Name, kind, objMeta.Namespace, objMeta.Annotations)
	if err != nil {
		logboek.LogErrorF("WARNING: %s\n", err)
		return nil, nil
	}

	return multitrackSpec, nil
}

func prepareMultitrackSpec(resourceName, kind, namespace string, annotations map[string]string) (*multitrack.MultitrackSpec, error) {
	multitrackSpec := &multitrack.MultitrackSpec{
		ResourceName:                 resourceName,
		Namespace:                    namespace,
		LogWatchRegexByContainerName: map[string]*regexp.Regexp{},
	}

mainLoop:
	for annoName, annoValue := range annotations {
		invalidAnnoValueError := fmt.Errorf("%s/%s annotation %s with invalid value %s", kind, resourceName, annoName, annoValue)

		switch annoName {
		case TrackAnnoName:
			trackValue := TrackAnno(annoValue)
			values := []TrackAnno{TrackAnnoEnabledValue, TrackAnnoDisabledValue}
			for _, value := range values {
				if value == trackValue {
					if value == TrackAnnoDisabledValue {
						return nil, nil
					}

					continue mainLoop
				}
			}

			return nil, fmt.Errorf("%s: choose one of %v", invalidAnnoValueError, values)
		case FailModeAnnoName:
			failModeValue := multitrack.FailMode(annoValue)
			values := []multitrack.FailMode{multitrack.IgnoreAndContinueDeployProcess, multitrack.FailWholeDeployProcessImmediately, multitrack.HopeUntilEndOfDeployProcess}
			for _, value := range values {
				if value == failModeValue {
					multitrackSpec.FailMode = failModeValue
					continue mainLoop
				}
			}

			return nil, fmt.Errorf("%s: choose one of %v", invalidAnnoValueError, values)
		case AllowFailuresCountAnnoName:
			intValue, err := strconv.Atoi(annoValue)
			if err != nil || intValue <= 0 {
				return nil, fmt.Errorf("%s: positive integer expected", invalidAnnoValueError)
			}

			multitrackSpec.AllowFailuresCount = &intValue
		case LogWatchRegexAnnoName:
			regexpValue, err := regexp.Compile(annoValue)
			if err != nil {
				return nil, fmt.Errorf("%s: %s", invalidAnnoValueError, err)
			}

			multitrackSpec.LogWatchRegex = regexpValue
		case ShowLogsUntilAnnoName:
			deployConditionValue := multitrack.DeployCondition(annoValue)
			values := []multitrack.DeployCondition{multitrack.ControllerIsReady, multitrack.PodIsReady, multitrack.EndOfDeploy}
			for _, value := range values {
				if value == deployConditionValue {
					multitrackSpec.ShowLogsUntil = deployConditionValue
					continue mainLoop
				}
			}

			return nil, fmt.Errorf("%s: choose one of %v", invalidAnnoValueError, values)
		case SkipLogsForContainersAnnoName:
			var containerNames []string
			for _, v := range strings.Split(annoValue, ",") {
				containerName := strings.TrimSpace(v)
				if containerName == "" {
					return nil, fmt.Errorf("%s: containers names separated by comma expected", invalidAnnoValueError)
				}

				containerNames = append(containerNames, containerName)
			}

			multitrackSpec.SkipLogsForContainers = containerNames
		case ShowLogsOnlyForContainers:
			var containerNames []string
			for _, v := range strings.Split(annoValue, ",") {
				containerName := strings.TrimSpace(v)
				if containerName == "" {
					return nil, fmt.Errorf("%s: containers names separated by comma expected", invalidAnnoValueError)
				}

				containerNames = append(containerNames, containerName)
			}

			multitrackSpec.ShowLogsOnlyForContainers = containerNames
		default:
			if strings.HasPrefix(annoName, ContainerLogWatchRegexAnnoPrefix) {
				if containerName := strings.TrimPrefix(annoName, ContainerLogWatchRegexAnnoPrefix); containerName != "" {
					regexpValue, err := regexp.Compile(annoValue)
					if err != nil {
						return nil, fmt.Errorf("%s: %s", invalidAnnoValueError, err)
					}

					multitrackSpec.LogWatchRegexByContainerName[containerName] = regexpValue
				}
			}
		}
	}

	return multitrackSpec, nil
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
			if value.ObjectMeta.Annotations[TrackAnnoName] == string(TrackAnnoDisabledValue) {
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
