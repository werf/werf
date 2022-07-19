package helm

import (
	"context"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	flaggerv1beta1 "github.com/fluxcd/flagger/pkg/apis/flagger/v1beta1"
	flaggerscheme "github.com/fluxcd/flagger/pkg/client/clientset/versioned/scheme"
	helm_kube "helm.sh/helm/v3/pkg/kube"
	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	batchv1 "k8s.io/api/batch/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/kubedog/pkg/tracker"
	"github.com/werf/kubedog/pkg/tracker/resid"
	"github.com/werf/kubedog/pkg/trackers/elimination"
	"github.com/werf/kubedog/pkg/trackers/rollout/multitrack"
	"github.com/werf/kubedog/pkg/trackers/rollout/multitrack/generic"
	"github.com/werf/logboek"
)

func init() {
	flaggerscheme.AddToScheme(scheme.Scheme)
}

type ResourcesWaiter struct {
	KubeInitializer           KubeInitializer
	Client                    *helm_kube.Client
	LogsFromTime              time.Time
	StatusProgressPeriod      time.Duration
	HooksStatusProgressPeriod time.Duration
}

func NewResourcesWaiter(kubeInitializer KubeInitializer, client *helm_kube.Client, logsFromTime time.Time, statusProgressPeriod, hooksStatusProgressPeriod time.Duration) *ResourcesWaiter {
	return &ResourcesWaiter{
		KubeInitializer:           kubeInitializer,
		Client:                    client,
		LogsFromTime:              logsFromTime,
		StatusProgressPeriod:      statusProgressPeriod,
		HooksStatusProgressPeriod: hooksStatusProgressPeriod,
	}
}

func extractSpecReplicas(specReplicas *int32) int {
	if specReplicas != nil {
		return int(*specReplicas)
	}
	return 1
}

func (waiter *ResourcesWaiter) Wait(ctx context.Context, resources helm_kube.ResourceList, timeout time.Duration) error {
	if os.Getenv("WERF_DISABLE_RESOURCES_WAITER") == "1" {
		return nil
	}

	if waiter.KubeInitializer != nil {
		if err := waiter.KubeInitializer.Init(ctx); err != nil {
			return fmt.Errorf("kube initializer failed: %w", err)
		}
	}

	specs, err := makeMultitrackSpecsFromResList(ctx, resources, timeout, waiter.StatusProgressPeriod)
	if err != nil {
		return fmt.Errorf("error making multitrack specs: %w", err)
	}

	// NOTE: use context from resources-waiter object here, will be changed in helm 3
	logboek.Context(ctx).LogOptionalLn()
	return logboek.Context(ctx).LogProcess("Waiting for resources to become ready").
		DoError(func() error {
			return multitrack.Multitrack(kube.Client, *specs, multitrack.MultitrackOptions{
				StatusProgressPeriod: waiter.StatusProgressPeriod,
				Options: tracker.Options{
					Timeout:      timeout,
					LogsFromTime: waiter.LogsFromTime,
				},
				DynamicClient:   kube.DynamicClient,
				DiscoveryClient: kube.CachedDiscoveryClient,
				Mapper:          kube.Mapper,
			})
		})
}

func makeMultitrackSpec(ctx context.Context, objMeta *metav1.ObjectMeta, failuresCountOptions allowedFailuresCountOptions, kind string) (*multitrack.MultitrackSpec, error) {
	multitrackSpec, err := prepareMultitrackSpec(objMeta.Name, kind, objMeta.Namespace, objMeta.Annotations, failuresCountOptions)
	if err != nil {
		logboek.Context(ctx).Warn().LogLn()
		logboek.Context(ctx).Warn().LogF("WARNING %s\n", err)
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
		ResourceName:                             metadataName,
		Namespace:                                namespace,
		LogRegexByContainerName:                  map[string]*regexp.Regexp{},
		AllowFailuresCount:                       defaultAllowFailuresCount,
		IgnoreReadinessProbeFailsByContainerName: map[string]time.Duration{},
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
				return nil, fmt.Errorf("%s: bool expected: %w", invalidAnnoValueError, err)
			}

			multitrackSpec.SkipLogs = boolValue
		case ShowEventsAnnoName:
			boolValue, err := strconv.ParseBool(annoValue)
			if err != nil {
				return nil, fmt.Errorf("%s: bool expected: %w", invalidAnnoValueError, err)
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

			return nil, fmt.Errorf("%w: choose one of %v", invalidAnnoValueError, values)
		case FailModeAnnoName:
			failModeValue := multitrack.FailMode(annoValue)
			values := []multitrack.FailMode{multitrack.IgnoreAndContinueDeployProcess, multitrack.FailWholeDeployProcessImmediately, multitrack.HopeUntilEndOfDeployProcess}
			for _, value := range values {
				if value == failModeValue {
					multitrackSpec.FailMode = failModeValue
					continue mainLoop
				}
			}

			return nil, fmt.Errorf("%w: choose one of %v", invalidAnnoValueError, values)
		case FailuresAllowedPerReplicaAnnoName:
			intValue, err := strconv.Atoi(annoValue)
			if err != nil || intValue < 0 {
				return nil, fmt.Errorf("%w: positive or zero integer expected", invalidAnnoValueError)
			}

			allowFailuresCount := new(int)
			*allowFailuresCount = applyAllowedFailuresCountMultiplier(intValue, failuresCountOptions.multiplier)
			multitrackSpec.AllowFailuresCount = allowFailuresCount
		case LogRegexAnnoName:
			regexpValue, err := regexp.Compile(annoValue)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", invalidAnnoValueError, err)
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
					return nil, fmt.Errorf("%w: containers names separated by comma expected", invalidAnnoValueError)
				}

				containerNames = append(containerNames, containerName)
			}

			multitrackSpec.SkipLogsForContainers = containerNames
		case ShowLogsOnlyForContainers:
			var containerNames []string
			for _, v := range strings.Split(annoValue, ",") {
				containerName := strings.TrimSpace(v)
				if containerName == "" {
					return nil, fmt.Errorf("%w: containers names separated by comma expected", invalidAnnoValueError)
				}

				containerNames = append(containerNames, containerName)
			}

			multitrackSpec.ShowLogsOnlyForContainers = containerNames
		default:
			if strings.HasPrefix(annoName, LogRegexForAnnoPrefix) {
				if containerName := strings.TrimPrefix(annoName, LogRegexForAnnoPrefix); containerName != "" {
					regexpValue, err := regexp.Compile(annoValue)
					if err != nil {
						return nil, fmt.Errorf("%s: %w", invalidAnnoValueError, err)
					}

					multitrackSpec.LogRegexByContainerName[containerName] = regexpValue
				}
			}
			if strings.HasPrefix(annoName, IgnoreReadinessProbeFailsForPrefix) {
				if containerName := strings.TrimPrefix(annoName, IgnoreReadinessProbeFailsForPrefix); containerName != "" {
					ignoreDuration, err := time.ParseDuration(annoValue)
					if err != nil {
						return nil, fmt.Errorf("%s: %w", invalidAnnoValueError, err)
					}
					if math.Signbit(ignoreDuration.Seconds()) {
						return nil, fmt.Errorf("%w: can't be less than 0", invalidAnnoValueError)
					}

					multitrackSpec.IgnoreReadinessProbeFailsByContainerName[containerName] = ignoreDuration
				}
			}
		}
	}

	return multitrackSpec, nil
}

func makeGenericSpec(ctx context.Context, resID *resid.ResourceID, statusProgressPeriod, timeout time.Duration, annotations map[string]string) (*generic.Spec, error) {
	genericSpec := &generic.Spec{
		ResourceID:           resID,
		Timeout:              timeout,
		StatusProgressPeriod: statusProgressPeriod,
	}

mainLoop:
	for annoName, annoValue := range annotations {
		invalidAnnoValueError := fmt.Errorf("%s annotation %s with invalid value %s", resID, annoName, annoValue)

		switch annoName {
		case ShowEventsAnnoName:
			boolValue, err := strconv.ParseBool(annoValue)
			if err != nil {
				return nil, fmt.Errorf("%s: bool expected: %w", invalidAnnoValueError, err)
			}

			genericSpec.ShowServiceMessages = boolValue
		case TrackTerminationModeAnnoName:
			trackTerminationModeValue := generic.TrackTerminationMode(annoValue)
			values := []generic.TrackTerminationMode{generic.WaitUntilResourceReady, generic.NonBlocking}
			for _, value := range values {
				if value == trackTerminationModeValue {
					genericSpec.TrackTerminationMode = trackTerminationModeValue
					continue mainLoop
				}
			}

			return nil, fmt.Errorf("%w: choose one of %v", invalidAnnoValueError, values)
		case FailModeAnnoName:
			failModeValue := generic.FailMode(annoValue)
			values := []generic.FailMode{generic.IgnoreAndContinueDeployProcess, generic.FailWholeDeployProcessImmediately, generic.HopeUntilEndOfDeployProcess}
			for _, value := range values {
				if value == failModeValue {
					genericSpec.FailMode = failModeValue
					continue mainLoop
				}
			}

			return nil, fmt.Errorf("%w: choose one of %v", invalidAnnoValueError, values)
		case NoActivityTimeoutName:
			noActivityTimeout, err := time.ParseDuration(annoValue)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", invalidAnnoValueError, err)
			} else if noActivityTimeout.Seconds() < 1 {
				return nil, fmt.Errorf("%w: can't be less than 1 second", invalidAnnoValueError)
			}

			genericSpec.NoActivityTimeout = &noActivityTimeout
		}
	}

	return genericSpec, nil
}

func (waiter *ResourcesWaiter) WatchUntilReady(ctx context.Context, resources helm_kube.ResourceList, timeout time.Duration) error {
	if os.Getenv("WERF_DISABLE_RESOURCES_WAITER") == "1" {
		return nil
	}

	if waiter.KubeInitializer != nil {
		if err := waiter.KubeInitializer.Init(ctx); err != nil {
			return fmt.Errorf("kube initializer failed: %w", err)
		}
	}

	specs, err := makeMultitrackSpecsFromResList(ctx, resources, timeout, waiter.HooksStatusProgressPeriod)
	if err != nil {
		return fmt.Errorf("error making multitrack specs: %w", err)
	}

	// NOTE: use context from resources-waiter object here, will be changed in helm 3
	logboek.Context(ctx).LogOptionalLn()
	return logboek.Context(ctx).LogProcess("Waiting for helm hooks termination").
		DoError(func() error {
			return multitrack.Multitrack(kube.Client, *specs, multitrack.MultitrackOptions{
				StatusProgressPeriod: waiter.HooksStatusProgressPeriod,
				Options: tracker.Options{
					Timeout:      timeout,
					LogsFromTime: waiter.LogsFromTime,
				},
				DynamicClient:   kube.DynamicClient,
				DiscoveryClient: kube.CachedDiscoveryClient,
				Mapper:          kube.Mapper,
			})
		})
}

func asVersioned(info *resource.Info) runtime.Object {
	convertor := runtime.ObjectConvertor(scheme.Scheme)
	groupVersioner := runtime.GroupVersioner(schema.GroupVersions(scheme.Scheme.PrioritizedVersionsAllGroups()))
	if info.Mapping != nil {
		groupVersioner = info.Mapping.GroupVersionKind.GroupVersion()
	}
	if obj, err := convertor.ConvertToVersion(info.Object, groupVersioner); err == nil {
		return obj
	}
	return info.Object
}

func (waiter *ResourcesWaiter) WaitUntilDeleted(ctx context.Context, specs []*helm_kube.ResourcesWaiterDeleteResourceSpec, timeout time.Duration) error {
	if len(specs) == 0 {
		return nil
	}

	if waiter.KubeInitializer != nil {
		if err := waiter.KubeInitializer.Init(ctx); err != nil {
			return fmt.Errorf("kube initializer failed: %w", err)
		}
	}

	var eliminationSpecs []*elimination.EliminationTrackerSpec
	for _, spec := range specs {
		eliminationSpecs = append(eliminationSpecs, &elimination.EliminationTrackerSpec{
			ResourceName:         spec.ResourceName,
			Namespace:            spec.Namespace,
			GroupVersionResource: spec.GroupVersionResource,
		})
	}

	var resourcesDescParts []string
	for _, spec := range specs {
		resourcesDescParts = append(resourcesDescParts, fmt.Sprintf("%s/%s", strings.ToLower(spec.GroupVersionResource.Resource), spec.ResourceName))
	}

	return logboek.Context(ctx).Default().LogProcess("Waiting for resources elimination: %s", strings.Join(resourcesDescParts, ", ")).DoError(func() error {
		return elimination.TrackUntilEliminated(ctx, kube.DynamicClient, eliminationSpecs, elimination.EliminationTrackerOptions{Timeout: timeout, StatusProgressPeriod: waiter.StatusProgressPeriod})
	})
}

func makeMultitrackSpecsFromResList(ctx context.Context, resources helm_kube.ResourceList, timeout, statusProgressPeriod time.Duration) (*multitrack.MultitrackSpecs, error) {
	specs := &multitrack.MultitrackSpecs{}

	for _, v := range resources {
		switch value := asVersioned(v).(type) {
		case *appsv1.Deployment:
			spec, err := makeMultitrackSpec(ctx, &value.ObjectMeta, allowedFailuresCountOptions{multiplier: extractSpecReplicas(value.Spec.Replicas), defaultPerReplica: 1}, "deploy")
			if err != nil {
				return nil, fmt.Errorf("cannot track %s %s: %w", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.Deployments = append(specs.Deployments, *spec)
			}
		case *appsv1beta1.Deployment:
			spec, err := makeMultitrackSpec(ctx, &value.ObjectMeta, allowedFailuresCountOptions{multiplier: extractSpecReplicas(value.Spec.Replicas), defaultPerReplica: 1}, "deploy")
			if err != nil {
				return nil, fmt.Errorf("cannot track %s %s: %w", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.Deployments = append(specs.Deployments, *spec)
			}
		case *appsv1beta2.Deployment:
			spec, err := makeMultitrackSpec(ctx, &value.ObjectMeta, allowedFailuresCountOptions{multiplier: extractSpecReplicas(value.Spec.Replicas), defaultPerReplica: 1}, "deploy")
			if err != nil {
				return nil, fmt.Errorf("cannot track %s %s: %w", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.Deployments = append(specs.Deployments, *spec)
			}
		case *extensions.Deployment:
			spec, err := makeMultitrackSpec(ctx, &value.ObjectMeta, allowedFailuresCountOptions{multiplier: extractSpecReplicas(value.Spec.Replicas), defaultPerReplica: 1}, "deploy")
			if err != nil {
				return nil, fmt.Errorf("cannot track %s %s: %w", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.Deployments = append(specs.Deployments, *spec)
			}
		case *extensions.DaemonSet:
			// TODO: multiplier equals 3 because typically there are only 3 nodes in the cluster.
			// TODO: It is better to fetch number of nodes dynamically, but in the most cases multiplier=3 will work ok.
			spec, err := makeMultitrackSpec(ctx, &value.ObjectMeta, allowedFailuresCountOptions{multiplier: 3, defaultPerReplica: 1}, "ds")
			if err != nil {
				return nil, fmt.Errorf("cannot track %s %s: %w", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.DaemonSets = append(specs.DaemonSets, *spec)
			}
		case *appsv1.DaemonSet:
			// TODO: multiplier equals 3 because typically there are only 3 nodes in the cluster.
			// TODO: It is better to fetch number of nodes dynamically, but in the most cases multiplier=3 will work ok.
			spec, err := makeMultitrackSpec(ctx, &value.ObjectMeta, allowedFailuresCountOptions{multiplier: 3, defaultPerReplica: 1}, "ds")
			if err != nil {
				return nil, fmt.Errorf("cannot track %s %s: %w", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.DaemonSets = append(specs.DaemonSets, *spec)
			}
		case *appsv1beta2.DaemonSet:
			// TODO: multiplier equals 3 because typically there are only 3 nodes in the cluster.
			// TODO: It is better to fetch number of nodes dynamically, but in the most cases multiplier=3 will work ok.
			spec, err := makeMultitrackSpec(ctx, &value.ObjectMeta, allowedFailuresCountOptions{multiplier: 3, defaultPerReplica: 1}, "ds")
			if err != nil {
				return nil, fmt.Errorf("cannot track %s %s: %w", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.DaemonSets = append(specs.DaemonSets, *spec)
			}
		case *appsv1.StatefulSet:
			spec, err := makeMultitrackSpec(ctx, &value.ObjectMeta, allowedFailuresCountOptions{multiplier: extractSpecReplicas(value.Spec.Replicas), defaultPerReplica: 1}, "sts")
			if err != nil {
				return nil, fmt.Errorf("cannot track %s %s: %w", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.StatefulSets = append(specs.StatefulSets, *spec)
			}
		case *appsv1beta1.StatefulSet:
			spec, err := makeMultitrackSpec(ctx, &value.ObjectMeta, allowedFailuresCountOptions{multiplier: extractSpecReplicas(value.Spec.Replicas), defaultPerReplica: 1}, "sts")
			if err != nil {
				return nil, fmt.Errorf("cannot track %s %s: %w", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.StatefulSets = append(specs.StatefulSets, *spec)
			}
		case *appsv1beta2.StatefulSet:
			spec, err := makeMultitrackSpec(ctx, &value.ObjectMeta, allowedFailuresCountOptions{multiplier: extractSpecReplicas(value.Spec.Replicas), defaultPerReplica: 1}, "sts")
			if err != nil {
				return nil, fmt.Errorf("cannot track %s %s: %w", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.StatefulSets = append(specs.StatefulSets, *spec)
			}
		case *batchv1.Job:
			spec, err := makeMultitrackSpec(ctx, &value.ObjectMeta, allowedFailuresCountOptions{multiplier: 1, defaultPerReplica: 0}, "job")
			if err != nil {
				return nil, fmt.Errorf("cannot track %s %s: %w", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.Jobs = append(specs.Jobs, *spec)
			}
		case *flaggerv1beta1.Canary:
			spec, err := makeMultitrackSpec(ctx, &value.ObjectMeta, allowedFailuresCountOptions{multiplier: 1, defaultPerReplica: 0}, "canary")
			if err != nil {
				return nil, fmt.Errorf("cannot track %s %s: %w", value.Kind, value.Name, err)
			}
			if spec != nil {
				specs.Canaries = append(specs.Canaries, *spec)
			}
		default:
			obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(v.Object)
			if err != nil {
				return nil, fmt.Errorf("error converting object to unstructured: %w", err)
			}

			object := unstructured.Unstructured{
				Object: obj,
			}

			resourceID := resid.NewResourceID(object.GetName(), object.GroupVersionKind(), resid.NewResourceIDOptions{
				Namespace: object.GetNamespace(),
			})

			if spec, err := makeGenericSpec(ctx, resourceID, statusProgressPeriod, timeout, object.GetAnnotations()); err != nil {
				logboek.Context(ctx).Warn().LogLn()
				logboek.Context(ctx).Warn().LogF("WARNING %s\n", err)
			} else if spec != nil {
				specs.Generics = append(specs.Generics, spec)
			}
		}
	}

	return specs, nil
}
