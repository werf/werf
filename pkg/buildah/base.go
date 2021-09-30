package buildah

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/werf/werf/pkg/util"
)

type BaseBuildah struct {
	TmpDir string
}

func NewBaseBuildah(tmpDir string) (*BaseBuildah, error) {
	b := &BaseBuildah{
		TmpDir: tmpDir,
	}

	if err := os.MkdirAll(b.TmpDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create dir %q: %s", b.TmpDir, err)
	}

	return b, nil
}

func (b *BaseBuildah) NewSessionTmpDir() (string, error) {
	sessionTmpDir, err := ioutil.TempDir(b.TmpDir, "session")
	if err != nil {
		return "", fmt.Errorf("unable to create session tmp dir: %s", err)
	}

	return sessionTmpDir, nil
}

func (b *BaseBuildah) prepareBuildFromDockerfile(dockerfile []byte, contextTar io.Reader) (string, string, string, error) {
	sessionTmpDir, err := b.NewSessionTmpDir()
	if err != nil {
		return "", "", "", err
	}

	dockerfileTmpPath := filepath.Join(sessionTmpDir, "Dockerfile")
	if err := ioutil.WriteFile(dockerfileTmpPath, dockerfile, os.ModePerm); err != nil {
		return "", "", "", fmt.Errorf("error writing %q: %s", dockerfileTmpPath, err)
	}

	contextTmpDir := filepath.Join(sessionTmpDir, "context")
	if err := os.MkdirAll(contextTmpDir, os.ModePerm); err != nil {
		return "", "", "", fmt.Errorf("unable to create dir %q: %s", contextTmpDir, err)
	}

	if contextTar != nil {
		if err := util.ExtractTar(contextTar, contextTmpDir); err != nil {
			return "", "", "", fmt.Errorf("unable to extract context tar to tmp context dir: %s", err)
		}
	}

	return sessionTmpDir, contextTmpDir, dockerfileTmpPath, nil
}
