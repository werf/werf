package buildah

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/docker"
	"github.com/werf/werf/pkg/werf"
)

const (
	BuildahImage = "quay.io/buildah/stable:v1.22.3@sha256:b551986d37a2c097749220976da630305eaba0f03de705dddc28007d6cc47ab1"
)

type DockerWithFuseBuildah struct {
	HostStorageDir string
	HostTmpDir     string
}

func NewDockerWithFuseBuildah() (*DockerWithFuseBuildah, error) {
	b := &DockerWithFuseBuildah{
		HostStorageDir: filepath.Join(werf.GetHomeDir(), "buildah", "storage"),
		HostTmpDir:     filepath.Join(werf.GetHomeDir(), "buildah", "tmp"),
	}

	if err := os.MkdirAll(b.HostStorageDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create dir %q: %s", b.HostStorageDir, err)
	}

	if err := os.MkdirAll(b.HostTmpDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create dir %q: %s", b.HostTmpDir, err)
	}

	return b, nil
}

func (b *DockerWithFuseBuildah) BuildFromDockerfile(ctx context.Context, dockerfile []byte, opts BuildFromDockerfileOpts) (string, error) {
	sessionTmpDir, err := ioutil.TempDir(b.HostTmpDir, "werf-buildah")
	if err != nil {
		return "", fmt.Errorf("unable to prepare temp dir: %s", err)
	}
	defer func() {
		if err = os.RemoveAll(sessionTmpDir); err != nil {
			logboek.Warn().LogF("unable to remove temp dir %s: %s\n", sessionTmpDir, err)
		}
	}()

	dockerfileTmpPath := filepath.Join(sessionTmpDir, "Dockerfile")
	if err := ioutil.WriteFile(dockerfileTmpPath, dockerfile, os.ModePerm); err != nil {
		return "", fmt.Errorf("error writing %q: %s", dockerfileTmpPath, err)
	}

	contextTmpDir := filepath.Join(sessionTmpDir, "context")
	if err := os.MkdirAll(contextTmpDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("unable to create dir %q: %s", contextTmpDir, err)
	}

	if opts.ContextTar != nil {
		if err := myExtractTar(opts.ContextTar, contextTmpDir); err != nil {
			return "", fmt.Errorf("unable to extract context tar: %s", err)
		}
	}

	output, _, err := b.runBuildah(
		ctx,
		[]string{
			"--volume", fmt.Sprintf("%s:/.werf/buildah/tmp", sessionTmpDir),
			"--workdir", "/.werf/buildah/tmp/context",
		},
		[]string{"bud", "-f", "/.werf/buildah/tmp/Dockerfile"}, opts.LogWriter,
	)

	if err != nil {
		return "", err
	}

	fmt.Printf("OUTPUT:\n%s\n---\n", output)

	outputLines := scanLines(output)

	return outputLines[len(outputLines)-1], nil
}

func (b *DockerWithFuseBuildah) RunCommand(ctx context.Context, container string, command []string, opts RunCommandOpts) error {
	_, _, err := b.runBuildah(ctx, []string{}, append([]string{container, "run"}, command...), opts.LogWriter)
	return err
}

func (b *DockerWithFuseBuildah) runBuildah(ctx context.Context, dockerArgs []string, buildahArgs []string, logWriter io.Writer) (string, string, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	var stdoutWriter io.Writer
	var stderrWriter io.Writer

	if logWriter != nil {
		stdoutWriter = io.MultiWriter(stdout, logWriter)
		stderrWriter = io.MultiWriter(stderr, logWriter)
	} else {
		stdoutWriter = stdout
		stderrWriter = stderr
	}

	args := []string{"--rm"}
	args = append(args, dockerArgs...)
	args = append(args, buildahWithFuseDockerArgs(b.HostStorageDir)...)
	args = append(args, buildahArgs...)

	fmt.Printf("ARGS: %v\n", args)

	err := docker.CliRun_ProvidedOutput(ctx, stdoutWriter, stderrWriter, args...)
	return stdout.String(), stderr.String(), err
}

func buildahWithFuseDockerArgs(hostStorageDir string) []string {
	return []string{
		"--device", "/dev/fuse",
		"--security-opt", "seccomp=unconfined",
		"--security-opt", "apparmor=unconfined",
		"--volume", fmt.Sprintf("%s:/var/lib/containers/storage", hostStorageDir),
		BuildahImage, "buildah",
	}
}

func scanLines(data string) []string {
	var lines []string

	s := bufio.NewScanner(strings.NewReader(data))
	for s.Scan() {
		lines = append(lines, s.Text())
	}

	return lines
}

func myExtractTar(reader io.Reader, dstDir string) error {
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("error getting tar header: %s", err)
		}

		path := filepath.Join(dstDir, header.Name)
		info := header.FileInfo()

		if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
			return err
		}

		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return fmt.Errorf("ALO: %s", err)
		}
		defer file.Close()

		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}

	return nil
}
