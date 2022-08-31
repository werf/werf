package git_history_based_cleanup

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/config"
	"github.com/werf/werf/pkg/util"
)

func ScanReferencesHistory(ctx context.Context, gitRepository *git.Repository, refs []*ReferenceToScan, expectedStageIDCommitList map[string][]string) ([]string, map[string][]string, error) {
	var reachedStageIDs []string
	var stopCommitList []string
	stageIDHitCommitList := map[string][]string{}

	for i := len(refs) - 1; i >= 0; i-- {
		ref := refs[i]

		var refReachedStageIDs []string
		var refStopCommitList []string
		refStageIDHitCommitList := map[string][]string{}
		var err error

		var logProcessMessage string
		if ref.Reference.Name().IsTag() {
			logProcessMessage = "Tag " + ref.String()
		} else {
			logProcessMessage = "Reference " + ref.String()
		}

		if err := logboek.Context(ctx).Info().LogProcess(logProcessMessage).DoError(func() error {
			refReachedStageIDs, refStopCommitList, refStageIDHitCommitList, err = scanReferenceHistory(ctx, gitRepository, ref, expectedStageIDCommitList, stopCommitList)
			if err != nil {
				return fmt.Errorf("scan reference history failed: %w", err)
			}

			stopCommitList = util.AddNewStringsToStringArray(stopCommitList, refStopCommitList...)
			reachedStageIDs = util.AddNewStringsToStringArray(reachedStageIDs, refReachedStageIDs...)

			for refStageID, refCommitList := range refStageIDHitCommitList {
				hitCommitList, ok := stageIDHitCommitList[refStageID]
				if !ok {
					stageIDHitCommitList[refStageID] = refCommitList
					continue
				}

				stageIDHitCommitList[refStageID] = util.AddNewStringsToStringArray(hitCommitList, refCommitList...)
			}

			return nil
		}); err != nil {
			return nil, nil, err
		}
	}

	return reachedStageIDs, stageIDHitCommitList, nil
}

func applyImagesCleanupInPolicy(gitRepository *git.Repository, stageIDCommitList map[string][]string, in *time.Duration) map[string][]string {
	if in == nil {
		return stageIDCommitList
	}

	policyStageIDCommitList := map[string][]string{}
	for stageID, commitList := range stageIDCommitList {
		var resultCommitList []string
		for _, commit := range commitList {
			commitHash := plumbing.NewHash(commit)
			c, err := gitRepository.CommitObject(commitHash)
			if err != nil {
				panic("unexpected condition")
			}

			if c.Committer.When.After(time.Now().Add(-*in)) {
				resultCommitList = append(resultCommitList, commit)
			}
		}

		if len(resultCommitList) != 0 {
			policyStageIDCommitList[stageID] = resultCommitList
		}
	}

	return policyStageIDCommitList
}

type commitHistoryScanner struct {
	gitRepository             *git.Repository
	expectedStageIDCommitList map[string][]string
	reachedStageIDCommitList  map[string][]string
	reachedCommitList         []string
	stopCommitList            []string
	isAlreadyScannedCommit    map[string]bool
	scanDepth                 int

	referenceScanOptions
}

func (s *commitHistoryScanner) reachedStageIDList() []string {
	var reachedStageIDList []string
	for stageID, commitList := range s.reachedStageIDCommitList {
		if len(commitList) != 0 {
			reachedStageIDList = append(reachedStageIDList, stageID)
		}
	}

	return reachedStageIDList
}

func scanReferenceHistory(ctx context.Context, gitRepository *git.Repository, ref *ReferenceToScan, expectedStageIDCommitList map[string][]string, stopCommitList []string) ([]string, []string, map[string][]string, error) {
	filteredExpectedStageIDCommitList := applyImagesCleanupInPolicy(gitRepository, expectedStageIDCommitList, ref.imagesCleanupKeepPolicy.In)

	var refExpectedStageIDCommitList map[string][]string
	isImagesCleanupKeepPolicyOnlyInOrAndBoth := ref.imagesCleanupKeepPolicy.Last == nil || (ref.imagesCleanupKeepPolicy.Operator != nil && *ref.imagesCleanupKeepPolicy.Operator == config.AndOperator)
	if isImagesCleanupKeepPolicyOnlyInOrAndBoth {
		refExpectedStageIDCommitList = filteredExpectedStageIDCommitList
	} else {
		refExpectedStageIDCommitList = expectedStageIDCommitList
	}

	if len(refExpectedStageIDCommitList) == 0 {
		logboek.Context(ctx).Info().LogLn("Skip reference due to nothing to seek")
		return []string{}, stopCommitList, map[string][]string{}, nil
	}

	s := &commitHistoryScanner{
		gitRepository:             gitRepository,
		expectedStageIDCommitList: refExpectedStageIDCommitList,
		reachedStageIDCommitList:  map[string][]string{},
		stopCommitList:            stopCommitList,

		referenceScanOptions:   ref.referenceScanOptions,
		isAlreadyScannedCommit: map[string]bool{},
	}

	if err := s.scanCommitHistory(ctx, ref.HeadCommit.Hash.String()); err != nil {
		return nil, nil, nil, fmt.Errorf("scan commit %s history failed: %w", ref.HeadCommit.Hash.String(), err)
	}

	isImagesCleanupKeepPolicyLastWithoutLimit := s.referenceScanOptions.imagesCleanupKeepPolicy.Last != nil && *s.referenceScanOptions.imagesCleanupKeepPolicy.Last != -1
	if isImagesCleanupKeepPolicyLastWithoutLimit {
		if len(s.reachedStageIDList()) == *s.referenceScanOptions.imagesCleanupKeepPolicy.Last {
			return s.reachedStageIDList(), s.stopCommitList, s.stageIDHitCommitList(), nil
		} else if len(s.reachedStageIDList()) > *s.referenceScanOptions.imagesCleanupKeepPolicy.Last {
			logboek.Context(ctx).Info().LogF("Reached more tags than expected by last (%d/%d)\n", len(s.reachedStageIDList()), *s.referenceScanOptions.imagesCleanupKeepPolicy.Last)

			latestCommitStageIDs := s.latestCommitStageIDs()
			var latestCommitList []*object.Commit
			for latestCommit := range latestCommitStageIDs {
				latestCommitList = append(latestCommitList, latestCommit)
			}

			sort.Slice(latestCommitList, func(i, j int) bool {
				return latestCommitList[i].Committer.When.After(latestCommitList[j].Committer.When)
			})

			if s.referenceScanOptions.imagesCleanupKeepPolicy.In == nil {
				return s.handleExtraStageIDsByLast(ctx, latestCommitStageIDs, latestCommitList)
			} else {
				return s.handleExtraStageIDsByLastWithIn(ctx, latestCommitStageIDs, latestCommitList)
			}
		}
	}

	if !reflect.DeepEqual(expectedStageIDCommitList, refExpectedStageIDCommitList) {
		return s.reachedStageIDList(), s.stopCommitList, s.stageIDHitCommitList(), nil
	}

	return s.handleStopCommitList(ctx, ref)
}

func (s *commitHistoryScanner) handleStopCommitList(ctx context.Context, ref *ReferenceToScan) ([]string, []string, map[string][]string, error) {
	switch {
	case s.referenceScanOptions.scanDepthLimit != 0:
		if len(s.reachedStageIDList()) == len(s.expectedStageIDCommitList) {
			s.stopCommitList = append(s.stopCommitList, s.reachedCommitList[len(s.reachedCommitList)-1])
		} else {
			return s.reachedStageIDList(), s.stopCommitList, s.stageIDHitCommitList(), nil
		}
	case len(s.reachedStageIDList()) != 0:
		s.stopCommitList = append(s.stopCommitList, s.reachedCommitList[len(s.reachedCommitList)-1])
	default:
		s.stopCommitList = append(s.stopCommitList, ref.HeadCommit.Hash.String())
	}
	logboek.Context(ctx).Debug().LogF("Stop commit %s added\n", s.stopCommitList[len(s.stopCommitList)-1])

	return s.reachedStageIDList(), s.stopCommitList, s.stageIDHitCommitList(), nil
}

func (s *commitHistoryScanner) handleExtraStageIDsByLastWithIn(ctx context.Context, latestCommitStageIDs map[*object.Commit][]string, latestCommitList []*object.Commit) ([]string, []string, map[string][]string, error) {
	var latestCommitListByLast []*object.Commit
	var latestCommitListByIn []*object.Commit
	stageIDHitCommitList := map[string][]string{}

	for ind, latestCommit := range latestCommitList {
		if ind < *s.referenceScanOptions.imagesCleanupKeepPolicy.Last {
			latestCommitListByLast = append(latestCommitListByLast, latestCommit)
		}

		if latestCommit.Committer.When.After(time.Now().Add(-*s.referenceScanOptions.imagesCleanupKeepPolicy.In)) {
			latestCommitListByIn = append(latestCommitListByIn, latestCommit)
		}
	}

	var resultLatestCommitList []*object.Commit
	isImagesCleanupKeepPolicyOperatorAnd := s.referenceScanOptions.imagesCleanupKeepPolicy.Operator == nil || *s.referenceScanOptions.imagesCleanupKeepPolicy.Operator == config.AndOperator
	if isImagesCleanupKeepPolicyOperatorAnd {
		for _, commitByLast := range latestCommitListByLast {
			for _, commitByIn := range latestCommitListByIn {
				if commitByLast == commitByIn {
					resultLatestCommitList = append(resultLatestCommitList, commitByLast)
				}
			}
		}
	} else {
		resultLatestCommitList = latestCommitListByIn
	latestCommitListByLastLoop:
		for _, commitByLast := range latestCommitListByLast {
			for _, commitByIn := range latestCommitListByIn {
				if commitByLast == commitByIn {
					continue latestCommitListByLastLoop
				}
			}

			resultLatestCommitList = append(resultLatestCommitList, commitByLast)
		}
	}

	var reachedStageIDList []string
	for _, latestCommit := range resultLatestCommitList {
		stageIDs := latestCommitStageIDs[latestCommit]
		if len(stageIDs) > 1 {
			logboek.Context(ctx).Info().LogBlock("Counted tags as one due to identical related commit %s", latestCommit.Hash.String()).Do(func() {
				for _, stageID := range stageIDs {
					logboek.Context(ctx).Info().LogLn(stageID)
				}
			})
		}

		for _, stageID := range stageIDs {
			reachedStageIDList = append(reachedStageIDList, stageID)
			stageIDHitCommitList[stageID] = []string{latestCommit.Hash.String()}
		}
	}

	var skippedStageIDList []string
latestCommitStageIDLoop:
	for _, stageIDs := range latestCommitStageIDs {
		for _, stageID := range stageIDs {
			for _, reachedStageID := range reachedStageIDList {
				if stageID == reachedStageID {
					continue latestCommitStageIDLoop
				}
			}

			skippedStageIDList = append(skippedStageIDList, stageID)
		}
	}

	if len(skippedStageIDList) != 0 {
		logboek.Context(ctx).Info().LogBlock(fmt.Sprintf("Skipped tags by keep policy (%s)", s.imagesCleanupKeepPolicy.String())).Do(func() {
			for _, stageID := range skippedStageIDList {
				logboek.Context(ctx).Info().LogLn(stageID)
			}
		})
	}

	return reachedStageIDList, s.stopCommitList, stageIDHitCommitList, nil
}

func (s *commitHistoryScanner) handleExtraStageIDsByLast(ctx context.Context, latestCommitStageIDs map[*object.Commit][]string, latestCommitList []*object.Commit) ([]string, []string, map[string][]string, error) {
	var reachedStageIDList []string
	var skippedStageIDList []string
	stageIDHitCommitList := map[string][]string{}

	for ind, latestCommit := range latestCommitList {
		stageIDs := latestCommitStageIDs[latestCommit]
		if ind < *s.referenceScanOptions.imagesCleanupKeepPolicy.Last {
			if len(stageIDs) > 1 {
				logboek.Context(ctx).Info().LogBlock("Counted tags as one due to identical related commit %s", latestCommit.Hash.String()).Do(func() {
					for _, stageID := range stageIDs {
						logboek.Context(ctx).Info().LogLn(stageID)
					}
				})
			}

			for _, stageID := range stageIDs {
				reachedStageIDList = append(reachedStageIDList, stageID)
				stageIDHitCommitList[stageID] = []string{latestCommit.Hash.String()}
			}
		} else {
			skippedStageIDList = append(skippedStageIDList, stageIDs...)
		}
	}

	if len(skippedStageIDList) != 0 {
		logboek.Context(ctx).Info().LogBlock(fmt.Sprintf("Skipped tags by keep policy (%s)", s.imagesCleanupKeepPolicy.String())).Do(func() {
			for _, stageID := range skippedStageIDList {
				logboek.Context(ctx).Info().LogLn(stageID)
			}
		})
	}

	return reachedStageIDList, s.stopCommitList, stageIDHitCommitList, nil
}

func (s *commitHistoryScanner) scanCommitHistory(ctx context.Context, commit string) error {
	var currentIteration, nextIteration []string

	currentIteration = append(currentIteration, commit)
	for {
		s.scanDepth++

		for _, commit := range currentIteration {
			if s.isAlreadyScannedCommit[commit] {
				continue
			}

			if s.isStopCommit(commit) {
				logboek.Context(ctx).Debug().LogF("Stop scanning commit history %s due to stop commit reached\n", commit)
				continue
			}

			commitParents, err := s.handleCommit(ctx, commit)
			if err != nil {
				return err
			}

			if s.scanDepth == s.referenceScanOptions.scanDepthLimit {
				logboek.Context(ctx).Debug().LogF("Stop scanning commit history %s due to scanDepthLimit (%d)\n", commit, s.referenceScanOptions.scanDepthLimit)
				continue
			}

			if len(s.expectedStageIDCommitList) == len(s.reachedStageIDCommitList) {
				logboek.Context(ctx).Debug().LogLn("Stop scanning due to all expected tags reached")
				break
			}

			s.isAlreadyScannedCommit[commit] = true
			nextIteration = append(nextIteration, commitParents...)
		}

		if len(nextIteration) == 0 {
			break
		}

		currentIteration = nextIteration
		nextIteration = nil
	}

	return nil
}

func (s *commitHistoryScanner) handleCommit(ctx context.Context, commit string) ([]string, error) {
	var isReachedCommit bool

outerLoop:
	for stageID, commitList := range s.expectedStageIDCommitList {
		for _, c := range commitList {
			if c == commit {
				if s.imagesCleanupKeepPolicy.In != nil {
					commit, err := s.gitRepository.CommitObject(plumbing.NewHash(commit))
					if err != nil {
						panic("unexpected condition")
					}

					isImagesCleanupKeepPolicyOnlyInOrAndBoth := s.imagesCleanupKeepPolicy.Last == nil || s.imagesCleanupKeepPolicy.Operator == nil || *s.imagesCleanupKeepPolicy.Operator == config.AndOperator
					if isImagesCleanupKeepPolicyOnlyInOrAndBoth {
						if commit.Committer.When.Before(time.Now().Add(-*s.imagesCleanupKeepPolicy.In)) {
							break outerLoop
						}
					}
				}

				isReachedCommit = true
				s.reachedStageIDCommitList[stageID] = append(s.reachedStageIDCommitList[stageID], c)

				logboek.Context(ctx).Info().LogF(
					"Expected content digest %s was reached on commit %s\n",
					stageID,
					commit,
				)

				break
			}
		}
	}

	if isReachedCommit {
		s.reachedCommitList = append(s.reachedCommitList, commit)
	}

	co, err := s.gitRepository.CommitObject(plumbing.NewHash(commit))
	if err != nil {
		return nil, fmt.Errorf("commit hash %s resolve failed: %w", commit, err)
	}

	var parentHashes []string
	for _, commitHash := range co.ParentHashes {
		parentHashes = append(parentHashes, commitHash.String())
	}

	return parentHashes, nil
}

func (s *commitHistoryScanner) isStopCommit(commit string) bool {
	for _, c := range s.stopCommitList {
		if commit == c {
			return true
		}
	}

	return false
}

func (s *commitHistoryScanner) stageIDHitCommitList() map[string][]string {
	result := map[string][]string{}
	for commit, stageIDs := range s.latestCommitStageIDs() {
		for _, stageID := range stageIDs {
			result[stageID] = []string{commit.Hash.String()}
		}
	}

	return result
}

func (s *commitHistoryScanner) latestCommitStageIDs() map[*object.Commit][]string {
	latestCommitStageIDs := map[*object.Commit][]string{}
	commitObjectCache := map[string]*object.Commit{}
	for stageID, commitList := range s.reachedStageIDCommitList {
		var latestCommitObject *object.Commit
		for _, commit := range commitList {
			var commitObject *object.Commit

			var err error
			var ok bool
			if commitObject, ok = commitObjectCache[commit]; !ok {
				commitObject, err = s.gitRepository.CommitObject(plumbing.NewHash(commit))
				if err != nil {
					panic("unexpected condition")
				}

				commitObjectCache[commit] = commitObject
			}

			if latestCommitObject == nil || commitObject.Committer.When.After(latestCommitObject.Committer.When) {
				latestCommitObject = commitObject
			}
		}

		if latestCommitObject != nil {
			latestCommitStageIDs[latestCommitObject] = append(latestCommitStageIDs[latestCommitObject], stageID)
		}
	}

	return latestCommitStageIDs
}
