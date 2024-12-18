package helm_for_werf_helm

import "context"

type KubeInitializer interface {
	Init(ctx context.Context) error
}
