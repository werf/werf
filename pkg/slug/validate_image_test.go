package slug

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = DescribeTable("ValidateImage", func(name string, valid bool) {
	err := ValidateImage(name)
	if valid {
		Expect(err).To(Succeed())
	} else {
		Expect(err).To(MatchError(ContainSubstring("should comply with regex")))
	}
},
	Entry("plain name", "backend", true),
	Entry("digits", "app2", true),
	Entry("dashes", "base-stapel-shell", true),
	Entry("underscores", "dockerfile_base_image_cmd", true),
	Entry("mixed case", "FromExternalImage", true),
	Entry("dots", "app.v2", true),
	Entry("slash-separated segments", "base/vex", true),
	Entry("several segments", "modules/ingress/controller", true),
	Entry("digit-leading segments", "1c/8bit", true),

	Entry("empty", "", false),
	Entry("blank", "   ", false),
	Entry("whitespace inside", "my app", false),
	Entry("tab inside", "my\tapp", false),
	Entry("newline inside", "app\n", false),
	Entry("leading dash", "-app", false),
	Entry("leading dot", ".app", false),
	Entry("leading underscore", "_app", false),
	Entry("trailing dash", "app-", false),
	Entry("trailing dot", "app.", false),
	Entry("trailing underscore", "app_", false),
	Entry("trailing dash in segment", "modules/api-/x", false),
	Entry("single character", "a", true),
	Entry("leading slash", "/app", false),
	Entry("trailing slash", "app/", false),
	Entry("empty segment", "modules//controller", false),
	Entry("unrendered template left over", "{{ .ImageName }}", false),
	Entry("template rendered to nothing between segments", "modules/", false),
	Entry("go template no value", "<no value>", false),
	Entry("quotes", `"app"`, false),
	Entry("colon", "app:latest", false),
)
