package common

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v3/pkg/config"
)

var _ = Describe("GetIntrospectOptions", func() {
	It("names the config file the images were loaded from for an unknown image", func() {
		werfConfig := config.NewWerfConfig(&config.Meta{Project: "test-project"}, nil)
		werfConfig.SourceFilePath = "another.werf.yaml"
		cmdData := &CmdData{StagesToIntrospect: &[]string{"wrong-image/from"}}

		_, err := GetIntrospectOptions(cmdData, werfConfig)
		Expect(err).To(MatchError(`specified image "wrong-image" ("wrong-image/from") is not defined in another.werf.yaml`))
	})

	It("falls back to the default config name when the config has no source file", func() {
		werfConfig := config.NewWerfConfig(&config.Meta{Project: "test-project"}, nil)
		cmdData := &CmdData{StagesToIntrospect: &[]string{"wrong-image/from"}}

		_, err := GetIntrospectOptions(cmdData, werfConfig)
		Expect(err).To(MatchError(`specified image "wrong-image" ("wrong-image/from") is not defined in werf.yaml`))
	})
})
