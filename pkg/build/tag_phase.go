package build

import (
	"fmt"

	imagePkg "github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/logger"
	"github.com/flant/werf/pkg/tag_strategy"
	"github.com/flant/werf/pkg/util"
)

func NewTagPhase(repo string, opts TagOptions) *TagPhase {
	tagsByScheme := map[tag_strategy.TagStrategy][]string{
		tag_strategy.Custom:    opts.CustomTags,
		tag_strategy.GitBranch: opts.TagsByGitBranch,
		tag_strategy.GitTag:    opts.TagsByGitTag,
		tag_strategy.GitCommit: opts.TagsByGitCommit,
	}
	return &TagPhase{Repo: repo, TagsByScheme: tagsByScheme}
}

type TagPhase struct {
	Repo         string
	TagsByScheme map[tag_strategy.TagStrategy][]string
}

func (p *TagPhase) Run(c *Conveyor) error {
	if debug() {
		fmt.Fprintf(logger.GetOutStream(), "TagPhase.Run\n")
	}

	for _, image := range c.imagesInOrder {
		if !image.isArtifact {
			err := logger.LogServiceProcess(fmt.Sprintf("Tag %s", image.LogName()), "", func() error {
				if err := p.tagImage(c, image); err != nil {
					return fmt.Errorf("unable to tag image %s: %s", image.GetName(), err)
				}

				return nil
			})

			logger.LogOptionalLn()

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *TagPhase) tagImage(c *Conveyor, image *Image) error {
	var imageRepository string
	if image.GetName() != "" {
		imageRepository = fmt.Sprintf("%s/%s", p.Repo, image.GetName())
	} else {
		imageRepository = p.Repo
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

		err := logger.LogServiceProcess(fmt.Sprintf("%s tagging strategy", string(strategy)), "", func() error {
			for _, tag := range tags {
				imageName := fmt.Sprintf("%s:%s", imageRepository, tag)

				err := func() error {
					imageLockName := fmt.Sprintf("image.%s", util.Sha256Hash(imageName))

					if err := lock.Lock(imageLockName, lock.LockOptions{}); err != nil {
						return fmt.Errorf("failed to lock %s: %s", imageLockName, err)
					}

					defer lock.Unlock(imageLockName)

					tagImage := imagePkg.NewImage(c.GetStageImage(lastStageImage.Name()), imageName)

					tagImage.Container().ServiceCommitChangeOptions().AddLabel(map[string]string{
						imagePkg.WerfTagStrategyLabel: string(strategy),
						imagePkg.WerfImageLabel:       "true",
					})

					err := logger.LogProcess(fmt.Sprintf("Building final image with meta information"), "", func() error {
						if err := tagImage.Build(imagePkg.BuildOptions{}); err != nil {
							return fmt.Errorf("error building %s: %s", tag, err)
						}

						return nil
					})

					if err != nil {
						return err
					}

					err = logger.LogProcessInline(fmt.Sprintf("Tagging %s", tag), func() error {
						if err = tagImage.Tag(); err != nil {
							return fmt.Errorf("error tagging %s: %s", imageName, err)
						}

						return nil
					})

					if err != nil {
						return err
					}

					tagImageId, err := tagImage.MustGetId()
					if err != nil {
						return err
					}

					logger.LogInfoF(logImageInfoFormat, "id", tagImageId)
					logger.LogInfoF(logImageInfoFormat, "image", imageName)

					return nil
				}()

				logger.LogOptionalLn()

				if err != nil {
					return err
				}
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
