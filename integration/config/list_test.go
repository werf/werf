package config_test

import (
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/flant/werf/pkg/testing/utils"
)

type listEntry struct {
	extraArgs              []string
	expectedImagesNames    []string
	notExpectedImagesNames []string
}

var listItBody = func(entry listEntry) {
	testDirPath = utils.FixturePath("list")

	werfArgs := []string{"config", "list"}
	werfArgs = append(werfArgs, entry.extraArgs...)

	output := utils.SucceedCommandOutputString(
		testDirPath,
		werfBinPath,
		werfArgs...,
	)

	lines := utils.StringToLines(output)
	for _, name := range entry.expectedImagesNames {
		Ω(lines).Should(ContainElement(name))
	}

	for _, name := range entry.notExpectedImagesNames {
		Ω(lines).ShouldNot(ContainElement(name))
	}
}

var _ = DescribeTable("config list", listItBody,
	Entry("all", listEntry{
		extraArgs:           []string{},
		expectedImagesNames: []string{"image_a", "image_b", "image_c", "artifact"},
	}),
	Entry("images only", listEntry{
		extraArgs:              []string{"--images-only"},
		expectedImagesNames:    []string{"image_a", "image_b", "image_c"},
		notExpectedImagesNames: []string{"artifact"},
	}))
