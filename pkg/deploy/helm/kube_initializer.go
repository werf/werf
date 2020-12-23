package helm

import "context"

type KubeInitializer interface {
	Init(ctx context.Context) error
}
