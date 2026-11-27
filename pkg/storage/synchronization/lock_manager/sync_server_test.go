package lock_manager

import (
	"context"
	"errors"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/storage"
)

type syncServerRecordStorage struct {
	storage.StagesStorage
	postErr error
}

var _ storage.StagesStorage = (*syncServerRecordStorage)(nil)

func (s *syncServerRecordStorage) GetSyncServerRecords(_ context.Context, _ string, _ ...storage.Option) ([]*storage.SyncServerRecord, error) {
	return nil, nil
}

func (s *syncServerRecordStorage) PostSyncServerRecord(_ context.Context, _ string, _ *storage.SyncServerRecord) error {
	return s.postErr
}

var _ = ginkgo.Describe("Synchronization server record", func() {
	ginkgo.It("reports why the record could not be created", func(ctx ginkgo.SpecContext) {
		postErr := errors.New("registry rejected the record")
		stagesStorage := &syncServerRecordStorage{postErr: postErr}

		_, err := GetOrCreateSyncServer(ctx, "project", "https://sync.example.com", stagesStorage)

		gomega.Expect(err).To(gomega.MatchError(postErr))
	})
})
