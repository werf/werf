package storage

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/level"
	"github.com/werf/werf/v3/pkg/docker_registry"
)

type interruptedMigrationRegistry struct {
	docker_registry.Interface
	copies atomic.Int64
}

var _ docker_registry.Interface = (*interruptedMigrationRegistry)(nil)

func (r *interruptedMigrationRegistry) CopyImage(ctx context.Context, src, dst string, opts docker_registry.CopyImageOptions) error {
	if r.copies.Add(1) == 3 {
		return fmt.Errorf("interrupted copy")
	}
	return r.Interface.CopyImage(ctx, src, dst, opts)
}

var _ = Describe("Metadata migration progress", func() {
	var output bytes.Buffer
	var outputMu sync.Mutex
	var ctx context.Context
	var readOutput func() string

	BeforeEach(func() {
		output.Reset()
		writer := &migrationLogWriter{writer: &output, mu: &outputMu}
		logger := logboek.NewLogger(writer, writer)
		logger.Streams().DisablePrettyLog()
		logger.SetAcceptedLevel(level.Debug)
		ctx = logboek.NewContext(context.Background(), logger)
		readOutput = func() string {
			outputMu.Lock()
			defer outputMu.Unlock()
			return output.String()
		}
	})

	It("runs tasks concurrently within the limit and reports completed work", func() {
		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		var active, peak, completed atomic.Int64
		barrier := make(chan struct{})
		var release sync.Once
		err := runMetaMigrationPhase(ctx, "inspect", 12, 3, time.Hour, func(ctx context.Context, index int) error {
			return logboek.Context(ctx).Debug().LogProcess("record %d", index).DoError(func() error {
				n := active.Add(1)
				defer active.Add(-1)
				for old := peak.Load(); n > old && !peak.CompareAndSwap(old, n); old = peak.Load() {
				}
				if n == 3 {
					release.Do(func() { close(barrier) })
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-barrier:
					logboek.Context(ctx).Debug().LogF("end record %d\n", index)
					completed.Add(1)
					return nil
				}
			})
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(peak.Load()).To(Equal(int64(3)))
		Expect(completed.Load()).To(Equal(int64(12)))
		Expect(readOutput()).To(ContainSubstring("inspect: 12/12 completed"))
		Expect(readOutput()).To(ContainSubstring("(done)"))
		for i := range 12 {
			Expect(readOutput()).To(ContainSubstring(fmt.Sprintf("end record %d\n", i)))
		}
	})

	It("reports a heartbeat while a request is blocked and waits for cancellation", func() {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		result := make(chan error, 1)
		go func() {
			result <- runMetaMigrationPhase(ctx, "copy", 1, 1, 10*time.Millisecond, func(ctx context.Context, _ int) error {
				<-ctx.Done()
				return ctx.Err()
			})
		}()
		Eventually(func() int { return strings.Count(readOutput(), "(running)") }).WithTimeout(time.Second).Should(BeNumerically(">=", 2))
		cancel()
		var err error
		Eventually(result).WithTimeout(time.Second).Should(Receive(&err))
		Expect(err).To(MatchError(ContainSubstring("context canceled")))
		Expect(readOutput()).To(ContainSubstring("copy: 0/1 completed"))
		Expect(readOutput()).To(ContainSubstring("(interrupted)"))
		Expect(readOutput()).NotTo(ContainSubstring("(done)"))
	})

	It("preserves sources after a partial concurrent copy and resumes successfully", func() {
		reg := newMarkerRegistry()
		interrupted := &interruptedMigrationRegistry{Interface: reg}
		src := newRepoStorage(reg, "registry.example/source")
		dst := newRepoStorage(interrupted, "registry.example/meta")
		for i := range 12 {
			reg.put(fmt.Sprintf("registry.example/source:managed-image-app%d", i), map[string]string{})
		}
		opts := MigrateMetaRepoOptions{RemoveSource: true, ParallelTasksLimit: 3}
		Expect(MigrateMetaRepo(ctx, "project", src, dst, opts)).To(MatchError(ContainSubstring("interrupted copy")))
		for i := range 12 {
			Expect(reg.has(fmt.Sprintf("registry.example/source:managed-image-app%d", i))).To(BeTrue())
		}
		Expect(MigrateMetaRepo(ctx, "project", src, dst, opts)).To(Succeed())
		for i := range 12 {
			Expect(reg.has(fmt.Sprintf("registry.example/source:managed-image-app%d", i))).To(BeFalse())
			Expect(reg.has(fmt.Sprintf("registry.example/meta:managed-image-app%d", i))).To(BeTrue())
		}
	})
})
