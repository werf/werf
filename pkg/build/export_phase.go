package build

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"k8s.io/utils/strings/slices"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"
	build_image "github.com/werf/werf/v2/pkg/build/image"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/storage"
	"github.com/werf/werf/v2/pkg/util/parallel"
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

	images := e.Conveyor.imagesTree.GetImagesByName(true, build_image.WithExportImageNameList(e.ExportImageNameList))

	if err := parallel.DoTasks(ctx, len(images), parallel.DoTasksOptions{
		MaxNumberOfWorkers: int(e.Conveyor.ParallelTasksLimit),
	}, func(ctx context.Context, taskId int) error {
		pair := images[taskId]
		name, imagesToExport := pair.Unpair()

		targetPlatforms := util.MapFuncToSlice(imagesToExport, func(img *build_image.Image) string { return img.TargetPlatform })
		if len(targetPlatforms) == 1 {
			img := imagesToExport[0]
			if err := e.exportImage(ctx, img); err != nil {
				return fmt.Errorf("unable to export image %q: %w", img.Name, err)
			}
		} else {
			// FIXME(multiarch): Support multiplatform manifest by pushing local images to repo first, then create manifest list.
			// FIXME(multiarch): Also support multiplatform manifest in werf build command in local mode with enabled final-repo.
			if _, isLocal := e.Conveyor.StorageManager.GetStagesStorage().(*storage.LocalStagesStorage); isLocal {
				return fmt.Errorf("export command in multiplatform mode should be used with remote stages storage")
			}

			// multiplatform mode
			img := e.Conveyor.imagesTree.GetMultiplatformImage(name)
			if err := e.exportMultiplatformImage(ctx, img); err != nil {
				return fmt.Errorf("unable to export multiplatform image %q: %w", img.Name, err)
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("export failed: %w", err)
	}

	return nil
}

func (e *Exporter) RunFromReport(ctx context.Context, reportPath string) error {
	if len(e.ExportTagFuncList) == 0 {
		return nil
	}

	report, err := LoadBuildReportFromFile(reportPath)
	if err != nil {
		return fmt.Errorf("unable to load build report: %w", err)
	}

	imagesToExport := e.filterImagesFromReport(report)
	if len(imagesToExport) == 0 {
		return nil
	}

	if err := parallel.DoTasks(ctx, len(imagesToExport), parallel.DoTasksOptions{
		MaxNumberOfWorkers: int(e.Conveyor.ParallelTasksLimit),
	}, func(ctx context.Context, taskId int) error {
		imgRecord := imagesToExport[taskId]

		isMultiplatform := imgRecord.TargetPlatform == ""

		if isMultiplatform {
			if err := e.exportMultiplatformImageFromReport(ctx, imgRecord); err != nil {
				return fmt.Errorf("unable to export multiplatform image %q from report: %w", imgRecord.WerfImageName, err)
			}
		} else {
			if err := e.exportImageFromReport(ctx, imgRecord); err != nil {
				return fmt.Errorf("unable to export image %q from report: %w", imgRecord.WerfImageName, err)
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("export from report failed: %w", err)
	}

	return nil
}

func (e *Exporter) filterImagesFromReport(report *ImagesReport) []ReportImageRecord {
	var result []ReportImageRecord
	for imageName, record := range report.Images {
		if !record.Final {
			continue
		}
		if len(e.ExportImageNameList) > 0 && !slices.Contains(e.ExportImageNameList, imageName) {
			continue
		}
		result = append(result, record)
	}
	return result
}

func (e *Exporter) exportImageFromReport(ctx context.Context, record ReportImageRecord) error {
	if len(e.ExportTagFuncList) == 0 {
		return nil
	}

	return logboek.Context(ctx).Default().LogProcess(fmt.Sprintf("Exporting image %s (from report)", record.WerfImageName)).
		Options(func(options types.LogProcessOptionsInterface) {
			options.Style(style.Highlight())
		}).
		DoError(func() error {
			stageDesc := stageDescFromReportRecord(record)

			for _, tagFunc := range e.ExportTagFuncList {
				stageID := extractStageIDFromReport(record)
				tag := tagFunc(record.WerfImageName, stageID)
				if err := logboek.Context(ctx).Default().LogProcess("tag %s", tag).
					DoError(func() error {
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

func (e *Exporter) exportMultiplatformImageFromReport(ctx context.Context, record ReportImageRecord) error {
	if len(e.ExportTagFuncList) == 0 {
		return nil
	}

	return logboek.Context(ctx).Default().LogProcess(fmt.Sprintf("Exporting image %s (from report)", record.WerfImageName)).
		Options(func(options types.LogProcessOptionsInterface) {
			options.Style(style.Highlight())
		}).
		DoError(func() error {
			stageDesc := stageDescFromReportRecord(record)
			stageDesc.Info.IsIndex = true

			for _, tagFunc := range e.ExportTagFuncList {
				stageID := extractStageIDFromReport(record)
				tag := tagFunc(record.WerfImageName, stageID)
				if err := logboek.Context(ctx).Default().LogProcess("tag %s", tag).
					DoError(func() error {
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

func (e *Exporter) exportMultiplatformImage(ctx context.Context, img *build_image.MultiplatformImage) error {
	return logboek.Context(ctx).Default().LogProcess(fmt.Sprintf("Exporting image %s", img.Name)).
		Options(func(options types.LogProcessOptionsInterface) {
			options.Style(style.Highlight())
		}).
		DoError(func() error {
			for _, tagFunc := range e.ExportTagFuncList {
				tag := tagFunc(img.Name, img.GetStageID().String())
				if err := logboek.Context(ctx).Default().LogProcess("tag %s", tag).
					DoError(func() error {
						stageDesc := img.GetStageDesc()
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

func (e *Exporter) exportImage(ctx context.Context, img *build_image.Image) error {
	if !slices.Contains(e.ExportImageNameList, img.Name) {
		return nil
	}
	if len(e.ExportTagFuncList) == 0 {
		return nil
	}

	return logboek.Context(ctx).Default().LogProcess(fmt.Sprintf("Exporting image %s", img.Name)).
		Options(func(options types.LogProcessOptionsInterface) {
			options.Style(style.Highlight())
		}).
		DoError(func() error {
			for _, tagFunc := range e.ExportTagFuncList {
				tag := tagFunc(img.GetName(), img.GetStageID())
				if err := logboek.Context(ctx).Default().LogProcess("tag %s", tag).
					DoError(func() error {
						stageDesc := img.GetLastNonEmptyStage().GetStageImage().Image.GetStageDesc()
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

func stageDescFromReportRecord(record ReportImageRecord) *image.StageDesc {
	var lastStage ReportStageRecord
	if len(record.Stages) > 0 {
		lastStage = record.Stages[len(record.Stages)-1]
	}

	digest, creationTs := parseStageTag(record.DockerTag)

	return &image.StageDesc{
		StageID: &image.StageID{
			Digest:     digest,
			CreationTs: creationTs,
		},
		Info: &image.Info{
			Name:              record.DockerImageName,
			Repository:        record.DockerRepo,
			Tag:               record.DockerTag,
			RepoDigest:        fmt.Sprintf("%s@%s", record.DockerRepo, record.DockerImageDigest),
			ID:                record.DockerImageID,
			Size:              record.Size,
			CreatedAtUnixNano: lastStage.CreatedAt,
		},
	}
}

func parseStageTag(tag string) (digest string, creationTs int64) {
	parts := strings.SplitN(tag, "-", 2)
	if len(parts) == 1 {
		// Multiplatform tag: just digest (56 chars sha3-224)
		return parts[0], 0
	}

	// Regular tag: digest-creationTs
	digest = parts[0]
	if ts, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
		creationTs = ts
	}
	return digest, creationTs
}

func extractStageIDFromReport(record ReportImageRecord) string {
	digest, creationTs := parseStageTag(record.DockerTag)
	stageID := image.NewStageID(digest, creationTs)
	return stageID.String()
}
