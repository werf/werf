package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/flant/dapp/pkg/deploy"
	"github.com/flant/dapp/pkg/deploy/secret"
	"github.com/flant/dapp/pkg/docker"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/dapp/pkg/ruby2go"
	"github.com/flant/kubedog/pkg/kube"
)

func main() {
	ruby2go.RunCli("deploy", func(args map[string]interface{}) (interface{}, error) {
		cmd, err := ruby2go.CommandFieldFromArgs(args)
		if err != nil {
			return nil, err
		}

		switch cmd {
		case "secret_key_generate":
			key, err := secret.GenerateSecretKey()
			if err != nil {
				return nil, err
			}

			fmt.Printf("DAPP_SECRET_KEY=%s\n", string(key))

			return nil, nil
		case "secret_generate", "secret_extract":
			projectDir, err := ruby2go.StringOptionFromArgs("project_dir", args)
			if err != nil {
				return nil, err
			}

			rawOptions, err := ruby2go.StringOptionFromArgs("raw_command_options", args)
			if err != nil {
				return nil, err
			}

			options := &secretGenerateOptions{}
			err = json.Unmarshal([]byte(rawOptions), options)
			if err != nil {
				return nil, err
			}

			s, err := secret.GetSecret(projectDir)
			if err != nil {
				return nil, err
			}

			switch cmd {
			case "secret_generate":
				return nil, secretGenerate(s, *options)
			case "secret_extract":
				return nil, secretExtract(s, *options)
			}
		case "secret_regenerate":
			projectDir, err := ruby2go.StringOptionFromArgs("project_dir", args)
			if err != nil {
				return nil, err
			}

			oldKey, err := ruby2go.StringOptionFromArgs("old_key", args)
			if err != nil {
				return nil, err
			}

			secretValuesPaths, err := ruby2go.StringArrayOptionFromArgs("secret_values_paths", args)
			if err != nil {
				return nil, err
			}

			newSecret, err := secret.GetSecret(projectDir)
			if err != nil {
				return nil, err
			}

			oldSecret, err := secret.NewSecretByKey([]byte(oldKey))
			if err != nil {
				return nil, err
			}

			return nil, SecretsRegenerate(newSecret, oldSecret, projectDir, secretValuesPaths...)
		case "secret_edit":
			projectDir, err := ruby2go.StringOptionFromArgs("project_dir", args)
			if err != nil {
				return nil, err
			}

			filePath, err := ruby2go.StringOptionFromArgs("file_path", args)
			if err != nil {
				return nil, err
			}

			tmpDir, err := ruby2go.StringOptionFromArgs("tmp_dir", args)
			if err != nil {
				return nil, err
			}

			options, err := ruby2go.OptionsFieldFromArgs(args)
			if err != nil {
				return nil, err
			}

			values, err := ruby2go.BoolFieldFromMapInterface("values", options)
			if err != nil {
				return nil, err
			}

			s, err := secret.GetSecret(projectDir)
			if err != nil {
				return nil, err
			}

			return nil, secretEdit(s, filePath, values, tmpDir)
		case "deploy":
			var rubyCliOptions deployRubyCliOptions
			if value, hasKey := args["rubyCliOptions"]; hasKey {
				err = json.Unmarshal([]byte(value.(string)), &rubyCliOptions)
				if err != nil {
					return nil, err
				}
			}

			err = lock.Init()
			if err != nil {
				return nil, err
			}

			hostDockerConfigDir, err := ruby2go.StringOptionFromArgs("host_docker_config_dir", args)
			if err != nil {
				return nil, err
			}

			os.Setenv("DOCKER_CONFIG", hostDockerConfigDir)
			log.SetFlags(0)
			log.SetOutput(ioutil.Discard)

			if err := docker.Init(hostDockerConfigDir); err != nil {
				return nil, err
			}

			kubeContext := os.Getenv("KUBECONTEXT")
			if kubeContext == "" {
				kubeContext = rubyCliOptions.Context
			}
			err = kube.Init(kube.InitOptions{KubeContext: kubeContext})
			if err != nil {
				return nil, err
			}

			err = deploy.Init()
			if err != nil {
				return nil, err
			}

			value, hasKey := args["projectName"]
			if !hasKey {
				return nil, fmt.Errorf("projectName argument required!")
			}
			projectName := value.(string)

			value, hasKey = args["projectDir"]
			if !hasKey {
				return nil, fmt.Errorf("projectDir argument required!")
			}
			projectDir := value.(string)

			value, hasKey = args["releaseName"]
			if !hasKey {
				return nil, fmt.Errorf("releaseName argument required!")
			}
			releaseName := value.(string)

			value, hasKey = args["tag"]
			if !hasKey {
				return nil, fmt.Errorf("tag argument required!")
			}
			tag := value.(string)

			value, hasKey = args["repo"]
			if !hasKey {
				return nil, fmt.Errorf("repo argument required!")
			}
			repo := value.(string)

			var dimgs []*deploy.DimgInfoGetterStub
			if value, hasKey := args["dimgs"]; hasKey {
				err = json.Unmarshal([]byte(value.(string)), &dimgs)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("dimgs argument required!")
			}

			return nil, runDeploy(projectName, projectDir, releaseName, tag, kubeContext, repo, dimgs, rubyCliOptions)

		default:
			return nil, fmt.Errorf("command `%s` isn't supported", cmd)
		}

		return nil, nil
	})
}
