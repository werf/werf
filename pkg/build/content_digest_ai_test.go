package build

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// anchorDigest is a thin wrapper that exercises the anchor branch of
// calculateDigest: pass HolisticInputs, get the platform-scoped hash.
// Semantically identical to the pre-merge calculateContentDigest so
// previously published content-tag images remain reusable byte-for-byte.
func anchorDigest(targetPlatform string, deps []string) string {
	d, err := calculateDigest(context.Background(), "anchor", "", nil, nil, calculateDigestOptions{
		TargetPlatform: targetPlatform,
		HolisticInputs: deps,
	})
	if err != nil {
		panic(err)
	}
	return d
}

var _ = Describe("anchor holistic digest", func() {
	const targetPlatform = "linux/amd64"

	It("is deterministic for the same contributing stages", func() {
		deps := []string{"from-digest", "git-archive-digest", "run-digest"}
		Expect(anchorDigest(targetPlatform, deps)).
			To(Equal(anchorDigest(targetPlatform, deps)))
	})

	It("ignores empty stage contributions (gitCache/gitLatestPatch)", func() {
		withGitArchiveOnly := []string{"from-digest", "git-archive-digest"}
		withEmptyGitStages := []string{"from-digest", "git-archive-digest", "", ""}

		Expect(anchorDigest(targetPlatform, withGitArchiveOnly)).
			To(Equal(anchorDigest(targetPlatform, withEmptyGitStages)))
	})

	It("is unaffected by the position of empty contributions", func() {
		base := []string{"from-digest", "git-archive-digest", "run-digest"}
		withEmptyInterleaved := []string{"from-digest", "", "git-archive-digest", "", "run-digest", ""}

		Expect(anchorDigest(targetPlatform, base)).
			To(Equal(anchorDigest(targetPlatform, withEmptyInterleaved)))
	})

	It("is unaffected by absent optional empty stages (install/setup/dependencies)", func() {
		withoutOptional := []string{"from-digest", "git-archive-digest"}
		withOptionalEmpty := []string{"", "from-digest", "", "git-archive-digest", ""}

		Expect(anchorDigest(targetPlatform, withoutOptional)).
			To(Equal(anchorDigest(targetPlatform, withOptionalEmpty)))
	})

	It("changes when a contributing stage changes", func() {
		Expect(anchorDigest(targetPlatform, []string{"from-digest", "git-archive-v1"})).
			NotTo(Equal(anchorDigest(targetPlatform, []string{"from-digest", "git-archive-v2"})))
	})

	It("changes when the target platform changes", func() {
		deps := []string{"from-digest", "git-archive-digest"}
		Expect(anchorDigest("linux/amd64", deps)).
			NotTo(Equal(anchorDigest("linux/arm64", deps)))
	})
})
