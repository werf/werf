package build

import (
	"fmt"

	imagePkg "github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/util"
)

func NewTagPhase(repo string, opts TagOptions) *TagPhase {
	tagsByScheme := map[TagScheme][]string{
		CustomScheme:    opts.Tags,
		CIScheme:        opts.TagsByCI,
		GitBranchScheme: opts.TagsByGitBranch,
		GitTagScheme:    opts.TagsByGitTag,
		GitCommitScheme: opts.TagsByGitCommit,
	}
	return &TagPhase{Repo: repo, TagsByScheme: tagsByScheme}
}

type TagPhase struct {
	Repo         string
	TagsByScheme map[TagScheme][]string
}

func (p *TagPhase) Run(c *Conveyor) error {
	if debug() {
		fmt.Printf("TagPhase.Run\n")
	}

	for _, image := range c.imagesInOrder {
		if !image.isArtifact {
			if image.GetName() == "" {
				fmt.Printf("# Tagging image\n")
			} else {
				fmt.Printf("# Tagging image/%s\n", image.GetName())
			}

			err := p.tagImage(c, image)
			if err != nil {
				return fmt.Errorf("unable to tag image %s: %s", image.GetName(), err)
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

	for scheme, tags := range p.TagsByScheme {
		for _, tag := range tags {
			imageImageName := fmt.Sprintf("%s:%s", imageRepository, tag)

			err := func() error {
				var err error

				imageLockName := fmt.Sprintf("image.%s", util.Sha256Hash(imageImageName))
				err = lock.Lock(imageLockName, lock.LockOptions{})
				if err != nil {
					return fmt.Errorf("failed to lock %s: %s", imageLockName, err)
				}
				defer lock.Unlock(imageLockName)

				fmt.Printf("# Build %s layer with tag scheme '%s'\n", imageImageName, scheme)

				tagImage := imagePkg.NewImage(c.GetStageImage(lastStageImage.Name()), imageImageName)

				tagImage.Container().ServiceCommitChangeOptions().AddLabel(map[string]string{
					"werf-tag-scheme": string(scheme),
					"werf-image":      "true",
				})

				err = tagImage.Build(imagePkg.BuildOptions{})
				if err != nil {
					return fmt.Errorf("error building %s with tag scheme '%s': %s", imageImageName, scheme, err)
				}

				if image.GetName() == "" {
					fmt.Printf("# Tagging image %s for image\n", imageImageName)
				} else {
					fmt.Printf("# Tagging image %s for image/%s\n", imageImageName, image.GetName())
				}

				err = tagImage.Tag()
				if err != nil {
					return fmt.Errorf("error tagging %s: %s", imageImageName, err)
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
