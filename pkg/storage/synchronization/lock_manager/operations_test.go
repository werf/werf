package lock_manager

import (
	"errors"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/lockgate"
	"github.com/werf/werf/v2/pkg/opstats"
)

var _ = ginkgo.DescribeTable("storage lock acquisition statistics",
	func(ctx ginkgo.SpecContext, kubernetes, fail bool) {
		collector := opstats.NewCollector()
		observedCtx := opstats.NewContext(ctx, collector)
		locker := &observedTestLocker{onAcquire: func() {
			gomega.Expect(collector.Summary()).To(gomega.BeEmpty())
		}}
		if fail {
			locker.err = errors.New("acquisition failed")
		}

		var manager Interface = NewGeneric(locker)
		if kubernetes {
			manager = &Kubernetes{LockerPerProject: map[string]lockgate.Locker{"project": locker}}
		}
		handle, err := manager.LockStage(observedCtx, "project", "digest")
		if fail {
			gomega.Expect(err).To(gomega.MatchError(locker.err))
		} else {
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}

		summary := collector.Summary()
		gomega.Expect(summary).To(gomega.HaveLen(1))
		gomega.Expect(summary[0].Operation).To(gomega.Equal(opstats.Operation("sync: lock acquire")))
		gomega.Expect(summary[0].Count).To(gomega.Equal(1))
		if !fail {
			gomega.Expect(manager.Unlock(observedCtx, handle)).To(gomega.Succeed())
			gomega.Expect(collector.Summary()).To(gomega.Equal(summary))
		}
	},
	ginkgo.Entry("HTTP/local lock success", false, false),
	ginkgo.Entry("HTTP/local lock failure", false, true),
	ginkgo.Entry("Kubernetes lock success", true, false),
	ginkgo.Entry("Kubernetes lock failure", true, true),
)
