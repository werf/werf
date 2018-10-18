package main

import (
	"encoding/json"
	"fmt"

	"github.com/flant/dapp/pkg/git_repo"
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

	ruby2go.RunCli("git-repo", func(args map[string]interface{}) (interface{}, error) {
		if state, hasKey := args["LocalGitRepo"]; hasKey {
			repo := &git_repo.Local{}
			json.Unmarshal([]byte(state.(string)), repo)

			switch method := args["method"]; method {
			case "IsCommitExists":
				res := make(map[string]interface{})

				commit, hasKey := args["commit"]
				if !hasKey {
					return nil, fmt.Errorf("commit argument required!")
				}

				resValue, resErr := repo.IsCommitExists(commit.(string))

				res["result"] = resValue

				newState, err := json.Marshal(repo)
				if err != nil {
					return nil, err
				}
				res["LocalGitRepo"] = string(newState)

				return res, resErr

			case "LatestBranchCommit":
				res := make(map[string]interface{})

				branch, hasKey := args["branch"]
				if !hasKey {
					return nil, fmt.Errorf("branch argument required!")
				}

				resValue, resErr := repo.LatestBranchCommit(branch.(string))

				res["result"] = resValue

				newState, err := json.Marshal(repo)
				if err != nil {
					return nil, err
				}
				res["LocalGitRepo"] = string(newState)

				return res, resErr

			case "FindCommitIdByMessage":
				res := make(map[string]interface{})

				regex, hasKey := args["regex"]
				if !hasKey {
					return nil, fmt.Errorf("regex argument required!")
				}

				resValue, resErr := repo.FindCommitIdByMessage(regex.(string))

				res["result"] = resValue

				newState, err := json.Marshal(repo)
				if err != nil {
					return nil, err
				}
				res["LocalGitRepo"] = string(newState)

				return res, resErr

			case "IsEmpty":
				res := make(map[string]interface{})
				resValue, resErr := repo.IsEmpty()

				res["result"] = resValue

				newState, err := json.Marshal(repo)
				if err != nil {
					return nil, err
				}
				res["LocalGitRepo"] = string(newState)

				return res, resErr

			case "HeadCommit":
				res := make(map[string]interface{})
				resValue, resErr := repo.HeadCommit()

				res["result"] = resValue

				newState, err := json.Marshal(repo)
				if err != nil {
					return nil, err
				}
				res["LocalGitRepo"] = string(newState)

				return res, resErr

			case "HeadBranchName":
				res := make(map[string]interface{})
				resValue, resErr := repo.HeadBranchName()

				res["result"] = resValue

				newState, err := json.Marshal(repo)
				if err != nil {
					return nil, err
				}
				res["LocalGitRepo"] = string(newState)

				return res, resErr

			case "RemoteOriginUrl":
				res := make(map[string]interface{})
				resValue, resErr := repo.RemoteOriginUrl()

				res["result"] = resValue

				newState, err := json.Marshal(repo)
				if err != nil {
					return nil, err
				}
				res["LocalGitRepo"] = string(newState)

				return res, resErr

			default:
				return nil, fmt.Errorf("unknown method \"%s\"", method)
			}
		} else if state, hasKey := args["RemoteGitRepo"]; hasKey {
			repo := &git_repo.Remote{}
			json.Unmarshal([]byte(state.(string)), repo)

			switch method := args["method"]; method {
			case "FindCommitIdByMessage":
				res := make(map[string]interface{})

				regex, hasKey := args["regex"]
				if !hasKey {
					return nil, fmt.Errorf("regex argument required!")
				}

				resValue, resErr := repo.FindCommitIdByMessage(regex.(string))

				res["result"] = resValue

				newState, err := json.Marshal(repo)
				if err != nil {
					return nil, err
				}
				res["RemoteGitRepo"] = string(newState)

				return res, resErr

			case "IsEmpty":
				res := make(map[string]interface{})
				resValue, resErr := repo.IsEmpty()

				res["result"] = resValue

				newState, err := json.Marshal(repo)
				if err != nil {
					return nil, err
				}
				res["RemoteGitRepo"] = string(newState)

				return res, resErr

			case "HeadCommit":
				res := make(map[string]interface{})
				resValue, resErr := repo.HeadCommit()

				res["result"] = resValue

				newState, err := json.Marshal(repo)
				if err != nil {
					return nil, err
				}
				res["RemoteGitRepo"] = string(newState)

				return res, resErr

			case "HeadBranchName":
				res := make(map[string]interface{})
				resValue, resErr := repo.HeadBranchName()

				res["result"] = resValue

				newState, err := json.Marshal(repo)
				if err != nil {
					return nil, err
				}
				res["RemoteGitRepo"] = string(newState)

				return res, resErr

			case "CloneAndFetch":
				res := make(map[string]interface{})
				resErr := repo.CloneAndFetch()

				newState, err := json.Marshal(repo)
				if err != nil {
					return nil, err
				}
				res["RemoteGitRepo"] = string(newState)

				return res, resErr

			case "IsCommitExists":
				res := make(map[string]interface{})

				commit, hasKey := args["commit"]
				if !hasKey {
					return nil, fmt.Errorf("commit argument required!")
				}

				resValue, resErr := repo.IsCommitExists(commit.(string))

				res["result"] = resValue

				newState, err := json.Marshal(repo)
				if err != nil {
					return nil, err
				}
				res["RemoteGitRepo"] = string(newState)

				return res, resErr

			case "LatestBranchCommit":
				res := make(map[string]interface{})

				branch, hasKey := args["branch"]
				if !hasKey {
					return nil, fmt.Errorf("branch argument required!")
				}

				resValue, resErr := repo.LatestBranchCommit(branch.(string))

				res["result"] = resValue

				newState, err := json.Marshal(repo)
				if err != nil {
					return nil, err
				}
				res["RemoteGitRepo"] = string(newState)

				return res, resErr

			case "RemoteOriginUrl":
				res := make(map[string]interface{})
				resValue, resErr := repo.RemoteOriginUrl()

				res["result"] = resValue

				newState, err := json.Marshal(repo)
				if err != nil {
					return nil, err
				}
				res["RemoteGitRepo"] = string(newState)

				return res, resErr

			default:
				return nil, fmt.Errorf("unknown method \"%s\"", method)
			}
		} else {
			return nil, fmt.Errorf("bad args %+v", args)
		}
	})
}
