package gitdata

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
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

func GetExistingGitArchives(cacheVersionRoot string) ([]*GitArchiveDesc, error) {
	var res []*GitArchiveDesc

	if _, err := os.Stat(cacheVersionRoot); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("error accessing dir %q: %w", cacheVersionRoot, err)
	}

	files, err := ioutil.ReadDir(cacheVersionRoot)
	if err != nil {
		return nil, fmt.Errorf("error reading dir %q: %w", cacheVersionRoot, err)
	}

	for _, finfo := range files {
		repoArchivesRootDir := filepath.Join(cacheVersionRoot, finfo.Name())

		if !finfo.IsDir() {
			if err := os.RemoveAll(repoArchivesRootDir); err != nil {
				return nil, fmt.Errorf("unable to remove %q: %w", repoArchivesRootDir, err)
			}
		}

		hashDirs, err := ioutil.ReadDir(repoArchivesRootDir)
		if err != nil {
			return nil, fmt.Errorf("error reading repo archives dir %q: %w", repoArchivesRootDir, err)
		}

		for _, finfo := range hashDirs {
			hashDir := filepath.Join(repoArchivesRootDir, finfo.Name())

			if !finfo.IsDir() {
				if err := os.RemoveAll(hashDir); err != nil {
					return nil, fmt.Errorf("unable to remove %q: %w", hashDir, err)
				}
			}

			archivesFiles, err := ioutil.ReadDir(hashDir)
			if err != nil {
				return nil, fmt.Errorf("error reading repo archives from dir %q: %w", hashDir, err)
			}

			for _, finfo := range archivesFiles {
				path := filepath.Join(hashDir, finfo.Name())

				if finfo.IsDir() {
					if err := os.RemoveAll(path); err != nil {
						return nil, fmt.Errorf("unable to remove %q: %w", path, err)
					}
				}

				if !strings.HasSuffix(finfo.Name(), ".meta.json") {
					continue
				}

				desc := &GitArchiveDesc{MetadataPath: path, CacheBasePath: cacheVersionRoot}
				res = append(res, desc)

				if data, err := ioutil.ReadFile(path); err != nil {
					return nil, fmt.Errorf("error reading metadata file %q: %w", path, err)
				} else {
					if err := json.Unmarshal(data, &desc.Metadata); err != nil {
						return nil, fmt.Errorf("error unmarshalling json from %q: %w", path, err)
					}
				}

				archivePath := filepath.Join(hashDir, fmt.Sprintf("%s.tar", strings.TrimSuffix(finfo.Name(), ".meta.json")))

				archiveInfo, err := os.Stat(archivePath)
				if err != nil {
					if os.IsNotExist(err) {
						continue
					}

					return nil, fmt.Errorf("error accessing %q: %w", archivePath, err)
				}

				desc.ArchivePath = archivePath
				desc.Size = uint64(archiveInfo.Size())
			}
		}
	}

	return res, nil
}
