package helm

import (
	"io"

	"k8s.io/helm/pkg/kube"
)

type KubeClient struct {
	*kube.Client

	createPerformHookFunc func() error
	updatePerformHookFunc func() error
}

func (c *KubeClient) CreateWithOptions(namespace string, reader io.Reader, opts kube.CreateOptions) error {
	opts.PostPerformHook = c.createPerformHookFunc
	return c.Client.CreateWithOptions(namespace, reader, opts)
}

func (c *KubeClient) UpdateWithOptions(namespace string, originalReader, targetReader io.Reader, opts kube.UpdateOptions) error {
	opts.PostPerformHook = c.updatePerformHookFunc
	return c.Client.UpdateWithOptions(namespace, originalReader, targetReader, opts)
}
