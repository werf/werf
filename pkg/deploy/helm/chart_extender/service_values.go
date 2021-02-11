package chart_extender

import (
	"context"
	"fmt"
	"strings"

	"github.com/ghodss/yaml"

	"github.com/werf/logboek"
	imagePkg "github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/werf"
)

type ServiceValuesOptions struct {
	Namespace     string
	Env           string
	IsStub        bool
	CustomTagFunc func(string) string
}

func GetServiceValues(ctx context.Context, projectName string, repo string, imageInfoGetters []*imagePkg.InfoGetter, opts ServiceValuesOptions) (map[string]interface{}, error) {
	globalInfo := map[string]interface{}{
		"werf": map[string]interface{}{
			"name":    projectName,
			"version": werf.Version,
		},
		"env": opts.Env,
	}

	werfInfo := map[string]interface{}{
		"name":  projectName,
		"repo":  repo,
		"env":   opts.Env,
		"image": map[string]interface{}{},
	}

	if opts.CustomTagFunc != nil {
		werfInfo["image_digest"] = map[string]interface{}{}
	}

	if opts.Namespace != "" {
		werfInfo["namespace"] = opts.Namespace
	}

	if opts.IsStub {
		werfInfo["is_stub"] = true
		werfInfo["stub_image"] = fmt.Sprintf("%s:TAG", repo)
	}

	for _, imageInfoGetter := range imageInfoGetters {
		var image string
		var imageDigest string

		if opts.CustomTagFunc == nil {
			image = imageInfoGetter.GetName()
		} else {
			image = strings.Join([]string{repo, opts.CustomTagFunc(imageInfoGetter.GetName())}, ":")
			imageDigest = imageInfoGetter.GetTag()
		}

		if imageInfoGetter.IsNameless() {
			werfInfo["is_nameless_image"] = true
			werfInfo["nameless_image"] = image

			if imageDigest != "" {
				werfInfo["nameless_image_digest"] = imageInfoGetter.GetTag()
			}
		} else {
			werfInfo["image"].(map[string]interface{})[imageInfoGetter.GetWerfImageName()] = image

			if imageDigest != "" {
				werfInfo["image_digest"].(map[string]interface{})[imageInfoGetter.GetWerfImageName()] = imageDigest
			}
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
