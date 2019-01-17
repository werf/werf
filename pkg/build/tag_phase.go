package build

import (
	"fmt"

	"github.com/flant/werf/pkg/image"
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

	for _, dimg := range c.dimgsInOrder {
		if !dimg.isArtifact {
			if dimg.GetName() == "" {
				fmt.Printf("# Tagging dimg\n")
			} else {
				fmt.Printf("# Tagging dimg/%s\n", dimg.GetName())
			}

			err := p.tagDimg(c, dimg)
			if err != nil {
				return fmt.Errorf("unable to tag dimg %s: %s", dimg.GetName(), err)
			}
		}
	}

	return nil
}

func (p *TagPhase) tagDimg(c *Conveyor, dimg *Dimg) error {
	var dimgRepository string
	if dimg.GetName() != "" {
		dimgRepository = fmt.Sprintf("%s/%s", p.Repo, dimg.GetName())
	} else {
		dimgRepository = p.Repo
	}

	stages := dimg.GetStages()
	lastStageImage := stages[len(stages)-1].GetImage()

	for scheme, tags := range p.TagsByScheme {
		for _, tag := range tags {
			dimgImageName := fmt.Sprintf("%s:%s", dimgRepository, tag)

			err := func() error {
				var err error

				imageLockName := fmt.Sprintf("image.%s", util.Sha256Hash(dimgImageName))
				err = lock.Lock(imageLockName, lock.LockOptions{})
				if err != nil {
					return fmt.Errorf("failed to lock %s: %s", imageLockName, err)
				}
				defer lock.Unlock(imageLockName)

				fmt.Printf("# Build %s layer with tag scheme '%s'\n", dimgImageName, scheme)

				tagImage := image.NewDimgImage(c.GetImage(lastStageImage.Name()), dimgImageName)

				tagImage.Container().ServiceCommitChangeOptions().AddLabel(map[string]string{
					"werf-tag-scheme": string(scheme),
					"werf-dimg":       "true",
				})

				err = tagImage.Build(image.BuildOptions{})
				if err != nil {
					return fmt.Errorf("error building %s with tag scheme '%s': %s", dimgImageName, scheme, err)
				}

				if dimg.GetName() == "" {
					fmt.Printf("# Tagging image %s for dimg\n", dimgImageName)
				} else {
					fmt.Printf("# Tagging image %s for dimg/%s\n", dimgImageName, dimg.GetName())
				}

				err = tagImage.Tag()
				if err != nil {
					return fmt.Errorf("error tagging %s: %s", dimgImageName, err)
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
