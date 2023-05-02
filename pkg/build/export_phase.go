package build

import (
	"context"
	"fmt"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"k8s.io/utils/strings/slices"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"
	build_image "github.com/werf/werf/pkg/build/image"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/storage"
)

type ExportPhase struct {
	BasePhase
	ExportPhaseOptions
}

type ExportPhaseOptions struct {
	ExportImageNameList []string
	ExportTagFuncList   []image.ExportTagFunc
	MutateConfigFunc    func(config v1.Config) (v1.Config, error)
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

func (phase *ExportPhase) AfterImages(ctx context.Context) error {
	if len(phase.ExportTagFuncList) == 0 {
		return nil
	}

	targetPlatforms, err := phase.Conveyor.GetTargetPlatforms()
	if err != nil {
		return fmt.Errorf("unable to get target platforms: %w", err)
	}
	if len(targetPlatforms) == 0 {
		targetPlatforms = []string{phase.Conveyor.ContainerBackend.GetDefaultPlatform()}
	}

	if len(targetPlatforms) == 1 {
		// single platform mode
		for _, desc := range phase.Conveyor.imagesTree.GetImagesByName(true) {
			_, images := desc.Unpair()
			img := images[0]
			if !slices.Contains(phase.ExportImageNameList, img.Name) {
				continue
			}
			if err := phase.exportImage(ctx, img); err != nil {
				return fmt.Errorf("unable to export image %q: %w", img.Name, err)
			}
		}
	} else {
		// FIXME(multiarch): Support multiplatform manifest by pushing local images to repo first, then create manifest list.
		// FIXME(multiarch): Also support multiplatform manifest in werf build command in local mode with enabled final-repo.
		if _, isLocal := phase.Conveyor.StorageManager.GetStagesStorage().(*storage.LocalStagesStorage); isLocal {
			return fmt.Errorf("export command is not supported in multiplatform mode")
		}

		// multiplatform mode
		for _, img := range phase.Conveyor.imagesTree.GetMultiplatformImages() {
			if !slices.Contains(phase.ExportImageNameList, img.Name) {
				continue
			}
			if err := phase.exportMultiplatformImage(ctx, img); err != nil {
				return fmt.Errorf("unable to export multiplatform image %q: %w", img.Name, err)
			}
		}
	}

	return nil
}

func (phase *ExportPhase) exportMultiplatformImage(ctx context.Context, img *build_image.MultiplatformImage) error {
	return logboek.Context(ctx).Default().LogProcess("Exporting image...").
		Options(func(options types.LogProcessOptionsInterface) {
			options.Style(style.Highlight())
		}).
		DoError(func() error {
			for _, tagFunc := range phase.ExportTagFuncList {
				tag := tagFunc(img.Name, img.GetStageID().String())
				if err := logboek.Context(ctx).Default().LogProcess("tag %s", tag).
					DoError(func() error {
						desc := img.GetStageDescription()
						if err := phase.Conveyor.StorageManager.GetStagesStorage().ExportStage(ctx, desc, tag, phase.MutateConfigFunc); err != nil {
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

func (phase *ExportPhase) exportImage(ctx context.Context, img *build_image.Image) error {
	if !slices.Contains(phase.ExportImageNameList, img.Name) {
		return nil
	}
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
						if err := phase.Conveyor.StorageManager.GetStagesStorage().ExportStage(ctx, stageDesc, tag, phase.MutateConfigFunc); err != nil {
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
