package e2e_export_test

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/test/pkg/utils"
	"github.com/werf/werf/v2/test/pkg/werf"
)

type simpleTestOptions struct {
	ExtraArgs []string
}

func ManifestGet(imageName string) *v1.IndexManifest {
	bytes, err := crane.Manifest(imageName)
	if err != nil {
		fmt.Errorf("Error %s", err)
	}
	var data v1.IndexManifest
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		fmt.Errorf("Error %s", err)
	}
	return &data
}

func LabelsExist(imageName string) (result bool) {
	bytes, err := crane.Config(imageName)
	if err != nil {
		fmt.Errorf("Error, %s", err)
	}
	var data v1.ConfigFile
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		fmt.Errorf("Error %s", err)
	}
	if data.Config.Labels == nil {
		result = false
	} else {
		result = true
	}
	return
}

var _ = Describe("Simple export", Label("e2e", "export", "simple"), func() {
	DescribeTable("should succeed and export images",
		func(opts simpleTestOptions) {
			By("initializating")
			{
				SuiteData.WerfRepo = strings.Join([]string{SuiteData.RegistryLocalAddress, SuiteData.ProjectName}, "/")
				SuiteData.Stubs.SetEnv("WERF_REPO", SuiteData.WerfRepo)
				SuiteData.Stubs.SetEnv("DOCKER_BUILDKIT", "1")
				repoDirname := "repo"
				fixtureRelPath := "simple"

				By("preparing test repo")
				SuiteData.InitTestRepo(repoDirname, fixtureRelPath)

				By("running export")
				imageName := fmt.Sprintf("%s/werf-export-%s", SuiteData.RegistryLocalAddress, utils.GetRandomString(10))
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))
				exportOut := werfProject.Export(&werf.ExportOptions{
					CommonOptions: werf.CommonOptions{
						ExtraArgs: append([]string{
							"--tag",
							imageName,
						},
							opts.ExtraArgs...),
					},
				})
				Expect(exportOut).To(ContainSubstring("Exporting image..."))
				By("check image architecture")
				manifest := ManifestGet(imageName)
				for i := range manifest.Manifests {
					if manifest.Manifests[i].Platform.Architecture == "amd64" {
						Expect(manifest.Manifests[i].Platform.Architecture).To(Equal("amd64"))
					} else {
						if manifest.Manifests[i].Platform.Architecture != "unknown" {
							Expect(manifest.Manifests[i].Platform.Architecture).To(Equal("arm64"))
						}
					}
				}
				By("check labels absence")
				Expect(LabelsExist(imageName)).To(Equal(false))
			}
		},
		Entry("Export single platform image", simpleTestOptions{
			ExtraArgs: []string{},
		}),
		Entry("Export multiplatform image", simpleTestOptions{
			ExtraArgs: []string{
				"--platform",
				"linux/amd64,linux/arm64",
			},
		}),
	)
})
