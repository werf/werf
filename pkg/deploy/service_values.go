package deploy

import (
	"github.com/ghodss/yaml"

	"github.com/flant/logboek"

	"github.com/werf/werf/pkg/images_manager"
	"github.com/werf/werf/pkg/tag_strategy"
)

const (
	TemplateEmptyValue = "\"-\""
)

type ServiceValuesOptions struct {
	Env string
}

func GetServiceValues(projectName string, imagesRepository, namespace, commonTag string, tagStrategy tag_strategy.TagStrategy, images []images_manager.ImageInfoGetter, opts ServiceValuesOptions) (map[string]interface{}, error) {
	res := make(map[string]interface{})

	ciInfo := map[string]interface{}{
		"is_tag":    false,
		"is_branch": false,
		"is_custom": false,
		"branch":    TemplateEmptyValue,
		"tag":       TemplateEmptyValue,
		"ref":       TemplateEmptyValue,
	}

	werfInfo := map[string]interface{}{
		"name": projectName,
		"repo": imagesRepository,
		"ci":   ciInfo,
	}
	if commonTag != "" {
		werfInfo["docker_tag"] = commonTag
	}

	globalInfo := map[string]interface{}{
		"namespace": namespace,
		"werf":      werfInfo,
	}
	if opts.Env != "" {
		globalInfo["env"] = opts.Env
	}
	res["global"] = globalInfo

	switch tagStrategy {
	case tag_strategy.GitTag:
		ciInfo["tag"] = commonTag
		ciInfo["ref"] = commonTag
		ciInfo["is_tag"] = true

	case tag_strategy.GitBranch:
		ciInfo["branch"] = commonTag
		ciInfo["ref"] = commonTag
		ciInfo["is_branch"] = true

	case tag_strategy.Custom:
		ciInfo["is_custom_tag"] = true
	case tag_strategy.StagesSignature:
		ciInfo["is_tag_by_stages_signatures"] = true
	}

	imagesInfo := make(map[string]interface{})
	werfInfo["image"] = imagesInfo

	for _, image := range images {
		imageData := make(map[string]interface{})

		if image.IsNameless() {
			werfInfo["is_nameless_image"] = true
			werfInfo["image"] = imageData
		} else {
			werfInfo["is_nameless_image"] = false
			imagesInfo[image.GetName()] = imageData
		}

		imageData["docker_image"] = image.GetImageName()
		imageData["docker_tag"] = image.GetImageTag()

		if tagStrategy == tag_strategy.GitBranch || tagStrategy == tag_strategy.Custom {
			setKey := func(key, value string) {
				if value == "" {
					imageData[key] = TemplateEmptyValue
				} else {
					imageData[key] = value
				}
				logboek.Debug.LogF("ServiceValues: %s.%s=%s", image.GetImageName(), key, value)
			}

			imageID, err := image.GetImageID()
			if err != nil {
				return nil, err
			}

			setKey("docker_image_id", imageID)

			imageDigest, err := image.GetImageDigest()
			if err != nil {
				return nil, err
			}

			setKey("docker_image_digest", imageDigest)
		}
	}

	data, err := yaml.Marshal(res)
	logboek.Debug.LogF("GetServiceValues result (err=%s):\n%s\n", err, data)

	return res, nil
}
