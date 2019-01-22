package build

import (
	"fmt"

	imagePkg "github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/werf"
)

func NewPrepareImagesPhase() *PrepareImagesPhase {
	return &PrepareImagesPhase{}
}

type PrepareImagesPhase struct{}

const WerfCacheVersionLabel = "werf-cache-version"

func (p *PrepareImagesPhase) Run(c *Conveyor) error {
	if debug() {
		fmt.Printf("PrepareImagesPhase.Run\n")
	}

	for _, image := range c.imagesInOrder {
		if debug() {
			fmt.Printf("  image: '%s'\n", image.GetName())
		}

		var prevImage, prevBuiltImage imagePkg.ImageInterface

		err := image.PrepareBaseImage(c)
		if err != nil {
			return fmt.Errorf("error preparing base image %s of image %s: %s", image.GetBaseImage().Name(), image.GetName(), err)
		}

		prevImage = image.baseImage
		for _, s := range image.GetStages() {
			if prevImage.IsExists() {
				prevBuiltImage = prevImage
			}

			stageImage := s.GetImage()

			if c.GetImageBySignature(s.GetSignature()) != nil || stageImage.IsExists() {
				prevImage = stageImage
				continue
			}

			if debug() {
				fmt.Printf("    %s\n", s.Name())
			}

			imageServiceCommitChangeOptions := stageImage.Container().ServiceCommitChangeOptions()
			imageServiceCommitChangeOptions.AddLabel(map[string]string{
				"werf":                c.projectName(),
				"werf-version":        werf.Version,
				WerfCacheVersionLabel: BuildCacheVersion,
				"werf-image":          "false",
				"werf-dev-mode":       "false",
			})

			if c.sshAuthSock != "" {
				imageRunOptions := stageImage.Container().RunOptions()
				imageRunOptions.AddVolume(fmt.Sprintf("%s:/tmp/werf-ssh-agent", c.sshAuthSock))
				imageRunOptions.AddEnv(map[string]string{"SSH_AUTH_SOCK": "/tmp/werf-ssh-agent"})
			}

			err := s.PrepareImage(c, prevBuiltImage, stageImage)
			if err != nil {
				return fmt.Errorf("error preparing stage %s: %s", s.Name(), err)
			}

			c.SetImageBySignature(s.GetSignature(), stageImage)

			if image.GetName() == "" {
				fmt.Printf("# Prepared for build image %s for image %s\n", stageImage.Name(), fmt.Sprintf("stage/%s", s.Name()))
			} else {
				fmt.Printf("# Prepared for build image %s for image/%s %s\n", stageImage.Name(), image.GetName(), fmt.Sprintf("stage/%s", s.Name()))
			}

			prevImage = stageImage
		}
	}

	return nil
}
