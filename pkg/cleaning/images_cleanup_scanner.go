package cleaning

import (
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

func scanReferencesHistory(gitRepository *git.Repository, refs []*referenceToScan, expectedContentSignatureCommitHashes map[string][]plumbing.Hash) ([]string, error) {
	var reachedContentSignatureList []string
	var stopCommitHashes []plumbing.Hash

	for i := len(refs) - 1; i >= 0; i-- {
		ref := refs[i]

		var refReachedContentSignatureList []string
		var refStopCommitHashes []plumbing.Hash
		var err error

		var logProcessMessage string
		if ref.Reference.Name().IsTag() {
			logProcessMessage = "Tag " + ref.String()
		} else {
			logProcessMessage = "Reference " + ref.String()
		}

		if err := logboek.Info.LogProcess(logProcessMessage, logboek.LevelLogProcessOptions{}, func() error {
			refReachedContentSignatureList, refStopCommitHashes, err = scanReferenceHistory(gitRepository, ref, expectedContentSignatureCommitHashes, stopCommitHashes)
			if err != nil {
				return fmt.Errorf("scan reference history failed: %s", err)
			}

			stopCommitHashes = append(stopCommitHashes, refStopCommitHashes...)

		outerLoop:
			for _, c1 := range refReachedContentSignatureList {
				for _, c2 := range reachedContentSignatureList {
					if c1 == c2 {
						continue outerLoop
					}
				}

				reachedContentSignatureList = append(reachedContentSignatureList, c1)
			}

			return nil
		}); err != nil {
			return nil, err
		}

		if len(reachedContentSignatureList) == len(expectedContentSignatureCommitHashes) {
			logboek.Info.LogLn("Scanning stopped due to all expected commit hashes were reached")
			break
		}
	}

	return reachedContentSignatureList, nil
}

func applyImagesCleanupPublishedInPolicy(gitRepository *git.Repository, contentSignatureCommitHashes map[string][]plumbing.Hash, publishedIn *time.Duration) map[string][]plumbing.Hash {
	if publishedIn == nil {
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

			if commit.Committer.When.After(time.Now().Add(-*publishedIn)) {
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

func scanReferenceHistory(gitRepository *git.Repository, ref *referenceToScan, expectedContentSignatureCommitHashes map[string][]plumbing.Hash, stopCommitHashes []plumbing.Hash) ([]string, []plumbing.Hash, error) {
	filteredExpectedContentSignatureCommitHashes := applyImagesCleanupPublishedInPolicy(gitRepository, expectedContentSignatureCommitHashes, ref.imagesCleanupKeepPolicy.PublishedIn)

	var refExpectedContentSignatureCommitHashes map[string][]plumbing.Hash
	if ref.imagesCleanupKeepPolicy.Last == nil || (ref.imagesCleanupKeepPolicy.Operator != nil && *ref.imagesCleanupKeepPolicy.Operator == config.AndOperator) {
		refExpectedContentSignatureCommitHashes = filteredExpectedContentSignatureCommitHashes
	} else {
		refExpectedContentSignatureCommitHashes = expectedContentSignatureCommitHashes
	}

	if len(refExpectedContentSignatureCommitHashes) == 0 {
		logboek.Info.LogLn("Skip reference due to nothing to seek")
		return []string{}, stopCommitHashes, nil
	}

	s := &commitHistoryScanner{
		gitRepository:                        gitRepository,
		expectedContentSignatureCommitHashes: refExpectedContentSignatureCommitHashes,
		reachedContentSignatureCommitHashes:  map[string][]plumbing.Hash{},
		stopCommitHashes:                     stopCommitHashes,

		referenceScanOptions:   ref.referenceScanOptions,
		isAlreadyScannedCommit: map[plumbing.Hash]bool{},
	}

	if err := s.scanCommitHistory(ref.HeadCommit.Hash); err != nil {
		return nil, nil, fmt.Errorf("scan commit %s history failed: %s", ref.HeadCommit.Hash.String(), err)
	}

	if s.referenceScanOptions.imagesCleanupKeepPolicy.Last != nil && *s.referenceScanOptions.imagesCleanupKeepPolicy.Last != -1 {
		if len(s.reachedContentSignatureList()) == *s.referenceScanOptions.imagesCleanupKeepPolicy.Last {
			return s.reachedContentSignatureList(), s.stopCommitHashes, nil
		} else if len(s.reachedContentSignatureList()) > *s.referenceScanOptions.imagesCleanupKeepPolicy.Last {
			logboek.Info.LogF("Reached more content signatures than expected by last (%d/%d)\n", len(s.reachedContentSignatureList()), *s.referenceScanOptions.imagesCleanupKeepPolicy.Last)

			latestCommitContentSignature := s.latestCommitContentSignature()
			var latestCommitList []*object.Commit
			for latestCommit := range latestCommitContentSignature {
				latestCommitList = append(latestCommitList, latestCommit)
			}

			sort.Slice(latestCommitList, func(i, j int) bool {
				return latestCommitList[i].Committer.When.After(latestCommitList[j].Committer.When)
			})

			if s.referenceScanOptions.imagesCleanupKeepPolicy.PublishedIn == nil {
				return s.handleExtraContentSignaturesByLast(latestCommitContentSignature, latestCommitList)
			} else {
				return s.handleExtraContentSignaturesByLastWithPublishedIn(latestCommitContentSignature, latestCommitList)
			}
		}
	}

	if !reflect.DeepEqual(expectedContentSignatureCommitHashes, refExpectedContentSignatureCommitHashes) {
		return s.reachedContentSignatureList(), s.stopCommitHashes, nil
	}

	return s.handleStopCommitHashes(ref)
}

func (s *commitHistoryScanner) handleStopCommitHashes(ref *referenceToScan) ([]string, []plumbing.Hash, error) {
	if s.referenceScanOptions.scanDepthLimit != 0 {
		if len(s.reachedContentSignatureList()) == len(s.expectedContentSignatureCommitHashes) {
			s.stopCommitHashes = append(s.stopCommitHashes, s.reachedCommitHashes[len(s.reachedCommitHashes)-1])
		} else {
			return s.reachedContentSignatureList(), s.stopCommitHashes, nil
		}
	} else if len(s.reachedContentSignatureList()) != 0 {
		s.stopCommitHashes = append(s.stopCommitHashes, s.reachedCommitHashes[len(s.reachedCommitHashes)-1])
	} else {
		s.stopCommitHashes = append(s.stopCommitHashes, ref.HeadCommit.Hash)
	}
	logboek.Debug.LogF("Stop commit %s added\n", s.stopCommitHashes[len(s.stopCommitHashes)-1].String())

	return s.reachedContentSignatureList(), s.stopCommitHashes, nil
}

func (s *commitHistoryScanner) handleExtraContentSignaturesByLastWithPublishedIn(latestCommitContentSignature map[*object.Commit]string, latestCommitList []*object.Commit) ([]string, []plumbing.Hash, error) {
	var latestCommitListByLast []*object.Commit
	var latestCommitListByPublishedIn []*object.Commit

	for ind, latestCommit := range latestCommitList {
		if ind < *s.referenceScanOptions.imagesCleanupKeepPolicy.Last {
			latestCommitListByLast = append(latestCommitListByLast, latestCommit)
		}

		if latestCommit.Committer.When.After(time.Now().Add(-*s.referenceScanOptions.imagesCleanupKeepPolicy.PublishedIn)) {
			latestCommitListByPublishedIn = append(latestCommitListByPublishedIn, latestCommit)
		}
	}

	var resultLatestCommitList []*object.Commit
	if s.referenceScanOptions.imagesCleanupKeepPolicy.Operator == nil || *s.referenceScanOptions.imagesCleanupKeepPolicy.Operator == config.AndOperator {
		for _, commitByLast := range latestCommitListByLast {
			for _, commitByPublishedIn := range latestCommitListByPublishedIn {
				if commitByLast == commitByPublishedIn {
					resultLatestCommitList = append(resultLatestCommitList, commitByLast)
				}
			}
		}
	} else {
		resultLatestCommitList = latestCommitListByPublishedIn[:]
	latestCommitListByLastLoop:
		for _, commitByLast := range latestCommitListByLast {
			for _, commitByPublishedIn := range latestCommitListByPublishedIn {
				if commitByLast == commitByPublishedIn {
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
		_ = logboek.Info.LogBlock(fmt.Sprintf("Skipped content signatures by keep policy (%s)", s.imagesCleanupKeepPolicy.String()), logboek.LevelLogBlockOptions{}, func() error {
			for _, contentSignature := range skippedContentSignatureList {
				logboek.Info.LogLn(contentSignature)
			}
			return nil
		})
	}

	return reachedContentSignatureList, s.stopCommitHashes, nil
}

func (s *commitHistoryScanner) handleExtraContentSignaturesByLast(latestCommitContentSignature map[*object.Commit]string, latestCommitList []*object.Commit) ([]string, []plumbing.Hash, error) {
	var reachedContentSignatureList []string
	var skippedContentSignatureList []string
	for ind, latestCommit := range latestCommitList {
		contentSignature := latestCommitContentSignature[latestCommit]
		if ind < *s.referenceScanOptions.imagesCleanupKeepPolicy.Last {
			reachedContentSignatureList = append(reachedContentSignatureList, contentSignature)
		} else {
			skippedContentSignatureList = append(skippedContentSignatureList, contentSignature)
		}
	}

	_ = logboek.Info.LogBlock(fmt.Sprintf("Skipped content signatures by keep policy (%s)", s.imagesCleanupKeepPolicy.String()), logboek.LevelLogBlockOptions{}, func() error {
		for _, contentSignature := range skippedContentSignatureList {
			logboek.Info.LogLn(contentSignature)
		}
		return nil
	})

	return reachedContentSignatureList, s.stopCommitHashes, nil
}

func (s *commitHistoryScanner) scanCommitHistory(commitHash plumbing.Hash) error {
	var currentIteration, nextIteration []plumbing.Hash

	currentIteration = append(currentIteration, commitHash)
	for {
		s.scanDepth++

		for _, commitHash := range currentIteration {
			if s.isAlreadyScannedCommit[commitHash] {
				continue
			}

			if s.isStopCommitHash(commitHash) {
				logboek.Debug.LogF("Stop scanning commit history %s due to stop commit reached\n", commitHash.String())
				continue
			}

			commitParents, err := s.handleCommitHash(commitHash)
			if err != nil {
				return err
			}

			if s.scanDepth == s.referenceScanOptions.scanDepthLimit {
				logboek.Debug.LogF("Stop scanning commit history %s due to scanDepthLimit (%d)\n", commitHash.String(), s.referenceScanOptions.scanDepthLimit)
				continue
			}

			if len(s.expectedContentSignatureCommitHashes) == len(s.reachedContentSignatureCommitHashes) {
				logboek.Debug.LogLn("Stop scanning due to all expected content signatures reached")
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

func (s *commitHistoryScanner) handleCommitHash(commitHash plumbing.Hash) ([]plumbing.Hash, error) {
outerLoop:
	for contentSignature, commitHashes := range s.expectedContentSignatureCommitHashes {
		for _, c := range commitHashes {
			if c == commitHash {
				for _, reachedC := range s.reachedCommitHashes {
					if reachedC == c {
						break outerLoop
					}
				}

				if s.imagesCleanupKeepPolicy.PublishedIn != nil {
					commit, err := s.gitRepository.CommitObject(commitHash)
					if err != nil {
						panic("unexpected condition")
					}

					if s.imagesCleanupKeepPolicy.Last == nil || s.imagesCleanupKeepPolicy.Operator == nil || *s.imagesCleanupKeepPolicy.Operator == config.AndOperator {
						if commit.Committer.When.Before(time.Now().Add(-*s.imagesCleanupKeepPolicy.PublishedIn)) {
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
					logboek.Info.LogF(
						"Expected content signature %s was reached on commit %s\n",
						contentSignature,
						commitHash.String(),
					)
				} else {
					logboek.Info.LogF(
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
