package file_reader_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/giterminism_manager/file_reader"
	"github.com/werf/werf/v2/pkg/logging"
	"github.com/werf/werf/v2/pkg/path_matcher"
	"github.com/werf/werf/v2/test/mock"
)

var _ = Describe("LoadChartDir", func() {
	t := GinkgoT()

	var (
		reader            file_reader.FileReader
		sharedOptions     *MocksharedOptions
		giterminismConfig *MockgiterminismConfig
		pathMatcher       *mock.MockPathMatcher
		gitRepo           *mock.MockGitRepo
		projectDir        string
	)

	BeforeEach(func() {
		ctrl := gomock.NewController(t, gomock.WithOverridableExpectations())

		sharedOptions = NewMocksharedOptions(ctrl)
		giterminismConfig = NewMockgiterminismConfig(ctrl)
		pathMatcher = mock.NewMockPathMatcher(ctrl)
		gitRepo = mock.NewMockGitRepo(ctrl)

		projectDir = t.TempDir()

		sharedOptions.EXPECT().ProjectDir().Return(projectDir).AnyTimes()
		sharedOptions.EXPECT().RelativeToGitProjectDir().Return("").AnyTimes()
		sharedOptions.EXPECT().LocalGitRepo().Return(gitRepo).AnyTimes()
		sharedOptions.EXPECT().HeadCommit(gomock.Any()).Return("head commit").AnyTimes()
		sharedOptions.EXPECT().Dev().Return(false).AnyTimes()
		sharedOptions.EXPECT().LooseGiterminism().Return(true).AnyTimes()
		gitRepo.EXPECT().GetWorkTreeDir().DoAndReturn(func() string { return projectDir }).AnyTimes()

		pathMatcher.EXPECT().IsDirOrSubmodulePathMatched(gomock.Any()).Return(true).AnyTimes()
		pathMatcher.EXPECT().IsPathMatched(gomock.Any()).Return(true).AnyTimes()
		giterminismConfig.EXPECT().UncommittedHelmFilePathMatcher().Return(pathMatcher).AnyTimes()

		reader = file_reader.NewFileReader(sharedOptions)
		reader.SetGiterminismConfig(giterminismConfig)
	})

	writeChart := func(files map[string]string) string {
		chartDir := filepath.Join(projectDir, ".helm")
		for relPath, data := range files {
			absPath := filepath.Join(chartDir, relPath)
			Expect(os.MkdirAll(filepath.Dir(absPath), 0o755)).To(Succeed())
			Expect(os.WriteFile(absPath, []byte(data), 0o644)).To(Succeed())
		}
		return chartDir
	}

	loadedNames := func(ctx context.Context, chartDir string) []string {
		files, err := reader.LoadChartDir(ctx, chartDir)
		Expect(err).NotTo(HaveOccurred())

		names := make([]string, 0, len(files))
		for _, f := range files {
			names = append(names, f.Name)
		}
		return names
	}

	It("loads every file when the chart has no .helmignore", func(ctx SpecContext) {
		chartDir := writeChart(map[string]string{
			"Chart.yaml":            "name: test",
			"templates/kept.yaml":   "kept",
			"templates/second.yaml": "second",
		})

		Expect(loadedNames(logging.WithLogger(ctx), chartDir)).To(ConsistOf(
			"Chart.yaml", "templates/kept.yaml", "templates/second.yaml",
		))
	})

	It("drops the files matched by .helmignore", func(ctx SpecContext) {
		chartDir := writeChart(map[string]string{
			".helmignore":            ".helmignore\ntemplates/ignored.yaml\n",
			"Chart.yaml":             "name: test",
			"templates/kept.yaml":    "kept",
			"templates/ignored.yaml": "ignored",
		})

		Expect(loadedNames(logging.WithLogger(ctx), chartDir)).To(ConsistOf(
			"Chart.yaml", "templates/kept.yaml",
		))
	})

	It("drops a whole directory matched by .helmignore", func(ctx SpecContext) {
		chartDir := writeChart(map[string]string{
			".helmignore":            "ignoreddir/\n",
			"Chart.yaml":             "name: test",
			"ignoreddir/a.yaml":      "a",
			"ignoreddir/deep/b.yaml": "b",
			"templates/kept.yaml":    "kept",
		})

		Expect(loadedNames(logging.WithLogger(ctx), chartDir)).To(ConsistOf(
			".helmignore", "Chart.yaml", "templates/kept.yaml",
		))
	})

	It("drops the dotfiles under templates without a .helmignore", func(ctx SpecContext) {
		chartDir := writeChart(map[string]string{
			"Chart.yaml":          "name: test",
			"templates/.gitkeep":  "",
			"templates/kept.yaml": "kept",
		})

		Expect(loadedNames(logging.WithLogger(ctx), chartDir)).To(ConsistOf(
			"Chart.yaml", "templates/kept.yaml",
		))
	})

	It("does not resolve a symlink inside an ignored directory", func(ctx SpecContext) {
		chartDir := writeChart(map[string]string{
			".helmignore": "ignoreddir/\n",
			"Chart.yaml":  "name: test",
		})

		outsideTarget := filepath.Join(t.TempDir(), "outside.txt")
		Expect(os.WriteFile(outsideTarget, []byte("outside"), 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(chartDir, "ignoreddir"), 0o755)).To(Succeed())
		Expect(os.Symlink(outsideTarget, filepath.Join(chartDir, "ignoreddir", "link.txt"))).To(Succeed())

		Expect(loadedNames(logging.WithLogger(ctx), chartDir)).To(ConsistOf(
			".helmignore", "Chart.yaml",
		))
	})

	// The rules must be matched against the chart-relative path with the symlink parts kept,
	// the way helm names the file inside the chart, and not against the resolved path.
	It("applies the rules when a chart subdirectory is a symlink", func(ctx SpecContext) {
		chartDir := writeChart(map[string]string{
			".helmignore": "templates/ignored.yaml\n",
			"Chart.yaml":  "name: test",
		})

		sharedDir := filepath.Join(projectDir, "shared")
		Expect(os.MkdirAll(sharedDir, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(sharedDir, "ignored.yaml"), []byte("ignored"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(sharedDir, "kept.yaml"), []byte("kept"), 0o644)).To(Succeed())
		Expect(os.Symlink(sharedDir, filepath.Join(chartDir, "templates"))).To(Succeed())

		Expect(loadedNames(logging.WithLogger(ctx), chartDir)).To(ConsistOf(
			".helmignore", "Chart.yaml", "templates/kept.yaml",
		))
	})

	It("applies the rules when the chart directory itself is a symlink", func(ctx SpecContext) {
		realChartDir := filepath.Join(projectDir, "charts", "mychart")
		for relPath, data := range map[string]string{
			".helmignore":            "templates/ignored.yaml\n",
			"Chart.yaml":             "name: test",
			"templates/ignored.yaml": "ignored",
			"templates/kept.yaml":    "kept",
		} {
			absPath := filepath.Join(realChartDir, relPath)
			Expect(os.MkdirAll(filepath.Dir(absPath), 0o755)).To(Succeed())
			Expect(os.WriteFile(absPath, []byte(data), 0o644)).To(Succeed())
		}

		chartDir := filepath.Join(projectDir, ".helm")
		Expect(os.Symlink(realChartDir, chartDir)).To(Succeed())

		Expect(loadedNames(logging.WithLogger(ctx), chartDir)).To(ConsistOf(
			".helmignore", "Chart.yaml", "templates/kept.yaml",
		))
	})

	It("reports an unparsable .helmignore", func(ctx SpecContext) {
		chartDir := writeChart(map[string]string{
			".helmignore": "templates/**/ignored.yaml\n",
			"Chart.yaml":  "name: test",
		})

		_, err := reader.LoadChartDir(logging.WithLogger(ctx), chartDir)
		Expect(err).To(MatchError(ContainSubstring("double-star")))
	})

	Context("when giterminism is enforced", func() {
		// An ignored file must not be checked against git at all, otherwise .helmignore
		// cannot be used to exclude uncommitted files as the giterminism docs promise.
		BeforeEach(func() {
			sharedOptions.EXPECT().LooseGiterminism().Return(false).AnyTimes()
			pathMatcher.EXPECT().IsPathMatched(gomock.Any()).Return(false).AnyTimes()

			gitRepo.EXPECT().ValidateStatusResult(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, pathMatcher path_matcher.PathMatcher) error {
					if strings.Contains(pathMatcher.String(), "untracked.yaml") {
						return git_repo.UntrackedFilesFoundError{PathList: []string{".helm/templates/untracked.yaml"}}
					}
					return nil
				}).AnyTimes()

			gitRepo.EXPECT().IsCommitTreeEntryExist(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()
			gitRepo.EXPECT().ListCommitFilesWithGlob(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _, dir, _ string) ([]string, error) {
					var list []string
					err := filepath.WalkDir(filepath.Join(projectDir, dir), func(path string, d fs.DirEntry, err error) error {
						if err != nil || d.IsDir() {
							return err
						}
						rel, err := filepath.Rel(projectDir, path)
						if err != nil {
							return err
						}
						list = append(list, filepath.ToSlash(rel))
						return nil
					})
					return list, err
				}).AnyTimes()
			readFromWorkTree := func(_ context.Context, _, relPath string) ([]byte, error) {
				return os.ReadFile(filepath.Join(projectDir, relPath))
			}
			gitRepo.EXPECT().ReadCommitTreeEntryContent(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(readFromWorkTree).AnyTimes()
			gitRepo.EXPECT().ReadCommitFile(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(readFromWorkTree).AnyTimes()
			gitRepo.EXPECT().IsCommitTreeEntryDirectory(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
			gitRepo.EXPECT().ResolveAndCheckCommitFilePath(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _, relPath string, _ func(string) error) (string, error) {
					return relPath, nil
				}).AnyTimes()
		})

		It("fails on an uncommitted file that .helmignore does not cover", func(ctx SpecContext) {
			chartDir := writeChart(map[string]string{
				".helmignore":              "unrelated.yaml\n",
				"Chart.yaml":               "name: test",
				"templates/untracked.yaml": "untracked",
			})

			_, err := reader.LoadChartDir(logging.WithLogger(ctx), chartDir)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("untracked"))
		})

		It("loads the chart when .helmignore covers the uncommitted file", func(ctx SpecContext) {
			chartDir := writeChart(map[string]string{
				".helmignore":              "templates/untracked.yaml\n",
				"Chart.yaml":               "name: test",
				"templates/untracked.yaml": "untracked",
			})

			Expect(loadedNames(logging.WithLogger(ctx), chartDir)).To(ConsistOf(
				".helmignore", "Chart.yaml",
			))
		})
	})
})
