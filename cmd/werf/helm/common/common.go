package common

import (
	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/pkg/tag_strategy"
)

func GetImagesRepoOrStub(imagesRepoOption string) string {
	if imagesRepoOption == "" {
		return "IMAGES_REPO"
	}
	return imagesRepoOption
}

func GetEnvironmentOrStub(environmentOption string) string {
	if environmentOption == "" {
		return "ENV"
	}
	return environmentOption
}

func GetTagOrStub(commonCmdData *common.CmdData) (string, tag_strategy.TagStrategy, error) {
	tag, tagStrategy, err := common.GetDeployTag(commonCmdData, common.TagOptionsGetterOptions{Optional: true})
	if err != nil {
		return "", "", err
	}

	if tag == "" {
		tag, tagStrategy = "TAG", tag_strategy.Custom
	}

	return tag, tagStrategy, nil
}
