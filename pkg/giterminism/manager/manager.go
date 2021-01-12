package manager

import (
	"github.com/werf/werf/pkg/git_repo"
	"github.com/werf/werf/pkg/giterminism"
	"github.com/werf/werf/pkg/giterminism/config"
	"github.com/werf/werf/pkg/giterminism/file_reader"
)

type Manager struct {
	projectDir   string
	headCommit   string
	localGitRepo git_repo.Local

	fileReader giterminism.FileReader
	config     giterminism.Config

	devMode                        bool
	looseGiterminism               bool
	nonStrictGiterminismInspection bool
}

type NewManagerOptions struct {
	LooseGiterminism bool
}

func NewManager(projectDir string, localGitRepo git_repo.Local, headCommit string, options NewManagerOptions) (giterminism.Manager, error) {
	m := Manager{
		projectDir:       projectDir,
		localGitRepo:     localGitRepo,
		headCommit:       headCommit,
		looseGiterminism: options.LooseGiterminism,
	}

	c, err := config.NewConfig(projectDir)
	if err != nil {
		return nil, err
	}

	m.config = c
	m.fileReader = file_reader.NewFileReader(m)

	return m, nil
}

func (m Manager) ProjectDir() string {
	return m.projectDir
}

func (m Manager) HeadCommit() string {
	return m.headCommit
}

func (m Manager) LocalGitRepo() *git_repo.Local {
	return &m.localGitRepo
}

func (m Manager) FileReader() giterminism.FileReader {
	return m.fileReader
}

func (m Manager) Config() giterminism.Config {
	return m.config
}

func (m Manager) LooseGiterminism() bool {
	return m.looseGiterminism
}
