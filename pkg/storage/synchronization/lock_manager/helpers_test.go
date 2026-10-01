package lock_manager

import (
	"context"
	"sync"

	"github.com/werf/werf/v3/pkg/storage"
)

type clientRecordStorage struct {
	storage.StagesStorage
	mutex    sync.Mutex
	records  []*storage.ClientIDRecord
	readErr  error
	writeErr error
}

var _ storage.StagesStorage = (*clientRecordStorage)(nil)

func (s *clientRecordStorage) GetClientIDRecords(_ context.Context, _ string, _ ...storage.Option) ([]*storage.ClientIDRecord, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return append([]*storage.ClientIDRecord(nil), s.records...), s.readErr
}

func (s *clientRecordStorage) PostClientIDRecord(_ context.Context, _ string, record *storage.ClientIDRecord) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.writeErr != nil {
		return s.writeErr
	}
	s.records = append(s.records, record)
	return nil
}
