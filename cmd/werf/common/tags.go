package common

import (
	"fmt"

	"github.com/werf/werf/pkg/build"
	"github.com/werf/werf/pkg/slug"
	"github.com/werf/werf/pkg/tag_strategy"
)

type TagOptionsGetterOptions struct {
	Optional bool
}

func GetDeployTag(cmdData *CmdData, opts TagOptionsGetterOptions) (string, tag_strategy.TagStrategy, error) {
	optionsCount := 0
	if len(*cmdData.TagCustom) > 0 {
		optionsCount += len(*cmdData.TagCustom)
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
	if *cmdData.TagByStagesSignature {
		optionsCount++
	}

	if optionsCount > 1 {
		return "", "", fmt.Errorf("exactly one tag option should be specified for deploy")
	}

	tagOpts, err := GetTagOptions(cmdData, opts)
	if err != nil {
		return "", "", err
	}

	if tagOpts.TagByStagesSignature {
		return "", tag_strategy.StagesSignature, nil
	} else if len(tagOpts.CustomTags) > 0 {
		return tagOpts.CustomTags[0], tag_strategy.Custom, nil
	} else if len(tagOpts.TagsByGitBranch) > 0 {
		return tagOpts.TagsByGitBranch[0], tag_strategy.GitBranch, nil
	} else if len(tagOpts.TagsByGitTag) > 0 {
		return tagOpts.TagsByGitTag[0], tag_strategy.GitTag, nil
	} else if len(tagOpts.TagsByGitCommit) > 0 {
		return tagOpts.TagsByGitCommit[0], tag_strategy.GitCommit, nil
	}

	if !opts.Optional {
		panic("tagOpts should contain at least one tag!")
	}

	return "", "", nil
}

func GetTagOptions(cmdData *CmdData, opts TagOptionsGetterOptions) (build.TagOptions, error) {
	emptyTags := true

	res := build.TagOptions{}

	for _, tag := range *cmdData.TagCustom {
		err := slug.ValidateDockerTag(tag)
		if err != nil {
			return build.TagOptions{}, fmt.Errorf("bad --tag-custom parameter '%s' specified: %s", tag, err)
		}

		res.CustomTags = append(res.CustomTags, tag)
		emptyTags = false
	}

	if tag := *cmdData.TagGitBranch; tag != "" {
		err := slug.ValidateDockerTag(tag)
		if err != nil {
			return build.TagOptions{}, fmt.Errorf("bad --tag-git-branch parameter '%s' specified: %s", tag, err)
		}

		res.TagsByGitBranch = append(res.TagsByGitBranch, tag)
		emptyTags = false
	}

	if tag := *cmdData.TagGitTag; tag != "" {
		err := slug.ValidateDockerTag(tag)
		if err != nil {
			return build.TagOptions{}, fmt.Errorf("bad --tag-git-tag parameter '%s' specified: %s", tag, err)
		}

		res.TagsByGitTag = append(res.TagsByGitTag, tag)
		emptyTags = false
	}

	if tag := *cmdData.TagGitCommit; tag != "" {
		err := slug.ValidateDockerTag(tag)
		if err != nil {
			return build.TagOptions{}, fmt.Errorf("bad --tag-git-commit parameter '%s' specified: %s", tag, err)
		}

		res.TagsByGitCommit = append(res.TagsByGitCommit, tag)
		emptyTags = false
	}

	if *cmdData.TagByStagesSignature {
		res.TagByStagesSignature = true
		emptyTags = false
	}

	if emptyTags && !opts.Optional {
		return build.TagOptions{}, fmt.Errorf("tag should be specified with --tag-by-stages-signature, --tag-custom, --tag-git-tag, --tag-git-branch or --tag-git-commit options")
	}

	return res, nil
}
