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

func (p *PrepareImagesPhase) Run(c *Conveyor) error {
	if debug() {
		fmt.Printf("PrepareImagesPhase.Run\n")
	}

	for _, dimg := range c.GetDimgsInOrder() {
		var prevBuiltImage image.Image

		err := dimg.PrepareBaseImage()
		if err != nil {
			return fmt.Errorf("error preparing base image %s of dimg %s: %s", dimg.GetBaseImage().Name(), dimg.GetName(), err)
		}

		for _, stage := range dimg.GetStages() {
			image := stage.GetImage()

			imageServiceCommitChangeOptions := image.Container().ServiceCommitChangeOptions()
			imageServiceCommitChangeOptions.AddLabel(map[string]string{
				"dapp-version":       dapp.Version,
				"dapp-cache-version": BuildCacheVersion,
				"dapp-dimg":          "false",
				"dapp-dev-mode":      "false",
			})

			if c.SshAuthSock != "" {
				imageRunOptions := image.Container().RunOptions()
				imageRunOptions.AddVolume(fmt.Sprintf("%s:/tmp/dapp-ssh-agent", c.SshAuthSock))
				imageRunOptions.AddEnv(map[string]string{"SSH_AUTH_SOCK": "/tmp/dapp-ssh-agent"})
			}

			err := stage.PrepareImage(prevBuiltImage, image)
			if err != nil {
				return fmt.Errorf("error preparing stage %s: %s", stage.Name(), err)
			}

			if image.IsExists() {
				prevBuiltImage = image
			}
		}
	}

	return nil
}
