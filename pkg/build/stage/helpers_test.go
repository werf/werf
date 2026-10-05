package stage

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/storage"
)

var _ Conveyor = (*importChecksumConveyorStub)(nil)

type importChecksumConveyorStub struct {
	Conveyor
	readOnly bool
	metadata *storage.ImportMetadata
	err      error
	fetches  int
}

func (c *importChecksumConveyorStub) IsImagesReadOnly(context.Context) bool { return c.readOnly }

func (c *importChecksumConveyorStub) FetchImportMetadata(context.Context, string, string) (*storage.ImportMetadata, error) {
	return c.metadata, c.err
}

func (c *importChecksumConveyorStub) GetImageContentDigest(string, string) string {
	return "source-content"
}

func (c *importChecksumConveyorStub) FetchLastNonEmptyImageStage(context.Context, string, string) error {
	c.fetches++
	return fmt.Errorf("checksum source fetch requested")
}

func listTree(root string) []string {
	ginkgo.GinkgoHelper()
	var res []string
	gomega.Expect(filepath.WalkDir(root, func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		res = append(res, relative)
		return nil
	})).To(gomega.Succeed())
	return res
}

type gitRepoNameStub struct{ git_repo.GitRepo }

var _ git_repo.GitRepo = (*gitRepoNameStub)(nil)

func (*gitRepoNameStub) GetName() string { return "fixture" }
