package deploy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/werf/werf/pkg/deploy/werf_chart"
	"github.com/werf/werf/pkg/git_repo"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/deploy/secret"
)

func GetSafeSecretManager(ctx context.Context, projectDir, helmChartDir string, secretValues []string, localGitRepo *git_repo.Local, disableDeterminism bool, ignoreSecretKey bool) (secret.Manager, error) {
	isSecretsExists := false

	if disableDeterminism || localGitRepo == nil {
		if _, err := os.Stat(filepath.Join(helmChartDir, werf_chart.SecretDirName)); !os.IsNotExist(err) {
			isSecretsExists = true
		}
		if _, err := os.Stat(filepath.Join(helmChartDir, werf_chart.DefaultSecretValuesFileName)); !os.IsNotExist(err) {
			isSecretsExists = true
		}
	} else {
		commit, err := localGitRepo.HeadCommit(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to get local repo head commit: %s", err)
		}

		if exists, err := localGitRepo.IsDirectoryExists(ctx, filepath.Join(helmChartDir, werf_chart.SecretDirName), commit); err != nil {
			return nil, fmt.Errorf("error checking existance of the directory %q in the local git repo commit %s: %s", filepath.Join(helmChartDir, werf_chart.SecretDirName), err)
		} else if exists {
			isSecretsExists = true
		}

		if exists, err := localGitRepo.IsFileExists(ctx, commit, filepath.Join(helmChartDir, werf_chart.DefaultSecretValuesFileName)); err != nil {
			return nil, fmt.Errorf("error checking existance of the file %q in the local git repo commit %s: %s", filepath.Join(helmChartDir, werf_chart.DefaultSecretValuesFileName), err)
		} else if exists {
			isSecretsExists = true
		}
	}

	if len(secretValues) > 0 {
		isSecretsExists = true
	}

	if isSecretsExists {
		if ignoreSecretKey {
			logboek.Context(ctx).Default().LogLnDetails("Secrets decryption disabled")
			return secret.NewSafeManager()
		}

		key, err := secret.GetSecretKey(projectDir)
		if err != nil {
			return nil, err
		}

		return secret.NewManager(key)
	} else {
		return secret.NewSafeManager()
	}
}
