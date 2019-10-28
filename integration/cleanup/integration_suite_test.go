// +build integration integration_k8s

package cleanup_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/flant/werf/integration/utils"
	utilsDocker "github.com/flant/werf/integration/utils/docker"
)

func TestIntegration(t *testing.T) {
	if !utils.MeetsRequirements(requiredSuiteTools, requiredSuiteEnvs) {
		fmt.Println("Missing required tools")
		os.Exit(1)
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Cleanup Suite")
}

var requiredSuiteTools = []string{"git", "docker"}
var requiredSuiteEnvs []string

var tmpDir string
var werfBinPath string

var _ = SynchronizedBeforeSuite(func() []byte {
	pathToWerf := utils.ProcessWerfBinPath()
	return []byte(pathToWerf)
}, func(computedPathToWerf []byte) {
	werfBinPath = string(computedPathToWerf)
})

var _ = BeforeEach(func() {
	var err error
	tmpDir, err = utils.GetTempDir()
	Ω(err).ShouldNot(HaveOccurred())

	utils.BeforeEachOverrideWerfProjectName()
})

var _ = AfterEach(func() {
	err := os.RemoveAll(tmpDir)
	Ω(err).ShouldNot(HaveOccurred())

	utils.ResetEnviron()
})

var _ = SynchronizedAfterSuite(func() {}, func() {
	gexec.CleanupBuildArtifacts()
})

func tmpPath(paths ...string) string {
	pathsToJoin := append([]string{tmpDir}, paths...)
	return filepath.Join(pathsToJoin...)
}

func fixturePath(paths ...string) string {
	pathsToJoin := append([]string{"_fixtures"}, paths...)
	return filepath.Join(pathsToJoin...)
}

func LocalProjectStagesCount() int {
	filterSet := filters.NewArgs()
	filterSet.Add("reference", strings.Join([]string{"werf-stages-storage", os.Getenv("WERF_PROJECT_NAME")}, "/"))
	options := types.ImageListOptions{Filters: filterSet}
	images, err := utilsDocker.Images(options)
	Ω(err).ShouldNot(HaveOccurred())
	return len(images)
}
