package host_cleaning

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("autoHostCleanupArgs", func() {
	It("should always pin the host cleanup report off", func() {
		// The detached child inherits os.Environ(), so without this argv entry
		// WERF_SAVE_HOST_CLEANUP_REPORT would make it write a report behind the user's back.
		Expect(autoHostCleanupArgs(AutoHostCleanupOptions{})).To(ContainElement("--save-host-cleanup-report=false"))
	})

	It("should keep the report pinned off whatever else is set", func() {
		args := autoHostCleanupArgs(AutoHostCleanupOptions{
			HostCleanupOptions: HostCleanupOptions{DryRun: true, Force: true},
		})

		Expect(args).To(ContainElement("--save-host-cleanup-report=false"))
		Expect(args).NotTo(ContainElement("--save-host-cleanup-report=true"))
	})
})
