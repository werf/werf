package build

import (
	"fmt"
	"strings"

	"github.com/docker/docker/pkg/stringid"

	"github.com/flant/logboek"
	imagePkg "github.com/flant/werf/pkg/image"
)

func NewBuildStagesPhase(stagesRepo string, opts BuildStagesOptions) *BuildStagesPhase {
	return &BuildStagesPhase{StagesRepo: stagesRepo, BuildStagesOptions: opts}
}

type BuildStagesOptions struct {
	ImageBuildOptions imagePkg.BuildOptions
}

type BuildStagesPhase struct {
	StagesRepo string
	BuildStagesOptions
}

func (p *BuildStagesPhase) Run(c *Conveyor) (err error) {
	logProcessOptions := logboek.LogProcessOptions{ColorizeMsgFunc: logboek.ColorizeHighlight}
	return logboek.LogProcess("Building stages", logProcessOptions, func() error {
		return p.run(c)
	})
}

func (p *BuildStagesPhase) run(c *Conveyor) error {
	/*
	 * TODO: Build stages phase should push result into stagesRepo if non :local repo has been used
	 */

	images := c.imagesInOrder
	for _, image := range images {
		if err := logboek.LogProcess(image.LogDetailedName(), logboek.LogProcessOptions{ColorizeMsgFunc: image.LogProcessColorizeFunc()}, func() error {
			return p.runImage(image, c)
		}); err != nil {
			return err
		}
	}

	return nil
}

func (p *BuildStagesPhase) runImage(image *Image, c *Conveyor) error {
	stages := image.GetStages()

	var prevStageImageSize int64

	for _, s := range stages {
		img := s.GetImage()

		isUsingCache := img.IsExists()

		if isUsingCache {
			logboek.LogHighlightF("Use cache image for %s\n", s.LogDetailedName())

			logImageInfo(img, prevStageImageSize, isUsingCache)

			logboek.LogOptionalLn()

			prevStageImageSize = img.Inspect().Size

			continue
		}

		infoSectionFunc := func(err error) {
			if err != nil {
				_ = logboek.WithIndent(func() error {
					logImageCommands(img)
					return nil
				})

				return
			}

			logImageInfo(img, prevStageImageSize, isUsingCache)
		}

		logProcessOptions := logboek.LogProcessOptions{InfoSectionFunc: infoSectionFunc, ColorizeMsgFunc: logboek.ColorizeHighlight}
		err := logboek.LogProcess(fmt.Sprintf("Building %s", s.LogDetailedName()), logProcessOptions, func() (err error) {
			if err := s.PreRunHook(c); err != nil {
				return fmt.Errorf("%s preRunHook failed: %s", s.LogDetailedName(), err)
			}

			if err := logboek.WithTag(fmt.Sprintf("%s/%s", image.LogName(), s.Name()), image.LogTagColorizeFunc(), func() error {
				if err := logboek.WithFittedStreamsOutputOn(func() error {
					if err := img.Build(p.ImageBuildOptions); err != nil {
						return fmt.Errorf("failed to build %s: %s", img.Name(), err)
					}

					return nil
				}); err != nil {
					return err
				}

				return nil
			}); err != nil {
				return err
			}

			if err := img.SaveInCache(); err != nil {
				return fmt.Errorf("failed to save in cache image %s: %s", img.Name(), err)
			}

			return nil
		})

		if err != nil {
			return err
		}

		imageLockName := imagePkg.ImageLockName(img.Name())
		if err := c.ReleaseGlobalLock(imageLockName); err != nil {
			return fmt.Errorf("failed to unlock %s: %s", imageLockName, err)
		}

		prevStageImageSize = img.Inspect().Size
	}

	return nil
}

var (
	logImageInfoLeftPartWidth = 12
	logImageInfoFormat        = fmt.Sprintf("  %%%ds: %%s\n", logImageInfoLeftPartWidth)
)

func logImageInfo(img imagePkg.ImageInterface, prevStageImageSize int64, isUsingCache bool) {
	parts := strings.Split(img.Name(), ":")
	repository, tag := parts[0], parts[1]

	logboek.LogInfoF(logImageInfoFormat, "repository", repository)
	logboek.LogInfoF(logImageInfoFormat, "image_id", stringid.TruncateID(img.ID()))
	logboek.LogInfoF(logImageInfoFormat, "created", img.Inspect().Created)
	logboek.LogInfoF(logImageInfoFormat, "tag", tag)

	if prevStageImageSize == 0 {
		logboek.LogInfoF(logImageInfoFormat, "size", byteCountBinary(img.Inspect().Size))
	} else {
		logboek.LogInfoF(logImageInfoFormat, "diff", byteCountBinary(img.Inspect().Size-prevStageImageSize))
	}

	if !isUsingCache {
		changes := img.Container().UserCommitChanges()
		if len(changes) != 0 {
			fitTextOptions := logboek.FitTextOptions{ExtraIndentWidth: logImageInfoLeftPartWidth + 4}
			formattedCommands := strings.TrimLeft(logboek.FitText(strings.Join(changes, "\n"), fitTextOptions), " ")
			logboek.LogInfoF(logImageInfoFormat, "instructions", formattedCommands)
		}

		logImageCommands(img)
	}
}

func logImageCommands(img imagePkg.ImageInterface) {
	commands := img.Container().UserRunCommands()
	if len(commands) != 0 {
		fitTextOptions := logboek.FitTextOptions{ExtraIndentWidth: logImageInfoLeftPartWidth + 4}
		formattedCommands := strings.TrimLeft(logboek.FitText(strings.Join(commands, "\n"), fitTextOptions), " ")
		logboek.LogInfoF(logImageInfoFormat, "commands", formattedCommands)
	}
}

func byteCountBinary(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
