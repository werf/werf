package tmp_manager

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func CreateKubeConfigFromBase64(ctx context.Context, base64KubeConfig io.Reader) (string, error) {
	base64Decoder := base64.NewDecoder(base64.StdEncoding, base64KubeConfig)
	kubeConfig, err := ioutil.ReadAll(base64Decoder)
	if err != nil {
		return "", fmt.Errorf("unable to base64 decode kubeconfig: %w", err)
	}

	tmpDir, err := newTmpDir(KubeConfigDirPrefix)
	if err != nil {
		return "", fmt.Errorf("unable to create kubeconfig tmp dir: %w", err)
	}

	if err := os.Chmod(tmpDir, 0o700); err != nil {
		return "", fmt.Errorf("unable to create tmp kubeconfigs service dir: %w", err)
	}

	kubeConfigPath := filepath.Join(tmpDir, "kubeconfig")
	if err := os.WriteFile(kubeConfigPath, kubeConfig, 0o600); err != nil {
		return "", fmt.Errorf("unable to write file kubeconfig: %w", err)
	}

	if err := registerCreatedPath(tmpDir, filepath.Join(GetCreatedTmpDirs(), kubeConfigsServiceDir)); err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("unable to register created kubeconfigs service dir: %w", err)
	}

	return kubeConfigPath, nil
}
