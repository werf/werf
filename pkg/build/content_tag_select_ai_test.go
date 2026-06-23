package build

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	imagePkg "github.com/werf/werf/v2/pkg/image"
)

func newContentTagDesc(digest string, creationTs int64) *imagePkg.StageDesc {
	return &imagePkg.StageDesc{
		StageID: imagePkg.NewStageID(digest, creationTs),
		Info:    &imagePkg.Info{},
	}
}

var _ = Describe("selectLatestStageDesc", func() {
	const digest = "deadbeef"

	It("returns nil for an empty or nil set", func() {
		Expect(selectLatestStageDesc(nil)).To(BeNil())
		Expect(selectLatestStageDesc(imagePkg.NewStageDescSet())).To(BeNil())
	})

	It("picks the descriptor with the greatest creation timestamp", func() {
		older := newContentTagDesc(digest, 100)
		newer := newContentTagDesc(digest, 200)

		Expect(selectLatestStageDesc(imagePkg.NewStageDescSet(older, newer))).To(Equal(newer))
	})

	It("breaks creation-timestamp ties deterministically by stage id string", func() {
		a := newContentTagDesc("aaa", 100)
		b := newContentTagDesc("bbb", 100)

		fromAB := selectLatestStageDesc(imagePkg.NewStageDescSet(a, b))
		fromBA := selectLatestStageDesc(imagePkg.NewStageDescSet(b, a))

		Expect(fromAB).To(Equal(b))
		Expect(fromBA).To(Equal(b))
	})
})
