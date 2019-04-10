package helm

import (
	"io"

	"k8s.io/helm/pkg/kube"
	"k8s.io/helm/pkg/tiller/environment"
)

type WerfKubeClient struct {
	KubeClient environment.KubeClient
}

func (client *WerfKubeClient) Create(namespace string, reader io.Reader, timeout int64, shouldWait bool) error {
	return client.KubeClient.Create(namespace, reader, timeout, shouldWait)
}

func (client *WerfKubeClient) Get(namespace string, reader io.Reader) (string, error) {
	return client.KubeClient.Get(namespace, reader)
}

func (client *WerfKubeClient) Delete(namespace string, reader io.Reader) error {
	return client.KubeClient.Delete(namespace, reader)
}

func (client *WerfKubeClient) WatchUntilReady(namespace string, reader io.Reader, timeout int64, shouldWait bool) error {
	return client.KubeClient.WatchUntilReady(namespace, reader, timeout, shouldWait)
}

func (client *WerfKubeClient) Update(namespace string, originalReader, modifiedReader io.Reader, force bool, recreate bool, timeout int64, shouldWait bool) error {
	return client.KubeClient.Update(namespace, originalReader, modifiedReader, force, recreate, timeout, shouldWait)
}

func (client *WerfKubeClient) UpdateWithOptions(namespace string, originalReader, modifiedReader io.Reader, opts kube.UpdateOptions) error {
	return client.KubeClient.UpdateWithOptions(namespace, originalReader, modifiedReader, opts)
}

func (client *WerfKubeClient) Build(namespace string, reader io.Reader) (kube.Result, error) {
	return client.KubeClient.Build(namespace, reader)
}

func (client *WerfKubeClient) BuildUnstructured(namespace string, reader io.Reader) (kube.Result, error) {
	return client.KubeClient.BuildUnstructured(namespace, reader)
}

// func (client *WerfKubeClient) WaitAndGetCompletedPodPhase(namespace string, reader io.Reader, timeout time.Duration) (v1.PodPhase, error) {
// 	return client.KubeClient.WaitAndGetCompletedPodPhase(namespace, reader, timeout).(v1.PodPhase)
// }
