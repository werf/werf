package config_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/test/pkg/utils"
)

type listEntry struct {
	extraArgs              []string
	expectedImagesNames    []string
	notExpectedImagesNames []string
}

var listItBody = func(entry listEntry) {
	SuiteData.CommitProjectWorktree(SuiteData.ProjectName, utils.FixturePath("list"), "initial commit")

	werfArgs := []string{"config", "list"}
	werfArgs = append(werfArgs, entry.extraArgs...)

	output := utils.SucceedCommandOutputString(
		SuiteData.GetProjectWorktree(SuiteData.ProjectName),
		SuiteData.WerfBinPath,
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
