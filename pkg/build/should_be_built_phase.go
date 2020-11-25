package build

import (
	"fmt"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/build/stage"
)

type ShouldBeBuiltPhase struct{}

func NewShouldBeBuiltPhase() *ShouldBeBuiltPhase {
	return &ShouldBeBuiltPhase{}
}

func (p *ShouldBeBuiltPhase) Run(c *Conveyor) error {
	logProcessOptions := logboek.LogProcessOptions{ColorizeMsgFunc: logboek.ColorizeHighlight}
	return logboek.LogProcess("Checking built stages cache", logProcessOptions, func() error {
		return p.run(c)
	})
}

func (p *ShouldBeBuiltPhase) run(c *Conveyor) error {
	var isBadImageExist bool
	var isBadDockerfileImageExist bool

	for _, image := range c.imagesInOrder {
		var badStages []stage.Interface

		for _, s := range image.GetStages() {
			i := s.GetImage()
			if i.IsExists() {
				continue
			}

			isBadImageExist = true
			if !isBadDockerfileImageExist {
				isBadDockerfileImageExist = image.isDockerfileImage
			}

			badStages = append(badStages, s)
		}

		for _, s := range badStages {
			logboek.LogErrorF("%s %s is not exist in stages storage\n", image.LogDetailedName(), s.LogDetailedName())
		}
	}

	if isBadImageExist {
		var reasonNumber int
		reasonNumberFunc := func() string {
			reasonNumber++
			return fmt.Sprintf("(%d) ", reasonNumber)
		}

		logboek.LogErrorLn()
		logboek.LogErrorLn(`There are some possible reasons:`)
		logboek.LogErrorLn()

		if isBadDockerfileImageExist {
			logboek.LogErrorLn(reasonNumberFunc() + `Dockerfile has COPY or ADD instruction which uses non-permanent data that affects stage signature: 
- .git directory which should be excluded with .dockerignore file (https://docs.docker.com/engine/reference/builder/#dockerignore-file)
- auto-generated file`)
			logboek.LogErrorLn()
		}

		logboek.LogErrorLn(reasonNumberFunc() + `werf.yaml has non-permanent data that affects stage signature:
- environment variable (e.g. {{ env "JOB_ID" }})
- dynamic go template function (e.g. one of sprig date functions http://masterminds.github.io/sprig/date.html)
- auto-generated file content (e.g. {{ .Files.Get "hash_sum_of_something" }})`)
		logboek.LogErrorLn()

		logboek.LogErrorLn(`To quickly find the problem compare current and previous rendered werf configurations.
Get the path at the beginning of command output by the following prefix 'Using werf config render file: '.
E.g.:

  diff /tmp/werf-config-render-502883762 /tmp/werf-config-render-837625028`)
		logboek.LogErrorLn()

		logboek.LogErrorLn(reasonNumberFunc() + `Stages have not been built yet or stages have been removed:
- automatically with werf cleanup command
- manually with werf purge, werf stages purge or werf host purge commands`)
		logboek.LogErrorLn()

		return fmt.Errorf("stages required")
	}

	return nil
}
