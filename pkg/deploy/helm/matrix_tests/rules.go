package matrix_tests

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// FIXME: No custom rules support provided and no configuration for rules settings added.
var (
	LintRulesConfiguration = map[string]bool{
		"ContainerNameDuplicates":         true,
		"ContainerEnvVariablesDuplicates": true,
		"ContainerImageTagLatest":         false,
		"ContainerPullPolicyIfNotPresent": false,

		"ObjectRecommendedLabels": false,
	}
)

func getContainers(object unstructured.Unstructured) ([]v1.Container, error) {
	var containers []v1.Container

	converter := runtime.DefaultUnstructuredConverter

	switch object.GetKind() {
	case "Deployment":
		deployment := new(appsv1.Deployment)
		converter.FromUnstructured(object.Object, deployment)

		containers = deployment.Spec.Template.Spec.Containers
	case "DaemonSet":
		daemonSet := new(appsv1.DaemonSet)
		converter.FromUnstructured(object.Object, daemonSet)

		containers = daemonSet.Spec.Template.Spec.Containers
	case "StatefulSet":
		statefulSet := new(appsv1.StatefulSet)
		converter.FromUnstructured(object.Object, statefulSet)

		containers = statefulSet.Spec.Template.Spec.Containers
	case "Pod":
		pod := new(v1.Pod)
		converter.FromUnstructured(object.Object, pod)

		containers = pod.Spec.Containers
	case "Job":
		job := new(batchv1.Job)
		converter.FromUnstructured(object.Object, job)

		containers = job.Spec.Template.Spec.Containers
	case "CronJob":
		cronJob := new(batchv1beta1.CronJob)
		converter.FromUnstructured(object.Object, cronJob)

		containers = cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers
	}
	return containers, nil
}

func applyContainerRules(object unstructured.Unstructured) error {
	containers, err := getContainers(object)
	if err != nil {
		return err
	}
	if len(containers) == 0 {
		return nil
	}

	if LintRulesConfiguration["ContainerNameDuplicates"] {
		err = containerNameDuplicates(containers)
		if err != nil {
			return err
		}
	}

	if LintRulesConfiguration["ContainerEnvVariablesDuplicates"] {
		err = containerEnvVariablesDuplicates(containers)
		if err != nil {
			return err
		}
	}

	if LintRulesConfiguration["ContainerImageTagLatest"] {
		err = containerImageTagLatest(containers)
		if err != nil {
			return err
		}
	}

	if LintRulesConfiguration["ContainerPullPolicyIfNotPresent"] {
		err = containerImagePullPolicyIfNotPresent(containers)
		if err != nil {
			return err
		}
	}
	return nil
}

func containerNameDuplicates(containers []v1.Container) error {
	names := make(map[string]struct{})
	for _, c := range containers {
		if _, ok := names[c.Name]; ok {
			return fmt.Errorf("container %q already exists", c.Name)
		}
		names[c.Name] = struct{}{}
	}
	return nil
}

func containerEnvVariablesDuplicates(containers []v1.Container) error {
	for _, c := range containers {
		envVariables := make(map[string]struct{})
		for _, variable := range c.Env {
			if _, ok := envVariables[variable.Name]; ok {
				return fmt.Errorf("container %q has two env variables with same name: %s", c.Name, variable.Name)
			}
			envVariables[variable.Name] = struct{}{}
		}
	}
	return nil
}

func containerImageTagLatest(containers []v1.Container) error {
	for _, c := range containers {
		imageParts := strings.Split(c.Image, ":")
		if len(imageParts) != 2 {
			return fmt.Errorf("can't parse image for container %q", c.Name)
		}
		if imageParts[1] == "latest" {
			return fmt.Errorf("image tag \"latest\" used for container %q", c.Name)
		}
	}
	return nil
}

func containerImagePullPolicyIfNotPresent(containers []v1.Container) error {
	for _, c := range containers {
		if c.ImagePullPolicy != "" && c.ImagePullPolicy != "IfNotPresent" {
			return fmt.Errorf("container %q imagePullPolicy should be unspecified or \"IfNotPresent\", recive: %s", c.Name, c.ImagePullPolicy)
		}
	}
	return nil
}

func applyObjectRules(object unstructured.Unstructured) error {

	if LintRulesConfiguration["ObjectRecommendedLabels"] {
		err := objectRecommendedLabels(object)
		if err != nil {
			return err
		}
	}

	return nil
}

func objectRecommendedLabels(object unstructured.Unstructured) error {
	labels := object.GetLabels()
	if _, ok := labels["module"]; !ok {
		return fmt.Errorf("object %q does not have label \"module\": %v", object.GetName(), labels)
	}
	if _, ok := labels["heritage"]; !ok {
		return fmt.Errorf("object %q does not have label \"heritage\": %v", object.GetName(), labels)
	}
	return nil
}

func ApplyLintRules(objectStore UnstructuredObjectStore) error {
	for _, object := range objectStore.Storage {
		err := applyObjectRules(object)
		if err != nil {
			return err
		}

		err = applyContainerRules(object)
		if err != nil {
			return err
		}

	}
	return nil
}
