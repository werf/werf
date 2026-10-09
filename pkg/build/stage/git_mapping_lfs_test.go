package stage_test

import (
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v3/pkg/build/stage"
)

var _ = ginkgo.Describe("automatic LFS Stapel cache", func() {
	ginkgo.It("does not reuse Git stages that contained pointer files", func() {
		mapping := stage.NewGitMapping()
		mapping.SetGitRepo(NewGitRepoStub("own", true, "commit"))
		gomega.Expect(mapping.GetFullName()).To(gomega.Equal("own"))
		gomega.Expect(mapping.GetParamshash()).NotTo(gomega.Equal(util.Sha256Hash("own" + strings.Repeat(":::", 9))))
	})
})
