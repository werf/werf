package deploy

import (
	"fmt"

	"github.com/ghodss/yaml"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/tag_strategy"
)

const (
	TemplateEmptyValue = "\"-\""
)

type ImageInfoGetter interface {
	IsNameless() bool
	GetName() string
	GetImageName() string
	GetImageId() (string, error)
}

func GetImagesInfoGetters(configImages []*config.Image, imagesRepo, tag string, withoutRegistry bool) []ImageInfoGetter {
	var images []ImageInfoGetter

	for _, image := range configImages {
		d := &ImageInfo{Config: image, WithoutRegistry: withoutRegistry, ImagesRepo: imagesRepo, Tag: tag}
		images = append(images, d)
	}

	return images
}

type ServiceValuesOptions struct {
	Env string
}

func GetServiceValues(projectName, repo, namespace, tag string, tagStrategy tag_strategy.TagStrategy, images []ImageInfoGetter, opts ServiceValuesOptions) (map[string]interface{}, error) {
	res := make(map[string]interface{})

	ciInfo := map[string]interface{}{
		"is_tag":    false,
		"is_branch": false,
		"branch":    TemplateEmptyValue,
		"tag":       TemplateEmptyValue,
		"ref":       TemplateEmptyValue,
	}

	werfInfo := map[string]interface{}{
		"name":       projectName,
		"repo":       repo,
		"docker_tag": tag,
		"ci":         ciInfo,
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
		ciInfo["tag"] = tag
		ciInfo["ref"] = tag
		ciInfo["is_tag"] = true

	case tag_strategy.GitBranch:
		ciInfo["branch"] = tag
		ciInfo["ref"] = tag
		ciInfo["is_branch"] = true
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
		imageData["docker_image_id"] = TemplateEmptyValue

		imageID, err := image.GetImageId()
		if err != nil {
			return nil, err
		}

		if debug() {
			fmt.Fprintf(logboek.GetOutStream(), "GetServiceValues got image id of %s: %#v", image.GetImageName(), imageID)
		}

		var value string
		if imageID == "" {
			value = TemplateEmptyValue
		} else {
			value = imageID
		}
		imageData["docker_image_id"] = value
	}

	if debug() {
		data, err := yaml.Marshal(res)
		fmt.Fprintf(logboek.GetOutStream(), "GetServiceValues result (err=%s):\n%s\n", err, data)
	}

	return res, nil
}
