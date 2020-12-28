package werf_chart

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"sigs.k8s.io/yaml"

	"github.com/werf/werf/pkg/deploy/secret"
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/giterminism_inspector"
)

func DecodeSecretValuesFileFromGitCommit(ctx context.Context, m secret.Manager, localGitRepo git_repo.Local, commit string, projectDir string, relPath string) (map[string]interface{}, error) {
	var data []byte
	if accepted, err := giterminism_inspector.IsHelmUncommittedFileAccepted(relPath); err != nil {
	} else if accepted {
		data, err = ioutil.ReadFile(filepath.Join(projectDir, relPath))
		if err != nil {
			return nil, err
		}
	} else {
		data, err = git_repo.ReadCommitFileAndCompareWithProjectFile(ctx, localGitRepo, commit, projectDir, relPath)
		if err != nil {
			return nil, err
		}
	}

	decodedData, err := m.DecryptYamlData(data)
	if err != nil {
		return nil, fmt.Errorf("cannot decode file %q secret data: %s", relPath, err)
	}

	rawValues := map[string]interface{}{}
	if err := yaml.Unmarshal(decodedData, &rawValues); err != nil {
		return nil, fmt.Errorf("cannot unmarshal secret values file %s: %s", relPath, err)
	}

	return rawValues, nil
}

func DecodeSecretValuesFileFromFilesystem(ctx context.Context, path string, m secret.Manager) (map[string]interface{}, error) {
	var data []byte

	if d, err := ioutil.ReadFile(path); err != nil {
		return nil, fmt.Errorf("cannot read file %q: %s", path, err)
	} else {
		data = d
	}

	decodedData, err := m.DecryptYamlData(data)
	if err != nil {
		return nil, fmt.Errorf("cannot decode file %q secret data: %s", path, err)
	}

	rawValues := map[string]interface{}{}
	if err := yaml.Unmarshal(decodedData, &rawValues); err != nil {
		return nil, fmt.Errorf("cannot unmarshal secret values file %s: %s", path, err)
	}

	return rawValues, nil
}
