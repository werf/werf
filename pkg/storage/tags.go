package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/werf/global_warnings"
)

const (
	cleanupTriggerTagCount                = 100
	repoMetaImagesWarningThresholdPercent = 80
	repoCleanupLastTimeNever              = "never"
	repoCleanupOverdueThreshold           = 14 * 24 * time.Hour
	warnTemplateLastCleanupOverdue        = "The `werf cleanup` command has not been run for a long time: %s. This may lead to an accumulation of outdated images and metadata in the container registry.\n"
	warnTemplateMetaTagsExceed            = "The number of meta tags in the %s repository exceeds the normal threshold of %.0f%%! (%d of %d tags are meta tags). This may impact tag listing performance and increase command execution time.\n"
	warnTemplateAdvice                    = `— To clean up the container registry from outdated images and meta tags, it is recommended to periodically run the 'werf cleanup' command.
— To disable the Git history-based cleanup policy and prevent meta images from being published to the container registry while keeping other cleanup strategies, set 'cleanup.disableGitHistoryBasedPolicy: true' in the werf configuration file.
— To disable all cleanup policies and suppress this warning, set 'cleanup.disable: true'.`
)

func (storage *RepoStagesStorage) checkMeta(ctx context.Context, tags []string, _ ...docker_registry.Option) error {
	if len(tags) > cleanupTriggerTagCount || storage.cleanupDisabled {
		return nil
	}

	onceIface, _ := storage.warnMetaTagsOverflowOnce.LoadOrStore(storage.RepoAddress, new(sync.Once))
	once := onceIface.(*sync.Once)

	var onceErr error
	once.Do(func() {
		var warnMessage strings.Builder

		if err := runCleanupNeededChecks(ctx, storage, tags, &warnMessage); err != nil {
			onceErr = err
			return
		}

		if warnMessage.Len() != 0 {
			warnMessage.WriteString(warnTemplateAdvice)
			global_warnings.GlobalWarningLn(ctx, warnMessage.String())
		}
	})

	return onceErr
}

func runCleanupNeededChecks(ctx context.Context, storage *RepoStagesStorage, tags []string, b *strings.Builder) error {
	if err := checkLastCleanup(ctx, storage, tags, b); errors.Is(err, ErrCleanupNotOverdue) {
		return nil
	}

	extraChecks := []func(ctx context.Context, storage *RepoStagesStorage, tags []string, b *strings.Builder) error{
		checkMetaTags,
	}

	for _, checkFunc := range extraChecks {
		if err := checkFunc(ctx, storage, tags, b); err != nil {
			return err
		}
	}

	return nil
}

func checkMetaTags(_ context.Context, storage *RepoStagesStorage, tags []string, b *strings.Builder) error {
	metaCount := countMetaTags(tags)

	if metaCount == 0 {
		return nil
	}

	total := len(tags)
	percent := float64(metaCount) / float64(total) * 100
	if percent > repoMetaImagesWarningThresholdPercent {
		b.WriteString(fmt.Sprintf(
			warnTemplateMetaTagsExceed,
			storage.RepoAddress, percent, metaCount, total,
		))
	}

	return nil
}

func countMetaTags(tags []string) int {
	metaCount := 0
	for _, t := range tags {
		if strings.HasPrefix(t, RepoImageMetadataByCommitRecord_ImageTagPrefix) {
			metaCount++
		}
	}
	return metaCount
}

var ErrCleanupNotOverdue = fmt.Errorf("cleanup is not overdue, no need to warning")

func checkLastCleanup(ctx context.Context, storage *RepoStagesStorage, tags []string, b *strings.Builder) error {
	if len(tags) == 0 {
		return nil
	}
	lastCleanup, err := getLastCleanupRecord(ctx, storage.DockerRegistry, storage.RepoAddress, tags)
	if err != nil {
		return fmt.Errorf("getting last cleanup record: %w", err)
	}

	lastCleanupTime := formatLastCleanupTime(lastCleanup)
	if lastCleanupTime == repoCleanupLastTimeNever || isCleanupOverdue(lastCleanup) {
		b.WriteString(fmt.Sprintf(warnTemplateLastCleanupOverdue, lastCleanupTime))
		return nil
	} else {
		return ErrCleanupNotOverdue
	}
}

func formatLastCleanupTime(lastCleanup *CleanupRecord) string {
	if lastCleanup != nil && lastCleanup.TimestampMillisec > 0 {
		t := time.UnixMilli(lastCleanup.TimestampMillisec)
		return t.Format(time.RFC3339)
	}
	return repoCleanupLastTimeNever
}

func isCleanupOverdue(lastCleanup *CleanupRecord) bool {
	if lastCleanup == nil || lastCleanup.TimestampMillisec == 0 {
		return true
	}
	return time.Since(time.UnixMilli(lastCleanup.TimestampMillisec)) > repoCleanupOverdueThreshold
}
