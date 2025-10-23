package common

import (
	"context"
	"fmt"

	kube_legacy "github.com/werf/kubedog-for-werf-helm/pkg/kube"
	"github.com/werf/kubedog/pkg/kube"
)

var ondemandKubeInitializer *OndemandKubeInitializer

type OndemandKubeInitializer struct {
	KubeContext             string
	KubeConfig              string
	KubeConfigBase64        string
	KubeConfigPathMergeList []string

	initialized bool
}

// TODO(v3): why do we even need this? Can we get rid from these kube dependencies for building?
func SetupOndemandKubeInitializer(kubeContext, kubeConfig, kubeConfigBase64 string, kubeConfigPathMergeList []string) {
	ondemandKubeInitializer = &OndemandKubeInitializer{
		KubeContext:             kubeContext,
		KubeConfig:              kubeConfig,
		KubeConfigBase64:        kubeConfigBase64,
		KubeConfigPathMergeList: kubeConfigPathMergeList,
	}
}

func GetOndemandKubeInitializer() *OndemandKubeInitializer {
	return ondemandKubeInitializer
}

func (initializer *OndemandKubeInitializer) Init(ctx context.Context) error {
	if initializer.initialized {
		return nil
	}

	if err := kube.Init(kube.InitOptions{KubeConfigOptions: kube.KubeConfigOptions{
		Context:             initializer.KubeContext,
		ConfigPath:          initializer.KubeConfig,
		ConfigDataBase64:    initializer.KubeConfigBase64,
		ConfigPathMergeList: initializer.KubeConfigPathMergeList,
	}}); err != nil {
		return fmt.Errorf("cannot initialize kube: %w", err)
	}

	if err := kube_legacy.Init(kube_legacy.InitOptions{KubeConfigOptions: kube_legacy.KubeConfigOptions{
		Context:             initializer.KubeContext,
		ConfigPath:          initializer.KubeConfig,
		ConfigDataBase64:    initializer.KubeConfigBase64,
		ConfigPathMergeList: initializer.KubeConfigPathMergeList,
	}}); err != nil {
		return fmt.Errorf("cannot initialize legacy kube: %w", err)
	}

	if err := InitKubedog(ctx); err != nil {
		return fmt.Errorf("cannot init kubedog: %w", err)
	}

	initializer.initialized = true

	return nil
}
