package build_context

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/werf/werf/pkg/util"
)

type BuildContext struct {
	ContextTarReader io.ReadCloser
	TmpDir           string

	contextTmpDir string
}

func NewBuildContext(tmpDir string, contextTarReader io.ReadCloser) *BuildContext {
	return &BuildContext{TmpDir: tmpDir, ContextTarReader: contextTarReader}
}

func (c *BuildContext) GetContextDir(ctx context.Context) (string, error) {
	if c.contextTmpDir != "" {
		return c.contextTmpDir, nil
	}

	contextTmpDir, err := ioutil.TempDir(c.TmpDir, "context")
	if err != nil {
		return "", fmt.Errorf("unable to create context tmp dir: %w", err)
	}

	if err := util.ExtractTar(c.ContextTarReader, contextTmpDir, util.ExtractTarOptions{}); err != nil {
		return "", fmt.Errorf("unable to extract context tar to tmp context dir: %w", err)
	}
	if err := c.ContextTarReader.Close(); err != nil {
		return "", fmt.Errorf("error closing context tar: %w", err)
	}

	c.contextTmpDir = contextTmpDir

	return c.contextTmpDir, nil
}

func (c *BuildContext) Terminate() error {
	if c.contextTmpDir == "" {
		return nil
	}
	if err := os.RemoveAll(c.contextTmpDir); err != nil {
		return fmt.Errorf("unable to remove dir %q: %w", c.contextTmpDir, err)
	}
	return nil
}
