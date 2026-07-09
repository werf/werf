package container_backend

import (
	"context"

	"github.com/werf/werf/v2/pkg/docker"
)

type LegacyStageImage struct {
	*legacyBaseImage
	builtID        string
	targetPlatform string

	buildServiceLabels map[string]string
}

func NewLegacyStageImage(name string, containerBackend ContainerBackend, targetPlatform string) *LegacyStageImage {
	stage := &LegacyStageImage{}
	stage.legacyBaseImage = newLegacyBaseImage(name, containerBackend)
	stage.targetPlatform = targetPlatform
	return stage
}

func (i *LegacyStageImage) GetTargetPlatform() string {
	return i.targetPlatform
}

func (i *LegacyStageImage) GetCopy() LegacyImageInterface {
	ni := NewLegacyStageImage(i.name, i.ContainerBackend, i.targetPlatform)
	if stageDesc := i.GetStageDesc(); stageDesc != nil {
		ni.SetStageDesc(stageDesc.GetCopy())
	} else if info := i.GetInfo(); info != nil {
		ni.SetInfo(info.GetCopy())
	}
	return ni
}

func (i *LegacyStageImage) GetID() string {
	if stageDesc := i.legacyBaseImage.GetStageDesc(); stageDesc != nil && stageDesc.Info != nil {
		return stageDesc.Info.Name
	}
	return i.legacyBaseImage.Name()
}

func (i *LegacyStageImage) SetBuiltID(builtID string) {
	i.builtID = builtID
}

func (i *LegacyStageImage) BuiltID() string {
	return i.builtID
}

func (i *LegacyStageImage) Tag(ctx context.Context, name string) error {
	return docker.CliTag(ctx, i.GetID(), name)
}

func (i *LegacyStageImage) Pull(ctx context.Context) error {
	var args []string
	if i.targetPlatform != "" {
		args = append(args, "--platform", i.targetPlatform)
	}
	args = append(args, i.name)

	if err := docker.CliPullWithRetries(ctx, args...); err != nil {
		return err
	}

	i.legacyBaseImage.UnsetInfo()

	return nil
}

func (i *LegacyStageImage) Push(ctx context.Context) error {
	return docker.CliPushWithRetries(ctx, i.name)
}

func (i *LegacyStageImage) SetBuildServiceLabels(labels map[string]string) {
	i.buildServiceLabels = labels
}

func (i *LegacyStageImage) GetBuildServiceLabels() map[string]string {
	return i.buildServiceLabels
}
