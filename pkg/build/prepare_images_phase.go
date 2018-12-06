package build

import (
	"fmt"

	"github.com/flant/dapp/pkg/dapp"
	"github.com/flant/dapp/pkg/image"
)

func NewPrepareImagesPhase() *PrepareImagesPhase {
	return &PrepareImagesPhase{}
}

type PrepareImagesPhase struct{}

const DappCacheVersionLabel = "dapp-cache-version"

func (p *PrepareImagesPhase) Run(c *Conveyor) error {
	if debug() {
		fmt.Printf("PrepareImagesPhase.Run\n")
	}

	for _, dimg := range c.DimgsInOrder {
		if debug() {
			fmt.Printf("  dimg: '%s'\n", dimg.GetName())
		}

		var prevImage, prevBuiltImage image.Image

		err := dimg.PrepareBaseImage(c)
		if err != nil {
			return fmt.Errorf("error preparing base image %s of dimg %s: %s", dimg.GetBaseImage().Name(), dimg.GetName(), err)
		}

		prevImage = dimg.baseImage
		for _, s := range dimg.GetStages() {
			if prevImage.IsExists() {
				prevBuiltImage = prevImage
			}

			img := s.GetImage()
			if !img.IsExists() {
				if debug() {
					fmt.Printf("    %s\n", s.Name())
				}

				imageServiceCommitChangeOptions := img.Container().ServiceCommitChangeOptions()
				imageServiceCommitChangeOptions.AddLabel(map[string]string{
					"dapp":                c.ProjectName,
					"dapp-version":        dapp.Version,
					DappCacheVersionLabel: BuildCacheVersion,
					"dapp-dimg":           "false",
					"dapp-dev-mode":       "false",
				})

				if c.SSHAuthSock != "" {
					imageRunOptions := img.Container().RunOptions()
					imageRunOptions.AddVolume(fmt.Sprintf("%s:/tmp/dapp-ssh-agent", c.SSHAuthSock))
					imageRunOptions.AddEnv(map[string]string{"SSH_AUTH_SOCK": "/tmp/dapp-ssh-agent"})
				}

				err := s.PrepareImage(c, prevBuiltImage, img)
				if err != nil {
					return fmt.Errorf("error preparing stage %s: %s", s.Name(), err)
				}
			}

			prevImage = img
		}
	}

	return nil
}
