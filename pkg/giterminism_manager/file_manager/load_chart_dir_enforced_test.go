package filemanager

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/werf/werf/v2/pkg/git_repo"
	"github.com/werf/werf/v2/pkg/giterminism_manager/file_reader"
	"github.com/werf/werf/v2/pkg/includes"
	"github.com/werf/werf/v2/pkg/logging"
	"github.com/werf/werf/v2/pkg/path_matcher"
	"github.com/werf/werf/v2/test/mock"
)

// FileManager assembles the chart from the local source and the includes, and under enforced
// giterminism the local half arrives as a flat commit listing rather than a filesystem walk. That
// path has its own filtering, so it needs its own coverage.
var _ = Describe("LoadChartDir under enforced giterminism", func() {
	t := GinkgoT()

	const includeCommit = "include commit"

	var (
		projectDir string
		reader     file_reader.FileReader
		gitRepo    *mock.MockGitRepo
		repo       *mock.MockGitRepository
		// commitOnlyFiles are served by the commit listing without existing in the worktree.
		commitOnlyFiles map[string]string
		includeFiles    map[string]string
		readPaths       map[string]bool
	)

	BeforeEach(func() {
		ctrl := gomock.NewController(t, gomock.WithOverridableExpectations())

		projectDir = t.TempDir()
		commitOnlyFiles = make(map[string]string)
		readPaths = make(map[string]bool)

		gitRepo = mock.NewMockGitRepo(ctrl)
		repo = mock.NewMockGitRepository(ctrl)
		repo.EXPECT().GetName().Return("include repo").AnyTimes()
		includeFiles = make(map[string]string)
		repo.EXPECT().ReadCommitFile(gomock.Any(), includeCommit, gomock.Any()).
			DoAndReturn(func(_ context.Context, _, relPath string) ([]byte, error) {
				data, ok := includeFiles[filepath.ToSlash(relPath)]
				if !ok {
					return nil, fmt.Errorf("file not found in include: %s", relPath)
				}
				return []byte(data), nil
			}).AnyTimes()

		reader = file_reader.NewFileReader(fakeSharedOptions{
			projectDir:          projectDir,
			localGitRepo:        gitRepo,
			enforcedGiterminism: true,
		})
		reader.SetGiterminismConfig(fakeGiterminismConfig{uncommittedHelmFilesRejected: true})

		gitRepo.EXPECT().GetName().Return("local repo").AnyTimes()
		gitRepo.EXPECT().GetWorkTreeDir().DoAndReturn(func() string { return projectDir }).AnyTimes()
		gitRepo.EXPECT().IsCommitTreeEntryExist(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()
		gitRepo.EXPECT().IsCommitTreeEntryDirectory(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
		gitRepo.EXPECT().IsCommitFileExist(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _, relPath string) (bool, error) {
				if _, ok := commitOnlyFiles[filepath.ToSlash(relPath)]; ok {
					return true, nil
				}
				_, err := os.Stat(filepath.Join(projectDir, relPath))
				return err == nil, nil
			}).AnyTimes()
		gitRepo.EXPECT().StatusPathList(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		gitRepo.EXPECT().ValidateStatusResult(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, matcher path_matcher.PathMatcher) error {
				if strings.Contains(matcher.String(), "uncommitted.yaml") {
					return git_repo.UntrackedFilesFoundError{PathList: []string{".helm/templates/uncommitted.yaml"}}
				}
				return nil
			}).AnyTimes()
		gitRepo.EXPECT().ResolveAndCheckCommitFilePath(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _, relPath string, _ func(string) error) (string, error) {
				return relPath, nil
			}).AnyTimes()

		// The commit listing is the union of the worktree and the commit-only files, so a file
		// deleted from the worktree still reaches the filter the way a real commit would.
		gitRepo.EXPECT().ListCommitFilesWithGlob(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _, dir, _ string) ([]string, error) {
				var list []string
				root := filepath.Join(projectDir, dir)
				if _, err := os.Stat(root); err == nil {
					if err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
						if err != nil || d.IsDir() {
							return err
						}
						rel, err := filepath.Rel(root, path)
						if err != nil {
							return err
						}
						list = append(list, filepath.ToSlash(filepath.Join(dir, rel)))
						return nil
					}); err != nil {
						return nil, err
					}
				}
				for relPath := range commitOnlyFiles {
					if strings.HasPrefix(relPath, filepath.ToSlash(dir)+"/") {
						list = append(list, relPath)
					}
				}
				return list, nil
			}).AnyTimes()

		readFromCommit := func(_ context.Context, _, relPath string) ([]byte, error) {
			norm := filepath.ToSlash(relPath)
			readPaths[norm] = true
			if data, ok := commitOnlyFiles[norm]; ok {
				return []byte(data), nil
			}
			return os.ReadFile(filepath.Join(projectDir, relPath))
		}
		gitRepo.EXPECT().ReadCommitTreeEntryContent(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(readFromCommit).AnyTimes()
		gitRepo.EXPECT().ReadCommitFile(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(readFromCommit).AnyTimes()
	})

	writeLocalChart := func(files map[string]string) {
		for relPath, data := range files {
			absPath := filepath.Join(projectDir, ".helm", relPath)
			Expect(os.MkdirAll(filepath.Dir(absPath), 0o755)).To(Succeed())
			Expect(os.WriteFile(absPath, []byte(data), 0o644)).To(Succeed())
		}
	}

	newInclude := func(files map[string]string) *includes.Include {
		objects := make(map[string]string, len(files))
		for relPath, data := range files {
			mountPath := filepath.ToSlash(filepath.Join(".helm", relPath))
			objects[mountPath] = mountPath
			includeFiles[mountPath] = data
		}
		return includes.NewInclude(repo, includeCommit, objects)
	}

	newFileManager := func(includeList ...*includes.Include) *FileManager {
		return &FileManager{
			fileReader:       reader,
			includes:         includeList,
			caches:           &caches{dockerFiles: make(map[string][]byte)},
			customProjectDir: projectDir,
		}
	}

	loadedNames := func(ctx context.Context, manager *FileManager) ([]string, error) {
		files, err := manager.LoadChartDir(ctx, ".helm")
		if err != nil {
			return nil, err
		}
		names := make([]string, 0, len(files))
		for _, file := range files {
			names = append(names, file.Name)
		}
		return names, nil
	}

	// The regression guard for the diagnostic: an excluded file must not be demanded as a commit,
	// which is what the removed error-path re-read with helm's defaults used to do.
	It("blames .helmignore rather than the uncommitted file it excludes", func(ctx SpecContext) {
		writeLocalChart(map[string]string{
			".helmignore":                "*\n",
			"Chart.yaml":                 "name: local",
			"templates/uncommitted.yaml": "broken",
		})

		_, err := newFileManager().LoadChartDir(logging.WithLogger(ctx), ".helm")
		Expect(err).To(HaveOccurred())
		Expect(err).NotTo(MatchError(ContainSubstring("must be committed")))
		Expect(err).To(MatchError(ContainSubstring(".helmignore")))
	})

	It("merges the local chart with the includes, letting the local file win", func(ctx SpecContext) {
		writeLocalChart(map[string]string{
			"Chart.yaml":            "name: local",
			"templates/shared.yaml": "local",
		})
		include := newInclude(map[string]string{
			"templates/shared.yaml":   "included",
			"templates/imported.yaml": "included",
		})

		files, err := newFileManager(include).LoadChartDir(logging.WithLogger(ctx), ".helm")
		Expect(err).NotTo(HaveOccurred())

		contents := make(map[string]string, len(files))
		for _, file := range files {
			contents[file.Name] = string(file.Data)
		}
		Expect(contents).To(HaveLen(3))
		Expect(contents).To(HaveKeyWithValue("Chart.yaml", "name: local"))
		Expect(contents).To(HaveKeyWithValue("templates/imported.yaml", "included"))
		Expect(contents).To(HaveKeyWithValue("templates/shared.yaml", "local"))
	})

	// A directory rule cannot match a file path, so the flat commit listing needs the parent
	// directories checked explicitly — and the excluded file must never be read.
	It("excludes a commit-only file matched by a directory rule without reading it", func(ctx SpecContext) {
		writeLocalChart(map[string]string{
			".helmignore": "secrets/\n",
			"Chart.yaml":  "name: local",
		})
		// The directories exist in the worktree; only the files are commit-only, which is what a
		// file deleted from the worktree but still present in the commit looks like.
		Expect(os.MkdirAll(filepath.Join(projectDir, ".helm", "secrets"), 0o755)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(projectDir, ".helm", "templates"), 0o755)).To(Succeed())
		commitOnlyFiles[".helm/secrets/token.yaml"] = "secret"
		commitOnlyFiles[".helm/templates/kept.yaml"] = "kept"

		names, err := loadedNames(logging.WithLogger(ctx), newFileManager())
		Expect(err).NotTo(HaveOccurred())
		Expect(names).To(ConsistOf(".helmignore", "Chart.yaml", "templates/kept.yaml"))
		Expect(readPaths).NotTo(HaveKey(".helm/secrets/token.yaml"),
			fmt.Sprintf("an excluded file must be filtered before it is read, got reads: %v", readPaths))
	})
})
