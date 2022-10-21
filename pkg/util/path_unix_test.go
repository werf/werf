//go:build !windows

package util_test

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/util"
)

type expandPathTest struct {
	path               string
	expectedPathFormat string
}

var _ = DescribeTable("expand path",
	func(e expandPathTest) {
		usr, err := user.Current()
		Ω(err).ShouldNot(HaveOccurred())

		wd, err := os.Getwd()
		Ω(err).ShouldNot(HaveOccurred())

		expectedPath := fmt.Sprintf(e.expectedPathFormat, usr.HomeDir, wd)
		Ω(util.ExpandPath(filepath.FromSlash(e.path))).Should(Equal(filepath.FromSlash(expectedPath)))
	},
	Entry("~", expandPathTest{
		path:               "~",
		expectedPathFormat: "%[1]s",
	}),
	Entry("~/", expandPathTest{
		path:               "~/",
		expectedPathFormat: "%[1]s",
	}),
	Entry("~/path", expandPathTest{
		path:               "~/path",
		expectedPathFormat: "%[1]s/path",
	}),
	Entry("path", expandPathTest{
		path:               "path",
		expectedPathFormat: "%[2]s/path",
	}),
	Entry("path1/../path2", expandPathTest{
		path:               "path1/../path2",
		expectedPathFormat: "%[2]s/path2",
	}),
)

type splitPathTest struct {
	path              string
	expectedPathParts []string
}

var _ = DescribeTable("split unix path",
	func(t splitPathTest) {
		parts := util.SplitFilepath(t.path)
		Expect(parts).To(Equal(t.expectedPathParts))
	},
	Entry("root path", splitPathTest{
		path:              "/",
		expectedPathParts: nil,
	}),
	Entry("root dir", splitPathTest{
		path:              "/mydir/",
		expectedPathParts: []string{"mydir"},
	}),
	Entry("unnormalized root path", splitPathTest{
		path:              "////",
		expectedPathParts: nil,
	}),
	Entry("empty path", splitPathTest{
		path:              "",
		expectedPathParts: nil,
	}),
	Entry("absolute path", splitPathTest{
		path:              "/path/to/dir/or/file",
		expectedPathParts: []string{"path", "to", "dir", "or", "file"},
	}),
	Entry("relative path", splitPathTest{
		path:              "path/to/dir/or/file",
		expectedPathParts: []string{"path", "to", "dir", "or", "file"},
	}),
)
