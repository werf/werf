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

func TestAI_DeleteRejectedStageImage_HappyPath(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	const ts int64 = 1700000000
	stageRef := "registry.example/project:" + digest + "-1700000000"

	r := newFakeRegistry()
	r.tryGetInfo[stageRef] = &image.Info{Name: stageRef}
	s := &RepoStagesStorage{RepoAddress: "registry.example/project", DockerRegistry: r}

	err := s.DeleteRejectedStageImage(context.Background(), *image.NewStageID(digest, ts), DeleteImageOptions{})
	require.NoError(t, err)
	assert.Equal(t, 1, r.deleteCall[stageRef])
}

func TestAI_DeleteRejectedStageImage_AlreadyGone(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"

	r := newFakeRegistry()
	s := &RepoStagesStorage{RepoAddress: "registry.example/project", DockerRegistry: r}

	err := s.DeleteRejectedStageImage(context.Background(), *image.NewStageID(digest, 1700000000), DeleteImageOptions{})
	require.NoError(t, err)
	assert.Empty(t, r.deleteCall, "no delete attempt when stage image absent")
}

func TestAI_DeleteRejectedStageImage_BrokenFallback(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	const ts int64 = 1700000000
	stageRef := "registry.example/project:" + digest + "-1700000000"

	r := newFakeRegistry()
	r.tryGetInfo[stageRef] = &image.Info{Name: stageRef}
	r.deleteErrs[stageRef] = []error{errors.New("BLOB_UNKNOWN: corrupted blob"), nil}

	s := &RepoStagesStorage{RepoAddress: "registry.example/project", DockerRegistry: r}

	err := s.DeleteRejectedStageImage(context.Background(), *image.NewStageID(digest, ts), DeleteImageOptions{})
	require.NoError(t, err)
	assert.Equal(t, 2, r.deleteCall[stageRef], "stage delete retried after dummy push")
	assert.Equal(t, 1, r.pushCall[stageRef], "dummy push exactly once")
}

func TestAI_DeleteRejectedStageImage_NonBrokenErrorPropagates(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	const ts int64 = 1700000000
	stageRef := "registry.example/project:" + digest + "-1700000000"

	r := newFakeRegistry()
	r.tryGetInfo[stageRef] = &image.Info{Name: stageRef}
	r.deleteErrs[stageRef] = []error{errors.New("UNAUTHORIZED")}

	s := &RepoStagesStorage{RepoAddress: "registry.example/project", DockerRegistry: r}

	err := s.DeleteRejectedStageImage(context.Background(), *image.NewStageID(digest, ts), DeleteImageOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "UNAUTHORIZED")
	assert.Equal(t, 0, r.pushCall[stageRef], "must not push on non-broken errors")
}

func TestAI_DeleteRejectedStageImage_PushFallbackFails(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	const ts int64 = 1700000000
	stageRef := "registry.example/project:" + digest + "-1700000000"

	r := newFakeRegistry()
	r.tryGetInfo[stageRef] = &image.Info{Name: stageRef}
	r.deleteErrs[stageRef] = []error{errors.New("MANIFEST_INVALID")}
	r.pushErrs[stageRef] = errors.New("registry write denied")

	s := &RepoStagesStorage{RepoAddress: "registry.example/project", DockerRegistry: r}

	err := s.DeleteRejectedStageImage(context.Background(), *image.NewStageID(digest, ts), DeleteImageOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "overwrite broken image")
	assert.Contains(t, err.Error(), "MANIFEST_INVALID")
}

func TestAI_DeleteRejectedStageImage_FallbackVanishedAfterPushTreatedAsDeleted(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	const ts int64 = 1700000000
	stageRef := "registry.example/project:" + digest + "-1700000000"

	r := newFakeRegistry()
	origInfo := &image.Info{Name: stageRef}
	r.deleteErrs[stageRef] = []error{errors.New("DIGEST_INVALID")}
	calls := 0
	wrap := &vanishingRegistry{fakeRegistry: r, origInfo: origInfo, ref: stageRef, calls: &calls}
	s := &RepoStagesStorage{RepoAddress: "registry.example/project", DockerRegistry: wrap}

	err := s.DeleteRejectedStageImage(context.Background(), *image.NewStageID(digest, ts), DeleteImageOptions{})
	require.NoError(t, err)
	assert.Equal(t, 1, r.deleteCall[stageRef], "no retry needed when stage vanished after dummy push")
	assert.Equal(t, 1, r.pushCall[stageRef])
}

func TestAI_DeleteRejectedStageImage_FallbackPushSucceedsRetryFails(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	const ts int64 = 1700000000
	stageRef := "registry.example/project:" + digest + "-1700000000"

	r := newFakeRegistry()
	r.tryGetInfo[stageRef] = &image.Info{Name: stageRef}
	r.deleteErrs[stageRef] = []error{errors.New("MANIFEST_INVALID"), errors.New("BLOB_UNKNOWN: still corrupt")}

	s := &RepoStagesStorage{RepoAddress: "registry.example/project", DockerRegistry: r}

	err := s.DeleteRejectedStageImage(context.Background(), *image.NewStageID(digest, ts), DeleteImageOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "after dummy overwrite")
	assert.Contains(t, err.Error(), "BLOB_UNKNOWN")
	assert.Contains(t, err.Error(), "MANIFEST_INVALID", "original delete error must be preserved")
}

func TestAI_DeleteRejectedStageImage_FallbackPushImmutableTag(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	const ts int64 = 1700000000
	stageRef := "registry.example/project:" + digest + "-1700000000"

	r := newFakeRegistry()
	r.tryGetInfo[stageRef] = &image.Info{Name: stageRef}
	r.deleteErrs[stageRef] = []error{errors.New("MANIFEST_INVALID")}
	r.pushErrs[stageRef] = errors.New("ImageTagAlreadyExistsException: tag is immutable")

	s := &RepoStagesStorage{RepoAddress: "registry.example/project", DockerRegistry: r}

	err := s.DeleteRejectedStageImage(context.Background(), *image.NewStageID(digest, ts), DeleteImageOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "overwrite broken image")
	assert.Contains(t, err.Error(), "immutable")
	assert.Contains(t, err.Error(), "MANIFEST_INVALID")
}

func TestAI_DeleteRejectedStageRecord_HappyPath(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	const ts int64 = 1700000000
	rejectedRef := "registry.example/project:" + digest + "-1700000000-rejected"

	r := newFakeRegistry()
	r.tryGetInfo[rejectedRef] = &image.Info{Name: rejectedRef}
	s := &RepoStagesStorage{RepoAddress: "registry.example/project", DockerRegistry: r}

	err := s.DeleteRejectedStageRecord(context.Background(), *image.NewStageID(digest, ts), DeleteImageOptions{})
	require.NoError(t, err)
	assert.Equal(t, 1, r.deleteCall[rejectedRef])
}

func TestAI_DeleteRejectedStageRecord_AlreadyGone(t *testing.T) {
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"

	r := newFakeRegistry()
	s := &RepoStagesStorage{RepoAddress: "registry.example/project", DockerRegistry: r}

	err := s.DeleteRejectedStageRecord(context.Background(), *image.NewStageID(digest, 1700000000), DeleteImageOptions{})
	require.NoError(t, err)
	assert.Empty(t, r.deleteCall, "no delete attempt when marker absent")
}

func TestAI_DeleteRejectedStageRecord_BrokenErrorPropagatesNoFallback(t *testing.T) {
	// Marker is a metadata record without business payload — a broken marker must
	// surface as an error, not be silently replaced with an empty dummy.
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	const ts int64 = 1700000000
	rejectedRef := "registry.example/project:" + digest + "-1700000000-rejected"

	r := newFakeRegistry()
	r.tryGetInfo[rejectedRef] = &image.Info{Name: rejectedRef}
	r.deleteErrs[rejectedRef] = []error{errors.New("BLOB_UNKNOWN: corrupted marker")}

	s := &RepoStagesStorage{RepoAddress: "registry.example/project", DockerRegistry: r}

	err := s.DeleteRejectedStageRecord(context.Background(), *image.NewStageID(digest, ts), DeleteImageOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "BLOB_UNKNOWN")
	assert.Equal(t, 0, r.pushCall[rejectedRef], "marker delete must NOT use broken-image fallback")
}

func TestAI_DeleteStage_DoesNotTouchRejectedMarker(t *testing.T) {
	// Regression guard: DeleteStage has a single responsibility (remove the stage
	// image only). The rejected marker, if any, is cleaned up by the
	// deleteRejectedStagesWithLinkedTags phase, not by DeleteStage.
	digest := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	const ts int64 = 1700000000
	stageRef := "registry.example/project:" + digest + "-1700000000"
	rejectedRef := stageRef + "-rejected"

	r := newFakeRegistry()
	r.tryGetInfo[stageRef] = &image.Info{Name: stageRef}
	r.tryGetInfo[rejectedRef] = &image.Info{Name: rejectedRef}
	s := &RepoStagesStorage{RepoAddress: "registry.example/project", DockerRegistry: r}

	stageDesc := &image.StageDesc{
		StageID: image.NewStageID(digest, ts),
		Info:    &image.Info{Name: stageRef},
	}
	err := s.DeleteStage(context.Background(), stageDesc, DeleteImageOptions{})
	require.NoError(t, err)
	assert.Equal(t, 1, r.deleteCall[stageRef], "stage image deleted")
	assert.Equal(t, 0, r.deleteCall[rejectedRef], "marker MUST NOT be touched by DeleteStage")
}

func TestAI_DeleteStageCustomTag_HappyPath(t *testing.T) {
	tag := "latest"
	customRef := "registry.example/project:latest"

	r := newFakeRegistry()
	r.tryGetInfo[customRef] = &image.Info{Name: customRef}
	s := &RepoStagesStorage{RepoAddress: "registry.example/project", DockerRegistry: r}

	err := s.DeleteStageCustomTag(context.Background(), tag)
	require.NoError(t, err)
	assert.Equal(t, 1, r.deleteCall[customRef])
	assert.Equal(t, 0, r.pushCall[customRef], "must not push when delete works")
}

func TestAI_DeleteStageCustomTag_Missing(t *testing.T) {
	r := newFakeRegistry()
	s := &RepoStagesStorage{RepoAddress: "registry.example/project", DockerRegistry: r}

	err := s.DeleteStageCustomTag(context.Background(), "missing")
	require.NoError(t, err)
	assert.Empty(t, r.deleteCall, "no delete when tag absent")
}

func TestAI_DeleteStageCustomTag_BrokenErrorPropagatesNoFallback(t *testing.T) {
	// Custom tag carries data the user pushed under that tag — a broken custom
	// tag must surface as an error, not be silently replaced with an empty dummy.
	tag := "v1.0.0"
	customRef := "registry.example/project:v1.0.0"

	r := newFakeRegistry()
	r.tryGetInfo[customRef] = &image.Info{Name: customRef}
	r.deleteErrs[customRef] = []error{errors.New("BLOB_UNKNOWN: corrupted custom tag")}

	s := &RepoStagesStorage{RepoAddress: "registry.example/project", DockerRegistry: r}

	err := s.DeleteStageCustomTag(context.Background(), tag)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "BLOB_UNKNOWN")
	assert.Equal(t, 0, r.pushCall[customRef], "custom tag delete must NOT use broken-image fallback")
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
