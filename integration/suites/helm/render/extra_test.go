package render_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"helm.sh/helm/v3/pkg/releaseutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"

	"github.com/werf/werf/test/pkg/utils"
)

var _ = Describe("helm render with extra annotations and labels", func() {
	BeforeEach(func() {
		SuiteData.CommitProjectWorktree(SuiteData.ProjectName, utils.FixturePath("base"), "initial commit")
	})

	It("should be rendered with extra annotations (except hooks)", func() {
		output := utils.SucceedCommandOutputString(
			SuiteData.GetProjectWorktree(SuiteData.ProjectName),
			SuiteData.WerfBinPath,
			"render", "--add-annotation=anno1=value1", "--add-annotation=anno2=value2",
		)

		splitManifests := releaseutil.SplitManifests(output)
		for _, content := range splitManifests {
			var obj unstructured.Unstructured
			Expect(yaml.Unmarshal([]byte(content), &obj)).To(Succeed())

			annotations := obj.GetAnnotations()
			labels := obj.GetLabels()

			// Hooks not supported yet by the helm
			if _, isHook := annotations["helm.sh/hook"]; isHook {
				continue
			}

			Expect(annotations["anno1"]).To(Equal("value1"))
			Expect(annotations["anno2"]).To(Equal("value2"))

			_, exists := labels["anno1"]
			Expect(exists).To(BeFalse())
			_, exists = labels["anno2"]
			Expect(exists).To(BeFalse())
		}
	})

	It("should be rendered with extra labels (except hooks)", func() {
		output := utils.SucceedCommandOutputString(
			SuiteData.GetProjectWorktree(SuiteData.ProjectName),
			SuiteData.WerfBinPath,
			"render", "--add-label=label1=value1", "--add-label=label2=value2",
		)

		splitManifests := releaseutil.SplitManifests(output)
		for _, content := range splitManifests {
			var obj unstructured.Unstructured
			Expect(yaml.Unmarshal([]byte(content), &obj)).To(Succeed())

			annotations := obj.GetAnnotations()
			labels := obj.GetLabels()

			// Hooks not supported yet by the helm
			if _, isHook := annotations["helm.sh/hook"]; isHook {
				continue
			}

			Expect(labels["label1"]).To(Equal("value1"))
			Expect(labels["label2"]).To(Equal("value2"))

			_, exists := annotations["label1"]
			Expect(exists).To(BeFalse())
			_, exists = annotations["label2"]
			Expect(exists).To(BeFalse())
		}
	})
})
