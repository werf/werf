package gitdata

import (
	"context"
	"encoding/json"
	"errors"
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

const (
	archiveMetadataSuffix = ".meta.json"
	archivePayloadSuffix  = ".tar"
)

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
//
// An entry whose metadata cannot be read is preserved with its payload and
// left out of the result; its error is joined into the returned error.
func GetGitArchivesAndRemoveInvalid(ctx context.Context, cacheVersionRoot string, options ScanOptions) ([]GitDataEntry, error) {
	var res []GitDataEntry
	var errs []error

	fileStat, err := os.Stat(cacheVersionRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error accessing dir %q: %w", cacheVersionRoot, err)
	}
	if !fileStat.IsDir() {
		logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: not a directory\n", cacheVersionRoot)
		if err := removePath(cacheVersionRoot, options); err != nil {
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
			if err := removePath(repoHashDir, options); err != nil {
				return nil, fmt.Errorf("unable to remove %q: %w", repoHashDir, err)
			}
			continue
		}

		hashPrefixes, err := ioutil.ReadDir(repoHashDir)
		if err != nil {
			errs = append(errs, fmt.Errorf("read repo archives dir %q: %w", repoHashDir, err))
			continue
		}

		for _, hashPrefixInfo := range hashPrefixes {
			hashPrefixDir := filepath.Join(repoHashDir, hashPrefixInfo.Name())

			if !hashPrefixInfo.IsDir() {
				logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: not a directory\n", hashPrefixDir)
				if err := removePath(hashPrefixDir, options); err != nil {
					return nil, fmt.Errorf("unable to remove %q: %w", hashPrefixDir, err)
				}
				continue
			}

			archiveFiles, err := ioutil.ReadDir(hashPrefixDir)
			if err != nil {
				errs = append(errs, fmt.Errorf("read repo archives from dir %q: %w", hashPrefixDir, err))
				continue
			}

			regularFiles := make(map[string]os.FileInfo, len(archiveFiles))
			for _, fileInfo := range archiveFiles {
				filePath := filepath.Join(hashPrefixDir, fileInfo.Name())

				if !fileInfo.Mode().IsRegular() {
					logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: not a regular file\n", filePath)
					if err := removePath(filePath, options); err != nil {
						return nil, fmt.Errorf("unable to remove %q: %w", filePath, err)
					}
					continue
				}

				regularFiles[fileInfo.Name()] = fileInfo
			}

			keptNames := make(map[string]bool, len(regularFiles))

			for _, fileInfo := range archiveFiles {
				name := fileInfo.Name()
				if _, ok := regularFiles[name]; !ok {
					continue
				}
				if !strings.HasSuffix(name, archiveMetadataSuffix) {
					continue
				}

				metadataPath := filepath.Join(hashPrefixDir, name)
				payloadName := strings.TrimSuffix(name, archiveMetadataSuffix) + archivePayloadSuffix
				desc := &GitArchiveDesc{MetadataPath: metadataPath, CacheBasePath: cacheVersionRoot}

				data, err := ioutil.ReadFile(metadataPath)
				if err != nil {
					errs = append(errs, fmt.Errorf("read metadata file %q: %w", metadataPath, err))
					keptNames[name] = true
					keptNames[payloadName] = true
					continue
				}

				if err := json.Unmarshal(data, &desc.Metadata); err != nil {
					logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: unable to unmarshal json: %s\n", metadataPath, err)
					continue
				}
				if desc.Metadata == nil {
					logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: empty metadata\n", metadataPath)
					continue
				}

				payloadInfo, ok := regularFiles[payloadName]
				if !ok {
					logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: archive file does not exist\n", filepath.Join(hashPrefixDir, payloadName))
					continue
				}

				desc.ArchivePath = filepath.Join(hashPrefixDir, payloadName)
				desc.Size = uint64(payloadInfo.Size())
				keptNames[name] = true
				keptNames[payloadName] = true

				res = append(res, desc)
			}

			for _, fileInfo := range archiveFiles {
				name := fileInfo.Name()
				if _, ok := regularFiles[name]; !ok || keptNames[name] {
					continue
				}

				filePath := filepath.Join(hashPrefixDir, name)
				logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: no valid archive entry owns it\n", filePath)
				if err := removePath(filePath, options); err != nil {
					return res, fmt.Errorf("unable to remove %q: %w", filePath, err)
				}
			}
		}
	}

	return res, errors.Join(errs...)
}
