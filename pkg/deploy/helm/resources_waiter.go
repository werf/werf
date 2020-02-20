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
	Client                    *helmKube.Client
	LogsFromTime              time.Time
	StatusProgressPeriod      time.Duration
	HooksStatusProgressPeriod time.Duration
}

func extractSpecReplicas(specReplicas *int32) int {
	if specReplicas != nil {
		return int(*specReplicas)
	}
	return 1
}

func (waiter *ResourcesWaiter) WaitForResources(timeout time.Duration, created helmKube.Result) error {
	specs := multitrack.MultitrackSpecs{}

	for _, v := range created {
		switch value := asVersioned(v).(type) {
		case *appsv1.Deployment:
			spec, err := makeMultitrackSpec(&value.ObjectMeta, allowedFailuresCountOptions{multiplier: extractSpecReplicas(value.Spec.Replicas), defaultPerReplica: 1}, "deploy")
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.Deployments = append(specs.Deployments, *spec)
			}
		case *appsv1beta1.Deployment:
			spec, err := makeMultitrackSpec(&value.ObjectMeta, allowedFailuresCountOptions{multiplier: extractSpecReplicas(value.Spec.Replicas), defaultPerReplica: 1}, "deploy")
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.Deployments = append(specs.Deployments, *spec)
			}
		case *appsv1beta2.Deployment:
			spec, err := makeMultitrackSpec(&value.ObjectMeta, allowedFailuresCountOptions{multiplier: extractSpecReplicas(value.Spec.Replicas), defaultPerReplica: 1}, "deploy")
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.Deployments = append(specs.Deployments, *spec)
			}
		case *extensions.Deployment:
			spec, err := makeMultitrackSpec(&value.ObjectMeta, allowedFailuresCountOptions{multiplier: extractSpecReplicas(value.Spec.Replicas), defaultPerReplica: 1}, "deploy")
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.Deployments = append(specs.Deployments, *spec)
			}
		case *extensions.DaemonSet:
			// TODO: multiplier equals 3 because typically there are only 3 nodes in the cluster.
			// TODO: It is better to fetch number of nodes dynamically, but in the most cases multiplier=3 will work ok.
			spec, err := makeMultitrackSpec(&value.ObjectMeta, allowedFailuresCountOptions{multiplier: 3, defaultPerReplica: 1}, "ds")
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.DaemonSets = append(specs.DaemonSets, *spec)
			}
		case *appsv1.DaemonSet:
			// TODO: multiplier equals 3 because typically there are only 3 nodes in the cluster.
			// TODO: It is better to fetch number of nodes dynamically, but in the most cases multiplier=3 will work ok.
			spec, err := makeMultitrackSpec(&value.ObjectMeta, allowedFailuresCountOptions{multiplier: 3, defaultPerReplica: 1}, "ds")
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.DaemonSets = append(specs.DaemonSets, *spec)
			}
		case *appsv1beta2.DaemonSet:
			// TODO: multiplier equals 3 because typically there are only 3 nodes in the cluster.
			// TODO: It is better to fetch number of nodes dynamically, but in the most cases multiplier=3 will work ok.
			spec, err := makeMultitrackSpec(&value.ObjectMeta, allowedFailuresCountOptions{multiplier: 3, defaultPerReplica: 1}, "ds")
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.DaemonSets = append(specs.DaemonSets, *spec)
			}
		case *appsv1.StatefulSet:
			spec, err := makeMultitrackSpec(&value.ObjectMeta, allowedFailuresCountOptions{multiplier: extractSpecReplicas(value.Spec.Replicas), defaultPerReplica: 1}, "sts")
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.StatefulSets = append(specs.StatefulSets, *spec)
			}
		case *appsv1beta1.StatefulSet:
			spec, err := makeMultitrackSpec(&value.ObjectMeta, allowedFailuresCountOptions{multiplier: extractSpecReplicas(value.Spec.Replicas), defaultPerReplica: 1}, "sts")
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.StatefulSets = append(specs.StatefulSets, *spec)
			}
		case *appsv1beta2.StatefulSet:
			spec, err := makeMultitrackSpec(&value.ObjectMeta, allowedFailuresCountOptions{multiplier: extractSpecReplicas(value.Spec.Replicas), defaultPerReplica: 1}, "sts")
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.StatefulSets = append(specs.StatefulSets, *spec)
			}
		case *batchv1.Job:
			spec, err := makeMultitrackSpec(&value.ObjectMeta, allowedFailuresCountOptions{multiplier: 1, defaultPerReplica: 0}, "job")
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.Jobs = append(specs.Jobs, *spec)
			}
		case *v1.ReplicationController:
		case *extensions.ReplicaSet:
		case *appsv1beta2.ReplicaSet:
		case *appsv1.ReplicaSet:
		case *v1.PersistentVolumeClaim:
		case *v1.Service:
		}
	}

	logboek.LogOptionalLn()
	return logboek.LogProcess("Waiting for release resources to become ready", logboek.LogProcessOptions{}, func() error {
		return multitrack.Multitrack(kube.Kubernetes, specs, multitrack.MultitrackOptions{
			StatusProgressPeriod: waiter.StatusProgressPeriod,
			Options: tracker.Options{
				Timeout:      timeout,
				LogsFromTime: waiter.LogsFromTime,
			},
		})
	})
}

func makeMultitrackSpec(objMeta *metav1.ObjectMeta, failuresCountOptions allowedFailuresCountOptions, kind string) (*multitrack.MultitrackSpec, error) {
	multitrackSpec, err := prepareMultitrackSpec(objMeta.Name, kind, objMeta.Namespace, objMeta.Annotations, failuresCountOptions)
	if err != nil {
		logboek.LogWarnF("WARNING: %s\n", err)
		return nil, nil
	}

	return multitrackSpec, nil
}

type allowedFailuresCountOptions struct {
	multiplier        int
	defaultPerReplica int
}

func applyAllowedFailuresCountMultiplier(value, multiplier int) int {
	if multiplier > 0 {
		return value * multiplier
	}
	return value
}

func prepareMultitrackSpec(metadataName, resourceNameOrKind, namespace string, annotations map[string]string, failuresCountOptions allowedFailuresCountOptions) (*multitrack.MultitrackSpec, error) {
	defaultAllowFailuresCount := new(int)
	// Allow 1 fail per replica by default
	*defaultAllowFailuresCount = applyAllowedFailuresCountMultiplier(failuresCountOptions.defaultPerReplica, failuresCountOptions.multiplier)

	multitrackSpec := &multitrack.MultitrackSpec{
		ResourceName:            metadataName,
		Namespace:               namespace,
		LogRegexByContainerName: map[string]*regexp.Regexp{},
		AllowFailuresCount:      defaultAllowFailuresCount,
	}

mainLoop:
	for annoName, annoValue := range annotations {
		invalidAnnoValueError := fmt.Errorf("%s/%s annotation %s with invalid value %s", resourceNameOrKind, metadataName, annoName, annoValue)

		switch annoName {
		case ShowLogsUntilAnnoName:
			return nil, fmt.Errorf("%s/%s annotation %s not supported yet", resourceNameOrKind, metadataName, annoName)
		case SkipLogsAnnoName:
			boolValue, err := strconv.ParseBool(annoValue)
			if err != nil {
				return nil, fmt.Errorf("%s: bool expected: %s", invalidAnnoValueError, err)
			}

			multitrackSpec.SkipLogs = boolValue
		case ShowEventsAnnoName:
			boolValue, err := strconv.ParseBool(annoValue)
			if err != nil {
				return nil, fmt.Errorf("%s: bool expected: %s", invalidAnnoValueError, err)
			}

			multitrackSpec.ShowServiceMessages = boolValue
		case TrackTerminationModeAnnoName:
			trackTerminationModeValue := multitrack.TrackTerminationMode(annoValue)
			values := []multitrack.TrackTerminationMode{multitrack.WaitUntilResourceReady, multitrack.NonBlocking}
			for _, value := range values {
				if value == trackTerminationModeValue {
					multitrackSpec.TrackTerminationMode = trackTerminationModeValue
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
		case FailuresAllowedPerReplicaAnnoName:
			intValue, err := strconv.Atoi(annoValue)
			if err != nil || intValue < 0 {
				return nil, fmt.Errorf("%s: positive or zero integer expected", invalidAnnoValueError)
			}

			allowFailuresCount := new(int)
			*allowFailuresCount = applyAllowedFailuresCountMultiplier(intValue, failuresCountOptions.multiplier)
			multitrackSpec.AllowFailuresCount = allowFailuresCount
		case LogRegexAnnoName:
			regexpValue, err := regexp.Compile(annoValue)
			if err != nil {
				return nil, fmt.Errorf("%s: %s", invalidAnnoValueError, err)
			}

			multitrackSpec.LogRegex = regexpValue
		// case ShowLogsUntilAnnoName:
		// 	deployConditionValue := multitrack.DeployCondition(annoValue)
		// 	values := []multitrack.DeployCondition{multitrack.ControllerIsReady, multitrack.PodIsReady, multitrack.EndOfDeploy}
		// 	for _, value := range values {
		// 		if value == deployConditionValue {
		// 			multitrackSpec.ShowLogsUntil = deployConditionValue
		// 			continue mainLoop
		// 		}
		// 	}

		// 	return nil, fmt.Errorf("%s: choose one of %v", invalidAnnoValueError, values)
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
			if strings.HasPrefix(annoName, LogRegexForAnnoPrefix) {
				if containerName := strings.TrimPrefix(annoName, LogRegexForAnnoPrefix); containerName != "" {
					regexpValue, err := regexp.Compile(annoValue)
					if err != nil {
						return nil, fmt.Errorf("%s: %s", invalidAnnoValueError, err)
					}

					multitrackSpec.LogRegexByContainerName[containerName] = regexpValue
				}
			}
		}
	}

	return multitrackSpec, nil
}

func (waiter *ResourcesWaiter) WatchUntilReady(namespace string, reader io.Reader, timeout time.Duration) error {
	infos, err := waiter.Client.BuildUnstructured(namespace, reader)
	if err != nil {
		return err
	}

	for _, info := range infos {
		name := info.Name
		kind := info.Mapping.GroupVersionKind.Kind

		switch value := asVersioned(info).(type) {
		case *batchv1.Job:
			specs := multitrack.MultitrackSpecs{}

			spec, err := makeMultitrackSpec(&value.ObjectMeta, allowedFailuresCountOptions{multiplier: 1, defaultPerReplica: 0}, "job")
			if err != nil {
				return fmt.Errorf("cannot track %s %s: %s", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.Jobs = append(specs.Jobs, *spec)
			}

			return logboek.LogProcess(fmt.Sprintf("Waiting for helm hook job/%s termination", name), logboek.LogProcessOptions{}, func() error {
				return multitrack.Multitrack(kube.Kubernetes, specs, multitrack.MultitrackOptions{
					StatusProgressPeriod: waiter.HooksStatusProgressPeriod,
					Options: tracker.Options{
						Timeout:      timeout,
						LogsFromTime: waiter.LogsFromTime,
					},
				})
			})

		default:
			logboek.LogWarnF("WARNING: Will not track helm hook %s/%s: %s kind not supported for tracking\n", strings.ToLower(kind), name, kind)
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
