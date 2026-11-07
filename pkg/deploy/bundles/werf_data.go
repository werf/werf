package bundles

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/samber/lo"

	"github.com/werf/nelm/v2/pkg/action"
	"github.com/werf/nelm/v2/pkg/helm/pkg/chart/chartyaml"
	chartcommon "github.com/werf/nelm/v2/pkg/helm/pkg/chart/common"
	"github.com/werf/nelm/v2/pkg/helm/pkg/chart/loader/archive"
	"github.com/werf/nelm/v2/pkg/werfdata"
)

// ResolveRelease fills in the release name and namespace not given with the ones werf bundle publish
// recorded in the bundle at chartPath. They are recorded for the environment of the publication,
// werf.env of the .werf-data extra values, and converge would resolve them differently for another
// one, so with another non-empty env the missing ones are an error. What is neither given nor
// recorded is returned empty.
func ResolveRelease(ctx context.Context, chartPath, release, namespace, env string) (string, string, error) {
	if env != "" && (release == "" || namespace == "") {
		if err := checkRecordedReleaseEnv(chartPath, release, namespace, env); err != nil {
			return "", "", err
		}
	}

	release, namespace, err := action.ResolveRecordedRelease(ctx, chartPath, release, namespace, action.ResolveRecordedReleaseOptions{})
	if err != nil {
		return "", "", fmt.Errorf("resolve release: %w", err)
	}

	return release, namespace, nil
}

func checkRecordedReleaseEnv(chartPath, release, namespace, env string) error {
	files, err := readWerfDataFiles(chartPath)
	if err != nil {
		return fmt.Errorf("read %s of chart %q: %w", werfdata.DirName, chartPath, err)
	}

	if len(files) == 0 {
		return nil
	}

	data, err := werfdata.Parse(files)
	if err != nil {
		return fmt.Errorf("parse %s: %w", werfdata.DirName, err)
	}

	if data.Metadata.Release == nil {
		return nil
	}

	werfVals, _ := data.ExtraValues["werf"].(map[string]any)
	recordedEnv, _ := werfVals["env"].(string)
	if recordedEnv == env {
		return nil
	}

	var missing []string
	if release == "" {
		missing = append(missing, "--release")
	}

	if namespace == "" {
		missing = append(missing, "--namespace")
	}

	return fmt.Errorf("release defaults recorded for env %q are not used with --env %q: pass %s", recordedEnv, env, strings.Join(missing, " and "))
}

// readWerfDataFiles reads the .werf-data files of the chart directory or archive at chartPath,
// without loading the rest of the chart.
func readWerfDataFiles(chartPath string) ([]*chartcommon.File, error) {
	info, err := os.Stat(chartPath)
	if err != nil {
		return nil, fmt.Errorf("stat chart: %w", err)
	}

	var files []*chartcommon.File
	if info.IsDir() {
		werfDataDir := filepath.Join(chartPath, werfdata.DirName)
		if err := filepath.WalkDir(werfDataDir, func(path string, entry fs.DirEntry, err error) error {
			if errors.Is(err, fs.ErrNotExist) && path == werfDataDir {
				return fs.SkipDir
			} else if err != nil {
				return err
			}

			if entry.IsDir() {
				return nil
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("read %q: %w", path, err)
			}

			name, err := filepath.Rel(chartPath, path)
			if err != nil {
				return fmt.Errorf("get relative path of %q: %w", path, err)
			}

			files = append(files, &chartcommon.File{Name: filepath.ToSlash(name), Data: data})

			return nil
		}); err != nil {
			return nil, err
		}

		return files, nil
	}

	archiveFile, err := os.Open(chartPath)
	if err != nil {
		return nil, fmt.Errorf("open chart archive: %w", err)
	}
	defer archiveFile.Close()

	archiveFiles, err := archive.LoadArchiveFiles(archiveFile)
	if err != nil {
		return nil, fmt.Errorf("read chart archive: %w", err)
	}

	for _, f := range archiveFiles {
		if werfdata.IsWerfDataFile(f.Name) {
			files = append(files, &chartcommon.File{Name: f.Name, ModTime: f.ModTime, Data: f.Data})
		}
	}

	return files, nil
}

// ChartImagesFromServiceValues returns the helm.sh/images entries of the werf images in the service
// values.
func ChartImagesFromServiceValues(ctx context.Context, serviceValues map[string]interface{}) []chartyaml.Image {
	globalVals, _ := serviceValues["global"].(map[string]interface{})
	werfVals, _ := globalVals["werf"].(map[string]interface{})
	imagesVals, _ := werfVals["images"].(map[string]interface{})

	names := lo.Keys(imagesVals)
	sort.Strings(names)

	images := []chartyaml.Image{}
	for _, name := range names {
		imageVals, ok := imagesVals[name].(map[string]interface{})
		if !ok {
			continue
		}

		imageRef, _ := imageVals["ref"].(string)
		if digest, _ := imageVals["digest"].(string); digest == "" {
			imageRef, _ = imageVals["ref_tag"].(string)
		}

		images = append(images, chartyaml.Image{Name: name, Image: imageRef})
	}

	return images
}
