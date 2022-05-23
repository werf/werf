package allow_list

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/werf/logboek"
)

func DeployedDockerImages(ctx context.Context, kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]string, error) {
	var deployedDockerImages []string

	images, err := getPodsImages(kubernetesClient, kubernetesNamespace)
	if err != nil {
		return nil, fmt.Errorf("cannot get Pods images: %w", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	images, err = getReplicationControllersImages(kubernetesClient, kubernetesNamespace)
	if err != nil {
		return nil, fmt.Errorf("cannot get ReplicationControllers images: %w", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	images, err = getDeploymentsImages(kubernetesClient, kubernetesNamespace)
	if err != nil {
		return nil, fmt.Errorf("cannot get Deployments images: %w", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	images, err = getStatefulSetsImages(kubernetesClient, kubernetesNamespace)
	if err != nil {
		return nil, fmt.Errorf("cannot get StatefulSets images: %w", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	images, err = getDaemonSetsImages(kubernetesClient, kubernetesNamespace)
	if err != nil {
		return nil, fmt.Errorf("cannot get DaemonSets images: %w", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	images, err = getReplicaSetsImages(kubernetesClient, kubernetesNamespace)
	if err != nil {
		return nil, fmt.Errorf("cannot get ReplicaSets images: %w", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	images, err = getCronJobsImages(kubernetesClient, kubernetesNamespace)
	if err != nil {
		return nil, fmt.Errorf("cannot get CronJobs images: %w", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	images, err = getJobsImages(ctx, kubernetesClient, kubernetesNamespace)
	if err != nil {
		return nil, fmt.Errorf("cannot get Jobs images: %w", err)
	}

	deployedDockerImages = append(deployedDockerImages, images...)

	return deployedDockerImages, nil
}

func getPodsImages(kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]string, error) {
	var images []string
	list, err := kubernetesClient.CoreV1().Pods(kubernetesNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, pod := range list.Items {
		for _, container := range append(
			pod.Spec.Containers,
			pod.Spec.InitContainers...,
		) {
			images = append(images, container.Image)
		}
	}

	return images, nil
}

func getReplicationControllersImages(kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]string, error) {
	var images []string
	list, err := kubernetesClient.CoreV1().ReplicationControllers(kubernetesNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, replicationController := range list.Items {
		for _, container := range append(
			replicationController.Spec.Template.Spec.Containers,
			replicationController.Spec.Template.Spec.InitContainers...,
		) {
			images = append(images, container.Image)
		}
	}

	return images, nil
}

func getDeploymentsImages(kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]string, error) {
	var images []string
	list, err := kubernetesClient.AppsV1().Deployments(kubernetesNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, deployment := range list.Items {
		for _, container := range append(
			deployment.Spec.Template.Spec.Containers,
			deployment.Spec.Template.Spec.InitContainers...,
		) {
			images = append(images, container.Image)
		}
	}

	return images, nil
}

func getStatefulSetsImages(kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]string, error) {
	var images []string
	list, err := kubernetesClient.AppsV1().StatefulSets(kubernetesNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, statefulSet := range list.Items {
		for _, container := range append(
			statefulSet.Spec.Template.Spec.Containers,
			statefulSet.Spec.Template.Spec.InitContainers...,
		) {
			images = append(images, container.Image)
		}
	}

	return images, nil
}

func getDaemonSetsImages(kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]string, error) {
	var images []string
	list, err := kubernetesClient.AppsV1().DaemonSets(kubernetesNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, daemonSets := range list.Items {
		for _, container := range append(
			daemonSets.Spec.Template.Spec.Containers,
			daemonSets.Spec.Template.Spec.InitContainers...,
		) {
			images = append(images, container.Image)
		}
	}

	return images, nil
}

func getReplicaSetsImages(kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]string, error) {
	var images []string
	list, err := kubernetesClient.AppsV1().ReplicaSets(kubernetesNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, replicaSet := range list.Items {
		for _, container := range append(
			replicaSet.Spec.Template.Spec.Containers,
			replicaSet.Spec.Template.Spec.InitContainers...,
		) {
			images = append(images, container.Image)
		}
	}

	return images, nil
}

func getCronJobsImages(kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]string, error) {
	var images []string
	list, err := kubernetesClient.BatchV1beta1().CronJobs(kubernetesNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, cronJob := range list.Items {
		for _, container := range append(
			cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers,
			cronJob.Spec.JobTemplate.Spec.Template.Spec.InitContainers...,
		) {
			images = append(images, container.Image)
		}
	}

	return images, nil
}

func getJobsImages(ctx context.Context, kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]string, error) {
	var images []string
	list, err := kubernetesClient.BatchV1().Jobs(kubernetesNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

FindActiveJobs:
	for _, job := range list.Items {
		for _, c := range job.Status.Conditions {
			switch c.Type {
			case batchv1.JobComplete:
				if c.Status == corev1.ConditionTrue {
					logboek.Context(ctx).Debug().LogF("Ignore complete job/%s: images in this resource are not used anymore and can be safely removed\n", job.Name)
					continue FindActiveJobs
				}
			case batchv1.JobFailed:
				if c.Status == corev1.ConditionTrue {
					logboek.Context(ctx).Debug().LogF("Ignore failed job/%s: images in this resource are not used anymore and can be safely removed\n", job.Name)
					continue FindActiveJobs
				}
			}
		}

		for _, container := range append(
			job.Spec.Template.Spec.Containers,
			job.Spec.Template.Spec.InitContainers...,
		) {
			images = append(images, container.Image)
		}
	}

	return images, nil
}
