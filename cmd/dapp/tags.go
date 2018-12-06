package main

import (
	"fmt"
	"os"
	"path"

	"github.com/flant/dapp/pkg/build"
	"github.com/flant/dapp/pkg/git_repo"
)

func getPushOptions(projectDir string, tagOption []string, tagBranchOption bool, tagCommitOption bool, tagBuildIDOption bool, tagCIOption bool, withStages bool) (build.PushOptions, error) {
	emptyTags := true

	opts := build.PushOptions{WithStages: withStages}
	opts.Tags = tagOption
	if len(tagOption) > 0 {
		emptyTags = false
	}

	if tagBranchOption {
		localGitRepo := &git_repo.Local{
			Path:   projectDir,
			GitDir: path.Join(projectDir, ".git"),
		}

		branch, err := localGitRepo.HeadBranchName()
		if err != nil {
			return build.PushOptions{}, fmt.Errorf("cannot detect local git branch for --tag-branch option: %s", err)
		}

		opts.TagsByGitBranch = append(opts.TagsByGitBranch, branch)
		emptyTags = false
	}

	if tagCommitOption {
		localGitRepo := &git_repo.Local{
			Path:   projectDir,
			GitDir: path.Join(projectDir, ".git"),
		}

		commit, err := localGitRepo.HeadCommit()
		if err != nil {
			return build.PushOptions{}, fmt.Errorf("cannot detect local git HEAD commit for --tag-commit option: %s", err)
		}

		opts.TagsByGitCommit = append(opts.TagsByGitCommit, commit)
		emptyTags = false
	}

	if tagBuildIDOption {
		var buildID string

		if os.Getenv("GITLAB_CI") != "" {
			buildID = os.Getenv("CI_BUILD_ID")
			if buildID == "" {
				buildID = os.Getenv("CI_JOB_ID")
			}
		} else if os.Getenv("TRAVIS") != "" {
			buildID = os.Getenv("TRAVIS_BUILD_NUMBER")
		} else {
			return build.PushOptions{}, fmt.Errorf("GITLAB_CI or TRAVIS environment variables has not been found for --tag-build-id option")
		}

		if buildID != "" {
			opts.TagsByCI = append(opts.TagsByCI, buildID)
			emptyTags = false
		}
	}

	if tagCIOption {
		var gitBranch, gitTag string

		if os.Getenv("GITLAB_CI") != "" {
			gitTag = os.Getenv("CI_BUILD_TAG")
			if gitTag != "" {
				gitTag = os.Getenv("CI_COMMIT_TAG")
			}

			gitBranch = os.Getenv("CI_BUILD_REF_NAME")
			if gitBranch != "" {
				gitBranch = os.Getenv("CI_COMMIT_REF_NAME")
			}
		} else if os.Getenv("TRAVIS") != "" {
			gitTag = os.Getenv("TRAVIS_TAG")
			gitBranch = os.Getenv("TRAVIS_BRANCH")
		} else {
			return build.PushOptions{}, fmt.Errorf("GITLAB_CI or TRAVIS environment variables has not been found for --tag-ci option")
		}

		if gitTag != "" {
			opts.TagsByGitTag = append(opts.TagsByGitTag, gitTag)
		}
		if gitBranch != "" {
			opts.TagsByGitBranch = append(opts.TagsByGitBranch, gitBranch)
		}
	}

	if emptyTags {
		opts.Tags = append(opts.Tags, "latest")
	}

	return opts, nil
}
