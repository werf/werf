package build

import (
	"context"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"
	"github.com/werf/werf/pkg/image"
)

type ExportPhase struct {
	BasePhase
	ExportPhaseOptions
}

type ExportPhaseOptions struct {
	ExportTagFuncList []image.ExportTagFunc
}

func NewExportPhase(c *Conveyor, opts ExportPhaseOptions) *ExportPhase {
	return &ExportPhase{
		BasePhase:          BasePhase{c},
		ExportPhaseOptions: opts,
	}
}

func (phase *ExportPhase) Name() string {
	return "export"
}

func (phase *ExportPhase) AfterImageStages(ctx context.Context, img *Image) error {
	if img.isArtifact {
		return nil
	}

	if err := phase.exportLastStageImage(ctx, img); err != nil {
		return err
	}

	return nil
}

func (phase *ExportPhase) exportLastStageImage(ctx context.Context, img *Image) error {
	if len(phase.ExportTagFuncList) == 0 {
		return nil
	}

	return logboek.Context(ctx).Default().LogProcess("Exporting image...").
		Options(func(options types.LogProcessOptionsInterface) {
			options.Style(style.Highlight())
		}).
		DoError(func() error {
			for _, tagFunc := range phase.ExportTagFuncList {
				tag := tagFunc(img.GetName(), img.GetStageID())
				if err := logboek.Context(ctx).Default().LogProcess("tag %s", tag).
					DoError(func() error {
						stageDesc := img.GetLastNonEmptyStage().GetStageImage().Image.GetStageDescription()
						if err := phase.Conveyor.StorageManager.GetStagesStorage().ExportStage(ctx, stageDesc, tag); err != nil {
							return err
						}

						return nil
					}); err != nil {
					return err
				}
			}

			return nil
		})
}

func (phase *ExportPhase) Clone() Phase {
	u := *phase
	return &u
}
