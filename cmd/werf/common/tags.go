package common

import (
	"fmt"
	"github.com/flant/werf/pkg/build"
	"github.com/flant/werf/pkg/slug"
)

func GetDeployTag(cmdData *CmdData) (string, error) {
	optionsCount := 0
	if len(*cmdData.Tag) > 0 {
		optionsCount += len(*cmdData.Tag)
	}

	if *cmdData.TagGitBranch != "" {
		optionsCount++
	}
	if *cmdData.TagGitTag != "" {
		optionsCount++
	}
	if *cmdData.TagGitCommit != "" {
		optionsCount++
	}

	if optionsCount > 1 {
		return "", fmt.Errorf("exactly one tag should be specified for deploy")
	}

	opts, err := GetTagOptions(cmdData)
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

func GetTagOptions(cmdData *CmdData) (build.TagOptions, error) {
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

	if *cmdData.TagGitBranch != "" {
		opts.TagsByGitBranch = append(opts.TagsByGitBranch, slug.DockerTag(*cmdData.TagGitBranch))
		emptyTags = false
	}

	if *cmdData.TagGitTag != "" {
		opts.TagsByGitTag = append(opts.TagsByGitTag, slug.DockerTag(*cmdData.TagGitTag))
		emptyTags = false
	}

	if *cmdData.TagGitCommit != "" {
		opts.TagsByGitCommit = append(opts.TagsByGitCommit, *cmdData.TagGitCommit)
		emptyTags = false
	}

	if emptyTags {
		opts.Tags = append(opts.Tags, "latest")
	}

	return opts, nil
}
