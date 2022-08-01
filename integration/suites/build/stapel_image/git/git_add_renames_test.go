package git_test

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/alessio/shellescape"
	. "github.com/onsi/ginkgo/v2"
	"gopkg.in/yaml.v3"

	"github.com/werf/werf/test/pkg/utils"
	"github.com/werf/werf/test/pkg/utils/docker"
)

var _ = Describe("Git add file renames", func() {
	rootOfHostFixtures := "git_add_renames"
	rootOfMountedInContainerGitRepo := "/host"
	rootOfBuiltIntoImageGitToFiles := "/"

	createCommandsFunc := func() []string {
		simpleGitMappings := getSimpleGitMappingsFromFirstImageInWerfConfig(filepath.FromSlash(filepath.Join(SuiteData.TestDirPath, "werf.yaml")))

		var commands []string
		for _, simpleGitMapping := range simpleGitMappings {
			gitAddHostFilePath := filepath.FromSlash(filepath.Join(SuiteData.TestDirPath, simpleGitMapping.Add))
			gitAddHostFileInfo, err := os.Lstat(gitAddHostFilePath)
			if err != nil {
				Fail(err.Error())
			}

			var expectedPerms uint64
			switch mode := gitAddHostFileInfo.Mode(); {
			case mode.IsDir():
				expectedPerms = uint64(os.FileMode(0o755))
			case mode.IsRegular():
				expectedPerms = uint64(os.FileMode(0o644))
			case mode&os.ModeSymlink == os.ModeSymlink:
				expectedPerms = uint64(os.FileMode(0o777))
			default:
				Fail(fmt.Sprintf("unexpected file mode for gitAddHostFilePath %q: %s", gitAddHostFilePath, mode))
			}

			mountedInContainerGitAddPath := shellescape.Quote(path.Join(rootOfMountedInContainerGitRepo, simpleGitMapping.Add))
			builtIntoImageGitToPath := shellescape.Quote(path.Join(rootOfBuiltIntoImageGitToFiles, simpleGitMapping.To))

			commands = append(
				commands,
				fmt.Sprintf("echo 'Checking mounted git[].add path %q and corresponding git[].to path %q inside of the container'", mountedInContainerGitAddPath, builtIntoImageGitToPath),
				fmt.Sprintf("diff <(echo %s) <(stat -c %%a %s)", strconv.FormatUint(expectedPerms, 8), builtIntoImageGitToPath),
				fmt.Sprintf("diff -r --strip-trailing-cr --no-dereference %s %s", mountedInContainerGitAddPath, builtIntoImageGitToPath),
			)
		}
		return commands
	}

	doByTestFunc := func(fixtureRelPath string) {
		filePatternsToCommit := []string{"werf.yaml"}

		utils.RunSucceedCommand(
			SuiteData.TestDirPath,
			"git",
			"rm", "--ignore-unmatch", "-rf", ".",
		)

		utils.CopyIn(utils.FixturePath(rootOfHostFixtures, fixtureRelPath), SuiteData.TestDirPath)

		simpleGitMappings := getSimpleGitMappingsFromFirstImageInWerfConfig(filepath.FromSlash(filepath.Join(SuiteData.TestDirPath, "werf.yaml")))
		for _, simpleGitMapping := range simpleGitMappings {
			filePatternsToCommit = append(filePatternsToCommit, filepath.Clean(filepath.Join(SuiteData.TestDirPath, simpleGitMapping.Add)))
		}

		utils.RunSucceedCommand(
			SuiteData.TestDirPath,
			"git",
			append(
				[]string{"add"},
				filePatternsToCommit...,
			)...,
		)

		utils.RunSucceedCommand(
			SuiteData.TestDirPath,
			"git",
			"commit", "--allow-empty", "-m", "+",
		)

		utils.RunSucceedCommand(
			SuiteData.TestDirPath,
			SuiteData.WerfBinPath,
			"build", "--debug",
		)

		extraDockerOptions := []string{
			fmt.Sprintf("-v %s:%s", SuiteData.TestDirPath, rootOfMountedInContainerGitRepo),
		}

		docker.RunSucceedContainerCommandWithStapel(
			SuiteData.WerfBinPath,
			SuiteData.TestDirPath,
			extraDockerOptions,
			createCommandsFunc(),
		)
	}

	BeforeEach(func() {
		utils.RunSucceedCommand(
			SuiteData.TestDirPath,
			"git",
			"init",
		)
	})

	It("should be handled correctly", func() {
		By(
			"Building with renamed files, dirs and symlinks when they are added with git[].add's different from corresponding git[].to's",
			func() { doByTestFunc("commit001-initial") },
		)

		By(
			"Rebuilding with some files, dirs and symlinks removed from git repository and werf.yaml",
			func() { doByTestFunc("commit002-remove-some-files") },
		)

		By(
			"Rebuilding with contents of files, dirs and symlinks changed",
			func() { doByTestFunc("commit003-change-files-content") },
		)

		By(
			"Rebuilding with renamed for the second time files, dirs, and symlinks",
			func() { doByTestFunc("commit004-change-filenames-again") },
		)
	})
})

type simpleImageConfig struct {
	Image string              `yaml:"image,omitempty"`
	Git   []*simpleGitMapping `yaml:"git,omitempty"`
}

type simpleGitMapping struct {
	Add string `yaml:"add,omitempty"`
	To  string `yaml:"to,omitempty"`
}

func getSimpleGitMappingsFromFirstImageInWerfConfig(werfConfigPath string) []*simpleGitMapping {
	werfConfigFile, err := os.Open(werfConfigPath)
	if err != nil {
		Fail(err.Error())
	}

	yamlDecoder := yaml.NewDecoder(werfConfigFile)
	var simpleGitMappings []*simpleGitMapping
	for {
		simpleImageCfg := new(simpleImageConfig)

		if err := yamlDecoder.Decode(&simpleImageCfg); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			Fail(err.Error())
		}

		if simpleImageCfg.Image == "" || simpleImageCfg.Git == nil {
			continue
		}

		simpleGitMappings = simpleImageCfg.Git
		break
	}

	if simpleGitMappings == nil {
		Fail(fmt.Sprintf("no image git mappings found in %q", werfConfigPath))
	}

	return simpleGitMappings
}
