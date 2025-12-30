package build

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/build/image"
	imagePkg "github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/storage"
)

const (
	ReportJSON    ReportFormat = "json"
	ReportEnvFile ReportFormat = "envfile"
)

const (
	BaseImageSourceTypeRepo      = "repo"
	BaseImageSourceTypeSecondary = "secondary"
)

type ReportFormat string

type ReportImageRecord struct {
	WerfImageName     string
	DockerRepo        string
	DockerTag         string
	DockerImageID     string
	DockerImageDigest string
	DockerImageName   string
	TargetPlatform    string
	Rebuilt           bool
	Final             bool
	Size              int64
	BuildTime         string
	Stages            []ReportStageRecord
}

type ReportStageRecord struct {
	Name              string
	DockerImageName   string
	DockerTag         string
	DockerImageID     string
	DockerImageDigest string
	CreatedAt         int64
	Size              int64
	SourceType        string
	BaseImagePulled   bool
	Rebuilt           bool
	BuildTime         string
}

type ImagesReport struct {
	mux              sync.Mutex
	Images           map[string]ReportImageRecord
	ImagesByPlatform map[string]map[string]ReportImageRecord
}

func NewImagesReport() *ImagesReport {
	return &ImagesReport{
		Images:           make(map[string]ReportImageRecord),
		ImagesByPlatform: make(map[string]map[string]ReportImageRecord),
	}
}

func (report *ImagesReport) SetImageRecord(name string, imageRecord ReportImageRecord) {
	report.mux.Lock()
	defer report.mux.Unlock()
	report.Images[name] = imageRecord
}

func (report *ImagesReport) SetImageByPlatformRecord(targetPlatform, name string, imageRecord ReportImageRecord) {
	report.mux.Lock()
	defer report.mux.Unlock()

	if _, hasKey := report.ImagesByPlatform[name]; !hasKey {
		report.ImagesByPlatform[name] = make(map[string]ReportImageRecord)
	}
	report.ImagesByPlatform[name][targetPlatform] = imageRecord
}

func (report *ImagesReport) ToJsonData() ([]byte, error) {
	report.mux.Lock()
	defer report.mux.Unlock()

	data, err := json.MarshalIndent(report, "", "\t")
	if err != nil {
		return nil, err
	}
	data = append(data, []byte("\n")...)

	return data, nil
}

func (report *ImagesReport) ToEnvFileData() []byte {
	report.mux.Lock()
	defer report.mux.Unlock()

	buf := bytes.NewBuffer([]byte{})
	for img, record := range report.Images {
		buf.WriteString(GenerateImageEnv(img, record.DockerImageName))
		buf.WriteString("\n")
	}

	return buf.Bytes()
}

func createBuildReport(ctx context.Context, phase *BuildPhase, imagePairs []util.Pair[string, []*image.Image]) error {
	for _, desc := range imagePairs {
		name, images := desc.Unpair()
		targetPlatforms := util.MapFuncToSlice(images, func(img *image.Image) string { return img.TargetPlatform })

		for _, img := range images {
			stageImage := img.GetLastNonEmptyStage().GetStageImage().Image
			stageDesc := stageImage.GetFinalStageDesc()
			if stageDesc == nil {
				stageDesc = stageImage.GetStageDesc()
			}

			stages := getStagesReport(img, false)

			record := ReportImageRecord{
				WerfImageName:     img.GetName(),
				DockerRepo:        stageDesc.Info.Repository,
				DockerTag:         stageDesc.Info.Tag,
				DockerImageID:     stageDesc.Info.ID,
				DockerImageDigest: stageDesc.Info.GetDigest(),
				DockerImageName:   stageDesc.Info.Name,
				TargetPlatform:    img.TargetPlatform,
				Rebuilt:           img.GetRebuilt(),
				Final:             img.IsFinal,
				Size:              stageDesc.Info.Size,
				BuildTime:         fmt.Sprintf("%.2f", img.BuildDuration.Seconds()),
				Stages:            stages,
			}

			if os.Getenv("WERF_ENABLE_REPORT_BY_PLATFORM") == "1" {
				phase.ImagesReport.SetImageByPlatformRecord(img.TargetPlatform, img.GetName(), record)
			}
			if len(targetPlatforms) == 1 {
				phase.ImagesReport.SetImageRecord(img.Name, record)
			}
		}

		if _, isLocal := phase.Conveyor.StorageManager.GetStagesStorage().(*storage.LocalStagesStorage); !isLocal {
			if len(targetPlatforms) > 1 {
				img := phase.Conveyor.imagesTree.GetMultiplatformImage(name)

				isRebuilt := false
				for _, pImg := range img.Images {
					isRebuilt = (isRebuilt || pImg.GetRebuilt())
				}

				stageDesc := img.GetFinalStageDesc()
				if stageDesc == nil {
					stageDesc = img.GetStageDesc()
				}

				buildDuration := 0.0
				stages := []ReportStageRecord{}
				for _, pImg := range img.Images {
					for _, stage := range getStagesReport(pImg, true) {
						stages = append(stages, stage)
					}
					buildDuration += pImg.BuildDuration.Seconds()
				}

				record := ReportImageRecord{
					WerfImageName:     img.Name,
					DockerRepo:        stageDesc.Info.Repository,
					DockerTag:         stageDesc.Info.Tag,
					DockerImageID:     stageDesc.Info.ID,
					DockerImageDigest: stageDesc.Info.GetDigest(),
					DockerImageName:   stageDesc.Info.Name,
					TargetPlatform:    "", // multiplatform manifest list
					Rebuilt:           isRebuilt,
					Final:             img.IsFinal,
					Size:              stageDesc.Info.Size,
					BuildTime:         fmt.Sprintf("%.2f", buildDuration),
					Stages:            stages,
				}
				phase.ImagesReport.SetImageRecord(img.Name, record)
			}
		}
	}

	debugJsonData, err := phase.ImagesReport.ToJsonData()
	logboek.Context(ctx).Debug().LogF("ImagesReport: (err: %v)\n%s", err, debugJsonData)

	if phase.ReportPath != "" {
		var data []byte
		var err error
		switch phase.ReportFormat {
		case ReportJSON:
			if data, err = phase.ImagesReport.ToJsonData(); err != nil {
				return fmt.Errorf("unable to prepare report json: %w", err)
			}
			logboek.Context(ctx).Debug().LogF("Writing json report to the %q:\n%s", phase.ReportPath, data)
		case ReportEnvFile:
			data = phase.ImagesReport.ToEnvFileData()
			logboek.Context(ctx).Debug().LogF("Writing envfile report to the %q:\n%s", phase.ReportPath, data)
		default:
			panic(fmt.Sprintf("unknown report format %q", phase.ReportFormat))
		}

		if err := os.WriteFile(phase.ReportPath, data, 0o644); err != nil {
			return fmt.Errorf("unable to write report to %s: %w", phase.ReportPath, err)
		}
	}

	return nil
}

func setBuildTime(b bool, t string) string {
	if !b {
		return "0.00"
	}
	return t
}

func getStagesReport(img *image.Image, multiplatform bool) []ReportStageRecord {
	var stagesRecords []ReportStageRecord
	for _, stg := range img.GetStages() {
		stgImg := stg.GetStageImage()
		if stgImg == nil || stgImg.Image == nil || stgImg.Image.GetStageDesc() == nil {
			continue
		}
		stgMeta := stg.GetMeta()
		stgDesc := stgImg.Image.GetStageDesc()
		name := string(stg.Name())
		if multiplatform {
			name = fmt.Sprintf("%s (%s)", name, img.TargetPlatform)
		}
		record := ReportStageRecord{
			Name:              name,
			DockerImageName:   stgDesc.Info.Name,
			DockerTag:         stgDesc.Info.Tag,
			DockerImageID:     stgDesc.Info.ID,
			DockerImageDigest: stgDesc.Info.GetDigest(),
			CreatedAt:         stgDesc.Info.CreatedAtUnixNano,
			Size:              stgDesc.Info.Size,
			SourceType:        stgMeta.BaseImageSourceType,
			BaseImagePulled:   stgMeta.BaseImagePulled,
			Rebuilt:           stgMeta.Rebuilt,
			BuildTime:         setBuildTime(stgMeta.Rebuilt, stgMeta.BuildTime),
		}
		stagesRecords = append(stagesRecords, record)
	}
	return stagesRecords
}

func LoadBuildReportFromFile(path string) (*ImagesReport, error) {
	file, err := os.Open(path) // opens it for reading only
	if err != nil {
		return nil, fmt.Errorf("unable to read build report file %q: %w", path, err)
	}
	defer file.Close()

	report, err := parseBuildReport(file)
	if err != nil {
		return nil, fmt.Errorf("unable to parse build report file %q: %w", path, err)
	}

	if err := validateBuildReport(report); err != nil {
		return nil, fmt.Errorf("invalid build report file %q: %w", path, err)
	}

	return report, nil
}

func parseBuildReport(reader io.Reader) (*ImagesReport, error) {
	decoder := json.NewDecoder(reader)

	var report ImagesReport
	if err := decoder.Decode(&report); err != nil {
		return nil, fmt.Errorf("unable to decode build report: %w", err)
	}

	return &report, nil
}

func validateBuildReport(report *ImagesReport) error {
	if len(report.Images) == 0 {
		return fmt.Errorf("build report contains no images")
	}

	for imageName, record := range report.Images {
		if err := validateImageRecord(imageName, record); err != nil {
			return err
		}
	}

	for imageName, platformRecords := range report.ImagesByPlatform {
		for platform, record := range platformRecords {
			if err := validateImageRecord(fmt.Sprintf("%s[%s]", imageName, platform), record); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateImageRecord(imageName string, record ReportImageRecord) error {
	if record.WerfImageName == "" {
		return fmt.Errorf("image %q has empty WerfImageName", imageName)
	}
	if record.DockerImageName == "" {
		return fmt.Errorf("image %q has empty DockerImageName", imageName)
	}
	if record.DockerRepo == "" {
		return fmt.Errorf("image %q has empty DockerRepo", imageName)
	}
	if record.DockerTag == "" {
		return fmt.Errorf("image %q has empty DockerTag", imageName)
	}
	if record.DockerImageID == "" {
		return fmt.Errorf("image %q has empty DockerImageID", imageName)
	}
	if record.DockerImageDigest == "" {
		return fmt.Errorf("image %q has empty DockerImageDigest", imageName)
	}

	for i, stage := range record.Stages {
		if err := validateStageRecord(imageName, i, stage); err != nil {
			return err
		}
	}

	return nil
}

func validateStageRecord(imageName string, stageIndex int, stage ReportStageRecord) error {
	stageRef := fmt.Sprintf("image %q stage #%d", imageName, stageIndex)

	if stage.Name == "" {
		return fmt.Errorf("%s has empty Name", stageRef)
	}

	stageRef = fmt.Sprintf("image %q stage %q", imageName, stage.Name)

	if stage.DockerImageName == "" {
		return fmt.Errorf("%s has empty DockerImageName", stageRef)
	}
	if stage.DockerImageID == "" {
		return fmt.Errorf("%s has empty DockerImageID", stageRef)
	}
	if stage.DockerImageDigest == "" {
		return fmt.Errorf("%s has empty DockerImageDigest", stageRef)
	}

	return nil
}

func (report *ImagesReport) ToImageInfoGetters(opts imagePkg.InfoGetterOptions) []*imagePkg.InfoGetter {
	report.mux.Lock()
	defer report.mux.Unlock()

	var infoGetters []*imagePkg.InfoGetter
	for _, record := range report.Images {
		if opts.OnlyFinal && !record.Final {
			continue
		}

		infoGetters = append(infoGetters, imagePkg.NewInfoGetter(record.WerfImageName, record.DockerImageName, opts))
	}

	return infoGetters
}
