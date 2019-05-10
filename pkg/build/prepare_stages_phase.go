package build

import (
	"fmt"

	"github.com/flant/logboek"
	imagePkg "github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/werf"
)

func NewPrepareStagesPhase() *PrepareStagesPhase {
	return &PrepareStagesPhase{}
}

type PrepareStagesPhase struct{}

func (p *PrepareStagesPhase) Run(c *Conveyor) error {
	logProcessOptions := logboek.LogProcessOptions{ColorizeMsgFunc: logboek.ColorizeHighlight}
	return logboek.LogProcess("Preparing stages build instructions", logProcessOptions, func() error {
		return p.run(c)
	})
}

func (p *PrepareStagesPhase) run(c *Conveyor) (err error) {
	for _, image := range c.imagesInOrder {
		if err := logboek.LogProcess(image.LogDetailedName(), logboek.LogProcessOptions{ColorizeMsgFunc: image.LogProcessColorizeFunc()}, func() error {
			return p.runImage(image, c)
		}); err != nil {
			return err
		}
	}

	return nil
}

func (p *PrepareStagesPhase) runImage(image *Image, c *Conveyor) (err error) {
	var prevImage, prevBuiltImage imagePkg.ImageInterface

	err = image.PrepareBaseImage(c)
	if err != nil {
		return fmt.Errorf("prepare base image %s failed: %s", image.GetBaseImage().Name(), err)
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
			imageRunOptions.AddVolume(fmt.Sprintf("%s:/.werf/tmp/ssh-auth-sock", c.sshAuthSock))
			imageRunOptions.AddEnv(map[string]string{"SSH_AUTH_SOCK": "/.werf/tmp/ssh-auth-sock"})
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
