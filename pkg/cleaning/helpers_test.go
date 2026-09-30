package cleaning

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"

	"github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/pkg/storage"
	"github.com/werf/werf/v3/pkg/storage/manager"
)

var _ manager.StorageManagerInterface = (*fakeStorageManager)(nil)

func newTestStageDesc(repository string, stageID *image.StageID) *image.StageDesc {
	return &image.StageDesc{
		StageID: stageID,
		Info: &image.Info{
			Repository: repository,
			Tag:        stageID.String(),
			Name:       repository + ":" + stageID.String(),
		},
	}
}

func newFakeContextClient(contextName, contextNamespace string, podImagesByNamespace map[string][]string) *ContextClient {
	var objects []runtime.Object
	for namespace, images := range podImagesByNamespace {
		for ind, img := range images {
			objects = append(objects, &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-pod-%d", namespace, ind), Namespace: namespace},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "app", Image: img}},
				},
			})
		}
	}

	return &ContextClient{
		ContextName:      contextName,
		ContextNamespace: contextNamespace,
		Client:           fake.NewClientset(objects...),
	}
}

func failPodsListInNamespace(contextClient *ContextClient, namespace string) {
	contextClient.Client.(*fake.Clientset).PrependReactor("list", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetNamespace() != namespace {
			return false, nil, nil
		}

		return true, nil, fmt.Errorf("pods is forbidden in namespace %q", namespace)
	})
}

func (f *fakeStorageManager) ForEachDeleteFinalStage(ctx context.Context, _ manager.ForEachDeleteStageOptions, stages image.StageDescSet, onDelete func(context.Context, *image.StageDesc, error) error) error {
	for stage := range stages.Iter() {
		f.deletedFinalStages = append(f.deletedFinalStages, stage)
		if err := onDelete(ctx, stage, nil); err != nil {
			return err
		}
	}
	return nil
}

func (f *fakeStorageManager) ForEachDeleteStage(ctx context.Context, _ manager.ForEachDeleteStageOptions, stages image.StageDescSet, onDelete func(context.Context, *image.StageDesc, error) error) error {
	for stage := range stages.Iter() {
		f.deletedStages = append(f.deletedStages, stage)
		if err := onDelete(ctx, stage, nil); err != nil {
			return err
		}
	}
	return nil
}

func (f *fakeStorageManager) ForEachGetStageCustomTagMetadata(_ context.Context, _ []string, _ func(context.Context, string, *storage.CustomTagMetadata, error) error) error {
	return nil
}

func (f *fakeStorageManager) GetFinalStagesStorage() storage.StagesStorage {
	return nil
}

func (f *fakePrimaryStagesStorage) GetAllAndGroupImageMetadataByImageName(_ context.Context, _ string, _ []string, _ ...storage.Option) (map[string]map[string][]string, map[string]map[string][]string, error) {
	return nil, nil, nil
}

func (f *fakePrimaryStagesStorage) GetStageCustomTagMetadataIDs(_ context.Context, _ ...storage.Option) ([]string, error) {
	return nil, nil
}

func (f *fakePrimaryStagesStorage) PostLastCleanupRecord(_ context.Context, projectName string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lastCleanupRecords = append(f.lastCleanupRecords, projectName)
	return nil
}

var (
	_ manager.StorageManagerInterface = (*snapshotStorageManager)(nil)
	_ storage.PrimaryStagesStorage    = (*snapshotMetaStorage)(nil)
)

type snapshotStorageManager struct {
	*fakeStorageManager
	metaSnapshot    *snapshotMetaStorage
	final           storage.StagesStorage
	publish         bool
	published       bool
	primarySnapshot image.StageDescSet
	finalErr        error
	events          []string
	metadataNames   []string
}

type snapshotMetaStorage struct {
	*fakePrimaryStagesStorage
	manager *snapshotStorageManager
	shared  bool
}

func newSnapshotStorageManager(splitMeta bool) *snapshotStorageManager {
	m := &snapshotStorageManager{fakeStorageManager: newFakeStorageManager(), publish: true}
	m.metaSnapshot = &snapshotMetaStorage{fakePrimaryStagesStorage: newFakePrimaryStagesStorage(), manager: m, shared: !splitMeta}
	m.final = m.metaSnapshot
	return m
}

func (m *snapshotStorageManager) GetMetaStorage() storage.PrimaryStagesStorage { return m.metaSnapshot }
func (m *snapshotStorageManager) GetFinalStagesStorage() storage.StagesStorage { return m.final }

func (m *snapshotStorageManager) snapshotPrimary() image.StageDescSet {
	if m.primarySnapshot == nil {
		m.primarySnapshot = image.NewStageDescSet()
		if m.published {
			m.primarySnapshot.Add(newTestStageDesc("primary", image.NewStageID("new-stage", 1)))
		}
	}
	return m.primarySnapshot
}

func (m *snapshotStorageManager) GetStageDescSetWithCache(context.Context) (image.StageDescSet, error) {
	m.events = append(m.events, "primary")
	return m.snapshotPrimary(), nil
}

func (m *snapshotStorageManager) GetFinalStageDescSet(context.Context) (image.StageDescSet, error) {
	m.events = append(m.events, "final")
	if m.finalErr != nil {
		return nil, m.finalErr
	}
	m.published = m.publish
	return image.NewStageDescSet(newTestStageDesc("final", image.NewStageID("new-stage", 1))), nil
}

func (s *snapshotMetaStorage) GetManagedImages(context.Context, string, ...storage.Option) ([]string, error) {
	s.manager.events = append(s.manager.events, "managed")
	if s.shared {
		s.manager.snapshotPrimary()
	}
	return []string{"shared", "published", "shared"}, nil
}

func (s *snapshotMetaStorage) GetAllAndGroupImageMetadataByImageName(_ context.Context, _ string, names []string, _ ...storage.Option) (map[string]map[string][]string, map[string]map[string][]string, error) {
	s.manager.metadataNames = append([]string(nil), names...)
	return nil, nil, nil
}

func (s *snapshotMetaStorage) Address() string { return "final" }
func (s *snapshotMetaStorage) String() string  { return "final" }

func (f *fakePrimaryStagesStorage) GetManagedImages(context.Context, string, ...storage.Option) ([]string, error) {
	return nil, nil
}
