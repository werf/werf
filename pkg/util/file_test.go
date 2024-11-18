package util

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type isSubpathOfBasePathEntry struct {
	basePath  string
	path      string
	isSubpath bool
}

var _ = DescribeTable("IsSubpathOfBasePath", func(e isSubpathOfBasePathEntry) {
	Expect(IsSubpathOfBasePath(e.basePath, e.path)).Should(BeEquivalentTo(e.isSubpath))
},
	Entry(`equal paths ("")`, isSubpathOfBasePathEntry{
		basePath:  "",
		path:      "",
		isSubpath: false,
	}),
	Entry(`equal paths ("/")`, isSubpathOfBasePathEntry{
		basePath:  "/",
		path:      "/",
		isSubpath: false,
	}),
	Entry(`equal paths ("dir")`, isSubpathOfBasePathEntry{
		basePath:  "dir",
		path:      "dir",
		isSubpath: false,
	}),
	Entry(`subpath of empty`, isSubpathOfBasePathEntry{
		basePath:  "",
		path:      "file",
		isSubpath: true,
	}),
	Entry(`subpath of root`, isSubpathOfBasePathEntry{
		basePath:  "/",
		path:      "/file",
		isSubpath: true,
	}),
	Entry(`subpath of dir`, isSubpathOfBasePathEntry{
		basePath:  "dir",
		path:      "dir/file",
		isSubpath: true,
	}),
	Entry(`not subpath (1)`, isSubpathOfBasePathEntry{
		basePath:  "dir",
		path:      "",
		isSubpath: false,
	}),
	Entry(`not subpath (2)`, isSubpathOfBasePathEntry{
		basePath:  "dir1",
		path:      "dir2",
		isSubpath: false,
	}),
	Entry(`not subpath (3)`, isSubpathOfBasePathEntry{
		basePath:  "/dir",
		path:      "/",
		isSubpath: false,
	}),
)
