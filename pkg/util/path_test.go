package util_test

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/util"
)

type entry struct {
	path               string
	expectedPathFormat string
}

var _ = DescribeTable("expand path",
	func(e entry) {
		usr, err := user.Current()
		Ω(err).ShouldNot(HaveOccurred())

		wd, err := os.Getwd()
		Ω(err).ShouldNot(HaveOccurred())

		expectedPath := fmt.Sprintf(e.expectedPathFormat, usr.HomeDir, wd)
		Ω(util.ExpandPath(filepath.FromSlash(e.path))).Should(Equal(filepath.FromSlash(expectedPath)))
	},
	Entry("~", entry{
		path:               "~",
		expectedPathFormat: "%[1]s",
	}),
	Entry("~/", entry{
		path:               "~/",
		expectedPathFormat: "%[1]s",
	}),
	Entry("~/path", entry{
		path:               "~/path",
		expectedPathFormat: "%[1]s/path",
	}),
	Entry("path", entry{
		path:               "path",
		expectedPathFormat: "%[2]s/path",
	}),
	Entry("path1/../path2", entry{
		path:               "path1/../path2",
		expectedPathFormat: "%[2]s/path2",
	}),
)
