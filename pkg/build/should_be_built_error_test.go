package build

import (
	"bytes"
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/logboek"
	"github.com/werf/werf/v3/pkg/build/stage"
)

var _ = ginkgo.Describe("Missing stage report", func() {
	printError := func() string {
		phase := newTestBuildPhase(&checkModeStorageManager{stagesStorage: &checkModeStorage{}}, nil)
		phase.ShouldBeBuiltMode = true
		img := newTestImage("app", true)
		stg := stage.NewBaseStage(stage.Setup, &stage.BaseStageOptions{ImageName: "app"})
		stg.SetDigest("missing-digest")

		var stdout, stderr bytes.Buffer
		ctx := logboek.NewContext(context.Background(), logboek.NewLogger(&stdout, &stderr))
		phase.printShouldBeBuiltError(ctx, img, stg)

		return stderr.String()
	}

	ginkgo.It("explains that only the primary stages storage was searched", func() {
		output := printError()

		gomega.Expect(output).To(gomega.ContainSubstring("check-mode-test"), "the primary stages storage address must be named")
		gomega.Expect(output).To(gomega.ContainSubstring("--secondary-repo"), "the skipped secondary stages storages must be named")
		gomega.Expect(output).To(gomega.ContainSubstring("werf build --repo check-mode-test"), "an ordinary build must be advised")
		gomega.Expect(output).To(gomega.ContainSubstring("--check-built-images"))
		gomega.Expect(output).To(gomega.ContainSubstring("WERF_CHECK_BUILT_IMAGES"))
		gomega.Expect(output).To(gomega.ContainSubstring("Stages have not been built yet"), "the still applicable reasons must be kept")
	})
})
