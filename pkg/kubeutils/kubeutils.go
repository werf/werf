package kubeutils

import (
	"context"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"
)

// Not namespaced resource specifies without namespace
func RemoveResourceAndWaitUntilRemoved(ctx context.Context, name, kind, namespace string) error {
	isNamespacedResource := namespace != ""

	groupVersionResource, err := kube.GroupVersionResourceByKind(kube.Client, kind)
	if err != nil {
		return err
	}

	var res dynamic.ResourceInterface
	if isNamespacedResource {
		res = kube.DynamicClient.Resource(groupVersionResource).Namespace(namespace)
	} else {
		res = kube.DynamicClient.Resource(groupVersionResource)
	}

	isExist := func() (bool, error) {
		_, err := res.Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}

			return true, err
		}

		return true, nil
	}

	exist, err := isExist()
	if err != nil {
		return err
	} else if !exist {
		return nil
	}

	var logProcessMsg string
	if isNamespacedResource {
		logProcessMsg = fmt.Sprintf("Deleting %s/%s from namespace %q", groupVersionResource.Resource, name, namespace)
	} else {
		logProcessMsg = fmt.Sprintf("Deleting %s/%s", groupVersionResource.Resource, name)
	}

	return logboek.Context(ctx).LogProcessInline(logProcessMsg).
		Options(func(options types.LogProcessInlineOptionsInterface) {
			options.Style(style.Details())
		}).
		DoError(func() error {
			deletePropagation := metav1.DeletePropagationForeground
			deleteOptions := metav1.DeleteOptions{
				PropagationPolicy: &deletePropagation,
			}
			err = res.Delete(context.Background(), name, deleteOptions)
			if err != nil {
				return err
			}

			for {
				exist, err := isExist()
				if err != nil {
					return err
				} else if !exist {
					break
				}

				time.Sleep(500 * time.Millisecond)
			}

			return nil
		})
}
