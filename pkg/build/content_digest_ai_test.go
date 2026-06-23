package build

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("calculateContentDigest", func() {
	const targetPlatform = "linux/amd64"

	It("is deterministic for the same contributing stages", func() {
		deps := []string{"from-digest", "git-archive-digest", "run-digest"}
		Expect(calculateContentDigest(targetPlatform, deps)).
			To(Equal(calculateContentDigest(targetPlatform, deps)))
	})

	It("ignores empty stage contributions (gitCache/gitLatestPatch)", func() {
		withGitArchiveOnly := []string{"from-digest", "git-archive-digest"}
		withEmptyGitStages := []string{"from-digest", "git-archive-digest", "", ""}

		Expect(calculateContentDigest(targetPlatform, withGitArchiveOnly)).
			To(Equal(calculateContentDigest(targetPlatform, withEmptyGitStages)))
	})

	It("is unaffected by the position of empty contributions", func() {
		base := []string{"from-digest", "git-archive-digest", "run-digest"}
		withEmptyInterleaved := []string{"from-digest", "", "git-archive-digest", "", "run-digest", ""}

		Expect(calculateContentDigest(targetPlatform, base)).
			To(Equal(calculateContentDigest(targetPlatform, withEmptyInterleaved)))
	})

	It("is unaffected by absent optional empty stages (install/setup/dependencies)", func() {
		withoutOptional := []string{"from-digest", "git-archive-digest"}
		withOptionalEmpty := []string{"", "from-digest", "", "git-archive-digest", ""}

		Expect(calculateContentDigest(targetPlatform, withoutOptional)).
			To(Equal(calculateContentDigest(targetPlatform, withOptionalEmpty)))
	})

	It("changes when a contributing stage changes", func() {
		Expect(calculateContentDigest(targetPlatform, []string{"from-digest", "git-archive-v1"})).
			NotTo(Equal(calculateContentDigest(targetPlatform, []string{"from-digest", "git-archive-v2"})))
	})

	It("changes when the target platform changes", func() {
		deps := []string{"from-digest", "git-archive-digest"}
		Expect(calculateContentDigest("linux/amd64", deps)).
			NotTo(Equal(calculateContentDigest("linux/arm64", deps)))
	})
})
