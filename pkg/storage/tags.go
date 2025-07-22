package storage

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

const (
	repoMetaImagesWarningThresholdPercent = 80
	repoCleanupLastTimeNever              = "never"
)

var (
	overThresholdWarningTemplate  = "WARNING: The number of meta tags exceeds %.0f%% in %s repository! (%d of %d)\n"
	lastCleanupRecordTemplate     = "Last cleanup was: %s\nRun `werf cleanup` to remove old meta tags and free up space in the repository.\n"
	disableCleanupWarningTemplate = `Meta images are only required for Git history-based cleanup.
To disable this policy and prevent meta images from being published to the container registry while keeping other cleanup strategies, set 'cleanup.disableGitHistoryBasedPolicy: true' in the werf configuration file.
To disable all cleanup policies, set 'cleanup.disable: true'.`
)

func (storage *RepoStagesStorage) analyzeMetaTags(ctx context.Context, tags []string, opts ...docker_registry.Option) error {
	if len(tags) == 0 {
		return nil
	}
	if storage.gitHistoryBasedCleanupDisabled || storage.cleanupDisabled {
		return nil
	}
	onceIface, _ := storage.warnMetaTagsOverflowOnce.LoadOrStore(storage.RepoAddress, new(sync.Once))
	once := onceIface.(*sync.Once)

	var onceErr error
	once.Do(func() {
		metaCount := 0
		for _, t := range tags {
			if strings.HasPrefix(t, RepoImageMetadataByCommitRecord_ImageTagPrefix) {
				metaCount++
			}
		}

		if metaCount == 0 {
			return
		}

		total := len(tags)
		percent := float64(metaCount) / float64(total) * 100

		if percent > repoMetaImagesWarningThresholdPercent {
			var b strings.Builder

			b.WriteString(fmt.Sprintf(
				overThresholdWarningTemplate,
				percent, storage.RepoAddress, metaCount, total,
			))

			lastCleanup, err := getLastCleanupRecord(ctx, storage.DockerRegistry, storage.RepoAddress, tags)
			if err != nil {
				onceErr = err
				return
			}

			lc := repoCleanupLastTimeNever
			if lastCleanup != nil && lastCleanup.TimestampMillisec > 0 {
				t := time.UnixMilli(lastCleanup.TimestampMillisec)
				lc = t.Format(time.RFC3339)
			}

			b.WriteString(fmt.Sprintf(lastCleanupRecordTemplate, lc))
			b.WriteString(disableCleanupWarningTemplate)

			global_warnings.GlobalWarningLn(ctx, b.String())
		}
	})

	return onceErr
}
