package werf_chart

import (
	"context"
	"fmt"

	"github.com/ghodss/yaml"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/image"
)

type ServiceValuesOptions struct {
	Env    string
	IsStub bool
}

func GetServiceValues(ctx context.Context, projectName string, repo, namespace string, imageInfoGetters []*image.InfoGetter, opts ServiceValuesOptions) (map[string]interface{}, error) {
	res := make(map[string]interface{})

	werfInfo := map[string]interface{}{
		"name":    projectName,
		"repo":    repo,
		"is_stub": opts.IsStub,
	}

	if opts.IsStub {
		werfInfo["stub_image"] = fmt.Sprintf("%s:TAG", repo)
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

	for _, imageInfoGetter := range imageInfoGetters {
		imageData := make(map[string]interface{})

		if imageInfoGetter.IsNameless() {
			werfInfo["is_nameless_image"] = true
			werfInfo["image"] = imageData
		} else {
			werfInfo["is_nameless_image"] = false
			imagesInfo[imageInfoGetter.GetWerfImageName()] = imageData
		}

		imageData["docker_image"] = imageInfoGetter.GetName()
		imageData["docker_tag"] = imageInfoGetter.GetTag()
	}

	data, err := yaml.Marshal(res)
	logboek.Context(ctx).Debug().LogF("GetServiceValues result (err=%s):\n%s\n", err, data)

	return res, nil
}
