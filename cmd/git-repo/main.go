package main

import (
	"encoding/json"
	"fmt"

	"github.com/flant/dapp/pkg/git_repo"
	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/dapp/pkg/ruby2go"
)

func main() {
	err := lock.Init()
	if err != nil {
		panic(err)
	}

	ruby2go.RunCli("git-repo", func(args map[string]interface{}) (interface{}, error) {
		if state, hasKey := args["LocalGitRepo"]; hasKey {
			repo := &git_repo.Local{}
			json.Unmarshal([]byte(state.(string)), repo)

			switch method := args["method"]; method {
			default:
				return nil, fmt.Errorf("unknown method \"%s\"", method)
			}
		} else if state, hasKey := args["RemoteGitRepo"]; hasKey {
			repo := &git_repo.Remote{}
			json.Unmarshal([]byte(state.(string)), repo)

			switch method := args["method"]; method {
			case "CloneAndFetch":
				res := make(map[string]interface{})
				resErr := repo.CloneAndFetch()

				newState, err := json.Marshal(repo)
				if err != nil {
					return nil, err
				}
				res["state"] = string(newState)

				return res, resErr

			default:
				return nil, fmt.Errorf("unknown method \"%s\"", method)
			}
		} else {
			return nil, fmt.Errorf("bad args %+v", args)
		}
	})
}
