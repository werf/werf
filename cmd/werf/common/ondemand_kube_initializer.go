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

	BearerToken     string
	BearerTokenFile string

	initialized bool
}

// TODO(major): why do we even need this? Can we get rid from these kube dependencies for building?
func SetupOndemandKubeInitializer(
	kubeContext, kubeConfig, kubeConfigBase64 string,
	kubeConfigPathMergeList []string,
	bearerToken, bearerTokenFile string,
) {
	ondemandKubeInitializer = &OndemandKubeInitializer{
		KubeContext:             kubeContext,
		KubeConfig:              kubeConfig,
		KubeConfigBase64:        kubeConfigBase64,
		KubeConfigPathMergeList: kubeConfigPathMergeList,
		BearerToken:             bearerToken,
		BearerTokenFile:         bearerTokenFile,
	}
}

func GetOndemandKubeInitializer() *OndemandKubeInitializer {
	return ondemandKubeInitializer
}

func (initializer *OndemandKubeInitializer) Init(ctx context.Context) error {
	if initializer.initialized {
		return nil
	}

	kubeOpts := kube.KubeConfigOptions{
		Context:             initializer.KubeContext,
		ConfigPath:          initializer.KubeConfig,
		ConfigDataBase64:    initializer.KubeConfigBase64,
		ConfigPathMergeList: initializer.KubeConfigPathMergeList,
		BearerToken:         initializer.BearerToken,
		BearerTokenFile:     initializer.BearerTokenFile,
	}

	if err := kube.Init(kube.InitOptions{KubeConfigOptions: kubeOpts}); err != nil {
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
