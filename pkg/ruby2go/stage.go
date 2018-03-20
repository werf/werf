package ruby2go

import (
	"github.com/flant/dapp/pkg/image"
)

type Stage struct {
	From                 *Stage                      `json:"from"`
	Name                 string                      `json:"name"`
	ContainerName        string                      `json:"container_name"`
	BuildId              string                      `json:"built_id"`
	BashCommands         []string                    `json:"bash_commands"`
	PreparedBashCommand  string                      `json:"prepared_bash_command"`
	ServiceBashCommands  []string                    `json:"service_bash_commands"`
	Options              image.StageContainerOptions `json:"options"`
	ServiceOptions       image.StageContainerOptions `json:"service_options"`
	ChangeOptions        image.StageContainerOptions `json:"change_options"`
	ServiceChangeOptions image.StageContainerOptions `json:"service_change_options"`
	ImageInspect         interface{}                 `json:"image_inspect"`
	BuiltImageInspect    interface{}                 `json:"built_image_inspect"`
}

func rubyStageToImageStage(rubyStage *Stage) *image.Stage {
	var from *image.Stage
	if rubyStage.From != nil {
		from = rubyStageToImageStage(rubyStage.From)
	}

	stageImage := image.NewStageImage(from, rubyStage.Name, rubyStage.BuildId)
	stageImage.Container.Name = rubyStage.ContainerName
	stageImage.Container.RunCommands = rubyStage.BashCommands
	stageImage.Container.ServiceRunCommands = rubyStage.ServiceBashCommands
	stageImage.Container.PreparedRunCommand = rubyStage.PreparedBashCommand
	stageImage.Container.RunOptions = &rubyStage.Options
	stageImage.Container.ServiceRunOptions = &rubyStage.ServiceOptions
	stageImage.Container.CommitChangeOptions = &rubyStage.ChangeOptions
	stageImage.Container.ServiceCommitChangeOptions = &rubyStage.ServiceChangeOptions
	return stageImage
}

func imageStageToRubyStage(imageStage *image.Stage) *Stage {
	rubyImage := &Stage{}

	if imageStage.From != nil {
		rubyImage.From = imageStageToRubyStage(imageStage.From)
	}

	rubyImage.Name = imageStage.Name
	rubyImage.BuildId = imageStage.BuiltId
	rubyImage.ContainerName = imageStage.Container.Name
	rubyImage.BashCommands = imageStage.Container.RunCommands
	rubyImage.ServiceBashCommands = imageStage.Container.ServiceRunCommands
	rubyImage.PreparedBashCommand = imageStage.Container.PreparedRunCommand
	rubyImage.Options = *imageStage.Container.RunOptions
	rubyImage.ServiceOptions = *imageStage.Container.ServiceRunOptions
	rubyImage.ChangeOptions = *imageStage.Container.CommitChangeOptions
	rubyImage.ServiceChangeOptions = *imageStage.Container.ServiceCommitChangeOptions
	rubyImage.ImageInspect = imageStage.Inspect
	rubyImage.BuiltImageInspect = imageStage.BuiltInspect
	return rubyImage
}
