package main

import (
	"encoding/json"
	"fmt"

	"github.com/flant/dapp/pkg/build"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/dapp/pkg/ruby2go"
	"github.com/flant/dapp/pkg/true_git"
)

func main() {
	var err error

	err = lock.Init()
	if err != nil {
		panic(err)
	}

	err = true_git.Init()
	if err != nil {
		panic(err)
	}

	ruby2go.RunCli("git-artifact", func(args map[string]interface{}) (interface{}, error) {
		res := make(map[string]interface{})

		ga := &build.GitArtifact{}
		if state, hasKey := args["GitArtifact"]; hasKey {
			json.Unmarshal([]byte(state.(string)), ga)
		}

		var state []byte

		switch method := args["method"]; method {
		case "LatestCommit":
			resultValue, resErr := ga.LatestCommit()
			res["result"] = resultValue

			state, err = json.Marshal(ga)
			if err != nil {
				return nil, err
			}
			res["GitArtifact"] = string(state)

			return res, resErr

		case "ApplyPatchCommand":
			stage := &build.StubStage{}
			if state, hasKey := args["Stage"]; hasKey {
				err := json.Unmarshal([]byte(state.(string)), stage)
				if err != nil {
					return nil, err
				}
			}

			resultValue, resErr := ga.ApplyPatchCommand(stage)

			res["result"] = resultValue

			state, err = json.Marshal(ga)
			if err != nil {
				return nil, err
			}
			res["GitArtifact"] = string(state)

			state, err = json.Marshal(stage)
			if err != nil {
				return nil, err
			}
			res["Stage"] = string(state)

			return res, resErr

		case "ApplyArchiveCommand":
			stage := &build.StubStage{}
			if state, hasKey := args["Stage"]; hasKey {
				err := json.Unmarshal([]byte(state.(string)), stage)
				if err != nil {
					return nil, err
				}
			}

			resultValue, resErr := ga.ApplyArchiveCommand(stage)

			res["result"] = resultValue

			state, err = json.Marshal(ga)
			if err != nil {
				return nil, err
			}
			res["GitArtifact"] = string(state)

			state, err = json.Marshal(stage)
			if err != nil {
				return nil, err
			}
			res["Stage"] = string(state)

			return res, resErr

		case "StageDependenciesChecksum":
			stageName, hasKey := args["StageName"]
			if !hasKey {
				return nil, fmt.Errorf("StageName argument required!")
			}
			resultValue, resErr := ga.StageDependenciesChecksum(stageName.(string))

			res["result"] = resultValue

			state, err = json.Marshal(ga)
			if err != nil {
				return nil, err
			}
			res["GitArtifact"] = string(state)

			return res, resErr

		case "PatchSize":
			fromCommit, hasKey := args["FromCommit"]
			if !hasKey {
				return nil, fmt.Errorf("FromCommit argument required!")
			}
			resultValue, resErr := ga.PatchSize(fromCommit.(string))

			res["result"] = resultValue

			state, err = json.Marshal(ga)
			if err != nil {
				return nil, err
			}
			res["GitArtifact"] = string(state)

			return res, resErr

		default:
			return nil, fmt.Errorf("unknown method \"%s\"", method)
		}
	})
}
