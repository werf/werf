package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/flant/dapp/pkg/deploy"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/dapp/pkg/ruby2go"
	"github.com/flant/dapp/pkg/secret"
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
			key, err := secret.GenerateAexSecretKey()
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

			s, err := deploy.GetSecret(projectDir)
			if err != nil {
				return nil, err
			}

			var secretGenerator *deploy.SecretGenerator
			switch cmd {
			case "secret_generate":
				if secretGenerator, err = newSecretGenerateGenerator(s); err != nil {
					return nil, err
				}

				return nil, secretGenerate(secretGenerator, *options)
			case "secret_extract":
				if secretGenerator, err = newSecretExtractGenerator(s); err != nil {
					return nil, err
				}

				return nil, secretExtract(secretGenerator, *options)
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

			newSecret, err := deploy.GetSecret(projectDir)
			if err != nil {
				return nil, err
			}

			oldSecret, err := secret.NewSecret([]byte(oldKey))
			if err != nil {
				return nil, err
			}

			return nil, SecretsRegenerate(newSecret, oldSecret, projectDir, secretValuesPaths...)

		case "secret_edit":
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

			value, hasKey := args["projectDir"]
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

			return nil, runDeploy(projectDir, releaseName, tag, kubeContext, repo, rubyCliOptions)

		default:
			return nil, fmt.Errorf("command `%s` isn't supported", cmd)
		}

		return nil, nil
	})
}
