package build

import (
	"fmt"

	"github.com/flant/werf/pkg/docker_registry"
	imagePkg "github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/logger"
	"github.com/flant/werf/pkg/tag_scheme"
	"github.com/flant/werf/pkg/util"
)

const RepoImageStageTagFormat = "image-stage-%s"

func NewPublishImagesPhase(imagesRepo string, opts PublishImagesOptions) *PublishImagesPhase {
	tagsByScheme := map[tag_scheme.TagScheme][]string{
		tag_scheme.CustomScheme:    opts.Tags,
		tag_scheme.GitBranchScheme: opts.TagsByGitBranch,
		tag_scheme.GitTagScheme:    opts.TagsByGitTag,
		tag_scheme.GitCommitScheme: opts.TagsByGitCommit,
	}
	return &PublishImagesPhase{ImagesRepo: imagesRepo, TagsByScheme: tagsByScheme}
}

type PublishImagesPhase struct {
	WithStages   bool
	ImagesRepo   string
	TagsByScheme map[tag_scheme.TagScheme][]string
}

func (p *PublishImagesPhase) Run(c *Conveyor) error {
	return logger.LogServiceProcess("Push images", "", func() error {
		return logger.WithoutIndent(func() error { return p.run(c) })
	})
}

func (p *PublishImagesPhase) run(c *Conveyor) error {
	// TODO: Push stages should occur on the BuildStagesPhase

	for _, image := range c.imagesInOrder {
		err := logger.WithTag(image.LogName(), func() error {
			if p.WithStages {
				err := logger.LogServiceProcess("Push stages cache", "", func() error {
					if err := p.pushImageStages(c, image); err != nil {
						return fmt.Errorf("unable to push image %s stages: %s", image.GetName(), err)
					}

					return nil
				})

				logger.LogOptionalLn()

				if err != nil {
					return err
				}
			}

			if !image.isArtifact {
				if err := p.pushImage(c, image); err != nil {
					return fmt.Errorf("unable to push image %s: %s", image.GetName(), err)
				}

				logger.LogOptionalLn()
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (p *PublishImagesPhase) pushImageStages(c *Conveyor, image *Image) error {
	stages := image.GetStages()

	existingStagesTags, err := docker_registry.ImageStagesTags(p.ImagesRepo)
	if err != nil {
		return fmt.Errorf("error fetching existing stages cache list %s: %s", p.ImagesRepo, err)
	}

	for _, stage := range stages {
		stageTagName := fmt.Sprintf(RepoImageStageTagFormat, stage.GetSignature())
		stageImageName := fmt.Sprintf("%s:%s", p.ImagesRepo, stageTagName)

		if util.IsStringsContainValue(existingStagesTags, stageTagName) {
			logger.LogState(fmt.Sprintf("%s", stage.Name()), "[EXISTS]")

			logRepoImageInfo(stageImageName)

			logger.LogOptionalLn()

			continue
		}

		err := func() error {
			imageLockName := fmt.Sprintf("image.%s", util.Sha256Hash(stageImageName))

			if err := lock.Lock(imageLockName, lock.LockOptions{}); err != nil {
				return fmt.Errorf("failed to lock %s: %s", imageLockName, err)
			}

			defer lock.Unlock(imageLockName)

			stageImage := c.GetStageImage(stage.GetImage().Name())

			return logger.LogProcess(fmt.Sprintf("%s", stage.Name()), "[PUSHING]", func() error {
				if err := stageImage.Export(stageImageName); err != nil {
					return fmt.Errorf("error pushing %s: %s", stageImageName, err)
				}

				return nil
			})
		}()

		logRepoImageInfo(stageImageName)

		logger.LogOptionalLn()

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

	existingTags, err := docker_registry.ImageTags(imageRepository)
	if err != nil {
		return fmt.Errorf("error fetch existing tags of image %s: %s", imageRepository, err)
	}

	stages := image.GetStages()
	lastStageImage := stages[len(stages)-1].GetImage()

	var nonEmptySchemeInOrder []tag_scheme.TagScheme
	for scheme, tags := range p.TagsByScheme {
		if len(tags) == 0 {
			continue
		}

		nonEmptySchemeInOrder = append(nonEmptySchemeInOrder, scheme)
	}

	for _, scheme := range nonEmptySchemeInOrder {
		tags := p.TagsByScheme[scheme]

		if len(tags) == 0 {
			continue
		}

		err := logger.LogServiceProcess(fmt.Sprintf("%s scheme", string(scheme)), "", func() error {
		ProcessingTags:
			for _, tag := range tags {
				imageName := fmt.Sprintf("%s:%s", imageRepository, tag)

				if util.IsStringsContainValue(existingTags, tag) {
					parentID, err := docker_registry.ImageParentId(imageName)
					if err != nil {
						return fmt.Errorf("unable to get image %s parent id: %s", imageName, err)
					}

					if lastStageImage.ID() == parentID {
						logger.LogState(tag, "[EXISTS]")
						logRepoImageInfo(imageName)

						logger.LogOptionalLn()

						continue ProcessingTags
					}
				}

				err := func() error {
					imageLockName := fmt.Sprintf("image.%s", util.Sha256Hash(imageName))

					if err = lock.Lock(imageLockName, lock.LockOptions{}); err != nil {
						return fmt.Errorf("failed to lock %s: %s", imageLockName, err)
					}

					defer lock.Unlock(imageLockName)

					pushImage := imagePkg.NewImage(c.GetStageImage(lastStageImage.Name()), imageName)

					pushImage.Container().ServiceCommitChangeOptions().AddLabel(map[string]string{
						imagePkg.WerfTagSchemeLabel: string(scheme),
						imagePkg.WerfImageLabel:     "true",
					})

					err := logger.LogProcessInline("Building final image with meta information", func() error {
						if err := pushImage.Build(imagePkg.BuildOptions{}); err != nil {
							return fmt.Errorf("error building %s with tag scheme '%s': %s", imageName, scheme, err)
						}

						return nil
					})

					if err != nil {
						return err
					}

					return logger.LogProcess(tag, "[PUSHING]", func() error {
						if err := pushImage.Export(); err != nil {
							return fmt.Errorf("error pushing %s: %s", imageName, err)
						}

						return nil
					})
				}()

				if err != nil {
					return err
				}

				logRepoImageInfo(imageName)

				logger.LogOptionalLn()
			}

			return nil
		})

		logger.LogOptionalLn()

		if err != nil {
			return err
		}
	}

	return nil
}

func logRepoImageInfo(imageName string) {
	logger.LogInfoF(logImageInfoFormat, "repo-image", imageName)
}
