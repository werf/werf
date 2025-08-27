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

// GetGitPatchesAndRemoveInvalid scans the given cacheVersionRoot directory and returns
// a list of GitPatchDesc for each valid .meta.json file found. It removes invalid
// entries and handles errors appropriately.
//
// The directory structure expected is as follows:
// ├── 0f1ddce0c13406a1178a3e8df39e356fb0ab629e7b3f3db26f04cb668a2c3b2a/
// │   ├── a3/
// │   │   ├── a3be8a34b216b93516c0f50964a15cace97662cf20a278bb3f50f511649249bb.patch.92f0dd52eb4e3cc2deb6761be83a42fa9d1d07e1c6476a5ac2c2ba9e62b43c10.paths_list
// │   │   ├── a3d24f9f2203e37ce6400f8198246e1a8d28a69728e8873e785d0e9adbf1e85e.meta.json
// │   │   ├── a3d24f9f2203e37ce6400f8198246e1a8d28a69728e8873e785d0e9adbf1e85e.patch
// │   │   └── ... (other patch files)
// │   └── ... (other hash groups)
// └── ... (other repositories)
func GetGitPatchesAndRemoveInvalid(ctx context.Context, cacheVersionRoot string) ([]GitDataEntry, error) {
	var res []GitDataEntry

	if _, err := os.Stat(cacheVersionRoot); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("error accessing dir %q: %w", cacheVersionRoot, err)
	}

	repoDirs, err := ioutil.ReadDir(cacheVersionRoot)
	if err != nil {
		return nil, fmt.Errorf("error reading dir %q: %w", cacheVersionRoot, err)
	}

	for _, repoDirInfo := range repoDirs {
		repoDir := filepath.Join(cacheVersionRoot, repoDirInfo.Name())

		if !repoDirInfo.IsDir() {
			logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: not a directory\n", repoDir)
			if err := os.RemoveAll(repoDir); err != nil {
				return nil, fmt.Errorf("unable to remove %q: %w", repoDir, err)
			}
			continue
		}

		hashGroupDirs, err := ioutil.ReadDir(repoDir)
		if err != nil {
			return nil, fmt.Errorf("error reading repo archives dir %q: %w", repoDir, err)
		}

		for _, hashGroupDirInfo := range hashGroupDirs {
			hashGroupDir := filepath.Join(repoDir, hashGroupDirInfo.Name())

			if !hashGroupDirInfo.IsDir() {
				logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: not a directory\n", hashGroupDir)
				if err := os.RemoveAll(hashGroupDir); err != nil {
					return nil, fmt.Errorf("unable to remove %q: %w", hashGroupDir, err)
				}
				continue
			}

			patchFiles, err := ioutil.ReadDir(hashGroupDir)
			if err != nil {
				return nil, fmt.Errorf("error reading repo patches from dir %q: %w", hashGroupDir, err)
			}

			for _, metaOrPatchFileInfo := range patchFiles {
				metaOrPathFilePath := filepath.Join(hashGroupDir, metaOrPatchFileInfo.Name())

				// Remove invalid entry: file must be a regular file.
				if metaOrPatchFileInfo.IsDir() {
					logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: not a regular file\n", metaOrPathFilePath)
					if err := os.RemoveAll(metaOrPathFilePath); err != nil {
						return nil, fmt.Errorf("unable to remove %q: %w", metaOrPathFilePath, err)
					}

					continue
				}

				if strings.HasSuffix(metaOrPatchFileInfo.Name(), ".meta.json") {
					desc := &GitPatchDesc{MetadataPath: metaOrPathFilePath}

					data, err := ioutil.ReadFile(metaOrPathFilePath)
					if err != nil {
						return nil, fmt.Errorf("error reading metadata file %q: %w", metaOrPathFilePath, err)
					}

					err = json.Unmarshal(data, &desc.Metadata)
					if err != nil {
						logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: unable to unmarshal json: %w\n", metaOrPathFilePath, err)
						if err := os.RemoveAll(metaOrPathFilePath); err != nil {
							return nil, fmt.Errorf("unable to remove %q: %w", metaOrPathFilePath, err)
						}
						continue
					}

					patchPath := filepath.Join(hashGroupDir, fmt.Sprintf("%s.patch", strings.TrimSuffix(metaOrPatchFileInfo.Name(), ".meta.json")))
					patchInfo, err := os.Stat(patchPath)
					if err != nil {
						if os.IsNotExist(err) {
							logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: patch file does not exist\n", patchPath)
							if err := os.RemoveAll(metaOrPathFilePath); err != nil {
								return nil, fmt.Errorf("unable to remove %q: %w", metaOrPathFilePath, err)
							}

							continue
						}
						return nil, fmt.Errorf("error accessing %q: %w", patchPath, err)
					}

					desc.PatchPath = patchPath
					desc.Size = uint64(patchInfo.Size())
					res = append(res, desc)
				} else if strings.HasSuffix(metaOrPatchFileInfo.Name(), ".paths_list") {
				} else if strings.HasSuffix(metaOrPatchFileInfo.Name(), ".patch") {
				} else {
					logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: unknown file type\n", metaOrPathFilePath)
					if err := os.RemoveAll(metaOrPathFilePath); err != nil {
						return nil, fmt.Errorf("unable to remove %q: %w", metaOrPathFilePath, err)
					}
				}
			}
		}
	}

	return res, nil
}
