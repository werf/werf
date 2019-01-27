package build

import (
	"fmt"

	"github.com/flant/werf/pkg/docker_registry"
	imagePkg "github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/logger"
	"github.com/flant/werf/pkg/util"
)

func NewPushPhase(repo string, opts PushOptions) *PushPhase {
	tagsByScheme := map[TagScheme][]string{
		CustomScheme:    opts.Tags,
		CIScheme:        opts.TagsByCI,
		GitBranchScheme: opts.TagsByGitBranch,
		GitTagScheme:    opts.TagsByGitTag,
		GitCommitScheme: opts.TagsByGitCommit,
	}
	return &PushPhase{Repo: repo, TagsByScheme: tagsByScheme, WithStages: opts.WithStages}
}

const (
	CustomScheme    TagScheme = "custom"
	GitTagScheme    TagScheme = "git_tag"
	GitBranchScheme TagScheme = "git_branch"
	GitCommitScheme TagScheme = "git_commit"
	CIScheme        TagScheme = "ci"

	RepoImageStageTagFormat = "image-stage-%s"
)

type TagScheme string

type PushPhase struct {
	WithStages   bool
	Repo         string
	TagsByScheme map[TagScheme][]string
}

func (p *PushPhase) Run(c *Conveyor) error {
	if debug() {
		fmt.Printf("PushPhase.Run\n")
	}

	err := c.GetDockerAuthorizer().LoginForPush(p.Repo)
	if err != nil {
		return fmt.Errorf("login into '%s' for push failed: %s", p.Repo, err)
	}

	for ind, image := range c.imagesInOrder {
		isLastImage := ind == len(c.imagesInOrder)-1
		err := logger.LogServiceProcess(fmt.Sprintf("Push %s", image.LogName()), "", func() error {
			if p.WithStages {
				err := logger.LogServiceProcess("Push stages cache", "", func() error {
					err := p.pushImageStages(c, image)
					if err != nil {
						return fmt.Errorf("unable to push image %s stages: %s", image.GetName(), err)
					}

					return nil
				})

				if err != nil {
					return err
				}

				if !image.isArtifact {
					fmt.Println()
				}
			}

			if !image.isArtifact {
				err := p.pushImage(c, image)
				if err != nil {
					return fmt.Errorf("unable to push image %s: %s", image.GetName(), err)
				}

				return nil

				if err != nil {
					return err
				}
			}

			return nil
		})

		if err != nil {
			return err
		}

		if !isLastImage {
			fmt.Println()
		}
	}

	return nil
}

func (p *PushPhase) pushImageStages(c *Conveyor, image *Image) error {
	stages := image.GetStages()

	existingStagesTags, err := docker_registry.ImageStagesTags(p.Repo)
	if err != nil {
		return fmt.Errorf("error fetching existing stages cache list %s: %s", p.Repo, err)
	}

	for ind, stage := range stages {
		isLastStage := ind == len(stages)-1

		stageTagName := fmt.Sprintf(RepoImageStageTagFormat, stage.GetSignature())
		stageImageName := fmt.Sprintf("%s:%s", p.Repo, stageTagName)

		if util.IsStringsContainValue(existingStagesTags, stageTagName) {
			logger.LogState(fmt.Sprintf("%s", stage.Name()), "[EXISTS]")

			logRepoImageInfo(stageImageName)

			if !isLastStage {
				fmt.Println()
			}

			continue
		}

		err := func() error {
			var err error

			imageLockName := fmt.Sprintf("image.%s", util.Sha256Hash(stageImageName))
			err = lock.Lock(imageLockName, lock.LockOptions{})
			if err != nil {
				return fmt.Errorf("failed to lock %s: %s", imageLockName, err)
			}
			defer lock.Unlock(imageLockName)

			stageImage := c.GetStageImage(stage.GetImage().Name())

			return logger.LogProcess(fmt.Sprintf("%s", stage.Name()), "[PUSHING]", func() error {
				err = stageImage.Export(stageImageName)
				if err != nil {
					return fmt.Errorf("error pushing %s: %s", stageImageName, err)
				}

				return nil
			})
		}()

		logRepoImageInfo(stageImageName)

		if err != nil {
			return err
		}

		if !isLastStage {
			fmt.Println()
		}
	}

	return nil
}

func (p *PushPhase) pushImage(c *Conveyor, image *Image) error {
	var imageRepository string
	if image.GetName() != "" {
		imageRepository = fmt.Sprintf("%s/%s", p.Repo, image.GetName())
	} else {
		imageRepository = p.Repo
	}

	existingTags, err := docker_registry.ImageTags(imageRepository)
	if err != nil {
		return fmt.Errorf("error fetch existing tags of image %s: %s", imageRepository, err)
	}

	stages := image.GetStages()
	lastStageImage := stages[len(stages)-1].GetImage()

	var nonEmptySchemeInOrder []TagScheme
	var lastNonEmptyTagScheme TagScheme
	for scheme, tags := range p.TagsByScheme {
		if len(tags) == 0 {
			continue
		}

		nonEmptySchemeInOrder = append(nonEmptySchemeInOrder, scheme)
		lastNonEmptyTagScheme = scheme
	}

	for _, scheme := range nonEmptySchemeInOrder {
		tags := p.TagsByScheme[scheme]

		if len(tags) == 0 {
			continue
		}

		err = logger.LogServiceProcess(fmt.Sprintf("%s scheme", string(scheme)), "", func() error {
		ProcessingTags:
			for ind, tag := range tags {
				isLastTag := ind == len(tags)-1

				imageName := fmt.Sprintf("%s:%s", imageRepository, tag)

				if util.IsStringsContainValue(existingTags, tag) {
					parentID, err := docker_registry.ImageParentId(imageName)
					if err != nil {
						return fmt.Errorf("unable to get image %s parent id: %s", imageName, err)
					}

					if lastStageImage.ID() == parentID {
						logger.LogState(tag, "[EXISTS]")
						logRepoImageInfo(imageName)

						if !isLastTag {
							fmt.Println()
						}

						continue ProcessingTags
					}
				}

				err := func() error {
					var err error

					imageLockName := fmt.Sprintf("image.%s", util.Sha256Hash(imageName))
					err = lock.Lock(imageLockName, lock.LockOptions{})
					if err != nil {
						return fmt.Errorf("failed to lock %s: %s", imageLockName, err)
					}
					defer lock.Unlock(imageLockName)

					pushImage := imagePkg.NewImage(c.GetStageImage(lastStageImage.Name()), imageName)

					pushImage.Container().ServiceCommitChangeOptions().AddLabel(map[string]string{
						imagePkg.WerfTagSchemeLabel: string(scheme),
						imagePkg.WerfImageLabel:     "true",
					})

					err = logger.LogProcessInline("Building final image with meta information", func() error {
						err = pushImage.Build(imagePkg.BuildOptions{})
						if err != nil {
							return fmt.Errorf("error building %s with tag scheme '%s': %s", imageName, scheme, err)
						}

						return nil
					})
					if err != nil {
						return err
					}

					return logger.LogProcess(tag, "[PUSHING]", func() error {
						err = pushImage.Export()
						if err != nil {
							return fmt.Errorf("error pushing %s: %s", imageName, err)
						}

						return nil
					})
				}()

				if err != nil {
					return err
				}

				logRepoImageInfo(imageName)

				if !isLastTag {
					fmt.Println()
				}
			}

			return nil
		})

		if err != nil {
			return err
		}

		if scheme != lastNonEmptyTagScheme {
			fmt.Println()
		}
	}

	return nil
}

func logRepoImageInfo(imageName string) {
	logger.LogInfoF(logImageInfoFormat, "repo-image", imageName)
}
