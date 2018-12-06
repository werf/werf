package build

import (
	"fmt"

	"github.com/flant/dapp/pkg/image"
	"github.com/flant/dapp/pkg/lock"

	"github.com/flant/dapp/pkg/docker_registry"
	"github.com/flant/dapp/pkg/util"
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

	for _, dimg := range c.DimgsInOrder {
		stages := dimg.GetStages()

		if p.WithStages {
			existingStagesTags, err := docker_registry.DimgstageTags(p.Repo)
			if err != nil {
				return fmt.Errorf("error fetching existing stages cache list %s: %s", p.Repo, err)
			}

			for _, stage := range stages {
				stageTagName := fmt.Sprintf("dimgstage-%s", stage.GetSignature())
				stageImageName := fmt.Sprintf("%s:%s", p.Repo, stageTagName)

				if util.IsStringsContainValue(existingStagesTags, stageTagName) {
					fmt.Printf("Stage %s EXIST\n", stageImageName)
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

					fmt.Printf("Push %s\n", stageImageName)

					stageImage := c.GetImage(stage.GetImage().Name())

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
		}

		var dimgRef string
		if dimg.GetName() != "" {
			dimgRef = fmt.Sprintf("%s/%s", p.Repo, dimg.GetName())
		} else {
			dimgRef = p.Repo
		}

		existingTags, err := docker_registry.DimgTags(dimgRef)
		if err != nil {
			return fmt.Errorf("error fetch existing tags of dimg %s: %s", dimgRef, err)
		}

		lastStageImage := stages[len(stages)-1].GetImage()

		for scheme, tags := range p.TagsByScheme {
		ProcessingTags:
			for _, tag := range tags {
				fullRef := fmt.Sprintf("%s:%s", dimgRef, tag)

				if util.IsStringsContainValue(existingTags, tag) {
					parentID, err := docker_registry.ImageParentId(fullRef)
					if err != nil {
						return fmt.Errorf("unable to get image %s parent id: %s", fullRef, err)
					}

					if lastStageImage.ID() == parentID {
						fmt.Printf("%s EXIST\n", fullRef)
						continue ProcessingTags
					}
				}

				err := func() error {
					var err error

					imageLockName := fmt.Sprintf("image.%s", util.Sha256Hash(fullRef))
					err = lock.Lock(imageLockName, lock.LockOptions{})
					if err != nil {
						return fmt.Errorf("failed to lock %s: %s", imageLockName, err)
					}
					defer lock.Unlock(imageLockName)

					fmt.Printf("Build %s layer with tag scheme '%s'\n", fullRef, scheme)

					pushImage := image.NewDimgImage(c.GetImage(lastStageImage.Name()), fullRef)

					pushImage.Container().ServiceCommitChangeOptions().AddLabel(map[string]string{
						"dapp-tag-scheme": string(scheme),
						"dapp-dimg":       "true",
					})

					err = pushImage.Build2(image.BuildOptions{})
					if err != nil {
						return fmt.Errorf("error building %s with tag scheme '%s': %s", fullRef, scheme, err)
					}

					fmt.Printf("Push %s\n", fullRef)

					err = pushImage.Export()
					if err != nil {
						return fmt.Errorf("error pushing %s: %s", fullRef, err)
					}

					return nil
				}()

				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
