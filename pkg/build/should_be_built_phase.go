package build

import (
	"fmt"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/build/stage"
)

type ShouldBeBuiltPhase struct {
	BasePhase

	IsBadDockerfileImageExists bool
	BadImages                  []*Image
	BadStagesByImage           map[string][]stage.Interface
}

func NewShouldBeBuiltPhase(c *Conveyor) *ShouldBeBuiltPhase {
	return &ShouldBeBuiltPhase{BasePhase: BasePhase{c}, BadStagesByImage: make(map[string][]stage.Interface)}
}

func (phase *ShouldBeBuiltPhase) Name() string {
	return "shouldBeBuilt"
}

func (phase *ShouldBeBuiltPhase) BeforeImageStages(img *Image) error {
	return nil
}

func (phase *ShouldBeBuiltPhase) AfterImageStages(img *Image) error {
	if len(phase.BadStagesByImage[img.GetName()]) > 0 {
		phase.BadImages = append(phase.BadImages, img)

		if !phase.IsBadDockerfileImageExists {
			phase.IsBadDockerfileImageExists = img.isDockerfileImage
		}
	}

	return nil
}

func (phase *ShouldBeBuiltPhase) ImageProcessingShouldBeStopped(img *Image) bool {
	return len(phase.BadStagesByImage[img.GetName()]) > 0
}

func (phase *ShouldBeBuiltPhase) OnImageStage(img *Image, stg stage.Interface) (bool, error) {
	if !stg.GetImage().IsExists() {
		phase.BadStagesByImage[img.GetName()] = append(phase.BadStagesByImage[img.GetName()], stg)
	}
	return true, nil
}

func (phase *ShouldBeBuiltPhase) BeforeImages() error {
	return nil
}

func (phase *ShouldBeBuiltPhase) AfterImages() error {
	if len(phase.BadImages) == 0 {
		return nil
	}

	logProcessOptions := logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()}
	return logboek.Default.LogProcess("Built stages cache check", logProcessOptions, func() error {
		for _, img := range phase.BadImages {
			for _, stg := range phase.BadStagesByImage[img.GetName()] {
				logboek.LogWarnF("%s %s is not exist in stages storage\n", img.LogDetailedName(), stg.LogDetailedName())
			}
		}

		var reasonNumber int
		reasonNumberFunc := func() string {
			reasonNumber++
			return fmt.Sprintf("(%d) ", reasonNumber)
		}

		logboek.LogWarnLn()
		logboek.LogWarnLn("There are some possible reasons:")
		logboek.LogWarnLn()

		if phase.IsBadDockerfileImageExists {
			logboek.LogWarnLn(reasonNumberFunc() + `Dockerfile has COPY or ADD instruction which uses non-permanent data that affects stage signature:
- .git directory which should be excluded with .dockerignore file (https://docs.docker.com/engine/reference/builder/#dockerignore-file)
- auto-generated file`)
			logboek.LogWarnLn()
		}

		logboek.LogWarnLn(reasonNumberFunc() + `werf.yaml has non-permanent data that affects stage signature:
- environment variable (e.g. {{ env "JOB_ID" }})
- dynamic go template function (e.g. one of sprig date functions http://masterminds.github.io/sprig/date.html)
- auto-generated file content (e.g. {{ .Files.Get "hash_sum_of_something" }})`)
		logboek.LogWarnLn()

		logboek.LogWarnLn(`Stage signature dependencies can be found here, https://werf.io/documentation/reference/stages_and_images.html#stage-dependencies.

To quickly find the problem compare current and previous rendered werf configurations.
Get the path at the beginning of command output by the following prefix 'Using werf config render file: '.
E.g.:

  diff /tmp/werf-config-render-502883762 /tmp/werf-config-render-837625028`)
		logboek.LogWarnLn()

		logboek.LogWarnLn(reasonNumberFunc() + `Stages have not been built yet or stages have been removed:
- automatically with werf cleanup command
- manually with werf purge, werf stages purge or werf host purge commands`)
		logboek.LogWarnLn()

		return fmt.Errorf("stages required")
	})
}
