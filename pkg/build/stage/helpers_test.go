package stage

import (
	"context"
	"fmt"

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
