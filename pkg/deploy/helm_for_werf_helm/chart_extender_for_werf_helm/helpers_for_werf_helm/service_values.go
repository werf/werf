package helpers_for_werf_helm

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/samber/lo"
	"sigs.k8s.io/yaml"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/image"
	"github.com/werf/werf/v2/pkg/werf"
)

func NewChartExtenderServiceValuesData() *ChartExtenderServiceValuesData {
	return &ChartExtenderServiceValuesData{ServiceValues: make(map[string]interface{})}
}

type ChartExtenderServiceValuesData struct {
	ServiceValues map[string]interface{}
}

func (d *ChartExtenderServiceValuesData) GetServiceValues() map[string]interface{} {
	return d.ServiceValues
}

func (d *ChartExtenderServiceValuesData) SetServiceValues(vals map[string]interface{}) {
	d.ServiceValues = vals
}

type ServiceValuesOptions struct {
	Namespace         string
	Env               string
	IsStub            bool
	StubImageNameList []string
	// disable env stub used in the werf-render command
	DisableEnvStub bool
	CommitHash     string
	CommitDate     *time.Time

	SetDockerConfigJsonValue bool
	DockerConfigPath         string
}

func GetEnvServiceValues(env string) map[string]interface{} {
	werfInfo := map[string]interface{}{"env": env}

	return map[string]interface{}{
		"werf": werfInfo,
		"global": map[string]interface{}{
			"werf": werfInfo,
			"env":  env,
		},
	}
}

func GetServiceValues(ctx context.Context, projectName, repo string, imageInfoGetters []*image.InfoGetter, opts ServiceValuesOptions) (map[string]interface{}, error) {
	imagesInfo := map[string]interface{}{}

	werfInfo := map[string]interface{}{
		"name":    projectName,
		"version": werf.Version,
		"repo":    repo,
		"commit": map[string]interface{}{
			"hash": opts.CommitHash,
			"date": map[string]interface{}{
				"human": opts.CommitDate.String(),
				"unix":  opts.CommitDate.Unix(),
			},
		},
	}

	legacyImageInfo := map[string]interface{}{
		"image": map[string]interface{}{},
		"tag":   map[string]interface{}{},
	}

	if opts.Namespace != "" {
		werfInfo["namespace"] = opts.Namespace
	}

	if opts.IsStub {
		stubTag := "TAG"
		stubImage := fmt.Sprintf("%s:%s", repo, stubTag)

		werfInfo["is_stub"] = true
		werfInfo["stub_image"] = stubImage
		for _, name := range opts.StubImageNameList {
			legacyImageInfo["image"].(map[string]interface{})[name] = stubImage
			legacyImageInfo["tag"].(map[string]interface{})[name] = stubTag
			imagesInfo[name] = image.BuildStubImageValuesMap(repo, stubTag)
		}
	}

	for _, imageInfoGetter := range imageInfoGetters {
		tag := imageInfoGetter.GetTag()
		imageName := imageInfoGetter.GetName()

		legacyImageInfo["image"].(map[string]interface{})[imageInfoGetter.GetWerfImageName()] = imageName
		legacyImageInfo["tag"].(map[string]interface{})[imageInfoGetter.GetWerfImageName()] = tag

		imageDetails, err := image.BuildImageValuesMap(imageInfoGetter)
		if err != nil {
			return nil, fmt.Errorf("get image details for %q: %w", imageName, err)
		}

		imagesInfo[imageInfoGetter.GetWerfImageName()] = imageDetails
	}

	globalRes := map[string]interface{}{}

	if opts.Env != "" {
		globalRes["env"] = opts.Env
		werfInfo["env"] = opts.Env
	} else if opts.IsStub && !opts.DisableEnvStub {
		globalRes["env"] = ""
		werfInfo["env"] = ""
	}

	globalRes["werf"] = lo.Assign(werfInfo, map[string]interface{}{
		"images": imagesInfo,
	})

	res := map[string]interface{}{
		"werf": lo.Assign(werfInfo, legacyImageInfo),
	}

	if exposeGlobalServiceValues() {
		res["global"] = globalRes
	}

	if opts.SetDockerConfigJsonValue {
		if err := writeDockerConfigJsonValue(ctx, res, opts.DockerConfigPath); err != nil {
			return nil, fmt.Errorf("error writing docker config value: %w", err)
		}
	}

	data, err := yaml.Marshal(res)
	logboek.Context(ctx).Debug().LogF("GetServiceValues result (err=%s):\n%s\n", err, data)

	return res, nil
}

func GetBundleServiceValues(ctx context.Context, opts ServiceValuesOptions) (map[string]interface{}, error) {
	werfInfo := map[string]interface{}{
		"version": werf.Version,
	}

	globalInfo := map[string]interface{}{
		"werf": werfInfo,
	}

	if opts.Env != "" {
		globalInfo["env"] = opts.Env
		werfInfo["env"] = opts.Env
	}

	if opts.Namespace != "" {
		werfInfo["namespace"] = opts.Namespace
	}

	res := map[string]interface{}{
		"werf":   werfInfo,
		"global": globalInfo,
	}

	if opts.SetDockerConfigJsonValue {
		if err := writeDockerConfigJsonValue(ctx, res, opts.DockerConfigPath); err != nil {
			return nil, fmt.Errorf("error writing docker config value: %w", err)
		}
	}

	data, err := yaml.Marshal(res)
	logboek.Context(ctx).Debug().LogF("GetBundleServiceValues result (err=%s):\n%s\n", err, data)

	return res, nil
}

func writeDockerConfigJsonValue(ctx context.Context, values map[string]interface{}, dockerConfigPath string) error {
	if dockerConfigPath == "" {
		dockerConfigPath = filepath.Join(os.Getenv("HOME"), ".docker")
	}
	configJsonPath := filepath.Join(dockerConfigPath, "config.json")

	if _, err := os.Stat(configJsonPath); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("error accessing %q: %w", configJsonPath, err)
	}

	if data, err := ioutil.ReadFile(configJsonPath); err != nil {
		return fmt.Errorf("error reading %q: %w", configJsonPath, err)
	} else {
		values["dockerconfigjson"] = base64.StdEncoding.EncodeToString(data)
	}

	logboek.Context(ctx).Default().LogF("NOTE: ### --set-docker-config-json-value option has been specified ###\n")
	logboek.Context(ctx).Default().LogF("NOTE:\n")
	logboek.Context(ctx).Default().LogF("NOTE: Werf sets .Values.dockerconfigjson with the current docker config content %q with --set-docker-config-json-value option.\n", configJsonPath)
	logboek.Context(ctx).Default().LogF("NOTE: This docker config may contain temporal login credentials created using temporal short-lived token (CI_JOB_TOKEN for example),\n")
	logboek.Context(ctx).Default().LogF("NOTE: and in such case should not be used as imagePullSecrets.\n")

	return nil
}

// TODO(3.0): remove global service values completely
func exposeGlobalServiceValues() bool {
	return !util.GetBoolEnvironmentDefaultFalse("WERF_EXPERIMENT_NO_GLOBAL_SERVICE_VALUES")
}
