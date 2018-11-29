package build

import (
	"fmt"

	"github.com/flant/dapp/pkg/build/stage"
	"github.com/flant/dapp/pkg/dapp"
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
		var prevImage stage.Image

		err := dimg.PrepareBaseImage()
		if err != nil {
			return fmt.Errorf("error preparing base image %s of dimg %s: %s", dimg.GetBaseImage().GetName(), dimg.GetName(), err)
		}

		for _, stage := range dimg.GetStages() {
			image := stage.GetImage()

			image.AddServiceChangeLabel("dapp-version", dapp.Version)
			image.AddServiceChangeLabel("dapp-cache-version", BuildCacheVersion)
			image.AddServiceChangeLabel("dapp-dimg", "false")
			image.AddServiceChangeLabel("dapp-dev-mode", "false")

			if c.SshAuthSock != "" {
				image.AddVolume(fmt.Sprintf("%s:/tmp/dapp-ssh-agent", c.SshAuthSock))
				image.AddEnv(map[string]interface{}{
					"SSH_AUTH_SOCK": "/tmp/dapp-ssh-agent",
				})
			}

			err := stage.PrepareImage(prevImage, image)
			if err != nil {
				return fmt.Errorf("error preparing stage %s: %s", stage.Name(), err)
			}

			prevImage = image
		}
	}

	return nil
}
