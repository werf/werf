package main

import (
	"encoding/json"
	"fmt"

	"github.com/flant/dapp/pkg/git_artifact"
	"github.com/flant/dapp/pkg/ruby2go"
)

func main() {
	ruby2go.RunCli("git-artifact", func(args map[string]interface{}) (map[string]interface{}, error) {
		res := make(map[string]interface{})

		ga := &git_artifact.GitArtifact{}
		if state, hasState := args["state"]; hasState {
			json.Unmarshal([]byte(state.(string)), ga)
		}

		switch command := args["command"]; command {
		case "LatestCommit":
			resultValue, resErr := ga.LatestCommit()
			res["result"] = resultValue

			newState, err := json.Marshal(ga)
			if err != nil {
				return nil, err
			}
			res["state"] = string(newState)

			return res, resErr
		default:
			return nil, fmt.Errorf("unknown command \"%s\"", command)
		}
	})
}
