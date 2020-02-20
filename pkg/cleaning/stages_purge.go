package cleaning

import (
	"github.com/flant/logboek"
)

type StagesPurgeOptions struct {
	ProjectName                   string
	DryRun                        bool
	RmContainersThatUseWerfImages bool
}

func StagesPurge(options StagesPurgeOptions) error {
	return logboek.Default.LogProcess(
		"Running stages purge",
		logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()},
		func() error {
			return stagesPurge(options)
		},
	)
}

func stagesPurge(options StagesPurgeOptions) error {
	var commonProjectOptions CommonProjectOptions
	commonProjectOptions.ProjectName = options.ProjectName
	commonProjectOptions.CommonOptions = CommonOptions{
		RmiForce:                      true,
		RmForce:                       options.RmContainersThatUseWerfImages,
		RmContainersThatUseWerfImages: options.RmContainersThatUseWerfImages,
		DryRun:                        options.DryRun,
	}

	if err := projectStagesPurge(commonProjectOptions); err != nil {
		return err
	}

	return nil
}

func projectStagesPurge(options CommonProjectOptions) error {
	if err := werfImagesFlushByFilterSet(projectImageStageFilterSet(options), options.CommonOptions); err != nil {
		return err
	}

	return nil
}
