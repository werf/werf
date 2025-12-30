package build

import (
	"context"
	"fmt"
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

func (e *Exporter) getMaxNumberOfWorkers() int {
	if !e.Conveyor.Parallel {
		return 1
	}
	return int(e.Conveyor.ParallelTasksLimit)
}

func (e *Exporter) Run(ctx context.Context) error {
	if len(e.ExportTagFuncList) == 0 {
		return nil
	}

	images := e.Conveyor.imagesTree.GetImagesByName(true, build_image.WithExportImageNameList(e.ExportImageNameList))

	if err := parallel.DoTasks(ctx, len(images), parallel.DoTasksOptions{
		MaxNumberOfWorkers: e.getMaxNumberOfWorkers(),
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
		MaxNumberOfWorkers: e.getMaxNumberOfWorkers(),
	}, func(ctx context.Context, taskId int) error {
		imgRecord := imagesToExport[taskId]
		isMultiplatform := imgRecord.TargetPlatform == ""

		if err := e.exportImageFromReportRecord(ctx, imgRecord, isMultiplatform); err != nil {
			return fmt.Errorf("unable to export image %q from report: %w", imgRecord.WerfImageName, err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("export from report failed: %w", err)
	}

	return nil
}

func (e *Exporter) filterImagesFromReport(report *ImagesReport) []ReportImageRecord {
	var result []ReportImageRecord

	if len(e.ExportImageNameList) == 0 {
		return result
	}

	for imageName, record := range report.Images {
		if !record.Final {
			continue
		}
		if !slices.Contains(e.ExportImageNameList, imageName) {
			continue
		}
		result = append(result, record)
	}
	return result
}

func (e *Exporter) exportImageFromReportRecord(ctx context.Context, record ReportImageRecord, isMultiplatform bool) error {
	if len(e.ExportTagFuncList) == 0 {
		return nil
	}

	return logboek.Context(ctx).Default().LogProcess(fmt.Sprintf("Exporting image %s (from report)", record.WerfImageName)).
		Options(func(options types.LogProcessOptionsInterface) {
			options.Style(style.Highlight())
		}).
		DoError(func() error {
			stageDesc, err := stageDescFromReportRecord(record)
			if err != nil {
				return fmt.Errorf("unable to get stage desc from report record: %w", err)
			}

			if isMultiplatform {
				stageDesc.Info.IsIndex = true
			}

			for _, tagFunc := range e.ExportTagFuncList {
				stageID, err := extractStageIDFromReport(record)
				if err != nil {
					return fmt.Errorf("unable to extract stage id from report record: %w", err)
				}

				tag := tagFunc(record.WerfImageName, stageID)
				if err := logboek.Context(ctx).Default().LogProcess("tag %s", tag).
					DoError(func() error {
						if err := e.Conveyor.StorageManager.GetStagesStorage().ExportStage(ctx, stageDesc, tag, e.MutateConfigFunc); err != nil {
							return fmt.Errorf("unable to export stage %s: %w", stageDesc.StageID.String(), err)
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
							return fmt.Errorf("unable to export stage %s: %w", stageDesc.StageID.String(), err)
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

func stageDescFromReportRecord(record ReportImageRecord) (*image.StageDesc, error) {
	var createdAtUnixNano int64
	if len(record.Stages) > 0 {
		createdAtUnixNano = record.Stages[len(record.Stages)-1].CreatedAt
	}

	digest, creationTs, err := image.GetDigestAndCreationTsFromLocalStageImageTag(record.DockerTag)
	if err != nil {
		return nil, fmt.Errorf("unable to parse stage tag %q: %w", record.DockerTag, err)
	}

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
			CreatedAtUnixNano: createdAtUnixNano,
		},
	}, nil
}

func extractStageIDFromReport(record ReportImageRecord) (string, error) {
	parts := strings.SplitN(record.DockerTag, "-", 2)
	if len(parts) == 1 {
		// Multiplatform tag: just digest (56 chars sha3-224)
		return parts[0], nil
	}

	digest, creationTs, err := image.GetDigestAndCreationTsFromLocalStageImageTag(record.DockerTag)
	if err != nil {
		return "", fmt.Errorf("unable to parse stage tag %q: %w", record.DockerTag, err)
	}

	stageID := image.NewStageID(digest, creationTs)

	return stageID.String(), nil
}
