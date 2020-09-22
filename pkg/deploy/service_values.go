package deploy

import (
	"context"

	"github.com/ghodss/yaml"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/images_manager"
)

type ServiceValuesOptions struct {
	Env string
}

func GetServiceValues(ctx context.Context, projectName string, imagesRepository, namespace string, images []images_manager.ImageInfoGetter, opts ServiceValuesOptions) (map[string]interface{}, error) {
	res := make(map[string]interface{})

	werfInfo := map[string]interface{}{
		"name": projectName,
		"repo": imagesRepository,
	}

	globalInfo := map[string]interface{}{
		"namespace": namespace,
		"werf":      werfInfo,
	}
	if opts.Env != "" {
		globalInfo["env"] = opts.Env
	}
	res["global"] = globalInfo

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
	}

	data, err := yaml.Marshal(res)
	logboek.Context(ctx).Debug().LogF("GetServiceValues result (err=%s):\n%s\n", err, data)

	return res, nil
}
