package main

import (
	"encoding/json"
	"fmt"

	"github.com/flant/dapp/pkg/ruby2go"
)

func main() {
	ruby2go.RunCli("build", func(args map[string]interface{}) (interface{}, error) {
		cmd, err := ruby2go.CommandFieldFromArgs(args)
		if err != nil {
			return nil, err
		}

		switch cmd {
		case "build":
			var rubyCliOptions buildRubyCliOptions
			if value, hasKey := args["rubyCliOptions"]; hasKey {
				err = json.Unmarshal([]byte(value.(string)), &rubyCliOptions)
				if err != nil {
					return nil, err
				}
			}

			value, hasKey := args["projectDir"]
			if !hasKey {
				return nil, fmt.Errorf("projectDir argument required!")
			}
			projectDir := value.(string)

			return nil, runBuild(projectDir, rubyCliOptions)

		default:
			return nil, fmt.Errorf("command `%s` isn't supported", cmd)
		}

		return nil, nil
	})
}
