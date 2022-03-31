package buildah

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/werf/werf/pkg/buildah/thirdparty"
	"github.com/werf/werf/pkg/util"
)

type BaseBuildah struct {
	Isolation               thirdparty.Isolation
	TmpDir                  string
	InstanceTmpDir          string
	ConfigTmpDir            string
	SignaturePolicyPath     string
	RegistriesConfigPath    string
	RegistriesConfigDirPath string
	Insecure                bool
}

type BaseBuildahOpts struct {
	Isolation thirdparty.Isolation
	Insecure  bool
}

func NewBaseBuildah(tmpDir string, opts BaseBuildahOpts) (*BaseBuildah, error) {
	b := &BaseBuildah{
		Isolation: opts.Isolation,
		TmpDir:    tmpDir,
		Insecure:  opts.Insecure,
	}

	if err := os.MkdirAll(b.TmpDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create dir %q: %w", b.TmpDir, err)
	}

	var err error
	b.InstanceTmpDir, err = ioutil.TempDir(b.TmpDir, "instance")
	if err != nil {
		return nil, fmt.Errorf("unable to create instance tmp dir: %w", err)
	}

	b.ConfigTmpDir = filepath.Join(b.InstanceTmpDir, "config")
	if err := os.MkdirAll(b.ConfigTmpDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create dir %q: %w", b.ConfigTmpDir, err)
	}

	b.SignaturePolicyPath = filepath.Join(b.ConfigTmpDir, "policy.json")
	if err := ioutil.WriteFile(b.SignaturePolicyPath, []byte(DefaultSignaturePolicy), os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to write file %q: %w", b.SignaturePolicyPath, err)
	}

	b.RegistriesConfigPath = filepath.Join(b.ConfigTmpDir, "registries.conf")
	if err := ioutil.WriteFile(b.RegistriesConfigPath, []byte(DefaultRegistriesConfig), os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to write file %q: %w", b.RegistriesConfigPath, err)
	}

	b.RegistriesConfigDirPath = filepath.Join(b.ConfigTmpDir, "registries.conf.d")
	if err := os.MkdirAll(b.RegistriesConfigDirPath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create dir %q: %w", b.RegistriesConfigDirPath, err)
	}

	return b, nil
}

func (b *BaseBuildah) NewSessionTmpDir() (string, error) {
	sessionTmpDir, err := ioutil.TempDir(b.TmpDir, "session")
	if err != nil {
		return "", fmt.Errorf("unable to create session tmp dir: %w", err)
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
		return "", "", "", fmt.Errorf("error writing %q: %w", dockerfileTmpPath, err)
	}

	contextTmpDir := filepath.Join(sessionTmpDir, "context")
	if err := os.MkdirAll(contextTmpDir, os.ModePerm); err != nil {
		return "", "", "", fmt.Errorf("unable to create dir %q: %w", contextTmpDir, err)
	}

	if contextTar != nil {
		if err := util.ExtractTar(contextTar, contextTmpDir); err != nil {
			return "", "", "", fmt.Errorf("unable to extract context tar to tmp context dir: %w", err)
		}
	}

	return sessionTmpDir, contextTmpDir, dockerfileTmpPath, nil
}
