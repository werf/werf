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

type GitPatchDesc struct {
	MetadataPath  string
	PatchPath     string
	Metadata      *PatchMetadata
	Size          uint64
	CacheBasePath string
}

func (entry *GitPatchDesc) GetPaths() []string {
	return []string{entry.MetadataPath, entry.PatchPath}
}

func (entry *GitPatchDesc) GetSize() uint64 {
	return entry.Size
}

func (entry *GitPatchDesc) GetLastAccessAt() time.Time {
	return time.Unix(entry.Metadata.LastAccessTimestamp, 0)
}

func (entry *GitPatchDesc) GetCacheBasePath() string {
	return entry.CacheBasePath
}

func GetExistingGitPatches(cacheVersionRoot string) ([]*GitPatchDesc, error) {
	var res []*GitPatchDesc

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
		repoPatchesRootDir := filepath.Join(cacheVersionRoot, finfo.Name())

		if !finfo.IsDir() {
			if err := os.RemoveAll(repoPatchesRootDir); err != nil {
				return nil, fmt.Errorf("unable to remove %q: %w", repoPatchesRootDir, err)
			}
		}

		hashDirs, err := ioutil.ReadDir(repoPatchesRootDir)
		if err != nil {
			return nil, fmt.Errorf("error reading repo archives dir %q: %w", repoPatchesRootDir, err)
		}

		for _, finfo := range hashDirs {
			hashDir := filepath.Join(repoPatchesRootDir, finfo.Name())

			if !finfo.IsDir() {
				if err := os.RemoveAll(hashDir); err != nil {
					return nil, fmt.Errorf("unable to remove %q: %w", hashDir, err)
				}
			}

			patchesFiles, err := ioutil.ReadDir(hashDir)
			if err != nil {
				return nil, fmt.Errorf("error reading repo patches from dir %q: %w", hashDir, err)
			}

			for _, finfo := range patchesFiles {
				path := filepath.Join(hashDir, finfo.Name())

				if finfo.IsDir() {
					if err := os.RemoveAll(path); err != nil {
						return nil, fmt.Errorf("unable to remove %q: %w", path, err)
					}
				}

				if !strings.HasSuffix(finfo.Name(), ".meta.json") {
					continue
				}

				desc := &GitPatchDesc{MetadataPath: path}
				res = append(res, desc)

				if data, err := ioutil.ReadFile(path); err != nil {
					return nil, fmt.Errorf("error reading metadata file %q: %w", path, err)
				} else {
					if err := json.Unmarshal(data, &desc.Metadata); err != nil {
						return nil, fmt.Errorf("error unmarshalling json from %q: %w", path, err)
					}
				}

				patchPath := filepath.Join(hashDir, fmt.Sprintf("%s.patch", strings.TrimSuffix(finfo.Name(), ".meta.json")))

				patchInfo, err := os.Stat(patchPath)
				if err != nil {
					if os.IsNotExist(err) {
						continue
					}

					return nil, fmt.Errorf("error accessing %q: %w", patchPath, err)
				}

				desc.PatchPath = patchPath
				desc.Size = uint64(patchInfo.Size())
			}
		}
	}

	return res, nil
}
