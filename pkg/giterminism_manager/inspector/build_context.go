package inspector

import (
	"context"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/git_repo/status"
	"github.com/werf/werf/pkg/path_matcher"
	"github.com/werf/werf/pkg/util"
)

func (i Inspector) InspectBuildContextFiles(ctx context.Context, matcher path_matcher.PathMatcher) error {
	if i.sharedOptions.LooseGiterminism() {
		return nil
	}

	logProcess := logboek.Context(ctx).Debug().LogProcess("status (%s)", matcher.String())
	logProcess.Start()
	result, err := i.sharedOptions.LocalGitRepo().Status(ctx, matcher)
	if err != nil {
		logProcess.Fail()
		return err
	} else {
		logProcess.End()
	}

	filePathList := result.FilePathList(status.FilterOptions{WorktreeOnly: i.sharedOptions.Dev()})

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
