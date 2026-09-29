package e2e_converge_test

import (
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/werf/werf/v3/test/pkg/utils"
	"github.com/werf/werf/v3/test/pkg/werf"
)

var _ = ginkgo.Describe("Live resource policy", ginkgo.Label("e2e", "converge", "simple", "live-resource-policy"), func() {
	var werfProject *werf.Project

	ginkgo.AfterEach(func(ctx ginkgo.SpecContext) {
		if werfProject == nil {
			return
		}

		werfProject.KubeCtl(ctx, &werf.KubeCtlOptions{
			CommonOptions: werf.CommonOptions{
				ExtraArgs: []string{"delete", "namespace", "--ignore-not-found", werfProject.Namespace(ctx)},
			},
		})
	})

	ginkgo.DescribeTable("preserves live-only retention policies",
		func(ctx ginkgo.SpecContext, uninstall bool) {
			setupEnv()
			SuiteData.InitTestRepo(ctx, "repo0", "simple/state0")
			repoPath := SuiteData.GetTestRepoPath("repo0")
			utils.CopyIn("testdata/live-resource-policy", filepath.Join(repoPath, ".helm"))
			utils.RunSucceedCommand(ctx, repoPath, "git", "add", ".helm")
			utils.RunSucceedCommand(ctx, repoPath, "git", "commit", "-m", "add retention resources")
			werfProject = werf.NewProject(SuiteData.WerfBinPath, repoPath)

			werfProject.Converge(ctx, nil)

			configMaps := werf.NewKubeClientFactory(ctx).Static().CoreV1().ConfigMaps(werfProject.Namespace(ctx))
			policies := []struct {
				name, annotation, value string
			}{
				{"helm-keep", "helm.sh/resource-policy", "keep"},
				{"werf-keep", "werf.io/resource-policy", "keep"},
				{"werf-skip-delete", "werf.io/resource-policy", "skip-delete"},
			}
			uids := make(map[string]types.UID)
			for _, policy := range policies {
				cm, err := configMaps.Get(ctx, policy.name, metav1.GetOptions{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				uids[policy.name] = cm.UID
				if cm.Annotations == nil {
					cm.Annotations = make(map[string]string)
				}
				cm.Annotations[policy.annotation] = policy.value
				_, err = configMaps.Update(ctx, cm, metav1.UpdateOptions{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			}
			_, err := configMaps.Get(ctx, "unprotected", metav1.GetOptions{})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			if uninstall {
				werfProject.RunCommand(ctx, []string{"dismiss"}, werf.CommonOptions{})
			} else {
				werfProject.Converge(ctx, &werf.ConvergeOptions{
					CommonOptions: werf.CommonOptions{ExtraArgs: []string{"--set", "removeResources=true"}},
				})
			}

			for _, policy := range policies {
				cm, err := configMaps.Get(ctx, policy.name, metav1.GetOptions{})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(cm.UID).To(gomega.Equal(uids[policy.name]))
				gomega.Expect(cm.Data).To(gomega.Equal(map[string]string{"retained": "original"}))
				gomega.Expect(cm.Annotations).To(gomega.HaveKeyWithValue(policy.annotation, policy.value))
			}
			_, err = configMaps.Get(ctx, "unprotected", metav1.GetOptions{})
			gomega.Expect(apierrors.IsNotFound(err)).To(gomega.BeTrue(), "unprotected resource must be deleted: %v", err)
		},
		ginkgo.Entry("when resources are removed from the chart", false),
		ginkgo.Entry("when the release is uninstalled", true),
	)
})
