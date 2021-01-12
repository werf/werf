package werf_chart

import (
	"context"
	"fmt"

	"github.com/ghodss/yaml"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/werf"
)

type ServiceValuesOptions struct {
	Namespace string
	Env       string
	IsStub    bool
}

func GetServiceValues(ctx context.Context, projectName string, repo string, imageInfoGetters []*image.InfoGetter, opts ServiceValuesOptions) (map[string]interface{}, error) {
	globalInfo := map[string]interface{}{
		"werf": map[string]interface{}{
			"version": werf.Version,
		},
		"env": opts.Env,
	}

	werfInfo := map[string]interface{}{
		"name": projectName,
		"repo": repo,
		"env":  opts.Env,
	}

	if opts.Namespace != "" {
		werfInfo["namespace"] = opts.Namespace
	}

	if opts.IsStub {
		werfInfo["is_stub"] = true
		werfInfo["stub_image"] = fmt.Sprintf("%s:TAG", repo)
	}

	for _, imageInfoGetter := range imageInfoGetters {
		if imageInfoGetter.IsNameless() {
			werfInfo["is_nameless_image"] = true
			werfInfo["image"] = imageInfoGetter.GetName()
		} else {
			if werfInfo["image"] == nil {
				werfInfo["image"] = map[string]interface{}{}
			}
			werfInfo["image"].(map[string]interface{})[imageInfoGetter.GetWerfImageName()] = imageInfoGetter.GetName()
		}
	}

	res := map[string]interface{}{
		"werf":   werfInfo,
		"global": globalInfo,
	}

	data, err := yaml.Marshal(res)
	logboek.Context(ctx).Debug().LogF("GetServiceValues result (err=%s):\n%s\n", err, data)

	return res, nil
}
