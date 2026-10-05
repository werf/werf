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

type GitPatchDesc struct {
	MetadataPath  string
	PatchPath     string
	SidecarPaths  []string
	Metadata      *PatchMetadata
	Size          uint64
	CacheBasePath string
}

func (entry *GitPatchDesc) GetPaths() []string {
	return append([]string{entry.MetadataPath, entry.PatchPath}, entry.SidecarPaths...)
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

const (
	patchMetadataSuffix = ".meta.json"
	patchPayloadSuffix  = ".patch"
)

func patchSidecarParentName(name string) (string, bool) {
	if !strings.HasSuffix(name, ".paths_list") && !strings.HasSuffix(name, ".archive") {
		return "", false
	}

	ind := strings.Index(name, patchPayloadSuffix+".")
	if ind < 0 {
		return "", false
	}

	return name[:ind+len(patchPayloadSuffix)], true
}

// GetGitPatchesAndRemoveInvalid scans the given cacheVersionRoot directory and returns
// a list of GitPatchDesc for each valid .meta.json file found. It removes invalid
// entries and handles errors appropriately.
//
// An entry is the metadata file, its patch payload and every sidecar file of that
// payload: they are created and reused together, so GC removes them together. Any
// file left without a valid metadata-plus-payload pair is removed as garbage.
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
//
// An entry whose metadata cannot be read is preserved with its payload and
// sidecars and left out of the result; its error is joined into the returned
// error.
func GetGitPatchesAndRemoveInvalid(ctx context.Context, cacheVersionRoot string, options ScanOptions) ([]GitDataEntry, error) {
	var res []GitDataEntry
	var errs []error

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
			if err := removePath(repoDir, options); err != nil {
				return nil, fmt.Errorf("unable to remove %q: %w", repoDir, err)
			}
			continue
		}

		hashGroupDirs, err := ioutil.ReadDir(repoDir)
		if err != nil {
			errs = append(errs, fmt.Errorf("read repo patches dir %q: %w", repoDir, err))
			continue
		}

		for _, hashGroupDirInfo := range hashGroupDirs {
			hashGroupDir := filepath.Join(repoDir, hashGroupDirInfo.Name())

			if !hashGroupDirInfo.IsDir() {
				logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: not a directory\n", hashGroupDir)
				if err := removePath(hashGroupDir, options); err != nil {
					return nil, fmt.Errorf("unable to remove %q: %w", hashGroupDir, err)
				}
				continue
			}

			patchFiles, err := ioutil.ReadDir(hashGroupDir)
			if err != nil {
				errs = append(errs, fmt.Errorf("read repo patches from dir %q: %w", hashGroupDir, err))
				continue
			}

			regularFiles := make(map[string]os.FileInfo, len(patchFiles))
			sidecarNames := make(map[string][]string)
			for _, fileInfo := range patchFiles {
				filePath := filepath.Join(hashGroupDir, fileInfo.Name())

				if !fileInfo.Mode().IsRegular() {
					logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: not a regular file\n", filePath)
					if err := removePath(filePath, options); err != nil {
						return nil, fmt.Errorf("unable to remove %q: %w", filePath, err)
					}
					continue
				}

				regularFiles[fileInfo.Name()] = fileInfo

				if parentName, ok := patchSidecarParentName(fileInfo.Name()); ok {
					sidecarNames[parentName] = append(sidecarNames[parentName], fileInfo.Name())
				}
			}

			keptNames := make(map[string]bool, len(regularFiles))

			for _, fileInfo := range patchFiles {
				name := fileInfo.Name()
				if _, ok := regularFiles[name]; !ok {
					continue
				}
				if !strings.HasSuffix(name, patchMetadataSuffix) {
					continue
				}

				metadataPath := filepath.Join(hashGroupDir, name)
				payloadName := strings.TrimSuffix(name, patchMetadataSuffix) + patchPayloadSuffix
				desc := &GitPatchDesc{MetadataPath: metadataPath, CacheBasePath: cacheVersionRoot}

				data, err := ioutil.ReadFile(metadataPath)
				if err != nil {
					errs = append(errs, fmt.Errorf("read metadata file %q: %w", metadataPath, err))
					keptNames[name] = true
					keptNames[payloadName] = true
					for _, sidecarName := range sidecarNames[payloadName] {
						keptNames[sidecarName] = true
					}
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
					logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: patch file does not exist\n", filepath.Join(hashGroupDir, payloadName))
					continue
				}

				desc.PatchPath = filepath.Join(hashGroupDir, payloadName)
				desc.Size = uint64(payloadInfo.Size())
				keptNames[name] = true
				keptNames[payloadName] = true

				for _, sidecarName := range sidecarNames[payloadName] {
					desc.SidecarPaths = append(desc.SidecarPaths, filepath.Join(hashGroupDir, sidecarName))
					desc.Size += uint64(regularFiles[sidecarName].Size())
					keptNames[sidecarName] = true
				}

				res = append(res, desc)
			}

			for _, fileInfo := range patchFiles {
				name := fileInfo.Name()
				if _, ok := regularFiles[name]; !ok || keptNames[name] {
					continue
				}

				filePath := filepath.Join(hashGroupDir, name)
				logboek.Context(ctx).Warn().LogF("Removing invalid entry %q: no valid patch entry owns it\n", filePath)
				if err := removePath(filePath, options); err != nil {
					return res, fmt.Errorf("unable to remove %q: %w", filePath, err)
				}
			}
		}
	}

	return res, errors.Join(errs...)
}
