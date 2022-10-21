//go:build windows

package util_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/pkg/util"
)

type splitPathTest struct {
	path              string
	expectedPathParts []string
}

var _ = DescribeTable("split windows path",
	func(t splitPathTest) {
		parts := util.SplitFilepath(t.path)
		Expect(parts).To(Equal(t.expectedPathParts))
	},
	Entry("UNC absolute path", splitPathTest{
		path:              `\\Server2\Share\Test\Foo.txt`,
		expectedPathParts: []string{"Server2", "Share", "Test", "Foo.txt"},
	}),
	Entry("UNC absolute non-normalized path", splitPathTest{
		path:              `\\\\\\Server2\\\Share\\\Test\Foo.txt`,
		expectedPathParts: []string{"Server2", "Share", "Test", "Foo.txt"},
	}),
	Entry("UNC root path", splitPathTest{
		path:              `\\`,
		expectedPathParts: nil,
	}),
	Entry("root disk path", splitPathTest{
		path:              `D:\`,
		expectedPathParts: []string{`D:`},
	}),
	Entry("unnormalized root disk path", splitPathTest{
		path:              `D:\\\\\\`,
		expectedPathParts: []string{`D:`},
	}),
	Entry("empty path", splitPathTest{
		path:              "",
		expectedPathParts: nil,
	}),
	Entry("absolute path to file", splitPathTest{
		path:              `D:\path\to\file`,
		expectedPathParts: []string{`D:`, "path", "to", "file"},
	}),
	Entry("absolute path to dir", splitPathTest{
		path:              `D:\path\to\dir\`,
		expectedPathParts: []string{`D:`, "path", "to", "dir"},
	}),
	Entry("relative path", splitPathTest{
		path:              `path\to\dir\or\file`,
		expectedPathParts: []string{"path", "to", "dir", "or", "file"},
	}),
)
