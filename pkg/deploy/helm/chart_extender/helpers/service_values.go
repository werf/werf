package helpers

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/werf"
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
	Namespace string
	Env       string
	IsStub    bool

	SetDockerConfigJsonValue bool
	DockerConfigPath         string
}

func GetServiceValues(ctx context.Context, projectName string, repo string, imageInfoGetters []*image.InfoGetter, opts ServiceValuesOptions) (map[string]interface{}, error) {
	globalInfo := map[string]interface{}{
		"werf": map[string]interface{}{
			"name":    projectName,
			"version": werf.Version,
		},
	}

	werfInfo := map[string]interface{}{
		"name":    projectName,
		"version": werf.Version,
		"repo":    repo,
		"image":   map[string]interface{}{},
	}

	if opts.Env != "" {
		globalInfo["env"] = opts.Env
		werfInfo["env"] = opts.Env
	} else if opts.IsStub {
		globalInfo["env"] = ""
		werfInfo["env"] = ""
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
			werfInfo["nameless_image"] = imageInfoGetter.GetName()
		} else {
			werfInfo["image"].(map[string]interface{})[imageInfoGetter.GetWerfImageName()] = imageInfoGetter.GetName()
		}
	}

	res := map[string]interface{}{
		"werf":   werfInfo,
		"global": globalInfo,
	}

	if opts.SetDockerConfigJsonValue {
		if err := writeDockerConfigJsonValue(ctx, res, opts.DockerConfigPath); err != nil {
			return nil, fmt.Errorf("error writing docker config value: %s", err)
		}
	}

	data, err := yaml.Marshal(res)
	logboek.Context(ctx).Debug().LogF("GetServiceValues result (err=%s):\n%s\n", err, data)

	return res, nil
}

func GetBundleServiceValues(ctx context.Context, opts ServiceValuesOptions) (map[string]interface{}, error) {
	globalInfo := map[string]interface{}{
		"werf": map[string]interface{}{
			"version": werf.Version,
		},
	}

	werfInfo := map[string]interface{}{
		"version": werf.Version,
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
			return nil, fmt.Errorf("error writing docker config value: %s", err)
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
		return fmt.Errorf("error accessing %q: %s", configJsonPath, err)
	}

	if data, err := ioutil.ReadFile(configJsonPath); err != nil {
		return fmt.Errorf("error reading %q: %s", configJsonPath, err)
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
