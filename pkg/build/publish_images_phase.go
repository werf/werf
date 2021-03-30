package build

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"

	"github.com/werf/werf/pkg/build/stage"
	"github.com/werf/werf/pkg/container_runtime"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/tag_strategy"
	"github.com/werf/werf/pkg/util"
)

func NewPublishImagesPhase(c *Conveyor, imagesRepo storage.ImagesRepo, opts PublishImagesOptions) *PublishImagesPhase {
	tagsByScheme := map[tag_strategy.TagStrategy][]string{
		tag_strategy.Custom:    opts.CustomTags,
		tag_strategy.GitBranch: opts.TagsByGitBranch,
		tag_strategy.GitTag:    opts.TagsByGitTag,
		tag_strategy.GitCommit: opts.TagsByGitCommit,
	}
	return &PublishImagesPhase{
		BasePhase:            BasePhase{c},
		ImagesToPublish:      opts.ImagesToPublish,
		TagsByScheme:         tagsByScheme,
		TagByStagesSignature: opts.TagByStagesSignature,
		ImagesRepo:           imagesRepo,
		PublishReport:        &PublishReport{Images: make(map[string]PublishReportImageRecord)},
		PublishReportPath:    opts.PublishReportPath,
		PublishReportFormat:  opts.PublishReportFormat,
	}
}

type PublishImagesPhase struct {
	BasePhase
	ImagesToPublish      []string
	TagsByScheme         map[tag_strategy.TagStrategy][]string
	TagByStagesSignature bool
	ImagesRepo           storage.ImagesRepo

	PublishReport       *PublishReport
	PublishReportPath   string
	PublishReportFormat PublishReportFormat
}

type PublishReportFormat string

const (
	PublishReportJSON PublishReportFormat = "json"
)

type PublishReport struct {
	mux    sync.Mutex
	Images map[string]PublishReportImageRecord
}

func (report *PublishReport) SetImageRecord(name string, imageRecord PublishReportImageRecord) {
	report.mux.Lock()
	defer report.mux.Unlock()
	report.Images[name] = imageRecord
}

func (report *PublishReport) ToJson() ([]byte, error) {
	report.mux.Lock()
	defer report.mux.Unlock()
	return json.Marshal(report)
}

type PublishReportImageRecord struct {
	WerfImageName string
	DockerRepo    string
	DockerTag     string
	DockerImageID string
}

func (phase *PublishImagesPhase) Name() string {
	return "publish"
}

func (phase *PublishImagesPhase) BeforeImages(ctx context.Context) error {
	return nil
}

func (phase *PublishImagesPhase) AfterImages(ctx context.Context) error {
	if data, err := phase.PublishReport.ToJson(); err != nil {
		return fmt.Errorf("unable to prepare publish report json: %s", err)
	} else {
		logboek.Context(ctx).Debug().LogF("Publish report:\n%s\n", data)

		if phase.PublishReportPath != "" && phase.PublishReportFormat == PublishReportJSON {
			if err := ioutil.WriteFile(phase.PublishReportPath, append(data, []byte("\n")...), 0644); err != nil {
				return fmt.Errorf("unable to write publish report to %s: %s", phase.PublishReportPath, err)
			}
		}
	}

	return nil
}

func (phase *PublishImagesPhase) BeforeImageStages(ctx context.Context, img *Image) error {
	return nil
}

func (phase *PublishImagesPhase) OnImageStage(ctx context.Context, img *Image, stg stage.Interface) error {
	return nil
}

func (phase *PublishImagesPhase) AfterImageStages(ctx context.Context, img *Image) error {
	if img.isArtifact {
		return nil
	}

	if len(phase.ImagesToPublish) == 0 {
		return phase.publishImage(ctx, img)
	}

	for _, name := range phase.ImagesToPublish {
		if name == img.GetName() {
			return phase.publishImage(ctx, img)
		}
	}

	return nil
}

func (phase *PublishImagesPhase) ImageProcessingShouldBeStopped(_ context.Context, _ *Image) bool {
	return false
}

func (phase *PublishImagesPhase) publishImage(ctx context.Context, img *Image) error {
	var nonEmptySchemeInOrder []tag_strategy.TagStrategy
	for strategy, tags := range phase.TagsByScheme {
		if len(tags) == 0 {
			continue
		}

		nonEmptySchemeInOrder = append(nonEmptySchemeInOrder, strategy)
	}

	localGitRepo := phase.Conveyor.GetLocalGitRepo()
	if localGitRepo != nil {
		if err := logboek.Context(ctx).Info().LogProcess(fmt.Sprintf("Processing image %s git metadata", img.GetName())).
			DoError(func() error {
				headCommit, err := localGitRepo.HeadCommit(ctx)
				if err != nil {
					return err
				}

				if metadata, err := phase.Conveyor.StorageManager.StagesStorage.GetImageMetadataByCommit(ctx, phase.Conveyor.projectName(), img.GetName(), headCommit); err != nil {
					return fmt.Errorf("unable to get image %s metadata by commit %s: %s", img.GetName(), headCommit, err)
				} else if metadata != nil {
					if metadata.ContentSignature != img.GetContentSignature() {
						// TODO: Check image existance and automatically allow republish if no images found by this commit. What if multiple images are published by multiple tagging strategies (including custom)?
						// TODO: allowInconsistentPublish: true option for werf.yaml
						// FIXME: return fmt.Errorf("inconsistent build: found already published image with stages-signature %s by commit %s, cannot publish a new image with stages-signature %s by the same commit", metadata.ContentSignature, headCommit, img.GetContentSignature())
						return phase.Conveyor.StorageManager.StagesStorage.PutImageCommit(ctx, phase.Conveyor.projectName(), img.GetName(), headCommit, &storage.ImageMetadata{ContentSignature: img.GetContentSignature()})
					}
					return nil
				} else {
					return phase.Conveyor.StorageManager.StagesStorage.PutImageCommit(ctx, phase.Conveyor.projectName(), img.GetName(), headCommit, &storage.ImageMetadata{ContentSignature: img.GetContentSignature()})
				}
			}); err != nil {
			return err
		}
	}

	var existingTags []string
	if tags, err := phase.fetchExistingTags(ctx, img.GetName()); err != nil {
		return err
	} else {
		existingTags = tags
	}

	for _, strategy := range nonEmptySchemeInOrder {
		imageMetaTags := phase.TagsByScheme[strategy]

		if err := logboek.Context(ctx).Info().LogProcess("%s tagging strategy", string(strategy)).
			Options(func(options types.LogProcessOptionsInterface) {
				options.Style(style.Highlight())
			}).
			DoError(func() error {
				for _, imageMetaTag := range imageMetaTags {
					if err := phase.publishImageByTag(ctx, img, imageMetaTag, strategy, publishImageByTagOptions{ExistingTagsList: existingTags, CheckAlreadyExistingTagByContentSignatureLabel: true}); err != nil {
						return fmt.Errorf("error publishing image %s by tag %s: %s", img.LogName(), imageMetaTag, err)
					}
				}

				return nil
			}); err != nil {
			return err
		}
	}

	if phase.TagByStagesSignature {
		if err := logboek.Context(ctx).Info().LogProcess("%s tagging strategy", tag_strategy.StagesSignature).
			Options(func(options types.LogProcessOptionsInterface) {
				options.Style(style.Highlight())
			}).
			DoError(func() error {
				if err := phase.publishImageByTag(ctx, img, img.GetContentSignature(), tag_strategy.StagesSignature, publishImageByTagOptions{ExistingTagsList: existingTags}); err != nil {
					return fmt.Errorf("error publishing image %s by image signature %s: %s", img.GetName(), img.GetContentSignature(), err)
				}

				return nil
			}); err != nil {
			return err
		}
	}

	return nil
}

func (phase *PublishImagesPhase) fetchExistingTags(ctx context.Context, imageName string) (existingTags []string, err error) {
	logProcessMsg := fmt.Sprintf("Fetching existing repo tags")
	logboek.Context(ctx).Info().LogProcessInline(logProcessMsg).Do(func() {
		existingTags, err = phase.ImagesRepo.GetAllImageRepoTags(ctx, imageName)
	})
	logboek.Context(ctx).Info().LogOptionalLn()

	if err != nil {
		return existingTags, fmt.Errorf("error fetching existing tags from image repository %s: %s", phase.ImagesRepo.String(), err)
	}
	return existingTags, nil
}

type publishImageByTagOptions struct {
	CheckAlreadyExistingTagByContentSignatureLabel bool
	ExistingTagsList                               []string
}

func (phase *PublishImagesPhase) publishImageByTag(ctx context.Context, img *Image, imageMetaTag string, tagStrategy tag_strategy.TagStrategy, opts publishImageByTagOptions) error {
	imageRepository := phase.ImagesRepo.ImageRepositoryName(img.GetName())
	imageName := phase.ImagesRepo.ImageRepositoryNameWithTag(img.GetName(), imageMetaTag)
	imageActualTag := phase.ImagesRepo.ImageRepositoryTag(img.GetName(), imageMetaTag)

	alreadyExists, alreadyExistingDockerImageID, err := phase.checkImageAlreadyExists(ctx, opts.ExistingTagsList, img.GetName(), imageMetaTag, img.GetContentSignature(), opts.CheckAlreadyExistingTagByContentSignatureLabel)
	if err != nil {
		return fmt.Errorf("error checking image %s already exists in the images repo: %s", img.LogName(), err)
	}

	if alreadyExists {
		logboek.Context(ctx).Default().LogFHighlight("%s tag %s is up-to-date\n", string(tagStrategy), imageActualTag)

		logboek.Context(ctx).Streams().DoWithIndent(func() {
			logboek.Context(ctx).Default().LogFDetails("images-repo: %s\n", imageRepository)
			logboek.Context(ctx).Default().LogFDetails("      image: %s\n", imageName)
		})

		logboek.Context(ctx).LogOptionalLn()

		phase.PublishReport.SetImageRecord(img.GetName(), PublishReportImageRecord{
			WerfImageName: img.GetName(),
			DockerRepo:    imageRepository,
			DockerTag:     imageActualTag,
			DockerImageID: alreadyExistingDockerImageID,
		})

		return nil
	}

	publishImage := container_runtime.NewWerfImage(phase.Conveyor.GetStageImage(img.GetLastNonEmptyStage().GetImage().Name()), imageName, phase.Conveyor.ContainerRuntime.(*container_runtime.LocalDockerServerRuntime))

	publishImage.Container().ServiceCommitChangeOptions().AddLabel(map[string]string{
		image.WerfDockerImageName:       imageName,
		image.WerfTagStrategyLabel:      string(tagStrategy),
		image.WerfImageLabel:            "true",
		image.WerfImageNameLabel:        img.GetName(),
		image.WerfImageTagLabel:         imageMetaTag,
		image.WerfContentSignatureLabel: img.GetContentSignature(),
		image.WerfImageVersionLabel:     image.WerfImageVersion,
	})

	successInfoSectionFunc := func() {
		logboek.Context(ctx).Streams().DoWithIndent(func() {
			logboek.Context(ctx).Default().LogFDetails("images-repo: %s\n", imageRepository)
			logboek.Context(ctx).Default().LogFDetails("      image: %s\n", imageName)
		})
	}

	publishingFunc := func() error {
		if err := phase.Conveyor.StorageManager.FetchStage(ctx, img.GetLastNonEmptyStage()); err != nil {
			return fmt.Errorf("unable to fetch last non empty stage %q: %s", img.GetLastNonEmptyStage().GetImage().Name(), err)
		}

		if err := logboek.Context(ctx).Info().LogProcess("Building final image with meta information").DoError(func() error {
			if err := publishImage.Build(ctx, container_runtime.BuildOptions{}); err != nil {
				return fmt.Errorf("error building %s with tagging strategy '%s': %s", imageName, tagStrategy, err)
			}
			return nil
		}); err != nil {
			return err
		}

		if lock, err := phase.Conveyor.StorageLockManager.LockImage(ctx, phase.Conveyor.projectName(), imageName); err != nil {
			return fmt.Errorf("error locking image %s: %s", imageName, err)
		} else {
			defer phase.Conveyor.StorageLockManager.Unlock(ctx, lock)
		}

		existingTags, err := phase.fetchExistingTags(ctx, img.GetName())
		if err != nil {
			return err
		}

		alreadyExists, alreadyExistingImageID, err := phase.checkImageAlreadyExists(ctx, existingTags, img.GetName(), imageMetaTag, img.GetContentSignature(), opts.CheckAlreadyExistingTagByContentSignatureLabel)
		if err != nil {
			return fmt.Errorf("error checking image %s already exists in the images repo: %s", img.LogName(), err)
		}

		if alreadyExists {
			logboek.Context(ctx).Default().LogFHighlight("%s tag %s is up-to-date\n", string(tagStrategy), imageActualTag)
			logboek.Context(ctx).Streams().DoWithIndent(func() {
				logboek.Context(ctx).Info().LogFDetails("discarding newly built image %s\n", publishImage.MustGetBuiltId())
				logboek.Context(ctx).Default().LogFDetails("images-repo: %s\n", imageRepository)
				logboek.Context(ctx).Default().LogFDetails("      image: %s\n", imageName)
			})

			logboek.Context(ctx).LogOptionalLn()

			phase.PublishReport.SetImageRecord(img.GetName(), PublishReportImageRecord{
				WerfImageName: img.GetName(),
				DockerRepo:    imageRepository,
				DockerTag:     imageActualTag,
				DockerImageID: alreadyExistingImageID,
			})

			return nil
		}

		if err := phase.ImagesRepo.PublishImage(ctx, publishImage); err != nil {
			return err
		}

		phase.PublishReport.SetImageRecord(img.GetName(), PublishReportImageRecord{
			WerfImageName: img.GetName(),
			DockerRepo:    imageRepository,
			DockerTag:     imageActualTag,
			DockerImageID: publishImage.MustGetBuiltId(),
		})

		return nil
	}

	return logboek.Context(ctx).Default().LogProcess("Publishing image %s by %s tag %s", img.LogName(), tagStrategy, imageMetaTag).
		Options(func(options types.LogProcessOptionsInterface) {
			options.SuccessInfoSectionFunc(successInfoSectionFunc)
			options.Style(style.Highlight())
		}).
		DoError(publishingFunc)
}

func (phase *PublishImagesPhase) checkImageAlreadyExists(ctx context.Context, existingTags []string, werfImageName, imageMetaTag, imageContentSignature string, checkAlreadyExistingTagByContentSignatureFromLabels bool) (bool, string, error) {
	imageActualTag := phase.ImagesRepo.ImageRepositoryTag(werfImageName, imageMetaTag)

	if !util.IsStringsContainValue(existingTags, imageActualTag) {
		return false, "", nil
	} else if !checkAlreadyExistingTagByContentSignatureFromLabels {
		return true, "", nil
	}

	var repoImageContentSignature string
	var repoDockerImageID string
	var err error
	getImageContentSignature := func() error {
		repoImage, err := phase.ImagesRepo.GetRepoImage(ctx, werfImageName, imageMetaTag)
		if err != nil {
			return err
		}
		repoImageContentSignature = repoImage.Labels[image.WerfContentSignatureLabel]
		repoDockerImageID = repoImage.ID
		return nil
	}

	err = logboek.Context(ctx).Info().LogProcessInline("Getting existing tag %s manifest", imageActualTag).DoError(getImageContentSignature)
	if err != nil {
		return false, "", fmt.Errorf("unable to get image %s parent id: %s", werfImageName, err)
	}

	logboek.Context(ctx).Debug().LogF("Current image content signature: %s\n", imageContentSignature)
	logboek.Context(ctx).Debug().LogF("Already published image content signature: %s\n", repoImageContentSignature)
	logboek.Context(ctx).Debug().LogF("Already published image docker ID: %s\n", repoDockerImageID)

	return imageContentSignature == repoImageContentSignature, repoDockerImageID, nil
}

func (phase *PublishImagesPhase) Clone() Phase {
	u := *phase
	return &u
}
