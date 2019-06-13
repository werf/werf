package build

import (
	"fmt"
	"github.com/flant/werf/pkg/lock"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/docker"
	imagePkg "github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/tag_strategy"
)

func NewBuildAndPublishImagesFromDockerfilePhase(imagesRepo string, opts PublishImagesOptions) *BuildAndPublishImagesFromDockerfilePhase {
	tagsByScheme := map[tag_strategy.TagStrategy][]string{
		tag_strategy.Custom:    opts.CustomTags,
		tag_strategy.GitBranch: opts.TagsByGitBranch,
		tag_strategy.GitTag:    opts.TagsByGitTag,
		tag_strategy.GitCommit: opts.TagsByGitCommit,
	}
	return &BuildAndPublishImagesFromDockerfilePhase{ImagesRepo: imagesRepo, TagsByScheme: tagsByScheme}
}

type BuildAndPublishImagesFromDockerfilePhase struct {
	ImagesRepo   string
	TagsByScheme map[tag_strategy.TagStrategy][]string
}

func (p *BuildAndPublishImagesFromDockerfilePhase) Run(c *Conveyor) error {
	if len(c.imageFromDockerfile) == 0 {
		return nil
	}

	logProcessOptions := logboek.LogProcessOptions{ColorizeMsgFunc: logboek.ColorizeHighlight}
	return logboek.LogProcess("Building and publishing images from dockerfile", logProcessOptions, func() error {
		return p.run(c)
	})
}

func (p *BuildAndPublishImagesFromDockerfilePhase) run(c *Conveyor) error {
	// TODO: Push stages should occur on the BuildStagesPhase

	for _, image := range c.imageFromDockerfile {
		logProcessOptions := logboek.LogProcessOptions{ColorizeMsgFunc: image.LogProcessColorizeFunc()}
		if err := logboek.LogProcess(image.LogDetailedName(), logProcessOptions, func() error {
			return p.runImage(c, image)
		}); err != nil {
			return err
		}
	}

	return nil
}

func (p *BuildAndPublishImagesFromDockerfilePhase) runImage(c *Conveyor, image *ImageFromDockerfile) error {
	var imageRepository string
	if image.GetName() != "" {
		imageRepository = fmt.Sprintf("%s/%s", p.ImagesRepo, image.GetName())
	} else {
		imageRepository = p.ImagesRepo
	}

	var stageImageTag string
	if image.GetName() == "" {
		stageImageTag = "latest"
	} else {
		stageImageTag = image.GetName() + "-latest"
	}

	stageImageName := fmt.Sprintf(LocalImageStageImageFormat, c.projectName(), stageImageTag)

	imageLockName := imagePkg.ImageLockName(stageImageName)
	if err := lock.Lock(imageLockName, lock.LockOptions{}); err != nil {
		return fmt.Errorf("failed to lock %s: %s", imageLockName, err)
	}

	defer lock.Unlock(imageLockName)

	logProcessOptions := logboek.LogProcessOptions{ColorizeMsgFunc: logboek.ColorizeHighlight}
	if err := logboek.LogProcess(fmt.Sprintf("Building image %s", image.GetName()), logProcessOptions, func() error {
		var buildArgs []string
		buildArgs = append(buildArgs, fmt.Sprintf("--tag=%s", stageImageName))
		buildArgs = append(buildArgs, image.DockerBuildArgs()...)
		return docker.CliBuild(buildArgs...)
	}); err != nil {
		return err
	}

	var nonEmptySchemeInOrder []tag_strategy.TagStrategy
	for strategy, tags := range p.TagsByScheme {
		if len(tags) == 0 {
			continue
		}

		nonEmptySchemeInOrder = append(nonEmptySchemeInOrder, strategy)
	}

	for _, strategy := range nonEmptySchemeInOrder {
		tags := p.TagsByScheme[strategy]

		if len(tags) == 0 {
			continue
		}

		logProcessOptions := logboek.LogProcessOptions{ColorizeMsgFunc: logboek.ColorizeHighlight}
		err := logboek.LogProcess(fmt.Sprintf("%s tagging strategy", string(strategy)), logProcessOptions, func() error {
			for _, tag := range tags {
				logProcessOptions := logboek.LogProcessOptions{ColorizeMsgFunc: logboek.ColorizeHighlight}
				if err := logboek.LogProcess(fmt.Sprintf("Publishing tag %s", tag), logProcessOptions, func() error {
					var buildArgs []string

					imageName := fmt.Sprintf("%s:%s", imageRepository, tag)

					if err := logboek.LogProcess("Building and tagging final image with meta information", logboek.LogProcessOptions{}, func() error {
						buildArgs = append(buildArgs, fmt.Sprintf("--label=%s=%s", imagePkg.WerfDockerImageName, imageName))
						buildArgs = append(buildArgs, fmt.Sprintf("--label=%s=%s", imagePkg.WerfTagStrategyLabel, string(strategy)))
						buildArgs = append(buildArgs, fmt.Sprintf("--label=%s=%s", imagePkg.WerfImageLabel, "true"))
						buildArgs = append(buildArgs, fmt.Sprintf("--tag=%s", imageName))
						buildArgs = append(buildArgs, image.DockerBuildArgs()...)

						return docker.CliBuild(buildArgs...)
					}); err != nil {
						return err
					}

					if err := logboek.LogProcess(fmt.Sprintf("Pushing %s", imageName), logboek.LogProcessOptions{}, func() error {
						return docker.CliPush(imageName)
					}); err != nil {
						return err
					}

					if err := logboek.LogProcess(fmt.Sprintf("Untagging %s", imageName), logboek.LogProcessOptions{}, func() error {
						return docker.CliRmi(imageName)
					}); err != nil {
						return err
					}

					return nil
				}); err != nil {
					return err
				}
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}
