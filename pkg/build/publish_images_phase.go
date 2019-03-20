package build

import (
	"fmt"

	"github.com/flant/werf/pkg/docker_registry"
	imagePkg "github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/logger"
	"github.com/flant/werf/pkg/tag_strategy"
	"github.com/flant/werf/pkg/util"
)

const RepoImageStageTagFormat = "image-stage-%s"

func NewPublishImagesPhase(imagesRepo string, opts PublishImagesOptions) *PublishImagesPhase {
	tagsByScheme := map[tag_strategy.TagStrategy][]string{
		tag_strategy.Custom:    opts.CustomTags,
		tag_strategy.GitBranch: opts.TagsByGitBranch,
		tag_strategy.GitTag:    opts.TagsByGitTag,
		tag_strategy.GitCommit: opts.TagsByGitCommit,
	}
	return &PublishImagesPhase{ImagesRepo: imagesRepo, TagsByScheme: tagsByScheme}
}

type PublishImagesPhase struct {
	WithStages   bool
	ImagesRepo   string
	TagsByScheme map[tag_strategy.TagStrategy][]string
}

func (p *PublishImagesPhase) Run(c *Conveyor) error {
	return logger.LogProcess("Publishing images", logger.LogProcessOptions{}, func() error {
		return p.run(c)
	})
}

func (p *PublishImagesPhase) run(c *Conveyor) error {
	// TODO: Push stages should occur on the BuildStagesPhase

	for _, image := range c.imagesInOrder {
		if err := logger.LogProcess(image.LogProcessName(), logger.LogProcessOptions{ColorizeMsgFunc: image.LogProcessColorizeFunc()}, func() error {
			if p.WithStages {
				err := logger.LogSecondaryProcess("Pushing stages cache", logger.LogProcessOptions{}, func() error {
					if err := p.pushImageStages(c, image); err != nil {
						return fmt.Errorf("unable to push image %s stages: %s", image.GetName(), err)
					}

					return nil
				})

				if err != nil {
					return err
				}
			}

			if !image.isArtifact {
				if err := p.pushImage(c, image); err != nil {
					return fmt.Errorf("unable to push image %s: %s", image.GetName(), err)
				}
			}

			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

func (p *PublishImagesPhase) pushImageStages(c *Conveyor, image *Image) error {
	stages := image.GetStages()

	existingTags, err := docker_registry.Tags(p.ImagesRepo)
	if err != nil {
		return fmt.Errorf("error fetching existing stages cache list %s: %s", p.ImagesRepo, err)
	}

	for _, stage := range stages {
		stageTagName := fmt.Sprintf(RepoImageStageTagFormat, stage.GetSignature())
		stageImageName := fmt.Sprintf("%s:%s", p.ImagesRepo, stageTagName)

		if util.IsStringsContainValue(existingTags, stageTagName) {
			logger.LogHighlightLn(stage.Name())

			logger.LogInfoF("stages-repo: %s\n", p.ImagesRepo)
			logger.LogInfoF("      image: %s\n", stageImageName)

			logger.OptionalLnModeOn()

			continue
		}

		err := func() error {
			imageLockName := imagePkg.GetImageLockName(stageImageName)

			if err := lock.Lock(imageLockName, lock.LockOptions{}); err != nil {
				return fmt.Errorf("failed to lock %s: %s", imageLockName, err)
			}

			defer lock.Unlock(imageLockName)

			stageImage := c.GetStageImage(stage.GetImage().Name())

			infoSectionFunc := func(err error) {
				if err == nil {
					_ = logger.WithIndent(func() error {
						logger.LogInfoF("stages-repo: %s\n", p.ImagesRepo)
						logger.LogInfoF("      image: %s\n", stageImageName)

						return nil
					})
				}
			}

			logProcessOptions := logger.LogProcessOptions{InfoSectionFunc: infoSectionFunc}
			return logger.LogProcess(fmt.Sprintf("Publishing %s", stage.Name()), logProcessOptions, func() error {
				if err := stageImage.Export(stageImageName); err != nil {
					return fmt.Errorf("error pushing %s: %s", stageImageName, err)
				}

				return nil
			})
		}()

		if err != nil {
			return err
		}
	}

	return nil
}

func (p *PublishImagesPhase) pushImage(c *Conveyor, image *Image) error {
	var imageRepository string
	if image.GetName() != "" {
		imageRepository = fmt.Sprintf("%s/%s", p.ImagesRepo, image.GetName())
	} else {
		imageRepository = p.ImagesRepo
	}

	var existingTags []string
	var err error
	fetchExistingTagsFunc := func() error {
		existingTags, err = docker_registry.Tags(imageRepository)
		return err
	}

	if debug() {
		err = logger.LogSecondaryProcessInline("Fetching existing image tags", fetchExistingTagsFunc)
		logger.OptionalLnModeOn()
	} else {
		err = fetchExistingTagsFunc()
	}

	if err != nil {
		return fmt.Errorf("error fetch existing tags of image %s: %s", imageRepository, err)
	}

	stages := image.GetStages()
	lastStageImage := stages[len(stages)-1].GetImage()

	var nonEmptySchemeInOrder []tag_strategy.TagStrategy
	for strategy, tags := range p.TagsByScheme {
		if len(tags) == 0 {
			continue
		}

		nonEmptySchemeInOrder = append(nonEmptySchemeInOrder, strategy)
	}

	for _, strategy := range nonEmptySchemeInOrder {
		tags := p.TagsByScheme[strategy]

		if len(tags) == 0 {
			continue
		}

		err := logger.LogProcess(fmt.Sprintf("%s tagging strategy", string(strategy)), logger.LogProcessOptions{}, func() error {
		ProcessingTags:
			for _, tag := range tags {
				tagLogName := fmt.Sprintf("tag %s", tag)
				imageName := fmt.Sprintf("%s:%s", imageRepository, tag)

				if util.IsStringsContainValue(existingTags, tag) {
					var parentID string
					var err error
					getImageParentIDFunc := func() error {
						parentID, err = docker_registry.ImageParentId(imageName)
						return err
					}

					if debug() {
						logProcessMsg := fmt.Sprintf("Getting existing tag %s parent id", tag)
						err = logger.LogSecondaryProcessInline(logProcessMsg, getImageParentIDFunc)
						logger.OptionalLnModeOn()
					} else {
						err = getImageParentIDFunc()
					}

					if err != nil {
						return fmt.Errorf("unable to get image %s parent id: %s", imageName, err)
					}

					if lastStageImage.ID() == parentID {
						logger.LogHighlightF("Tag %s is up-to-date\n", tag)
						_ = logger.WithIndent(func() error {
							logger.LogInfoF("images-repo: %s\n", imageRepository)
							logger.LogInfoF("      image: %s\n", imageName)

							return nil
						})

						logger.OptionalLnModeOn()

						continue ProcessingTags
					}
				}

				err := func() error {
					imageLockName := imagePkg.GetImageLockName(imageName)
					if err = lock.Lock(imageLockName, lock.LockOptions{}); err != nil {
						return fmt.Errorf("failed to lock %s: %s", imageLockName, err)
					}
					defer lock.Unlock(imageLockName)

					pushImage := imagePkg.NewImage(c.GetStageImage(lastStageImage.Name()), imageName)

					pushImage.Container().ServiceCommitChangeOptions().AddLabel(map[string]string{
						imagePkg.WerfDockerImageName:  imageName,
						imagePkg.WerfTagStrategyLabel: string(strategy),
						imagePkg.WerfImageLabel:       "true",
					})

					infoSectionFunc := func(err error) {
						if err == nil {
							_ = logger.WithIndent(func() error {
								logger.LogInfoF("images-repo: %s\n", imageRepository)
								logger.LogInfoF("      image: %s\n", imageName)

								return nil
							})
						}
					}
					logProcessOptions := logger.LogProcessOptions{InfoSectionFunc: infoSectionFunc}
					return logger.LogProcess(fmt.Sprintf("Publishing %s", tagLogName), logProcessOptions, func() error {
						if err := logger.LogSecondaryProcess("Building final image with meta information", logger.LogProcessOptions{}, func() error {
							if err := pushImage.Build(imagePkg.BuildOptions{}); err != nil {
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
