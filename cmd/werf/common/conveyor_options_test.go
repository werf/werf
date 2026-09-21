package common

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/slug"
)

var _ = DescribeTable("custom tag image name substitutions", func(tagFormat string, expectError bool) {
	repoAddress := "registry.example.com/project"
	imageName := "libstdc++"
	cmdData := &CmdData{Repo: &RepoData{Address: &repoAddress}}
	imagesToProcess := config.ImagesToProcess{FinalImageNameList: []string{imageName}}

	customTagFuncList, err := getCustomTagFuncList([]string{tagFormat}, cmdData, imagesToProcess)
	if expectError {
		Expect(err).To(HaveOccurred())
		return
	}

	Expect(err).NotTo(HaveOccurred())
	Expect(customTagFuncList).To(HaveLen(1))
	Expect(customTagFuncList[0](imageName, "content-based-tag")).To(Equal("libstdc"))
	Expect(slug.ValidateDockerTag(customTagFuncList[0](imageName, "content-based-tag"))).To(Succeed())
},
	Entry("image name", "%image%", true),
	Entry("slugged image name", "%image_slug%", false),
)
