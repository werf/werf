package build

import (
	"fmt"
	"strings"

	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/build/stage"
	"github.com/flant/werf/pkg/container_runtime"
	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/storage"
	"github.com/flant/werf/pkg/tag_strategy"
	"github.com/flant/werf/pkg/util"
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
	}
}

type PublishImagesPhase struct {
	BasePhase
	ImagesToPublish      []string
	TagsByScheme         map[tag_strategy.TagStrategy][]string
	TagByStagesSignature bool
	ImagesRepo           storage.ImagesRepo
}

func (phase *PublishImagesPhase) Name() string {
	return "publish"
}

func (phase *PublishImagesPhase) BeforeImages() error {
	return nil
}

func (phase *PublishImagesPhase) AfterImages() error {
	return nil
}

func (phase *PublishImagesPhase) BeforeImageStages(img *Image) error {
	return nil
}

func (phase *PublishImagesPhase) OnImageStage(img *Image, stg stage.Interface) error {
	return nil
}

func (phase *PublishImagesPhase) AfterImageStages(img *Image) error {
	if img.isArtifact {
		return nil
	}

	if len(phase.ImagesToPublish) == 0 {
		return phase.publishImage(img)
	}

	for _, name := range phase.ImagesToPublish {
		if name == img.GetName() {
			return phase.publishImage(img)
		}
	}

	return nil
}

func (phase *PublishImagesPhase) ImageProcessingShouldBeStopped(img *Image) bool {
	return false
}

/*
func (p *PublishImagesPhase) Run(c *Conveyor) error {
	logProcessOptions := logboek.LogProcessOptions{ColorizeMsgFunc: logboek.ColorizeHighlight}
	return logboek.LogProcess("Publishing images", logProcessOptions, func() error {
		return p.run(c)
	})
}

func (p *PublishImagesPhase) run(c *Conveyor) error {
	// TODO: Push stages should occur on the BuildStagesPhase

	for _, image := range imagesToPublish {
		if image.isArtifact { // FIXME: distributed stages
			continue
		}

		if err := logboek.LogProcess(image.LogDetailedName(), logboek.LogProcessOptions{ColorizeMsgFunc: image.LogProcessColorizeFunc()}, func() error {
			//if p.WithStages {
			//	err := logboek.LogProcess("Pushing stages cache", logboek.LogProcessOptions{}, func() error {
			//		if err := p.publishImageStages(c, image); err != nil {
			//			return fmt.Errorf("unable to push image %s stages: %s", image.GetName(), err)
			//		}
			//
			//		return nil
			//	})
			//
			//	if err != nil {
			//		return err
			//	}
			//}

			if !image.isArtifact {
				if err := p.publishImage(c, image); err != nil {
					return fmt.Errorf("unable to push image %s: %s", image.LogName(), err)
				}
			}

			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

//func (p *PublishImagesPhase) publishImageStages(c *Conveyor, image *Image) error {
//	stages := image.GetStages()
//
//	existingTags, err := docker_registry.Tags(p.ImagesRepo)
//	if err != nil {
//		return fmt.Errorf("error fetching existing stages cache list %s: %s", p.ImagesRepo, err)
//	}
//
//	for _, stage := range stages {
//		stageTagName := fmt.Sprintf(RepoImageStageTagFormat, stage.GetSignature())
//		stageImageName := fmt.Sprintf("%s:%s", p.ImagesRepo, stageTagName)
//
//		if util.IsStringsContainValue(existingTags, stageTagName) {
//			logboek.LogHighlightLn(stage.Name())
//
//			logboek.LogInfoF("stages-repo: %s\n", p.ImagesRepo)
//			logboek.LogInfoF("      image: %s\n", stageImageName)
//
//			logboek.LogOptionalLn()
//
//			continue
//		}
//
//		err := func() error {
//			imageLockName := image.ImageLockName(stageImageName)
//			if err := shluz.Lock(imageLockName, shluz.LockOptions{}); err != nil {
//				return fmt.Errorf("failed to lock %s: %s", imageLockName, err)
//			}
//
//			defer shluz.Unlock(imageLockName)
//
//			stageImage := c.GetStageImage(stage.GetImage().Name())
//
//			successInfoSectionFunc := func() {
//				_ = logboek.WithIndent(func() error {
//					logboek.LogInfoF("stages-repo: %s\n", p.ImagesRepo)
//					logboek.LogInfoF("      image: %s\n", stageImageName)
//
//					return nil
//				})
//			}
//
//			logProcessOptions := logboek.LogProcessOptions{SuccessInfoSectionFunc: successInfoSectionFunc, ColorizeMsgFunc: logboek.ColorizeHighlight}
//			return logboek.LogProcess(fmt.Sprintf("Publishing %s", stage.Name()), logProcessOptions, func() error {
//				if err := stageImage.Export(stageImageName); err != nil {
//					return fmt.Errorf("error pushing %s: %s", stageImageName, err)
//				}
//
//				return nil
//			})
//		}()
//
//		if err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
*/

func (phase *PublishImagesPhase) publishImage(img *Image) error {
	existingTags, err := phase.fetchExistingTags(img.GetName())
	if err != nil {
		return err
	}

	var nonEmptySchemeInOrder []tag_strategy.TagStrategy
	for strategy, tags := range phase.TagsByScheme {
		if len(tags) == 0 {
			continue
		}

		nonEmptySchemeInOrder = append(nonEmptySchemeInOrder, strategy)
	}

	for _, strategy := range nonEmptySchemeInOrder {
		imageMetaTags := phase.TagsByScheme[strategy]

		if err := logboek.Info.LogProcess(
			fmt.Sprintf("%s tagging strategy", string(strategy)),
			logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()},
			func() error {
				for _, imageMetaTag := range imageMetaTags {
					if err := phase.publishImageByTag(img, imageMetaTag, strategy, existingTags); err != nil {
						return fmt.Errorf("error publishing image %s by tag %s: %s", img.LogName(), imageMetaTag, err)
					}
				}

				return nil
			},
		); err != nil {
			return err
		}
	}

	if phase.TagByStagesSignature {
		if err := logboek.Info.LogProcess(
			fmt.Sprintf("%s tagging strategy", tag_strategy.StagesSignature),
			logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()},
			func() error {

				if err := phase.publishImageByTag(img, img.GetStagesSignature(), tag_strategy.StagesSignature, existingTags); err != nil {
					return fmt.Errorf("error publishing image %s by image signature %s: %s", img.GetName(), img.GetStagesSignature(), err)
				}

				return nil
			},
		); err != nil {
			return err
		}
	}

	return nil
}

func (phase *PublishImagesPhase) fetchExistingTags(imageName string) (existingTags []string, err error) {
	logProcessMsg := fmt.Sprintf("Fetching existing repo tags")
	_ = logboek.Info.LogProcessInline(logProcessMsg, logboek.LevelLogProcessInlineOptions{}, func() error {
		existingTags, err = phase.ImagesRepo.GetAllImageRepoTags(imageName)
		return nil
	})
	logboek.Info.LogOptionalLn()

	if err != nil {
		return existingTags, fmt.Errorf("error fetching existing tags from image repository %s: %s", phase.ImagesRepo.String(), err)
	}
	return existingTags, nil
}

func (phase *PublishImagesPhase) publishImageByTag(img *Image, imageMetaTag string, tagStrategy tag_strategy.TagStrategy, initialExistingTagsList []string) error {
	imageRepository := phase.ImagesRepo.ImageRepositoryName(img.GetName())
	lastStageImage := img.GetLastNonEmptyStage().GetImage()
	imageName := phase.ImagesRepo.ImageRepositoryNameWithTag(img.GetName(), imageMetaTag)
	imageActualTag := phase.ImagesRepo.ImageRepositoryTag(img.GetName(), imageMetaTag)

	alreadyExists, err := phase.checkImageAlreadyExists(initialExistingTagsList, img.GetName(), imageMetaTag, lastStageImage)
	if err != nil {
		return fmt.Errorf("error checking image %s already exists in the images repo: %s", img.LogName(), err)
	}

	if alreadyExists {
		logboek.Default.LogFHighlight("%s tag %s is up-to-date\n", strings.Title(string(tagStrategy)), imageActualTag)

		_ = logboek.WithIndent(func() error {
			logboek.Default.LogFDetails("images-repo: %s\n", imageRepository)
			logboek.Default.LogFDetails("      image: %s\n", imageName)
			return nil
		})

		logboek.LogOptionalLn()

		return nil
	}

	publishImage := container_runtime.NewWerfImage(phase.Conveyor.GetStageImage(lastStageImage.Name()), imageName, phase.Conveyor.ContainerRuntime.(*container_runtime.LocalDockerServerRuntime))

	publishImage.Container().ServiceCommitChangeOptions().AddLabel(map[string]string{
		image.WerfDockerImageName:  imageName,
		image.WerfTagStrategyLabel: string(tagStrategy),
		image.WerfImageLabel:       "true",
		image.WerfImageNameLabel:   img.GetName(),
		image.WerfImageTagLabel:    imageMetaTag,
	})

	successInfoSectionFunc := func() {
		_ = logboek.WithIndent(func() error {
			logboek.Default.LogFDetails("images-repo: %s\n", imageRepository)
			logboek.Default.LogFDetails("      image: %s\n", imageName)
			return nil
		})
	}

	publishingFunc := func() error {
		if err := fetchStage(phase.Conveyor.StagesStorage, img.GetLastNonEmptyStage()); err != nil {
			return err
		}

		if err := logboek.Info.LogProcess("Building final image with meta information", logboek.LevelLogProcessOptions{}, func() error {
			if err := publishImage.Build(container_runtime.BuildOptions{}); err != nil {
				return fmt.Errorf("error building %s with tagging strategy '%s': %s", imageName, tagStrategy, err)
			}
			return nil
		}); err != nil {
			return err
		}

		if err := phase.Conveyor.StorageLockManager.LockImage(imageName); err != nil {
			return fmt.Errorf("error locking image %s: %s", imageName, err)
		}
		defer phase.Conveyor.StorageLockManager.UnlockImage(imageName)

		existingTags, err := phase.fetchExistingTags(img.GetName())
		if err != nil {
			return err
		}

		alreadyExists, err := phase.checkImageAlreadyExists(existingTags, img.GetName(), imageMetaTag, lastStageImage)
		if err != nil {
			return fmt.Errorf("error checking image %s already exists in the images repo: %s", img.LogName(), err)
		}

		if alreadyExists {
			logboek.Default.LogFHighlight("%s tag %s is up-to-date\n", strings.Title(string(tagStrategy)), imageActualTag)
			_ = logboek.WithIndent(func() error {
				logboek.Info.LogFDetails("discarding newly built image %s\n", publishImage.MustGetBuiltId())
				logboek.Default.LogFDetails("images-repo: %s\n", imageRepository)
				logboek.Default.LogFDetails("      image: %s\n", imageName)

				return nil
			})

			logboek.LogOptionalLn()

			return nil
		}

		return phase.ImagesRepo.PublishImage(publishImage)

		return nil
	}

	return logboek.Default.LogProcess(
		fmt.Sprintf("Publishing image %s by %s tag %s", img.LogName(), tagStrategy, imageMetaTag),
		logboek.LevelLogProcessOptions{
			SuccessInfoSectionFunc: successInfoSectionFunc,
			Style:                  logboek.HighlightStyle(),
		},
		publishingFunc)
}

func (phase *PublishImagesPhase) checkImageAlreadyExists(existingTags []string, werfImageName, imageMetaTag string, lastStageImage container_runtime.ImageInterface) (bool, error) {
	imageActualTag := phase.ImagesRepo.ImageRepositoryTag(werfImageName, imageMetaTag)

	if !util.IsStringsContainValue(existingTags, imageActualTag) {
		return false, nil
	}

	var parentID string
	var err error
	getImageParentIDFunc := func() error {
		repoImage, err := phase.ImagesRepo.GetRepoImage(werfImageName, imageMetaTag)
		if err != nil {
			return err
		}

		parentID = repoImage.ParentID
		return nil
	}

	logProcessMsg := fmt.Sprintf("Getting existing tag %s parent id", imageActualTag)
	err = logboek.Info.LogProcessInline(logProcessMsg, logboek.LevelLogProcessInlineOptions{}, getImageParentIDFunc)
	if err != nil {
		return false, fmt.Errorf("unable to get image %s parent id: %s", werfImageName, err)
	}

	return lastStageImage.GetStagesStorageImageInfo().ID == parentID, nil
}
