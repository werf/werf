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
	"github.com/werf/werf/pkg/util"
)

type Exporter struct {
	ExportOptions
	Conveyor *Conveyor
}

type ExportOptions struct {
	ExportImageNameList []string
	ExportTagFuncList   []image.ExportTagFunc
	MutateConfigFunc    func(config v1.Config) (v1.Config, error)
}

func NewExporter(c *Conveyor, opts ExportOptions) *Exporter {
	return &Exporter{
		Conveyor:      c,
		ExportOptions: opts,
	}
}

func (e *Exporter) Run(ctx context.Context) error {
	if len(e.ExportTagFuncList) == 0 {
		return nil
	}

	for _, desc := range e.Conveyor.imagesTree.GetImagesByName(true) {
		name, images := desc.Unpair()
		if !slices.Contains(e.ExportImageNameList, name) {
			continue
		}

		targetPlatforms := util.MapFuncToSlice(images, func(img *build_image.Image) string { return img.TargetPlatform })
		if len(targetPlatforms) == 1 {
			img := images[0]
			if err := e.exportImage(ctx, img); err != nil {
				return fmt.Errorf("unable to export image %q: %w", img.Name, err)
			}
		} else {
			// FIXME(multiarch): Support multiplatform manifest by pushing local images to repo first, then create manifest list.
			// FIXME(multiarch): Also support multiplatform manifest in werf build command in local mode with enabled final-repo.
			if _, isLocal := e.Conveyor.StorageManager.GetStagesStorage().(*storage.LocalStagesStorage); isLocal {
				return fmt.Errorf("export command is not supported in multiplatform mode")
			}

			// multiplatform mode
			img := e.Conveyor.imagesTree.GetMultiplatformImage(name)
			if err := e.exportMultiplatformImage(ctx, img); err != nil {
				return fmt.Errorf("unable to export multiplatform image %q: %w", img.Name, err)
			}
		}
	}

	return nil
}

func (e *Exporter) exportMultiplatformImage(ctx context.Context, img *build_image.MultiplatformImage) error {
	return logboek.Context(ctx).Default().LogProcess("Exporting image...").
		Options(func(options types.LogProcessOptionsInterface) {
			options.Style(style.Highlight())
		}).
		DoError(func() error {
			for _, tagFunc := range e.ExportTagFuncList {
				tag := tagFunc(img.Name, img.GetStageID().String())
				if err := logboek.Context(ctx).Default().LogProcess("tag %s", tag).
					DoError(func() error {
						desc := img.GetStageDescription()
						if err := e.Conveyor.StorageManager.GetStagesStorage().ExportStage(ctx, desc, tag, e.MutateConfigFunc); err != nil {
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

func (e *Exporter) exportImage(ctx context.Context, img *build_image.Image) error {
	if !slices.Contains(e.ExportImageNameList, img.Name) {
		return nil
	}
	if len(e.ExportTagFuncList) == 0 {
		return nil
	}

	return logboek.Context(ctx).Default().LogProcess("Exporting image...").
		Options(func(options types.LogProcessOptionsInterface) {
			options.Style(style.Highlight())
		}).
		DoError(func() error {
			for _, tagFunc := range e.ExportTagFuncList {
				tag := tagFunc(img.GetName(), img.GetStageID())
				if err := logboek.Context(ctx).Default().LogProcess("tag %s", tag).
					DoError(func() error {
						stageDesc := img.GetLastNonEmptyStage().GetStageImage().Image.GetStageDescription()
						if err := e.Conveyor.StorageManager.GetStagesStorage().ExportStage(ctx, stageDesc, tag, e.MutateConfigFunc); err != nil {
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
