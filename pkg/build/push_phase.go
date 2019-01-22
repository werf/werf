package build

import (
	"fmt"

	"github.com/flant/werf/pkg/docker_registry"
	imagePkg "github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/lock"
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

	for _, image := range c.imagesInOrder {
		if p.WithStages {
			if image.GetName() == "" {
				fmt.Printf("# Pushing image stages cache\n")
			} else {
				fmt.Printf("# Pushing image/%s stages cache\n", image.GetName())
			}

			err := p.pushImageStages(c, image)
			if err != nil {
				return fmt.Errorf("unable to push image %s stages: %s", image.GetName(), err)
			}
		}

		if !image.isArtifact {
			if image.GetName() == "" {
				fmt.Printf("# Pushing image\n")
			} else {
				fmt.Printf("# Pushing image/%s\n", image.GetName())
			}

			err := p.pushImage(c, image)
			if err != nil {
				return fmt.Errorf("unable to push image %s: %s", image.GetName(), err)
			}
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

	for _, stage := range stages {
		stageTagName := fmt.Sprintf(RepoImageStageTagFormat, stage.GetSignature())
		stageImageName := fmt.Sprintf("%s:%s", p.Repo, stageTagName)

		if util.IsStringsContainValue(existingStagesTags, stageTagName) {
			if image.GetName() == "" {
				fmt.Printf("# Ignore existing in repo image %s for image stage/%s\n", stageImageName, stage.Name())
			} else {
				fmt.Printf("# Ignore existing in repo image %s for image/%s stage/%s\n", stageImageName, image.GetName(), stage.Name())
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

			if image.GetName() == "" {
				fmt.Printf("# Pushing image %s for image stage/%s\n", stageImageName, stage.Name())
			} else {
				fmt.Printf("# Pushing image %s for image/%s stage/%s\n", stageImageName, image.GetName(), stage.Name())
			}

			stageImage := c.GetStageImage(stage.GetImage().Name())

			err = stageImage.Export(stageImageName)
			if err != nil {
				return fmt.Errorf("error pushing %s: %s", stageImageName, err)
			}

			return nil
		}()

		if err != nil {
			return err
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

	for scheme, tags := range p.TagsByScheme {
	ProcessingTags:
		for _, tag := range tags {
			imageImageName := fmt.Sprintf("%s:%s", imageRepository, tag)

			if util.IsStringsContainValue(existingTags, tag) {
				parentID, err := docker_registry.ImageParentId(imageImageName)
				if err != nil {
					return fmt.Errorf("unable to get image %s parent id: %s", imageImageName, err)
				}

				if lastStageImage.ID() == parentID {
					if image.GetName() == "" {
						fmt.Printf("# Ignore existing in repo image %s for image\n", imageImageName)
					} else {
						fmt.Printf("# Ignore existing in repo image %s for image/%s\n", imageImageName, image.GetName())
					}
					continue ProcessingTags
				}
			}

			err := func() error {
				var err error

				imageLockName := fmt.Sprintf("image.%s", util.Sha256Hash(imageImageName))
				err = lock.Lock(imageLockName, lock.LockOptions{})
				if err != nil {
					return fmt.Errorf("failed to lock %s: %s", imageLockName, err)
				}
				defer lock.Unlock(imageLockName)

				fmt.Printf("# Build %s layer with tag scheme '%s'\n", imageImageName, scheme)

				pushImage := imagePkg.NewImage(c.GetStageImage(lastStageImage.Name()), imageImageName)

				pushImage.Container().ServiceCommitChangeOptions().AddLabel(map[string]string{
					"werf-tag-scheme": string(scheme),
					"werf-image":      "true",
				})

				err = pushImage.Build(imagePkg.BuildOptions{})
				if err != nil {
					return fmt.Errorf("error building %s with tag scheme '%s': %s", imageImageName, scheme, err)
				}

				if image.GetName() == "" {
					fmt.Printf("# Pushing image %s for image\n", imageImageName)
				} else {
					fmt.Printf("# Pushing image %s for image/%s\n", imageImageName, image.GetName())
				}

				err = pushImage.Export()
				if err != nil {
					return fmt.Errorf("error pushing %s: %s", imageImageName, err)
				}

				return nil
			}()

			if err != nil {
				return err
			}
		}
	}

	return nil
}
