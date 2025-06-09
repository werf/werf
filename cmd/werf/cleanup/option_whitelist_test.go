package cleanup

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/werf/werf/v2/pkg/cleaning"
)

var _ = Describe("whitelist", func() {
	DescribeTable(
		"parseWhitelist",
		func(fileContent string, expectedValue, expectedErr types.GomegaMatcher) {
			file, err := os.CreateTemp(GinkgoT().TempDir(), "whitelist-")
			Expect(err).To(Succeed())

			_, err = file.WriteString(fileContent)
			Expect(err).To(Succeed())

			whitelist, err := parseWhitelist(file.Name())
			Expect(err).To(expectedErr)
			Expect(whitelist).To(expectedValue)
		},
		Entry(
			"should return err if whitelist contains invalid data",
			"some invalid data",
			BeNil(),
			HaveOccurred(),
		),
		Entry(
			"should return empty whitelist if no content",
			"",
			Equal(cleaning.NewWhitelistWithSize(0)),
			Succeed(),
		),
		Entry(
			"should return empty whitelist if file contains only empty lines",
			"\n\n",
			Equal(cleaning.NewWhitelistWithSize(0)),
			Succeed(),
		),
		Entry(
			"should return filled whitelist with data",
			"1e09fb543b4ef442ce5ed36bfeee6b27866bf1e68541db5995962b24-1749456960043\n18c3b56662bedc24f4b8fd9e13845b01cc25c49295f479ac33397e27-1749456950030\n",
			Equal(
				cleaning.NewWhitelist(
					"1e09fb543b4ef442ce5ed36bfeee6b27866bf1e68541db5995962b24-1749456960043",
					"18c3b56662bedc24f4b8fd9e13845b01cc25c49295f479ac33397e27-1749456950030",
				),
			),
			Succeed(),
		),
	)
})
