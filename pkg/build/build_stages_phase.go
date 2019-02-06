package build

import (
	"fmt"
	"strings"

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
	return p.run(c)
}

func (p *BuildStagesPhase) run(c *Conveyor) error {
	if debug() {
		fmt.Fprintf(logger.GetOutStream(), "BuildStagesPhase.Run\n")
	}

	/*
	 * TODO: Build stages phase should push result into stagesRepo if non :local repo has been used
	 */

	images := c.imagesInOrder
	for ind, image := range images {
		isLastImage := ind == len(images)-1

		err := logger.LogServiceProcess(fmt.Sprintf("Build %s stages", image.LogName()), "", func() error {
			return p.runImage(image, c)
		})

		if err != nil {
			return err
		}

		if !isLastImage {
			fmt.Fprintln(logger.GetOutStream())
		}
	}

	return nil
}

func (p *BuildStagesPhase) runImage(image *Image, c *Conveyor) error {
	if debug() {
		fmt.Fprintf(logger.GetOutStream(), "  image: '%s'\n", image.GetName())
	}

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

		imageLockName := fmt.Sprintf("%s.image.%s", c.projectName(), img.Name())
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
	for ind, s := range stages {
		img := s.GetImage()
		msg := fmt.Sprintf("%s", s.Name())

		isLastStage := ind == len(stages)-1
		isUsingCache := img.IsExists()

		if isUsingCache {
			logger.LogServiceState(msg, "[USING CACHE]")

			logImageInfo(img, prevStageImageSize, isUsingCache)

			if !isLastStage {
				fmt.Fprintln(logger.GetOutStream())
			}

			prevStageImageSize = img.Inspect().Size

			continue
		}

		err := logger.LogProcess(msg, "[BUILDING]", func() error {
			if debug() {
				fmt.Fprintf(logger.GetOutStream(), "    %s\n", s.Name())
			}

			if err := s.PreRunHook(c); err != nil {
				return fmt.Errorf("stage '%s' preRunHook failed: %s", s.Name(), err)
			}

			if err := img.Build(p.ImageBuildOptions); err != nil {
				return fmt.Errorf("failed to build %s: %s", img.Name(), err)
			}

			if err := img.SaveInCache(); err != nil {
				return fmt.Errorf("failed to save in cache image %s: %s", img.Name(), err)
			}

			return nil
		})

		if err != nil {
			logger.WithLogIndent(func() error {
				logImageCommands(img)
				return nil
			})

			return err
		}

		logImageInfo(img, prevStageImageSize, isUsingCache)

		if !isLastStage {
			fmt.Fprintln(logger.GetOutStream())
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
	logger.LogInfoF(logImageInfoFormat, "image", img.Name())

	logger.LogInfoF(logImageInfoFormat, "created", img.Inspect().Created)

	if prevStageImageSize == 0 {
		logger.LogInfoF(logImageInfoFormat, "size", byteCountBinary(img.Inspect().Size))
	} else {
		logger.LogInfoF(logImageInfoFormat, "diff", byteCountBinary(img.Inspect().Size-prevStageImageSize))
	}

	if !isUsingCache {
		changes := img.Container().CommitChanges()
		if len(changes) != 0 {
			formattedCommands := strings.TrimLeft(logger.FitTextWithIndent(strings.Join(changes, "\n"), logImageInfoLeftPartWidth+4), " ")
			logger.LogInfoF(logImageInfoFormat, "instructions", formattedCommands)
		}

		logImageCommands(img)
	}
}

func logImageCommands(img imagePkg.ImageInterface) {
	commands := img.Container().AllRunCommands()
	if len(commands) != 0 {
		formattedCommands := strings.TrimLeft(logger.FitTextWithIndent(strings.Join(commands, "\n"), logImageInfoLeftPartWidth+4), " ")
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
