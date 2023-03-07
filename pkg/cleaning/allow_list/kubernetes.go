package allow_list

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/werf/logboek"
)

type DeployedImage struct {
	Name           string
	ResourcesNames []string
}

func AppendDeployedImages(deployedImages []*DeployedImage, newDeployedImages ...*DeployedImage) (res []*DeployedImage) {
	for _, desc := range deployedImages {
		res = append(res, &DeployedImage{
			Name:           desc.Name,
			ResourcesNames: desc.ResourcesNames,
		})
	}

AppendNewImages:
	for _, newDesc := range newDeployedImages {
		for _, desc := range res {
			if desc.Name == newDesc.Name {
				desc.ResourcesNames = append(desc.ResourcesNames, newDesc.ResourcesNames...)
				continue AppendNewImages
			}
		}

		res = append(res, &DeployedImage{
			Name:           newDesc.Name,
			ResourcesNames: newDesc.ResourcesNames,
		})
	}

	return
}

func DeployedDockerImages(ctx context.Context, kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]*DeployedImage, error) {
	var deployedDockerImages []*DeployedImage

	images, err := getPodsImages(kubernetesClient, kubernetesNamespace)
	if err != nil {
		return nil, fmt.Errorf("cannot get Pods images: %w", err)
	}
	deployedDockerImages = AppendDeployedImages(deployedDockerImages, images...)

	images, err = getReplicationControllersImages(kubernetesClient, kubernetesNamespace)
	if err != nil {
		return nil, fmt.Errorf("cannot get ReplicationControllers images: %w", err)
	}
	deployedDockerImages = AppendDeployedImages(deployedDockerImages, images...)

	images, err = getDeploymentsImages(kubernetesClient, kubernetesNamespace)
	if err != nil {
		return nil, fmt.Errorf("cannot get Deployments images: %w", err)
	}
	deployedDockerImages = AppendDeployedImages(deployedDockerImages, images...)

	images, err = getStatefulSetsImages(kubernetesClient, kubernetesNamespace)
	if err != nil {
		return nil, fmt.Errorf("cannot get StatefulSets images: %w", err)
	}
	deployedDockerImages = AppendDeployedImages(deployedDockerImages, images...)

	images, err = getDaemonSetsImages(kubernetesClient, kubernetesNamespace)
	if err != nil {
		return nil, fmt.Errorf("cannot get DaemonSets images: %w", err)
	}
	deployedDockerImages = AppendDeployedImages(deployedDockerImages, images...)

	images, err = getReplicaSetsImages(kubernetesClient, kubernetesNamespace)
	if err != nil {
		return nil, fmt.Errorf("cannot get ReplicaSets images: %w", err)
	}
	deployedDockerImages = AppendDeployedImages(deployedDockerImages, images...)

	images, err = getCronJobsImages(ctx, kubernetesClient, kubernetesNamespace)
	if err != nil {
		return nil, fmt.Errorf("cannot get CronJobs images: %w", err)
	}
	deployedDockerImages = AppendDeployedImages(deployedDockerImages, images...)

	images, err = getJobsImages(ctx, kubernetesClient, kubernetesNamespace)
	if err != nil {
		return nil, fmt.Errorf("cannot get Jobs images: %w", err)
	}
	deployedDockerImages = AppendDeployedImages(deployedDockerImages, images...)

	return deployedDockerImages, nil
}

func getPodsImages(kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]*DeployedImage, error) {
	var images []*DeployedImage
	list, err := kubernetesClient.CoreV1().Pods(kubernetesNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, pod := range list.Items {
		for _, container := range append(
			pod.Spec.Containers,
			pod.Spec.InitContainers...,
		) {
			images = AppendDeployedImages(images, &DeployedImage{
				Name:           container.Image,
				ResourcesNames: []string{fmt.Sprintf("ns/%s pod/%s container/%s", pod.Namespace, pod.Name, container.Name)},
			})
		}
	}

	return images, nil
}

func getReplicationControllersImages(kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]*DeployedImage, error) {
	var images []*DeployedImage
	list, err := kubernetesClient.CoreV1().ReplicationControllers(kubernetesNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, replicationController := range list.Items {
		for _, container := range append(
			replicationController.Spec.Template.Spec.Containers,
			replicationController.Spec.Template.Spec.InitContainers...,
		) {
			images = AppendDeployedImages(images, &DeployedImage{
				Name:           container.Image,
				ResourcesNames: []string{fmt.Sprintf("ns/%s rc/%s container/%s", replicationController.Namespace, replicationController.Name, container.Name)},
			})
		}
	}

	return images, nil
}

func getDeploymentsImages(kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]*DeployedImage, error) {
	var images []*DeployedImage
	list, err := kubernetesClient.AppsV1().Deployments(kubernetesNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, deployment := range list.Items {
		for _, container := range append(
			deployment.Spec.Template.Spec.Containers,
			deployment.Spec.Template.Spec.InitContainers...,
		) {
			images = AppendDeployedImages(images, &DeployedImage{
				Name:           container.Image,
				ResourcesNames: []string{fmt.Sprintf("ns/%s deploy/%s container/%s", deployment.Namespace, deployment.Name, container.Name)},
			})
		}
	}

	return images, nil
}

func getStatefulSetsImages(kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]*DeployedImage, error) {
	var images []*DeployedImage
	list, err := kubernetesClient.AppsV1().StatefulSets(kubernetesNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, statefulSet := range list.Items {
		for _, container := range append(
			statefulSet.Spec.Template.Spec.Containers,
			statefulSet.Spec.Template.Spec.InitContainers...,
		) {
			images = AppendDeployedImages(images, &DeployedImage{
				Name:           container.Image,
				ResourcesNames: []string{fmt.Sprintf("ns/%s sts/%s container/%s", statefulSet.Namespace, statefulSet.Name, container.Name)},
			})
		}
	}

	return images, nil
}

func getDaemonSetsImages(kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]*DeployedImage, error) {
	var images []*DeployedImage
	list, err := kubernetesClient.AppsV1().DaemonSets(kubernetesNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, daemonSet := range list.Items {
		for _, container := range append(
			daemonSet.Spec.Template.Spec.Containers,
			daemonSet.Spec.Template.Spec.InitContainers...,
		) {
			images = AppendDeployedImages(images, &DeployedImage{
				Name:           container.Image,
				ResourcesNames: []string{fmt.Sprintf("ns/%s ds/%s container/%s", daemonSet.Namespace, daemonSet.Name, container.Name)},
			})
		}
	}

	return images, nil
}

func getReplicaSetsImages(kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]*DeployedImage, error) {
	var images []*DeployedImage
	list, err := kubernetesClient.AppsV1().ReplicaSets(kubernetesNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, replicaSet := range list.Items {
		for _, container := range append(
			replicaSet.Spec.Template.Spec.Containers,
			replicaSet.Spec.Template.Spec.InitContainers...,
		) {
			images = AppendDeployedImages(images, &DeployedImage{
				Name:           container.Image,
				ResourcesNames: []string{fmt.Sprintf("ns/%s rs/%s container/%s", replicaSet.Namespace, replicaSet.Name, container.Name)},
			})
		}
	}

	return images, nil
}

func getCronJobsImages(ctx context.Context, kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]*DeployedImage, error) {
	images, err := getCronJobsImagesBatchV1(kubernetesClient, kubernetesNamespace)
	if apierrors.IsNotFound(err) {
		logboek.Context(ctx).Warn().LogF("\n")
		logboek.Context(ctx).Warn().LogF("WARNING: Unable to query CronJobs in batch/v1: %s\n", err)
		logboek.Context(ctx).Warn().LogF("WARNING: Will fallback to CronJobs in batch/v1beta1, which is not officially supported anymore\n")

		if images, fallbackErr := getCronJobsImagesBatchV1beta1(kubernetesClient, kubernetesNamespace); fallbackErr == nil {
			return images, nil
		}
	}
	return images, err
}

func getCronJobsImagesBatchV1(kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]*DeployedImage, error) {
	var images []*DeployedImage

	list, err := kubernetesClient.BatchV1().CronJobs(kubernetesNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, cronJob := range list.Items {
		for _, container := range append(
			cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers,
			cronJob.Spec.JobTemplate.Spec.Template.Spec.InitContainers...,
		) {
			images = AppendDeployedImages(images, &DeployedImage{
				Name:           container.Image,
				ResourcesNames: []string{fmt.Sprintf("ns/%s cronjob/%s container/%s", cronJob.Namespace, cronJob.Name, container.Name)},
			})
		}
	}

	return images, nil
}

func getCronJobsImagesBatchV1beta1(kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]*DeployedImage, error) {
	var images []*DeployedImage

	list, err := kubernetesClient.BatchV1beta1().CronJobs(kubernetesNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, cronJob := range list.Items {
		for _, container := range append(
			cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers,
			cronJob.Spec.JobTemplate.Spec.Template.Spec.InitContainers...,
		) {
			images = AppendDeployedImages(images, &DeployedImage{
				Name:           container.Image,
				ResourcesNames: []string{fmt.Sprintf("ns/%s cronjob/%s container/%s", cronJob.Namespace, cronJob.Name, container.Name)},
			})
		}
	}

	return images, nil
}

func getJobsImages(ctx context.Context, kubernetesClient kubernetes.Interface, kubernetesNamespace string) ([]*DeployedImage, error) {
	var images []*DeployedImage
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
					logboek.Context(ctx).Info().LogF("Ignore complete job/%s: images in this resource are not used anymore and can be safely removed\n", job.Name)
					continue FindActiveJobs
				}
			case batchv1.JobFailed:
				if c.Status == corev1.ConditionTrue {
					logboek.Context(ctx).Info().LogF("Ignore failed job/%s: images in this resource are not used anymore and can be safely removed\n", job.Name)
					continue FindActiveJobs
				}
			}
		}

		for _, container := range append(
			job.Spec.Template.Spec.Containers,
			job.Spec.Template.Spec.InitContainers...,
		) {
			images = AppendDeployedImages(images, &DeployedImage{
				Name:           container.Image,
				ResourcesNames: []string{fmt.Sprintf("ns/%s job/%s container/%s", job.Namespace, job.Name, container.Name)},
			})
		}
	}

	return images, nil
}
