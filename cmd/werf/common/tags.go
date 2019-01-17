package common

import (
	"fmt"
	"os"
	"path"

	"github.com/flant/werf/pkg/build"
	"github.com/flant/werf/pkg/git_repo"
	"github.com/flant/werf/pkg/slug"
)

func GetDeployTag(cmdData *CmdData, projectDir string) (string, error) {
	optionsCount := 0
	if len(*cmdData.Tag) > 0 {
		optionsCount += len(*cmdData.Tag)
	}

	if *cmdData.TagBranch {
		optionsCount++
	}
	if *cmdData.TagCommit {
		optionsCount++
	}
	if *cmdData.TagBuildID {
		optionsCount++
	}
	if *cmdData.TagCI {
		optionsCount++
	}

	if optionsCount > 1 {
		return "", fmt.Errorf("exactly one tag should be specified for deploy")
	}

	opts, err := GetTagOptions(cmdData, projectDir)
	if err != nil {
		return "", err
	}

	tags := []string{}
	tags = append(tags, opts.Tags...)
	tags = append(tags, opts.TagsByCI...)
	tags = append(tags, opts.TagsByGitBranch...)
	tags = append(tags, opts.TagsByGitCommit...)
	tags = append(tags, opts.TagsByGitTag...)

	return tags[0], nil
}

func GetTagOptions(cmdData *CmdData, projectDir string) (build.TagOptions, error) {
	emptyTags := true

	opts := build.TagOptions{}

	for _, tag := range *cmdData.Tag {
		err := slug.ValidateDockerTag(tag)
		if err != nil {
			return build.TagOptions{}, fmt.Errorf("bad --tag parameter '%s' specified: %s", tag, err)
		}

		opts.Tags = append(opts.Tags, tag)
		emptyTags = false
	}

	if *cmdData.TagBranch {
		localGitRepo := &git_repo.Local{
			Path:   projectDir,
			GitDir: path.Join(projectDir, ".git"),
		}

		branch, err := localGitRepo.HeadBranchName()
		if err != nil {
			return build.TagOptions{}, fmt.Errorf("cannot detect local git branch for --tag-branch option: %s", err)
		}

		opts.TagsByGitBranch = append(opts.TagsByGitBranch, slug.DockerTag(branch))
		emptyTags = false
	}

	if *cmdData.TagCommit {
		localGitRepo := &git_repo.Local{
			Path:   projectDir,
			GitDir: path.Join(projectDir, ".git"),
		}

		commit, err := localGitRepo.HeadCommit()
		if err != nil {
			return build.TagOptions{}, fmt.Errorf("cannot detect local git HEAD commit for --tag-commit option: %s", err)
		}

		opts.TagsByGitCommit = append(opts.TagsByGitCommit, commit)
		emptyTags = false
	}

	if *cmdData.TagBuildID {
		var buildID string

		if os.Getenv("GITLAB_CI") != "" {
			buildID = os.Getenv("CI_BUILD_ID")
			if buildID == "" {
				buildID = os.Getenv("CI_JOB_ID")
			}
		} else if os.Getenv("TRAVIS") != "" {
			buildID = os.Getenv("TRAVIS_BUILD_NUMBER")
		} else {
			return build.TagOptions{}, fmt.Errorf("GITLAB_CI or TRAVIS environment variables has not been found for --tag-build-id option")
		}

		if buildID != "" {
			opts.TagsByCI = append(opts.TagsByCI, buildID)
			emptyTags = false
		}
	}

	if *cmdData.TagCI {
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
			return build.TagOptions{}, fmt.Errorf("GITLAB_CI or TRAVIS environment variables has not been found for --tag-ci option")
		}

		if gitTag != "" {
			opts.TagsByGitTag = append(opts.TagsByGitTag, slug.DockerTag(gitTag))
			emptyTags = false
		}
		if gitBranch != "" {
			opts.TagsByGitBranch = append(opts.TagsByGitBranch, slug.DockerTag(gitBranch))
			emptyTags = false
		}
	}

	if emptyTags {
		opts.Tags = append(opts.Tags, "latest")
	}

	return opts, nil
}
