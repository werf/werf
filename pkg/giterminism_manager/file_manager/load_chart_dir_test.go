package filemanager

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/werf/werf/v2/pkg/giterminism_manager/file_reader"
	"github.com/werf/werf/v2/pkg/includes"
	"github.com/werf/werf/v2/pkg/logging"
	"github.com/werf/werf/v2/test/mock"
)

var _ = Describe("LoadChartDir", func() {
	t := GinkgoT()

	const commitHash = "include commit"

	var (
		projectDir string
		reader     file_reader.FileReader
		repo       *mock.MockGitRepository
	)

	BeforeEach(func() {
		ctrl := gomock.NewController(t)

		projectDir = t.TempDir()
		repo = mock.NewMockGitRepository(ctrl)
		repo.EXPECT().GetName().Return("include repo").AnyTimes()

		reader = file_reader.NewFileReader(fakeSharedOptions{projectDir: projectDir})
		reader.SetGiterminismConfig(fakeGiterminismConfig{})
	})

	writeLocalChart := func(files map[string]string) {
		for relPath, data := range files {
			absPath := filepath.Join(projectDir, ".helm", relPath)
			Expect(os.MkdirAll(filepath.Dir(absPath), 0o755)).To(Succeed())
			Expect(os.WriteFile(absPath, []byte(data), 0o644)).To(Succeed())
		}
	}

	// The include is keyed by the chart-relative path the file is mounted at, and its content is
	// served from the repository the same way a real include reads it out of a commit.
	newInclude := func(files map[string]string) *includes.Include {
		objects := make(map[string]string, len(files))
		for relPath, data := range files {
			mountPath := filepath.ToSlash(filepath.Join(".helm", relPath))
			objects[mountPath] = mountPath
			repo.EXPECT().ReadCommitFile(gomock.Any(), commitHash, mountPath).
				Return([]byte(data), nil).AnyTimes()
		}
		return includes.NewInclude(repo, commitHash, objects)
	}

	newFileManager := func(includeList ...*includes.Include) *FileManager {
		return &FileManager{
			fileReader:       reader,
			includes:         includeList,
			caches:           &caches{dockerFiles: make(map[string][]byte)},
			customProjectDir: projectDir,
		}
	}

	loadedNames := func(ctx context.Context, manager *FileManager) []string {
		files, err := manager.LoadChartDir(ctx, ".helm")
		Expect(err).NotTo(HaveOccurred())

		names := make([]string, 0, len(files))
		for _, f := range files {
			names = append(names, f.Name)
		}
		return names
	}

	It("merges the local chart with the includes when there is no .helmignore", func(ctx SpecContext) {
		writeLocalChart(map[string]string{
			"Chart.yaml":           "name: local",
			"templates/local.yaml": "local",
		})
		include := newInclude(map[string]string{
			"templates/local.yaml":    "from include",
			"templates/imported.yaml": "imported",
		})

		Expect(loadedNames(logging.WithLogger(ctx), newFileManager(include))).To(ConsistOf(
			"Chart.yaml", "templates/local.yaml", "templates/imported.yaml",
		))
	})

	It("keeps the local file when the same path comes from an include", func(ctx SpecContext) {
		writeLocalChart(map[string]string{
			"Chart.yaml":           "name: local",
			"templates/local.yaml": "local wins",
		})
		include := newInclude(map[string]string{"templates/local.yaml": "from include"})

		files, err := newFileManager(include).LoadChartDir(logging.WithLogger(ctx), ".helm")
		Expect(err).NotTo(HaveOccurred())

		for _, f := range files {
			if f.Name == "templates/local.yaml" {
				Expect(string(f.Data)).To(Equal("local wins"))
			}
		}
	})

	It("drops an imported file matched by the local .helmignore", func(ctx SpecContext) {
		writeLocalChart(map[string]string{
			".helmignore": "templates/imported.yaml\n",
			"Chart.yaml":  "name: local",
		})
		include := newInclude(map[string]string{
			"templates/imported.yaml": "imported",
			"templates/kept.yaml":     "kept",
		})

		Expect(loadedNames(logging.WithLogger(ctx), newFileManager(include))).To(ConsistOf(
			".helmignore", "Chart.yaml", "templates/kept.yaml",
		))
	})

	// Without this, a path excluded from the local source is indistinguishable from a path that
	// was never there, so the include loop adds it back and silently substitutes its content.
	It("does not bring an excluded local file back from an include", func(ctx SpecContext) {
		writeLocalChart(map[string]string{
			".helmignore":            "templates/ignored.yaml\n",
			"Chart.yaml":             "name: local",
			"templates/ignored.yaml": "local ignored",
		})
		include := newInclude(map[string]string{"templates/ignored.yaml": "from include"})

		Expect(loadedNames(logging.WithLogger(ctx), newFileManager(include))).To(ConsistOf(
			".helmignore", "Chart.yaml",
		))
	})

	It("drops an imported file under a directory rule of the local .helmignore", func(ctx SpecContext) {
		writeLocalChart(map[string]string{
			".helmignore": "ignoreddir/\n",
			"Chart.yaml":  "name: local",
		})
		include := newInclude(map[string]string{
			"ignoreddir/a.yaml":      "a",
			"ignoreddir/deep/b.yaml": "b",
			"templates/kept.yaml":    "kept",
		})

		Expect(loadedNames(logging.WithLogger(ctx), newFileManager(include))).To(ConsistOf(
			".helmignore", "Chart.yaml", "templates/kept.yaml",
		))
	})

	// The chart has no local directory at all, so the fix has to reach the .helmignore that the
	// include itself delivers, otherwise such a chart is never filtered and werf keeps diverging
	// from helm on it.
	It("applies the .helmignore delivered by the include when the chart is not local", func(ctx SpecContext) {
		include := newInclude(map[string]string{
			".helmignore":             "templates/imported.yaml\n",
			"Chart.yaml":              "name: imported",
			"templates/imported.yaml": "imported",
			"templates/kept.yaml":     "kept",
		})

		Expect(loadedNames(logging.WithLogger(ctx), newFileManager(include))).To(ConsistOf(
			".helmignore", "Chart.yaml", "templates/kept.yaml",
		))
	})

	It("prefers the local .helmignore over the one from an include", func(ctx SpecContext) {
		writeLocalChart(map[string]string{
			".helmignore": "templates/local-rule.yaml\n",
			"Chart.yaml":  "name: local",
		})
		include := newInclude(map[string]string{
			".helmignore":                 "templates/include-rule.yaml\n",
			"templates/local-rule.yaml":   "excluded by the local rule",
			"templates/include-rule.yaml": "kept, the include rule does not apply",
		})

		Expect(loadedNames(logging.WithLogger(ctx), newFileManager(include))).To(ConsistOf(
			".helmignore", "Chart.yaml", "templates/include-rule.yaml",
		))
	})

	It("hints at .helmignore when the chart has one and nothing is left", func(ctx SpecContext) {
		writeLocalChart(map[string]string{
			".helmignore": "*\n",
			"Chart.yaml":  "name: local",
		})

		_, err := newFileManager().LoadChartDir(logging.WithLogger(ctx), ".helm")
		Expect(err).To(MatchError(ContainSubstring("not found in the project git repository or includes")))
		Expect(err).To(MatchError(ContainSubstring("the chart has a .helmignore")))
	})

	It("reports a missing chart directory as not found rather than excluded", func(ctx SpecContext) {
		_, err := newFileManager().LoadChartDir(logging.WithLogger(ctx), ".helm")
		Expect(err).To(MatchError(ContainSubstring("not found in the project git repository or includes")))
	})

	// An existing but empty directory has nothing to exclude, so blaming .helmignore there would
	// send the user looking for a rule that does not exist.
	It("reports an empty chart directory as not found rather than excluded", func(ctx SpecContext) {
		Expect(os.MkdirAll(filepath.Join(projectDir, ".helm"), 0o755)).To(Succeed())

		_, err := newFileManager().LoadChartDir(logging.WithLogger(ctx), ".helm")
		Expect(err).To(MatchError(ContainSubstring("not found in the project git repository or includes")))
	})

	// Helm's own defaults empty this chart, so there is no .helmignore to send the user looking for.
	It("reports a chart emptied by helm's defaults without hinting at .helmignore", func(ctx SpecContext) {
		writeLocalChart(map[string]string{"templates/.gitkeep": ""})

		_, err := newFileManager().LoadChartDir(logging.WithLogger(ctx), ".helm")
		Expect(err).To(MatchError(ContainSubstring("not found in the project git repository or includes")))
		Expect(err).NotTo(MatchError(ContainSubstring("the chart has a .helmignore")))
	})

	It("reports an included chart emptied by helm's defaults without hinting at .helmignore", func(ctx SpecContext) {
		include := newInclude(map[string]string{"templates/.gitkeep": ""})

		_, err := newFileManager(include).LoadChartDir(logging.WithLogger(ctx), ".helm")
		Expect(err).To(MatchError(ContainSubstring("not found in the project git repository or includes")))
		Expect(err).NotTo(MatchError(ContainSubstring("the chart has a .helmignore")))
	})

	It("hints at .helmignore when the only one comes from an include", func(ctx SpecContext) {
		include := newInclude(map[string]string{
			".helmignore": "*\n",
			"Chart.yaml":  "name: included",
		})

		_, err := newFileManager(include).LoadChartDir(logging.WithLogger(ctx), ".helm")
		Expect(err).To(MatchError(ContainSubstring("not found in the project git repository or includes")))
		Expect(err).To(MatchError(ContainSubstring("the chart has a .helmignore")))
	})

	It("applies helm's default rules to imported files", func(ctx SpecContext) {
		writeLocalChart(map[string]string{"Chart.yaml": "name: local"})
		include := newInclude(map[string]string{
			"templates/.gitkeep":  "",
			"templates/kept.yaml": "kept",
		})

		Expect(loadedNames(logging.WithLogger(ctx), newFileManager(include))).To(ConsistOf(
			"Chart.yaml", "templates/kept.yaml",
		))
	})
})
