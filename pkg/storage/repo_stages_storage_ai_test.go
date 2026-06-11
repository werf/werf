package storage

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/werf/werf/v2/pkg/docker_registry"
	"github.com/werf/werf/v2/pkg/image"
)

type fakeRegistry struct {
	docker_registry.Interface

	mu sync.Mutex

	tags    []string
	tagsErr error

	tryGetInfo map[string]*image.Info
	tryGetErr  map[string]error

	deleteErrs map[string][]error
	deleteCall map[string]int

	pushErrs map[string]error
	pushCall map[string]int
}

func newFakeRegistry() *fakeRegistry {
	return &fakeRegistry{
		tryGetInfo: map[string]*image.Info{},
		tryGetErr:  map[string]error{},
		deleteErrs: map[string][]error{},
		deleteCall: map[string]int{},
		pushErrs:   map[string]error{},
		pushCall:   map[string]int{},
	}
}

func (r *fakeRegistry) Tags(_ context.Context, _ string, _ ...docker_registry.Option) ([]string, error) {
	return r.tags, r.tagsErr
}

func (r *fakeRegistry) TryGetRepoImage(_ context.Context, reference string) (*image.Info, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err, ok := r.tryGetErr[reference]; ok {
		return nil, err
	}
	return r.tryGetInfo[reference], nil
}

func (r *fakeRegistry) DeleteRepoImage(_ context.Context, info *image.Info) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	ref := info.Name
	r.deleteCall[ref]++
	errs := r.deleteErrs[ref]
	if len(errs) == 0 {
		return nil
	}
	err := errs[0]
	r.deleteErrs[ref] = errs[1:]
	return err
}

func (r *fakeRegistry) PushImage(_ context.Context, reference string, _ *docker_registry.PushImageOptions) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.pushCall[reference]++
	return r.pushErrs[reference]
}

func TestAI_GetRejectedStageIDs(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"

	tests := []struct {
		name     string
		tags     []string
		expectN  int
		expectTs []int64
	}{
		{
			name:    "empty tag list returns nothing",
			tags:    nil,
			expectN: 0,
		},
		{
			name:    "ignores non-rejected tags",
			tags:    []string{digest + "-1700000000", digest},
			expectN: 0,
		},
		{
			name:     "picks rejected tags",
			tags:     []string{digest + "-1700000000-rejected", digest + "-1700000001-rejected", digest + "-1700000002"},
			expectN:  2,
			expectTs: []int64{1700000000, 1700000001},
		},
		{
			name:    "skips malformed rejected tag",
			tags:    []string{"garbage-rejected", digest + "-1700000000-rejected"},
			expectN: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := newFakeRegistry()
			r.tags = tc.tags
			s := &RepoStagesStorage{RepoAddress: "registry.example/project", DockerRegistry: r}

			got, err := s.GetRejectedStageIDs(context.Background())
			require.NoError(t, err)
			assert.Len(t, got, tc.expectN)
			for i, ts := range tc.expectTs {
				assert.Equal(t, digest, got[i].Digest)
				assert.Equal(t, ts, got[i].CreationTs)
			}
		})
	}
}

func TestAI_DeleteRejectedStage_HappyPath(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	const ts int64 = 1700000000
	rejectedRef := "registry.example/project:" + digest + "-1700000000-rejected"

	r := newFakeRegistry()
	r.tryGetInfo[rejectedRef] = &image.Info{Name: rejectedRef}
	s := &RepoStagesStorage{RepoAddress: "registry.example/project", DockerRegistry: r}

	err := s.DeleteRejectedStage(context.Background(), *image.NewStageID(digest, ts), DeleteImageOptions{})
	require.NoError(t, err)
	assert.Equal(t, 1, r.deleteCall[rejectedRef])
	assert.Equal(t, 0, r.pushCall[rejectedRef], "must not push when delete works")
}

func TestAI_DeleteRejectedStage_MissingMarker(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	r := newFakeRegistry()
	s := &RepoStagesStorage{RepoAddress: "registry.example/project", DockerRegistry: r}

	err := s.DeleteRejectedStage(context.Background(), *image.NewStageID(digest, 1700000000), DeleteImageOptions{})
	require.NoError(t, err)
	assert.Empty(t, r.deleteCall, "no delete attempt when marker absent")
}

func TestAI_DeleteRejectedStage_BrokenFallback(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	const ts int64 = 1700000000
	rejectedRef := "registry.example/project:" + digest + "-1700000000-rejected"

	r := newFakeRegistry()
	// first TryGet returns broken info, second (after push) returns fresh info
	r.tryGetInfo[rejectedRef] = &image.Info{Name: rejectedRef}
	// first delete fails with broken image error, second succeeds
	r.deleteErrs[rejectedRef] = []error{errors.New("BLOB_UNKNOWN: corrupted blob"), nil}

	s := &RepoStagesStorage{RepoAddress: "registry.example/project", DockerRegistry: r}

	err := s.DeleteRejectedStage(context.Background(), *image.NewStageID(digest, ts), DeleteImageOptions{})
	require.NoError(t, err)
	assert.Equal(t, 2, r.deleteCall[rejectedRef], "expected delete retried after dummy push")
	assert.Equal(t, 1, r.pushCall[rejectedRef], "expected dummy push exactly once")
}

func TestAI_DeleteRejectedStage_NonBrokenErrorPropagates(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	const ts int64 = 1700000000
	rejectedRef := "registry.example/project:" + digest + "-1700000000-rejected"

	r := newFakeRegistry()
	r.tryGetInfo[rejectedRef] = &image.Info{Name: rejectedRef}
	r.deleteErrs[rejectedRef] = []error{errors.New("UNAUTHORIZED")}

	s := &RepoStagesStorage{RepoAddress: "registry.example/project", DockerRegistry: r}

	err := s.DeleteRejectedStage(context.Background(), *image.NewStageID(digest, ts), DeleteImageOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "UNAUTHORIZED")
	assert.Equal(t, 0, r.pushCall[rejectedRef], "must not push on non-broken errors")
}

func TestAI_DeleteRejectedStage_PushFallbackFails(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	const ts int64 = 1700000000
	rejectedRef := "registry.example/project:" + digest + "-1700000000-rejected"

	r := newFakeRegistry()
	r.tryGetInfo[rejectedRef] = &image.Info{Name: rejectedRef}
	r.deleteErrs[rejectedRef] = []error{errors.New("MANIFEST_INVALID")}
	r.pushErrs[rejectedRef] = errors.New("registry write denied")

	s := &RepoStagesStorage{RepoAddress: "registry.example/project", DockerRegistry: r}

	err := s.DeleteRejectedStage(context.Background(), *image.NewStageID(digest, ts), DeleteImageOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "overwrite broken image")
	assert.Contains(t, err.Error(), "MANIFEST_INVALID")
}

func TestAI_DeleteRejectedStage_FallbackVanishedTreatedAsDeleted(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	const ts int64 = 1700000000
	rejectedRef := "registry.example/project:" + digest + "-1700000000-rejected"

	r := newFakeRegistry()
	// first call returns broken image; after push registry still returns nil meaning the tag
	// is gone (some registries garbage-collect during overwrite).
	calls := 0
	r.tryGetErr = nil
	origInfo := &image.Info{Name: rejectedRef}
	r.deleteErrs[rejectedRef] = []error{errors.New("DIGEST_INVALID")}
	// custom TryGet: first call returns info, second returns nil
	wrap := &vanishingRegistry{fakeRegistry: r, origInfo: origInfo, ref: rejectedRef, calls: &calls}
	s := &RepoStagesStorage{RepoAddress: "registry.example/project", DockerRegistry: wrap}

	err := s.DeleteRejectedStage(context.Background(), *image.NewStageID(digest, ts), DeleteImageOptions{})
	require.NoError(t, err)
	assert.Equal(t, 1, r.deleteCall[rejectedRef], "no retry needed when fallback finds tag gone")
	assert.Equal(t, 1, r.pushCall[rejectedRef])
}

type vanishingRegistry struct {
	*fakeRegistry
	origInfo *image.Info
	ref      string
	calls    *int
}

func (r *vanishingRegistry) TryGetRepoImage(_ context.Context, reference string) (*image.Info, error) {
	if reference != r.ref {
		return r.fakeRegistry.TryGetRepoImage(context.Background(), reference)
	}
	*r.calls++
	if *r.calls == 1 {
		return r.origInfo, nil
	}
	return nil, nil
}
