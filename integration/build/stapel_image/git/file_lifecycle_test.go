package git_test

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/alessio/shellescape"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/integration/utils"
	"github.com/werf/werf/integration/utils/docker"
)

var _ = Describe("file lifecycle", func() {
	var fixturesPathParts []string
	gitToPath := "/app"

	fileDataToAdd := []byte("test")
	fileDataToModify := []byte("test2")

	gitExecutableFilePerm := os.FileMode(0755)
	gitOrdinaryFilePerm := os.FileMode(0644)

	type fileLifecycleEntry struct {
		relPath string
		data    []byte
		perm    os.FileMode
		delete  bool
		devMode bool
	}

	createFileFunc := func(fileName string, fileData []byte, filePerm os.FileMode) {
		filePath := filepath.Join(testDirPath, fileName)
		utils.CreateFile(filePath, fileData)

		if runtime.GOOS == "windows" {
			gitArgs := []string{"add"}
			if filePerm == gitExecutableFilePerm {
				gitArgs = append(gitArgs, "--chmod=+x")
			} else {
				gitArgs = append(gitArgs, "--chmod=-x")
			}
			gitArgs = append(gitArgs, fileName)

			utils.RunSucceedCommand(
				testDirPath,
				"git",
				gitArgs...,
			)
		} else {
			Ω(os.Chmod(filePath, filePerm)).Should(Succeed())
		}
	}

	fileLifecycleEntryItBody := func(entry fileLifecycleEntry) {
		var commitMsg string
		filePath := filepath.Join(testDirPath, entry.relPath)
		if entry.delete {
			Ω(os.Remove(filePath)).Should(Succeed())
			commitMsg = "Delete file " + entry.relPath
		} else {
			createFileFunc(entry.relPath, entry.data, entry.perm)
			commitMsg = "Add/Modify file " + entry.relPath
		}

		if entry.devMode {
			stubs.SetEnv("WERF_DEV", "1")
			addFile(testDirPath, entry.relPath)
		} else {
			addAndCommitFile(testDirPath, entry.relPath, commitMsg)
		}

		utils.RunSucceedCommand(
			testDirPath,
			werfBinPath,
			"build",
		)

		var cmd []string
		var extraDockerOptions []string
		if entry.delete {
			cmd = append(cmd, docker.CheckContainerFileCommand(path.Join(gitToPath, entry.relPath), false, false))
		} else {
			cmd = append(cmd, docker.CheckContainerFileCommand(path.Join(gitToPath, entry.relPath), false, true))
			cmd = append(cmd, fmt.Sprintf("diff <(stat -c %%a %s) <(echo %s)", shellescape.Quote(path.Join(gitToPath, entry.relPath)), strconv.FormatUint(uint64(entry.perm), 8)))
			cmd = append(cmd, fmt.Sprintf("diff %s %s", shellescape.Quote(path.Join(gitToPath, entry.relPath)), shellescape.Quote(path.Join("/host", entry.relPath))))

			extraDockerOptions = append(extraDockerOptions, fmt.Sprintf("-v %s:%s", testDirPath, "/host"))
		}

		docker.RunSucceedContainerCommandWithStapel(
			werfBinPath,
			testDirPath,
			extraDockerOptions,
			cmd,
		)
	}

	BeforeEach(func() {
		fixturesPathParts = []string{"file_lifecycle"}
		commonBeforeEach(testDirPath, utils.FixturePath(fixturesPathParts...))
	})

	type test struct {
		relPathToAdd          string
		relPathToAddAndModify string
	}

	tests := []test{
		{
			"test",
			"test2",
		},
		{
			"dir/test",
			"dir/test2",
		},
	}

	if runtime.GOOS == "windows" {
		tests = append(tests, []test{
			{
				"普 通 话",
				"华语",
			},
			{
				"普 通 话/华语",
				"普 通 话/华语 2",
			},
		}...)
	} else {
		tests = append(tests, []test{
			{
				"file with !%s $chars один! два 'три' & ? .",
				"file with !%s $chars один! два 'три' & ? .. 2",
			},
			{
				"d i r/file with !%s $chars один! два 'три' & ? .",
				"d i r/file with !%s $chars один! два 'три' & ? .. 2",
			},
		}...)
	}

	for _, t := range tests {
		relPathToAdd := t.relPathToAdd
		relPathToAddAndModify := t.relPathToAddAndModify

		pathLogFunc := func(path string) string {
			return fmt.Sprintf(" (%s)", path)
		}

		forNormalAndDevMode(func(extraDescription string, devMode bool) {
			DescribeTable("processing file with archive apply"+extraDescription+pathLogFunc(relPathToAdd),
				fileLifecycleEntryItBody,
				Entry("should add file (0755)", fileLifecycleEntry{
					relPath: relPathToAdd,
					data:    fileDataToAdd,
					perm:    gitExecutableFilePerm,
					devMode: devMode,
				}),
				Entry("should add file (0644)", fileLifecycleEntry{
					relPath: relPathToAdd,
					data:    fileDataToAdd,
					perm:    gitOrdinaryFilePerm,
					devMode: devMode,
				}),
			)

			When("gitArchive stage with file is built"+extraDescription+pathLogFunc(relPathToAdd), func() {
				BeforeEach(func() {
					createFileFunc(relPathToAddAndModify, fileDataToAdd, gitExecutableFilePerm)
					addAndCommitFile(testDirPath, relPathToAddAndModify, "Add file "+relPathToAddAndModify)

					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						"build",
					)
				})

				DescribeTable("processing file with patch apply"+extraDescription,
					fileLifecycleEntryItBody,
					Entry("should add file (0755)", fileLifecycleEntry{
						relPath: relPathToAdd,
						data:    fileDataToAdd,
						perm:    gitExecutableFilePerm,
						devMode: devMode,
					}),
					Entry("should add file (0644)", fileLifecycleEntry{
						relPath: relPathToAdd,
						data:    fileDataToAdd,
						perm:    gitOrdinaryFilePerm,
						devMode: devMode,
					}),
					Entry("should modify file", fileLifecycleEntry{
						relPath: relPathToAddAndModify,
						data:    fileDataToModify,
						perm:    gitExecutableFilePerm,
						devMode: devMode,
					}),
					Entry("should change file permission (0755->0644)", fileLifecycleEntry{
						relPath: relPathToAddAndModify,
						data:    fileDataToAdd,
						perm:    gitOrdinaryFilePerm,
						devMode: devMode,
					}),
					Entry("should modify and change file permission (0755->0644)", fileLifecycleEntry{
						relPath: relPathToAddAndModify,
						data:    fileDataToModify,
						perm:    gitOrdinaryFilePerm,
						devMode: devMode,
					}),
					Entry("should delete file", fileLifecycleEntry{
						relPath: relPathToAddAndModify,
						delete:  true,
						devMode: devMode,
					}),
				)
			})

			When("file is symlink"+extraDescription+pathLogFunc(relPathToAdd), func() {
				linkToAdd := "werf.yaml"
				linkToModify := "none"

				type symlinkFileLifecycleEntry struct {
					relPath string
					link    string
					delete  bool
					devMode bool
				}

				symlinkFileLifecycleEntryItBody := func(entry symlinkFileLifecycleEntry) {
					var commitMsg string
					filePath := filepath.Join(testDirPath, entry.relPath)
					if entry.delete {
						Ω(os.Remove(filePath)).Should(Succeed())
						commitMsg = "Delete file " + entry.relPath
					} else {
						hashBytes, _ := utils.RunCommandWithOptions(
							testDirPath,
							"git",
							[]string{"hash-object", "-w", "--stdin"},
							utils.RunCommandOptions{
								ToStdin:       entry.link,
								ShouldSucceed: true,
							},
						)

						utils.RunSucceedCommand(
							testDirPath,
							"git",
							"update-index", "--add", "--cacheinfo", "120000", string(bytes.TrimSpace(hashBytes)), entry.relPath,
						)

						utils.RunSucceedCommand(
							testDirPath,
							"git",
							"checkout", entry.relPath,
						)

						commitMsg = "Add/Modify file " + entry.relPath
					}

					if entry.devMode {
						stubs.SetEnv("WERF_DEV", "1")
						addFile(testDirPath, entry.relPath)
					} else {
						addAndCommitFile(testDirPath, entry.relPath, commitMsg)
					}

					utils.RunSucceedCommand(
						testDirPath,
						werfBinPath,
						"build",
					)

					var cmd []string
					if entry.delete {
						cmd = append(cmd, checkContainerSymlinkFileCommand(path.Join(gitToPath, entry.relPath), false))
					} else {
						cmd = append(cmd, checkContainerSymlinkFileCommand(path.Join(gitToPath, entry.relPath), true))
						readlinkCmd := fmt.Sprintf("readlink %s", shellescape.Quote(path.Join(gitToPath, entry.relPath)))
						cmd = append(cmd, fmt.Sprintf("diff <(%s) <(echo %s)", readlinkCmd, shellescape.Quote(entry.link)))
					}

					docker.RunSucceedContainerCommandWithStapel(
						werfBinPath,
						testDirPath,
						[]string{},
						cmd,
					)
				}

				DescribeTable("processing symlink file with archive apply"+extraDescription,
					symlinkFileLifecycleEntryItBody,
					Entry("should add symlink", symlinkFileLifecycleEntry{
						relPath: relPathToAdd,
						link:    linkToAdd,
						devMode: devMode,
					}),
				)

				When("gitArchive stage with file is built"+extraDescription, func() {
					BeforeEach(func() {
						symlinkFileLifecycleEntryItBody(symlinkFileLifecycleEntry{
							relPath: relPathToAddAndModify,
							link:    linkToAdd,
						})
					})

					DescribeTable("processing symlink file with patch apply"+extraDescription,
						symlinkFileLifecycleEntryItBody,
						Entry("should add symlink", symlinkFileLifecycleEntry{
							relPath: relPathToAdd,
							link:    linkToAdd,
							devMode: devMode,
						}),
						Entry("should modify file", symlinkFileLifecycleEntry{
							relPath: relPathToAddAndModify,
							link:    linkToModify,
							devMode: devMode,
						}),
						Entry("should delete file", symlinkFileLifecycleEntry{
							relPath: relPathToAddAndModify,
							delete:  true,
							devMode: devMode,
						}))
				})
			})
		})
	}
})

func forNormalAndDevMode(f func(string, bool)) {
	for _, devMode := range []bool{false, true} {
		var extraDescription string
		if devMode {
			extraDescription = " in developer mode"
		}

		f(extraDescription, devMode)
	}
}

func checkContainerSymlinkFileCommand(containerDirPath string, exist bool) string {
	var cmd string

	if exist {
		cmd = fmt.Sprintf("test -h %s", shellescape.Quote(containerDirPath))
	} else {
		cmd = fmt.Sprintf("! test -h %s", shellescape.Quote(containerDirPath))
	}

	return cmd
}
