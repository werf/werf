package build

import (
	"fmt"

	"github.com/flant/werf/pkg/logger"

	imagePkg "github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/werf"
)

func NewPrepareStagesPhase() *PrepareStagesPhase {
	return &PrepareStagesPhase{}
}

type PrepareStagesPhase struct{}

func (p *PrepareStagesPhase) Run(c *Conveyor) error {
	return logger.LogProcess("Preparing stages build instructions", logger.LogProcessOptions{}, func() error {
		return p.run(c)
	})
}

func (p *PrepareStagesPhase) run(c *Conveyor) (err error) {
	if debug() {
		fmt.Fprintf(logger.GetOutStream(), "PrepareStagesPhase.Run\n")
	}

	for _, image := range c.imagesInOrder {
		if err := logger.LogProcess(image.LogProcessName(), logger.LogProcessOptions{ColorizeMsgFunc: image.LogProcessColorizeFunc()}, func() error {
			return p.runImage(image, c)
		}); err != nil {
			return err
		}
	}

	return nil
}

func (p *PrepareStagesPhase) runImage(image *Image, c *Conveyor) (err error) {
	if debug() {
		fmt.Fprintf(logger.GetOutStream(), "  image: '%s'\n", image.GetName())
	}

	var prevImage, prevBuiltImage imagePkg.ImageInterface

	err = image.PrepareBaseImage(c)
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
			fmt.Fprintf(logger.GetOutStream(), "    %s\n", s.Name())
		}

		imageServiceCommitChangeOptions := stageImage.Container().ServiceCommitChangeOptions()
		imageServiceCommitChangeOptions.AddLabel(map[string]string{
			imagePkg.WerfDockerImageName:   stageImage.Name(),
			imagePkg.WerfLabel:             c.projectName(),
			imagePkg.WerfVersionLabel:      werf.Version,
			imagePkg.WerfCacheVersionLabel: BuildCacheVersion,
			imagePkg.WerfImageLabel:        "false",
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

		prevImage = stageImage
	}

	return
}
