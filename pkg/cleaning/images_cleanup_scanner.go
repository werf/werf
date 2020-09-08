package cleaning

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
)

func scanReferencesHistory(ctx context.Context, gitRepository *git.Repository, refs []*referenceToScan, expectedContentSignatureCommitHashes map[string][]plumbing.Hash) ([]string, []plumbing.Hash, error) {
	var reachedContentSignatureList []string
	var stopCommitHashes []plumbing.Hash
	var hitCommitHashes []plumbing.Hash

	for i := len(refs) - 1; i >= 0; i-- {
		ref := refs[i]

		var refReachedContentSignatureList []string
		var refStopCommitHashes []plumbing.Hash
		var refHitCommitHashes []plumbing.Hash
		var err error

		var logProcessMessage string
		if ref.Reference.Name().IsTag() {
			logProcessMessage = "Tag " + ref.String()
		} else {
			logProcessMessage = "Reference " + ref.String()
		}

		if err := logboek.Context(ctx).Info().LogProcess(logProcessMessage).DoError(func() error {
			refReachedContentSignatureList, refStopCommitHashes, refHitCommitHashes, err = scanReferenceHistory(ctx, gitRepository, ref, expectedContentSignatureCommitHashes, stopCommitHashes)
			if err != nil {
				return fmt.Errorf("scan reference history failed: %s", err)
			}

			stopCommitHashes = append(stopCommitHashes, refStopCommitHashes...)

		reachedContentSignatureListLoop:
			for _, c1 := range refReachedContentSignatureList {
				for _, c2 := range reachedContentSignatureList {
					if c1 == c2 {
						continue reachedContentSignatureListLoop
					}
				}

				reachedContentSignatureList = append(reachedContentSignatureList, c1)
			}

		refHitCommitHashesLoop:
			for _, refHitCommitHash := range refHitCommitHashes {
				for _, hitCommitHash := range hitCommitHashes {
					if refHitCommitHash == hitCommitHash {
						continue refHitCommitHashesLoop
					}
				}

				hitCommitHashes = append(hitCommitHashes, refHitCommitHash)
			}

			return nil
		}); err != nil {
			return nil, nil, err
		}
	}

	return reachedContentSignatureList, hitCommitHashes, nil
}

func applyImagesCleanupInPolicy(gitRepository *git.Repository, contentSignatureCommitHashes map[string][]plumbing.Hash, in *time.Duration) map[string][]plumbing.Hash {
	if in == nil {
		return contentSignatureCommitHashes
	}

	policyContentSignatureCommitHashes := map[string][]plumbing.Hash{}
	for contentSignature, commitHashList := range contentSignatureCommitHashes {
		var resultCommitHashList []plumbing.Hash
		for _, commitHash := range commitHashList {
			commit, err := gitRepository.CommitObject(commitHash)
			if err != nil {
				panic("unexpected condition")
			}

			if commit.Committer.When.After(time.Now().Add(-*in)) {
				resultCommitHashList = append(resultCommitHashList, commitHash)
			}
		}

		if len(resultCommitHashList) != 0 {
			policyContentSignatureCommitHashes[contentSignature] = resultCommitHashList
		}
	}

	return policyContentSignatureCommitHashes
}

type commitHistoryScanner struct {
	gitRepository                        *git.Repository
	expectedContentSignatureCommitHashes map[string][]plumbing.Hash
	reachedContentSignatureCommitHashes  map[string][]plumbing.Hash
	reachedCommitHashes                  []plumbing.Hash
	stopCommitHashes                     []plumbing.Hash
	isAlreadyScannedCommit               map[plumbing.Hash]bool
	scanDepth                            int

	referenceScanOptions
}

func (s *commitHistoryScanner) reachedContentSignatureList() []string {
	var reachedContentSignatureList []string
	for contentSignature, commitHashes := range s.reachedContentSignatureCommitHashes {
		if len(commitHashes) != 0 {
			reachedContentSignatureList = append(reachedContentSignatureList, contentSignature)
		}
	}

	return reachedContentSignatureList
}

func scanReferenceHistory(ctx context.Context, gitRepository *git.Repository, ref *referenceToScan, expectedContentSignatureCommitHashes map[string][]plumbing.Hash, stopCommitHashes []plumbing.Hash) ([]string, []plumbing.Hash, []plumbing.Hash, error) {
	filteredExpectedContentSignatureCommitHashes := applyImagesCleanupInPolicy(gitRepository, expectedContentSignatureCommitHashes, ref.imagesCleanupKeepPolicy.In)

	var refExpectedContentSignatureCommitHashes map[string][]plumbing.Hash
	isImagesCleanupKeepPolicyOnlyInOrAndBoth := ref.imagesCleanupKeepPolicy.Last == nil || (ref.imagesCleanupKeepPolicy.Operator != nil && *ref.imagesCleanupKeepPolicy.Operator == config.AndOperator)
	if isImagesCleanupKeepPolicyOnlyInOrAndBoth {
		refExpectedContentSignatureCommitHashes = filteredExpectedContentSignatureCommitHashes
	} else {
		refExpectedContentSignatureCommitHashes = expectedContentSignatureCommitHashes
	}

	if len(refExpectedContentSignatureCommitHashes) == 0 {
		logboek.Context(ctx).Info().LogLn("Skip reference due to nothing to seek")
		return []string{}, stopCommitHashes, []plumbing.Hash{}, nil
	}

	s := &commitHistoryScanner{
		gitRepository:                        gitRepository,
		expectedContentSignatureCommitHashes: refExpectedContentSignatureCommitHashes,
		reachedContentSignatureCommitHashes:  map[string][]plumbing.Hash{},
		stopCommitHashes:                     stopCommitHashes,

		referenceScanOptions:   ref.referenceScanOptions,
		isAlreadyScannedCommit: map[plumbing.Hash]bool{},
	}

	if err := s.scanCommitHistory(ctx, ref.HeadCommit.Hash); err != nil {
		return nil, nil, nil, fmt.Errorf("scan commit %s history failed: %s", ref.HeadCommit.Hash.String(), err)
	}

	isImagesCleanupKeepPolicyLastWithoutLimit := s.referenceScanOptions.imagesCleanupKeepPolicy.Last != nil && *s.referenceScanOptions.imagesCleanupKeepPolicy.Last != -1
	if isImagesCleanupKeepPolicyLastWithoutLimit {
		if len(s.reachedContentSignatureList()) == *s.referenceScanOptions.imagesCleanupKeepPolicy.Last {
			return s.reachedContentSignatureList(), s.stopCommitHashes, s.hitCommitHashes(), nil
		} else if len(s.reachedContentSignatureList()) > *s.referenceScanOptions.imagesCleanupKeepPolicy.Last {
			logboek.Context(ctx).Info().LogF("Reached more content signatures than expected by last (%d/%d)\n", len(s.reachedContentSignatureList()), *s.referenceScanOptions.imagesCleanupKeepPolicy.Last)

			latestCommitContentSignature := s.latestCommitContentSignature()
			var latestCommitList []*object.Commit
			for latestCommit := range latestCommitContentSignature {
				latestCommitList = append(latestCommitList, latestCommit)
			}

			sort.Slice(latestCommitList, func(i, j int) bool {
				return latestCommitList[i].Committer.When.After(latestCommitList[j].Committer.When)
			})

			if s.referenceScanOptions.imagesCleanupKeepPolicy.In == nil {
				return s.handleExtraContentSignaturesByLast(ctx, latestCommitContentSignature, latestCommitList)
			} else {
				return s.handleExtraContentSignaturesByLastWithIn(ctx, latestCommitContentSignature, latestCommitList)
			}
		}
	}

	if !reflect.DeepEqual(expectedContentSignatureCommitHashes, refExpectedContentSignatureCommitHashes) {
		return s.reachedContentSignatureList(), s.stopCommitHashes, s.hitCommitHashes(), nil
	}

	return s.handleStopCommitHashes(ctx, ref)
}

func (s *commitHistoryScanner) handleStopCommitHashes(ctx context.Context, ref *referenceToScan) ([]string, []plumbing.Hash, []plumbing.Hash, error) {
	if s.referenceScanOptions.scanDepthLimit != 0 {
		if len(s.reachedContentSignatureList()) == len(s.expectedContentSignatureCommitHashes) {
			s.stopCommitHashes = append(s.stopCommitHashes, s.reachedCommitHashes[len(s.reachedCommitHashes)-1])
		} else {
			return s.reachedContentSignatureList(), s.stopCommitHashes, s.hitCommitHashes(), nil
		}
	} else if len(s.reachedContentSignatureList()) != 0 {
		s.stopCommitHashes = append(s.stopCommitHashes, s.reachedCommitHashes[len(s.reachedCommitHashes)-1])
	} else {
		s.stopCommitHashes = append(s.stopCommitHashes, ref.HeadCommit.Hash)
	}
	logboek.Context(ctx).Debug().LogF("Stop commit %s added\n", s.stopCommitHashes[len(s.stopCommitHashes)-1].String())

	return s.reachedContentSignatureList(), s.stopCommitHashes, s.hitCommitHashes(), nil
}

func (s *commitHistoryScanner) handleExtraContentSignaturesByLastWithIn(ctx context.Context, latestCommitContentSignature map[*object.Commit]string, latestCommitList []*object.Commit) ([]string, []plumbing.Hash, []plumbing.Hash, error) {
	var latestCommitListByLast []*object.Commit
	var latestCommitListByIn []*object.Commit
	var hitCommitHashes []plumbing.Hash

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

	var reachedContentSignatureList []string
	for _, latestCommit := range resultLatestCommitList {
		contentSignature := latestCommitContentSignature[latestCommit]
		reachedContentSignatureList = append(reachedContentSignatureList, contentSignature)
		hitCommitHashes = append(hitCommitHashes, latestCommit.Hash)
	}

	var skippedContentSignatureList []string
latestCommitContentSignatureLoop:
	for _, contentSignature := range latestCommitContentSignature {
		for _, reachedContentSignature := range reachedContentSignatureList {
			if contentSignature == reachedContentSignature {
				continue latestCommitContentSignatureLoop
			}
		}

		skippedContentSignatureList = append(skippedContentSignatureList, contentSignature)
	}

	if len(skippedContentSignatureList) != 0 {
		logboek.Context(ctx).Info().LogBlock(fmt.Sprintf("Skipped content signatures by keep policy (%s)", s.imagesCleanupKeepPolicy.String())).Do(func() {
			for _, contentSignature := range skippedContentSignatureList {
				logboek.Context(ctx).Info().LogLn(contentSignature)
			}
		})
	}

	return reachedContentSignatureList, s.stopCommitHashes, hitCommitHashes, nil
}

func (s *commitHistoryScanner) handleExtraContentSignaturesByLast(ctx context.Context, latestCommitContentSignature map[*object.Commit]string, latestCommitList []*object.Commit) ([]string, []plumbing.Hash, []plumbing.Hash, error) {
	var reachedContentSignatureList []string
	var skippedContentSignatureList []string
	var hitCommitHashes []plumbing.Hash

	for ind, latestCommit := range latestCommitList {
		contentSignature := latestCommitContentSignature[latestCommit]
		if ind < *s.referenceScanOptions.imagesCleanupKeepPolicy.Last {
			reachedContentSignatureList = append(reachedContentSignatureList, contentSignature)
			hitCommitHashes = append(hitCommitHashes, latestCommit.Hash)
		} else {
			skippedContentSignatureList = append(skippedContentSignatureList, contentSignature)
		}
	}

	logboek.Context(ctx).Info().LogBlock(fmt.Sprintf("Skipped content signatures by keep policy (%s)", s.imagesCleanupKeepPolicy.String())).Do(func() {
		for _, contentSignature := range skippedContentSignatureList {
			logboek.Context(ctx).Info().LogLn(contentSignature)
		}
	})

	return reachedContentSignatureList, s.stopCommitHashes, hitCommitHashes, nil
}

func (s *commitHistoryScanner) scanCommitHistory(ctx context.Context, commitHash plumbing.Hash) error {
	var currentIteration, nextIteration []plumbing.Hash

	currentIteration = append(currentIteration, commitHash)
	for {
		s.scanDepth++

		for _, commitHash := range currentIteration {
			if s.isAlreadyScannedCommit[commitHash] {
				continue
			}

			if s.isStopCommitHash(commitHash) {
				logboek.Context(ctx).Debug().LogF("Stop scanning commit history %s due to stop commit reached\n", commitHash.String())
				continue
			}

			commitParents, err := s.handleCommitHash(ctx, commitHash)
			if err != nil {
				return err
			}

			if s.scanDepth == s.referenceScanOptions.scanDepthLimit {
				logboek.Context(ctx).Debug().LogF("Stop scanning commit history %s due to scanDepthLimit (%d)\n", commitHash.String(), s.referenceScanOptions.scanDepthLimit)
				continue
			}

			if len(s.expectedContentSignatureCommitHashes) == len(s.reachedContentSignatureCommitHashes) {
				logboek.Context(ctx).Debug().LogLn("Stop scanning due to all expected content signatures reached")
				break
			}

			s.isAlreadyScannedCommit[commitHash] = true
			nextIteration = append(nextIteration, commitParents...)
		}

		if len(nextIteration) == 0 {
			break
		}

		currentIteration = nextIteration
		nextIteration = []plumbing.Hash{}
	}

	return nil
}

func (s *commitHistoryScanner) handleCommitHash(ctx context.Context, commitHash plumbing.Hash) ([]plumbing.Hash, error) {
outerLoop:
	for contentSignature, commitHashes := range s.expectedContentSignatureCommitHashes {
		for _, c := range commitHashes {
			if c == commitHash {
				for _, reachedC := range s.reachedCommitHashes {
					if reachedC == c {
						break outerLoop
					}
				}

				if s.imagesCleanupKeepPolicy.In != nil {
					commit, err := s.gitRepository.CommitObject(commitHash)
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

				s.reachedCommitHashes = append(s.reachedCommitHashes, c)

				reachedCommitHashes, ok := s.reachedContentSignatureCommitHashes[contentSignature]
				if !ok {
					reachedCommitHashes = []plumbing.Hash{}
				}
				reachedCommitHashes = append(reachedCommitHashes, c)
				s.reachedContentSignatureCommitHashes[contentSignature] = reachedCommitHashes

				if !ok {
					logboek.Context(ctx).Info().LogF(
						"Expected content signature %s was reached on commit %s\n",
						contentSignature,
						commitHash.String(),
					)
				} else {
					logboek.Context(ctx).Info().LogF(
						"Expected content signature %s was reached again on another commit %s\n",
						contentSignature,
						commitHash.String(),
					)
				}

				break outerLoop
			}
		}
	}

	co, err := s.gitRepository.CommitObject(commitHash)
	if err != nil {
		return nil, fmt.Errorf("commit hash %s resolve failed: %s", commitHash.String(), err)
	}

	return co.ParentHashes, nil
}

func (s *commitHistoryScanner) isStopCommitHash(commitHash plumbing.Hash) bool {
	for _, c := range s.stopCommitHashes {
		if commitHash == c {
			return true
		}
	}

	return false
}

func (s *commitHistoryScanner) hitCommitHashes() []plumbing.Hash {
	var result []plumbing.Hash
	for commit, _ := range s.latestCommitContentSignature() {
		result = append(result, commit.Hash)
	}

	return result
}

func (s *commitHistoryScanner) latestCommitContentSignature() map[*object.Commit]string {
	contentSignatureLatestCommit := map[*object.Commit]string{}
	for contentSignature, commitHashes := range s.reachedContentSignatureCommitHashes {
		var latestCommit *object.Commit
		for _, commitHash := range commitHashes {
			commit, err := s.gitRepository.CommitObject(commitHash)
			if err != nil {
				panic("unexpected condition")
			}

			if latestCommit == nil || commit.Committer.When.After(latestCommit.Committer.When) {
				latestCommit = commit
			}
		}

		if latestCommit != nil {
			contentSignatureLatestCommit[latestCommit] = contentSignature
		}
	}

	return contentSignatureLatestCommit
}
