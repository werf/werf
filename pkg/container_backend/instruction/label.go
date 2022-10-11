package instruction

import (
	"context"
	"fmt"

	"github.com/werf/werf/pkg/buildah"
	"github.com/werf/werf/pkg/container_backend/build_context"
	dockerfile_instruction "github.com/werf/werf/pkg/dockerfile/instruction"
)

type Label struct {
	dockerfile_instruction.Label
}

func NewLabel(i dockerfile_instruction.Label) *Label {
	return &Label{Label: i}
}

func (i *Label) UsesBuildContext() bool {
	return false
}

func (i *Label) LabelsAsList() []string {
	var labels []string
	for k, v := range i.Labels {
		labels = append(labels, fmt.Sprintf("%s=%s", k, v))
	}
	return labels
}

func (i *Label) Apply(ctx context.Context, containerName string, drv buildah.Buildah, drvOpts buildah.CommonOpts, buildContext *build_context.BuildContext) error {
	if err := drv.Config(ctx, containerName, buildah.ConfigOpts{CommonOpts: drvOpts, Labels: i.LabelsAsList()}); err != nil {
		return fmt.Errorf("error setting labels %v for container %s: %w", i.LabelsAsList(), containerName, err)
	}
	return nil
}
