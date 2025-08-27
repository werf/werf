package gitdata

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/werf/logboek"
)

type GitArchiveDesc struct {
	MetadataPath  string
	ArchivePath   string
	Metadata      *ArchiveMetadata
	Size          uint64
	CacheBasePath string
}

func (entry *GitArchiveDesc) GetPaths() []string {
	return []string{entry.MetadataPath, entry.ArchivePath}
}

func (entry *GitArchiveDesc) GetSize() uint64 {
	return entry.Size
}

func (entry *GitArchiveDesc) GetLastAccessAt() time.Time {
	return time.Unix(entry.Metadata.LastAccessTimestamp, 0)
}

func (entry *GitArchiveDesc) GetCacheBasePath() string {
	return entry.CacheBasePath
}

// GetGitArchivesAndRemoveInvalid scans the given cacheVersionRoot directory and returns
// a list of GitArchiveDesc for each valid git archive found. It removes invalid
// entries and handles errors appropriately.
//
// The directory structure expected is as follows:
// ├── 39e4985a993e1688a3a7e548e9bbf007ea53f4654d746e966b7b6a5011b72ffa/
// │   ├── 29/
// │   │   ├── 296f52bea4934b503f8141226900ab3798ce9eeeefcbde068bd316c687e40320.meta.json
// │   │   ├── 296f52bea4934b503f8141226900ab3798ce9eeeefcbde068bd316c687e40320.tar
// │   │   └── ... (other archive files)
// │   └── ... (other hash prefixes)
// └── ... (other repository hashes)
func GetGitArchivesAndRemoveInvalid(ctx context.Context, cacheVersionRoot string) ([]GitDataEntry, error) {
	var res []GitDataEntry

	fileStat, err := os.Stat(cacheVersionRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error accessing dir %q: %w", cacheVersionRoot, err)
	}
	if !fileStat.IsDir() {
		logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: not a directory\n", cacheVersionRoot)
		if err := os.RemoveAll(cacheVersionRoot); err != nil {
			return nil, fmt.Errorf("unable to remove %q: %w", cacheVersionRoot, err)
		}
		return nil, nil
	}

	repoHashes, err := ioutil.ReadDir(cacheVersionRoot)
	if err != nil {
		return nil, fmt.Errorf("error reading dir %q: %w", cacheVersionRoot, err)
	}

	for _, repoHashInfo := range repoHashes {
		repoHashDir := filepath.Join(cacheVersionRoot, repoHashInfo.Name())

		if !repoHashInfo.IsDir() {
			logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: not a directory\n", repoHashDir)
			if err := os.RemoveAll(repoHashDir); err != nil {
				return nil, fmt.Errorf("unable to remove %q: %w", repoHashDir, err)
			}
			continue
		}

		hashPrefixes, err := ioutil.ReadDir(repoHashDir)
		if err != nil {
			return nil, fmt.Errorf("error reading repo archives dir %q: %w", repoHashDir, err)
		}

		for _, hashPrefixInfo := range hashPrefixes {
			hashPrefixDir := filepath.Join(repoHashDir, hashPrefixInfo.Name())

			if !hashPrefixInfo.IsDir() {
				logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: not a directory\n", hashPrefixDir)
				if err := os.RemoveAll(hashPrefixDir); err != nil {
					return nil, fmt.Errorf("unable to remove %q: %w", hashPrefixDir, err)
				}
				continue
			}

			archiveFiles, err := ioutil.ReadDir(hashPrefixDir)
			if err != nil {
				return nil, fmt.Errorf("error reading repo archives from dir %q: %w", hashPrefixDir, err)
			}

			for _, archiveMetaOrTarFileInfo := range archiveFiles {
				archiveMetaOrTarFilePath := filepath.Join(hashPrefixDir, archiveMetaOrTarFileInfo.Name())

				if archiveMetaOrTarFileInfo.IsDir() {
					logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: not a file\n", archiveMetaOrTarFilePath)
					if err := os.RemoveAll(archiveMetaOrTarFilePath); err != nil {
						return nil, fmt.Errorf("unable to remove %q: %w", archiveMetaOrTarFilePath, err)
					}
					continue
				}

				if strings.HasSuffix(archiveMetaOrTarFileInfo.Name(), ".meta.json") {
					desc := &GitArchiveDesc{MetadataPath: archiveMetaOrTarFilePath, CacheBasePath: cacheVersionRoot}

					data, err := ioutil.ReadFile(archiveMetaOrTarFilePath)
					if err != nil {
						return nil, fmt.Errorf("error reading metadata file %q: %w", archiveMetaOrTarFilePath, err)
					}
					if err := json.Unmarshal(data, &desc.Metadata); err != nil {
						logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: unable to unmarshal json: %w\n", archiveMetaOrTarFilePath, err)
						if err := os.RemoveAll(archiveMetaOrTarFilePath); err != nil {
							return nil, fmt.Errorf("unable to remove %q: %w", archiveMetaOrTarFilePath, err)
						}
						continue
					}

					archivePath := filepath.Join(hashPrefixDir, fmt.Sprintf("%s.tar", strings.TrimSuffix(archiveMetaOrTarFileInfo.Name(), ".meta.json")))

					archiveInfo, err := os.Stat(archivePath)
					if err != nil {
						if os.IsNotExist(err) {
							logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: archive file does not exist\n", archivePath)
							if err := os.RemoveAll(archiveMetaOrTarFilePath); err != nil {
								return nil, fmt.Errorf("unable to remove %q: %w", archiveMetaOrTarFilePath, err)
							}
							continue
						}
						return nil, fmt.Errorf("error accessing %q: %w", archivePath, err)
					}

					desc.ArchivePath = archivePath
					desc.Size = uint64(archiveInfo.Size())
					res = append(res, desc)
				} else if strings.HasSuffix(archiveMetaOrTarFileInfo.Name(), ".tar") {
					// This is a valid tar file, do nothing.
				} else {
					logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: unknown file type\n", archiveMetaOrTarFilePath)
					if err := os.RemoveAll(archiveMetaOrTarFilePath); err != nil {
						return nil, fmt.Errorf("unable to remove %q: %w", archiveMetaOrTarFilePath, err)
					}
				}
			}
		}
	}

	return res, nil
}
