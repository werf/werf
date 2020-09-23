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
				return fmt.Errorf("scan reference history failed: %s", err)
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

	refExpectedStageIDCommitList := map[string][]string{}
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
		return nil, nil, nil, fmt.Errorf("scan commit %s history failed: %s", ref.HeadCommit.Hash.String(), err)
	}

	isImagesCleanupKeepPolicyLastWithoutLimit := s.referenceScanOptions.imagesCleanupKeepPolicy.Last != nil && *s.referenceScanOptions.imagesCleanupKeepPolicy.Last != -1
	if isImagesCleanupKeepPolicyLastWithoutLimit {
		if len(s.reachedStageIDList()) == *s.referenceScanOptions.imagesCleanupKeepPolicy.Last {
			return s.reachedStageIDList(), s.stopCommitList, s.stageIDHitCommitList(), nil
		} else if len(s.reachedStageIDList()) > *s.referenceScanOptions.imagesCleanupKeepPolicy.Last {
			logboek.Context(ctx).Info().LogF("Reached more content signatures than expected by last (%d/%d)\n", len(s.reachedStageIDList()), *s.referenceScanOptions.imagesCleanupKeepPolicy.Last)

			latestCommitStageID := s.latestCommitStageID()
			var latestCommitList []*object.Commit
			for latestCommit := range latestCommitStageID {
				latestCommitList = append(latestCommitList, latestCommit)
			}

			sort.Slice(latestCommitList, func(i, j int) bool {
				return latestCommitList[i].Committer.When.After(latestCommitList[j].Committer.When)
			})

			if s.referenceScanOptions.imagesCleanupKeepPolicy.In == nil {
				return s.handleExtraStageIDsByLast(ctx, latestCommitStageID, latestCommitList)
			} else {
				return s.handleExtraStageIDsByLastWithIn(ctx, latestCommitStageID, latestCommitList)
			}
		}
	}

	if !reflect.DeepEqual(expectedStageIDCommitList, refExpectedStageIDCommitList) {
		return s.reachedStageIDList(), s.stopCommitList, s.stageIDHitCommitList(), nil
	}

	return s.handleStopCommitList(ctx, ref)
}

func (s *commitHistoryScanner) handleStopCommitList(ctx context.Context, ref *ReferenceToScan) ([]string, []string, map[string][]string, error) {
	if s.referenceScanOptions.scanDepthLimit != 0 {
		if len(s.reachedStageIDList()) == len(s.expectedStageIDCommitList) {
			s.stopCommitList = append(s.stopCommitList, s.reachedCommitList[len(s.reachedCommitList)-1])
		} else {
			return s.reachedStageIDList(), s.stopCommitList, s.stageIDHitCommitList(), nil
		}
	} else if len(s.reachedStageIDList()) != 0 {
		s.stopCommitList = append(s.stopCommitList, s.reachedCommitList[len(s.reachedCommitList)-1])
	} else {
		s.stopCommitList = append(s.stopCommitList, ref.HeadCommit.Hash.String())
	}
	logboek.Context(ctx).Debug().LogF("Stop commit %s added\n", s.stopCommitList[len(s.stopCommitList)-1])

	return s.reachedStageIDList(), s.stopCommitList, s.stageIDHitCommitList(), nil
}

func (s *commitHistoryScanner) handleExtraStageIDsByLastWithIn(ctx context.Context, latestCommitStageID map[*object.Commit]string, latestCommitList []*object.Commit) ([]string, []string, map[string][]string, error) {
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
		resultLatestCommitList = latestCommitListByIn[:]
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
		stageID := latestCommitStageID[latestCommit]
		reachedStageIDList = append(reachedStageIDList, stageID)
		stageIDHitCommitList[stageID] = []string{latestCommit.Hash.String()}
	}

	var skippedStageIDList []string
latestCommitStageIDLoop:
	for _, stageID := range latestCommitStageID {
		for _, reachedStageID := range reachedStageIDList {
			if stageID == reachedStageID {
				continue latestCommitStageIDLoop
			}
		}

		skippedStageIDList = append(skippedStageIDList, stageID)
	}

	if len(skippedStageIDList) != 0 {
		logboek.Context(ctx).Info().LogBlock(fmt.Sprintf("Skipped content signatures by keep policy (%s)", s.imagesCleanupKeepPolicy.String())).Do(func() {
			for _, stageID := range skippedStageIDList {
				logboek.Context(ctx).Info().LogLn(stageID)
			}
		})
	}

	return reachedStageIDList, s.stopCommitList, stageIDHitCommitList, nil
}

func (s *commitHistoryScanner) handleExtraStageIDsByLast(ctx context.Context, latestCommitStageID map[*object.Commit]string, latestCommitList []*object.Commit) ([]string, []string, map[string][]string, error) {
	var reachedStageIDList []string
	var skippedStageIDList []string
	stageIDHitCommitList := map[string][]string{}

	for ind, latestCommit := range latestCommitList {
		stageID := latestCommitStageID[latestCommit]
		if ind < *s.referenceScanOptions.imagesCleanupKeepPolicy.Last {
			reachedStageIDList = append(reachedStageIDList, stageID)
			stageIDHitCommitList[stageID] = []string{latestCommit.Hash.String()}
		} else {
			skippedStageIDList = append(skippedStageIDList, stageID)
		}
	}

	logboek.Context(ctx).Info().LogBlock(fmt.Sprintf("Skipped content signatures by keep policy (%s)", s.imagesCleanupKeepPolicy.String())).Do(func() {
		for _, stageID := range skippedStageIDList {
			logboek.Context(ctx).Info().LogLn(stageID)
		}
	})

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
				logboek.Context(ctx).Debug().LogLn("Stop scanning due to all expected content signatures reached")
				break
			}

			s.isAlreadyScannedCommit[commit] = true
			nextIteration = append(nextIteration, commitParents...)
		}

		if len(nextIteration) == 0 {
			break
		}

		currentIteration = nextIteration
		nextIteration = []string{}
	}

	return nil
}

func (s *commitHistoryScanner) handleCommit(ctx context.Context, commit string) ([]string, error) {
outerLoop:
	for stageID, commitList := range s.expectedStageIDCommitList {
		for _, c := range commitList {
			if c == commit {
				for _, reachedC := range s.reachedCommitList {
					if reachedC == c {
						break outerLoop
					}
				}

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

				s.reachedCommitList = append(s.reachedCommitList, c)

				reachedCommitList, ok := s.reachedStageIDCommitList[stageID]
				if !ok {
					reachedCommitList = []string{}
				}
				reachedCommitList = append(reachedCommitList, c)
				s.reachedStageIDCommitList[stageID] = reachedCommitList

				if !ok {
					logboek.Context(ctx).Info().LogF(
						"Expected content signature %s was reached on commit %s\n",
						stageID,
						commit,
					)
				} else {
					logboek.Context(ctx).Info().LogF(
						"Expected content signature %s was reached again on another commit %s\n",
						stageID,
						commit,
					)
				}

				break outerLoop
			}
		}
	}

	co, err := s.gitRepository.CommitObject(plumbing.NewHash(commit))
	if err != nil {
		return nil, fmt.Errorf("commit hash %s resolve failed: %s", commit, err)
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
	for commit, stageID := range s.latestCommitStageID() {
		result[stageID] = []string{commit.Hash.String()}
	}

	return result
}

func (s *commitHistoryScanner) latestCommitStageID() map[*object.Commit]string {
	stageIDLatestCommit := map[*object.Commit]string{}
	for stageID, commitList := range s.reachedStageIDCommitList {
		var latestCommit *object.Commit
		for _, commit := range commitList {
			commit, err := s.gitRepository.CommitObject(plumbing.NewHash(commit))
			if err != nil {
				panic("unexpected condition")
			}

			if latestCommit == nil || commit.Committer.When.After(latestCommit.Committer.When) {
				latestCommit = commit
			}
		}

		if latestCommit != nil {
			stageIDLatestCommit[latestCommit] = stageID
		}
	}

	return stageIDLatestCommit
}
