package container_backend

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LegacyStageImageContainer imageRef", func() {
	const (
		tmpUUIDName    = "80324e3c-81f8-43cc-a055-6a95a7928690"
		committedID    = "sha256:deadbeef"
		targetPlatform = "linux/amd64"
	)

	DescribeTable("returns the expected image reference", func(name string, platform string, committed bool, expected string) {
		img := NewLegacyStageImage(nil, name, nil, platform)
		if committed {
			img.buildImage = newLegacyBaseImage(committedID, nil)
			img.builtID = committedID
		}

		Expect(img.container.imageRef(img)).To(Equal(expected))
	},
		Entry("committed image with target platform returns committed id, not name", tmpUUIDName, targetPlatform, true, committedID),
		Entry("committed image without target platform returns committed id", tmpUUIDName, "", true, committedID),
		Entry("non-committed image with target platform returns name", "registry.example.com/project:stage", targetPlatform, false, "registry.example.com/project:stage"),
		Entry("non-committed image without target platform returns id from name", "registry.example.com/project:stage", "", false, "registry.example.com/project:stage"),
	)
})
