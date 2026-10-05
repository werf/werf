package storage

import (
	"context"

	"github.com/werf/werf/v2/pkg/container_backend"
	"github.com/werf/werf/v2/pkg/docker_registry"
	registry_api "github.com/werf/werf/v2/pkg/docker_registry/api"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/opstats"
)

var _ docker_registry.Interface = (*brokenStageRegistry)(nil)

var _ container_backend.ContainerBackend = (*brokenStageBackend)(nil)

type brokenStageRegistry struct {
	docker_registry.Interface
	err error
}

func (r *brokenStageRegistry) GetRepoImage(_ context.Context, _ string) (*image.Info, error) {
	return nil, r.err
}

func (r *brokenStageRegistry) MutateAndPushImage(_ context.Context, _, _ string, _ ...registry_api.MutateOption) error {
	return r.err
}

type brokenStageBackend struct {
	container_backend.ContainerBackend
	err error
}

func (b *brokenStageBackend) PullImageFromRegistry(_ context.Context, _ container_backend.LegacyImageInterface) error {
	return b.err
}

func brokenCount(collector *opstats.Collector) int {
	for _, e := range collector.EventSummary() {
		if e.Event == opstats.EventStageBroken {
			return e.Count
		}
	}
	return 0
}
