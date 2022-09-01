package helpers

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"sigs.k8s.io/yaml"

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
	Namespace       string
	Env             string
	IsStub          bool
	StubImagesNames []string
	// disable env stub used in the werf-render command
	DisableEnvStub bool
	CommitHash     string
	CommitDate     *time.Time

	SetDockerConfigJsonValue bool
	DockerConfigPath         string
}

func GetEnvServiceValues(env string) map[string]interface{} {
	return map[string]interface{}{
		"werf":   map[string]interface{}{"env": env},
		"global": map[string]interface{}{"env": env},
	}
}

func GetServiceValues(ctx context.Context, projectName, repo string, imageInfoGetters []*image.InfoGetter, opts ServiceValuesOptions) (map[string]interface{}, error) {
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
		"tag":     map[string]interface{}{},
		"commit": map[string]interface{}{
			"hash": opts.CommitHash,
			"date": map[string]interface{}{
				"human": opts.CommitDate.String(),
				"unix":  opts.CommitDate.Unix(),
			},
		},
	}

	if opts.Env != "" {
		globalInfo["env"] = opts.Env
		werfInfo["env"] = opts.Env
	} else if opts.IsStub && !opts.DisableEnvStub {
		globalInfo["env"] = ""
		werfInfo["env"] = ""
	}

	if opts.Namespace != "" {
		werfInfo["namespace"] = opts.Namespace
	}

	if opts.IsStub {
		stubTag := "TAG"
		stubImage := fmt.Sprintf("%s:%s", repo, stubTag)

		werfInfo["is_stub"] = true
		werfInfo["stub_image"] = stubImage
		for _, name := range opts.StubImagesNames {
			werfInfo["image"].(map[string]interface{})[name] = stubImage
			werfInfo["tag"].(map[string]interface{})[name] = stubTag
		}
	}

	for _, imageInfoGetter := range imageInfoGetters {
		tag := imageInfoGetter.GetTag()
		image := imageInfoGetter.GetName()

		if imageInfoGetter.IsNameless() {
			werfInfo["is_nameless_image"] = true
			werfInfo["nameless_image"] = image
		} else {
			werfInfo["image"].(map[string]interface{})[imageInfoGetter.GetWerfImageName()] = image
			werfInfo["tag"].(map[string]interface{})[imageInfoGetter.GetWerfImageName()] = tag
		}
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
