package e2e_converge_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/werf/kubedog/pkg/kube"
	"github.com/werf/werf/test/pkg/utils"
	"github.com/werf/werf/test/pkg/werf"
)

var _ = Describe("Simple converge", Label("e2e", "converge", "simple"), func() {
	var repoDirname string

	BeforeEach(func() {
		Expect(kube.Init(kube.InitOptions{})).To(Succeed())
	})

	AfterEach(func() {
		utils.RunSucceedCommand(SuiteData.GetTestRepoPath(repoDirname), SuiteData.WerfBinPath, "dismiss", "--with-namespace")
	})

	It("should succeed and deploy expected resources",
		func() {
			By("initializing")
			repoDirname = "repo0"
			setupEnv()

			By("state0: starting")
			{
				fixtureRelPath := "simple/state0"

				By("state0: preparing test repo")
				SuiteData.InitTestRepo(repoDirname, fixtureRelPath)
				werfProject := werf.NewProject(SuiteData.WerfBinPath, SuiteData.GetTestRepoPath(repoDirname))

				By("state0: execute converge")
				combinedOut := werfProject.Converge(&werf.ConvergeOptions{})

				By("state0: check converge output")
				Expect(combinedOut).To(ContainSubstring("STATUS: deployed"))
				Expect(combinedOut).To(ContainSubstring("REVISION: 1"))

				By("state0: check deployed resources in cluster")
				cm, err := kube.Client.CoreV1().ConfigMaps(werfProject.Namespace()).Get(context.Background(), "test1", metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
				Expect(cm.Data).To(Equal(map[string]string{"key1": "value1"}))
			}
		},
	)
})
