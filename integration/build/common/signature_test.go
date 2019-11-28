// +build integration

package common_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/flant/werf/integration/utils"
)

var _ = Describe("persistent stage signatures", func() {
	BeforeEach(func() {
		utils.CopyIn(fixturePath("signature"), testDirPath)

		utils.RunSucceedCommand(
			testDirPath,
			"git",
			"init",
		)

		utils.RunSucceedCommand(
			testDirPath,
			"git",
			"add", "-A",
		)

		utils.RunSucceedCommand(
			testDirPath,
			"git",
			"commit", "-m", "Initial commit",
		)
	})

	It("should not be changed", func() {
		output, err := utils.RunCommand(
			testDirPath,
			werfBinPath,
			"run", "-s", ":local",
		)

		Ω(err).Should(HaveOccurred())

		for _, signature := range []string{
			"dockerfile:             268b26f069427bdeb1f7f29756d8c5b6ccb58a569bde747b1933c2e33c9007b8",
			"from:                   1fd881d714bc93857277419810a4b0bc276ad90435b1b16d21a8a36507354303",
			"beforeInstall:          39505e9134789f71716aeb1dfdbef6e163b3dc6b79f768ab35f63eb21499b38e",
			"importsBeforeInstall:   8083303ecaa2699bd3baaca1fd8d3c79699a047c18e60f2ded1f496e01d1c110",
			"install:                962fad5e56e72d80ff9dbe8530fdc21ead8340eef6f064b746a07f62a3f52e18",
			"beforeSetup:            e08c7c6211e028a058cd144d8ea6218d0e1bc7e52765e8e569467a693d47c2a5",
			"setup:                  805dc430725ef3c9e788c415da116d3369aa2804bbab9409cd7f3b5b1785954f",
			"importsAfterSetup:      e1a251bab5af242f14a6ff034ad33e84a68f0040aa312a64d72aaeec8fdc8167",
			"dockerInstructions:     7a5a8f49b55a9dfc03c29d9208fe6e04304fb10e0d92785931051a8350f53bb0",
		} {
			Ω(string(output)).Should(ContainSubstring(signature))
		}
	})
})
