package common

import (
	"errors"
	"net/http"
	"os"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/werf/werf/v3/pkg/storage"
	"github.com/werf/werf/v3/pkg/storage/synchronization/lock_manager"
	"github.com/werf/werf/v3/pkg/storage/synchronization/server"
)

var _ = ginkgo.Describe("repository synchronization policy", func() {
	ginkgo.DescribeTable("only permits fallback when the public endpoint was implicit", func(ctx ginkgo.SpecContext, source string, existingID bool) {
		previousEnv, present := os.LookupEnv("WERF_SYNCHRONIZATION")
		ginkgo.DeferCleanup(func() {
			if present {
				gomega.Expect(os.Setenv("WERF_SYNCHRONIZATION", previousEnv)).To(gomega.Succeed())
			} else {
				gomega.Expect(os.Unsetenv("WERF_SYNCHRONIZATION")).To(gomega.Succeed())
			}
		})
		gomega.Expect(os.Unsetenv("WERF_SYNCHRONIZATION")).To(gomega.Succeed())
		if source == "env" {
			gomega.Expect(os.Setenv("WERF_SYNCHRONIZATION", server.DefaultAddress)).To(gomega.Succeed())
		}
		command := &cobra.Command{}
		data := &CmdData{}
		SetupSynchronization(data, command)
		if source == "flag" {
			gomega.Expect(command.Flags().Set("synchronization", server.DefaultAddress)).To(gomega.Succeed())
		}
		store := synchronizationTestStorage(ctx)
		if existingID {
			gomega.Expect(store.PostClientIDRecord(ctx, "project", &storage.ClientIDRecord{ClientID: "existing", TimestampMillisec: 1})).To(gomega.Succeed())
		}
		original := http.DefaultTransport
		http.DefaultTransport = &synchronizationDNSFailureTransport{next: original}
		ginkgo.DeferCleanup(func() { http.DefaultTransport = original })
		synchronization, err := GetSynchronization(ctx, data, "project", store)
		if err == nil {
			manager, managerErr := synchronization.GetStorageLockManager(ctx)
			gomega.Expect(managerErr).NotTo(gomega.HaveOccurred())
			var handle lock_manager.LockHandle
			handle, err = manager.LockStage(ctx, "project", "stage")
			if err == nil {
				gomega.Expect(manager.Unlock(ctx, handle)).To(gomega.Succeed())
			}
		}
		if source == "implicit" {
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		} else {
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("no such host"))
		}
	},
		ginkgo.Entry("implicit bootstrap", "implicit", false),
		ginkgo.Entry("explicit flag bootstrap", "flag", false),
		ginkgo.Entry("explicit environment bootstrap", "env", false),
		ginkgo.Entry("implicit existing ID", "implicit", true),
		ginkgo.Entry("explicit flag existing ID", "flag", true),
		ginkgo.Entry("explicit environment existing ID", "env", true))

	ginkgo.It("does not pin read-only initialization and pins the first successful publication only once", func(ctx ginkgo.SpecContext) {
		store := synchronizationTestStorage(ctx)
		address := ":local"
		_, err := GetSynchronization(ctx, &CmdData{Synchronization: &address}, "project", store)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		_, found, err := store.GetSynchronizationMarker(ctx, "project")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(found).To(gomega.BeFalse())
		locker := &synchronizationTestLocker{acquireErr: errors.New("lock unavailable")}
		manager := &markerLockManager{Interface: locker, store: store, projectName: "project", address: address}
		_, err = manager.LockStage(ctx, "project", "stage")
		gomega.Expect(err).To(gomega.MatchError("lock unavailable"))
		_, found, err = store.GetSynchronizationMarker(ctx, "project")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(found).To(gomega.BeFalse())
		locker.acquireErr = nil
		handle, err := manager.LockStage(ctx, "project", "stage")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(manager.Unlock(ctx, handle)).To(gomega.Succeed())
		fingerprint, found, err := store.GetSynchronizationMarker(ctx, "project")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(found).To(gomega.BeTrue())
		gomega.Expect(fingerprint).To(gomega.Equal(synchronizationFingerprint(address)))
		store.DockerRegistry = &synchronizationErrorRegistry{Interface: store.DockerRegistry, readErr: errors.New("must not reread on every stage")}
		_, err = manager.LockStage(ctx, "project", "another-stage")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})
	ginkgo.DescribeTable("rejects missing or mismatching configuration before contacting synchronization", func(ctx ginkgo.SpecContext, address string) {
		store := synchronizationTestStorage(ctx)
		gomega.Expect(store.PutSynchronizationMarker(ctx, "project", synchronizationFingerprint(":local"))).To(gomega.Succeed())
		_, err := GetSynchronization(ctx, &CmdData{Synchronization: &address}, "project", store)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("synchronization"))
		gomega.Expect(err.Error()).NotTo(gomega.ContainSubstring("secret-password"))
	}, ginkgo.Entry("automatic", ""), ginkgo.Entry("another backend", "https://user:secret-password@invalid.test"))
	ginkgo.It("fails on unreadable marker before attempting public fallback", func(ctx ginkgo.SpecContext) {
		store := synchronizationTestStorage(ctx)
		failure := errors.New("registry access denied")
		store.DockerRegistry = &synchronizationErrorRegistry{Interface: store.DockerRegistry, readErr: failure}
		address := ""
		_, err := GetSynchronization(ctx, &CmdData{Synchronization: &address}, "project", store)
		gomega.Expect(errors.Is(err, failure)).To(gomega.BeTrue())
	})
	ginkgo.It("never pins implicit synchronization", func(ctx ginkgo.SpecContext) {
		store := synchronizationTestStorage(ctx)
		manager := &markerLockManager{Interface: &synchronizationTestLocker{}, store: store, projectName: "project"}
		_, err := manager.LockStage(ctx, "project", "stage")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		_, found, err := store.GetSynchronizationMarker(ctx, "project")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(found).To(gomega.BeFalse())
	})
	ginkgo.DescribeTable("releases the actual lock when policy changes while acquiring", func(ctx ginkgo.SpecContext, address string) {
		store := synchronizationTestStorage(ctx)
		locker := &synchronizationTestLocker{onAcquire: func() {
			gomega.Expect(store.PutSynchronizationMarker(ctx, "project", synchronizationFingerprint("another-backend"))).To(gomega.Succeed())
		}}
		manager := &markerLockManager{Interface: locker, store: store, projectName: "project", address: address}
		_, err := manager.LockStage(ctx, "project", "stage")
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(locker.released).To(gomega.Equal([]lock_manager.LockHandle{{ProjectName: "project/stage"}}))
	}, ginkgo.Entry("explicit", ":local"), ginkgo.Entry("implicit", ""))
	ginkgo.It("preserves both marker write and release failures", func(ctx ginkgo.SpecContext) {
		store := synchronizationTestStorage(ctx)
		writeFailure, releaseFailure := errors.New("write denied"), errors.New("release failed")
		store.DockerRegistry = &synchronizationErrorRegistry{Interface: store.DockerRegistry, writeErr: writeFailure}
		locker := &synchronizationTestLocker{releaseErr: releaseFailure}
		manager := &markerLockManager{Interface: locker, store: store, projectName: "project", address: ":local"}
		_, err := manager.LockStage(ctx, "project", "stage")
		gomega.Expect(errors.Is(err, writeFailure)).To(gomega.BeTrue())
		gomega.Expect(errors.Is(err, releaseFailure)).To(gomega.BeTrue())
		gomega.Expect(locker.released).To(gomega.Equal([]lock_manager.LockHandle{{ProjectName: "project/stage"}}))
	})
})
