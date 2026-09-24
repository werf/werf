package config

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("StapelImageBase.exportsAutoExcluding", func() {
	DescribeTable("should succeed without error",
		func(img *StapelImageBase) {
			Expect(img.exportsAutoExcluding()).To(Succeed())
		},
		Entry("two imports to same directory", &StapelImageBase{
			Import: []*Import{
				newTestImport("/src1", "/app"),
				newTestImport("/src2", "/app"),
			},
			raw: &rawStapelImage{doc: &doc{}},
		}),
		Entry("three imports to same directory", &StapelImageBase{
			Import: []*Import{
				newTestImport("/src1", "/app"),
				newTestImport("/src2", "/app"),
				newTestImport("/src3", "/app"),
			},
			raw: &rawStapelImage{doc: &doc{}},
		}),
		Entry("two imports with partially overlapping paths", &StapelImageBase{
			Import: []*Import{
				newTestImport("/src1", "/app"),
				newTestImport("/src2", "/app/data"),
			},
			raw: &rawStapelImage{doc: &doc{}},
		}),
	)

	DescribeTable("should return error",
		func(img *StapelImageBase) {
			Expect(img.exportsAutoExcluding()).ToNot(Succeed())
		},
		Entry("git local + import overlap without includePaths", &StapelImageBase{
			Git:    &GitManager{Local: []*GitLocal{newTestGitLocal(&ExportBase{Add: "/", To: "/app"})}},
			Import: []*Import{newTestImport("/src", "/app")},
			raw:    &rawStapelImage{doc: &doc{}},
		}),
		Entry("git remote + import overlap without includePaths", &StapelImageBase{
			Git:    &GitManager{Remote: []*GitRemote{newTestGitRemote(&ExportBase{Add: "/", To: "/app"})}},
			Import: []*Import{newTestImport("/src", "/app")},
			raw:    &rawStapelImage{doc: &doc{}},
		}),
		Entry("two git locals overlap without includePaths", &StapelImageBase{
			Git: &GitManager{Local: []*GitLocal{
				newTestGitLocal(&ExportBase{Add: "/a", To: "/app"}),
				newTestGitLocal(&ExportBase{Add: "/b", To: "/app"}),
			}},
			raw: &rawStapelImage{doc: &doc{}},
		}),
	)

	DescribeTable("should auto-exclude and succeed",
		func(makeImg func() (*StapelImageBase, *ExportBase), expectedExclude string) {
			img, target := makeImg()
			Expect(img.exportsAutoExcluding()).To(Succeed())
			Expect(target.ExcludePaths).To(ContainElement(expectedExclude))
		},
		Entry("git local auto-excluded by import with includePaths",
			func() (*StapelImageBase, *ExportBase) {
				eb := &ExportBase{Add: "/", To: "/"}
				return &StapelImageBase{
					Git:    &GitManager{Local: []*GitLocal{newTestGitLocal(eb)}},
					Import: []*Import{newTestImport("/src", "/app", "data")},
					raw:    &rawStapelImage{doc: &doc{}},
				}, eb
			}, "app/data",
		),
		Entry("git local auto-excluded by import targeting root with includePaths",
			func() (*StapelImageBase, *ExportBase) {
				eb := &ExportBase{Add: "/", To: "/"}
				return &StapelImageBase{
					Git:    &GitManager{Local: []*GitLocal{newTestGitLocal(eb)}},
					Import: []*Import{newTestImport("/src", "/", "data")},
					raw:    &rawStapelImage{doc: &doc{}},
				}, eb
			}, "data",
		),
		Entry("git local auto-excluded by another git local with includePaths",
			func() (*StapelImageBase, *ExportBase) {
				eb := &ExportBase{Add: "/a", To: "/"}
				return &StapelImageBase{
					Git: &GitManager{Local: []*GitLocal{
						newTestGitLocal(eb),
						newTestGitLocal(&ExportBase{Add: "/b", To: "/", IncludePaths: []string{"stuff"}}),
					}},
					raw: &rawStapelImage{doc: &doc{}},
				}, eb
			}, "stuff",
		),
	)
})

func newTestImport(add, to string, includePaths ...string) *Import {
	return &Import{
		Export: &Export{ExportBase: &ExportBase{Add: add, To: to, IncludePaths: includePaths}},
		raw:    &rawImport{rawStapelImage: &rawStapelImage{doc: &doc{}}},
	}
}

func newTestGitLocal(eb *ExportBase) *GitLocal {
	return &GitLocal{
		GitLocalExport: &GitLocalExport{GitExportBase: &GitExportBase{
			GitExport: &GitExport{ExportBase: eb, raw: &rawGitExport{}},
		}},
		raw: &rawGit{rawStapelImage: &rawStapelImage{doc: &doc{}}},
	}
}

func newTestGitRemote(eb *ExportBase) *GitRemote {
	return &GitRemote{
		GitRemoteExport: &GitRemoteExport{GitLocalExport: &GitLocalExport{GitExportBase: &GitExportBase{
			GitExport: &GitExport{ExportBase: eb, raw: &rawGitExport{}},
		}}},
		raw: &rawGit{rawStapelImage: &rawStapelImage{doc: &doc{}}},
	}
}
