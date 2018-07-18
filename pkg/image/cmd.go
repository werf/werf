package image

import (
	"encoding/json"
	"fmt"

	"github.com/docker/docker/api/types"
)

type CmdStage struct {
	From                 *CmdStage             `json:"from"`
	Name                 string                `json:"name"`
	ContainerName        string                `json:"container_name"`
	BuildId              string                `json:"built_id"`
	BashCommands         []string              `json:"bash_commands"`
	ServiceBashCommands  []string              `json:"service_bash_commands"`
	Options              StageContainerOptions `json:"options"`
	ChangeOptions        StageContainerOptions `json:"change_options"`
	ServiceChangeOptions StageContainerOptions `json:"service_change_options"`
	ImageInspect         *types.ImageInspect   `json:"image_inspect"`
	BuiltImageInspect    *types.ImageInspect   `json:"built_image_inspect"`
}

func cmdStageToImageStage(cmdStage *CmdStage) *Stage {
	var from *Stage
	if cmdStage.From != nil {
		from = cmdStageToImageStage(cmdStage.From)
	}

	stageImage := NewStageImage(from, cmdStage.Name, cmdStage.BuildId)
	stageImage.Inspect = cmdStage.ImageInspect
	stageImage.BuiltInspect = cmdStage.BuiltImageInspect
	stageImage.Container.RunCommands = cmdStage.BashCommands
	stageImage.Container.ServiceRunCommands = cmdStage.ServiceBashCommands
	stageImage.Container.RunOptions = &cmdStage.Options
	stageImage.Container.CommitChangeOptions = &cmdStage.ChangeOptions
	stageImage.Container.ServiceCommitChangeOptions = &cmdStage.ServiceChangeOptions
	return stageImage
}

func imageStageToRubyStage(imageStage *Stage) *CmdStage {
	cmdImage := &CmdStage{}

	if imageStage.From != nil {
		cmdImage.From = imageStageToRubyStage(imageStage.From)
	}

	cmdImage.Name = imageStage.Name
	cmdImage.BuildId = imageStage.BuiltId
	cmdImage.ContainerName = imageStage.Container.Name
	cmdImage.BashCommands = imageStage.Container.RunCommands
	cmdImage.ServiceBashCommands = imageStage.Container.ServiceRunCommands
	cmdImage.Options = *imageStage.Container.RunOptions
	cmdImage.ChangeOptions = *imageStage.Container.CommitChangeOptions
	cmdImage.ServiceChangeOptions = *imageStage.Container.ServiceCommitChangeOptions
	cmdImage.ImageInspect = imageStage.Inspect
	cmdImage.BuiltImageInspect = imageStage.BuiltInspect
	return cmdImage
}

func CommandWithImage(args map[string]interface{}, command func(stageImage *Stage) error) (map[string]interface{}, error) {
	stageImage, err := stageImageFromArgs(args)
	if err != nil {
		return nil, err
	}

	if err := command(stageImage); err != nil {
		return nil, err
	}

	resultMap, err := stageImageToArgs(stageImage, make(map[string]interface{}))
	if err != nil {
		return nil, err
	}

	return resultMap, nil
}

func stageImageFromArgs(args map[string]interface{}) (*Stage, error) {
	cmdImage := &CmdStage{}

	switch args["image"].(type) {
	case string:
		if err := json.Unmarshal([]byte(args["image"].(string)), cmdImage); err != nil {
			return nil, fmt.Errorf("image field unmarshaling failed: %s", err.Error())
		}
		return cmdStageToImageStage(cmdImage), nil
	default:
		return nil, fmt.Errorf("image field value `%v` isn't supported", args["image"])
	}
}

func stageImageToArgs(stageImage *Stage, args map[string]interface{}) (map[string]interface{}, error) {
	raw, err := json.Marshal(imageStageToRubyStage(stageImage))
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("stage marshaling failed: %s", err.Error()))
	}
	args["image"] = string(raw)
	return args, nil
}
