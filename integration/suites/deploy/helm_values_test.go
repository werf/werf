package deploy_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/werf/pkg/util"
	"github.com/werf/werf/test/pkg/utils"
	"github.com/werf/werf/test/pkg/utils/liveexec"
)

func getValues(params ...string) map[string]interface{} {
	output := utils.SucceedCommandOutputString(
		SuiteData.GetProjectWorktree(SuiteData.ProjectName),
		SuiteData.WerfBinPath,
		append([]string{"helm", "get", "values", SuiteData.ProjectName, "--namespace", SuiteData.ProjectName}, params...)...,
	)

	lines := util.SplitLines(output)
	lines = lines[1:]
	output = strings.Join(lines, "\n")

	var data map[string]interface{}
	err := yaml.Unmarshal([]byte(output), &data)
	Expect(err).To(Succeed())

	return data
}

func getUserSuppliedValues() map[string]interface{} {
	return getValues()
}

func getComputedValues() map[string]interface{} {
	return getValues("--all")
}

var _ = Describe("Helm values", func() {
	BeforeEach(func() {
		Expect(kube.Init(kube.InitOptions{})).To(Succeed())
	})

	Context("explicit values param may break default values changes in further deploys: https://github.com/werf/werf/issues/4478", func() {
		It("ignores chagnes in the .helm/values.yaml", func() {
			By("Installing release first time with basic .helm/values.yaml")
			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "helm_values1-001", "initial commit")
			Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).To(Succeed())

			{
				uv := getUserSuppliedValues()
				Expect(len(uv)).To(Equal(0))

				cv := getComputedValues()
				Expect(cv).To(Equal(map[string]interface{}{
					"arr": []interface{}{"one", "two", "three"},
				}))
			}

			By("Upgrading release with explicit --values param and the same values file")
			Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{}, "--values", ".helm/values.yaml")).To(Succeed())

			{
				uv := getUserSuppliedValues()
				Expect(uv).To(Equal(map[string]interface{}{
					"arr": []interface{}{"one", "two", "three"},
				}))

				cv := getComputedValues()
				Expect(cv).To(Equal(map[string]interface{}{
					"arr": []interface{}{"one", "two", "three"},
				}))
			}

			By("Upgrading release without explicit values param and changed values file")
			SuiteData.CommitProjectWorktree(SuiteData.ProjectName, "helm_values1-002", "append new value into default values array")
			Expect(werfConverge(SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).To(Succeed())

			{
				uv := getUserSuppliedValues()
				Expect(uv).To(Equal(map[string]interface{}{
					"arr": []interface{}{"one", "two", "three"},
				}))

				cv := getComputedValues()
				Expect(cv).To(Equal(map[string]interface{}{
					"arr": []interface{}{"one", "two", "three"},
				}))
			}
		})
	})
})
