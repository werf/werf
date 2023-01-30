package instruction

import (
	"context"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend"
)

type Label struct {
	instructions.LabelCommand
}

func NewLabel(i instructions.LabelCommand) *Label {
	return &Label{LabelCommand: i}
}

func (i *Label) UsesBuildContext() bool {
	return false
}

func (i *Label) LabelsAsList() []string {
	var labels []string
	for _, item := range i.Labels {
		labels = append(labels, fmt.Sprintf("%s=%s", item.Key, item.Value))
	}
	return labels
}

func (i *Label) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContextArchive container_backend.BuildContextArchiver) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{CommonOpts: drvOpts, Labels: i.LabelsAsList()}); err != nil {
		return fmt.Errorf("error setting labels %v for container %s: %w", i.LabelsAsList(), containerName, err)
	}
	return nil
}
