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

	if *cmdData.TagGitBranch || os.Getenv("WERF_AUTOTAG_GIT_BRANCH") != "" {
		optionsCount++
	}
	if *cmdData.TagGitTag || os.Getenv("WERF_AUTOTAG_GIT_TAG") != "" {
		optionsCount++
	}
	if *cmdData.TagGitCommit {
		optionsCount++
	}

	if optionsCount > 1 {
		return "", fmt.Errorf("exactly one tag should be specified for deploy")
	}

	opts, err := GetTagOptions(cmdData, projectDir)
	if err != nil {
		return "", err
	}

	var tags []string
	tags = append(tags, opts.Tags...)
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

	if os.Getenv("WERF_AUTOTAG_GIT_BRANCH") != "" {
		opts.TagsByGitBranch = append(opts.TagsByGitBranch, slug.DockerTag(os.Getenv("WERF_AUTOTAG_GIT_BRANCH")))
	} else if *cmdData.TagGitBranch {
		localGitRepo := &git_repo.Local{
			Path:   projectDir,
			GitDir: path.Join(projectDir, ".git"),
		}

		branch, err := localGitRepo.HeadBranchName()
		if err != nil {
			return build.TagOptions{}, fmt.Errorf("cannot detect local git branch for --tag-git-branch option: %s", err)
		}

		opts.TagsByGitBranch = append(opts.TagsByGitBranch, slug.DockerTag(branch))
		emptyTags = false
	}

	if os.Getenv("WERF_AUTOTAG_GIT_TAG") != "" {
		opts.TagsByGitTag = append(opts.TagsByGitTag, slug.DockerTag(os.Getenv("WERF_AUTOTAG_GIT_TAG")))
	} else if *cmdData.TagGitTag {
		localGitRepo := &git_repo.Local{
			Path:   projectDir,
			GitDir: path.Join(projectDir, ".git"),
		}

		branch, err := localGitRepo.HeadTagName()
		if err != nil {
			return build.TagOptions{}, fmt.Errorf("cannot detect local git tag for --tag-git-tag option: %s", err)
		}

		opts.TagsByGitBranch = append(opts.TagsByGitBranch, slug.DockerTag(branch))
		emptyTags = false
	}

	if *cmdData.TagGitCommit {
		localGitRepo := &git_repo.Local{
			Path:   projectDir,
			GitDir: path.Join(projectDir, ".git"),
		}

		commit, err := localGitRepo.HeadCommit()
		if err != nil {
			return build.TagOptions{}, fmt.Errorf("cannot detect local git HEAD commit for --tag-git-commit option: %s", err)
		}

		opts.TagsByGitCommit = append(opts.TagsByGitCommit, commit)
		emptyTags = false
	}

	if emptyTags {
		opts.Tags = append(opts.Tags, "latest")
	}

	return opts, nil
}
