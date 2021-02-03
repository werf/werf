package inspector

import (
	"context"

	"github.com/werf/werf/pkg/git_repo/status"
	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/util"
)

func (i Inspector) InspectBuildContextFiles(ctx context.Context, matcher path_matcher.PathMatcher) error {
	if i.sharedOptions.LooseGiterminism() {
		return nil
	}

	if err := i.sharedOptions.LocalGitRepo().ValidateSubmodules(ctx, matcher); err != nil {
		return i.fileReader.HandleValidateSubmodulesErr(err)
	}

	filePathList, err := i.sharedOptions.LocalGitRepo().GetModifiedLocallyFilePathList(ctx, matcher, status.FilterOptions{
		WorktreeOnly:     i.sharedOptions.Dev(),
		IgnoreSubmodules: true,
	})
	if err != nil {
		return err
	}

	if len(filePathList) == 0 {
		return nil
	}

	var relativeToProjectDirPathList []string
	for _, path := range filePathList {
		relativeToProjectDirPathList = append(
			relativeToProjectDirPathList,
			util.GetRelativeToBaseFilepath(i.sharedOptions.RelativeToGitProjectDir(), path),
		)
	}

	return i.fileReader.ExtraWindowsCheckFilesModifiedLocally(ctx, relativeToProjectDirPathList...)
}
