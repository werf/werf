package chart_extender

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/werf/logboek"
)

func NewExtraValuesData() *ExtraValuesData {
	return &ExtraValuesData{extraValues: make(map[string]interface{})}
}

type ExtraValuesData struct {
	extraValues map[string]interface{}
}

func (d *ExtraValuesData) GetExtraValues() map[string]interface{} {
	return d.extraValues
}

func WriteDockerConfigJsonValue(ctx context.Context, values map[string]interface{}, dockerConfigPath string) error {
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
