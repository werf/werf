package deploy_test

import (
	"context"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/werf/v2/test/pkg/utils"
	"github.com/werf/werf/v2/test/pkg/utils/liveexec"
)

func getValues(ctx context.Context, params ...string) map[string]interface{} {
	output := utils.SucceedCommandOutputString(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), SuiteData.WerfBinPath, append([]string{"helm", "get", "values", SuiteData.ProjectName, "--namespace", SuiteData.ProjectName}, params...)...)

	lines := util.SplitLines(output)
	lines = lines[1:]
	output = strings.Join(lines, "\n")

	var data map[string]interface{}
	err := yaml.Unmarshal([]byte(output), &data)
	Expect(err).To(Succeed())

	return data
}

func getUserSuppliedValues(ctx context.Context) map[string]interface{} {
	return getValues(ctx)
}

func getComputedValues(ctx context.Context) map[string]interface{} {
	return getValues(ctx, "--all")
}

var _ = Describe("Helm values", Pending, func() {
	BeforeEach(func() {
		Expect(kube.Init(kube.InitOptions{})).To(Succeed())
	})

	Context("explicit values param may break default values changes in further deploys: https://github.com/werf/werf/issues/4478", func() {
		It("ignores chagnes in the .helm/values.yaml", func(ctx SpecContext) {
			By("Installing release first time with basic .helm/values.yaml")
			SuiteData.CommitProjectWorktree(ctx, SuiteData.ProjectName, "helm_values1-001", "initial commit")
			Expect(werfConverge(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).To(Succeed())

			{
				uv := getUserSuppliedValues(ctx)
				Expect(len(uv)).To(Equal(0))

				cv := getComputedValues(ctx)
				Expect(cv).To(Equal(map[string]interface{}{
					"arr": []interface{}{"one", "two", "three"},
				}))
			}

			By("Upgrading release with explicit --values param and the same values file")
			Expect(werfConverge(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{}, "--values", ".helm/values.yaml")).To(Succeed())

			{
				uv := getUserSuppliedValues(ctx)
				Expect(uv).To(Equal(map[string]interface{}{
					"arr": []interface{}{"one", "two", "three"},
				}))

				cv := getComputedValues(ctx)
				Expect(cv).To(Equal(map[string]interface{}{
					"arr": []interface{}{"one", "two", "three"},
				}))
			}

			By("Upgrading release without explicit values param and changed values file")
			SuiteData.CommitProjectWorktree(ctx, SuiteData.ProjectName, "helm_values1-002", "append new value into default values array")
			Expect(werfConverge(ctx, SuiteData.GetProjectWorktree(SuiteData.ProjectName), liveexec.ExecCommandOptions{})).To(Succeed())

			{
				uv := getUserSuppliedValues(ctx)
				Expect(uv).To(Equal(map[string]interface{}{
					"arr": []interface{}{"one", "two", "three"},
				}))

				cv := getComputedValues(ctx)
				Expect(cv).To(Equal(map[string]interface{}{
					"arr": []interface{}{"one", "two", "three"},
				}))
			}
		})
	})
})
