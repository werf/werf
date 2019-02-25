package build

import (
	"fmt"
	"strings"

	"github.com/docker/docker/pkg/stringid"

	imagePkg "github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/lock"
	"github.com/flant/werf/pkg/logger"
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
	return logger.LogProcess("Building stages", logger.LogProcessOptions{}, func() error {
		return p.run(c)
	})
}

func (p *BuildStagesPhase) run(c *Conveyor) error {
	/*
	 * TODO: Build stages phase should push result into stagesRepo if non :local repo has been used
	 */

	images := c.imagesInOrder
	for _, image := range images {
		if err := logger.LogProcess(image.LogProcessName(), logger.LogProcessOptions{ColorizeMsgFunc: image.LogProcessColorizeFunc()}, func() error {
			return p.runImage(image, c)
		}); err != nil {
			return err
		}
	}

	return nil
}

func (p *BuildStagesPhase) runImage(image *Image, c *Conveyor) error {
	var acquiredLocks []string

	unlockLock := func() {
		var lockName string
		lockName, acquiredLocks = acquiredLocks[0], acquiredLocks[1:]
		lock.Unlock(lockName)
	}

	unlockLocks := func() {
		for len(acquiredLocks) > 0 {
			unlockLock()
		}
	}

	defer unlockLocks()

	// lock
	for _, stage := range image.GetStages() {
		img := stage.GetImage()
		if img.IsExists() {
			continue
		}

		imageLockName := imagePkg.GetImageLockName(img.Name())

		if err := lock.Lock(imageLockName, lock.LockOptions{}); err != nil {
			return fmt.Errorf("failed to lock %s: %s", imageLockName, err)
		}

		acquiredLocks = append(acquiredLocks, imageLockName)

		if err := img.SyncDockerState(); err != nil {
			return err
		}
	}

	// build
	stages := image.GetStages()
	var prevStageImageSize int64
	for _, s := range stages {
		img := s.GetImage()
		stageLogName := fmt.Sprintf("stage %s", s.Name())

		isUsingCache := img.IsExists()

		if isUsingCache {
			logger.LogHighlightLn(stageLogName)

			logImageInfo(img, prevStageImageSize, isUsingCache)

			logger.OptionalLnModeOn()

			prevStageImageSize = img.Inspect().Size

			continue
		}

		infoSectionFunc := func(err error) {
			if err != nil {
				_ = logger.WithIndent(func() error {
					logImageCommands(img)
					return nil
				})

				return
			}

			logImageInfo(img, prevStageImageSize, isUsingCache)
		}

		logProcessOptions := logger.LogProcessOptions{InfoSectionFunc: infoSectionFunc}
		err := logger.LogProcess(fmt.Sprintf("Building %s", stageLogName), logProcessOptions, func() (err error) {
			if err := s.PreRunHook(c); err != nil {
				return fmt.Errorf("stage '%s' preRunHook failed: %s", s.Name(), err)
			}

			if err := logger.WithTag(fmt.Sprintf("%s/%s", image.LogName(), s.Name()), image.LogTagColorizeFunc(), func() error {
				if err := logger.WithFittedStreamsOutputOn(func() error {
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

		unlockLock()

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

	logger.LogInfoF(logImageInfoFormat, "repository", repository)
	logger.LogInfoF(logImageInfoFormat, "image_id", stringid.TruncateID(img.ID()))
	logger.LogInfoF(logImageInfoFormat, "created", img.Inspect().Created)
	logger.LogInfoF(logImageInfoFormat, "tag", tag)

	if prevStageImageSize == 0 {
		logger.LogInfoF(logImageInfoFormat, "size", byteCountBinary(img.Inspect().Size))
	} else {
		logger.LogInfoF(logImageInfoFormat, "diff", byteCountBinary(img.Inspect().Size-prevStageImageSize))
	}

	if !isUsingCache {
		changes := img.Container().UserCommitChanges()
		if len(changes) != 0 {
			fitTextOptions := logger.FitTextOptions{ExtraIndentWidth: logImageInfoLeftPartWidth + 4}
			formattedCommands := strings.TrimLeft(logger.FitText(strings.Join(changes, "\n"), fitTextOptions), " ")
			logger.LogInfoF(logImageInfoFormat, "instructions", formattedCommands)
		}

		logImageCommands(img)
	}
}

func logImageCommands(img imagePkg.ImageInterface) {
	commands := img.Container().UserRunCommands()
	if len(commands) != 0 {
		fitTextOptions := logger.FitTextOptions{ExtraIndentWidth: logImageInfoLeftPartWidth + 4}
		formattedCommands := strings.TrimLeft(logger.FitText(strings.Join(commands, "\n"), fitTextOptions), " ")
		logger.LogInfoF(logImageInfoFormat, "commands", formattedCommands)
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
