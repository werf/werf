package build

import (
	"fmt"

	image "github.com/flant/werf/pkg/image"

	"github.com/flant/logboek"
	"github.com/flant/shluz"
	"github.com/flant/werf/pkg/build/stage"
	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/tag_strategy"
	"github.com/flant/werf/pkg/util"
)

func NewPublishImagesPhase(c *Conveyor, imagesRepoManager ImagesRepoManager, opts PublishImagesOptions) *PublishImagesPhase {
	tagsByScheme := map[tag_strategy.TagStrategy][]string{
		tag_strategy.Custom:    opts.CustomTags,
		tag_strategy.GitBranch: opts.TagsByGitBranch,
		tag_strategy.GitTag:    opts.TagsByGitTag,
		tag_strategy.GitCommit: opts.TagsByGitCommit,
	}
	return &PublishImagesPhase{BasePhase: BasePhase{c}, TagsByScheme: tagsByScheme, ImageRepoManager: imagesRepoManager}
}

type PublishImagesPhase struct {
	BasePhase
	ImagesToPublish  []string
	TagsByScheme     map[tag_strategy.TagStrategy][]string
	ImageRepoManager ImagesRepoManager
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

func (phase *PublishImagesPhase) OnImageStage(img *Image, stg stage.Interface) (bool, error) {
	return true, nil
}

func (phase *PublishImagesPhase) AfterImageStages(img *Image) error {
	if len(phase.ImagesToPublish) == 0 {
		return phase.pushImage(img)
	}

	for _, name := range phase.ImagesToPublish {
		if name == img.GetName() {
			return phase.pushImage(img)
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
			//		if err := p.pushImageStages(c, image); err != nil {
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
				if err := p.pushImage(c, image); err != nil {
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

//func (p *PublishImagesPhase) pushImageStages(c *Conveyor, image *Image) error {
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

func (phase *PublishImagesPhase) pushImage(img *Image) error {
	imageRepository := phase.ImageRepoManager.ImageRepo(img.GetName())

	var existingTags []string
	var err error
	fetchExistingTagsFunc := func() error {
		existingTags, err = docker_registry.Tags(imageRepository)
		return err
	}

	if debug() {
		err = logboek.LogProcessInline("Fetching existing image tags", logboek.LogProcessInlineOptions{}, fetchExistingTagsFunc)
		logboek.LogOptionalLn()
	} else {
		err = fetchExistingTagsFunc()
	}

	if err != nil {
		return fmt.Errorf("error fetch existing tags of image %s: %s", imageRepository, err)
	}

	stages := img.GetStages()
	lastStageImage := stages[len(stages)-1].GetImage()

	var nonEmptySchemeInOrder []tag_strategy.TagStrategy
	for strategy, tags := range phase.TagsByScheme {
		if len(tags) == 0 {
			continue
		}

		nonEmptySchemeInOrder = append(nonEmptySchemeInOrder, strategy)
	}

	for _, strategy := range nonEmptySchemeInOrder {
		imageMetaTags := phase.TagsByScheme[strategy]

		if len(imageMetaTags) == 0 {
			continue
		}

		logProcessOptions := logboek.LogProcessOptions{ColorizeMsgFunc: logboek.ColorizeHighlight}
		err := logboek.LogProcess(fmt.Sprintf("%s tagging strategy", string(strategy)), logProcessOptions, func() error {
		ProcessingTags:
			for _, imageMetaTag := range imageMetaTags {
				imageName := phase.ImageRepoManager.ImageRepoWithTag(img.GetName(), imageMetaTag)
				imageTag := phase.ImageRepoManager.ImageRepoTag(img.GetName(), imageMetaTag)
				tagLogName := fmt.Sprintf("tag %s", imageTag)

				if util.IsStringsContainValue(existingTags, imageTag) {
					var parentID string
					var err error
					getImageParentIDFunc := func() error {
						parentID, err = docker_registry.ImageParentId(imageName)
						return err
					}

					if debug() {
						logProcessMsg := fmt.Sprintf("Getting existing tag %s parent id", imageTag)
						err = logboek.LogProcessInline(logProcessMsg, logboek.LogProcessInlineOptions{}, getImageParentIDFunc)
						logboek.LogOptionalLn()
					} else {
						err = getImageParentIDFunc()
					}

					if err != nil {
						return fmt.Errorf("unable to get image %s parent id: %s", imageName, err)
					}

					if lastStageImage.ID() == parentID {
						logboek.LogHighlightF("Tag %s is up-to-date\n", imageTag)
						_ = logboek.WithIndent(func() error {
							logboek.LogInfoF("images-repo: %s\n", imageRepository)
							logboek.LogInfoF("      image: %s\n", imageName)

							return nil
						})

						logboek.LogOptionalLn()

						continue ProcessingTags
					}
				}

				err := func() error {
					imageLockName := image.ImageLockName(imageName)
					if err = shluz.Lock(imageLockName, shluz.LockOptions{}); err != nil {
						return fmt.Errorf("failed to lock %s: %s", imageLockName, err)
					}
					defer shluz.Unlock(imageLockName)

					pushImage := image.NewImage(phase.Conveyor.GetStageImage(lastStageImage.Name()), imageName)

					pushImage.Container().ServiceCommitChangeOptions().AddLabel(map[string]string{
						image.WerfDockerImageName:  imageName,
						image.WerfTagStrategyLabel: string(strategy),
						image.WerfImageLabel:       "true",
						image.WerfImageNameLabel:   img.GetName(),
						image.WerfImageTagLabel:    imageMetaTag,
					})

					successInfoSectionFunc := func() {
						_ = logboek.WithIndent(func() error {
							logboek.LogInfoF("images-repo: %s\n", imageRepository)
							logboek.LogInfoF("      image: %s\n", imageName)

							return nil
						})
					}
					logProcessOptions := logboek.LogProcessOptions{SuccessInfoSectionFunc: successInfoSectionFunc, ColorizeMsgFunc: logboek.ColorizeHighlight}
					return logboek.LogProcess(fmt.Sprintf("Publishing %s", tagLogName), logProcessOptions, func() error {
						if err := logboek.LogProcess("Building final image with meta information", logboek.LogProcessOptions{}, func() error {
							if err := pushImage.Build(image.BuildOptions{}); err != nil {
								return fmt.Errorf("error building %s with tagging strategy '%s': %s", imageName, strategy, err)
							}

							return nil
						}); err != nil {
							return err
						}

						if err := pushImage.Export(); err != nil {
							return fmt.Errorf("error pushing %s: %s", imageName, err)
						}

						return nil
					})
				}()

				if err != nil {
					return err
				}
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}
