package migrate

import (
	"context"
	"io"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/lockgate/pkg/file_locker"
	"github.com/werf/logboek"
	"github.com/werf/werf/v3/pkg/docker_registry"
	"github.com/werf/werf/v3/pkg/image"
	"github.com/werf/werf/v3/pkg/storage"
	"github.com/werf/werf/v3/pkg/storage/synchronization/lock_manager"
	"github.com/werf/werf/v3/pkg/werf"
)

type migrationRegistry struct {
	docker_registry.Interface
	mu      sync.Mutex
	marker  *image.Info
	reads   atomic.Int64
	entered chan struct{}
	release chan struct{}
}

var _ docker_registry.Interface = (*migrationRegistry)(nil)

func (r *migrationRegistry) TryGetRepoImage(ctx context.Context, _ string) (*image.Info, error) {
	if r.reads.Add(1) == 1 {
		close(r.entered)
		select {
		case <-r.release:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.marker, nil
}

func (r *migrationRegistry) Tags(context.Context, string, ...docker_registry.Option) ([]string, error) {
	return nil, nil
}

func (r *migrationRegistry) PushImage(_ context.Context, _ string, opts *docker_registry.PushImageOptions) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.marker = &image.Info{Labels: opts.Labels}
	return nil
}

var _ = DescribeTable("Concurrent metadata migrations", func(secondDestination string, conflict bool) {
	Expect(werf.Init(GinkgoT().TempDir(), GinkgoT().TempDir())).To(Succeed())
	locker, err := file_locker.NewFileLocker(GinkgoT().TempDir())
	Expect(err).NotTo(HaveOccurred())
	locks := lock_manager.NewGeneric(locker)
	reg := &migrationRegistry{entered: make(chan struct{}), release: make(chan struct{})}
	var release sync.Once
	defer release.Do(func() { close(reg.release) })
	src := &storage.RepoStagesStorage{RepoAddress: "registry.example/source", DockerRegistry: reg}
	first := &storage.RepoStagesStorage{RepoAddress: "registry.example/meta", DockerRegistry: reg}
	second := &storage.RepoStagesStorage{RepoAddress: secondDestination, DockerRegistry: reg}
	run := func(dst *storage.RepoStagesStorage, result chan error) {
		ctx := logboek.NewContext(context.Background(), logboek.NewLogger(io.Discard, io.Discard))
		result <- migrateWithLock(ctx, "project", src, dst, locks, storage.MigrateMetaRepoOptions{}, time.Second)
	}
	firstResult, secondResult := make(chan error, 1), make(chan error, 1)
	go run(first, firstResult)
	Eventually(reg.entered).Should(BeClosed())
	go run(second, secondResult)
	Consistently(reg.reads.Load, 100*time.Millisecond).Should(Equal(int64(1)))
	release.Do(func() { close(reg.release) })
	Eventually(firstResult).WithTimeout(5 * time.Second).Should(Receive(&err))
	Expect(err).NotTo(HaveOccurred())
	Eventually(secondResult).WithTimeout(5 * time.Second).Should(Receive(&err))
	if conflict {
		Expect(err).To(MatchError(ContainSubstring("refusing to migrate")))
	} else {
		Expect(err).NotTo(HaveOccurred())
	}
	Expect(reg.marker.Labels[storage.RepoMetaRepoMarker_LabelMetaRepo]).To(Equal("registry.example/meta"))
},
	Entry("waits and succeeds for the same destination", "registry.example/meta", false),
	Entry("rechecks the marker before using a conflicting destination", "registry.example/other", true),
)

var _ = It("releases the migration lock after a timeout so a retry succeeds", func() {
	Expect(werf.Init(GinkgoT().TempDir(), GinkgoT().TempDir())).To(Succeed())
	locker, err := file_locker.NewFileLocker(GinkgoT().TempDir())
	Expect(err).NotTo(HaveOccurred())
	locks := lock_manager.NewGeneric(locker)
	reg := &migrationRegistry{entered: make(chan struct{}), release: make(chan struct{})}
	src := &storage.RepoStagesStorage{RepoAddress: "registry.example/source", DockerRegistry: reg}
	dst := &storage.RepoStagesStorage{RepoAddress: "registry.example/meta", DockerRegistry: reg}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ctx = logboek.NewContext(ctx, logboek.NewLogger(io.Discard, io.Discard))
	Expect(migrateWithLock(ctx, "project", src, dst, locks, storage.MigrateMetaRepoOptions{}, 10*time.Millisecond)).To(MatchError(ContainSubstring("context deadline exceeded")))
	Expect(ctx.Err()).NotTo(HaveOccurred())
	Expect(reg.marker).To(BeNil())
	Expect(migrateWithLock(ctx, "project", src, dst, locks, storage.MigrateMetaRepoOptions{}, time.Second)).To(Succeed())
})
