package config

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"gopkg.in/yaml.v2"
)

var _ = Describe("ExportBase", func() {
	DescribeTable("validate()",
		func(yamlMap map[string]any, errMessage string) {
			Expect(yamlMap).NotTo(HaveLen(0), "yamlMap should not be empty")

			testConfigSection := "test config"
			testDoc := &doc{
				Content:        []byte("some content"),
				Line:           -1,
				RenderFilePath: "some file path",
			}

			rawOriginMock := NewMockrawOrigin(gomock.NewController(GinkgoT()))
			rawOriginMock.EXPECT().configSection().Return(testConfigSection)
			rawOriginMock.EXPECT().doc().Return(testDoc)

			rawExBase := &rawExportBase{rawOrigin: rawOriginMock}
			exBase := &ExportBase{raw: rawExBase}

			rawYaml, err := yaml.Marshal(yamlMap)
			Expect(err).To(Succeed())

			Expect(yaml.UnmarshalStrict(rawYaml, &exBase)).To(Succeed())

			expectedErr := newDetailedConfigError(errMessage, testConfigSection, testDoc)

			Expect(exBase.validate()).To(Equal(expectedErr))
		},
		Entry(
			"should return err if b.Add = / and b.IncludePaths=[]",
			map[string]any{
				"add": "/",
				"to":  "/some/path",
			},
			"`add: '/'` requires not empty includePaths to interpret copy sources unambiguously",
		),
	)
})
